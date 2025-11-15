package handlers

import (
	"strings"
	"testing"
)

func TestValidateFindAnalogyRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *FindAnalogyRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &FindAnalogyRequest{
				SourceDomain:  "biology: immune system",
				TargetProblem: "How to protect computer network?",
				Constraints:   []string{"technical", "security"},
			},
			wantErr: false,
		},
		{
			name: "missing source domain",
			req: &FindAnalogyRequest{
				TargetProblem: "How to protect network?",
			},
			wantErr: true,
			errMsg:  "source_domain is required",
		},
		{
			name: "missing target problem",
			req: &FindAnalogyRequest{
				SourceDomain: "biology: immune system",
			},
			wantErr: true,
			errMsg:  "target_problem is required",
		},
		{
			name: "too many constraints",
			req: &FindAnalogyRequest{
				SourceDomain:  "biology",
				TargetProblem: "problem",
				Constraints:   make([]string, 51), // MaxConstraints is 50
			},
			wantErr: true,
			errMsg:  "too many constraints",
		},
		{
			name: "source domain too long",
			req: &FindAnalogyRequest{
				SourceDomain:  string(make([]byte, MaxContentLength+1)),
				TargetProblem: "problem",
			},
			wantErr: true,
			errMsg:  "source_domain exceeds max length",
		},
		{
			name: "target problem too long",
			req: &FindAnalogyRequest{
				SourceDomain:  "biology",
				TargetProblem: string(make([]byte, MaxContentLength+1)),
			},
			wantErr: true,
			errMsg:  "target_problem exceeds max length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFindAnalogyRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFindAnalogyRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateApplyAnalogyRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *ApplyAnalogyRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &ApplyAnalogyRequest{
				AnalogyID:     "analogy_123",
				TargetContext: "New security scenario",
			},
			wantErr: false,
		},
		{
			name: "missing analogy ID",
			req: &ApplyAnalogyRequest{
				TargetContext: "scenario",
			},
			wantErr: true,
			errMsg:  "analogy_id is required",
		},
		{
			name: "missing target context",
			req: &ApplyAnalogyRequest{
				AnalogyID: "analogy_123",
			},
			wantErr: true,
			errMsg:  "target_context is required",
		},
		{
			name: "analogy ID too long",
			req: &ApplyAnalogyRequest{
				AnalogyID:     string(make([]byte, MaxAnalogyIDLength+1)),
				TargetContext: "scenario",
			},
			wantErr: true,
			errMsg:  "analogy_id too long",
		},
		{
			name: "target context too long",
			req: &ApplyAnalogyRequest{
				AnalogyID:     "analogy_123",
				TargetContext: string(make([]byte, MaxContextLength+1)),
			},
			wantErr: true,
			errMsg:  "target_context exceeds max length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateApplyAnalogyRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateApplyAnalogyRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateDecomposeArgumentRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *DecomposeArgumentRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &DecomposeArgumentRequest{
				Argument: "We should adopt policy X because studies show it reduces Y by 30%",
			},
			wantErr: false,
		},
		{
			name: "empty argument",
			req: &DecomposeArgumentRequest{
				Argument: "",
			},
			wantErr: true,
			errMsg:  "argument is required",
		},
		{
			name: "argument too long",
			req: &DecomposeArgumentRequest{
				Argument: string(make([]byte, MaxContentLength+1)),
			},
			wantErr: true,
			errMsg:  "argument exceeds max length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDecomposeArgumentRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDecomposeArgumentRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateGenerateCounterArgumentsRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *GenerateCounterArgumentsRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &GenerateCounterArgumentsRequest{
				ArgumentID: "arg_123",
			},
			wantErr: false,
		},
		{
			name: "missing argument ID",
			req: &GenerateCounterArgumentsRequest{
				ArgumentID: "",
			},
			wantErr: true,
			errMsg:  "argument_id is required",
		},
		{
			name: "argument ID too long",
			req: &GenerateCounterArgumentsRequest{
				ArgumentID: string(make([]byte, MaxArgumentIDLength+1)),
			},
			wantErr: true,
			errMsg:  "argument_id too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGenerateCounterArgumentsRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGenerateCounterArgumentsRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateDetectFallaciesRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *DetectFallaciesRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &DetectFallaciesRequest{
				Content:       "Everyone says X is true, so it must be true",
				CheckFormal:   true,
				CheckInformal: true,
			},
			wantErr: false,
		},
		{
			name: "missing content",
			req: &DetectFallaciesRequest{
				CheckFormal:   true,
				CheckInformal: true,
			},
			wantErr: true,
			errMsg:  "content is required",
		},
		{
			name: "content too long",
			req: &DetectFallaciesRequest{
				Content:       string(make([]byte, MaxContentLength+1)),
				CheckFormal:   true,
				CheckInformal: true,
			},
			wantErr: true,
			errMsg:  "content exceeds max length",
		},
		{
			name: "valid with only formal check",
			req: &DetectFallaciesRequest{
				Content:     "Some logical argument",
				CheckFormal: true,
			},
			wantErr: false,
		},
		{
			name: "valid with only informal check",
			req: &DetectFallaciesRequest{
				Content:       "Some argument",
				CheckInformal: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDetectFallaciesRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDetectFallaciesRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateProcessEvidencePipelineRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *ProcessEvidencePipelineRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &ProcessEvidencePipelineRequest{
				Content:       "Study shows X increases Y by 30%",
				Source:        "Journal of Science 2024",
				ClaimID:       "claim_123",
				SupportsClaim: true,
			},
			wantErr: false,
		},
		{
			name: "valid request without claim ID",
			req: &ProcessEvidencePipelineRequest{
				Content:       "Study shows X increases Y by 30%",
				Source:        "Journal of Science 2024",
				SupportsClaim: true,
			},
			wantErr: false,
		},
		{
			name: "missing content",
			req: &ProcessEvidencePipelineRequest{
				Source:        "Journal",
				SupportsClaim: true,
			},
			wantErr: true,
			errMsg:  "content is required",
		},
		{
			name: "missing source",
			req: &ProcessEvidencePipelineRequest{
				Content:       "Study shows",
				SupportsClaim: true,
			},
			wantErr: true,
			errMsg:  "source is required",
		},
		{
			name: "content too long",
			req: &ProcessEvidencePipelineRequest{
				Content:       string(make([]byte, MaxContentLength+1)),
				Source:        "Journal",
				SupportsClaim: true,
			},
			wantErr: true,
			errMsg:  "content exceeds max length",
		},
		{
			name: "source too long",
			req: &ProcessEvidencePipelineRequest{
				Content:       "Study shows",
				Source:        string(make([]byte, MaxSourceLength+1)),
				SupportsClaim: true,
			},
			wantErr: true,
			errMsg:  "source exceeds max length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProcessEvidencePipelineRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProcessEvidencePipelineRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateAnalyzeTemporalCausalEffectsRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *AnalyzeTemporalCausalEffectsRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &AnalyzeTemporalCausalEffectsRequest{
				GraphID:          "graph_123",
				VariableID:       "marketing_spend",
				InterventionType: "increase",
			},
			wantErr: false,
		},
		{
			name: "missing graph ID",
			req: &AnalyzeTemporalCausalEffectsRequest{
				VariableID:       "marketing_spend",
				InterventionType: "increase",
			},
			wantErr: true,
			errMsg:  "graph_id is required",
		},
		{
			name: "missing variable ID",
			req: &AnalyzeTemporalCausalEffectsRequest{
				GraphID:          "graph_123",
				InterventionType: "increase",
			},
			wantErr: true,
			errMsg:  "variable_id is required",
		},
		{
			name: "missing intervention type",
			req: &AnalyzeTemporalCausalEffectsRequest{
				GraphID:    "graph_123",
				VariableID: "marketing_spend",
			},
			wantErr: true,
			errMsg:  "intervention_type is required",
		},
		{
			name: "invalid intervention type",
			req: &AnalyzeTemporalCausalEffectsRequest{
				GraphID:          "graph_123",
				VariableID:       "marketing_spend",
				InterventionType: "invalid_type",
			},
			wantErr: true,
			errMsg:  "invalid intervention_type",
		},
		{
			name: "graph ID too long",
			req: &AnalyzeTemporalCausalEffectsRequest{
				GraphID:          string(make([]byte, MaxGraphIDLength+1)),
				VariableID:       "marketing_spend",
				InterventionType: "increase",
			},
			wantErr: true,
			errMsg:  "graph_id too long",
		},
		{
			name: "variable ID too long",
			req: &AnalyzeTemporalCausalEffectsRequest{
				GraphID:          "graph_123",
				VariableID:       string(make([]byte, MaxVariableIDLength+1)),
				InterventionType: "increase",
			},
			wantErr: true,
			errMsg:  "variable_id too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAnalyzeTemporalCausalEffectsRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAnalyzeTemporalCausalEffectsRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateAnalyzeDecisionTimingRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *AnalyzeDecisionTimingRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &AnalyzeDecisionTimingRequest{
				Situation:     "When to launch product?",
				CausalGraphID: "graph_123",
			},
			wantErr: false,
		},
		{
			name: "valid request without causal graph ID",
			req: &AnalyzeDecisionTimingRequest{
				Situation: "When to launch product?",
			},
			wantErr: false,
		},
		{
			name: "missing situation",
			req: &AnalyzeDecisionTimingRequest{
				CausalGraphID: "graph_123",
			},
			wantErr: true,
			errMsg:  "situation is required",
		},
		{
			name: "situation too long",
			req: &AnalyzeDecisionTimingRequest{
				Situation:     string(make([]byte, MaxContentLength+1)),
				CausalGraphID: "graph_123",
			},
			wantErr: true,
			errMsg:  "situation exceeds max length",
		},
		{
			name: "causal graph ID too long",
			req: &AnalyzeDecisionTimingRequest{
				Situation:     "When to launch?",
				CausalGraphID: string(make([]byte, MaxGraphIDLength+1)),
			},
			wantErr: true,
			errMsg:  "causal_graph_id too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAnalyzeDecisionTimingRequest(tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAnalyzeDecisionTimingRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() != tt.errMsg && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Error message = %v, should contain %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestToJSONContent(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantLen int
	}{
		{
			name:    "simple struct",
			data:    map[string]string{"key": "value"},
			wantLen: 1,
		},
		{
			name:    "complex struct",
			data:    map[string]interface{}{"nested": map[string]int{"count": 5}},
			wantLen: 1,
		},
		{
			name:    "nil data",
			data:    nil,
			wantLen: 1,
		},
		{
			name:    "array",
			data:    []string{"item1", "item2"},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toJSONContent(tt.data)
			if len(result) != tt.wantLen {
				t.Errorf("toJSONContent() length = %v, want %v", len(result), tt.wantLen)
			}

			// Check that result contains TextContent
			if len(result) > 0 {
				if result[0] == nil {
					t.Error("toJSONContent() returned nil content")
				}
			}
		})
	}
}
