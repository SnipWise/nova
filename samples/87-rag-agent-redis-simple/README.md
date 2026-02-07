# Sample 87: RAG Agent with Redis Vector Store

This example demonstrates how to use the NOVA RAG Agent with Redis as a persistent vector storage backend.

## What This Example Shows

1. **Creating a RAG agent with Redis store** - Using `WithRedisStore` option
2. **Storing vector embeddings in Redis** - Persistent storage using HNSW indexing
3. **Searching similar content** - Finding semantically similar text using cosine similarity
4. **Redis persistence** - Data survives application restarts

## Prerequisites

### 1. Start Redis

Make sure Redis is running with vector search support:

```bash
# From the project root directory
./start-redis.sh

# Or manually:
docker-compose -f docker-compose.redis.yml up -d
```

Verify Redis is running:
```bash
docker exec -it nova-redis-vector-store redis-cli ping
# Should return: PONG
```

### 2. Start the Embedding Model

You need an embedding model running. For this example, we use `ai/mxbai-embed-large` which produces 1024-dimensional vectors:

```bash
# If using llama.cpp via docker:
docker run -d \
  -p 12434:12434 \
  --name llama-cpp-server \
  your-llama-cpp-image

# Or use Ollama:
ollama run mxbai-embed-large
```

## Running the Example

```bash
cd samples/79-rag-agent-redis-simple
go run main.go
```

## What Happens

1. **Creates a RAG agent** connected to Redis at `localhost:6379`
2. **Generates embeddings** for 10 sample texts about animals
3. **Stores vectors in Redis** using the index `nova_rag_demo_index`
4. **Performs similarity searches** for three different queries
5. **Shows results** ranked by semantic similarity

### Sample Output

```
═══════════════════════════════════════════════════
NOVA RAG Agent with Redis Vector Store
═══════════════════════════════════════════════════

═══════════════════════════════════════════════════
1. Creating RAG Agent with Redis Store
═══════════════════════════════════════════════════
✅ RAG Agent created with Redis vector store
   Redis: localhost:6379
   Index: nova_rag_demo_index
   Model: ai/mxbai-embed-large (1024 dimensions)

═══════════════════════════════════════════════════
2. Storing Vector Embeddings in Redis
═══════════════════════════════════════════════════
📝 Storing 10 text chunks as embeddings...
   ✓ Saved: "Squirrels run in the forest"
   ✓ Saved: "Birds fly in the sky"
   ...
✅ All embeddings saved to Redis

═══════════════════════════════════════════════════
3. Searching Similar Content
═══════════════════════════════════════════════════
🔍 Query: "Which animals swim?"
   Searching with similarity threshold >= 0.3...
   Found 4 matches:
   1. "Fishes swim in the sea"
      Similarity: 0.7234
   2. "Dolphins leap out of the ocean"
      Similarity: 0.6891
   ...
```

## Key Differences from In-Memory Store

| Feature | In-Memory | Redis |
|---------|-----------|-------|
| **Persistence** | Lost on restart | Survives restarts |
| **Sharing** | Single process | Multiple processes |
| **Scale** | Limited by RAM | Millions of vectors |
| **Speed** | Fastest | Very fast (HNSW) |
| **Setup** | None needed | Requires Redis |

## Inspecting Data in Redis

You can view the stored vectors using Redis CLI:

```bash
# Access Redis CLI
docker exec -it nova-redis-vector-store redis-cli

# List all indexes
FT._LIST

# View index details
FT.INFO nova_rag_demo_index

# List all document keys
KEYS doc:*

# View a specific document
HGETALL doc:<uuid>

# Count documents
DBSIZE
```

## Running Multiple Times

The beauty of Redis persistence:

```bash
# First run - stores vectors in Redis
go run main.go

# Stop the program (Ctrl+C or let it finish)

# Second run - vectors are still in Redis!
go run main.go
# Will add more vectors to the same index
```

**To start fresh:**
```bash
# Delete all data and restart
docker-compose -f ../../docker-compose.redis.yml down -v
docker-compose -f ../../docker-compose.redis.yml up -d
```

## Configuration Options

You can customize the Redis connection in `main.go`:

```go
rag.WithRedisStore(stores.RedisConfig{
    Address:   "localhost:6379",  // Redis server address
    Password:  "",                // Redis password (if required)
    DB:        0,                 // Redis database number
    IndexName: "my_custom_index", // Custom index name
}, 1024) // Embedding dimension (must match model!)
```

**Important**: The `dimension` parameter (1024 in this example) **must match** your embedding model's output dimension:
- `ai/mxbai-embed-large`: 1024
- `text-embedding-3-small`: 1536
- `text-embedding-3-large`: 3072
- Check your model's documentation

## Troubleshooting

### Redis Connection Failed

```
❌ Failed to create RAG agent: failed to connect to Redis: dial tcp [::1]:6379: connect: connection refused
```

**Solution**: Start Redis with `./start-redis.sh` from the project root

### Dimension Mismatch

```
Error: vector dimension mismatch
```

**Solution**: Ensure the dimension parameter matches your embedding model

### Index Already Exists

Redis reuses existing indexes. To create a fresh index:

```bash
docker exec -it nova-redis-vector-store redis-cli
FT.DROPINDEX nova_rag_demo_index DD  # DD deletes documents too
```

## Next Steps

- See `samples/80-rag-agent-redis-crew/` for using Redis RAG with Crew Agent
- Read `docs/rag-agent-redis.en.md` for complete documentation
- Check `REDIS-SETUP.md` for Redis configuration details

## Learn More

- [Redis Vector Search](https://redis.io/docs/stack/search/reference/vectors/)
- [HNSW Algorithm](https://arxiv.org/abs/1603.09320)
- [NOVA RAG Agent Documentation](../../docs/)
