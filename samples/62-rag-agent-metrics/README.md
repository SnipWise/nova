# RAG Agent with Embedding Metrics

This example demonstrates metrics tracking for RAG (Retrieval Augmented Generation) operations with embeddings.

## Important Note

**The RAG agent uses specialized embedding telemetry** (not LLM completion telemetry) because it's designed for embeddings. This example shows:
1. **Built-in telemetry**: `GetLastEmbeddingRequestJSON()` and `GetLastEmbeddingResponseJSON()` for API monitoring
2. **Performance metrics**: `rag.Metrics` type for tracking embedding generation and search performance

## Features

- Generate embeddings for multiple documents
- Track embedding generation performance
- Monitor similarity search operations
- Calculate cost estimates for embeddings
- Measure vector store efficiency

## Metrics Tracked

### Embedding Generation
- Total embeddings created
- Average dimensions per embedding
- Time per embedding generation
- Characters processed
- Throughput (ops/second)

### Similarity Search
- Number of search operations
- Search latency per query
- Results quality (similarity scores)
- Vector store lookup performance

### Cost Estimation
- Character-based cost calculation
- Per-document cost breakdown
- Total session costs

## Why RAG Agent is Different

Unlike other agents that perform LLM completions:
- RAG agent generates **embeddings** (vector representations)
- Uses embedding API instead of completion API
- Telemetry tracks embedding requests/responses (not chat completions)
- Metrics focus on **vector operations** (dimensions, search time, etc.)

## Using rag.Metrics

The `rag.Metrics` type provides convenient methods for tracking RAG performance:

```go
// Create metrics tracker
metrics := rag.NewMetrics()

// Record embedding
metrics.RecordEmbedding(content, dimensions, duration)

// Record search
metrics.RecordSearch(duration)

// Get averages
avgDims := metrics.AvgDimensions()
avgTime := metrics.AvgEmbeddingTime()
throughput := metrics.Throughput()

// Estimate costs
cost := metrics.EstimateCost(0.0001) // $0.0001 per 1000 chars
```

## Running

```bash
cd samples/62-rag-agent-metrics
go run main.go
```

## Expected Output

The example:
1. Creates embeddings for 5 knowledge base documents
2. Tracks generation time and dimensions for each
3. Performs similarity searches for 3 queries
4. Measures search performance
5. Calculates overall efficiency and costs

## Use Cases

Embedding metrics are useful for:
1. **Cost Optimization**: Track embedding API costs
2. **Performance Tuning**: Optimize embedding model selection
3. **Quality Monitoring**: Measure search result quality
4. **Capacity Planning**: Size vector stores appropriately
5. **Latency Analysis**: Identify bottlenecks in RAG pipeline
