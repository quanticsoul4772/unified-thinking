# Changelog

All notable changes to the Unified Thinking MCP Server project.

## [1.1.0] - 2025-10-07

### Added

**Major Features:**
- ✅ SQLite persistence layer with write-through caching
- ✅ Workflow orchestration system for automated multi-tool reasoning
- ✅ 22 new cognitive reasoning tools (33 total, up from 9)
- ✅ Causal reasoning with Pearl's do-calculus framework
- ✅ Temporal reasoning for short-term vs long-term analysis
- ✅ Perspective analysis for multi-stakeholder reasoning
- ✅ Enhanced fallacy detection (20+ fallacy types)
- ✅ Auto-validation system with confidence thresholds
- ✅ Comprehensive metrics framework

**New Tools:**
- `recent-branches` - Quick context switching for recently accessed branches
- `get-metrics` - System performance and usage metrics
- `probabilistic-reasoning` - Bayesian inference with belief updates (create/update/get/combine)
- `assess-evidence` - Evidence quality and reliability assessment
- `detect-contradictions` - Find contradictions among thoughts or within branches
- `make-decision` - Multi-criteria decision analysis with weighted scoring
- `decompose-problem` - Break complex problems into manageable subproblems
- `sensitivity-analysis` - Test robustness of conclusions to assumption changes
- `self-evaluate` - Metacognitive self-assessment of reasoning quality
- `detect-biases` - Identify cognitive biases and logical fallacies
- `analyze-perspectives` - Multi-stakeholder perspective analysis
- `analyze-temporal` - Short-term vs long-term temporal reasoning
- `identify-optimal-timing` - Determine optimal timing for decisions
- `compare-time-horizons` - Compare implications across time horizons
- `analyze-correlation-vs-causation` - Distinguish correlation from causation
- `build-causal-graph` - Construct causal graphs from observations
- `get-causal-graph` - Retrieve previously built causal graph
- `simulate-intervention` - Simulate interventions using do-calculus
- `generate-counterfactual` - Generate "what if" scenarios
- `synthesize-insights` - Synthesize insights from multiple reasoning modes
- `detect-emergent-patterns` - Detect emergent patterns across modes
- `execute-workflow` - Execute predefined multi-tool workflows
- `list-workflows` - List available automated workflows
- `register-workflow` - Register custom workflows for automation

**Storage Enhancements:**
- SQLite backend with persistent storage across server restarts
- Write-through caching for optimal performance
- FTS5 full-text search for fast content queries
- WAL mode for concurrent read access
- Schema versioning and automatic migrations
- Graceful fallback to in-memory storage on errors
- Storage factory pattern for extensibility

**Cognitive Architecture:**
- `internal/reasoning/` - Probabilistic inference, decision analysis, causal reasoning, temporal reasoning
- `internal/analysis/` - Evidence assessment, contradiction detection, perspective analysis, sensitivity testing
- `internal/metacognition/` - Self-evaluation and cognitive bias detection
- `internal/integration/` - Cross-mode synthesis and integration patterns
- `internal/orchestration/` - Workflow orchestration for automated tool chaining

**Testing Improvements:**
- Test coverage increased from 65% to 73.6%
- Storage layer coverage: 80.5%
- Validation package coverage: 94.2%
- New test files for concurrent operations
- Comprehensive test coverage for orchestration, workflows, and evidence pipelines
- Property-based testing infrastructure

### Changed

- Enhanced CLAUDE.md with detailed architecture and 33-tool documentation
- Updated README.md with complete feature list and cognitive reasoning capabilities
- Improved error handling across all modules
- Optimized storage append operations to avoid full get-modify-store cycles
- Enhanced auto mode detection with additional keyword triggers
- Refactored handlers into focused modules (thinking, branches, validation, search, enhanced)

### Fixed

- Branch priority calculation now properly weights cross-reference strengths
- Deep copy functions prevent external mutation of stored data
- Concurrent access patterns now fully thread-safe
- Auto-validation edge cases with confidence thresholds

### Configuration

**New Environment Variables:**
- `STORAGE_TYPE` - Choose between `memory` (default) or `sqlite`
- `SQLITE_PATH` - Path to SQLite database file
- `SQLITE_TIMEOUT` - Connection timeout in milliseconds (default: 5000)
- `STORAGE_FALLBACK` - Fallback storage type if primary fails
- `AUTO_VALIDATION_THRESHOLD` - Confidence threshold for auto-validation (default: 0.5)

### Documentation

- Added TESTING.md - Comprehensive testing guide
- Updated QUICKSTART.md with 33 tools and SQLite configuration
- Enhanced CLAUDE.md with workflow orchestration and causal reasoning details
- Consolidated test documentation (removed redundant files)
- Added inline documentation for all new cognitive modules

### Performance

- Write-through caching reduces SQLite latency
- FTS5 full-text search for fast content queries
- 64MB SQLite cache size for optimal performance
- Optimized batch insert operations
- Connection pooling for database reuse

### Migration Notes

**From 1.0.0 to 1.1.0:**

1. **Storage Configuration** (Optional):
   - To enable persistence, add `STORAGE_TYPE=sqlite` to your environment
   - Set `SQLITE_PATH` to desired database location
   - Data from in-memory mode cannot be migrated (no persistence in 1.0.0)

2. **New Tools Available**:
   - 22 new tools are automatically available (no configuration needed)
   - See QUICKSTART.md or README.md for complete tool list

3. **Backward Compatibility**:
   - All 1.0.0 tools continue to work unchanged
   - Default behavior (in-memory storage) is identical to 1.0.0
   - No breaking changes to existing tool interfaces

---

## [1.0.0] - 2025-09-30

### Added

- Initial unified server implementation consolidating 5 separate MCP servers
- 9 MCP tools: think, history, list-branches, focus-branch, branch-history, validate, prove, check-syntax, search
- Four cognitive thinking modes: linear, tree, divergent, and auto
- In-memory storage with thread-safe operations using RWMutex
- Deep copying strategy for all data retrieval to prevent race conditions
- Comprehensive input validation for all request types
- Logical validation with basic consistency checking
- Simplified proof attempt functionality
- Syntax validation for logical statements
- Cross-reference support between thinking branches
- Insight generation and tracking
- Full thought history and search capabilities
- Stdio-based MCP protocol transport
- Debug logging controlled by DEBUG environment variable
- Test suite achieving 65% code coverage

### Changed

- Replaced formatted text responses with pure JSON for AI consumption
- All tool handlers now return proper CallToolResult with Content field
- Storage Get methods return deep copies instead of internal pointers
- Response format optimized for Claude AI (84% size reduction vs formatted version)
- Documentation updated to reflect MCP server architecture and stdio communication

### Fixed

- MCP protocol violations where handlers returned nil for CallToolResult
- Data race conditions in concurrent access to storage maps
- Missing input validation allowing unbounded resource consumption
- Race condition in GetActiveBranch when active branch was deleted
- Error swallowing in handleThink where validation errors were ignored

### Technical Details

#### MCP Protocol Compliance
- All 9 tool handlers return valid CallToolResult structures
- Responses use TextContent with JSON-serialized data
- Proper error propagation through MCP error handling

#### Thread Safety
- All storage operations protected by sync.RWMutex
- Read operations use RLock for concurrent access
- Write operations use exclusive Lock
- Deep copy functions prevent external mutation: copyThought, copyBranch, copyInsight, copyCrossRef, copyValidation

#### Input Validation Limits
- Content: 100KB maximum
- Key points: 50 items, 1KB each
- Cross-references: 20 maximum
- Statements: 100 items, 10KB each
- Premises: 50 items, 10KB each
- Query strings: 1KB maximum
- IDs: 100 bytes maximum
- All strings validated for UTF-8 encoding

#### Replaced Servers
- sequential-thinking (linear mode)
- branch-thinking (tree mode)
- unreasonable-thinking-server (divergent mode with rebellion)
- mcp-logic (partial: prove and check-syntax tools)
- state-coordinator (partial: branch management)

### Migration

#### Tool Mapping
- sequential-thinking solve-problem → think with mode="linear"
- branch-thinking branch-thinking → think with mode="tree"
- unreasonable-thinking generate_unreasonable_thought → think with mode="divergent" force_rebellion=true
- mcp-logic prove → prove (same tool)
- mcp-logic check-well-formed → check-syntax

#### Configuration
Add to claude_desktop_config.json:
```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "path/to/unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "DEBUG": "true"
      }
    }
  }
}
```

### Dependencies
- Go 1.23 or higher
- github.com/modelcontextprotocol/go-sdk v0.8.0
- modernc.org/sqlite (for SQLite backend in 1.1.0+)

### Known Limitations
- Validation and proof implementations are simplified heuristics, not formal logic engines
- Auto mode detection uses keyword matching (may improve in future versions)

---

## Version History

- **1.1.0** (2025-10-07) - Major feature release with 33 tools, SQLite persistence, workflow orchestration
- **1.0.0** (2025-09-30) - Initial release with 9 tools and in-memory storage
