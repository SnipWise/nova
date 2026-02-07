# Sample 88: Multiple Agents with Shared Redis Knowledge Base

This example demonstrates how multiple AI agents can share a common Redis-backed knowledge base for collaborative tasks.

## What This Example Shows

1. **Shared Redis knowledge base** - Multiple agents accessing the same persistent data
2. **Specialized agents** - Researcher and Writer agents with different roles
3. **RAG-powered responses** - Agents use semantic search to find relevant context
4. **Persistent collaboration** - Knowledge survives restarts and can be reused

## Use Case

Imagine a company information system where:
- A **Researcher agent** finds and analyzes factual information
- A **Writer agent** creates marketing content
- Both agents share the same knowledge base stored in Redis
- Information is persistent and can be updated centrally

## Prerequisites

### 1. Start Redis

```bash
# From project root
./start-redis.sh

# Or manually
docker-compose -f docker-compose.redis.yml up -d
```

### 2. Start Language Model

You need a chat model for the agents:

```bash
# Example with llama.cpp
docker run -d -p 12434:12434 --name llama-cpp-server your-image

# Or use Ollama
ollama run qwen2.5:3b-instruct
```

### 3. Start Embedding Model

You need an embedding model (1024 dimensions):

```bash
ollama run mxbai-embed-large
```

## Running the Example

```bash
cd samples/88-rag-agent-redis-crew
go run main.go
```

## What Happens

### Step 1-2: Knowledge Loading
- Creates a RAG agent with Redis backend
- Loads 8 pieces of company information into Redis
- All data is stored in the `crew_knowledge_base` index

### Step 3: Agent Creation
- **Researcher**: Analyzes information, provides factual summaries (temperature=0.0)
- **Writer**: Creates engaging marketing content (temperature=0.5)

### Step 4: Shared Knowledge
- Both agents access the same Redis knowledge base
- Ensures consistency across the team

### Step 5: Research Task
- Query: "company founders and locations"
- RAG finds relevant information from Redis
- Researcher synthesizes a factual answer

### Step 6: Writing Task
- Query: "product and services"
- RAG finds relevant product information
- Writer creates marketing description

## Key Benefits Demonstrated

✅ **Shared Knowledge** - All agents access the same Redis data
✅ **Persistence** - Knowledge survives application restarts
✅ **Consistency** - Everyone works with the same information
✅ **Scalability** - Add more agents without duplicating data
✅ **Fast Search** - HNSW indexing for quick semantic search

## Comparison: In-Memory vs Redis

| Feature | In-Memory (per agent) | Redis (shared) |
|---------|---------------------|----------------|
| **Knowledge sharing** | Each agent has its own copy | All agents share one source |
| **Memory usage** | N × data size | 1 × data size |
| **Persistence** | Lost on restart | Survives restarts |
| **Updates** | Must update each agent | Update once, all agents see it |
| **Scaling** | Linear growth | Constant size |

## Example Output

```
═══════════════════════════════════════
NOVA Crew with Shared Redis Knowledge Base
═══════════════════════════════════════

═══════════════════════════════════════
1. Creating Shared RAG Agent with Redis
═══════════════════════════════════════
✅ Shared RAG Agent created
   Redis: localhost:6379
   Index: crew_knowledge_base
   All crew members will share this knowledge base

═══════════════════════════════════════
2. Loading Knowledge into Redis
═══════════════════════════════════════
📝 Loading 8 knowledge items into Redis...
   ✓ Saved: "The company was founded in 2020..."
   ...
✅ Knowledge base loaded into Redis

═══════════════════════════════════════
5. Task 1: Research Company Information
═══════════════════════════════════════
🔍 Searching knowledge base for: "company founders and locations"
   Found 3 relevant pieces of information:
   1. The company was founded in 2020 by Alice Johnson and Bob Smith
      Similarity: 0.7234
   2. We have offices in San Francisco, London, and Tokyo
      Similarity: 0.6891
   ...

📊 Researcher agent analyzing information...

🤖 Researcher: The company was founded by Alice Johnson and Bob Smith
   in 2020. The company has offices in San Francisco, London, and Tokyo.

═══════════════════════════════════════
6. Task 2: Product Information
═══════════════════════════════════════
🔍 Searching knowledge base for: "product and services"
   ...

✍️  Writer agent creating description...

🤖 Writer: TaskMaster is a powerful cloud-based project management tool
   designed to simplify team collaboration. Now available on both iOS
   and Android, it brings your team's productivity to new heights.
```

## Extending the Example

### Add More Agents

```go
// Create a fact-checker agent
factCheckerAgent, err := chat.NewAgent(ctx,
    agents.Config{
        Name: "FactChecker",
        SystemInstructions: "You verify information accuracy.",
        ...
    },
    ...
)

// All agents use the same ragAgent for knowledge retrieval
results, _ := ragAgent.SearchSimilar("...", 0.3)
```

### Update Knowledge Dynamically

```go
// Add new information to Redis
newFact := "The company opened a new office in Berlin in 2024"
ragAgent.SaveEmbedding(newFact)

// All agents immediately have access to this new information
```

### Multiple Knowledge Bases

```go
// Create separate knowledge bases for different topics
technicalRAG, _ := rag.NewAgent(ctx, ..., rag.WithRedisStore(
    stores.RedisConfig{IndexName: "technical_docs"}, 1024))

marketingRAG, _ := rag.NewAgent(ctx, ..., rag.WithRedisStore(
    stores.RedisConfig{IndexName: "marketing_content"}, 1024))
```

## Inspecting Redis Data

```bash
# Access Redis CLI
docker exec -it nova-redis-vector-store redis-cli

# View the index
FT.INFO crew_knowledge_base

# List all documents
KEYS doc:*

# Count documents
DBSIZE

# View specific document
HGETALL doc:<some-uuid>
```

## Cleanup

```bash
# Stop without deleting data
docker-compose -f ../../docker-compose.redis.yml stop

# Delete all data and restart fresh
docker-compose -f ../../docker-compose.redis.yml down -v
docker-compose -f ../../docker-compose.redis.yml up -d
```

## Real-World Applications

This pattern is useful for:
- **Customer support teams** - Shared knowledge base across multiple support agents
- **Content creation** - Research, writing, and editing agents collaborating
- **Data analysis** - Multiple analysts querying the same dataset
- **Documentation systems** - Technical writers accessing shared technical knowledge
- **Multi-agent workflows** - Complex tasks requiring specialized agents

## Next Steps

- See `samples/87-rag-agent-redis-simple/` for basic Redis RAG usage
- Read `docs/rag-agent-redis.en.md` for complete documentation
- Check `REDIS-SETUP.md` for Redis configuration details

## Learn More

- [Redis Vector Search](https://redis.io/docs/stack/search/reference/vectors/)
- [Multi-Agent Systems](https://en.wikipedia.org/wiki/Multi-agent_system)
- [RAG (Retrieval-Augmented Generation)](https://arxiv.org/abs/2005.11401)
