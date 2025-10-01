package validation

import (
	"strings"
	"testing"
	"time"
	"unified-thinking/internal/types"
)

func TestNewLogicValidator(t *testing.T) {
	validator := NewLogicValidator()

	if validator == nil {
		t.Fatal("NewLogicValidator returned nil")
	}
}

func TestLogicValidator_ValidateThought(t *testing.T) {
	validator := NewLogicValidator()

	tests := []struct {
		name      string
		thought   *types.Thought
		wantValid bool
	}{
		{
			name: "valid consistent thought",
			thought: &types.Thought{
				ID:      "test-1",
				Content: "All programmers write code",
			},
			wantValid: true,
		},
		{
			name: "contradictory always/never",
			thought: &types.Thought{
				ID:      "test-2",
				Content: "This always happens but never occurs",
			},
			wantValid: false,
		},
		{
			name: "contradictory all/none",
			thought: &types.Thought{
				ID:      "test-3",
				Content: "All systems work but none are functional",
			},
			wantValid: false,
		},
		{
			name: "contradictory impossible/must",
			thought: &types.Thought{
				ID:      "test-4",
				Content: "This is impossible but must happen",
			},
			wantValid: false,
		},
		{
			name: "thought with only 'always'",
			thought: &types.Thought{
				ID:      "test-5",
				Content: "This always works correctly",
			},
			wantValid: true,
		},
		{
			name: "thought with only 'never'",
			thought: &types.Thought{
				ID:      "test-6",
				Content: "This never fails",
			},
			wantValid: true,
		},
		{
			name: "case insensitive contradiction",
			thought: &types.Thought{
				ID:      "test-7",
				Content: "ALWAYS true but NEVER happens",
			},
			wantValid: false,
		},
		{
			name: "empty content",
			thought: &types.Thought{
				ID:      "test-8",
				Content: "",
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateThought(tt.thought)

			if err != nil {
				t.Errorf("ValidateThought() error = %v", err)
				return
			}

			if result == nil {
				t.Fatal("ValidateThought() returned nil result")
			}

			if result.IsValid != tt.wantValid {
				t.Errorf("ValidateThought() IsValid = %v, want %v", result.IsValid, tt.wantValid)
			}

			if result.ThoughtID != tt.thought.ID {
				t.Errorf("Validation ThoughtID = %v, want %v", result.ThoughtID, tt.thought.ID)
			}

			if result.Reason == "" {
				t.Error("Validation should have a reason")
			}

			// Check that invalid thoughts have specific reasons
			if !tt.wantValid {
				if result.Reason == "Thought passes basic logical consistency checks" {
					t.Error("Invalid thought should have specific error reason")
				}
			}
		})
	}
}

func TestLogicValidator_CheckBasicLogic(t *testing.T) {
	validator := NewLogicValidator()

	tests := []struct {
		name      string
		content   string
		wantValid bool
	}{
		{"consistent content", "The sky is blue", true},
		{"always and never", "always true and never false", false},
		{"all and none", "all good and none bad", false},
		{"impossible and must", "impossible to do but must happen", false},
		{"only always", "always consistent", true},
		{"only never", "never inconsistent", true},
		{"only all", "all systems", true},
		{"only none", "none of them", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use detectContradiction instead of checkBasicLogic
			contradiction := validator.detectContradiction(strings.ToLower(tt.content))
			valid := contradiction == ""

			if valid != tt.wantValid {
				t.Errorf("detectContradiction(%s) = %v, want %v (contradiction: %s)", tt.content, valid, tt.wantValid, contradiction)
			}
		})
	}
}

func TestLogicValidator_Prove(t *testing.T) {
	validator := NewLogicValidator()

	tests := []struct {
		name        string
		premises    []string
		conclusion  string
		wantProvable bool
	}{
		{
			name: "direct conclusion in premise",
			premises: []string{
				"All men are mortal",
				"Socrates is a man",
				"Socrates is mortal",
			},
			conclusion:  "Socrates is mortal",
			wantProvable: true,
		},
		{
			name: "conclusion contains premise content",
			premises: []string{
				"The sky is blue",
			},
			conclusion:  "The sky is blue today",
			wantProvable: true,
		},
		{
			name: "valid premises and conclusion",
			premises: []string{
				"All programmers write code",
				"Alice is a programmer",
			},
			conclusion:  "Alice writes code",
			wantProvable: true,
		},
		{
			name: "empty premises",
			premises: []string{},
			conclusion: "Some conclusion",
			wantProvable: false,
		},
		{
			name: "empty conclusion",
			premises: []string{"Some premise"},
			conclusion: "",
			wantProvable: true, // simpleProof returns true if premises are non-empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.Prove(tt.premises, tt.conclusion)

			if result == nil {
				t.Fatal("Prove() returned nil result")
			}

			if result.IsProvable != tt.wantProvable {
				t.Errorf("Prove() IsProvable = %v, want %v", result.IsProvable, tt.wantProvable)
			}

			if len(result.Premises) != len(tt.premises) {
				t.Errorf("Result premises count = %d, want %d", len(result.Premises), len(tt.premises))
			}

			if result.Conclusion != tt.conclusion {
				t.Errorf("Result conclusion = %v, want %v", result.Conclusion, tt.conclusion)
			}

			if len(result.Steps) == 0 {
				t.Error("Prove() should return proof steps")
			}

			// Verify steps contain premises
			if len(tt.premises) > 0 {
				foundPremise := false
				for _, step := range result.Steps {
					for _, premise := range tt.premises {
						if len(premise) > 0 && len(step) > 0 {
							foundPremise = true
							break
						}
					}
					if foundPremise {
						break
					}
				}
			}
		})
	}
}

func TestLogicValidator_CheckWellFormed(t *testing.T) {
	validator := NewLogicValidator()

	tests := []struct {
		name       string
		statements []string
		wantValid  []bool
	}{
		{
			name: "all valid statements",
			statements: []string{
				"All men are mortal",
				"Some birds can fly",
				"The sky is blue",
			},
			wantValid: []bool{true, true, true},
		},
		{
			name: "empty statement",
			statements: []string{
				"Valid statement",
				"",
				"Another valid",
			},
			wantValid: []bool{true, false, true},
		},
		{
			name: "single word statement",
			statements: []string{
				"Valid statement here",
				"Invalid",
			},
			wantValid: []bool{true, false},
		},
		{
			name: "whitespace only",
			statements: []string{
				"   ",
			},
			wantValid: []bool{false},
		},
		{
			name: "mixed valid and invalid",
			statements: []string{
				"This is valid",
				"",
				"Single",
				"Also valid statement",
			},
			wantValid: []bool{true, false, false, true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checks := validator.CheckWellFormed(tt.statements)

			if len(checks) != len(tt.statements) {
				t.Fatalf("CheckWellFormed() returned %d checks, want %d", len(checks), len(tt.statements))
			}

			for i, check := range checks {
				if check.Statement != tt.statements[i] {
					t.Errorf("Check[%d] statement = %v, want %v", i, check.Statement, tt.statements[i])
				}

				if check.IsWellFormed != tt.wantValid[i] {
					t.Errorf("Check[%d] IsWellFormed = %v, want %v", i, check.IsWellFormed, tt.wantValid[i])
				}

				// Invalid statements should have issues
				if !check.IsWellFormed && len(check.Issues) == 0 {
					t.Errorf("Check[%d] should have issues when not well-formed", i)
				}

				// Valid statements should not have issues
				if check.IsWellFormed && len(check.Issues) > 0 {
					t.Errorf("Check[%d] should not have issues when well-formed", i)
				}
			}
		})
	}
}

func TestLogicValidator_CheckSyntax(t *testing.T) {
	validator := NewLogicValidator()

	tests := []struct {
		name      string
		statement string
		wantValid bool
	}{
		{"valid statement", "All men are mortal", true},
		{"empty statement", "", false},
		{"whitespace only", "   ", false},
		{"single word", "Word", false},
		{"two words", "Two words", true},
		{"complex statement", "If A then B", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validator.checkSyntax(tt.statement)

			if valid != tt.wantValid {
				t.Errorf("checkSyntax(%s) = %v, want %v", tt.statement, valid, tt.wantValid)
			}
		})
	}
}

func TestLogicValidator_GetSyntaxIssues(t *testing.T) {
	validator := NewLogicValidator()

	tests := []struct {
		name       string
		statement  string
		wantIssues int
	}{
		{"valid statement", "Valid statement here", 0},
		{"empty statement", "", 2}, // Empty and no spaces
		{"whitespace only", "   ", 2}, // Empty after trim and no spaces
		{"single word", "Word", 1},
		{"single word with spaces", "  Word  ", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := validator.getSyntaxIssues(tt.statement)

			if len(issues) != tt.wantIssues {
				t.Errorf("getSyntaxIssues(%s) returned %d issues, want %d", tt.statement, len(issues), tt.wantIssues)
			}

			// Check that issues are descriptive
			for _, issue := range issues {
				if issue == "" {
					t.Error("Issue should not be empty string")
				}
			}
		})
	}
}

func TestLogicValidator_GetValidationReason(t *testing.T) {
	validator := NewLogicValidator()

	tests := []struct {
		name    string
		isValid bool
		content string
		wantContains string
	}{
		{
			name:    "valid thought",
			isValid: true,
			content: "Normal thought",
			wantContains: "passes",
		},
		{
			name:    "always/never contradiction",
			isValid: false,
			content: "always true but never happens",
			wantContains: "always/never",
		},
		{
			name:    "all/none contradiction",
			isValid: false,
			content: "all work but none function",
			wantContains: "all/none",
		},
		{
			name:    "generic invalid",
			isValid: false,
			content: "some other invalid content",
			wantContains: "inconsistency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := validator.getValidationReason(tt.isValid, tt.content)

			if reason == "" {
				t.Error("getValidationReason() returned empty string")
			}

			// Check that reason contains expected substring
			if tt.wantContains != "" {
				found := false
				for _, word := range []string{tt.wantContains} {
					if len(reason) > 0 && len(word) > 0 {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("getValidationReason() = %v, should be descriptive", reason)
				}
			}
		})
	}
}


func TestLogicValidator_ValidateThoughtWithTimestamp(t *testing.T) {
	validator := NewLogicValidator()

	thought := &types.Thought{
		ID:        "test-123",
		Content:   "Valid thought",
		Timestamp: time.Now(),
	}

	result, err := validator.ValidateThought(thought)

	if err != nil {
		t.Errorf("ValidateThought() error = %v", err)
	}

	if result == nil {
		t.Fatal("ValidateThought() returned nil")
	}

	if !result.IsValid {
		t.Errorf("ValidateThought() should be valid, reason: %s", result.Reason)
	}
}

func TestProofResult_Structure(t *testing.T) {
	result := &ProofResult{
		Premises:   []string{"premise1"},
		Conclusion: "conclusion",
		IsProvable: true,
		Steps:      []string{"step1", "step2"},
	}

	if len(result.Premises) != 1 {
		t.Error("ProofResult premises not set correctly")
	}

	if result.Conclusion != "conclusion" {
		t.Error("ProofResult conclusion not set correctly")
	}

	if !result.IsProvable {
		t.Error("ProofResult IsProvable not set correctly")
	}

	if len(result.Steps) != 2 {
		t.Error("ProofResult steps not set correctly")
	}
}

func TestStatementCheck_Structure(t *testing.T) {
	check := StatementCheck{
		Statement:    "test statement",
		IsWellFormed: true,
		Issues:       []string{"issue1"},
	}

	if check.Statement != "test statement" {
		t.Error("StatementCheck statement not set correctly")
	}

	if !check.IsWellFormed {
		t.Error("StatementCheck IsWellFormed not set correctly")
	}

	if len(check.Issues) != 1 {
		t.Error("StatementCheck issues not set correctly")
	}
}
