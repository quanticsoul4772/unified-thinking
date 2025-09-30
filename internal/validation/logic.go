package validation

import (
	"fmt"
	"strings"

	"unified-thinking/internal/types"
)

// LogicValidator implements logical validation
type LogicValidator struct{}

// NewLogicValidator creates a new logic validator
func NewLogicValidator() *LogicValidator {
	return &LogicValidator{}
}

// ValidateThought validates a thought for logical consistency
func (v *LogicValidator) ValidateThought(thought *types.Thought) (*types.Validation, error) {
	// Simplified validation - in production, use proper logic engine
	isValid := v.checkBasicLogic(thought.Content)

	validation := &types.Validation{
		ThoughtID: thought.ID,
		IsValid:   isValid,
		Reason:    v.getValidationReason(isValid, thought.Content),
	}

	return validation, nil
}

// Prove attempts to prove a conclusion from premises
func (v *LogicValidator) Prove(premises []string, conclusion string) *ProofResult {
	// Simplified proof - in production, use proper theorem prover
	result := &ProofResult{
		Premises:   premises,
		Conclusion: conclusion,
		IsProvable: v.simpleProof(premises, conclusion),
		Steps:      v.generateProofSteps(premises, conclusion),
	}

	return result
}

// CheckWellFormed validates statement syntax
func (v *LogicValidator) CheckWellFormed(statements []string) []StatementCheck {
	checks := make([]StatementCheck, len(statements))
	for i, stmt := range statements {
		checks[i] = StatementCheck{
			Statement:    stmt,
			IsWellFormed: v.checkSyntax(stmt),
			Issues:       v.getSyntaxIssues(stmt),
		}
	}
	return checks
}

func (v *LogicValidator) checkBasicLogic(content string) bool {
	// Simplified - check for basic contradictions
	lower := strings.ToLower(content)

	// Check for obvious contradictions
	if strings.Contains(lower, "always") && strings.Contains(lower, "never") {
		return false
	}

	if strings.Contains(lower, "all") && strings.Contains(lower, "none") {
		return false
	}

	if strings.Contains(lower, "impossible") && strings.Contains(lower, "must happen") {
		return false
	}

	return true
}

func (v *LogicValidator) simpleProof(premises []string, conclusion string) bool {
	// Simplified proof checking
	// Check if conclusion is directly stated in premises
	conclusionLower := strings.ToLower(conclusion)
	for _, premise := range premises {
		if strings.Contains(strings.ToLower(premise), conclusionLower) {
			return true
		}
	}

	// Check for basic logical patterns
	// In production, implement proper formal logic
	return len(premises) > 0 && len(conclusion) > 0
}

func (v *LogicValidator) generateProofSteps(premises []string, conclusion string) []string {
	steps := []string{"Given premises:"}
	for i, p := range premises {
		steps = append(steps, fmt.Sprintf("%d. %s", i+1, p))
	}
	steps = append(steps, fmt.Sprintf("Conclusion: %s", conclusion))
	return steps
}

func (v *LogicValidator) checkSyntax(statement string) bool {
	// Simplified syntax check
	trimmed := strings.TrimSpace(statement)
	
	if len(trimmed) == 0 {
		return false
	}

	// Check for basic statement structure
	if !strings.Contains(trimmed, " ") {
		return false
	}

	return true
}

func (v *LogicValidator) getSyntaxIssues(statement string) []string {
	issues := []string{}
	trimmed := strings.TrimSpace(statement)

	if len(trimmed) == 0 {
		issues = append(issues, "Statement is empty")
	}

	if !strings.Contains(trimmed, " ") {
		issues = append(issues, "Statement appears to be a single word")
	}

	return issues
}

func (v *LogicValidator) getValidationReason(isValid bool, content string) string {
	if isValid {
		return "Thought passes basic logical consistency checks"
	}

	// Provide specific reasons for invalid thoughts
	lower := strings.ToLower(content)
	
	if strings.Contains(lower, "always") && strings.Contains(lower, "never") {
		return "Contains contradictory absolute statements (always/never)"
	}
	
	if strings.Contains(lower, "all") && strings.Contains(lower, "none") {
		return "Contains contradictory universal quantifiers (all/none)"
	}

	return "Potential logical inconsistency detected"
}

// ProofResult represents the result of a logical proof attempt
type ProofResult struct {
	Premises   []string `json:"premises"`
	Conclusion string   `json:"conclusion"`
	IsProvable bool     `json:"is_provable"`
	Steps      []string `json:"steps"`
}

// StatementCheck represents syntax validation results
type StatementCheck struct {
	Statement    string   `json:"statement"`
	IsWellFormed bool     `json:"is_well_formed"`
	Issues       []string `json:"issues,omitempty"`
}
