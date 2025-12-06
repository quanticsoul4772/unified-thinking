# Unified Thinking Server - Configuration Guide

This guide explains how to configure the unified-thinking MCP server using environment variables, configuration files, and feature flags.

## Table of Contents

- [Quick Start](#quick-start)
- [Configuration Sources](#configuration-sources)
- [Server Settings](#server-settings)
- [Storage Settings](#storage-settings)
- [Feature Flags](#feature-flags)
- [Performance Settings](#performance-settings)
- [Logging Settings](#logging-settings)
- [Environment Variables](#environment-variables)
- [Configuration File](#configuration-file)
- [Examples](#examples)

## Quick Start

The server works out-of-the-box with sensible defaults. No configuration is required to get started.

```bash
# Run with defaults
./unified-thinking

# Override specific settings with environment variables
UT_LOGGING_LEVEL=debug ./unified-thinking

# Load from configuration file
./unified-thinking --config=config.json
```

## Configuration Sources

Configuration is loaded in the following order of precedence (highest to lowest):

1. **Environment Variables** (highest priority) - Overrides everything
2. **Configuration File** (if provided) - Overrides defaults
3. **Default Values** (lowest priority) - Built-in sensible defaults

This allows you to use a configuration file for most settings and override specific values with environment variables for deployment-specific configuration.

## Server Settings

### server.name

**Description**: Name of the server instance (for logging and identification)

**Default**: `"unified-thinking"`

**Environment Variable**: `UT_SERVER_NAME`

**Example**:
```bash
export UT_SERVER_NAME="my-thinking-server"
```

### server.version

**Description**: Version of the server

**Default**: `"1.0.0"`

**Environment Variable**: `UT_SERVER_VERSION`

### server.environment

**Description**: Deployment environment

**Default**: `"development"`

**Valid Values**: `development`, `staging`, `production`

**Environment Variable**: `UT_SERVER_ENVIRONMENT`

**Example**:
```bash
export UT_SERVER_ENVIRONMENT="production"
```

## Storage Settings

### storage.type

**Description**: Type of storage backend

**Default**: `"sqlite"`

**Valid Values**: `sqlite`, `memory`

**Environment Variable**: `STORAGE_TYPE` (Claude Desktop) or `UT_STORAGE_TYPE` (standalone)

**Note**: SQLite is the recommended and default storage type. It provides persistence and is required for full functionality including knowledge graph and trajectory storage.

**Example**:
```bash
export STORAGE_TYPE=sqlite
export SQLITE_PATH=./data/thoughts.db
```

### storage.max_thoughts

**Description**: Maximum number of thoughts to store (0 = unlimited)

**Default**: `0`

**Environment Variable**: `UT_STORAGE_MAX_THOUGHTS`

**Example**:
```bash
export UT_STORAGE_MAX_THOUGHTS=10000
```

### storage.max_branches

**Description**: Maximum number of branches to store (0 = unlimited)

**Default**: `0`

**Environment Variable**: `UT_STORAGE_MAX_BRANCHES`

### storage.enable_indexing

**Description**: Enable content indexing for faster search

**Default**: `true`

**Environment Variable**: `UT_STORAGE_ENABLE_INDEXING`

**Example**:
```bash
export UT_STORAGE_ENABLE_INDEXING=false
```

### SQLite-Specific Settings

#### SQLITE_PATH

**Description**: Path to SQLite database file (created if not exists)

**Default**: `"./data/thoughts.db"`

**Environment Variable**: `SQLITE_PATH`

**Example**:
```bash
export SQLITE_PATH="C:\\Users\\YourName\\AppData\\Roaming\\Claude\\unified-thinking.db"
```

#### SQLITE_TIMEOUT

**Description**: Connection timeout in milliseconds

**Default**: `5000`

**Environment Variable**: `SQLITE_TIMEOUT`

**Example**:
```bash
export SQLITE_TIMEOUT=10000
```

**Note**: The server uses fail-fast behavior. If the configured storage backend fails to initialize, the server will terminate immediately rather than falling back to an alternative storage type.

## Feature Flags

Feature flags allow you to enable or disable specific capabilities. All features are enabled by default.

### Core Thinking Modes

| Feature | Environment Variable | Default | Description |
|---------|---------------------|---------|-------------|
| `linear_mode` | `UT_FEATURES_LINEAR_MODE` | `true` | Sequential logical thinking |
| `tree_mode` | `UT_FEATURES_TREE_MODE` | `true` | Branching exploration thinking |
| `divergent_mode` | `UT_FEATURES_DIVERGENT_MODE` | `true` | Creative divergent thinking |
| `auto_mode` | `UT_FEATURES_AUTO_MODE` | `true` | Automatic mode selection |

### Validation Features

| Feature | Environment Variable | Default | Description |
|---------|---------------------|---------|-------------|
| `logical_validation` | `UT_FEATURES_LOGICAL_VALIDATION` | `true` | Validate logical consistency |
| `proof_generation` | `UT_FEATURES_PROOF_GENERATION` | `true` | Generate logical proofs |
| `syntax_checking` | `UT_FEATURES_SYNTAX_CHECKING` | `true` | Check logical syntax |

### Advanced Reasoning

| Feature | Environment Variable | Default | Description |
|---------|---------------------|---------|-------------|
| `probabilistic_reasoning` | `UT_FEATURES_PROBABILISTIC_REASONING` | `true` | Bayesian inference |
| `decision_making` | `UT_FEATURES_DECISION_MAKING` | `true` | Multi-criteria decisions |
| `problem_decomposition` | `UT_FEATURES_PROBLEM_DECOMPOSITION` | `true` | Break down complex problems |

### Analysis Capabilities

| Feature | Environment Variable | Default | Description |
|---------|---------------------|---------|-------------|
| `evidence_assessment` | `UT_FEATURES_EVIDENCE_ASSESSMENT` | `true` | Assess evidence quality |
| `contradiction_detection` | `UT_FEATURES_CONTRADICTION_DETECTION` | `true` | Detect contradictions |
| `sensitivity_analysis` | `UT_FEATURES_SENSITIVITY_ANALYSIS` | `true` | Test assumption robustness |

### Metacognition

| Feature | Environment Variable | Default | Description |
|---------|---------------------|---------|-------------|
| `self_evaluation` | `UT_FEATURES_SELF_EVALUATION` | `true` | Self-assess reasoning quality |
| `bias_detection` | `UT_FEATURES_BIAS_DETECTION` | `true` | Identify cognitive biases |

### Utilities

| Feature | Environment Variable | Default | Description |
|---------|---------------------|---------|-------------|
| `search_enabled` | `UT_FEATURES_SEARCH_ENABLED` | `true` | Enable search functionality |
| `history_enabled` | `UT_FEATURES_HISTORY_ENABLED` | `true` | Enable history tracking |
| `metrics_enabled` | `UT_FEATURES_METRICS_ENABLED` | `true` | Enable metrics collection |

### Example: Disable Specific Features

```bash
# Disable probabilistic reasoning
export UT_FEATURES_PROBABILISTIC_REASONING=false

# Disable bias detection
export UT_FEATURES_BIAS_DETECTION=false
```

## Performance Settings

### performance.max_concurrent_thoughts

**Description**: Maximum number of thoughts processed concurrently

**Default**: `100`

**Environment Variable**: `UT_PERFORMANCE_MAX_CONCURRENT_THOUGHTS`

**Example**:
```bash
export UT_PERFORMANCE_MAX_CONCURRENT_THOUGHTS=50
```

### performance.enable_deep_copy

**Description**: Enable deep copying for thread safety (recommended to keep enabled)

**Default**: `true`

**Environment Variable**: `UT_PERFORMANCE_ENABLE_DEEP_COPY`

### performance.cache_size

**Description**: Size of internal caches (0 = no caching)

**Default**: `1000`

**Environment Variable**: `UT_PERFORMANCE_CACHE_SIZE`

## Logging Settings

### logging.level

**Description**: Logging verbosity level

**Default**: `"info"`

**Valid Values**: `debug`, `info`, `warn`, `error`

**Environment Variable**: `UT_LOGGING_LEVEL`

**Example**:
```bash
export UT_LOGGING_LEVEL=debug
```

### logging.format

**Description**: Log output format

**Default**: `"text"`

**Valid Values**: `text`, `json`

**Environment Variable**: `UT_LOGGING_FORMAT`

**Example**:
```bash
export UT_LOGGING_FORMAT=json
```

### logging.enable_timestamps

**Description**: Include timestamps in log entries

**Default**: `true`

**Environment Variable**: `UT_LOGGING_ENABLE_TIMESTAMPS`

## Environment Variables

All environment variables follow the naming convention: `UT_<SECTION>_<KEY>`

### Boolean Values

Boolean environment variables accept various formats:

- **True**: `true`, `TRUE`, `1`, `yes`, `YES`, `on`, `ON`, `enabled`, `ENABLED`
- **False**: `false`, `FALSE`, `0`, `no`, `NO`, `off`, `OFF`, `disabled`, `DISABLED`, or empty

### Complete Environment Variable List

```bash
# Server
UT_SERVER_NAME=unified-thinking
UT_SERVER_VERSION=1.0.0
UT_SERVER_ENVIRONMENT=development

# Storage
UT_STORAGE_TYPE=sqlite
UT_STORAGE_MAX_THOUGHTS=0
UT_STORAGE_MAX_BRANCHES=0
UT_STORAGE_ENABLE_INDEXING=true

# Features - Core Modes
UT_FEATURES_LINEAR_MODE=true
UT_FEATURES_TREE_MODE=true
UT_FEATURES_DIVERGENT_MODE=true
UT_FEATURES_AUTO_MODE=true

# Features - Validation
UT_FEATURES_LOGICAL_VALIDATION=true
UT_FEATURES_PROOF_GENERATION=true
UT_FEATURES_SYNTAX_CHECKING=true

# Features - Advanced Reasoning
UT_FEATURES_PROBABILISTIC_REASONING=true
UT_FEATURES_DECISION_MAKING=true
UT_FEATURES_PROBLEM_DECOMPOSITION=true

# Features - Analysis
UT_FEATURES_EVIDENCE_ASSESSMENT=true
UT_FEATURES_CONTRADICTION_DETECTION=true
UT_FEATURES_SENSITIVITY_ANALYSIS=true

# Features - Metacognition
UT_FEATURES_SELF_EVALUATION=true
UT_FEATURES_BIAS_DETECTION=true

# Features - Utilities
UT_FEATURES_SEARCH_ENABLED=true
UT_FEATURES_HISTORY_ENABLED=true
UT_FEATURES_METRICS_ENABLED=true

# Performance
UT_PERFORMANCE_MAX_CONCURRENT_THOUGHTS=100
UT_PERFORMANCE_ENABLE_DEEP_COPY=true
UT_PERFORMANCE_CACHE_SIZE=1000

# Logging
UT_LOGGING_LEVEL=info
UT_LOGGING_FORMAT=text
UT_LOGGING_ENABLE_TIMESTAMPS=true
```

## Configuration File

Create a JSON configuration file to set multiple options at once:

### config.json Example

```json
{
  "server": {
    "name": "unified-thinking",
    "version": "1.0.0",
    "environment": "production"
  },
  "storage": {
    "type": "sqlite",
    "max_thoughts": 50000,
    "max_branches": 1000,
    "enable_indexing": true
  },
  "features": {
    "linear_mode": true,
    "tree_mode": true,
    "divergent_mode": true,
    "auto_mode": true,
    "logical_validation": true,
    "proof_generation": true,
    "syntax_checking": true,
    "probabilistic_reasoning": true,
    "decision_making": true,
    "problem_decomposition": true,
    "evidence_assessment": true,
    "contradiction_detection": true,
    "sensitivity_analysis": true,
    "self_evaluation": true,
    "bias_detection": true,
    "search_enabled": true,
    "history_enabled": true,
    "metrics_enabled": true
  },
  "performance": {
    "max_concurrent_thoughts": 100,
    "enable_deep_copy": true,
    "cache_size": 1000
  },
  "logging": {
    "level": "info",
    "format": "text",
    "enable_timestamps": true
  }
}
```

### Loading Configuration File

```bash
# Load from config file
./unified-thinking --config=config.json

# Override specific values with environment variables
UT_LOGGING_LEVEL=debug ./unified-thinking --config=config.json
```

## Examples

### Development Environment

```bash
# Verbose logging for development
export UT_SERVER_ENVIRONMENT=development
export UT_LOGGING_LEVEL=debug
export UT_LOGGING_FORMAT=text

./unified-thinking
```

### Production Environment with SQLite

```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:\\path\\to\\bin\\unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "DEBUG": "false",
        "STORAGE_TYPE": "sqlite",
        "SQLITE_PATH": "C:\\ProgramData\\unified-thinking\\thoughts.db",
        "AUTO_VALIDATION_THRESHOLD": "0.5"
      }
    }
  }
}
```

### High-Performance Production Environment

```json
{
  "server": {
    "name": "prod-thinking-server",
    "environment": "production"
  },
  "storage": {
    "type": "sqlite",
    "max_thoughts": 100000,
    "max_branches": 5000
  },
  "performance": {
    "max_concurrent_thoughts": 200,
    "cache_size": 5000
  },
  "logging": {
    "level": "warn",
    "format": "json",
    "enable_timestamps": true
  }
}
```

### Minimal Feature Set

If you only need basic thinking without advanced features:

```bash
# Enable only core features
export UT_FEATURES_LINEAR_MODE=true
export UT_FEATURES_TREE_MODE=true
export UT_FEATURES_DIVERGENT_MODE=false
export UT_FEATURES_AUTO_MODE=true

# Disable advanced reasoning
export UT_FEATURES_PROBABILISTIC_REASONING=false
export UT_FEATURES_DECISION_MAKING=false
export UT_FEATURES_PROBLEM_DECOMPOSITION=false

# Disable analysis features
export UT_FEATURES_EVIDENCE_ASSESSMENT=false
export UT_FEATURES_CONTRADICTION_DETECTION=false
export UT_FEATURES_SENSITIVITY_ANALYSIS=false

# Disable metacognition
export UT_FEATURES_SELF_EVALUATION=false
export UT_FEATURES_BIAS_DETECTION=false

./unified-thinking
```

### Performance Tuning

For high-throughput scenarios:

```bash
# Increase concurrency
export UT_PERFORMANCE_MAX_CONCURRENT_THOUGHTS=500

# Increase cache size
export UT_PERFORMANCE_CACHE_SIZE=10000

# Disable storage limits
export UT_STORAGE_MAX_THOUGHTS=0
export UT_STORAGE_MAX_BRANCHES=0

./unified-thinking
```

### Testing Environment

```bash
# Enable all features for testing
export UT_SERVER_ENVIRONMENT=development
export UT_LOGGING_LEVEL=debug

# Set limits for test data
export UT_STORAGE_MAX_THOUGHTS=1000
export UT_STORAGE_MAX_BRANCHES=100

./unified-thinking
```

## Validation

The configuration system validates all settings on startup. If validation fails, the server will not start and will display an error message explaining the issue.

Common validation errors:
- Empty server name
- Invalid environment (must be development, staging, or production)
- Invalid storage type (must be sqlite or memory)
- Negative values for max_thoughts, max_branches, or cache_size
- max_concurrent_thoughts less than 1
- Invalid log level or format

## Programmatic Access

From within the Go code, you can access configuration:

```go
import "unified-thinking/internal/config"

// Load configuration
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Check feature flags
if cfg.IsFeatureEnabled("probabilistic_reasoning") {
    // Feature is enabled
}

// Access settings
maxThoughts := cfg.Storage.MaxThoughts
logLevel := cfg.Logging.Level
```

## Best Practices

1. **Use defaults for development** - The defaults are designed for local development
2. **Use configuration files for deployment** - Easier to manage multiple settings
3. **Use environment variables for secrets** - Never commit sensitive data to config files
4. **Enable all features initially** - Disable features only if needed for specific use cases
5. **Monitor performance** - Adjust `max_concurrent_thoughts` and `cache_size` based on load
6. **Use structured logging in production** - Set `logging.format=json` for production environments
7. **Set appropriate storage limits** - Prevent unbounded memory growth in long-running servers

## Troubleshooting

### Server Won't Start

Check validation errors in the startup logs. Common issues:
- Invalid environment variable format
- Missing required settings
- Incompatible feature combinations

### Performance Issues

- Increase `performance.max_concurrent_thoughts` for high concurrency
- Increase `performance.cache_size` for better performance
- Enable `storage.enable_indexing` for faster searches

### Memory Usage

- Set `storage.max_thoughts` and `storage.max_branches` limits
- Reduce `performance.cache_size`
- Disable unused features

### Logging Not Working

- Check `logging.level` is set correctly
- Verify `logging.format` matches your log aggregation system
- Ensure `logging.enable_timestamps` is enabled if required

## Related Documentation

- [Quick Start Guide](../QUICKSTART.md)
- [API Reference](API.md)
- [Architecture Overview](ARCHITECTURE.md)

---

**Last Updated**: 2025-10-07
**Version**: 1.1.0

## Changes in Version 1.1.0

- Added SQLite storage backend support
- Added `STORAGE_TYPE`, `SQLITE_PATH`, `SQLITE_TIMEOUT` environment variables
- Added `AUTO_VALIDATION_THRESHOLD` for confidence-based auto-validation
- Updated examples to show SQLite configuration
- All storage backends now support FTS5 full-text search

## Changes in Version 1.2.0

- Removed `STORAGE_FALLBACK` environment variable (now uses fail-fast behavior)
- Server terminates immediately if configured storage backend fails to initialize
- No silent fallback to alternative storage types
