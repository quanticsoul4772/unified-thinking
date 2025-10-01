# Known Issues and Limitations

## Issues Discovered During Testing

### Issue 1: focus-branch Tool Error (Priority: Medium)

**Status:** Needs Investigation
**Discovered:** 2025-01-10 during manual testing
**Severity:** Medium

**Description:**
The focus-branch tool returns an error when attempting to focus on a branch during Claude Desktop testing.

**Expected Behavior:**
- User provides valid branch ID
- Server switches active branch
- Returns success status with new active branch ID

**Actual Behavior:**
- Error returned when calling focus-branch
- Branch switching fails

**Possible Causes:**
1. Branch ID format mismatch (branch might not exist in storage)
2. Timing issue (branch creation vs. focus attempt)
3. Server state issue

**Code Locations:**
- Handler: internal/server/server.go:314 (handleFocusBranch)
- Storage: internal/storage/memory.go:239 (SetActiveBranch)

**Investigation Steps:**
1. Verify branch exists before attempting to focus
2. Check branch ID format consistency
3. Add debug logging to identify exact error
4. Test with explicit branch creation followed by immediate focus

**Workaround:**
None currently. Branch focusing unavailable.

**Fix Priority:** Medium (feature is non-critical but useful for workflow)

---

## Known Limitations

### Limitation 1: Prove Tool Logical Inference (Priority: Low)

**Status:** By Design (Enhancement Opportunity)
**Discovered:** 2025-01-10 during manual testing
**Severity:** Low

**Description:**
The prove tool cannot prove common valid syllogisms like "All humans are mortal, Socrates is human → Socrates is mortal".

**Current Behavior:**
- Uses simple proof strategy based on content matching
- Returns is_provable: false for valid syllogisms

**Expected Future Behavior:**
- Implement proper logical inference engine
- Support modus ponens, modus tollens, syllogisms
- Recognize standard logical forms

**Code Location:**
- internal/validation/logic.go:simpleProof

**Impact:**
- Users cannot rely on prove tool for formal logic validation
- Tool is useful only for very simple direct implications

**Enhancement Considerations:**
- Implement SAT solver integration
- Add predicate logic parser
- Support first-order logic inference rules

**Priority:** Low (documented limitation, not breaking functionality)

---

### Limitation 2: Syntax Checker Missing Parentheses Detection (Priority: Low)

**Status:** Enhancement Opportunity
**Discovered:** 2025-01-10 during manual testing
**Severity:** Low

**Description:**
The check-syntax tool does not detect unbalanced parentheses in logical statements.

**Example:**
Input: "If (A then B"
Expected: Not well-formed (unbalanced parentheses)
Actual: Well-formed

**Current Behavior:**
- Checks for: empty statements, single words, minimum length
- Does NOT check for: balanced parentheses, malformed operators

**Code Location:**
- internal/validation/logic.go:checkSyntax (line 62)
- internal/validation/logic.go:getSyntaxIssues

**Enhancement:**
Already implemented in getSyntaxIssues but may not be fully utilized.

**Investigation:**
Review why balanced parentheses check is not catching this case.

**Priority:** Low (minor validation gap, not critical)

---

### Limitation 3: Cognitive Reasoning Tools Not Exposed (Priority: High)

**Status:** Implementation Complete, Integration Pending
**Impact:** High Value Features Unavailable

**Description:**
The following cognitive reasoning capabilities are implemented in code but not yet exposed as MCP tools:

1. **probabilistic-reasoning** - Bayesian inference and belief updates
2. **assess-evidence** - Evidence quality assessment
3. **detect-contradictions** - Cross-thought contradiction detection
4. **make-decision** - Multi-criteria decision analysis
5. **decompose-problem** - Problem breakdown with dependencies
6. **sensitivity-analysis** - Robustness testing
7. **self-evaluate** - Metacognitive self-assessment
8. **detect-biases** - Cognitive bias identification

**Code Locations:**
- Implementations: internal/reasoning/, internal/analysis/, internal/metacognition/
- Integration needed: internal/server/server.go (RegisterTools)

**Required Work:**
- Add MCP tool handlers for each feature
- Define request/response structures
- Add validation functions
- Update tool registration
- Update documentation

**Priority:** High (high-value features ready for deployment)

**Estimated Effort:** 2-4 hours per tool (16-32 hours total)

---

## Test Coverage Gaps

### Gap 1: Error Handling Tests Partially Complete

**Status:** 4/8 Tests Completed
**Tests Remaining:** 4 error handling tests

Error scenarios tested (PASS):
- Invalid mode ✓
- Missing required fields ✓
- Invalid branch IDs ✓
- Invalid thought IDs ✓

Error scenarios not yet tested:
- Empty content
- Very long content
- Special characters
- Invalid UTF-8

**Action:** Complete remaining 4 error handling tests

---

### Gap 2: Advanced Options Tests COMPLETE

**Status:** 3/3 Tests Completed
**All Tests:** PASS ✓

Features tested successfully:
- Confidence levels ✓
- Key points extraction ✓
- Validation requirements ✓

**Status:** COMPLETE

---

## Performance Observations

**Positive Findings:**
- All operations complete in < 1 second
- No memory issues observed during testing
- Server stable throughout test session
- MCP protocol communication working correctly
- Thread-safe operations verified through unit tests

**Areas Not Yet Tested:**
- High-volume thought creation (100+ thoughts)
- Many branches (50+ branches)
- Large history retrieval
- Concurrent operations under load
- Long-running server stability (24+ hours)

---

## Recommendations

### Immediate Actions
1. Debug focus-branch error (investigate branch ID handling)
2. Complete remaining 18 manual tests
3. Document prove tool limitations in user-facing docs

### Short-term Enhancements
1. Add cognitive reasoning tools as MCP tools (high value)
2. Enhance syntax checker for parentheses detection
3. Add better error messages for focus-branch failures

### Long-term Improvements
1. Implement proper logical inference engine for prove tool
2. Add performance profiling under load
3. Consider persistence layer for long-running sessions
4. Add metrics dashboard for system monitoring

---

## Version History

- 2025-01-10: Initial issues document created during manual testing
  - 1 issue identified (focus-branch)
  - 3 limitations documented
  - 2 test coverage gaps noted
  - Recommendations added
