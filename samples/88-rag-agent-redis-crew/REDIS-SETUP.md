# Redis Setup for NOVA RAG Agent

This guide explains how to set up Redis as a vector store backend for the NOVA RAG Agent.

## Quick Start

### Start Redis

```bash
# Start Redis Stack (includes RediSearch for vector search)
docker-compose -f docker-compose.redis.yml up -d

# Check status
docker-compose -f docker-compose.redis.yml ps

# View logs
docker-compose -f docker-compose.redis.yml logs -f redis-server
```

### Stop Redis

```bash
# Stop without removing data
docker-compose -f docker-compose.redis.yml stop

# Stop and remove containers (data persists in volume)
docker-compose -f docker-compose.redis.yml down

# Stop and remove EVERYTHING including data (⚠️ WARNING: destroys all stored vectors)
docker-compose -f docker-compose.redis.yml down -v
```

## What's Included

The `docker-compose.redis.yml` file provides:

- **Redis 8.2.3-alpine3.22**: Lightweight Alpine-based Redis with **built-in RediSearch** for vector similarity search
- **Port 6379**: Redis server (standard port)
- **Persistent storage**: Data survives container restarts (mapped to `./data` directory)
- **Health checks**: Automatic monitoring of Redis availability
- **Auto-restart**: Container restarts automatically if it crashes

**Important**: Redis 8.x includes RediSearch natively! No need for redis-stack or separate modules.

## Why Redis 8.2.3?

Starting with **Redis 8.0**, RediSearch is **included natively** in the core Redis distribution! This means:

✅ **No separate modules needed** - vector search works out-of-the-box
✅ **Lightweight** - Alpine image is only ~50MB
✅ **Same as VectorMind** - battle-tested configuration from RecallFlow
✅ **HNSW indexing** - efficient vector similarity search built-in
✅ **Production-ready** - stable and officially supported

**Sources:**
- [Redis 8 GA Release](https://redis.io/blog/redis-8-ga/) - RediSearch now native
- [VectorMind Project](https://github.com/RecallFlow/VectorMind) - Uses same configuration

## Redis CLI Access

You can interact with Redis directly using the CLI:

```bash
# Access Redis CLI
docker exec -it nova-redis-vector-store redis-cli

# Inside Redis CLI, try these commands:
PING                    # Test connection
FT._LIST                # List all search indexes
FT.INFO nova_rag_index  # View index details
KEYS doc:*              # List all documents
HGETALL doc:some-uuid   # View a specific document
```

**Optional: RedisInsight UI**

If you want a web interface, you can install RedisInsight separately:

```bash
docker run -d --name redisinsight \
  -p 8001:5540 \
  --network nova_nova-network \
  redis/redisinsight:latest
```

Then access it at http://localhost:8001 and connect to `redis-server:6379`

## Testing the Connection

```bash
# Test Redis is running
docker exec -it nova-redis-vector-store redis-cli ping
# Should return: PONG

# Check RediSearch module is loaded
docker exec -it nova-redis-vector-store redis-cli MODULE LIST
# Should show: redisearch

# Check if there are any indexes
docker exec -it nova-redis-vector-store redis-cli FT._LIST
```

## Data Persistence

Redis data is stored in a Docker volume named `redis-data`. This means:

✅ Data survives container restarts
✅ Data survives `docker-compose down`
❌ Data is deleted with `docker-compose down -v`

### Backup and Restore

```bash
# Backup: Trigger Redis save
docker exec nova-redis-vector-store redis-cli BGSAVE

# Export volume data
docker run --rm -v nova_redis-data:/data -v $(pwd):/backup alpine \
  tar czf /backup/redis-backup-$(date +%Y%m%d-%H%M%S).tar.gz -C /data .

# Restore from backup (⚠️ stops Redis first)
docker-compose -f docker-compose.redis.yml down
docker run --rm -v nova_redis-data:/data -v $(pwd):/backup alpine \
  sh -c "cd /data && tar xzf /backup/redis-backup-YYYYMMDD-HHMMSS.tar.gz"
docker-compose -f docker-compose.redis.yml up -d
```

## Troubleshooting

### Redis won't start

```bash
# Check if port 6379 is already in use
lsof -i :6379

# View detailed logs
docker-compose -f docker-compose.redis.yml logs redis-server

# Reset everything (⚠️ deletes data)
docker-compose -f docker-compose.redis.yml down -v
docker-compose -f docker-compose.redis.yml up -d
```

### Connection refused from Go application

- Ensure Redis is running: `docker-compose -f docker-compose.redis.yml ps`
- Check connection string in your code: `localhost:6379` (or `redis-server:6379` if your app runs in Docker)
- Verify network connectivity: `docker network ls`

### RedisInsight can't connect

- Check port 8001 is not in use: `lsof -i :8001`
- Access RedisInsight at: http://localhost:8001
- Use connection: Host=`localhost`, Port=`6379`, no password

## Production Considerations

For production deployments:

1. **Set a password**: Add `REDIS_ARGS=--requirepass yourpassword` to environment
2. **Use specific version**: Change `redis/redis-stack:latest` to `redis/redis-stack:7.4.0-v0` for stability
3. **Tune persistence**: Adjust `--save` and `--appendonly` based on your durability requirements
4. **Monitor memory**: Set `maxmemory` and `maxmemory-policy` for large datasets
5. **Use Redis Sentinel or Cluster** for high availability

Example production configuration:

```yaml
environment:
  - REDIS_ARGS=--requirepass ${REDIS_PASSWORD} --maxmemory 2gb --maxmemory-policy allkeys-lru --save 60 1000 --appendonly yes
```

## Next Steps

See the examples in `samples/` directory:
- `samples/XX-rag-agent-redis-simple/` - Basic Redis RAG usage
- `samples/XX-rag-agent-redis-crew/` - Redis RAG with Crew Agent

Read the documentation:
- `docs/rag-agent-redis.en.md` - English documentation
- `docs/rag-agent-redis.fr.md` - Documentation en français
