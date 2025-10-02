# Unified Thinking Server

A MCP server that consolidates multiple cognitive thinking patterns into a single, efficient Go-based implementation.

## Features

### Thinking Modes

- **Linear Mode**: Sequential step-by-step reasoning for systematic problem solving
- **Tree Mode**: Multi-branch parallel exploration with insights and cross-references
- **Divergent Mode**: Creative/unconventional ideation with "rebellion" capability
- **Auto Mode**: Automatic mode selection based on input content

### Core Capabilities

- Multi-mode thinking (linear, tree, divergent, auto)
- Branch management and exploration
- Insight generation and tracking
- Cross-reference support between branches
- Logical validation and consistency checking
- Formal proof attempts
- Syntax validation for logical statements
- Search across all thoughts
- Full history tracking

### Cognitive Reasoning Capabilities

The server includes advanced cognitive reasoning features that transform it from a thought recorder into a comprehensive reasoning assistant:

#### Probabilistic Reasoning
- Bayesian inference with prior and posterior belief updates
- Evidence-based probability updates
- Belief combination operations (AND/OR)
- Confidence estimation from evidence aggregation

#### Evidence Assessment
- Automatic quality classification (Strong/Moderate/Weak/Anecdotal)
- Reliability scoring based on source credibility and content
- Relevance calculation and weighted scoring
- Multi-source evidence aggregation

#### Analysis Tools
- Contradiction detection across thoughts (negations, absolutes, modals)
- Sensitivity analysis for testing assumption robustness
- Multi-perspective stakeholder analysis
- Temporal reasoning (short-term vs long-term implications)

#### Decision Support
- Multi-criteria decision analysis (MCDA) with weighted scoring
- Problem decomposition into manageable subproblems
- Dependency mapping and solution path determination
- Cross-mode insight synthesis and branch merging

#### Metacognition
- Self-evaluation of thought quality, completeness, and coherence
- Cognitive bias detection (confirmation, anchoring, availability, sunk cost, overconfidence, recency, groupthink)
- Bias severity classification and mitigation strategies
- Strength/weakness identification with improvement suggestions

#### Advanced Reasoning
- Analogical reasoning with cross-domain mapping
- Enhanced creative ideation techniques
- Pattern-based reasoning and inference

## Installation

### Prerequisites

- Go 1.23 or higher
- Git

### Build

```bash
go mod download
go build -o bin/unified-thinking.exe ./cmd/server
```

Or using make:

```bash
make build
```

## Configuration

Add to your Claude Desktop config (`%APPDATA%\Claude\claude_desktop_config.json` on Windows):

```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "/path/to/unified-thinking/bin/unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "DEBUG": "true"
      }
    }
  }
}
```

After saving the config, restart Claude Desktop.

### Available Tools

1. **think** - Main thinking tool
   ```json
   {
     "content": "Your thinking prompt",
     "mode": "auto|linear|tree|divergent",
     "confidence": 0.8,
     "key_points": ["point1", "point2"],
     "require_validation": true
   }
   ```

2. **history** - View thinking history
   ```json
   {
     "mode": "linear|tree|divergent",
     "branch_id": "optional"
   }
   ```

3. **list-branches** - List all branches (tree mode)

4. **focus-branch** - Switch active branch
   ```json
   {
     "branch_id": "branch-xxx"
   }
   ```

5. **branch-history** - Get detailed branch history
   ```json
   {
     "branch_id": "branch-xxx"
   }
   ```

6. **validate** - Validate a thought for logical consistency
   ```json
   {
     "thought_id": "thought-xxx"
   }
   ```

7. **prove** - Attempt to prove a logical conclusion
   ```json
   {
     "premises": ["All humans are mortal", "Socrates is human"],
     "conclusion": "Socrates is mortal"
   }
   ```

8. **check-syntax** - Validate logical statement syntax
   ```json
   {
     "statements": ["Statement 1", "Statement 2"]
   }
   ```

9. **search** - Search thoughts
   ```json
   {
     "query": "search term",
     "mode": "optional mode filter"
   }
   ```

10. **get-metrics** - Get system performance and usage metrics

11. **recent-branches** - Get recently accessed branches for quick context switching
    - Returns the last 10 accessed branches with timestamps
    - Shows active branch for context
    - Enables fast branch switching without remembering IDs

12. **probabilistic-reasoning** - Bayesian inference and belief updates
    ```json
    {
      "operation": "create|update|get|combine",
      "statement": "It will rain tomorrow",
      "prior_prob": 0.3,
      "belief_id": "belief-xxx",
      "likelihood": 0.8,
      "evidence_prob": 0.4
    }
    ```

13. **assess-evidence** - Evidence quality assessment
    ```json
    {
      "content": "Evidence content",
      "source": "Source reference",
      "claim_id": "claim-xxx",
      "supports_claim": true
    }
    ```

14. **detect-contradictions** - Find contradictions among thoughts
    ```json
    {
      "thought_ids": ["thought-1", "thought-2"],
      "branch_id": "branch-xxx",
      "mode": "linear|tree|divergent"
    }
    ```

15. **make-decision** - Multi-criteria decision analysis
    ```json
    {
      "question": "Which option should we choose?",
      "options": [{"name": "Option A", "scores": {"cost": 0.8}}],
      "criteria": [{"name": "Cost", "weight": 0.6, "maximize": false}]
    }
    ```

16. **decompose-problem** - Break down complex problems
    ```json
    {
      "problem": "Complex problem description"
    }
    ```

17. **sensitivity-analysis** - Test robustness of conclusions
    ```json
    {
      "target_claim": "Main conclusion",
      "assumptions": ["assumption1", "assumption2"],
      "base_confidence": 0.8
    }
    ```

18. **self-evaluate** - Metacognitive self-assessment
    ```json
    {
      "thought_id": "thought-xxx",
      "branch_id": "branch-xxx"
    }
    ```

19. **detect-biases** - Identify cognitive biases
    ```json
    {
      "thought_id": "thought-xxx",
      "branch_id": "branch-xxx"
    }
    ```

### Example Prompts

**Auto Mode (Recommended)**:
```
"Analyze this problem using the best thinking approach"
```

**Explicit Linear Mode**:
```
"Think step by step about solving this"
```

**Explicit Tree Mode**:
```
"Explore multiple branches of this idea with cross-references"
```

**Explicit Divergent Mode**:
```
"What's a creative, unconventional solution to this?"
"Challenge all assumptions about this problem" (with force_rebellion)
```

## Architecture

```
unified-thinking/
├── cmd/server/          # Main entry point
├── internal/
│   ├── types/          # Core data structures (extended with cognitive types)
│   ├── storage/        # In-memory storage
│   ├── modes/          # Thinking mode implementations
│   │   ├── linear.go
│   │   ├── tree.go
│   │   ├── divergent.go
│   │   └── auto.go
│   ├── reasoning/      # Probabilistic reasoning and decision making
│   ├── analysis/       # Evidence assessment, contradiction detection, sensitivity analysis
│   ├── metacognition/  # Self-evaluation and bias detection
│   ├── validation/     # Logic validation
│   └── server/         # MCP server implementation
└── TECHNICAL_PLAN.md   # Detailed technical documentation
```

### Cognitive Architecture

The server implements a modular cognitive architecture with three specialized packages:

- **reasoning**: Implements Bayesian probabilistic inference, multi-criteria decision analysis, and problem decomposition
- **analysis**: Provides evidence quality assessment, contradiction detection, and sensitivity analysis for robustness testing
- **metacognition**: Enables self-evaluation and cognitive bias detection with mitigation strategies

All components are thread-safe, composable, and maintain backward compatibility with existing functionality.

## Development

### Build

```bash
# Build the server binary
make build

# Clean build artifacts
make clean
```

```bash
# For protocol debugging only (waits for MCP messages on stdin)
go run ./cmd/server/main.go

# With debug logging
DEBUG=true go run ./cmd/server/main.go
```

### Testing

```bash
# Run tests
make test

# Run with verbose output
go test -v ./...
```

## Troubleshooting

### Server won't start

1. Check that Go is installed: `go version`
2. Verify the binary was built: Check `bin/` directory
3. Enable debug mode: Set `DEBUG=true` in env

### Tools not appearing

1. Restart Claude Desktop completely
2. Check config file syntax
3. Verify the executable path is correct

### Performance issues

- The server uses in-memory storage
- For long sessions, consider periodic restarts
- Monitor memory usage if processing many thoughts

## Contributing

See [TECHNICAL_PLAN.md](TECHNICAL_PLAN.md) for detailed architecture and implementation notes.

## License

MIT License

## Technical Details

### Data Structures

The system includes 15 specialized cognitive data structures:

- **Evidence**: Quality assessment with reliability and relevance scoring
- **ProbabilisticBelief**: Bayesian beliefs with prior/posterior tracking
- **Contradiction**: Cross-thought contradiction detection results
- **Perspective**: Stakeholder viewpoint analysis
- **TemporalAnalysis**: Short-term vs long-term reasoning
- **Decision**: Structured decision framework with criteria and options
- **ProblemDecomposition**: Complex problem breakdown into subproblems
- **Synthesis**: Cross-mode insight integration
- **SensitivityAnalysis**: Robustness testing of conclusions
- **Analogy**: Cross-domain analogical reasoning
- **SelfEvaluation**: Metacognitive self-assessment
- **CognitiveBias**: Detected biases with mitigation strategies

### Implementation Features

- Modular package design for composability
- Thread-safe operations with proper locking patterns
- Comprehensive test coverage for all cognitive capabilities
- Backward compatible with existing functionality
- In-memory storage optimized for performance

## Version

Version 1.0.0 - Initial unified release with cognitive reasoning capabilities
