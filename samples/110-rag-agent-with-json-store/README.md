# Sample 110: RAG Agent with JSON Store

This example demonstrates how to use the RAG Agent with JSON file-based storage and document initialization.

## Features Demonstrated

### 1. **WithJsonStore** - JSON File-Based Storage
- Automatically loads existing embeddings from a JSON file
- Falls back to an empty store if the file doesn't exist
- Allows persistence via `agent.PersistStore(filePath)`

### 2. **WithDocuments** - Document Initialization
- Initialize the RAG agent with predefined documents
- Automatically generates embeddings for each document
- Supports multiple loading modes:
  - `DocumentLoadModeOverwrite`: Clear existing data and load new documents
  - `DocumentLoadModeMerge`: Add new documents to existing data (default)
  - `DocumentLoadModeSkip`: Skip loading if store already has data
  - `DocumentLoadModeError`: Error if store already contains data

## Usage

```bash
go run main.go
```

### First Run
On the first run, the program will:
1. Create an empty JSON store
2. Load the predefined documents (animals)
3. Generate embeddings for each document
4. Save the store to `./store/animals.json`
5. Perform similarity searches

### Subsequent Runs
On subsequent runs, the program will:
1. Load the existing store from `./store/animals.json`
2. Overwrite with new documents (due to `DocumentLoadModeOverwrite`)
3. Save the updated store
4. Perform similarity searches

## Try Different Modes

### Merge Mode (Add Documents)
```go
rag.WithDocuments(txtChunks, rag.DocumentLoadModeMerge)
```
This will add new documents to existing ones.

### Skip Mode (Keep Existing)
```go
rag.WithDocuments(txtChunks, rag.DocumentLoadModeSkip)
```
This will skip loading if the store already has data.

### Error Mode (Prevent Overwrites)
```go
rag.WithDocuments(txtChunks, rag.DocumentLoadModeError)
```
This will log an error if the store is not empty.

## Configuration

- **Engine URL**: `http://localhost:12434/engines/llama.cpp/v1`
- **Model**: `ai/mxbai-embed-large:latest`
- **Store Path**: `./store/animals.json`
- **Similarity Threshold**: `0.6`

## Key Takeaways

1. **Order Matters**: Always apply `WithJsonStore` before `WithDocuments`
2. **Persistence**: Call `agent.PersistStore(filePath)` to save changes to disk
3. **Flexible Loading**: Choose the right loading mode for your use case
4. **Reusable Store**: The JSON file can be loaded in future runs for fast startup
