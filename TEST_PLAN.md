# Test Plan for Unified Thinking MCP Server

This document provides a comprehensive test plan for validating all functionality of the unified-thinking MCP server in Claude Desktop.

## Prerequisites

1. Server built and configured in Claude Desktop config
2. Claude Desktop restarted after configuration
3. Server appearing in available MCP tools

## Test Categories

### 1. Basic Thinking Modes

#### 1.1 Linear Mode
**Objective:** Test sequential step-by-step reasoning

**Test Case:**
```
Use the think tool with:
{
  "content": "Solve this problem step by step: If a train travels 60 mph for 2.5 hours, how far does it go?",
  "mode": "linear"
}
```

**Expected Result:**
- Thought created with mode "linear"
- Sequential reasoning steps
- Final answer: 150 miles

#### 1.2 Tree Mode
**Objective:** Test multi-branch parallel exploration

**Test Case 1 - Create branches:**
```
Use the think tool with:
{
  "content": "What are different approaches to solving climate change?",
  "mode": "tree"
}
```

**Expected Result:**
- Multiple branches created
- Each branch explores different approach
- Insights generated for promising paths

**Test Case 2 - List branches:**
```
Use the list-branches tool
```

**Expected Result:**
- Shows all active branches
- Displays branch IDs, states, priorities
- Shows thought counts per branch

**Test Case 3 - Focus on specific branch:**
```
Use the focus-branch tool with:
{
  "branch_id": "<branch-id-from-list>"
}
```

**Expected Result:**
- Active branch switched
- Confirmation message with branch details

**Test Case 4 - Get branch history:**
```
Use the branch-history tool with:
{
  "branch_id": "<branch-id>"
}
```

**Expected Result:**
- Full thought history for branch
- Insights and cross-references
- Timestamps and metadata

#### 1.3 Divergent Mode
**Objective:** Test creative/unconventional thinking

**Test Case:**
```
Use the think tool with:
{
  "content": "What's an unconventional solution to reduce traffic congestion?",
  "mode": "divergent"
}
```

**Expected Result:**
- Creative, non-standard ideas
- May include "rebellion" thoughts
- Challenges common assumptions

#### 1.4 Auto Mode
**Objective:** Test automatic mode selection

**Test Case 1 - Should select linear:**
```
Use the think tool with:
{
  "content": "Calculate the compound interest on $1000 at 5% for 3 years",
  "mode": "auto"
}
```

**Expected Result:**
- Mode automatically set to "linear"
- Step-by-step calculation

**Test Case 2 - Should select tree:**
```
Use the think tool with:
{
  "content": "Explore multiple architectural patterns for a microservices system",
  "mode": "auto"
}
```

**Expected Result:**
- Mode automatically set to "tree"
- Multiple branches created

**Test Case 3 - Should select divergent:**
```
Use the think tool with:
{
  "content": "Think outside the box: how could we reinvent education?",
  "mode": "auto"
}
```

**Expected Result:**
- Mode automatically set to "divergent"
- Creative, unconventional ideas

### 2. Logical Validation

#### 2.1 Validate Thought
**Objective:** Test logical consistency checking

**Test Case 1 - Valid thought:**
```
First create a thought, then use the validate tool with:
{
  "thought_id": "<thought-id-from-response>"
}
```

**Expected Result:**
- Validation result with is_valid: true
- Reason explaining why it passes

**Test Case 2 - Invalid thought with contradiction:**
```
Create thought with contradictory content:
{
  "content": "This system always works but never functions properly",
  "mode": "linear"
}

Then validate it.
```

**Expected Result:**
- Validation result with is_valid: false
- Reason identifying the "always/never" contradiction

#### 2.2 Prove Logical Conclusion
**Objective:** Test formal proof attempts

**Test Case 1 - Provable syllogism:**
```
Use the prove tool with:
{
  "premises": ["All humans are mortal", "Socrates is human"],
  "conclusion": "Socrates is mortal"
}
```

**Expected Result:**
- is_provable: true
- Proof steps showing logical derivation

**Test Case 2 - Unprovable conclusion:**
```
Use the prove tool with:
  "premises": ["Some birds can fly"],
  "conclusion": "All birds can fly"
}
```

**Expected Result:**
- is_provable: false
- Explanation of why conclusion doesn't follow

#### 2.3 Check Syntax
**Objective:** Test logical statement syntax validation

**Test Case:**
```
Use the check-syntax tool with:
{
  "statements": [
    "All men are mortal",
    "Invalid",
    "",
    "If (A then B",
    "This is a valid statement"
  ]
}
```

**Expected Result:**
- Array of validation results
- First statement: well-formed
- Second statement: not well-formed (single word)
- Third statement: not well-formed (empty)
- Fourth statement: not well-formed (unbalanced parentheses)
- Fifth statement: well-formed

### 3. History and Search

#### 3.1 View History
**Objective:** Test thought history retrieval

**Test Case 1 - All history:**
```
Use the history tool with no parameters
```

**Expected Result:**
- All thoughts across all modes
- Ordered by timestamp
- Includes metadata

**Test Case 2 - Mode-filtered history:**
```
Use the history tool with:
{
  "mode": "linear"
}
```

**Expected Result:**
- Only linear mode thoughts
- Ordered by timestamp

**Test Case 3 - Branch-specific history:**
```
Use the history tool with:
{
  "mode": "tree",
  "branch_id": "<branch-id>"
}
```

**Expected Result:**
- Only thoughts from specified branch
- Ordered by timestamp

#### 3.2 Search Thoughts
**Objective:** Test content search functionality

**Test Case 1 - Simple search:**
```
Use the search tool with:
{
  "query": "climate"
}
```

**Expected Result:**
- All thoughts containing "climate"
- Relevance-ranked results
- Includes thought IDs and snippets

**Test Case 2 - Mode-filtered search:**
```
Use the search tool with:
{
  "query": "solution",
  "mode": "divergent"
}
```

**Expected Result:**
- Only divergent mode thoughts containing "solution"

### 4. Branch Management

#### 4.1 Recent Branches
**Objective:** Test recent branch access tracking

**Test Case:**
```
1. Create several branches in tree mode
2. Focus on different branches in sequence
3. Use the recent-branches tool
```

**Expected Result:**
- Last 10 accessed branches
- Ordered by most recent access
- Shows access timestamps
- Indicates current active branch

### 5. System Metrics

#### 5.1 Get Metrics
**Objective:** Test performance and usage statistics

**Test Case:**
```
After performing several operations, use the get-metrics tool
```

**Expected Result:**
- Total thought counts by mode
- Branch statistics (active, suspended, completed)
- Validation statistics
- Search performance metrics
- Memory usage information

### 6. Advanced Options

#### 6.1 Confidence Levels
**Objective:** Test confidence scoring

**Test Case:**
```
Use the think tool with:
{
  "content": "Based on available evidence, climate change is occurring",
  "mode": "linear",
  "confidence": 0.95
}
```

**Expected Result:**
- Thought created with confidence: 0.95
- Higher confidence affects branch priority in tree mode

#### 6.2 Key Points Extraction
**Objective:** Test automatic key point identification

**Test Case:**
```
Use the think tool with:
{
  "content": "Three factors contribute to success: persistence, adaptability, and learning from failure",
  "mode": "linear",
  "key_points": ["persistence", "adaptability", "learning from failure"]
}
```

**Expected Result:**
- Thought stores key_points array
- Key points accessible in history/search

#### 6.3 Validation Requirements
**Objective:** Test automatic validation on thought creation

**Test Case:**
```
Use the think tool with:
{
  "content": "Testing validation requirement",
  "mode": "linear",
  "require_validation": true
}
```

**Expected Result:**
- Thought created and automatically validated
- Validation result included in response

### 7. Cognitive Reasoning Features

Note: These features are implemented in the codebase but require MCP tool integration in server.go to be accessible. Test these once the tools are added.

#### 7.1 Probabilistic Reasoning
**Planned Tool:** probabilistic-reasoning

**Future Test:**
- Create Bayesian belief
- Update belief with new evidence
- Combine multiple beliefs
- Verify probability calculations

#### 7.2 Evidence Assessment
**Planned Tool:** assess-evidence

**Future Test:**
- Submit evidence for quality assessment
- Verify automatic quality classification
- Test reliability and relevance scoring
- Aggregate multiple evidence pieces

#### 7.3 Contradiction Detection
**Planned Tool:** detect-contradictions

**Future Test:**
- Create thoughts with contradictions
- Detect direct negations
- Detect contradictory absolutes
- Verify severity classification

#### 7.4 Decision Making
**Planned Tool:** make-decision

**Future Test:**
- Create decision with multiple options
- Define weighted criteria
- Verify automatic scoring
- Check recommendation accuracy

#### 7.5 Problem Decomposition
**Planned Tool:** decompose-problem

**Future Test:**
- Submit complex problem
- Verify subproblem breakdown
- Check dependency mapping
- Validate solution path ordering

#### 7.6 Sensitivity Analysis
**Planned Tool:** sensitivity-analysis

**Future Test:**
- Test assumption variations
- Calculate robustness scores
- Identify key assumptions
- Measure impact magnitudes

#### 7.7 Self-Evaluation
**Planned Tool:** self-evaluate

**Future Test:**
- Evaluate thought quality
- Check completeness scoring
- Verify coherence assessment
- Review improvement suggestions

#### 7.8 Bias Detection
**Planned Tool:** detect-biases

**Future Test:**
- Submit thoughts for bias analysis
- Detect confirmation bias
- Detect anchoring bias
- Verify mitigation strategies

## Error Handling Tests

### 8.1 Invalid Inputs

**Test Case 1 - Invalid mode:**
```
Use the think tool with:
{
  "content": "Test",
  "mode": "invalid_mode"
}
```

**Expected Result:**
- Error message indicating invalid mode
- List of valid modes

**Test Case 2 - Missing required field:**
```
Use the think tool with:
{
  "mode": "linear"
}
```

**Expected Result:**
- Error message indicating missing content

**Test Case 3 - Invalid branch ID:**
```
Use the focus-branch tool with:
{
  "branch_id": "nonexistent-branch"
}
```

**Expected Result:**
- Error message indicating branch not found

**Test Case 4 - Invalid thought ID for validation:**
```
Use the validate tool with:
{
  "thought_id": "nonexistent-thought"
}
```

**Expected Result:**
- Error message indicating thought not found

### 8.2 Edge Cases

**Test Case 1 - Empty content:**
```
Use the think tool with:
{
  "content": "",
  "mode": "linear"
}
```

**Expected Result:**
- Error or thought created with empty content flag

**Test Case 2 - Very long content:**
```
Use the think tool with content exceeding 10,000 characters
```

**Expected Result:**
- Thought created successfully or appropriate truncation

**Test Case 3 - Special characters:**
```
Use the think tool with:
{
  "content": "Test with special chars: <>&\"'",
  "mode": "linear"
}
```

**Expected Result:**
- Content properly escaped and stored

## Performance Tests

### 9.1 Load Testing

**Test Case 1 - Many thoughts:**
```
Create 100+ thoughts in rapid succession
```

**Expected Result:**
- All thoughts created successfully
- Reasonable response times
- No memory leaks

**Test Case 2 - Many branches:**
```
Create 50+ branches in tree mode
```

**Expected Result:**
- All branches managed correctly
- list-branches returns all branches
- No performance degradation

**Test Case 3 - Large history:**
```
After creating many thoughts, use history tool
```

**Expected Result:**
- History returns in reasonable time
- Results properly paginated or limited
- No timeouts

### 9.2 Concurrent Operations

**Test Case:**
```
Perform multiple operations simultaneously:
- Create thoughts in different modes
- Search while creating
- Validate while exploring branches
```

**Expected Result:**
- All operations complete successfully
- No race conditions
- Thread-safe behavior

## Integration Tests

### 10.1 Full Workflow Tests

**Test Case 1 - Research workflow:**
```
1. Use auto mode to explore a topic
2. Switch to tree mode for deeper exploration
3. Create multiple branches
4. Use recent-branches to navigate
5. Validate key conclusions
6. Search across all findings
7. Get metrics on the research session
```

**Expected Result:**
- Smooth workflow with mode transitions
- All data accessible and consistent
- Metrics accurately reflect activity

**Test Case 2 - Problem-solving workflow:**
```
1. Use linear mode to break down problem
2. Switch to divergent for creative solutions
3. Use validation on proposed solutions
4. Prove logical connections
5. Review history of solution development
```

**Expected Result:**
- Each mode contributes appropriately
- Logical validation catches errors
- Complete audit trail in history

## Regression Tests

After any code changes, verify:

1. All existing thoughts remain accessible
2. Branch relationships intact
3. Search indexes functional
4. Validation logic unchanged (unless intentionally modified)
5. History ordering preserved
6. Metrics calculations accurate

## Test Execution Checklist

- [ ] All basic thinking modes tested
- [ ] Logical validation working
- [ ] History and search functional
- [ ] Branch management operational
- [ ] System metrics accurate
- [ ] Error handling appropriate
- [ ] Edge cases handled
- [ ] Performance acceptable
- [ ] Integration workflows smooth
- [ ] No regressions introduced

## Notes for Testers

1. Keep track of thought IDs and branch IDs for cross-referencing tests
2. Clear the server state between major test sections if needed (restart Claude Desktop)
3. Document any unexpected behavior or errors
4. Verify responses match expected JSON structure
5. Check that all fields are properly populated
6. Test both with and without optional parameters

## Future Test Additions

When cognitive reasoning tools are integrated:
- Add full test suite for probabilistic reasoning
- Test evidence assessment workflows
- Validate contradiction detection accuracy
- Test decision framework with real scenarios
- Verify metacognition features
- Test all bias detection patterns
