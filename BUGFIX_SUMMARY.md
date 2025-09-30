# Critical Bug Fix: Branch History Empty

## Issue
Test 3.3 revealed that branch-history tool returns empty thoughts array despite thoughts being created in that branch.

## Root Cause
In `internal/modes/tree.go`, the `ProcessThought` method was modifying a branch object but never persisting it back to storage.

**Problem code (lines 65-101):**
```go
// Update branch
branch, err := m.storage.GetBranch(branchID)  // Gets a DEEP COPY
if err != nil {
    return nil, err
}
branch.Thoughts = append(branch.Thoughts, thought)  // Modifies the COPY
// ... more modifications to insights and cross-refs ...
m.updateBranchMetrics(branch)
branch.UpdatedAt = time.Now()
// BUG: No StoreBranch() call - changes are lost!
```

## Explanation
1. `GetBranch()` returns a **deep copy** of the branch (for thread safety)
2. The code modifies this copy by appending thoughts, insights, and cross-refs
3. **The modified copy is never stored back** via `StoreBranch()`
4. When `branch-history` is called later, it retrieves the original unmodified branch
5. Result: Empty thoughts array

This is the exact data race bug identified in the performance analysis by go-performance-optimizer agent.

## Fix Applied

**File:** `internal/modes/tree.go`
**Lines:** 103-106 (added)

```go
// Update branch metrics
m.updateBranchMetrics(branch)
branch.UpdatedAt = time.Now()

// Store updated branch back to storage
if err := m.storage.StoreBranch(branch); err != nil {
    return nil, err
}
```

## Impact

**Before fix:**
- Thoughts created in tree mode were stored individually but not associated with branches
- Branch history showed empty thoughts array
- Branch metrics (priority, confidence) were calculated but discarded
- Cross-references and insights were lost

**After fix:**
- Thoughts properly associated with branches
- Branch history returns complete thought list
- Branch metrics persisted correctly
- Cross-references and insights stored properly

## Testing

**Affected Tests:**
- Test 3.3: branch-history tool (was returning empty thoughts)
- Test 2.5: Tree mode same branch continuation (thoughts should accumulate)
- Test 2.7: Tree mode with cross-references (cross-refs should persist)
- Test 3.4: Branch history with cross-refs (cross-refs should appear)
- All Section 7 tests: Cross-reference retrieval

**Verification Steps:**
1. Rebuild server: `go build -o bin/unified-thinking.exe ./cmd/server`
2. Restart Claude Desktop
3. Create 2 thoughts in same branch
4. Use branch-history tool on that branch
5. Verify thoughts array contains both thoughts
6. Verify cross-refs appear if created

## Related Issues

This bug was caused by the same deep-copy strategy that prevents data races:
- Deep copy prevents concurrent modification bugs (GOOD)
- But requires explicit store-back after modification (MISSED)

**Other locations to check for similar pattern:**
```bash
grep -n "GetBranch\|GetThought" internal/modes/*.go
```

Look for any code that:
1. Calls a Get method (returns deep copy)
2. Modifies the returned object
3. Does not call corresponding Store method

## Performance Impact

**Before:** No performance impact (but data was lost)
**After:** One additional `StoreBranch()` call per tree-mode thought

**Overhead:**
- StoreBranch: O(1) map write with exclusive lock
- Negligible impact (< 1ms)
- Required for correctness

## Commit Message

```
Fix critical bug: Branch updates not persisting in tree mode

TreeMode.ProcessThought was modifying a deep copy of the branch
but never calling StoreBranch() to persist changes. This caused:
- Empty thoughts array in branch-history
- Lost cross-references
- Discarded branch metrics updates

Added StoreBranch() call after branch modifications to persist
all changes including thoughts, insights, cross-refs, and metrics.

Fixes: Test 3.3 (branch-history empty thoughts)
Affects: Tests 2.5, 2.7, 3.4, all Section 7 tests
```

## Additional Notes

The go-performance-optimizer agent correctly identified this as a data race:
> "Lines 65-101: gets branch, modifies it, but never stores it back.
> The modification happens outside the lock, so it's lost."

The agent also noted:
> "tree.go line 69: Modifies branch without lock - DATA RACE"

While technically not a race condition (the copy is thread-safe), the practical effect is the same: data loss due to modifications not being persisted.
