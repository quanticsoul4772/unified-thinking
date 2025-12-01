# Knowledge Graph Setup and Testing

## Prerequisites

### 1. Start Neo4j

**Option A: Using Docker Compose (Recommended)**
```bash
# From the unified-thinking directory
docker-compose up -d

# Wait for Neo4j to be ready (takes ~30 seconds)
docker-compose logs -f neo4j

# You should see "Started" in the logs
```

**Option B: Using Docker directly**
```bash
docker run -d \
  --name unified-thinking-neo4j \
  -p 7474:7474 -p 7687:7687 \
  -e NEO4J_AUTH=neo4j/password \
  neo4j:latest
```

**Option C: Native Installation**
- Download from https://neo4j.com/download/
- Set password to `password` (or update tests)
- Ensure it's running on default ports (7474 HTTP, 7687 Bolt)

### 2. Verify Neo4j is Running

Open browser to http://localhost:7474
- Username: `neo4j`
- Password: `password`

Or test with cypher-shell:
```bash
docker exec -it unified-thinking-neo4j cypher-shell -u neo4j -p password
# Run: RETURN 1;
# Should return: 1
```

## Running Knowledge Graph Tests

### Full Integration Tests (Requires Neo4j)
```bash
# Run all knowledge graph tests with Neo4j
go test ./internal/knowledge/... -v

# Expected output:
# - TestKnowledgeGraph_StoreAndRetrieve: Tests entity storage and retrieval
# - TestKnowledgeGraph_SemanticSearch: Tests vector similarity search
# - TestKnowledgeGraph_HybridSearch: Tests combined semantic + graph traversal
# - TestTrajectoryExtractor_ExtractFromTrajectory: Tests episodic memory integration
# - TestRLContextRetriever_GetSimilarProblems: Tests RL context retrieval
# - TestRLContextRetriever_GetStrategyPerformance: Tests strategy performance tracking
```

### Coverage Report
```bash
# Generate coverage report
go test ./internal/knowledge/... -cover -coverprofile=coverage.out

# View coverage in browser
go tool cover -html=coverage.out
```

### Short Mode (Skips Neo4j Tests)
```bash
# Only runs unit tests, skips integration tests
go test ./internal/knowledge/... -short
```

## Expected Test Coverage

**With Neo4j running:**
- `internal/knowledge`: ~85%+ coverage
- `internal/knowledge/extraction`: 71.7% coverage

**Without Neo4j (short mode):**
- `internal/knowledge`: ~5% coverage (only unit tests)
- `internal/knowledge/extraction`: 71.7% coverage

## Troubleshooting

### "Neo4j not available" errors
1. Verify Neo4j is running: `docker ps | grep neo4j`
2. Check connection: `nc -zv localhost 7687`
3. Check logs: `docker logs unified-thinking-neo4j`

### "database is locked" errors
- This happens if multiple test runs overlap
- Solution: Wait for previous tests to complete, or restart Neo4j

### Performance Issues
- Increase Neo4j memory in docker-compose.yml:
  ```yaml
  - NEO4J_dbms_memory_heap_max__size=2G
  ```

## Cleanup

```bash
# Stop and remove Neo4j container
docker-compose down

# Remove data volumes (fresh start)
docker-compose down -v
```

## Environment Variables for Production

When deploying, set these environment variables:

```bash
NEO4J_ENABLED=true
NEO4J_URI=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=your-secure-password
NEO4J_DATABASE=neo4j
NEO4J_TIMEOUT_MS=5000

# Voyage AI for embeddings
VOYAGE_API_KEY=your-api-key
EMBEDDINGS_MODEL=voyage-3-lite

# SQLite for embedding cache
STORAGE_TYPE=sqlite
SQLITE_PATH=/path/to/unified-thinking.db
```

## What the Knowledge Graph Provides

### With Neo4j Enabled
1. **Hybrid Search**: Semantic similarity + graph traversal
2. **Entity Storage**: Problems, concepts, strategies, tools, files
3. **Relationship Tracking**: Causal links, enables/contradicts patterns
4. **RL Context**: Similar past problems inform strategy selection
5. **Episodic Integration**: Reasoning trajectories populate knowledge graph

### Without Neo4j
- Feature gracefully degrades (returns empty results)
- All other unified-thinking features work normally
- Thompson Sampling RL works without knowledge graph context
