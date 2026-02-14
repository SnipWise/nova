# RAG Agent with Redis - Skip Duplicates Demo

This example demonstrates the `DocumentLoadModeSkipDuplicates` feature, which prevents duplicate documents from being added to your vector store when restarting your application.

## Problem Solved

When using persistent stores like Redis, restarting your application would typically reload all documents, creating duplicates. With `DocumentLoadModeSkipDuplicates`, the RAG agent:

- ✅ Checks each document individually before adding it
- ✅ Skips documents that already exist (based on exact Prompt match)
- ✅ Only adds NEW documents
- ✅ Logs statistics: "X added, Y skipped (duplicates)"

## Prerequisites

1. **Redis** must be running on `localhost:6379`
   ```bash
   # Start Redis with Docker
   docker run -d --name redis-stack -p 6379:6379 redis/redis-stack-server:latest

   # Or with redis-stack (includes RedisInsight)
   docker run -d --name redis-stack -p 6379:6379 -p 8001:8001 redis/redis-stack:latest
   ```

2. **Ollama** with embedding model
   ```bash
   ollama pull mxbai-embed-large
   ```

3. **Environment variables** (create `.env` file)
   ```bash
   # Optional: Enable debug logs to see skip messages
   LOG_LEVEL=debug
   ```

## Usage

```bash
# First run - loads all documents
go run main.go

# Second run - skips all documents (they already exist)
go run main.go

# Third run - still no duplicates!
go run main.go
```

## What Happens

### First Run
```
📝 Documents to load:
   1. Squirrels run in the forest...
   2. Birds fly in the sky...
   3. Frogs swim in the pond...
   4. Bears hibernate in caves...
   5. Rabbits hop through meadows...

✅ RAG Agent created successfully
   Redis: localhost:6379
   Index: skip_duplicates_demo
   Mode: DocumentLoadModeSkipDuplicates

[DEBUG] Loading 5 documents into store
[DEBUG] Successfully saved embedding for document 0
[DEBUG] Successfully saved embedding for document 1
[DEBUG] Successfully saved embedding for document 2
[DEBUG] Successfully saved embedding for document 3
[DEBUG] Successfully saved embedding for document 4
[DEBUG] Document loading complete: 5 added, 0 skipped (duplicates)
```

### Second Run (and subsequent runs)
```
[DEBUG] Loading 5 documents into store
[DEBUG] Checking documents individually for duplicates (DocumentLoadModeSkipDuplicates)
[DEBUG] Document 0 already exists, skipping (duplicate)
[DEBUG] Document 1 already exists, skipping (duplicate)
[DEBUG] Document 2 already exists, skipping (duplicate)
[DEBUG] Document 3 already exists, skipping (duplicate)
[DEBUG] Document 4 already exists, skipping (duplicate)
[DEBUG] Document loading complete: 0 added, 5 skipped (duplicates)
```

## Code Explanation

The key configuration:

```go
ragAgent, err := rag.NewAgent(
    ctx,
    agents.Config{
        Name:      "SkipDuplicatesDemo",
        EngineURL: engineURL,
    },
    models.Config{
        Name: embeddingModel,
    },
    rag.WithRedisStore(stores.RedisConfig{
        Address:   "localhost:6379",
        Password:  "",
        DB:        0,
        IndexName: "skip_duplicates_demo",
    }, 1024),
    rag.WithDocuments(documents, rag.DocumentLoadModeSkipDuplicates), // 👈 The magic!
)
```

## Available Document Load Modes

- `DocumentLoadModeOverwrite` - Clears existing data before loading
- `DocumentLoadModeMerge` - Adds all documents to existing data (default, creates duplicates)
- `DocumentLoadModeSkip` - Does nothing if store already contains ANY data
- `DocumentLoadModeSkipDuplicates` - ✨ **Checks each document individually** and skips only duplicates
- `DocumentLoadModeError` - Returns error if store is not empty

## Comparison: Skip vs SkipDuplicates

### `DocumentLoadModeSkip`
- ❌ All-or-nothing: skips ALL documents if store is not empty
- ❌ Can't add new documents to an existing store
- ✅ Very fast (one check)

### `DocumentLoadModeSkipDuplicates`
- ✅ Granular: checks each document individually
- ✅ Can add new documents to an existing store
- ✅ Perfect for application restarts
- ⚠️ Slower on large stores (checks each document)

## Use Cases

Perfect for:
- 📱 Applications that restart frequently
- 🔄 Development/testing cycles
- 🚀 Production apps with persistent Redis
- 📊 Incremental data loading

## Clean Up

To reset the Redis index and start fresh:

```bash
# Connect to Redis CLI
docker exec -it redis-stack redis-cli

# Delete the index
> DEL skip_duplicates_demo:*
> KEYS skip_duplicates_demo:*  # Should return empty

# Or delete all keys (use with caution!)
> FLUSHALL
```

## Next Steps

- Try modifying the documents list and rerunning - new documents will be added, existing ones skipped!
- Experiment with different similarity thresholds in the search
- Check out the Redis data using RedisInsight at http://localhost:8001
