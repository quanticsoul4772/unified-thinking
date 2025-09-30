# Unified Thinking MCP Server Test Plan

This test plan verifies all 9 tools, 4 thinking modes, edge cases, validation limits, and performance characteristics of the Unified Thinking MCP Server.

## Prerequisites

1. Unified Thinking MCP Server is configured in Claude Desktop config at %APPDATA%\Claude\claude_desktop_config.json
2. Claude Desktop has been restarted after configuration
3. Server binary exists at configured path (bin/unified-thinking.exe)

## Test Environment Setup

Before starting tests, note the baseline:
- Time: Record start time for session duration testing
- Memory: Note if any performance degradation occurs during testing
- Session: All tests in a single Claude Desktop session to test state persistence

## Test Suite Overview

Total tools: 9
Total modes: 4 (linear, tree, divergent, auto)
Total test cases: 67
Estimated duration: 30-45 minutes

---

## SECTION 1: BASIC TOOL VERIFICATION (9 tests)

### Test 1.1: think tool - auto mode
**Objective:** Verify basic think functionality with auto mode selection

**Execution:**
Ask Claude to use the think tool with this exact request:
"Use the think tool with content='Analyze the benefits of test-driven development' and mode='auto'"

**Expected Result:**
- JSON response with fields: thought_id, mode (one of: linear/tree/divergent), status="success", confidence (0.0-1.0)
- thought_id format: thought-{UUID or timestamp-based ID}
- Response time: < 500ms

**Performance Metrics:**
- Record which mode was auto-selected
- Note response time
- Verify thought_id is non-empty string

---

### Test 1.2: think tool - linear mode
**Objective:** Verify explicit linear mode selection

**Execution:**
"Use the think tool with content='Step by step: How to build a REST API' and mode='linear'"

**Expected Result:**
- mode field in response = "linear"
- thought_id present
- status = "success"

**Performance Metrics:**
- Response time should be similar to Test 1.1 (within 100ms)

---

### Test 1.3: think tool - tree mode
**Objective:** Verify tree mode with branch creation

**Execution:**
"Use the think tool with content='Explore multiple approaches to database optimization' and mode='tree'"

**Expected Result:**
- mode = "tree"
- branch_id present in response (format: branch-{ID})
- priority field present (0.0-1.0)
- thought_id present

**Performance Metrics:**
- Record branch_id for later tests
- Note if response includes insight_count field

---

### Test 1.4: think tool - divergent mode
**Objective:** Verify divergent mode for creative thinking

**Execution:**
"Use the think tool with content='What is an unconventional solution to code documentation?' and mode='divergent'"

**Expected Result:**
- mode = "divergent"
- thought_id present
- confidence value present

**Performance Metrics:**
- Compare response time to linear mode (may be slightly higher due to creativity processing)

---

### Test 1.5: history tool
**Objective:** Retrieve all thinking history

**Execution:**
"Use the history tool with no parameters to show all thoughts"

**Expected Result:**
- JSON array of thoughts
- Should contain at least 4 thoughts from Tests 1.1-1.4
- Each thought has: id, content, mode, timestamp
- Thoughts ordered by timestamp

**Performance Metrics:**
- Response time: < 200ms for 4 thoughts
- Note if response time scales linearly with thought count

---

### Test 1.6: list-branches tool
**Objective:** List all thinking branches

**Execution:**
"Use the list-branches tool"

**Expected Result:**
- JSON with branches array and active_branch_id field
- At least 1 branch from Test 1.3
- Each branch has: id, state, priority, confidence, created_at

**Performance Metrics:**
- Record number of branches
- Record active_branch_id value

---

### Test 1.7: search tool
**Objective:** Search across all thoughts

**Execution:**
"Use the search tool with query='API'"

**Expected Result:**
- JSON array of matching thoughts
- Should find thought from Test 1.2 (contains "REST API")
- Case-insensitive search

**Performance Metrics:**
- Response time: < 300ms
- Verify search is substring match

---

### Test 1.8: check-syntax tool
**Objective:** Validate logical statement syntax

**Execution:**
"Use the check-syntax tool with statements=['All developers write tests', 'Some code has bugs', 'Testing improves quality']"

**Expected Result:**
- JSON array of checks
- 3 check objects, each with: statement, is_well_formed (boolean), issues (array)
- All should have is_well_formed = true

**Performance Metrics:**
- Response time: < 200ms for 3 statements

---

### Test 1.9: validate tool
**Objective:** Validate a thought for logical consistency

**Execution:**
"Use the validate tool with the thought_id from Test 1.2"

**Expected Result:**
- JSON with is_valid (boolean) and reason (string)
- is_valid = true expected
- reason = "Thought passes basic logical consistency checks"

**Performance Metrics:**
- Response time: < 150ms

---

## SECTION 2: MODE-SPECIFIC TESTING (12 tests)

### Test 2.1: Linear mode - sequential steps
**Objective:** Verify linear mode processes sequential reasoning

**Execution:**
"Use think tool: content='First establish requirements, then design architecture, finally implement features', mode='linear', key_points=['requirements', 'design', 'implementation']"

**Expected Result:**
- mode = "linear"
- Response includes thought_id
- key_points preserved in storage (verify with history)

**Performance Metrics:**
- Test with 3 key_points provided

---

### Test 2.2: Linear mode - with confidence
**Objective:** Test confidence parameter handling

**Execution:**
"Use think tool: content='Database indexing improves query performance', mode='linear', confidence=0.95"

**Expected Result:**
- confidence field in response = 0.95
- thought stored with specified confidence

**Performance Metrics:**
- Verify confidence precision (should be exact float match)

---

### Test 2.3: Linear mode - with validation
**Objective:** Test require_validation flag

**Execution:**
"Use think tool: content='All programs must be tested before deployment', mode='linear', require_validation=true"

**Expected Result:**
- is_valid field present in response
- is_valid = true
- Validation performed automatically

**Performance Metrics:**
- Response time increase due to validation: expect +50-100ms vs without validation

---

### Test 2.4: Tree mode - branch creation
**Objective:** Test explicit branch creation in tree mode

**Execution:**
"Use think tool: content='Approach 1: Monolithic architecture', mode='tree', type='exploration'"

**Expected Result:**
- branch_id present
- New branch created
- type field preserved

**Performance Metrics:**
- Record new branch_id (Branch-A)

---

### Test 2.5: Tree mode - same branch continuation
**Objective:** Add thought to existing branch

**Execution:**
"Use think tool: content='Pros: Simple deployment. Cons: Scaling challenges', mode='tree', branch_id={Branch-A from Test 2.4}"

**Expected Result:**
- Same branch_id returned as provided
- Thought added to existing branch

**Performance Metrics:**
- Verify branch thought count increased

---

### Test 2.6: Tree mode - parallel branch
**Objective:** Create parallel exploration branch

**Execution:**
"Use think tool: content='Approach 2: Microservices architecture', mode='tree', type='exploration'"

**Expected Result:**
- Different branch_id from Branch-A (Branch-B)
- Two parallel branches exist

**Performance Metrics:**
- list-branches should show 2 branches for tree mode

---

### Test 2.7: Tree mode - with cross-references
**Objective:** Test cross-reference between branches

**Execution:**
"Use think tool: content='Microservices provide better scaling than monolithic', mode='tree', branch_id={Branch-B}, cross_refs=[{to_branch: {Branch-A}, type: 'contradictory', reason: 'Different scaling approaches', strength: 0.8}]"

**Expected Result:**
- Cross-reference created
- Response successful
- Cross-ref stored with specified attributes

**Performance Metrics:**
- Verify cross-ref appears in branch history

---

### Test 2.8: Divergent mode - creative thinking
**Objective:** Test divergent mode basic creativity

**Execution:**
"Use think tool: content='Unusual way to handle errors: celebrate them as learning opportunities', mode='divergent'"

**Expected Result:**
- mode = "divergent"
- Thought processed successfully

---

### Test 2.9: Divergent mode - force rebellion
**Objective:** Test force_rebellion parameter

**Execution:**
"Use think tool: content='Challenge: Maybe documentation should be generated from conversations, not written', mode='divergent', force_rebellion=true"

**Expected Result:**
- Thought created with is_rebellion flag (stored internally)
- Response successful

---

### Test 2.10: Auto mode - problem-solving detection
**Objective:** Verify auto mode selects appropriate mode for step-by-step content

**Execution:**
"Use think tool: content='To solve this bug: 1) Reproduce the error, 2) Check logs, 3) Fix root cause, 4) Add test', mode='auto'"

**Expected Result:**
- Auto mode likely selects "linear" (sequential pattern detected)
- Verify mode field in response

**Performance Metrics:**
- Auto-selection algorithm processing time (should add minimal overhead)

---

### Test 2.11: Auto mode - exploration detection
**Objective:** Verify auto mode detects exploration needs

**Execution:**
"Use think tool: content='Let us explore various caching strategies and their tradeoffs', mode='auto'"

**Expected Result:**
- Auto mode may select "tree" (exploration keyword detected)
- Or "linear" depending on implementation heuristics

**Performance Metrics:**
- Document which mode was selected and reasoning

---

### Test 2.12: Auto mode - creative content detection
**Objective:** Verify auto mode handles creative prompts

**Execution:**
"Use think tool: content='Imagine a completely new paradigm for version control that does not use commits', mode='auto'"

**Expected Result:**
- Auto mode may select "divergent" (creative/unconventional keywords)
- Response successful regardless of selection

---

## SECTION 3: BRANCH OPERATIONS (8 tests)

### Test 3.1: focus-branch tool
**Objective:** Switch active branch

**Execution:**
"Use focus-branch tool with branch_id={Branch-A from Test 2.4}"

**Expected Result:**
- status = "success"
- active_branch_id = {Branch-A}

**Performance Metrics:**
- Response time: < 100ms

---

### Test 3.2: Verify active branch changed
**Objective:** Confirm focus-branch persisted

**Execution:**
"Use list-branches tool"

**Expected Result:**
- active_branch_id field matches Branch-A
- State change persisted in memory

---

### Test 3.3: branch-history tool
**Objective:** Get detailed branch history

**Execution:**
"Use branch-history tool with branch_id={Branch-A}"

**Expected Result:**
- JSON with branch details
- thoughts array with 2 thoughts (from Tests 2.4 and 2.5)
- insights array (may be empty)
- cross_refs array (may be empty for Branch-A)

**Performance Metrics:**
- Response time: < 200ms

---

### Test 3.4: Branch history with cross-refs
**Objective:** Verify cross-refs appear in branch history

**Execution:**
"Use branch-history tool with branch_id={Branch-B}"

**Expected Result:**
- cross_refs array contains 1 cross-reference to Branch-A
- Cross-ref has: to_branch, type='contradictory', reason, strength=0.8

---

### Test 3.5: History filtered by mode
**Objective:** Test mode filtering in history tool

**Execution:**
"Use history tool with mode='tree'"

**Expected Result:**
- Only thoughts with mode='tree' returned
- Should include thoughts from Tests 2.4, 2.5, 2.6, 2.7

**Performance Metrics:**
- Filtering efficiency: response time should be similar to unfiltered

---

### Test 3.6: History filtered by branch
**Objective:** Test branch_id filtering

**Execution:**
"Use history tool with branch_id={Branch-A}"

**Expected Result:**
- Only thoughts from Branch-A returned
- Should be 2 thoughts

---

### Test 3.7: Search filtered by mode
**Objective:** Test mode filtering in search

**Execution:**
"Use search tool with query='architecture', mode='tree'"

**Expected Result:**
- Only tree-mode thoughts matching 'architecture' returned
- Should find thoughts from Tests 2.4 and 2.6

---

### Test 3.8: Create thought in non-active branch
**Objective:** Verify thoughts can be added to non-active branches

**Execution:**
First ensure Branch-A is active, then:
"Use think tool: content='Additional thought for Branch B', mode='tree', branch_id={Branch-B}"

**Expected Result:**
- Thought added to Branch-B despite Branch-A being active
- Explicit branch_id parameter takes precedence

---

## SECTION 4: VALIDATION AND PROOF (8 tests)

### Test 4.1: Validate - consistent thought
**Objective:** Test validation on logically consistent content

**Execution:**
"Use think tool: content='Software testing reduces bugs in production', mode='linear'"
Then: "Use validate tool with the returned thought_id"

**Expected Result:**
- is_valid = true
- reason = "Thought passes basic logical consistency checks"

---

### Test 4.2: Validate - contradictory thought
**Objective:** Test detection of contradictions

**Execution:**
"Use think tool: content='This feature always works perfectly and never fails under any circumstances but also fails frequently', mode='linear'"
Then: "Use validate tool with the returned thought_id"

**Expected Result:**
- is_valid = false
- reason contains "contradictory" or "always/never"

**Performance Metrics:**
- Contradiction detection pattern matching

---

### Test 4.3: Validate - all/none contradiction
**Objective:** Test universal quantifier contradiction detection

**Execution:**
"Use think tool: content='All users prefer feature A and none of the users like feature A', mode='linear'"
Then: "Use validate tool with the returned thought_id"

**Expected Result:**
- is_valid = false
- reason mentions "all/none" contradiction

---

### Test 4.4: prove tool - valid syllogism
**Objective:** Test basic logical proof

**Execution:**
"Use prove tool with premises=['All humans are mortal', 'Socrates is human'], conclusion='Socrates is mortal'"

**Expected Result:**
- is_provable = true
- premises array matches input
- conclusion matches input
- steps array contains proof steps (minimum 3 steps)

**Performance Metrics:**
- Response time: < 200ms

---

### Test 4.5: prove tool - invalid conclusion
**Objective:** Test proof rejection for invalid logic

**Execution:**
"Use prove tool with premises=['Some developers use Python'], conclusion='All developers use Python'"

**Expected Result:**
- is_provable = false (simplified validator may return true - document behavior)
- steps array present

**Note:** Simplified validator has limited logic - document actual behavior

---

### Test 4.6: check-syntax - well-formed statements
**Objective:** Test syntax validation on proper statements

**Execution:**
"Use check-syntax tool with statements=['All functions should have tests', 'Code reviews improve quality', 'Documentation helps maintainability']"

**Expected Result:**
- 3 check objects
- All have is_well_formed = true
- issues arrays are empty

---

### Test 4.7: check-syntax - malformed statements
**Objective:** Test syntax error detection

**Execution:**
"Use check-syntax tool with statements=['ProperStatement', 'SingleWord', '   ', 'Another proper statement here']"

**Expected Result:**
- Statement 1: is_well_formed = false (single word)
- Statement 2: is_well_formed = false (single word)
- Statement 3: is_well_formed = false (empty)
- Statement 4: is_well_formed = true

**Performance Metrics:**
- Verify issues array contains helpful error descriptions

---

### Test 4.8: Validation with require_validation flag
**Objective:** Test automatic validation during thought creation

**Execution:**
"Use think tool: content='Testing is important for all software projects and no software needs testing', mode='linear', require_validation=true"

**Expected Result:**
- is_valid field in response = false
- Thought still created but flagged as invalid

**Performance Metrics:**
- Response time should include validation overhead (+50-100ms)

---

## SECTION 5: EDGE CASES AND ERROR HANDLING (15 tests)

### Test 5.1: Empty content
**Objective:** Test content validation

**Execution:**
"Use think tool with content='', mode='linear'"

**Expected Result:**
- Error: "validation error on field 'content': content cannot be empty"
- No thought created

---

### Test 5.2: Invalid mode
**Objective:** Test mode validation

**Execution:**
"Use think tool with content='test', mode='invalid_mode'"

**Expected Result:**
- Error: "validation error on field 'mode': invalid mode: invalid_mode (must be 'linear', 'tree', 'divergent', or 'auto')"

---

### Test 5.3: Confidence out of range - high
**Objective:** Test confidence bounds checking

**Execution:**
"Use think tool with content='test', mode='linear', confidence=1.5"

**Expected Result:**
- Error: "validation error on field 'confidence': confidence must be between 0.0 and 1.0"

---

### Test 5.4: Confidence out of range - low
**Objective:** Test negative confidence rejection

**Execution:**
"Use think tool with content='test', mode='linear', confidence=-0.5"

**Expected Result:**
- Error: "validation error on field 'confidence': confidence must be between 0.0 and 1.0"

---

### Test 5.5: Non-existent branch
**Objective:** Test branch existence validation

**Execution:**
"Use focus-branch tool with branch_id='nonexistent-branch-12345'"

**Expected Result:**
- Error indicating branch not found

---

### Test 5.6: Non-existent thought ID
**Objective:** Test thought existence validation

**Execution:**
"Use validate tool with thought_id='nonexistent-thought-99999'"

**Expected Result:**
- Error indicating thought not found

---

### Test 5.7: Maximum content length
**Objective:** Test content size limit (100KB)

**Execution:**
Create a string with 100001 bytes:
"Use think tool with content={100001 character string}, mode='linear'"

**Expected Result:**
- Error: "validation error on field 'content': content exceeds maximum length of 100000 bytes"

**Performance Metrics:**
- Validation should reject before processing

---

### Test 5.8: Maximum key points count
**Objective:** Test key_points array limit (50)

**Execution:**
"Use think tool with content='test', mode='linear', key_points={array of 51 strings}"

**Expected Result:**
- Error: "validation error on field 'key_points': too many key points (max 50)"

---

### Test 5.9: Maximum key point length
**Objective:** Test individual key point size limit (1KB)

**Execution:**
"Use think tool with content='test', mode='linear', key_points=[{1001 character string}]"

**Expected Result:**
- Error: "validation error on field 'key_points': key_points[0] exceeds max length of 1000"

---

### Test 5.10: Maximum cross-references
**Objective:** Test cross_refs array limit (20)

**Execution:**
"Use think tool with content='test', mode='tree', cross_refs={array of 21 cross-ref objects}"

**Expected Result:**
- Error: "validation error on field 'cross_refs': too many cross references (max 20)"

---

### Test 5.11: Invalid cross-ref type
**Objective:** Test cross-ref type validation

**Execution:**
"Use think tool with content='test', mode='tree', cross_refs=[{to_branch: 'branch-1', type: 'invalid_type', reason: 'test', strength: 0.5}]"

**Expected Result:**
- Error: "validation error on field 'cross_refs': cross_refs[0].type invalid (must be 'complementary', 'contradictory', 'builds_upon', or 'alternative')"

---

### Test 5.12: Cross-ref strength out of range
**Objective:** Test cross-ref strength validation

**Execution:**
"Use think tool with content='test', mode='tree', cross_refs=[{to_branch: 'branch-1', type: 'complementary', reason: 'test', strength: 1.5}]"

**Expected Result:**
- Error: "validation error on field 'cross_refs': cross_refs[0].strength must be 0.0-1.0"

---

### Test 5.13: Empty premises array
**Objective:** Test prove tool validation

**Execution:**
"Use prove tool with premises=[], conclusion='test conclusion'"

**Expected Result:**
- Error: "validation error on field 'premises': at least one premise is required"

---

### Test 5.14: Too many premises
**Objective:** Test premise count limit (50)

**Execution:**
"Use prove tool with premises={array of 51 strings}, conclusion='test'"

**Expected Result:**
- Error: "validation error on field 'premises': too many premises (max 50)"

---

### Test 5.15: Empty statements array
**Objective:** Test check-syntax validation

**Execution:**
"Use check-syntax tool with statements=[]"

**Expected Result:**
- Error: "validation error on field 'statements': at least one statement is required"

---

## SECTION 6: PERFORMANCE AND SCALING (10 tests)

### Test 6.1: Sequential thought creation - 20 thoughts
**Objective:** Test performance with moderate thought count

**Execution:**
Create 20 thoughts sequentially using think tool (any mode)

**Performance Metrics:**
- Time for thought 1 vs thought 20 (should be similar)
- Total time for 20 creations
- Response time degradation: < 10%
- Use history tool to verify all 20 stored

**Expected Result:**
- Consistent response times (variance < 100ms)
- All thoughts retrievable

---

### Test 6.2: History retrieval scaling
**Objective:** Test history tool performance with multiple thoughts

**Execution:**
After Test 6.1, use history tool

**Performance Metrics:**
- Response time for returning 20+ thoughts
- Target: < 500ms
- JSON serialization efficiency

**Expected Result:**
- All thoughts returned
- Proper JSON array formatting

---

### Test 6.3: Search performance with large dataset
**Objective:** Test search efficiency

**Execution:**
After building up 20+ thoughts, search for common term

**Performance Metrics:**
- Search response time: < 300ms
- Compare to baseline search (Test 1.7)

**Expected Result:**
- All matching thoughts returned
- Performance degradation < 50% vs baseline

---

### Test 6.4: Multiple branches - 10 branches
**Objective:** Test tree mode with many branches

**Execution:**
Create 10 different branches using tree mode

**Performance Metrics:**
- list-branches response time
- Memory stability (no degradation noted)

**Expected Result:**
- All 10 branches listed
- Response time: < 500ms

---

### Test 6.5: Branch switching performance
**Objective:** Test focus-branch efficiency

**Execution:**
Switch between 5 different branches sequentially

**Performance Metrics:**
- Average focus-branch response time
- Target: < 100ms per switch

**Expected Result:**
- All switches successful
- Consistent performance

---

### Test 6.6: Validation batch performance
**Objective:** Test validation on multiple thoughts

**Execution:**
Validate 10 different thoughts using validate tool

**Performance Metrics:**
- Average validation time
- Target: < 150ms per validation
- Consistency across validations

---

### Test 6.7: Proof complexity - maximum premises
**Objective:** Test prove tool with max premises (50)

**Execution:**
"Use prove tool with premises={array of 50 valid statements}, conclusion='complex conclusion'"

**Performance Metrics:**
- Response time with 50 premises
- Target: < 1000ms
- Steps array length

**Expected Result:**
- Proof processed successfully
- No timeout or error

---

### Test 6.8: Syntax check batch - maximum statements
**Objective:** Test check-syntax with max statements (100)

**Execution:**
"Use check-syntax tool with statements={array of 100 statements}"

**Performance Metrics:**
- Response time for 100 statements
- Target: < 2000ms
- Per-statement processing time: ~20ms

**Expected Result:**
- All 100 statements checked
- Complete checks array returned

---

### Test 6.9: Large content processing
**Objective:** Test near-maximum content size

**Execution:**
"Use think tool with content={99KB string}, mode='linear'"

**Performance Metrics:**
- Response time for large content
- JSON serialization time
- Target: < 1000ms

**Expected Result:**
- Thought created successfully
- Content preserved in full

---

### Test 6.10: Session longevity
**Objective:** Test memory stability over long session

**Execution:**
After completing all previous tests:
1. Count total thoughts created
2. Use history tool
3. Use list-branches tool
4. Perform new think operation

**Performance Metrics:**
- No performance degradation noted
- All tools still responsive
- Response times within normal ranges

**Expected Result:**
- All data still accessible
- No memory-related errors
- Stable performance

---

## SECTION 7: CROSS-REFERENCE AND RELATIONSHIPS (5 tests)

### Test 7.1: Cross-reference - complementary
**Objective:** Test complementary relationship type

**Execution:**
Create two tree branches, then:
"Use think tool with content='Branch 2 complements Branch 1 analysis', mode='tree', cross_refs=[{to_branch: {Branch-1}, type: 'complementary', reason: 'Provides additional perspective', strength: 0.9}]"

**Expected Result:**
- Cross-reference created
- Type preserved as 'complementary'

---

### Test 7.2: Cross-reference - builds_upon
**Objective:** Test builds_upon relationship

**Execution:**
"Use think tool with content='Building on previous architecture analysis', mode='tree', cross_refs=[{to_branch: {existing-branch}, type: 'builds_upon', reason: 'Extends the architecture discussion', strength: 0.85}]"

**Expected Result:**
- Cross-reference type = 'builds_upon'
- Proper storage and retrieval

---

### Test 7.3: Cross-reference - alternative
**Objective:** Test alternative relationship

**Execution:**
"Use think tool with content='Alternative approach to the problem', mode='tree', cross_refs=[{to_branch: {existing-branch}, type: 'alternative', reason: 'Different solution path', strength: 0.7}]"

**Expected Result:**
- Type = 'alternative'
- Strength = 0.7 preserved

---

### Test 7.4: Multiple cross-references in single thought
**Objective:** Test multiple cross-refs per thought

**Execution:**
"Use think tool with content='Synthesis of multiple branches', mode='tree', cross_refs=[{to_branch: {Branch-1}, type: 'builds_upon', reason: 'reason1', strength: 0.8}, {to_branch: {Branch-2}, type: 'complementary', reason: 'reason2', strength: 0.75}]"

**Expected Result:**
- Both cross-references stored
- Retrievable via branch-history

---

### Test 7.5: Cross-reference retrieval
**Objective:** Verify cross-refs appear in branch queries

**Execution:**
"Use branch-history tool on branch with cross-refs"

**Expected Result:**
- cross_refs array populated
- All cross-ref attributes present: to_branch, type, reason, strength

---

## TEST COMPLETION CHECKLIST

After completing all test sections, verify:

1. Tool Coverage:
   - [ ] think (Tests 1.1-1.4, 2.1-2.12, and others)
   - [ ] history (Tests 1.5, 3.5, 3.6, 6.2)
   - [ ] list-branches (Tests 1.6, 3.2, 6.4)
   - [ ] focus-branch (Tests 3.1, 3.2, 6.5)
   - [ ] branch-history (Tests 3.3, 3.4, 7.5)
   - [ ] validate (Tests 1.9, 4.1-4.3, 4.8, 6.6)
   - [ ] prove (Tests 4.4, 4.5, 6.7)
   - [ ] check-syntax (Tests 1.8, 4.6, 4.7, 6.8)
   - [ ] search (Tests 1.7, 3.7, 6.3)

2. Mode Coverage:
   - [ ] linear (Tests 1.2, 2.1-2.3)
   - [ ] tree (Tests 1.3, 2.4-2.7, 3.x, 7.x)
   - [ ] divergent (Tests 1.4, 2.8, 2.9)
   - [ ] auto (Tests 1.1, 2.10-2.12)

3. Edge Cases:
   - [ ] Input validation (Tests 5.1-5.15)
   - [ ] Boundary conditions (Tests 5.7-5.12)
   - [ ] Error handling (Tests 5.1-5.6)

4. Performance:
   - [ ] Scaling behavior (Tests 6.1-6.10)
   - [ ] Response time consistency
   - [ ] No memory degradation

5. Integration:
   - [ ] Cross-references (Tests 2.7, 7.1-7.5)
   - [ ] Branch operations (Tests 3.1-3.8)
   - [ ] Validation integration (Tests 2.3, 4.8)

## PERFORMANCE SUMMARY TEMPLATE

After completing all tests, document:

**Response Time Benchmarks:**
- think tool average: ___ ms
- history tool average: ___ ms
- validate tool average: ___ ms
- search tool average: ___ ms
- Other tools average: ___ ms

**Scaling Observations:**
- Thought count before degradation: ___
- Branch count tested: ___
- Performance degradation percentage: ___%

**Issues Found:**
1. Issue description, test number, severity
2. ...

**Auto Mode Behavior:**
- Sequential content -> mode selected: ___
- Exploration content -> mode selected: ___
- Creative content -> mode selected: ___

**Memory Stability:**
- Session duration: ___ minutes
- Total operations: ___
- Stability: Stable / Degraded / Failed

## COMMON ISSUES TO WATCH FOR

1. **JSON Formatting Errors:**
   - Malformed JSON responses
   - Missing required fields
   - Incorrect data types

2. **Validation Bypass:**
   - Invalid input accepted
   - Limits not enforced
   - UTF-8 validation failures

3. **State Management:**
   - Active branch not persisting
   - Thoughts appearing in wrong branches
   - Cross-references not stored

4. **Performance Degradation:**
   - Response times increasing over session
   - Memory usage growing unbounded
   - Search slowing with data growth

5. **Error Handling:**
   - Generic error messages
   - Missing validation errors
   - Tool crashes on invalid input

6. **Auto Mode Selection:**
   - Inconsistent mode selection
   - Unexpected mode choices
   - Mode selection not documented in response

## NOTES SECTION

Use this section to document:
- Unexpected behaviors
- Performance anomalies
- Feature requests
- Bug reports with reproduction steps

---

## Test Plan Version: 1.0
## Last Updated: 2025-09-30
## Target Server Version: v1.0.0
