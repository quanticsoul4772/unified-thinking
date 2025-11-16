# Phase 3 & Phase 4: Polish and Documentation

Strategic plan for completing the unified-thinking server refactoring.

## Phase 3: Polish - Extract Magic Numbers (Week 5)

### Objective
Replace magic numbers with well-named constants to improve code maintainability and semantic clarity.

### Analysis Summary
- **Total numeric literals found**: 1,266 across target packages
  - validation package: 548 occurrences (11 files)
  - memory package: 346 occurrences (6 files)
  - integration package: 372 occurrences (9 files)

### Priority Order (Decision Score: 0.90)
1. **Validation Package** (validation/calibration.go)
2. **Memory Package** (memory/episodic.go, memory/learning.go)
3. **Integration Package** (integration/synthesizer.go)

### Phase 3a: Validation Package Magic Numbers

**File**: `internal/validation/calibration.go`

**Magic Numbers Identified**:
```go
// Calibration bucket configuration
10.0          → CalibrationBucketCount (number of confidence buckets)
0.1           → CalibrationBucketWidth (width of each bucket)

// Calibration thresholds
0.05          → CalibrationBiasThresholdNone (threshold for "no bias")
0.15          → CalibrationBiasThresholdHigh (threshold for high bias)
0.1           → CalibrationWarningThreshold (general warning threshold)
0.2           → CalibrationBucketWarningThreshold (per-bucket warning)

// Minimum data requirements
5             → MinBucketSampleSize (minimum samples for bucket analysis)
```

**Constants to Add** (validation/calibration.go):
```go
const (
	// Calibration bucket configuration
	CalibrationBucketCount = 10    // Number of calibration buckets (0.0-0.1, 0.1-0.2, ..., 0.9-1.0)
	CalibrationBucketWidth = 0.1   // Width of each calibration bucket

	// Bias magnitude thresholds
	BiasThresholdNone    = 0.05    // Below this is considered "well-calibrated"
	BiasThresholdModerate = 0.15    // Above this is considered "high bias"

	// Calibration error thresholds
	CalibrationWarning        = 0.1  // General calibration warning threshold
	BucketCalibrationWarning  = 0.2  // Per-bucket calibration warning

	// Minimum sample requirements
	MinBucketSamples = 5  // Minimum predictions required for bucket reliability
)
```

**Impact**: ~12 replacements in calibration.go

### Phase 3b: Memory Package Magic Numbers

**Files**:
- `internal/memory/episodic.go`
- `internal/memory/learning.go`

**Magic Numbers Identified**:
```go
// Similarity thresholds
0.3           → MinSimilarityThreshold (minimum similarity for retrieval)
0.7           → HighSuccessThreshold (high success score cutoff)
0.4           → LowSuccessThreshold (low success score cutoff)
0.8           → SuccessRecommendationThreshold (for top recommendations)

// Similarity scoring weights
0.3           → DomainSimilarityWeight (weight for domain matching)
0.2           → ComplexitySimilarityWeight (weight for complexity matching)
0.2           → GoalSimilarityWeight (weight for goal overlap)

// Learning configuration
0.6           → MinSuccessRateForPattern (minimum success rate)
0.2           → ComplexityRangeDelta (complexity matching tolerance)
```

**Constants to Add** (memory/episodic.go):
```go
const (
	// Similarity retrieval thresholds
	MinSimilarityThreshold = 0.3  // Minimum similarity score for case retrieval

	// Success score thresholds for recommendations
	HighSuccessThreshold   = 0.7  // Threshold for successful pattern recommendations
	LowSuccessThreshold    = 0.4  // Below this triggers warnings
	SuccessRecommendationThreshold = 0.8  // For priority recommendations

	// Similarity scoring weights (must sum to 1.0)
	DomainSimilarityWeight     = 0.3  // Weight for domain matching
	ComplexitySimilarityWeight = 0.2  // Weight for complexity similarity
	GoalSimilarityWeight       = 0.2  // Weight for goal overlap
	ContextSimilarityWeight    = 0.3  // Weight for context matching (implicit)

	// Complexity matching tolerance
	ComplexityRangeDelta = 0.2  // +/- range for complexity matching
)
```

**Constants to Add** (memory/learning.go):
```go
const (
	// Pattern learning thresholds
	MinSuccessRateForPattern = 0.6  // Minimum success rate to establish a pattern
	MinTrajectoryCountForPattern = 3  // Minimum trajectories to form a pattern
)
```

**Impact**: ~25 replacements across memory package

### Phase 3c: Integration Package Magic Numbers

**File**: `internal/integration/synthesizer.go`

**Magic Numbers Identified**:
```go
// Synthesis requirements
2             → MinInputsForSynthesis (minimum inputs required)

// Confidence thresholds (to be identified during implementation)
```

**Constants to Add** (integration/synthesizer.go):
```go
const (
	// Synthesis requirements
	MinInputsForSynthesis = 2  // Minimum reasoning inputs for synthesis
)
```

**Impact**: ~5 replacements in integration package

### Total Phase 3 Impact
- **Files Modified**: 4 (calibration.go, episodic.go, learning.go, synthesizer.go)
- **Constants Added**: ~25 well-named constants
- **Magic Numbers Replaced**: ~42 occurrences
- **Lines Changed**: ~50 lines total

### Phase 3 Execution Strategy
1. **Phase 3a**: Extract validation constants (commit)
2. **Phase 3b**: Extract memory constants (commit)
3. **Phase 3c**: Extract integration constants (commit)
4. **Verify**: Run all tests after each phase
5. **Pattern**: Follow server/validation.go as example

---

## Phase 4: Documentation (Week 6)

### Objective
Add comprehensive godoc comments to exported functions and types in key packages.

### Documentation Gaps Analysis

**Packages Needing Enhancement**:
1. `internal/validation` - Good package docs, but individual functions need enhancement
2. `internal/memory` - Good package docs, exported functions need enhancement
3. `internal/integration` - Minimal package docs, needs comprehensive update
4. `internal/analysis` - Needs function-level documentation
5. `internal/reasoning` - Needs consistent documentation style

### Documentation Standards

**Package-Level Documentation** (already good in most packages):
```go
// Package [name] provides [concise description].
//
// [Detailed description of package purpose and capabilities]
//
// Key features:
//   - Feature 1
//   - Feature 2
//   - Feature 3
package name
```

**Function-Level Documentation** (to be added):
```go
// FunctionName performs [action] and returns [result].
//
// This function [detailed explanation of what it does, why it exists,
// and any important behavior or edge cases].
//
// Parameters:
//   - param1: Description of parameter 1
//   - param2: Description of parameter 2
//
// Returns:
//   - Description of return value(s)
//   - Error conditions and what they mean
//
// Example:
//   result, err := FunctionName(arg1, arg2)
//   if err != nil {
//       // Handle error
//   }
func FunctionName(param1 Type1, param2 Type2) (Result, error) {
    // implementation
}
```

**Type Documentation** (to be enhanced):
```go
// TypeName represents [what this type models].
//
// [Detailed description of the type's purpose, when to use it,
// and any important constraints or invariants].
//
// Fields:
//   - Field1: Purpose and constraints
//   - Field2: Purpose and constraints
type TypeName struct {
    Field1 Type1  `json:"field1"`  // Inline comment for clarity
    Field2 Type2  `json:"field2"`  // Inline comment for clarity
}
```

### Phase 4 Priorities

**Priority 1: Exported Functions in Core Packages**
- [ ] `validation/calibration.go` - CalibrationTracker methods
- [ ] `memory/episodic.go` - EpisodicMemory methods
- [ ] `memory/learning.go` - PatternLearner methods
- [ ] `integration/synthesizer.go` - Synthesizer methods

**Priority 2: Exported Types**
- [ ] All exported structs in validation, memory, integration packages
- [ ] Ensure all fields have inline comments explaining purpose

**Priority 3: Package-Level Documentation Enhancement**
- [ ] `integration` package - needs comprehensive overview
- [ ] `analysis` package - needs usage examples
- [ ] `reasoning` package - needs architecture overview

### Phase 4 Execution Strategy

**Phase 4a: Validation Package Documentation**
1. Enhance CalibrationTracker exported methods
2. Document all exported types (CalibrationReport, CalibrationBias, etc.)
3. Add usage examples for common scenarios
4. Commit: "docs: enhance validation package documentation"

**Phase 4b: Memory Package Documentation**
1. Enhance EpisodicMemory exported methods
2. Enhance PatternLearner exported methods
3. Document trajectory storage and retrieval patterns
4. Add examples of episodic learning workflows
5. Commit: "docs: enhance memory package documentation"

**Phase 4c: Integration Package Documentation**
1. Enhance Synthesizer exported methods
2. Add comprehensive package-level documentation
3. Document synthesis patterns and use cases
4. Commit: "docs: enhance integration package documentation"

**Phase 4d: Analysis & Reasoning Packages**
1. Document exported functions in analysis package
2. Document exported functions in reasoning package
3. Commit: "docs: enhance analysis and reasoning package documentation"

### Documentation Quality Metrics

**Target Standards**:
- [ ] 100% of exported functions have godoc comments
- [ ] 100% of exported types have godoc comments
- [ ] Package-level documentation explains purpose and usage
- [ ] All comments follow Go documentation conventions
- [ ] Examples provided for complex functions
- [ ] Error conditions documented

**Validation**:
```bash
# Check documentation coverage
go doc -all internal/validation | grep "FUNCTIONS" -A 100
go doc -all internal/memory | grep "FUNCTIONS" -A 100
go doc -all internal/integration | grep "FUNCTIONS" -A 100

# Generate documentation HTML for review
godoc -http=:6060
```

---

## Success Criteria

### Phase 3 Complete When:
- [ ] All magic numbers in validation package extracted to constants
- [ ] All magic numbers in memory package extracted to constants
- [ ] All magic numbers in integration package extracted to constants
- [ ] All tests passing with zero failures
- [ ] Constants have clear, semantic names following Go conventions
- [ ] Code review shows improved readability

### Phase 4 Complete When:
- [ ] All exported functions in target packages have comprehensive godoc
- [ ] All exported types have clear documentation
- [ ] Package-level documentation is complete and accurate
- [ ] godoc output is professional and helpful
- [ ] Examples provided for complex functions
- [ ] Documentation follows Go best practices

---

## Timeline Estimate

**Phase 3: Polish** (Estimated: 3-4 hours)
- Phase 3a (Validation): 1 hour
- Phase 3b (Memory): 1.5 hours
- Phase 3c (Integration): 0.5 hours
- Testing & Verification: 0.5-1 hour

**Phase 4: Documentation** (Estimated: 4-5 hours)
- Phase 4a (Validation): 1 hour
- Phase 4b (Memory): 1.5 hours
- Phase 4c (Integration): 1 hour
- Phase 4d (Analysis/Reasoning): 1.5 hours
- Review & Polish: 0.5-1 hour

**Total Estimated Time**: 7-9 hours

---

## Risks and Mitigations

### Phase 3 Risks
- **Risk**: Breaking tests by incorrectly extracting constants
  - **Mitigation**: Run tests after each file, verify semantics carefully

- **Risk**: Choosing poor constant names
  - **Mitigation**: Follow existing patterns in server/validation.go, use descriptive names

### Phase 4 Risks
- **Risk**: Documentation becomes outdated quickly
  - **Mitigation**: Focus on "what" and "why", not implementation details

- **Risk**: Time-consuming for marginal benefit
  - **Mitigation**: Prioritize exported functions, skip internal helpers

---

## References

**Existing Good Examples**:
- `internal/server/validation.go` - Excellent constant definitions with comments
- `internal/memory/episodic.go` - Good package-level documentation
- `internal/validation/fallacies.go` - Good type documentation

**Go Documentation Standards**:
- https://go.dev/doc/effective_go#commentary
- https://go.dev/blog/godoc

---

**Generated**: 2025-11-16
**Author**: Claude Code with unified-thinking analysis
**Status**: Ready for execution
