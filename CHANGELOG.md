# Changelog

All notable changes to the Unified Thinking MCP Server project.

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

### Known Limitations
- Validation and proof implementations are simplified heuristics, not formal logic engines
- In-memory storage is unbounded and will grow with usage
- No persistence across server restarts
- Windows-specific build process in current setup
