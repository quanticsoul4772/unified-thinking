# Claude Desktop Test Plan: Cognitive Reasoning Tools

**Version**: 1.0.0
**Date**: 2025-10-01
**Purpose**: Verify all 19 MCP tools work correctly in Claude Desktop
**Expected Duration**: 30-45 minutes

---

## Prerequisites

### 1. Install the MCP Server

Ensure the server is configured in Claude Desktop:

**File**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "unified-thinking": {
      "command": "C:/Development/Projects/MCP/project-root/mcp-servers/unified-thinking/bin/unified-thinking.exe",
      "transport": "stdio",
      "env": {
        "DEBUG": "true"
      }
    }
  }
}
```

### 2. Restart Claude Desktop

After updating the config, **completely quit and restart Claude Desktop**.

### 3. Verify Server Connection

In a new conversation, ask:
```
What MCP tools are available from unified-thinking?
```

**Expected**: Claude should list all 19 tools.

---

## Test Plan Structure

Each test includes:
- **Test ID**: Unique identifier
- **Tool**: Tool being tested
- **Prompt**: What to say to Claude
- **Expected Result**: What should happen
- **Pass Criteria**: How to verify success

---

## Core Tools Tests (Tools 1-11)

### Test 1.1: Basic Linear Thinking

**Tool**: `think`
**Prompt**:
```
Use the think tool in linear mode to analyze: "What are the benefits of remote work?"
```

**Expected Result**:
- Thought created with step-by-step reasoning
- Thought ID returned (e.g., `thought-xxx`)
- Mode should be "linear"

**Pass Criteria**: ✅ Thought created successfully with sequential reasoning

---

### Test 1.2: Tree Mode with Branches

**Tool**: `think`
**Prompt**:
```
Use the think tool in tree mode to explore: "Different approaches to learning a new programming language"
```

**Expected Result**:
- Thought created in tree mode
- Branch ID created (e.g., `branch-xxx`)
- Multiple perspectives or approaches listed

**Pass Criteria**: ✅ Branch created successfully with multi-perspective analysis

---

### Test 1.3: Divergent Mode

**Tool**: `think`
**Prompt**:
```
Use the think tool in divergent mode to generate creative ideas: "Unconventional uses for a coffee mug"
```

**Expected Result**:
- Thought created with creative/unusual ideas
- Should include non-obvious suggestions
- Mode should be "divergent"

**Pass Criteria**: ✅ Creative, unconventional ideas generated

---

### Test 1.4: Auto Mode Detection

**Tool**: `think`
**Prompt**:
```
Use the think tool in auto mode: "What if we could photosynthesize like plants?"
```

**Expected Result**:
- Auto mode should detect divergent thinking needed
- Creative exploration of the concept
- Mode automatically selected

**Pass Criteria**: ✅ Auto mode correctly selects appropriate thinking style

---

### Test 1.5: View History

**Tool**: `history`
**Prompt**:
```
Show me the thinking history from linear mode
```

**Expected Result**:
- List of all linear mode thoughts
- Includes thought IDs, content previews, timestamps
- At least Test 1.1 thought visible

**Pass Criteria**: ✅ History retrieved with all expected thoughts

---

### Test 1.6: List Branches

**Tool**: `list-branches`
**Prompt**:
```
List all thinking branches
```

**Expected Result**:
- List of all branches created
- Shows branch IDs, names, thought counts
- At least Test 1.2 branch visible

**Pass Criteria**: ✅ All branches listed correctly

---

### Test 1.7: Focus Branch

**Tool**: `focus-branch`
**Prompt**:
```
Focus on the branch created in Test 1.2 (use the branch ID from Test 1.6)
```

**Expected Result**:
- Active branch switched successfully
- Confirmation message with branch details

**Pass Criteria**: ✅ Branch focus changed without errors

---

### Test 1.8: Branch History

**Tool**: `branch-history`
**Prompt**:
```
Show the history of the currently focused branch
```

**Expected Result**:
- Detailed history of all thoughts in that branch
- Shows insights, cross-references if any
- Timeline of branch development

**Pass Criteria**: ✅ Branch history displayed with complete details

---

### Test 1.9: Validate Thought Logic

**Tool**: `validate`
**Prompt**:
```
Use the think tool to create a thought: "All birds can fly, but penguins cannot fly, therefore penguins are birds"

Then validate that thought for logical consistency.
```

**Expected Result**:
- Validation detects the logical issue
- Reports contradictions or inconsistencies
- Provides explanation

**Pass Criteria**: ✅ Validation identifies the contradiction correctly

---

### Test 1.10: Prove Logical Conclusion

**Tool**: `prove`
**Prompt**:
```
Use the prove tool with these premises:
- "All developers write code"
- "Sarah is a developer"

Conclusion: "Sarah writes code"
```

**Expected Result**:
- Proof result shows "valid" or "likely valid"
- Explanation of the logical inference
- Confidence score provided

**Pass Criteria**: ✅ Valid syllogism correctly identified

---

### Test 1.11: Check Syntax

**Tool**: `check-syntax`
**Prompt**:
```
Check the syntax of these statements:
- "If (P and Q) then R"
- "((Unbalanced parentheses"
- "All X are Y"
```

**Expected Result**:
- First statement: valid
- Second statement: unbalanced parentheses detected
- Third statement: valid

**Pass Criteria**: ✅ Syntax errors correctly identified

---

### Test 1.12: Search Thoughts

**Tool**: `search`
**Prompt**:
```
Search for thoughts containing the word "creative"
```

**Expected Result**:
- Returns Test 1.3 thought (divergent mode test)
- May return other thoughts with "creative" in content
- Shows thought IDs, excerpts, modes

**Pass Criteria**: ✅ Search finds relevant thoughts

---

### Test 1.13: Get Metrics

**Tool**: `get-metrics`
**Prompt**:
```
Show me the system metrics
```

**Expected Result**:
- Total thought count (at least 12 from tests so far)
- Branch count (at least 1)
- Thoughts by mode breakdown
- Performance metrics

**Pass Criteria**: ✅ Metrics accurately reflect activity

---

### Test 1.14: Recent Branches

**Tool**: `recent-branches`
**Prompt**:
```
Show me the recently accessed branches
```

**Expected Result**:
- List of up to 10 most recent branches
- Shows timestamps of last access
- Includes branch from Test 1.2 and Test 1.7

**Pass Criteria**: ✅ Recent branches listed with timestamps

---

## Cognitive Reasoning Tools Tests (Tools 12-19)

### Test 2.1: Probabilistic Reasoning - Create Belief

**Tool**: `probabilistic-reasoning`
**Prompt**:
```
Use probabilistic-reasoning to create a belief:
- Statement: "It will rain tomorrow"
- Prior probability: 0.3
```

**Expected Result**:
- Belief created with ID (e.g., `belief-xxx`)
- Prior probability set to 0.3
- Posterior initially equals prior

**Pass Criteria**: ✅ Belief created with correct probability

---

### Test 2.2: Probabilistic Reasoning - Update Belief

**Tool**: `probabilistic-reasoning`
**Prompt**:
```
Update the belief from Test 2.1 with new evidence:
- Operation: update
- Belief ID: [use ID from Test 2.1]
- Evidence: "Dark clouds are forming"
- Likelihood: 0.8
- Evidence probability: 0.4
```

**Expected Result**:
- Belief updated using Bayesian inference
- Posterior probability calculated (should be higher than 0.3)
- Shows prior → posterior transition

**Pass Criteria**: ✅ Posterior probability > prior probability

---

### Test 2.3: Probabilistic Reasoning - Combine Beliefs

**Tool**: `probabilistic-reasoning`
**Prompt**:
```
Create two beliefs:
1. "It will be cold tomorrow" (prior: 0.6)
2. "It will be windy tomorrow" (prior: 0.4)

Then combine them with AND operation
```

**Expected Result**:
- Two beliefs created
- Combined belief calculated
- AND operation gives P(A) × P(B) = 0.24

**Pass Criteria**: ✅ Combined probability = 0.24

---

### Test 2.4: Assess Evidence - Strong Evidence

**Tool**: `assess-evidence`
**Prompt**:
```
Assess this evidence:
- Content: "Peer-reviewed study of 10,000 participants published in Nature shows 95% correlation"
- Source: "Nature Journal"
- Claim: "Exercise improves mental health"
- Supports claim: true
```

**Expected Result**:
- Quality: "Strong" or "Moderate"
- Reliability score: 0.8-1.0
- Relevance score: 0.8-1.0
- Explanation of quality assessment

**Pass Criteria**: ✅ Classified as high-quality evidence

---

### Test 2.5: Assess Evidence - Weak Evidence

**Tool**: `assess-evidence`
**Prompt**:
```
Assess this evidence:
- Content: "My friend told me this works"
- Source: "Personal anecdote"
- Claim: "Exercise improves mental health"
- Supports claim: true
```

**Expected Result**:
- Quality: "Weak" or "Anecdotal"
- Reliability score: 0.2-0.4
- Relevance score: lower than Test 2.4
- Explanation noting anecdotal nature

**Pass Criteria**: ✅ Classified as low-quality evidence

---

### Test 2.6: Detect Contradictions

**Tool**: `detect-contradictions`
**Prompt**:
```
Create two thoughts:
1. "All remote work is always more productive"
2. "Remote work never increases productivity"

Then use detect-contradictions to find conflicts between them
```

**Expected Result**:
- Contradiction detected
- Type: "negation" (always vs never)
- Explanation of the conflict
- Severity assessment

**Pass Criteria**: ✅ Contradiction identified correctly

---

### Test 2.7: Make Decision - MCDA

**Tool**: `make-decision`
**Prompt**:
```
Help me decide between programming languages for a new project:

Options:
1. Python - Cost: 0.9, Speed: 0.5, Ecosystem: 0.9
2. Rust - Cost: 0.4, Speed: 0.9, Ecosystem: 0.6
3. Go - Cost: 0.7, Speed: 0.8, Ecosystem: 0.7

Criteria:
- Cost (weight: 0.3, minimize: false - higher score = lower cost)
- Speed (weight: 0.5, maximize: true)
- Ecosystem (weight: 0.2, maximize: true)
```

**Expected Result**:
- Weighted scores calculated for each option
- Ranking of options (likely: Rust > Go > Python)
- Justification for recommendation
- Sensitivity noted

**Pass Criteria**: ✅ Decision analysis completed with ranking

---

### Test 2.8: Decompose Problem

**Tool**: `decompose-problem`
**Prompt**:
```
Decompose this problem: "Build a production-ready web application"
```

**Expected Result**:
- Problem broken into subproblems:
  - Frontend development
  - Backend development
  - Database design
  - Testing
  - Deployment
  - Monitoring
- Dependencies identified
- Complexity estimated for each

**Pass Criteria**: ✅ Problem decomposed into logical subproblems with dependencies

---

### Test 2.9: Sensitivity Analysis

**Tool**: `sensitivity-analysis`
**Prompt**:
```
Perform sensitivity analysis:
- Target claim: "Remote work increases productivity by 20%"
- Assumptions:
  - "Workers have good internet"
  - "Workers have dedicated workspace"
  - "Communication tools are effective"
- Base confidence: 0.7
```

**Expected Result**:
- Each assumption tested
- Impact scores showing how critical each is
- Recommendations for most critical assumptions
- Confidence range calculated

**Pass Criteria**: ✅ Sensitivity analysis shows impact of each assumption

---

### Test 2.10: Self-Evaluate Thought Quality

**Tool**: `self-evaluate`
**Prompt**:
```
Create a thought: "The solution is obvious. Everyone knows that X is always better than Y."

Then use self-evaluate to assess this thought.
```

**Expected Result**:
- Quality score (likely low due to absolutes)
- Completeness assessment
- Coherence score
- Strengths identified (if any)
- Weaknesses identified:
  - Lacks evidence
  - Uses absolutes ("always")
  - Vague ("obvious", "everyone knows")
- Improvement suggestions

**Pass Criteria**: ✅ Evaluation identifies weakness (absolutes, lack of evidence)

---

### Test 2.11: Detect Biases - Confirmation Bias

**Tool**: `detect-biases`
**Prompt**:
```
Create a thought: "I only looked at data that supports my hypothesis and ignored the rest. This proves I was right."

Then detect biases in this thought.
```

**Expected Result**:
- Confirmation bias detected
- Severity: "High" or "Critical"
- Explanation of the bias
- Mitigation strategies suggested
- Confidence score (high)

**Pass Criteria**: ✅ Confirmation bias correctly identified

---

### Test 2.12: Detect Biases - Overconfidence Bias

**Tool**: `detect-biases`
**Prompt**:
```
Create a thought: "I'm 100% certain this will work. There's absolutely no chance of failure."

Then detect biases in this thought.
```

**Expected Result**:
- Overconfidence bias detected
- Severity assessment
- Evidence: absolutes ("100%", "absolutely no chance")
- Mitigation: suggest probability ranges

**Pass Criteria**: ✅ Overconfidence bias correctly identified

---

### Test 2.13: Detect Biases - Sunk Cost Fallacy

**Tool**: `detect-biases`
**Prompt**:
```
Create a thought: "We've already invested $50,000 in this project, so we must continue even though it's clearly failing."

Then detect biases in this thought.
```

**Expected Result**:
- Sunk cost fallacy detected
- Severity assessment
- Explanation of the fallacy
- Mitigation: focus on future value, not past costs

**Pass Criteria**: ✅ Sunk cost fallacy correctly identified

---

## Integration Tests (Cross-Tool)

### Test 3.1: Full Reasoning Workflow

**Tools**: Multiple
**Prompt**:
```
Let's solve a complex problem using multiple cognitive tools:

1. Use decompose-problem to break down: "Should our company adopt a 4-day work week?"

2. Use assess-evidence to evaluate:
   - Evidence A: "Iceland trial showed 86% of workers happier"
   - Evidence B: "My boss thinks it won't work"

3. Use probabilistic-reasoning to create belief: "4-day week improves productivity" (prior: 0.5)

4. Use make-decision with:
   Options: Adopt 4-day week, Keep 5-day week
   Criteria: Employee satisfaction (0.4), Productivity (0.4), Cost (0.2)

5. Use detect-biases on the final decision reasoning

6. Use self-evaluate on the entire analysis
```

**Expected Result**:
- All 6 tools work in sequence
- Each tool uses output from previous steps
- Final decision is well-reasoned
- Biases detected if present
- Self-evaluation provides meta-assessment

**Pass Criteria**: ✅ All tools work together coherently

---

### Test 3.2: Tree Mode + Contradiction Detection

**Tools**: `think`, `detect-contradictions`
**Prompt**:
```
1. Create two thoughts in tree mode on different branches:
   Branch A: "AI will replace all programming jobs"
   Branch B: "AI will augment programmers, not replace them"

2. Use detect-contradictions to analyze these branches
```

**Expected Result**:
- Two branches created
- Contradiction detected between branches
- Type identified (direct opposition)
- Explanation provided

**Pass Criteria**: ✅ Contradiction detected across branches

---

### Test 3.3: Evidence + Belief Update

**Tools**: `assess-evidence`, `probabilistic-reasoning`
**Prompt**:
```
1. Create belief: "Coffee improves focus" (prior: 0.6)

2. Assess evidence: "Meta-analysis of 50 studies shows moderate effect"

3. Update the belief based on the evidence assessment
```

**Expected Result**:
- Belief created
- Evidence assessed (likely Strong/Moderate quality)
- Belief updated with higher posterior
- Logical flow from evidence to belief

**Pass Criteria**: ✅ Evidence quality influences belief update appropriately

---

## Error Handling Tests

### Test 4.1: Invalid Probability Range

**Tool**: `probabilistic-reasoning`
**Prompt**:
```
Create a belief with prior probability 1.5
```

**Expected Result**:
- Error: "probability must be between 0.0 and 1.0"
- Helpful error message
- No crash

**Pass Criteria**: ✅ Validation error caught gracefully

---

### Test 4.2: Non-Existent Branch

**Tool**: `focus-branch`
**Prompt**:
```
Focus on branch "branch-doesnotexist"
```

**Expected Result**:
- Error: "branch not found: branch-doesnotexist"
- **Important**: Should show available branches
- Example: "available branches: [branch-xxx, branch-yyy]"

**Pass Criteria**: ✅ Error message includes list of available branches

---

### Test 4.3: Invalid Tool Parameters

**Tool**: `make-decision`
**Prompt**:
```
Use make-decision with no options
```

**Expected Result**:
- Error: "options required" or similar
- Helpful message about what's needed
- No crash

**Pass Criteria**: ✅ Validation error with helpful message

---

## Performance Tests

### Test 5.1: Large History

**Tool**: `think`, `history`
**Prompt**:
```
Create 20 thoughts in linear mode with various content, then retrieve the full history.
```

**Expected Result**:
- All 20 thoughts created successfully
- History retrieval completes in < 2 seconds
- All thoughts present in history

**Pass Criteria**: ✅ No performance degradation with 20+ thoughts

---

### Test 5.2: Multiple Branches

**Tool**: `think`, `list-branches`
**Prompt**:
```
Create 10 different branches in tree mode, then list all branches.
```

**Expected Result**:
- All 10 branches created
- list-branches returns all branches
- No performance issues

**Pass Criteria**: ✅ Branch management handles multiple branches efficiently

---

## Test Results Summary Template

```
==================================================
UNIFIED THINKING MCP SERVER - TEST RESULTS
==================================================
Date: [DATE]
Tester: Claude Desktop
Duration: [TIME]

CORE TOOLS (1-11):
[ ] Test 1.1:  Linear thinking
[ ] Test 1.2:  Tree mode branches
[ ] Test 1.3:  Divergent mode
[ ] Test 1.4:  Auto mode detection
[ ] Test 1.5:  View history
[ ] Test 1.6:  List branches
[ ] Test 1.7:  Focus branch
[ ] Test 1.8:  Branch history
[ ] Test 1.9:  Validate logic
[ ] Test 1.10: Prove conclusion
[ ] Test 1.11: Check syntax
[ ] Test 1.12: Search thoughts
[ ] Test 1.13: Get metrics
[ ] Test 1.14: Recent branches

COGNITIVE TOOLS (12-19):
[ ] Test 2.1:  Probabilistic reasoning - create
[ ] Test 2.2:  Probabilistic reasoning - update
[ ] Test 2.3:  Probabilistic reasoning - combine
[ ] Test 2.4:  Assess evidence - strong
[ ] Test 2.5:  Assess evidence - weak
[ ] Test 2.6:  Detect contradictions
[ ] Test 2.7:  Make decision (MCDA)
[ ] Test 2.8:  Decompose problem
[ ] Test 2.9:  Sensitivity analysis
[ ] Test 2.10: Self-evaluate thought
[ ] Test 2.11: Detect biases - confirmation
[ ] Test 2.12: Detect biases - overconfidence
[ ] Test 2.13: Detect biases - sunk cost

INTEGRATION TESTS:
[ ] Test 3.1: Full reasoning workflow
[ ] Test 3.2: Tree + contradiction detection
[ ] Test 3.3: Evidence + belief update

ERROR HANDLING:
[ ] Test 4.1: Invalid probability
[ ] Test 4.2: Non-existent branch
[ ] Test 4.3: Invalid parameters

PERFORMANCE:
[ ] Test 5.1: Large history (20+ thoughts)
[ ] Test 5.2: Multiple branches (10+)

==================================================
TOTAL: [ ]/36 PASSED
RESULT: [PASS/FAIL]
==================================================

NOTES:
[Any issues, observations, or recommendations]

CRITICAL FAILURES:
[List any blocking issues]

RECOMMENDATIONS:
[Suggestions for improvements]
```

---

## Troubleshooting

### Issue: Tools Not Appearing

**Solution**:
1. Check config path is correct
2. Restart Claude Desktop completely
3. Check server binary exists at specified path
4. Enable DEBUG=true in config
5. Check Claude Desktop logs

### Issue: Tool Returns Error

**Solution**:
1. Verify input format matches documentation
2. Check parameter types (strings vs numbers)
3. Ensure required fields are provided
4. Review error message for specific validation failure

### Issue: Slow Performance

**Solution**:
1. Check number of stored thoughts (use get-metrics)
2. Restart Claude Desktop if memory usage high
3. Server uses in-memory storage - long sessions may accumulate data

---

## Success Criteria

**Minimum for PASS**:
- All 14 core tool tests pass
- At least 12/13 cognitive tool tests pass
- All 3 integration tests pass
- All 3 error handling tests pass
- Both performance tests pass

**Result**: 33/36 tests = **PASS**

---

## Notes for Tester

1. **Run tests in order** - Some tests depend on previous tests (e.g., Test 1.7 needs branch from Test 1.2)

2. **Save IDs** - Note thought IDs, branch IDs, belief IDs for use in subsequent tests

3. **Fresh session recommended** - Start with a new conversation for clean state

4. **Error handling is important** - Test 4.2 specifically validates the bug fix from Sprint 1

5. **Timing** - Most tests should complete in < 5 seconds each

6. **Documentation** - Refer to README.md for tool parameter details if needed

---

**Test Plan Version**: 1.0
**Last Updated**: 2025-10-01
**Status**: Ready for Execution
