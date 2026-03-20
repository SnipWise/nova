package stores

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisConfig holds the configuration for Redis connection
type RedisConfig struct {
	Address   string // Redis server address (e.g., "localhost:6379")
	Password  string // Redis password (empty string for no password)
	DB        int    // Redis database number (default: 0)
	IndexName string // Name of the Redis search index (default: "nova_rag_index")
}

// RedisVectorStore implements VectorStore using Redis as the backend
// It uses Redis HNSW (Hierarchical Navigable Small World) indexing for efficient similarity search
type RedisVectorStore struct {
	client    *redis.Client
	ctx       context.Context
	config    RedisConfig
	dimension int // Embedding vector dimension (e.g., 384, 768, 1024, 3072)
}

// NewRedisVectorStore creates a new Redis-based vector store
// Parameters:
//   - ctx: context for Redis operations
//   - config: Redis connection configuration
//   - dimension: the dimension of embedding vectors (must match your embedding model)
//
// The function will:
//  1. Create a Redis client connection
//  2. Verify the connection with a PING
//  3. Create a vector search index if it doesn't exist
func NewRedisVectorStore(ctx context.Context, config RedisConfig, dimension int) (*RedisVectorStore, error) {
	// Set default values
	if config.IndexName == "" {
		config.IndexName = "nova_rag_index"
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.DB,
		Protocol: 2,
	})

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	store := &RedisVectorStore{
		client:    client,
		ctx:       ctx,
		config:    config,
		dimension: dimension,
	}

	// Create index if it doesn't exist
	if err := store.ensureIndexExists(); err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return store, nil
}

// Close closes the Redis connection
func (rvs *RedisVectorStore) Close() error {
	return rvs.client.Close()
}

// ensureIndexExists creates the Redis search index if it doesn't already exist
func (rvs *RedisVectorStore) ensureIndexExists() error {
	// Check if index exists
	_, err := rvs.client.Do(rvs.ctx, "FT.INFO", rvs.config.IndexName).Result()
	if err == nil {
		// Index already exists
		return nil
	}

	// Create index with HNSW vector field
	// FT.CREATE index_name ON HASH PREFIX 1 prefix: SCHEMA field_name type [options...]
	args := []interface{}{
		"FT.CREATE",
		rvs.config.IndexName,
		"ON", "HASH",
		"PREFIX", "1", "doc:",
		"SCHEMA",
		"prompt", "TEXT",
		"embedding", "VECTOR", "HNSW", "6",
		"TYPE", "FLOAT32",
		"DIM", strconv.Itoa(rvs.dimension),
		"DISTANCE_METRIC", "COSINE",
	}

	_, err = rvs.client.Do(rvs.ctx, args...).Result()
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// Save saves a vector record to Redis
// If the record doesn't have an ID, a new UUID will be generated
func (rvs *RedisVectorStore) Save(vectorRecord VectorRecord) (VectorRecord, error) {
	// Generate ID if not provided
	if vectorRecord.Id == "" {
		vectorRecord.Id = uuid.New().String()
	}

	// Convert embedding to bytes
	embeddingBytes := floatsToBytes(vectorRecord.Embedding)

	// Store in Redis as a hash
	key := fmt.Sprintf("doc:%s", vectorRecord.Id)
	pipe := rvs.client.Pipeline()
	pipe.HSet(rvs.ctx, key, "id", vectorRecord.Id)
	pipe.HSet(rvs.ctx, key, "prompt", vectorRecord.Prompt)
	pipe.HSet(rvs.ctx, key, "embedding", embeddingBytes)

	_, err := pipe.Exec(rvs.ctx)
	if err != nil {
		return VectorRecord{}, fmt.Errorf("failed to save vector record: %w", err)
	}

	return vectorRecord, nil
}

// GetAll retrieves all vector records from Redis
func (rvs *RedisVectorStore) GetAll() ([]VectorRecord, error) {
	// Find all keys matching the doc: prefix
	keys, err := rvs.client.Keys(rvs.ctx, "doc:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	records := make([]VectorRecord, 0, len(keys))
	for _, key := range keys {
		// Get hash data
		data, err := rvs.client.HGetAll(rvs.ctx, key).Result()
		if err != nil {
			continue // Skip failed records
		}

		// Parse record
		record := VectorRecord{
			Id:     data["id"],
			Prompt: data["prompt"],
		}

		// Convert bytes back to floats
		if embeddingBytes, ok := data["embedding"]; ok {
			record.Embedding = bytesToFloats([]byte(embeddingBytes))
		}

		records = append(records, record)
	}

	return records, nil
}

// SearchSimilarities searches for vector records with cosine similarity >= limit
// Parameters:
//   - embeddingFromQuestion: the vector record to search for (only Embedding field is used)
//   - limit: minimum cosine similarity threshold (0.0 to 1.0, where 1.0 is identical)
//
// Returns records sorted by similarity (highest first)
func (rvs *RedisVectorStore) SearchSimilarities(embeddingFromQuestion VectorRecord, limit float64) ([]VectorRecord, error) {
	// Convert limit (cosine similarity) to distance for Redis
	// Redis returns distance, we need to convert back to similarity
	// For COSINE metric in Redis: distance = 1 - cosine_similarity
	// So we search with a large K and filter afterwards

	queryVector := floatsToBytes(embeddingFromQuestion.Embedding)

	// Build FT.SEARCH query for vector similarity
	// We use KNN (K-Nearest Neighbors) search
	args := []interface{}{
		"FT.SEARCH",
		rvs.config.IndexName,
		"*=>[KNN 100 @embedding $query_vec AS score]",
		"PARAMS", "2", "query_vec", queryVector,
		"SORTBY", "score",
		"DIALECT", "2",
		"RETURN", "3", "id", "prompt", "score",
	}

	result, err := rvs.client.Do(rvs.ctx, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	records, err := rvs.parseSearchResults(result, limit)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// SearchTopNSimilarities searches for the top N most similar records
// Parameters:
//   - embeddingFromQuestion: the vector record to search for
//   - limit: minimum cosine similarity threshold
//   - max: maximum number of results to return
func (rvs *RedisVectorStore) SearchTopNSimilarities(embeddingFromQuestion VectorRecord, limit float64, max int) ([]VectorRecord, error) {
	records, err := rvs.SearchSimilarities(embeddingFromQuestion, limit)
	if err != nil {
		return nil, err
	}

	return getTopNVectorRecords(records, max), nil
}

// parseScoreField converts a raw Redis "score" field value to cosine similarity.
// Redis COSINE distance = 1 - cosine_similarity, so similarity = 1.0 - distance.
func parseScoreField(raw interface{}) (float64, bool) {
	switch v := raw.(type) {
	case string:
		distance, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, false
		}
		return 1.0 - distance, true
	case float64:
		return 1.0 - v, true
	}
	return 0, false
}

// parseDocumentFields parses the flat field-value pairs returned by FT.SEARCH
// for a single document and returns the populated VectorRecord.
func parseDocumentFields(id string, fields []interface{}) VectorRecord {
	record := VectorRecord{Id: id}
	for j := 0; j < len(fields); j += 2 {
		if j+1 >= len(fields) {
			break
		}
		fieldName, ok := fields[j].(string)
		if !ok {
			continue
		}
		switch fieldName {
		case "prompt":
			if prompt, ok := fields[j+1].(string); ok {
				record.Prompt = prompt
			}
		case "score":
			if similarity, ok := parseScoreField(fields[j+1]); ok {
				record.CosineSimilarity = similarity
			}
		}
	}
	return record
}

// parseSearchResults parses the FT.SEARCH result and converts it to VectorRecord slice
func (rvs *RedisVectorStore) parseSearchResults(result interface{}, similarityLimit float64) ([]VectorRecord, error) {
	resultArray, ok := result.([]interface{})
	if !ok || len(resultArray) == 0 {
		return []VectorRecord{}, nil
	}

	count, ok := resultArray[0].(int64)
	if !ok || count == 0 {
		return []VectorRecord{}, nil
	}

	records := make([]VectorRecord, 0)

	// Results come as pairs: [key, fields, key, fields, ...]  (first element is count)
	for i := 1; i < len(resultArray); i += 2 {
		if i+1 >= len(resultArray) {
			break
		}
		docKey, ok := resultArray[i].(string)
		if !ok {
			continue
		}
		fields, ok := resultArray[i+1].([]interface{})
		if !ok {
			continue
		}
		record := parseDocumentFields(strings.TrimPrefix(docKey, "doc:"), fields)
		if record.CosineSimilarity >= similarityLimit {
			records = append(records, record)
		}
	}

	return records, nil
}

// floatsToBytes converts a float64 slice to bytes for Redis storage (FLOAT32 encoding)
func floatsToBytes(floats []float64) []byte {
	bytes := make([]byte, len(floats)*4)
	for i, f := range floats {
		binary.LittleEndian.PutUint32(bytes[i*4:(i+1)*4], math.Float32bits(float32(f)))
	}
	return bytes
}

// bytesToFloats converts bytes back to float64 slice (FLOAT32 encoding)
func bytesToFloats(bytes []byte) []float64 {
	floats := make([]float64, len(bytes)/4)
	for i := range floats {
		floats[i] = float64(math.Float32frombits(binary.LittleEndian.Uint32(bytes[i*4 : (i+1)*4])))
	}
	return floats
}
