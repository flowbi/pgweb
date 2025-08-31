# Query and Metadata Caching

Flowbi/pgweb implements intelligent in-memory caching to improve performance by reducing database load and response times for frequently accessed data.

## Overview

The caching system consists of two independent caches:

- **Query Cache**: Caches results of SELECT queries that don't contain time-sensitive functions
- **Metadata Cache**: Caches database schema information like tables, columns, constraints, and indexes

## Cache Implementation Details

### Thread Safety

- Uses `sync.RWMutex` for concurrent read/write access
- Multiple queries can read from cache simultaneously
- Cache writes are serialized for data consistency

### Memory Management

- **Smart Memory Limits**: Query cache limited to 50MB, metadata cache to 100MB by default
- **Size-Based Eviction**: Automatically removes oldest items when memory limit is approached
- **Memory Estimation**: Uses reflection-based size calculation for cached values
- **Automatic Cleanup**: Expired entries cleaned every 5 minutes
- **TTL-based Expiration**: All cached items expire based on configured TTL

### Cache Key Strategy

- **Security**: MD5 hashing of query + connection string + user role for uniqueness
- **Role Isolation**: Different users with different roles get separate cache entries
- **Namespaced Keys**: Prefixed for organization:
  - `query:` for SELECT query results
  - `metadata:` for database schema information

## Configuration

All caching settings can be configured via environment variables or command-line flags:

### Environment Variables

```bash
# Disable query result caching (default: false)
PGWEB_DISABLE_QUERY_CACHE=false

# Disable metadata caching (default: false)
PGWEB_DISABLE_METADATA_CACHE=false

# Query cache TTL in seconds (default: 120)
PGWEB_QUERY_CACHE_TTL=300

# Metadata cache TTL in seconds (default: 600)
PGWEB_METADATA_CACHE_TTL=600
```

### Command-Line Flags

```bash
# Disable caches
./pgweb --no-query-cache --no-metadata-cache

# Configure TTL values
./pgweb --query-cache-ttl=300 --metadata-cache-ttl=1200
```

## Query Cache Behavior

### Cacheable Queries

Only SELECT queries are cached, and only when they don't contain time-sensitive functions:

**✅ Cached:**

```sql
SELECT * FROM users WHERE status = 'active'
SELECT COUNT(*) FROM orders
SELECT name, email FROM customers
```

**❌ Not Cached:**

```sql
INSERT INTO logs VALUES (...)  -- Not a SELECT
UPDATE users SET last_seen = NOW()  -- Contains time function
SELECT * FROM events WHERE created_at > NOW() - INTERVAL '1 hour'
SELECT random() as lucky_number  -- Contains random function
```

### Time-Sensitive Function Detection

The following functions prevent query caching:

- `now()`
- `current_timestamp`
- `random()`

### Cache Flow

1. Check if query is cacheable (SELECT + no time-sensitive functions)
2. Generate cache key from query + connection string
3. Look for existing cached result
4. If found and not expired, return cached result
5. If not found, execute query and cache result

## Metadata Cache Behavior

### Cached Metadata

- Database schemas list
- Table and view information
- Column definitions and types
- Table constraints (primary keys, foreign keys, checks)
- Index information
- Table statistics

### Cache Invalidation

Metadata cache entries expire based on TTL. Consider clearing cache after:

- Schema changes (CREATE/DROP TABLE)
- Column modifications (ALTER TABLE)
- Index creation/removal
- Constraint changes

## Cache Management

### Statistics Endpoint

Get cache performance metrics:

```bash
GET /api/cache/stats
```

Response:

```json
{
  "caching_enabled": {
    "query_cache": true,
    "metadata_cache": true
  },
  "cache_ttl": {
    "query_cache_ttl": 120,
    "metadata_cache_ttl": 600
  },
  "query_cache": {
    "total_items": 45,
    "expired_items": 3,
    "active_items": 42,
    "memory_used_mb": 12,
    "memory_limit_mb": 50,
    "memory_used_bytes": 12582912
  },
  "metadata_cache": {
    "total_items": 12,
    "expired_items": 0,
    "active_items": 12,
    "memory_used_mb": 3,
    "memory_limit_mb": 100,
    "memory_used_bytes": 3145728
  }
}
```

### Clear Cache

Clear all cached data:

```bash
POST /api/cache/clear
```

Response:

```json
{
  "message": "Caches cleared successfully",
  "cleared": ["query_cache", "metadata_cache"]
}
```

## Performance Benefits

### Query Cache Benefits

- **Faster Response Times**: Cached SELECT queries return instantly
- **Reduced Database Load**: Identical queries don't hit the database
- **Better Concurrency**: Multiple users can access cached results simultaneously

### Metadata Cache Benefits

- **Snappy UI**: Table lists and schema browsing respond immediately
- **Reduced Metadata Queries**: Schema information loaded once per TTL period
- **Improved User Experience**: Faster navigation between tables and schemas

## Best Practices

### When to Disable Caching

**Disable Query Cache** when:

- Working with rapidly changing data
- Real-time analytics requirements
- Memory constraints on the server

**Disable Metadata Cache** when:

- Frequently changing schema (development environments)
- Multiple users modifying database structure

### Optimal TTL Settings

**Short TTL (30-120 seconds)** for:

- Frequently changing application data
- Development environments
- Real-time dashboards

**Long TTL (300-3600 seconds)** for:

- Stable reference data
- Reports and analytics
- Production environments with stable schemas

### Memory Considerations

- **Built-in Limits**: Query cache limited to 50MB, metadata cache to 100MB
- **Automatic Management**: Cache automatically evicts oldest items when memory limit reached
- **Memory Monitoring**: Check memory usage via `/api/cache/stats` endpoint
- **ECS/Container Deployments**: Ensure container memory allocation accounts for cache limits plus application overhead
- **Row Count Limits**: Large query results (>10,000 rows) are not cached to prevent memory issues
- **Estimation**: Memory usage is estimated using reflection; actual usage may vary

## Deployment Considerations

### Container Memory Requirements

When deploying to ECS, Kubernetes, or Docker, ensure adequate memory allocation:

**Minimum Recommended Memory:**
- **Development**: 256MB (cache disabled or very low TTL)
- **Production**: 512MB-1GB depending on usage patterns
- **High Traffic**: 1GB+ for optimal performance

**Memory Breakdown:**
- Application base: ~200-300MB
- Query cache limit: 50MB (configurable)
- Metadata cache limit: 100MB (configurable)  
- Go runtime/GC: ~100-200MB
- Database connections: ~50-100MB
- **Total**: ~500-650MB minimum

**Example ECS Task Definition:**
```json
{
  "memory": 1024,
  "memoryReservation": 512,
  "environment": [
    {"name": "PGWEB_QUERY_CACHE_TTL", "value": "300"},
    {"name": "PGWEB_DISABLE_QUERY_CACHE", "value": "false"}
  ]
}
```

## Troubleshooting

### Cache Not Working

1. Verify caching is enabled via `/api/cache/stats`
2. Check query doesn't contain time-sensitive functions
3. Ensure query is a valid SELECT statement
4. **Check Environment Variables**: Ensure `PGWEB_QUERY_CACHE_TTL` is being read correctly

### High Memory Usage

1. **Check Cache Stats**: Review memory usage via `/api/cache/stats`
2. **Container Memory**: Ensure your container has enough memory (see deployment section above)
3. **Large Results**: Cache won't store results >10,000 rows, but check for many medium-sized results
4. **TTL Optimization**: Reduce TTL values to expire items sooner
5. **Manual Clear**: Use `/api/cache/clear` if needed
6. **Temporary Disable**: Set `PGWEB_DISABLE_QUERY_CACHE=true` as last resort

### Container Memory Issues

If you see OOMKilled errors or high memory usage:

1. **Check Actual Usage**: Memory stats may show higher usage due to Go's garbage collection
2. **Increase Container Memory**: Add 200-300MB buffer above cache limits
3. **Reduce Cache TTL**: Lower values mean faster cleanup
4. **Monitor Metrics**: Use `/api/cache/stats` to track memory trends

### Environment Variables Not Working

If your `PGWEB_QUERY_CACHE_TTL` setting isn't being applied:

1. **Verify Environment**: Check environment variables are properly set in container
2. **Restart Required**: Environment variables only read at startup
3. **Check Logs**: Look for debug messages about cache configuration
4. **Test Override**: Try command-line flags: `--query-cache-ttl=300`

### Stale Data Issues

1. Reduce cache TTL values
2. Clear cache after schema changes
3. Monitor data freshness requirements vs. performance gains
