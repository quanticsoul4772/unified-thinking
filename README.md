# Unified Thinking Server

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Test Coverage](https://img.shields.io/badge/coverage-84%25-brightgreen)](https://github.com/quanticsoul4772/unified-thinking)
[![Tests](https://img.shields.io/badge/tests-148_files-success)](https://github.com/quanticsoul4772/unified-thinking)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![MCP](https://img.shields.io/badge/MCP-compatible-purple)](https://modelcontextprotocol.io/)
[![Tools](https://img.shields.io/badge/tools-80-blue)](https://github.com/quanticsoul4772/unified-thinking)

A Model Context Protocol (MCP) server that consolidates multiple cognitive thinking patterns into a single Go-based implementation with 80 specialized reasoning tools.

## Quick Start

```bash
git clone https://github.com/quanticsoul4772/unified-thinking.git
cd unified-thinking
make build
# Configure Claude Desktop (see Installation section)
# Restart Claude Desktop
```

## Features

### Thinking Modes (7)

| Mode | Description |
|------|-------------|
| **Linear** | Sequential step-by-step reasoning |
| **Tree** | Multi-branch parallel exploration with cross-references |
| **Divergent** | Creative ideation with "rebellion" capability |
| **Reflection** | Metacognitive analysis of previous reasoning |
| **Backtracking** | Checkpoint-based with restore capabilities |
| **Auto** | Automatic mode selection based on input |
| **Graph** | Graph-of-Thoughts with aggregation, refinement, cyclic reasoning |

### Tool Categories (80 tools)

| Category | Tools | Description |
|----------|-------|-------------|
| Core Thinking | 11 | think, history, branches, validation, search |
| Probabilistic | 4 | Bayesian inference, evidence assessment |
| Decision | 2 | Multi-criteria analysis, problem decomposition |
| Metacognition | 3 | Self-evaluation, bias detection, blind spots |
| Hallucination | 5 | Verification, calibration tracking |
| Perspective | 4 | Stakeholder analysis, temporal reasoning |
| Causal | 5 | Causal graphs, interventions, counterfactuals |
| Integration | 6 | Cross-mode synthesis, workflow orchestration |
| Dual-Process | 1 | System 1/2 reasoning |
| Backtracking | 3 | Checkpoints, restoration |
| Abductive | 2 | Hypothesis generation and evaluation |
| Case-Based | 2 | Similar case retrieval, CBR cycles |
| Symbolic | 2 | Theorem proving, constraint checking |
| Enhanced | 8 | Analogies, argument analysis, fallacy detection |
| Episodic Memory | 5 | Session tracking, pattern learning, recommendations |
| Knowledge Graph | 3 | Entity storage, semantic search, relationships |
| Similarity | 1 | Semantic thought search |
| Graph-of-Thoughts | 8 | GoT operations (generate, aggregate, refine, score, prune) |
| Claude Code | 5 | Session export/import, presets, formatting |

See [API_REFERENCE.md](API_REFERENCE.md) for complete tool documentation.

## Installation

### Prerequisites
- Go 1.24+
- Git

### Build

**macOS/Linux:**
```bash
git clone https://github.com/quanticsoul4772/unified-thinking.git
cd unified-thinking
make build  # Output: bin/unified-thinking
```

**Windows:**
```bash
git clone https://github.com/quanticsoul4772/unified-thinking.git
cd unified-thinking
make build  # Output: bin\unified-thinking.exe
```

## Configuration

**Config file locations:**
- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux:** `~/.config/Claude/claude_desktop_config.json`

### Minimal (In-Memory)

```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "/path/to/bin/unified-thinking",
      "env": { "DEBUG": "true" }
    }
  }
}
```

### Recommended (SQLite + Embeddings)

```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "/path/to/bin/unified-thinking",
      "env": {
        "DEBUG": "true",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "~/Library/Application Support/Claude/unified-thinking.db",
        "VOYAGE_API_KEY": "your-voyage-api-key"
      }
    }
  }
}
```

### Full (All Features)

```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "/path/to/bin/unified-thinking",
      "env": {
        "DEBUG": "true",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "~/Library/Application Support/Claude/unified-thinking.db",
        "VOYAGE_API_KEY": "your-voyage-api-key",
        "ANTHROPIC_API_KEY": "your-anthropic-api-key",
        "NEO4J_ENABLED": "true",
        "NEO4J_URI": "neo4j+s://your-instance.databases.neo4j.io",
        "NEO4J_USERNAME": "neo4j",
        "NEO4J_PASSWORD": "your-password"
      }
    }
  }
}
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `STORAGE_TYPE` | `memory` | `memory` or `sqlite` |
| `SQLITE_PATH` | - | Database file path |
| `DEBUG` | `false` | Enable debug logging |
| `VOYAGE_API_KEY` | - | Voyage AI for semantic embeddings |
| `ANTHROPIC_API_KEY` | - | Required for Graph-of-Thoughts |
| `NEO4J_ENABLED` | `false` | Enable knowledge graph |
| `NEO4J_URI` | - | Neo4j connection URI |
| `CONTEXT_BRIDGE_ENABLED` | `true` | Cross-session context retrieval |

**Notes:**
- Trajectory persistence requires `STORAGE_TYPE=sqlite`
- Graph-of-Thoughts requires `ANTHROPIC_API_KEY`
- Knowledge graph requires both `NEO4J_ENABLED=true` and `VOYAGE_API_KEY`

## Documentation

- **[API Reference](API_REFERENCE.md)** - Complete tool documentation with parameters and examples
- **[Configuration Guide](docs/CONFIGURATION.md)** - Detailed configuration options
- **[Embeddings Guide](docs/EMBEDDINGS.md)** - Semantic embeddings setup
- **[Changelog](CHANGELOG.md)** - Version history and updates

## Development

```bash
make build          # Build binary
make test           # Run tests
make test-coverage  # Coverage report
make benchmark      # Run benchmarks
make clean          # Remove artifacts
```

**Test Coverage:** ~84% overall | 148 test files | 100% pass rate

| Package | Coverage |
|---------|----------|
| metrics | 100% |
| config, presets | 97%+ |
| reasoning, reinforcement | 90%+ |
| analysis, metacognition, validation | 87%+ |
| memory, orchestration, processing | 83%+ |

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Server won't start | Check `go version` (1.24+), verify binary in `bin/` |
| Tools not appearing | Restart Claude Desktop, check config syntax |
| Data lost on restart | Set `STORAGE_TYPE=sqlite` |
| GoT tools fail | Set `ANTHROPIC_API_KEY` |

## Contributing

- [Contributing Guidelines](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Security Policy](SECURITY.md)

## License

MIT License - see [LICENSE](LICENSE)

## Support

- [GitHub Issues](https://github.com/quanticsoul4772/unified-thinking/issues)
- [GitHub Discussions](https://github.com/quanticsoul4772/unified-thinking/discussions)

---

Built with the [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
