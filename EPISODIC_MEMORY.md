# Episodic Reasoning Memory System

## Overview

The Episodic Reasoning Memory System transforms the unified-thinking server from a stateless tool provider into a **learning cognitive partner** that improves over time by learning from past reasoning sessions.

## Architecture

### Core Components

```
Episodic Memory System
├── EpisodicMemoryStore (episodic.go)
│   ├── Trajectory Storage
│   ├── Pattern Storage  
│   ├── Multi-Index System (problem, domain, tags, tool sequences)
│   └── Similarity-based Retrieval
│
├── SessionTracker (tracker.go)
│   ├── Active Session Tracking
│   ├── Step Recording
│   ├── Quality Metrics Calculation
│   └── Trajectory Building
│
├── LearningEngine (learning.go)
│   ├── Pattern Recognition
│   ├── Success/Failure Analysis
│   ├── Strategy Extraction
│   └── Continuous Learning
│
└── EpisodicMemoryHandler (handlers/episodic.go)
    ├── start-reasoning-session
    ├── complete-reasoning-session
    ├── get-recommendations
    └── search-trajectories
```

### Key Data Structures

#### ReasoningTrajectory
Complete record of a reasoning session from problem to solution:
- **Problem**: Initial state, goals, constraints, complexity
- **Approach**: Strategy, modes used, tool sequence, key decisions
- **Steps**: Detailed execution trace with inputs/outputs
- **Outcome**: Success status, goals achieved, solution, confidence
- **Quality**: Efficiency, coherence, completeness, innovation metrics
- **Metadata**: Domain, tags, complexity, success score

#### TrajectoryPattern
Learned pattern from multiple successful trajectories:
- **Problem Signature**: Domain, type, complexity range, required capabilities
- **Successful Approach**: Proven strategy and tool sequences
- **Performance Metrics**: Success rate, average quality, usage count
- **Examples**: Links to exemplar trajectories

#### Recommendation
Adaptive guidance based on episodic memory:
- **Type**: tool_sequence, approach, warning, optimization
- **Suggestion**: Specific actionable advice
- **Reasoning**: Why this recommendation (based on past performance)
- **Success Rate**: Historical success rate of this approach
- **Estimated Impact**: Expected improvement

## How It Works

### 1. Session Tracking

**Start Session**:
```
User calls: start-reasoning-session
├── Creates ProblemDescription
├── Retrieves similar past trajectories
├── Generates initial recommendations
└── Returns suggestions based on historical success
```

**During Session**:
```
Each tool invocation automatically:
├── Records as ReasoningStep
├── Tracks tool usage patterns
├── Monitors mode transitions
└── Builds complete trajectory
```

**Complete Session**:
```
User calls: complete-reasoning-session
├── Calculates quality metrics
├── Computes success score
├── Stores trajectory in episodic memory
├── Triggers pattern learning
└── Returns performance analysis
```

### 2. Pattern Learning

**Continuous Learning Cycle**:
```
Every N trajectories (or time interval):
├── Group trajectories by problem signature
├── Identify common approaches in successful cases
├── Extract tool sequence patterns
├── Calculate success rates
├── Cache learned patterns
└── Update recommendations
```

**Pattern Recognition**:
- **Problem Similarity**: Domain, type, complexity matching
- **Approach Analysis**: Successful vs failed strategies
- **Tool Sequences**: Common patterns in high-performing trajectories
- **Failure Detection**: Approaches that historically fail

### 3. Adaptive Recommendations

**Recommendation Generation**:
```
When starting new problem:
├── Find similar past trajectories (similarity > 0.3)
├── Check learned patterns for problem type
├── Analyze successful approaches
├── Generate prioritized recommendations
└── Include warnings about failed approaches
```

**Recommendation Types**:
- **Tool Sequence**: "Consider using: [think, assess-evidence, make-decision]"
- **Approach**: "Similar problems solved well with parallel-exploration strategy"
- **Warning**: "Avoid systematic-linear approach - only 20% success rate for this problem type"
- **Optimization**: "Switch to tree mode at step 5 for better branching exploration"

### 4. Quality Feedback Loop

**Metrics Tracked**:
- **Efficiency**: Steps taken vs optimal path (fewer steps = more efficient)
- **Coherence**: Logical consistency (from contradiction detection)
- **Completeness**: Goal coverage (% of goals achieved)
- **Innovation**: Use of creative/advanced tools
- **Reliability**: Confidence in final result
- **Bias/Fallacy Counts**: Reasoning quality indicators

**Success Scoring**:
```
Success Score = (Outcome Status × 0.6) + (Quality × 0.4)
                × (0.7 + Confidence × 0.3)

Where:
- success → 0.9 base
- partial → 0.6 base
- failure → 0.2 base
```

## API Reference

### MCP Tools

#### start-reasoning-session
Starts tracking a new reasoning session.

**Parameters**:
- `session_id` (required): Unique session identifier
- `description` (required): Problem description
- `goals` (optional): List of goals to achieve
- `domain` (optional): Problem domain (e.g., "software-engineering", "science")
- `context` (optional): Additional context
- `complexity` (optional): Estimated complexity (0.0-1.0)

**Returns**:
- `session_id`: Session identifier
- `problem_id`: Problem fingerprint hash
- `status`: "active"
- `suggestions`: Array of recommendations based on similar past problems

**Example**:
```json
{
  "session_id": "session_123",
  "description": "Optimize database query performance",
  "goals": ["Reduce query time", "Improve user experience"],
  "domain": "software-engineering",
  "complexity": 0.6
}
```

#### complete-reasoning-session
Marks a session as complete and stores the trajectory.

**Parameters**:
- `session_id` (required): Session to complete
- `status` (required): "success", "partial", or "failure"
- `goals_achieved` (optional): List of achieved goals
- `goals_failed` (optional): List of failed goals
- `solution` (optional): Description of solution
- `confidence` (optional): Confidence in solution (0.0-1.0)
- `unexpected_outcomes` (optional): Unexpected results

**Returns**:
- `trajectory_id`: Stored trajectory ID
- `success_score`: Calculated success score (0.0-1.0)
- `quality_score`: Overall quality score (0.0-1.0)
- `status`: "completed"

**Example**:
```json
{
  "session_id": "session_123",
  "status": "success",
  "goals_achieved": ["Reduce query time"],
  "solution": "Added indexes and query optimization",
  "confidence": 0.85
}
```

#### get-recommendations
Get recommendations for a problem based on episodic memory.

**Parameters**:
- `description` (required): Problem description
- `goals` (optional): Problem goals
- `domain` (optional): Problem domain
- `context` (optional): Additional context
- `complexity` (optional): Estimated complexity
- `limit` (optional): Max recommendations (default: 5)

**Returns**:
- `recommendations`: Array of recommendations
- `similar_cases`: Count of similar past trajectories
- `learned_patterns`: Applicable learned patterns
- `count`: Number of recommendations

**Example**:
```json
{
  "description": "Need to implement user authentication",
  "domain": "security",
  "limit": 3
}
```

#### search-trajectories
Search for past reasoning trajectories.

**Parameters**:
- `domain` (optional): Filter by domain
- `tags` (optional): Filter by tags
- `min_success` (optional): Minimum success score
- `problem_type` (optional): Filter by problem type
- `limit` (optional): Max results (default: 10)

**Returns**:
- `trajectories`: Array of trajectory summaries
- `count`: Number of results

## Integration with Existing Systems

### Case-Based Reasoning (CBR)
The episodic memory system **extends** the existing CBR module:
- CBR stores solution cases
- Episodic memory stores complete reasoning trajectories
- Episodic memory can populate CBR with successful solutions
- Both use similarity-based retrieval

### Workflow Orchestration
Episodic memory informs workflow selection:
- Learned patterns suggest optimal tool sequences
- Successful workflows are identified and recommended
- Failed workflows trigger warnings

### Self-Evaluation & Metacognition
Quality metrics feed into pattern learning:
- Self-evaluation scores influence success scoring
- Bias detection counts affect quality metrics
- Metacognitive insights extracted as tags

## Use Cases

### 1. "Similar Problem Solved Before"
```
User: "Optimize database performance"
System retrieves: 3 similar trajectories
Recommendation: "Previous cases succeeded with:
  1. assess-evidence (query analysis)
  2. build-causal-graph (bottleneck identification)
  3. make-decision (optimization strategy selection)
  Success rate: 87%"
```

### 2. "Avoid Known Failure Patterns"
```
User: "Design microservices architecture"
System detects: 5 similar trajectories with systematic-linear approach
Warning: "Avoid purely linear approach - only 20% success rate.
  Recommend: multi-branch-creative strategy (75% success rate)"
```

### 3. "Continuous Improvement"
```
After 10 trajectories in "software-engineering" domain:
System learns: Tree mode + abductive-reasoning → 80% success
              Linear mode alone → 40% success
Auto-suggests tree mode for future software problems
```

### 4. "Cross-Domain Learning"
```
User: "Protect computer network"
System finds: Successful analogy from biology domain
Recommendation: "Similar problem solved using find-analogy
  from 'immune system' domain - 90% success rate"
```

## Performance Characteristics

### Storage
- **In-Memory**: Fast retrieval, no persistence
- **SQLite** (future): Persistent across restarts, slower but durable
- **Indexes**: O(1) lookup by problem hash, domain, tags

### Pattern Learning
- **Batch Processing**: Runs every N trajectories or time interval
- **Incremental**: New patterns added without reprocessing all data
- **Threshold**: Minimum 3 trajectories to form a pattern

### Recommendation Generation
- **Real-time**: Generated on session start (< 100ms)
- **Similarity Threshold**: 0.3 minimum (configurable)
- **Top-K**: Returns top 5 recommendations by relevance

## Future Enhancements

### Phase 2.1 (Near-term)
- [ ] SQLite persistence for trajectories
- [ ] Retrospective analysis tool for reviewing past sessions
- [ ] Pattern visualization and inspection tools
- [ ] Export trajectories for external analysis

### Phase 2.2 (Medium-term)
- [ ] Cross-session learning (learn from other users/agents)
- [ ] Confidence calibration based on historical accuracy
- [ ] Automated A/B testing of approaches
- [ ] "Reasoning coach" mode with real-time feedback

### Phase 2.3 (Long-term)
- [ ] Multi-agent collaboration patterns
- [ ] Transfer learning across domains
- [ ] Neural pattern recognition for complex similarities
- [ ] Automated workflow generation from successful trajectories

## Research Foundation

Based on 2025 AI agent research:
- **Episodic Memory**: ["Episodic memory is the missing piece for long-term LLM agents"](https://arxiv.org/html/2504.06943v2)
- **Case-Based Reasoning**: Psychologically grounded - humans solve problems by recalling analogous situations
- **Memory-Augmented Learning**: Enables continuous improvement without fine-tuning
- **Context Continuity**: Aligns with MCP's core goal of maintaining reasoning across interactions

## Implementation Status

### Completed (Phase 2 Initial)
- ✅ Core data structures (ReasoningTrajectory, TrajectoryPattern, etc.)
- ✅ EpisodicMemoryStore with multi-index system
- ✅ SessionTracker for active session monitoring
- ✅ LearningEngine with pattern recognition
- ✅ Similarity-based retrieval
- ✅ Quality metrics calculation
- ✅ Success scoring algorithm
- ✅ MCP handler functions (4 tools)

### In Progress
- 🔄 Server integration and tool registration
- 🔄 Automatic session tracking from existing tools

### Pending
- ⏳ Comprehensive test suite
- ⏳ SQLite persistence implementation
- ⏳ Documentation updates
- ⏳ Example usage demonstrations

## Configuration

### Environment Variables
```bash
# Enable episodic memory
EPISODIC_MEMORY_ENABLED=true

# Storage backend
EPISODIC_MEMORY_STORAGE=memory  # or "sqlite"

# SQLite database path (if using sqlite)
EPISODIC_MEMORY_DB_PATH=./data/episodic.db

# Learning configuration
EPISODIC_MIN_TRAJECTORIES=3      # Min trajectories to form pattern
EPISODIC_MIN_SUCCESS_RATE=0.6    # Min success rate for patterns
EPISODIC_LEARNING_INTERVAL=3600  # Learning cycle interval (seconds)

# Similarity threshold
EPISODIC_MIN_SIMILARITY=0.3      # Min similarity for recommendations
```

## Conclusion

The Episodic Reasoning Memory System represents a fundamental architectural shift from **stateless tools** to **stateful learning**. By learning from past reasoning sessions, the unified-thinking server becomes progressively more intelligent, providing:

1. **Proactive Guidance**: Suggests proven approaches before the user gets stuck
2. **Failure Prevention**: Warns about historically unsuccessful strategies
3. **Continuous Improvement**: Gets smarter with every reasoning session
4. **Cross-Session Learning**: Applies lessons from one problem to similar future problems
5. **Pattern Recognition**: Identifies and codifies successful reasoning strategies

This positions unified-thinking as a **genuinely novel MCP server** that learns and improves over time, differentiating it from all other stateless MCP implementations.
