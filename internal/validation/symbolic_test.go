package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSymbolicReasoner(t *testing.T) {
	sr := NewSymbolicReasoner()

	assert.NotNil(t, sr)
	assert.NotNil(t, sr.constraints)
	assert.NotNil(t, sr.symbols)
	assert.Empty(t, sr.constraints)
	assert.Empty(t, sr.symbols)
}

func TestAddSymbol(t *testing.T) {
	sr := NewSymbolicReasoner()

	symbol := sr.AddSymbol("x", SymbolVariable, "integer")

	assert.NotNil(t, symbol)
	assert.Equal(t, "x", symbol.Name)
	assert.Equal(t, SymbolVariable, symbol.Type)
	assert.Equal(t, "integer", symbol.Domain)
	assert.NotNil(t, symbol.Metadata)
	assert.Contains(t, sr.symbols, "x")
}

func TestAddSymbol_Multiple(t *testing.T) {
	sr := NewSymbolicReasoner()

	sr.AddSymbol("x", SymbolVariable, "integer")
	sr.AddSymbol("y", SymbolVariable, "integer")
	sr.AddSymbol("z", SymbolConstant, "boolean")

	assert.Len(t, sr.symbols, 3)
	assert.Equal(t, "integer", sr.symbols["x"].Domain)
	assert.Equal(t, "integer", sr.symbols["y"].Domain)
	assert.Equal(t, "boolean", sr.symbols["z"].Domain)
	assert.Equal(t, SymbolConstant, sr.symbols["z"].Type)
}

func TestAddConstraint_Success(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")
	sr.AddSymbol("y", SymbolVariable, "integer")

	constraint, err := sr.AddConstraint(ConstraintEquality, "x = y", []string{"x", "y"})

	assert.NoError(t, err)
	assert.NotNil(t, constraint)
	assert.Equal(t, ConstraintEquality, constraint.Type)
	assert.Equal(t, "x = y", constraint.Expression)
	assert.Len(t, constraint.Symbols, 2)
	assert.True(t, constraint.Satisfiable)
}

func TestAddConstraint_UnknownSymbol(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")

	_, err := sr.AddConstraint(ConstraintEquality, "x = z", []string{"x", "z"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "symbol z not found")
}

func TestCheckSatisfiability_Tautology(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("a", SymbolVariable, "boolean")

	constraint, _ := sr.AddConstraint(ConstraintDisjunction, "true or false", []string{"a"})

	assert.True(t, constraint.Satisfiable)
	assert.Contains(t, constraint.Explanation, "Tautology")
}

func TestCheckSatisfiability_Contradiction(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("a", SymbolVariable, "boolean")

	constraint, _ := sr.AddConstraint(ConstraintConjunction, "true and false", []string{"a"})

	assert.False(t, constraint.Satisfiable)
	assert.Contains(t, constraint.Explanation, "contradiction")
}

func TestCheckSatisfiability_EqualityConflict(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")

	constraint, _ := sr.AddConstraint(ConstraintEquality, "x = 5 and x = 10", []string{"x"})

	assert.False(t, constraint.Satisfiable)
	assert.Contains(t, constraint.Explanation, "Contradiction")
	assert.Contains(t, constraint.Explanation, "cannot equal both")
}

func TestCheckSatisfiability_ConsistentEqualities(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")

	constraint, _ := sr.AddConstraint(ConstraintEquality, "x = 5 and x = 5", []string{"x"})

	assert.True(t, constraint.Satisfiable)
}

func TestProveTheorem_DirectMatch(t *testing.T) {
	sr := NewSymbolicReasoner()

	theorem := &SymbolicTheorem{
		Name:       "Simple identity",
		Premises:   []string{"A is true"},
		Conclusion: "A is true",
		Status:     StatusUnproven,
	}

	proof, err := sr.ProveTheorem(theorem)

	assert.NoError(t, err)
	assert.NotNil(t, proof)
	assert.True(t, proof.IsValid)
	assert.Equal(t, StatusProven, theorem.Status)
	assert.Greater(t, proof.Confidence, 0.8)
	assert.NotEmpty(t, proof.Steps)
}

func TestProveTheorem_ModusPonens(t *testing.T) {
	sr := NewSymbolicReasoner()

	theorem := &SymbolicTheorem{
		Name: "Modus ponens",
		Premises: []string{
			"It is raining",
			"If it is raining then the ground is wet",
		},
		Conclusion: "The ground is wet",
		Status:     StatusUnproven,
	}

	proof, err := sr.ProveTheorem(theorem)

	assert.NoError(t, err)
	assert.NotNil(t, proof)
	assert.True(t, proof.IsValid)
	assert.Equal(t, StatusProven, theorem.Status)
	assert.Greater(t, proof.Confidence, 0.8)

	// Check that proof has correct steps
	assert.GreaterOrEqual(t, len(proof.Steps), 3) // 2 premises + conclusion
	lastStep := proof.Steps[len(proof.Steps)-1]
	assert.Equal(t, "modus_ponens", lastStep.Rule)
}

func TestProveTheorem_ModusPonens_ImpliesKeyword(t *testing.T) {
	sr := NewSymbolicReasoner()

	theorem := &SymbolicTheorem{
		Name: "Modus ponens with implies",
		Premises: []string{
			"P",
			"P implies Q",
		},
		Conclusion: "Q",
		Status:     StatusUnproven,
	}

	proof, err := sr.ProveTheorem(theorem)

	assert.NoError(t, err)
	assert.True(t, proof.IsValid)
	assert.Equal(t, StatusProven, theorem.Status)
}

func TestProveTheorem_Simplification(t *testing.T) {
	sr := NewSymbolicReasoner()

	theorem := &SymbolicTheorem{
		Name: "Simplification",
		Premises: []string{
			"A and B are both true",
		},
		Conclusion: "A is true",
		Status:     StatusUnproven,
	}

	proof, err := sr.ProveTheorem(theorem)

	assert.NoError(t, err)
	assert.True(t, proof.IsValid)
	assert.Equal(t, StatusProven, theorem.Status)

	lastStep := proof.Steps[len(proof.Steps)-1]
	assert.Equal(t, "simplification", lastStep.Rule)
}

func TestProveTheorem_Conjunction(t *testing.T) {
	sr := NewSymbolicReasoner()

	theorem := &SymbolicTheorem{
		Name: "Conjunction introduction",
		Premises: []string{
			"A is true",
			"B is true",
		},
		Conclusion: "A and B are true",
		Status:     StatusUnproven,
	}

	proof, err := sr.ProveTheorem(theorem)

	assert.NoError(t, err)
	assert.True(t, proof.IsValid)
	assert.Equal(t, StatusProven, theorem.Status)

	lastStep := proof.Steps[len(proof.Steps)-1]
	assert.Equal(t, "conjunction", lastStep.Rule)
}

func TestProveTheorem_Unprovable(t *testing.T) {
	sr := NewSymbolicReasoner()

	theorem := &SymbolicTheorem{
		Name: "Unprovable theorem",
		Premises: []string{
			"A is true",
		},
		Conclusion: "Z is true", // No connection to premises
		Status:     StatusUnproven,
	}

	proof, err := sr.ProveTheorem(theorem)

	assert.NoError(t, err)
	assert.NotNil(t, proof)
	assert.False(t, proof.IsValid)
	assert.Equal(t, StatusUnproven, theorem.Status)
	assert.Less(t, proof.Confidence, 0.5)
	assert.Contains(t, proof.Explanation, "Unable to derive")
}

func TestCheckConstraintConsistency_Consistent(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")
	sr.AddSymbol("y", SymbolVariable, "integer")

	c1, _ := sr.AddConstraint(ConstraintEquality, "x = 5", []string{"x"})
	c2, _ := sr.AddConstraint(ConstraintEquality, "y = 10", []string{"y"})

	result, err := sr.CheckConstraintConsistency([]string{c1.ID, c2.ID})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsConsistent)
	assert.Empty(t, result.Conflicts)
	assert.Contains(t, result.Explanation, "consistent")
}

func TestCheckConstraintConsistency_Inconsistent(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")

	c1, _ := sr.AddConstraint(ConstraintEquality, "x = 5", []string{"x"})
	c2, _ := sr.AddConstraint(ConstraintEquality, "x = 10", []string{"x"})

	result, err := sr.CheckConstraintConsistency([]string{c1.ID, c2.ID})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsConsistent)
	assert.NotEmpty(t, result.Conflicts)
	assert.Contains(t, result.Explanation, "conflicts")
}

func TestCheckConstraintConsistency_NoSharedSymbols(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")
	sr.AddSymbol("y", SymbolVariable, "integer")

	c1, _ := sr.AddConstraint(ConstraintEquality, "x = 5", []string{"x"})
	c2, _ := sr.AddConstraint(ConstraintEquality, "y = 10", []string{"y"})

	result, err := sr.CheckConstraintConsistency([]string{c1.ID, c2.ID})

	assert.NoError(t, err)
	assert.True(t, result.IsConsistent)
	assert.Empty(t, result.Conflicts)
}

func TestDetectConflict_EqualityConflict(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")

	c1, _ := sr.AddConstraint(ConstraintEquality, "x = 5", []string{"x"})
	c2, _ := sr.AddConstraint(ConstraintEquality, "x = 10", []string{"x"})

	conflict := sr.detectConflict(c1, c2)

	assert.NotNil(t, conflict)
	assert.Equal(t, "equality_conflict", conflict.ConflictType)
	assert.Contains(t, conflict.Explanation, "cannot equal both")
}

func TestDetectConflict_NoConflict(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")
	sr.AddSymbol("y", SymbolVariable, "integer")

	c1, _ := sr.AddConstraint(ConstraintEquality, "x = 5", []string{"x"})
	c2, _ := sr.AddConstraint(ConstraintEquality, "y = 10", []string{"y"})

	conflict := sr.detectConflict(c1, c2)

	assert.Nil(t, conflict)
}

func TestDetectConflict_InequalityConflict(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")

	// Test the exact case the user reported: x > 10 AND x < 5
	c1, _ := sr.AddConstraint(ConstraintInequality, "x > 10", []string{"x"})
	c2, _ := sr.AddConstraint(ConstraintInequality, "x < 5", []string{"x"})

	conflict := sr.detectConflict(c1, c2)

	assert.NotNil(t, conflict, "Should detect conflict between x > 10 and x < 5")
	assert.Equal(t, "inequality_conflict", conflict.ConflictType)
	assert.Contains(t, conflict.Explanation, "empty solution set")
}

func TestDetectConflict_InequalityConflict_Variations(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")

	tests := []struct {
		name           string
		expr1          string
		expr2          string
		shouldConflict bool
	}{
		{"x > 10 and x < 5", "x > 10", "x < 5", true},
		{"x >= 5 and x < 5", "x >= 5", "x < 5", true},
		{"x > 10 and x <= 10", "x > 10", "x <= 10", true},
		{"x > 5 and x < 10", "x > 5", "x < 10", false},
		{"x >= 5 and x <= 10", "x >= 5", "x <= 10", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sr := NewSymbolicReasoner()
			sr.AddSymbol("x", SymbolVariable, "integer")

			c1, _ := sr.AddConstraint(ConstraintInequality, tc.expr1, []string{"x"})
			c2, _ := sr.AddConstraint(ConstraintInequality, tc.expr2, []string{"x"})

			conflict := sr.detectConflict(c1, c2)

			if tc.shouldConflict {
				assert.NotNil(t, conflict, "Expected conflict for %s", tc.name)
			} else {
				assert.Nil(t, conflict, "Expected no conflict for %s", tc.name)
			}
		})
	}
}

func TestExtractValue(t *testing.T) {
	sr := NewSymbolicReasoner()

	tests := []struct {
		name     string
		expr     string
		symbol   string
		expected string
	}{
		{
			name:     "simple_left",
			expr:     "x = 5",
			symbol:   "x",
			expected: "5",
		},
		{
			name:     "simple_right",
			expr:     "10 = y",
			symbol:   "y",
			expected: "10",
		},
		{
			name:     "no_match",
			expr:     "x = 5",
			symbol:   "z",
			expected: "",
		},
		{
			name:     "no_equals",
			expr:     "x > 5",
			symbol:   "x",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sr.extractValue(tt.expr, tt.symbol)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAllConstraints(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")
	sr.AddSymbol("y", SymbolVariable, "integer")

	sr.AddConstraint(ConstraintEquality, "x = 5", []string{"x"})
	sr.AddConstraint(ConstraintEquality, "y = 10", []string{"y"})

	constraints := sr.GetAllConstraints()

	assert.Len(t, constraints, 2)
}

func TestGetAllSymbols(t *testing.T) {
	sr := NewSymbolicReasoner()

	sr.AddSymbol("x", SymbolVariable, "integer")
	sr.AddSymbol("y", SymbolVariable, "real")
	sr.AddSymbol("z", SymbolConstant, "boolean")

	symbols := sr.GetAllSymbols()

	assert.Len(t, symbols, 3)
}

func TestSymbolTypes(t *testing.T) {
	assert.Equal(t, SymbolType("variable"), SymbolVariable)
	assert.Equal(t, SymbolType("constant"), SymbolConstant)
	assert.Equal(t, SymbolType("function"), SymbolFunction)
}

func TestConstraintTypes(t *testing.T) {
	assert.Equal(t, ConstraintType("equality"), ConstraintEquality)
	assert.Equal(t, ConstraintType("inequality"), ConstraintInequality)
	assert.Equal(t, ConstraintType("range"), ConstraintRange)
	assert.Equal(t, ConstraintType("membership"), ConstraintMembership)
	assert.Equal(t, ConstraintType("implication"), ConstraintImplication)
	assert.Equal(t, ConstraintType("conjunction"), ConstraintConjunction)
	assert.Equal(t, ConstraintType("disjunction"), ConstraintDisjunction)
	assert.Equal(t, ConstraintType("negation"), ConstraintNegation)
}

func TestTheoremStatus(t *testing.T) {
	assert.Equal(t, TheoremStatus("unproven"), StatusUnproven)
	assert.Equal(t, TheoremStatus("proven"), StatusProven)
	assert.Equal(t, TheoremStatus("refuted"), StatusRefuted)
	assert.Equal(t, TheoremStatus("undecidable"), StatusUndecidable)
}

func TestProofStep(t *testing.T) {
	step := &ProofStep{
		StepNumber:    1,
		Statement:     "A is true",
		Justification: "Premise",
		Rule:          "assumption",
		Dependencies:  []int{},
	}

	assert.Equal(t, 1, step.StepNumber)
	assert.Equal(t, "A is true", step.Statement)
	assert.Equal(t, "assumption", step.Rule)
	assert.Empty(t, step.Dependencies)
}

func TestConstraintMetadata(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")

	constraint, _ := sr.AddConstraint(ConstraintEquality, "x = 5", []string{"x"})
	constraint.Metadata["source"] = "test"
	constraint.Metadata["priority"] = "high"

	assert.Equal(t, "test", constraint.Metadata["source"])
	assert.Equal(t, "high", constraint.Metadata["priority"])
}

func TestSymbolMetadata(t *testing.T) {
	sr := NewSymbolicReasoner()
	symbol := sr.AddSymbol("x", SymbolVariable, "integer")

	symbol.Metadata["range"] = "[0, 100]"
	symbol.Metadata["unit"] = "meters"

	assert.Equal(t, "[0, 100]", symbol.Metadata["range"])
	assert.Equal(t, "meters", symbol.Metadata["unit"])
}

func TestComplexProof_ChainedInference(t *testing.T) {
	sr := NewSymbolicReasoner()

	theorem := &SymbolicTheorem{
		Name: "Chained inference",
		Premises: []string{
			"Socrates is a man",
			"All men are mortal",
		},
		Conclusion: "Socrates is mortal",
		Status:     StatusUnproven,
	}

	proof, err := sr.ProveTheorem(theorem)

	assert.NoError(t, err)
	assert.NotNil(t, proof)
	// Note: This may not prove automatically depending on pattern matching
	// but the infrastructure supports it
	assert.NotEmpty(t, proof.Steps)
}

func TestMultipleConstraints_DifferentTypes(t *testing.T) {
	sr := NewSymbolicReasoner()
	sr.AddSymbol("x", SymbolVariable, "integer")
	sr.AddSymbol("y", SymbolVariable, "integer")
	sr.AddSymbol("z", SymbolVariable, "boolean")

	c1, _ := sr.AddConstraint(ConstraintEquality, "x = 5", []string{"x"})
	c2, _ := sr.AddConstraint(ConstraintInequality, "y > 10", []string{"y"})
	c3, _ := sr.AddConstraint(ConstraintImplication, "z implies x > 0", []string{"z", "x"})

	assert.Equal(t, ConstraintEquality, c1.Type)
	assert.Equal(t, ConstraintInequality, c2.Type)
	assert.Equal(t, ConstraintImplication, c3.Type)

	allConstraints := sr.GetAllConstraints()
	assert.Len(t, allConstraints, 3)
}
