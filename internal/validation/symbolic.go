// Package validation provides symbolic reasoning capabilities.
// This module extends logical validation with symbolic constraint tracking
// and basic theorem proving for enhanced reasoning quality.
package validation

import (
	"fmt"
	"strings"
)

// SymbolicReasoner provides symbolic constraint tracking and theorem proving
type SymbolicReasoner struct {
	constraints map[string]*SymbolicConstraint
	symbols     map[string]*Symbol
}

// NewSymbolicReasoner creates a new symbolic reasoner
func NewSymbolicReasoner() *SymbolicReasoner {
	return &SymbolicReasoner{
		constraints: make(map[string]*SymbolicConstraint),
		symbols:     make(map[string]*Symbol),
	}
}

// Symbol represents a symbolic variable or constant
type Symbol struct {
	Name     string
	Type     SymbolType
	Domain   string // e.g., "integer", "boolean", "real"
	Value    interface{}
	Metadata map[string]interface{}
}

// SymbolType categorizes symbols
type SymbolType string

const (
	SymbolVariable SymbolType = "variable"
	SymbolConstant SymbolType = "constant"
	SymbolFunction SymbolType = "function"
)

// SymbolicConstraint represents a symbolic constraint on variables
type SymbolicConstraint struct {
	ID          string
	Type        ConstraintType
	Expression  string
	Symbols     []string // Symbol names involved
	Satisfiable bool
	Explanation string
	Metadata    map[string]interface{}
}

// ConstraintType categorizes constraints
type ConstraintType string

const (
	ConstraintEquality    ConstraintType = "equality"    // x = y
	ConstraintInequality  ConstraintType = "inequality"  // x < y, x > y
	ConstraintRange       ConstraintType = "range"       // x in [a, b]
	ConstraintMembership  ConstraintType = "membership"  // x in Set
	ConstraintImplication ConstraintType = "implication" // A → B
	ConstraintConjunction ConstraintType = "conjunction" // A ∧ B
	ConstraintDisjunction ConstraintType = "disjunction" // A ∨ B
	ConstraintNegation    ConstraintType = "negation"    // ¬A
)

// SymbolicTheorem represents a theorem to be proved
type SymbolicTheorem struct {
	Name       string
	Premises   []string
	Conclusion string
	Proof      *TheoremProof
	Status     TheoremStatus
}

// TheoremStatus indicates proof status
type TheoremStatus string

const (
	StatusUnproven    TheoremStatus = "unproven"
	StatusProven      TheoremStatus = "proven"
	StatusRefuted     TheoremStatus = "refuted"
	StatusUndecidable TheoremStatus = "undecidable"
)

// TheoremProof contains proof steps and justification
type TheoremProof struct {
	Steps       []*ProofStep
	IsValid     bool
	Method      string
	Confidence  float64
	Explanation string
}

// ProofStep represents a single step in a proof
type ProofStep struct {
	StepNumber    int
	Statement     string
	Justification string
	Rule          string
	Dependencies  []int // Step numbers this depends on
}

// AddSymbol registers a symbolic variable or constant
func (sr *SymbolicReasoner) AddSymbol(name string, symbolType SymbolType, domain string) *Symbol {
	symbol := &Symbol{
		Name:     name,
		Type:     symbolType,
		Domain:   domain,
		Metadata: make(map[string]interface{}),
	}
	sr.symbols[name] = symbol
	return symbol
}

// AddConstraint adds a symbolic constraint
func (sr *SymbolicReasoner) AddConstraint(constraintType ConstraintType, expression string, symbols []string) (*SymbolicConstraint, error) {
	// Validate that all symbols exist
	for _, symName := range symbols {
		if _, exists := sr.symbols[symName]; !exists {
			return nil, fmt.Errorf("symbol %s not found", symName)
		}
	}

	constraint := &SymbolicConstraint{
		ID:          fmt.Sprintf("constraint-%d", len(sr.constraints)+1),
		Type:        constraintType,
		Expression:  expression,
		Symbols:     symbols,
		Satisfiable: true, // Assume satisfiable until proven otherwise
		Metadata:    make(map[string]interface{}),
	}

	// Check basic satisfiability
	constraint.Satisfiable = sr.checkSatisfiability(constraint)

	sr.constraints[constraint.ID] = constraint
	return constraint, nil
}

// checkSatisfiability performs basic satisfiability checking
func (sr *SymbolicReasoner) checkSatisfiability(constraint *SymbolicConstraint) bool {
	expr := strings.ToLower(constraint.Expression)

	// Check for obvious contradictions
	if strings.Contains(expr, "true and false") ||
		strings.Contains(expr, "false and true") {
		constraint.Explanation = "Direct contradiction: conjunction of true and false"
		return false
	}

	// Check for tautologies that are always satisfiable
	if strings.Contains(expr, "true or false") ||
		strings.Contains(expr, "false or true") {
		constraint.Explanation = "Tautology: always satisfiable"
		return true
	}

	// Check for equality contradictions (x = a AND x = b where a ≠ b)
	if constraint.Type == ConstraintEquality {
		// Look for pattern like "x = 5 and x = 10"
		parts := strings.Split(expr, " and ")
		if len(parts) > 1 {
			equalities := make(map[string][]string)
			for _, part := range parts {
				if strings.Contains(part, "=") {
					tokens := strings.Split(part, "=")
					if len(tokens) == 2 {
						varName := strings.TrimSpace(tokens[0])
						value := strings.TrimSpace(tokens[1])
						equalities[varName] = append(equalities[varName], value)
					}
				}
			}

			// Check if any variable has conflicting values
			for varName, values := range equalities {
				if len(values) > 1 {
					// Check if all values are the same
					first := values[0]
					for _, v := range values[1:] {
						if v != first {
							constraint.Explanation = fmt.Sprintf("Contradiction: %s cannot equal both %s and %s", varName, first, v)
							return false
						}
					}
				}
			}
		}
	}

	// Default: assume satisfiable
	return true
}

// ProveTheorem attempts to prove a theorem symbolically
func (sr *SymbolicReasoner) ProveTheorem(theorem *SymbolicTheorem) (*TheoremProof, error) {
	proof := &TheoremProof{
		Steps:      make([]*ProofStep, 0),
		Method:     "natural_deduction",
		Confidence: 0.0,
	}

	stepNum := 1

	// Add premises as initial steps
	for i, premise := range theorem.Premises {
		step := &ProofStep{
			StepNumber:    stepNum,
			Statement:     premise,
			Justification: fmt.Sprintf("Premise %d", i+1),
			Rule:          "assumption",
			Dependencies:  []int{},
		}
		proof.Steps = append(proof.Steps, step)
		stepNum++
	}

	// Try to derive conclusion using inference rules
	derived := sr.attemptDerivation(theorem.Premises, theorem.Conclusion, proof, &stepNum)

	if derived {
		proof.IsValid = true
		proof.Confidence = 0.9
		proof.Explanation = "Conclusion successfully derived from premises"
		theorem.Status = StatusProven
	} else {
		proof.IsValid = false
		proof.Confidence = 0.1
		proof.Explanation = "Unable to derive conclusion from premises with available rules"
		theorem.Status = StatusUnproven
	}

	theorem.Proof = proof
	return proof, nil
}

// attemptDerivation tries to derive conclusion from premises
func (sr *SymbolicReasoner) attemptDerivation(premises []string, conclusion string, proof *TheoremProof, stepNum *int) bool {
	conclusionLower := strings.ToLower(conclusion)

	// Direct premise match
	for _, premise := range premises {
		if strings.ToLower(premise) == conclusionLower {
			step := &ProofStep{
				StepNumber:    *stepNum,
				Statement:     conclusion,
				Justification: "Direct match with premise",
				Rule:          "identity",
				Dependencies:  []int{1},
			}
			proof.Steps = append(proof.Steps, step)
			*stepNum++
			return true
		}
	}

	// Modus ponens: If we have "A" and "A → B", derive "B"
	if sr.tryModusPonensSymbolic(premises, conclusion, proof, stepNum) {
		return true
	}

	// Simplification: From "A ∧ B", derive "A" or "B"
	if sr.trySimplification(premises, conclusion, proof, stepNum) {
		return true
	}

	// Conjunction: If we have "A" and "B", derive "A ∧ B"
	if sr.tryConjunction(premises, conclusion, proof, stepNum) {
		return true
	}

	return false
}

// tryModusPonensSymbolic attempts modus ponens inference
func (sr *SymbolicReasoner) tryModusPonensSymbolic(premises []string, conclusion string, proof *TheoremProof, stepNum *int) bool {
	// Look for pattern: "A" and "A → B" to derive "B"
	for i, p1 := range premises {
		for j, p2 := range premises {
			if i == j {
				continue
			}

			p1Lower := strings.ToLower(strings.TrimSpace(p1))
			p2Lower := strings.ToLower(strings.TrimSpace(p2))
			conclusionLower := strings.ToLower(strings.TrimSpace(conclusion))

			// Check if p2 is an implication
			for _, implWord := range []string{" implies ", " → ", " then ", "if ", " means "} {
				if strings.Contains(p2Lower, implWord) {
					var antecedent, consequent string

					// Handle "if A then B" pattern
					if strings.Contains(p2Lower, "if ") && strings.Contains(p2Lower, " then ") {
						parts := strings.Split(p2Lower, " then ")
						if len(parts) == 2 {
							antecedent = strings.TrimSpace(strings.TrimPrefix(parts[0], "if "))
							consequent = strings.TrimSpace(parts[1])
						}
					} else {
						// Handle "A implies B" pattern
						parts := strings.Split(p2Lower, implWord)
						if len(parts) == 2 {
							antecedent = strings.TrimSpace(parts[0])
							consequent = strings.TrimSpace(parts[1])
						}
					}

					if antecedent != "" && consequent != "" {
						// Normalize whitespace and check for match
						antecedentNorm := strings.Join(strings.Fields(antecedent), " ")
						p1Norm := strings.Join(strings.Fields(p1Lower), " ")
						consequentNorm := strings.Join(strings.Fields(consequent), " ")
						conclusionNorm := strings.Join(strings.Fields(conclusionLower), " ")

						// Check if p1 matches antecedent and conclusion matches consequent
						if strings.Contains(p1Norm, antecedentNorm) && strings.Contains(consequentNorm, conclusionNorm) {
							step := &ProofStep{
								StepNumber:    *stepNum,
								Statement:     conclusion,
								Justification: fmt.Sprintf("Modus ponens from steps %d and %d", i+1, j+1),
								Rule:          "modus_ponens",
								Dependencies:  []int{i + 1, j + 1},
							}
							proof.Steps = append(proof.Steps, step)
							*stepNum++
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// trySimplification attempts simplification rule
func (sr *SymbolicReasoner) trySimplification(premises []string, conclusion string, proof *TheoremProof, stepNum *int) bool {
	// From "A ∧ B", derive "A" or "B"
	conclusionLower := strings.ToLower(conclusion)

	for i, premise := range premises {
		premiseLower := strings.ToLower(premise)

		// Check for conjunction
		for _, conjWord := range []string{" and ", " ∧ "} {
			if strings.Contains(premiseLower, conjWord) {
				parts := strings.Split(premiseLower, conjWord)
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if strings.Contains(part, conclusionLower) || strings.Contains(conclusionLower, part) {
						step := &ProofStep{
							StepNumber:    *stepNum,
							Statement:     conclusion,
							Justification: fmt.Sprintf("Simplification from step %d", i+1),
							Rule:          "simplification",
							Dependencies:  []int{i + 1},
						}
						proof.Steps = append(proof.Steps, step)
						*stepNum++
						return true
					}
				}
			}
		}
	}
	return false
}

// tryConjunction attempts conjunction introduction
func (sr *SymbolicReasoner) tryConjunction(premises []string, conclusion string, proof *TheoremProof, stepNum *int) bool {
	// If conclusion is "A ∧ B", check if we have both "A" and "B"
	conclusionLower := strings.ToLower(strings.TrimSpace(conclusion))

	for _, conjWord := range []string{" and ", " ∧ "} {
		if strings.Contains(conclusionLower, conjWord) {
			parts := strings.Split(conclusionLower, conjWord)
			if len(parts) >= 2 {
				// Extract the key parts (handle "A is true and B is true" or "A and B")
				part1 := strings.TrimSpace(parts[0])
				part2 := strings.TrimSpace(parts[1])

				// Normalize: extract core statements
				part1Core := sr.extractCoreStatement(part1)
				part2Core := sr.extractCoreStatement(part2)

				// Find premises matching both parts
				foundPart1 := -1
				foundPart2 := -1

				for i, premise := range premises {
					premiseLower := strings.ToLower(strings.TrimSpace(premise))
					premiseCore := sr.extractCoreStatement(premiseLower)

					// Check if premise matches either part
					if sr.statementsMatch(premiseCore, part1Core) {
						foundPart1 = i
					}
					if sr.statementsMatch(premiseCore, part2Core) {
						foundPart2 = i
					}
				}

				if foundPart1 >= 0 && foundPart2 >= 0 {
					step := &ProofStep{
						StepNumber:    *stepNum,
						Statement:     conclusion,
						Justification: fmt.Sprintf("Conjunction from steps %d and %d", foundPart1+1, foundPart2+1),
						Rule:          "conjunction",
						Dependencies:  []int{foundPart1 + 1, foundPart2 + 1},
					}
					proof.Steps = append(proof.Steps, step)
					*stepNum++
					return true
				}
			}
		}
	}
	return false
}

// extractCoreStatement extracts the core statement from text
func (sr *SymbolicReasoner) extractCoreStatement(text string) string {
	// Remove common suffixes like "is true", "are true"
	text = strings.TrimSpace(text)
	text = strings.TrimSuffix(text, " is true")
	text = strings.TrimSuffix(text, " are true")
	text = strings.TrimSuffix(text, " true")

	// Normalize whitespace
	return strings.Join(strings.Fields(text), " ")
}

// statementsMatch checks if two statements match (with normalization)
func (sr *SymbolicReasoner) statementsMatch(s1, s2 string) bool {
	s1Norm := strings.Join(strings.Fields(strings.ToLower(s1)), " ")
	s2Norm := strings.Join(strings.Fields(strings.ToLower(s2)), " ")

	// Check for exact match or substring match
	return s1Norm == s2Norm || strings.Contains(s1Norm, s2Norm) || strings.Contains(s2Norm, s1Norm)
}

// CheckConstraintConsistency checks if a set of constraints are mutually consistent
func (sr *SymbolicReasoner) CheckConstraintConsistency(constraintIDs []string) (*ConsistencyResult, error) {
	result := &ConsistencyResult{
		IsConsistent: true,
		Conflicts:    make([]*ConstraintConflict, 0),
		Explanation:  "All constraints are mutually consistent",
	}

	// Get constraints
	constraints := make([]*SymbolicConstraint, 0)
	for _, id := range constraintIDs {
		if c, exists := sr.constraints[id]; exists {
			constraints = append(constraints, c)
		}
	}

	// Check pairwise consistency
	for i := 0; i < len(constraints); i++ {
		for j := i + 1; j < len(constraints); j++ {
			c1 := constraints[i]
			c2 := constraints[j]

			if conflict := sr.detectConflict(c1, c2); conflict != nil {
				result.IsConsistent = false
				result.Conflicts = append(result.Conflicts, conflict)
			}
		}
	}

	if !result.IsConsistent {
		result.Explanation = fmt.Sprintf("Found %d conflicts between constraints", len(result.Conflicts))
	}

	return result, nil
}

// detectConflict checks if two constraints conflict
func (sr *SymbolicReasoner) detectConflict(c1, c2 *SymbolicConstraint) *ConstraintConflict {
	// Check if constraints share symbols
	commonSymbols := make([]string, 0)
	for _, s1 := range c1.Symbols {
		for _, s2 := range c2.Symbols {
			if s1 == s2 {
				commonSymbols = append(commonSymbols, s1)
			}
		}
	}

	if len(commonSymbols) == 0 {
		return nil // No shared symbols = no conflict
	}

	// Check for contradictory constraints on same symbol
	c1Expr := strings.ToLower(c1.Expression)
	c2Expr := strings.ToLower(c2.Expression)

	// Look for pattern: "x = a" vs "x = b" where a ≠ b
	if c1.Type == ConstraintEquality && c2.Type == ConstraintEquality {
		for _, sym := range commonSymbols {
			// Extract values from expressions
			val1 := sr.extractValue(c1Expr, sym)
			val2 := sr.extractValue(c2Expr, sym)

			if val1 != "" && val2 != "" && val1 != val2 {
				return &ConstraintConflict{
					Constraint1:  c1.ID,
					Constraint2:  c2.ID,
					ConflictType: "equality_conflict",
					Explanation:  fmt.Sprintf("Symbol %s cannot equal both %s and %s", sym, val1, val2),
				}
			}
		}
	}

	// Check for contradictory inequalities
	if c1.Type == ConstraintInequality && c2.Type == ConstraintInequality {
		for _, sym := range commonSymbols {
			// Parse inequality constraints
			ineq1 := sr.parseInequality(c1Expr, sym)
			ineq2 := sr.parseInequality(c2Expr, sym)

			if ineq1 != nil && ineq2 != nil {
				// Check if constraints are mutually exclusive
				if conflict := sr.checkInequalityConflict(ineq1, ineq2, sym); conflict != "" {
					return &ConstraintConflict{
						Constraint1:  c1.ID,
						Constraint2:  c2.ID,
						ConflictType: "inequality_conflict",
						Explanation:  conflict,
					}
				}
			}
		}
	}

	return nil
}

// extractValue attempts to extract value from an equality expression
func (sr *SymbolicReasoner) extractValue(expr, symbol string) string {
	if !strings.Contains(expr, "=") {
		return ""
	}

	parts := strings.Split(expr, "=")
	if len(parts) != 2 {
		return ""
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	if strings.Contains(left, symbol) {
		return right
	}
	if strings.Contains(right, symbol) {
		return left
	}

	return ""
}

// InequalityConstraint represents parsed inequality
type InequalityConstraint struct {
	Symbol   string
	Operator string // ">", "<", ">=", "<="
	Value    float64
}

// parseInequality parses an inequality expression
func (sr *SymbolicReasoner) parseInequality(expr, symbol string) *InequalityConstraint {
	// Try to parse patterns like "x > 10", "x < 5", etc.
	var operator string
	var parts []string

	if strings.Contains(expr, ">=") {
		operator = ">="
		parts = strings.Split(expr, ">=")
	} else if strings.Contains(expr, "<=") {
		operator = "<="
		parts = strings.Split(expr, "<=")
	} else if strings.Contains(expr, ">") {
		operator = ">"
		parts = strings.Split(expr, ">")
	} else if strings.Contains(expr, "<") {
		operator = "<"
		parts = strings.Split(expr, "<")
	} else {
		return nil
	}

	if len(parts) != 2 {
		return nil
	}

	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	// Determine which side has the symbol
	var valueStr string
	if strings.Contains(left, symbol) {
		valueStr = right
	} else if strings.Contains(right, symbol) {
		valueStr = left
		// Flip operator if symbol is on right side
		switch operator {
		case ">":
			operator = "<"
		case "<":
			operator = ">"
		case ">=":
			operator = "<="
		case "<=":
			operator = ">="
		}
	} else {
		return nil
	}

	// Parse numeric value (simple implementation)
	var value float64
	_, err := fmt.Sscanf(valueStr, "%f", &value)
	if err != nil {
		return nil
	}

	return &InequalityConstraint{
		Symbol:   symbol,
		Operator: operator,
		Value:    value,
	}
}

// checkInequalityConflict checks if two inequalities conflict
func (sr *SymbolicReasoner) checkInequalityConflict(ineq1, ineq2 *InequalityConstraint, symbol string) string {
	// Check for obvious conflicts like x > 10 AND x < 5
	if ineq1.Operator == ">" && ineq2.Operator == "<" {
		if ineq1.Value >= ineq2.Value {
			return fmt.Sprintf("Symbol %s cannot be both > %.0f and < %.0f (empty solution set)",
				symbol, ineq1.Value, ineq2.Value)
		}
	}
	if ineq1.Operator == "<" && ineq2.Operator == ">" {
		if ineq2.Value >= ineq1.Value {
			return fmt.Sprintf("Symbol %s cannot be both < %.0f and > %.0f (empty solution set)",
				symbol, ineq1.Value, ineq2.Value)
		}
	}

	// Check for >= and < conflicts
	if ineq1.Operator == ">=" && ineq2.Operator == "<" {
		if ineq1.Value >= ineq2.Value {
			return fmt.Sprintf("Symbol %s cannot be both >= %.0f and < %.0f (empty solution set)",
				symbol, ineq1.Value, ineq2.Value)
		}
	}
	if ineq1.Operator == "<" && ineq2.Operator == ">=" {
		if ineq2.Value >= ineq1.Value {
			return fmt.Sprintf("Symbol %s cannot be both < %.0f and >= %.0f (empty solution set)",
				symbol, ineq1.Value, ineq2.Value)
		}
	}

	// Check for <= and > conflicts
	if ineq1.Operator == "<=" && ineq2.Operator == ">" {
		if ineq1.Value <= ineq2.Value {
			return fmt.Sprintf("Symbol %s cannot be both <= %.0f and > %.0f (empty solution set)",
				symbol, ineq1.Value, ineq2.Value)
		}
	}
	if ineq1.Operator == ">" && ineq2.Operator == "<=" {
		if ineq2.Value <= ineq1.Value {
			return fmt.Sprintf("Symbol %s cannot be both > %.0f and <= %.0f (empty solution set)",
				symbol, ineq1.Value, ineq2.Value)
		}
	}

	return ""
}

// ConsistencyResult contains constraint consistency check results
type ConsistencyResult struct {
	IsConsistent bool
	Conflicts    []*ConstraintConflict
	Explanation  string
}

// ConstraintConflict represents a conflict between constraints
type ConstraintConflict struct {
	Constraint1  string
	Constraint2  string
	ConflictType string
	Explanation  string
}

// GetAllConstraints returns all registered constraints
func (sr *SymbolicReasoner) GetAllConstraints() []*SymbolicConstraint {
	constraints := make([]*SymbolicConstraint, 0, len(sr.constraints))
	for _, c := range sr.constraints {
		constraints = append(constraints, c)
	}
	return constraints
}

// GetAllSymbols returns all registered symbols
func (sr *SymbolicReasoner) GetAllSymbols() []*Symbol {
	symbols := make([]*Symbol, 0, len(sr.symbols))
	for _, s := range sr.symbols {
		symbols = append(symbols, s)
	}
	return symbols
}
