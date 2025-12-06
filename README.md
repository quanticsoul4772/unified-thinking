# Unified Thinking Server

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Test Coverage](https://img.shields.io/badge/coverage-71%25-yellowgreen)](https://github.com/quanticsoul4772/unified-thinking)
[![Tests](https://img.shields.io/badge/tests-156_files-success)](https://github.com/quanticsoul4772/unified-thinking)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![MCP](https://img.shields.io/badge/MCP-compatible-purple)](https://modelcontextprotocol.io/)
[![Tools](https://img.shields.io/badge/tools-85-blue)](https://github.com/quanticsoul4772/unified-thinking)

A Model Context Protocol (MCP) server that consolidates multiple cognitive thinking patterns into a single Go-based implementation with 85 specialized reasoning tools.

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

### Tool Categories (85 tools)

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
| Graph-of-Thoughts | 9 | GoT operations (generate, aggregate, refine, score, prune, explore) |
| Claude Code | 5 | Session export/import, presets, formatting |
| Research | 1 | Web-augmented research with citations |
| Multimodal | 1 | Image embedding generation |
| Agentic | 2 | Autonomous tool-calling agent |

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

### Required Configuration

All infrastructure is **required** - the server fails fast if any required configuration is missing.

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
        "NEO4J_URI": "neo4j+s://your-instance.databases.neo4j.io",
        "NEO4J_USERNAME": "neo4j",
        "NEO4J_PASSWORD": "your-password",
        "NEO4J_DATABASE": "neo4j"
      }
    }
  }
}
```

### Environment Variables

**Required (server fails if missing):**

| Variable | Description |
|----------|-------------|
| `VOYAGE_API_KEY` | Voyage AI API key for embeddings |
| `ANTHROPIC_API_KEY` | Anthropic API key for GoT, agent, web search |
| `NEO4J_URI` | Neo4j connection URI |
| `NEO4J_USERNAME` | Neo4j username |
| `NEO4J_PASSWORD` | Neo4j password |

**Storage:**

| Variable | Default | Description |
|----------|---------|-------------|
| `STORAGE_TYPE` | `sqlite` | `sqlite` (recommended) or `memory` |
| `SQLITE_PATH` | `./data/unified-thinking.db` | Database file path |

**Optional:**

| Variable | Default | Description |
|----------|---------|-------------|
| `DEBUG` | `false` | Enable debug logging |
| `NEO4J_DATABASE` | `neo4j` | Neo4j database name |
| `EMBEDDINGS_MODEL` | `voyage-3-lite` | Embedding model |
| `GOT_MODEL` | `claude-sonnet-4-5-20250929` | Model for Graph-of-Thoughts |

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

**Test Coverage:** 71% overall | 156 test files | 100% pass rate

| Package | Coverage |
|---------|----------|
| metrics | 100% |
| presets | 98% |
| config | 97% |
| similarity | 95% |
| format | 95% |
| reinforcement, reasoning | 90% |
| analysis, modes, metacognition | 87% |
| memory, orchestration, validation | 83% |

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
