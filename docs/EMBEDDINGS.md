# Semantic Embeddings for Episodic Memory

## Overview

The unified-thinking server uses semantic embeddings for episodic memory retrieval. This uses Voyage AI to generate vector representations of problems, enabling semantic similarity search combined with hash-based matching via Reciprocal Rank Fusion (RRF).

## Features

- **Hybrid Search**: Combines structured (hash-based) and semantic (vector) search using RRF
- **Voyage AI Integration**: Uses Voyage AI embedding models
- **Free Tier**: 200M free tokens covers typical usage

## Setup

### 1. Get Voyage AI API Key

Get your free API key from https://dashboard.voyageai.com/

### 2. Configure Claude Desktop

Add VOYAGE_API_KEY to your Claude Desktop config:

```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "/path/to/unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "VOYAGE_API_KEY": "your-api-key-here",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "/path/to/unified-thinking.db"
      }
    }
  }
}
```

### 3. Restart Claude Desktop

Semantic embeddings work automatically once VOYAGE_API_KEY is set.

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `VOYAGE_API_KEY` | - | Voyage AI API key (required) |
| `EMBEDDINGS_MODEL` | `voyage-3-lite` | Model: voyage-3-lite, voyage-3, voyage-3-large |
| `EMBEDDINGS_RRF_K` | `60` | RRF parameter |
| `EMBEDDINGS_MIN_SIMILARITY` | `0.5` | Minimum cosine similarity threshold |

## Model Selection

| Model | Dimension | Speed | Quality |
|-------|-----------|-------|---------|
| `voyage-3-lite` | 512 | Fast | Good |
| `voyage-3` | 1024 | Medium | Better |
| `voyage-3-large` | 2048 | Slower | Best |

Recommended: `voyage-3-lite` for balanced performance.

## How It Works

### Embedding Generation

When a trajectory is stored:
1. Combines problem description, context, goals, domain, and type
2. Sends to Voyage AI to generate vector embedding
3. Stores embedding with problem data

### Hybrid Search

When searching for similar problems:
1. **Structured Search**: Hash-based similarity (domain, type, goals)
2. **Vector Search**: Cosine similarity between embeddings
3. **RRF Fusion**: Combines rankings

### Reciprocal Rank Fusion

```
score = sum(1/(k + rank_i))
```

Where k=60 and rank_i is position in each result list.

## Examples

Problems that match semantically:
- "Optimize database query performance" matches "Speed up SQL execution time"
- "Fix authentication bug" matches "Resolve login security issue"
- "Implement caching layer" matches "Add memoization for performance"

## Troubleshooting

### Check Configuration

```bash
echo $VOYAGE_API_KEY  # Should be set
```

### Check Logs

```bash
DEBUG=true ./bin/unified-thinking.exe
```

Look for:
- "Semantic auto mode selection enabled" - semantic detection working
- "Successfully generated embedding" - embedding generation working

### Verify API Key

```bash
curl -X POST https://api.voyageai.com/v1/embeddings \
  -H "Authorization: Bearer $VOYAGE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model": "voyage-3-lite", "input": ["test"]}'
```

## Cost

With 200M free tokens:
- Each problem uses ~100 tokens
- Free tier covers 2,000,000 problems
- Typical usage: <10,000 problems/year

## Technical Details

### Data Model

```go
type ProblemDescription struct {
    // ... existing fields ...
    Embedding     []float32          `json:"embedding,omitempty"`
    EmbeddingMeta *EmbeddingMetadata `json:"embedding_meta,omitempty"`
}
```

### Package Structure

```
internal/embeddings/
├── embedder.go       # Interface and configuration
├── voyage.go         # Voyage AI client
├── similarity.go     # Vector similarity functions
└── cache.go          # Embedding cache
```

## References

- [Voyage AI Documentation](https://docs.voyageai.com)
- [RRF Paper](https://plg.uwaterloo.ca/~gvcormac/cormacksigir09-rrf.pdf)
- [Anthropic Contextual Retrieval](https://www.anthropic.com/news/contextual-retrieval)
