package server

import (
	"strings"
	"testing"
)

func TestValidateThinkRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *ThinkRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid basic request",
			req: &ThinkRequest{
				Content:    "Test content",
				Mode:       "linear",
				Confidence: 0.8,
			},
			wantErr: false,
		},
		{
			name: "valid request with all fields",
			req: &ThinkRequest{
				Content:    "Test content",
				Mode:       "tree",
				Type:       "analysis",
				BranchID:   "branch-1",
				ParentID:   "parent-1",
				Confidence: 0.9,
				KeyPoints:  []string{"point1", "point2"},
				CrossRefs: []CrossRefInput{
					{ToBranch: "b1", Type: "complementary", Reason: "test", Strength: 0.8},
				},
			},
			wantErr: false,
		},
		{
			name: "empty content",
			req: &ThinkRequest{
				Content: "",
				Mode:    "linear",
			},
			wantErr: true,
			errMsg:  "content cannot be empty",
		},
		{
			name: "content too long",
			req: &ThinkRequest{
				Content: strings.Repeat("a", MaxContentLength+1),
				Mode:    "linear",
			},
			wantErr: true,
			errMsg:  "content exceeds maximum length",
		},
		{
			name: "invalid UTF-8 content",
			req: &ThinkRequest{
				Content: string([]byte{0xff, 0xfe, 0xfd}),
				Mode:    "linear",
			},
			wantErr: true,
			errMsg:  "content must be valid UTF-8",
		},
		{
			name: "invalid mode",
			req: &ThinkRequest{
				Content: "Test",
				Mode:    "invalid_mode",
			},
			wantErr: true,
			errMsg:  "invalid mode",
		},
		{
			name: "confidence out of range low",
			req: &ThinkRequest{
				Content:    "Test",
				Mode:       "linear",
				Confidence: -0.1,
			},
			wantErr: true,
			errMsg:  "confidence must be between 0.0 and 1.0",
		},
		{
			name: "confidence out of range high",
			req: &ThinkRequest{
				Content:    "Test",
				Mode:       "linear",
				Confidence: 1.5,
			},
			wantErr: true,
			errMsg:  "confidence must be between 0.0 and 1.0",
		},
		{
			name: "too many key points",
			req: &ThinkRequest{
				Content:   "Test",
				Mode:      "linear",
				KeyPoints: make([]string, MaxKeyPoints+1),
			},
			wantErr: true,
			errMsg:  "too many key points",
		},
		{
			name: "key point too long",
			req: &ThinkRequest{
				Content:   "Test",
				Mode:      "linear",
				KeyPoints: []string{strings.Repeat("a", MaxKeyPointLength+1)},
			},
			wantErr: true,
			errMsg:  "exceeds max length",
		},
		{
			name: "invalid UTF-8 in key point",
			req: &ThinkRequest{
				Content:   "Test",
				Mode:      "linear",
				KeyPoints: []string{string([]byte{0xff, 0xfe})},
			},
			wantErr: true,
			errMsg:  "must be valid UTF-8",
		},
		{
			name: "too many cross refs",
			req: &ThinkRequest{
				Content:   "Test",
				Mode:      "tree",
				CrossRefs: make([]CrossRefInput, MaxCrossRefs+1),
			},
			wantErr: true,
			errMsg:  "too many cross references",
		},
		{
			name: "cross ref empty to_branch",
			req: &ThinkRequest{
				Content: "Test",
				Mode:    "tree",
				CrossRefs: []CrossRefInput{
					{ToBranch: "", Type: "complementary", Strength: 0.8},
				},
			},
			wantErr: true,
			errMsg:  "to_branch cannot be empty",
		},
		{
			name: "cross ref invalid strength low",
			req: &ThinkRequest{
				Content: "Test",
				Mode:    "tree",
				CrossRefs: []CrossRefInput{
					{ToBranch: "b1", Type: "complementary", Strength: -0.1},
				},
			},
			wantErr: true,
			errMsg:  "strength must be 0.0-1.0",
		},
		{
			name: "cross ref invalid strength high",
			req: &ThinkRequest{
				Content: "Test",
				Mode:    "tree",
				CrossRefs: []CrossRefInput{
					{ToBranch: "b1", Type: "complementary", Strength: 1.5},
				},
			},
			wantErr: true,
			errMsg:  "strength must be 0.0-1.0",
		},
		{
			name: "cross ref invalid type",
			req: &ThinkRequest{
				Content: "Test",
				Mode:    "tree",
				CrossRefs: []CrossRefInput{
					{ToBranch: "b1", Type: "invalid", Strength: 0.8},
				},
			},
			wantErr: true,
			errMsg:  "type invalid",
		},
		{
			name: "type too long",
			req: &ThinkRequest{
				Content: "Test",
				Mode:    "linear",
				Type:    strings.Repeat("a", MaxTypeLength+1),
			},
			wantErr: true,
			errMsg:  "type exceeds maximum length",
		},
		{
			name: "branch ID too long",
			req: &ThinkRequest{
				Content:  "Test",
				Mode:     "tree",
				BranchID: strings.Repeat("a", MaxBranchIDLength+1),
			},
			wantErr: true,
			errMsg:  "branch_id too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateThinkRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateThinkRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateHistoryRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *HistoryRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid empty request",
			req:     &HistoryRequest{},
			wantErr: false,
		},
		{
			name: "valid with mode",
			req: &HistoryRequest{
				Mode: "linear",
			},
			wantErr: false,
		},
		{
			name: "valid with branch ID",
			req: &HistoryRequest{
				BranchID: "branch-123",
			},
			wantErr: false,
		},
		{
			name: "invalid mode",
			req: &HistoryRequest{
				Mode: "invalid",
			},
			wantErr: true,
			errMsg:  "invalid mode",
		},
		{
			name: "branch ID too long",
			req: &HistoryRequest{
				BranchID: strings.Repeat("a", MaxBranchIDLength+1),
			},
			wantErr: true,
			errMsg:  "branch_id too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHistoryRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHistoryRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateFocusBranchRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *FocusBranchRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &FocusBranchRequest{
				BranchID: "branch-123",
			},
			wantErr: false,
		},
		{
			name: "empty branch ID",
			req: &FocusBranchRequest{
				BranchID: "",
			},
			wantErr: true,
			errMsg:  "branch_id is required",
		},
		{
			name: "branch ID too long",
			req: &FocusBranchRequest{
				BranchID: strings.Repeat("a", MaxBranchIDLength+1),
			},
			wantErr: true,
			errMsg:  "branch_id too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFocusBranchRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFocusBranchRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateBranchHistoryRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *BranchHistoryRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &BranchHistoryRequest{
				BranchID: "branch-123",
			},
			wantErr: false,
		},
		{
			name: "empty branch ID",
			req: &BranchHistoryRequest{
				BranchID: "",
			},
			wantErr: true,
			errMsg:  "branch_id is required",
		},
		{
			name: "branch ID too long",
			req: &BranchHistoryRequest{
				BranchID: strings.Repeat("a", MaxBranchIDLength+1),
			},
			wantErr: true,
			errMsg:  "branch_id too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranchHistoryRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBranchHistoryRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateValidateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *ValidateRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &ValidateRequest{
				ThoughtID: "thought-123",
			},
			wantErr: false,
		},
		{
			name: "empty thought ID",
			req: &ValidateRequest{
				ThoughtID: "",
			},
			wantErr: true,
			errMsg:  "thought_id is required",
		},
		{
			name: "thought ID too long",
			req: &ValidateRequest{
				ThoughtID: strings.Repeat("a", MaxBranchIDLength+1),
			},
			wantErr: true,
			errMsg:  "thought_id too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateValidateRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateProveRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *ProveRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &ProveRequest{
				Premises:   []string{"All men are mortal", "Socrates is a man"},
				Conclusion: "Socrates is mortal",
			},
			wantErr: false,
		},
		{
			name: "empty premises",
			req: &ProveRequest{
				Premises:   []string{},
				Conclusion: "Conclusion",
			},
			wantErr: true,
			errMsg:  "at least one premise is required",
		},
		{
			name: "too many premises",
			req: &ProveRequest{
				Premises:   make([]string, MaxPremises+1),
				Conclusion: "Conclusion",
			},
			wantErr: true,
			errMsg:  "too many premises",
		},
		{
			name: "empty premise",
			req: &ProveRequest{
				Premises:   []string{"Valid premise", ""},
				Conclusion: "Conclusion",
			},
			wantErr: true,
			errMsg:  "premise[1] cannot be empty",
		},
		{
			name: "premise too long",
			req: &ProveRequest{
				Premises:   []string{strings.Repeat("a", MaxPremiseLength+1)},
				Conclusion: "Conclusion",
			},
			wantErr: true,
			errMsg:  "premise[0] too long",
		},
		{
			name: "invalid UTF-8 premise",
			req: &ProveRequest{
				Premises:   []string{string([]byte{0xff, 0xfe})},
				Conclusion: "Conclusion",
			},
			wantErr: true,
			errMsg:  "premise[0] not valid UTF-8",
		},
		{
			name: "empty conclusion",
			req: &ProveRequest{
				Premises:   []string{"Premise"},
				Conclusion: "",
			},
			wantErr: true,
			errMsg:  "conclusion is required",
		},
		{
			name: "conclusion too long",
			req: &ProveRequest{
				Premises:   []string{"Premise"},
				Conclusion: strings.Repeat("a", MaxPremiseLength+1),
			},
			wantErr: true,
			errMsg:  "conclusion too long",
		},
		{
			name: "invalid UTF-8 conclusion",
			req: &ProveRequest{
				Premises:   []string{"Premise"},
				Conclusion: string([]byte{0xff, 0xfe}),
			},
			wantErr: true,
			errMsg:  "conclusion not valid UTF-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize empty premises with valid data for the "too many" test
			if len(tt.req.Premises) > MaxPremises {
				for i := range tt.req.Premises {
					tt.req.Premises[i] = "premise"
				}
			}

			err := ValidateProveRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProveRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateCheckSyntaxRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CheckSyntaxRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &CheckSyntaxRequest{
				Statements: []string{"All men are mortal", "Some cats are black"},
			},
			wantErr: false,
		},
		{
			name: "empty statements",
			req: &CheckSyntaxRequest{
				Statements: []string{},
			},
			wantErr: true,
			errMsg:  "at least one statement is required",
		},
		{
			name: "too many statements",
			req: &CheckSyntaxRequest{
				Statements: make([]string, MaxStatements+1),
			},
			wantErr: true,
			errMsg:  "too many statements",
		},
		{
			name: "statement too long",
			req: &CheckSyntaxRequest{
				Statements: []string{strings.Repeat("a", MaxStatementLength+1)},
			},
			wantErr: true,
			errMsg:  "statement[0] too long",
		},
		{
			name: "invalid UTF-8 statement",
			req: &CheckSyntaxRequest{
				Statements: []string{string([]byte{0xff, 0xfe})},
			},
			wantErr: true,
			errMsg:  "statement[0] not valid UTF-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCheckSyntaxRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCheckSyntaxRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateSearchRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *SearchRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &SearchRequest{
				Query: "test query",
				Mode:  "linear",
			},
			wantErr: false,
		},
		{
			name: "valid empty query",
			req: &SearchRequest{
				Query: "",
				Mode:  "",
			},
			wantErr: false,
		},
		{
			name: "query too long",
			req: &SearchRequest{
				Query: strings.Repeat("a", MaxQueryLength+1),
			},
			wantErr: true,
			errMsg:  "query exceeds maximum length",
		},
		{
			name: "invalid UTF-8 query",
			req: &SearchRequest{
				Query: string([]byte{0xff, 0xfe}),
			},
			wantErr: true,
			errMsg:  "query must be valid UTF-8",
		},
		{
			name: "invalid mode",
			req: &SearchRequest{
				Query: "test",
				Mode:  "invalid",
			},
			wantErr: true,
			errMsg:  "invalid mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSearchRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSearchRequest() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errMsg != "" {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "test_field",
		Message: "test message",
	}

	expected := "validation error on field 'test_field': test message"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %v, want %v", err.Error(), expected)
	}
}
