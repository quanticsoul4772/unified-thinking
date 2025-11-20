# Bayesian Inference Correctness Specification

## Overview

This document specifies the mathematical requirements for Bayesian belief updating in the unified-thinking server, explaining why the original `UpdateBelief` method was removed and replaced with `UpdateBeliefFull` and `UpdateBeliefWithEvidence`.

## Mathematical Foundation

### Bayes' Theorem

The correct form of Bayes' theorem for updating beliefs given evidence is:

```
P(H|E) = P(E|H) × P(H) / P(E)
```

Where:
- `P(H|E)` = Posterior probability (updated belief after seeing evidence)
- `P(E|H)` = Likelihood (probability of evidence if hypothesis is true)
- `P(H)` = Prior probability (belief before seeing evidence)
- `P(E)` = Evidence probability (total probability of seeing this evidence)

### Expanded Form

Since `P(E)` must be calculated using the law of total probability:

```
P(E) = P(E|H) × P(H) + P(E|¬H) × P(¬H)
```

The complete formula becomes:

```
P(H|E) = P(E|H) × P(H) / [P(E|H) × P(H) + P(E|¬H) × P(¬H)]
```

**Critical Insight**: This formula requires **both** `P(E|H)` and `P(E|¬H)` for correct calculation.

## Why Two Likelihoods Are Required

### The Medical Test Example

Consider a medical test with:
- Disease prevalence: 1% (P(disease) = 0.01)
- Sensitivity: 99% (P(positive|disease) = 0.99)
- Specificity: 95% (P(negative|healthy) = 0.95, so P(positive|healthy) = 0.05)

If a patient tests positive, what's the probability they have the disease?

**Intuitive (Wrong) Answer**: 99% (the test sensitivity)

**Correct Answer**:
```
P(disease|positive) = 0.99 × 0.01 / [0.99 × 0.01 + 0.05 × 0.99]
                    = 0.0099 / [0.0099 + 0.0495]
                    = 0.0099 / 0.0594
                    ≈ 0.167 (16.7%)
```

The positive test only gives a **16.7% probability** of disease, not 99%!

### Why This Matters

Without `P(E|¬H)`, we cannot:
1. Properly normalize the posterior probability
2. Account for how common the evidence is under alternative hypotheses
3. Distinguish between strong and weak evidence

## Original Implementation Flaw

### Deprecated Method (REMOVED)

```go
// INCORRECT - DO NOT USE
func UpdateBelief(beliefID, evidenceID string, likelihood, evidenceProb float64) {
    // This was mathematically incorrect because:
    // 1. "likelihood" was ambiguous - P(E|H) or P(E|¬H)?
    // 2. "evidenceProb" conflated P(E) with one of the likelihoods
    // 3. Cannot properly calculate posterior without both likelihoods
}
```

### Why It Was Wrong

The ambiguous parameters made it impossible to correctly apply Bayes' theorem:
- `likelihood` could be interpreted as either P(E|H) or P(E|¬H)
- `evidenceProb` confused the marginal probability P(E) with conditional likelihoods
- No way to compute the denominator correctly

## Correct Implementation

### UpdateBeliefFull (Primary Method)

```go
func UpdateBeliefFull(
    beliefID string,
    evidenceID string,
    likelihoodIfTrue float64,  // P(E|H) - probability of evidence if hypothesis is true
    likelihoodIfFalse float64, // P(E|¬H) - probability of evidence if hypothesis is false
) (*ProbabilisticBelief, error)
```

**Implementation** (from `internal/reasoning/probabilistic.go:86-139`):

```go
prior := belief.Probability
priorNot := 1.0 - prior

numerator := likelihoodIfTrue * prior
denominator := (likelihoodIfTrue * prior) + (likelihoodIfFalse * priorNot)

var posterior float64
if denominator > 0 {
    posterior = numerator / denominator
} else {
    posterior = prior // No update if denominator is zero
}
```

### UpdateBeliefWithEvidence (Helper Method)

```go
func UpdateBeliefWithEvidence(
    beliefID string,
    evidence *types.Evidence,
) (*ProbabilisticBelief, error)
```

This method estimates both likelihoods from evidence quality scores:

```go
if evidence.SupportsClaim {
    // Evidence supports the belief
    likelihoodIfTrue = 0.5 + (evidence.OverallScore * 0.4)  // Range: 0.5-0.9
    likelihoodIfFalse = 0.5 - (evidence.OverallScore * 0.3) // Range: 0.2-0.5
} else {
    // Evidence refutes the belief
    likelihoodIfTrue = 0.5 - (evidence.OverallScore * 0.4)  // Range: 0.1-0.5
    likelihoodIfFalse = 0.5 + (evidence.OverallScore * 0.3) // Range: 0.5-0.8
}
```

Then calls `UpdateBeliefFull` with the estimated likelihoods.

## API Design Rationale

### Parameter Naming

- `likelihoodIfTrue` and `likelihoodIfFalse` are **explicit and unambiguous**
- Clearly distinguish the two conditional probabilities required
- Self-documenting: no need to look up what "likelihood" means

### Tool Interface

The `probabilistic-reasoning` tool (updated in commit 687d1af) now requires:

```json
{
  "operation": "update",
  "belief_id": "belief_123",
  "evidence_id": "ev_456",
  "likelihood_if_true": 0.8,
  "likelihood_if_false": 0.2
}
```

This forces users to think about both conditional probabilities, preventing mathematical errors.

## Validation and Edge Cases

### Uninformative Evidence Detection

From `internal/reasoning/probabilistic.go:103-115`:

```go
// Check for degenerate cases where both likelihoods are identical
if math.Abs(likelihoodIfTrue-likelihoodIfFalse) < 1e-10 {
    // Evidence is equally likely regardless of hypothesis truth
    // No update needed - posterior equals prior
    belief.Metadata["last_update_uninformative"] = true
    return belief, nil
}
```

If `P(E|H) = P(E|¬H)`, the evidence provides no information and should not change beliefs.

### Zero Denominator Handling

```go
if denominator > 0 {
    posterior = numerator / denominator
} else {
    posterior = prior // No update if denominator is zero
}
```

Prevents division by zero in edge cases.

### Probability Range Clamping

```go
posterior = math.Max(0, math.Min(1, posterior))
```

Ensures posterior probability stays in valid [0,1] range despite floating-point errors.

## Test Coverage

Updated test files (commit 687d1af):
- `internal/reasoning/probabilistic_test.go` - Core logic tests
- `internal/reasoning/probabilistic_validation_test.go` - Edge case validation
- `internal/reasoning/concurrent_test.go` - Thread safety tests
- `internal/server/handlers/probabilistic_test.go` - API integration tests

All tests verify:
1. Correct posterior calculation with both likelihoods
2. Uninformative evidence detection
3. Edge case handling (zero denominator, identical likelihoods)
4. API parameter validation

## Migration Guide

### Old Code (DO NOT USE)

```go
// INCORRECT
pr.UpdateBelief(beliefID, evidenceID, 0.8, 0.5)
```

### New Code (CORRECT)

```go
// Option 1: Explicit likelihoods
pr.UpdateBeliefFull(beliefID, evidenceID, 0.8, 0.2)

// Option 2: Estimate from evidence
evidence := &types.Evidence{
    ID: evidenceID,
    SupportsClaim: true,
    OverallScore: 0.8,
}
pr.UpdateBeliefWithEvidence(beliefID, evidence)
```

## References

1. **Pearl, J.** (1988). *Probabilistic Reasoning in Intelligent Systems*. Morgan Kaufmann.
2. **Jaynes, E.T.** (2003). *Probability Theory: The Logic of Science*. Cambridge University Press.
3. **Base Rate Fallacy**: Why ignoring P(E|¬H) leads to incorrect posteriors
4. **Likelihood Ratio**: The ratio P(E|H) / P(E|¬H) determines the strength of evidence

## Conclusion

The refactoring from `UpdateBelief` to `UpdateBeliefFull` addresses a fundamental mathematical correctness issue. By requiring both `likelihoodIfTrue` and `likelihoodIfFalse`, the implementation now:

1. Correctly applies Bayes' theorem with proper normalization
2. Forces users to consider both conditional probabilities
3. Prevents common base rate fallacy errors
4. Provides self-documenting API with unambiguous parameter names

This change is **not** a refactoring preference—it fixes a **mathematical incorrectness** that would have produced invalid probability updates.
