package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/analysis"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

func TestNewDecisionHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	decisionMaker := reasoning.NewDecisionMaker()
	problemDecomposer := reasoning.NewProblemDecomposer()
	sensitivityAnalyzer := analysis.NewSensitivityAnalyzer()

	handler := NewDecisionHandler(store, decisionMaker, problemDecomposer, sensitivityAnalyzer)

	if handler == nil {
		t.Fatal("NewDecisionHandler returned nil")
	}
	if handler.storage == nil {
		t.Error("storage not initialized")
	}
	if handler.decisionMaker == nil {
		t.Error("decisionMaker not initialized")
	}
	if handler.problemDecomposer == nil {
		t.Error("problemDecomposer not initialized")
	}
	if handler.sensitivityAnalyzer == nil {
		t.Error("sensitivityAnalyzer not initialized")
	}
	if handler.metadataGen == nil {
		t.Error("metadataGen not initialized")
	}
}

func TestDecisionHandler_HandleMakeDecision(t *testing.T) {
	store := storage.NewMemoryStorage()
	decisionMaker := reasoning.NewDecisionMaker()
	problemDecomposer := reasoning.NewProblemDecomposer()
	sensitivityAnalyzer := analysis.NewSensitivityAnalyzer()
	handler := NewDecisionHandler(store, decisionMaker, problemDecomposer, sensitivityAnalyzer)

	tests := []struct {
		name    string
		input   MakeDecisionRequest
		wantErr bool
	}{
		{
			name: "valid decision request",
			input: MakeDecisionRequest{
				Question: "Which database should we use?",
				Options: []*types.DecisionOption{
					{
						ID:          "pg",
						Name:        "PostgreSQL",
						Description: "Relational database",
						Scores:      map[string]float64{"cost": 0.8, "performance": 0.9},
						Pros:        []string{"ACID compliant"},
						Cons:        []string{"Complex setup"},
					},
					{
						ID:          "mongo",
						Name:        "MongoDB",
						Description: "Document database",
						Scores:      map[string]float64{"cost": 0.7, "performance": 0.8},
						Pros:        []string{"Flexible schema"},
						Cons:        []string{"No ACID"},
					},
				},
				Criteria: []*types.DecisionCriterion{
					{ID: "cost", Name: "Cost", Description: "Total cost", Weight: 0.5, Maximize: false},
					{ID: "performance", Name: "Performance", Description: "Speed", Weight: 0.5, Maximize: true},
				},
			},
			wantErr: false,
		},
		{
			name: "missing question",
			input: MakeDecisionRequest{
				Question: "",
				Options: []*types.DecisionOption{
					{ID: "opt1", Name: "Option 1", Scores: map[string]float64{"cost": 0.8}},
				},
				Criteria: []*types.DecisionCriterion{
					{ID: "cost", Name: "Cost", Weight: 1.0},
				},
			},
			wantErr: true,
		},
		{
			name: "missing options",
			input: MakeDecisionRequest{
				Question: "Test question?",
				Options:  []*types.DecisionOption{},
				Criteria: []*types.DecisionCriterion{
					{ID: "cost", Name: "Cost", Weight: 1.0},
				},
			},
			wantErr: true,
		},
		{
			name: "missing criteria",
			input: MakeDecisionRequest{
				Question: "Test question?",
				Options: []*types.DecisionOption{
					{ID: "opt1", Name: "Option 1", Scores: map[string]float64{"cost": 0.8}},
				},
				Criteria: []*types.DecisionCriterion{},
			},
			wantErr: true,
		},
		{
			name: "invalid option score",
			input: MakeDecisionRequest{
				Question: "Test question?",
				Options: []*types.DecisionOption{
					{ID: "opt1", Name: "Option 1", Scores: map[string]float64{"cost": 1.5}},
				},
				Criteria: []*types.DecisionCriterion{
					{ID: "cost", Name: "Cost", Weight: 1.0},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid criterion weight",
			input: MakeDecisionRequest{
				Question: "Test question?",
				Options: []*types.DecisionOption{
					{ID: "opt1", Name: "Option 1", Scores: map[string]float64{"cost": 0.8}},
				},
				Criteria: []*types.DecisionCriterion{
					{ID: "cost", Name: "Cost", Weight: 1.5},
				},
			},
			wantErr: true,
		},
		{
			name: "weights do not sum to 1.0",
			input: MakeDecisionRequest{
				Question: "Test question?",
				Options: []*types.DecisionOption{
					{ID: "opt1", Name: "Option 1", Scores: map[string]float64{"cost": 0.8, "perf": 0.9}},
				},
				Criteria: []*types.DecisionCriterion{
					{ID: "cost", Name: "Cost", Weight: 0.3},
					{ID: "perf", Name: "Performance", Weight: 0.3},
				},
			},
			wantErr: true,
		},
		{
			name: "option missing ID",
			input: MakeDecisionRequest{
				Question: "Test question?",
				Options: []*types.DecisionOption{
					{ID: "", Name: "Option 1", Scores: map[string]float64{"cost": 0.8}},
				},
				Criteria: []*types.DecisionCriterion{
					{ID: "cost", Name: "Cost", Weight: 1.0},
				},
			},
			wantErr: true,
		},
		{
			name: "option missing name",
			input: MakeDecisionRequest{
				Question: "Test question?",
				Options: []*types.DecisionOption{
					{ID: "opt1", Name: "", Scores: map[string]float64{"cost": 0.8}},
				},
				Criteria: []*types.DecisionCriterion{
					{ID: "cost", Name: "Cost", Weight: 1.0},
				},
			},
			wantErr: true,
		},
		{
			name: "option missing scores",
			input: MakeDecisionRequest{
				Question: "Test question?",
				Options: []*types.DecisionOption{
					{ID: "opt1", Name: "Option 1", Scores: map[string]float64{}},
				},
				Criteria: []*types.DecisionCriterion{
					{ID: "cost", Name: "Cost", Weight: 1.0},
				},
			},
			wantErr: true,
		},
		{
			name: "criterion missing ID",
			input: MakeDecisionRequest{
				Question: "Test question?",
				Options: []*types.DecisionOption{
					{ID: "opt1", Name: "Option 1", Scores: map[string]float64{"cost": 0.8}},
				},
				Criteria: []*types.DecisionCriterion{
					{ID: "", Name: "Cost", Weight: 1.0},
				},
			},
			wantErr: true,
		},
		{
			name: "criterion missing name",
			input: MakeDecisionRequest{
				Question: "Test question?",
				Options: []*types.DecisionOption{
					{ID: "opt1", Name: "Option 1", Scores: map[string]float64{"cost": 0.8}},
				},
				Criteria: []*types.DecisionCriterion{
					{ID: "cost", Name: "", Weight: 1.0},
				},
			},
			wantErr: true,
		},
		{
			name: "too many options",
			input: MakeDecisionRequest{
				Question: "Test question?",
				Options:  createManyOptions(25),
				Criteria: []*types.DecisionCriterion{
					{ID: "cost", Name: "Cost", Weight: 1.0},
				},
			},
			wantErr: true,
		},
		{
			name: "too many criteria",
			input: MakeDecisionRequest{
				Question: "Test question?",
				Options: []*types.DecisionOption{
					{ID: "opt1", Name: "Option 1", Scores: createManyScores(15)},
				},
				Criteria: createManyCriteria(15),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleMakeDecision(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleMakeDecision() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("MakeDecisionResponse should not be nil")
				}
				if response.Decision == nil {
					t.Error("Decision should not be nil")
				}
				if response.Status != "success" {
					t.Errorf("Status = %v, want success", response.Status)
				}
				if response.Metadata == nil {
					t.Error("Metadata should not be nil")
				}
			}
		})
	}
}

func TestDecisionHandler_HandleDecomposeProblem(t *testing.T) {
	store := storage.NewMemoryStorage()
	decisionMaker := reasoning.NewDecisionMaker()
	problemDecomposer := reasoning.NewProblemDecomposer()
	sensitivityAnalyzer := analysis.NewSensitivityAnalyzer()
	handler := NewDecisionHandler(store, decisionMaker, problemDecomposer, sensitivityAnalyzer)

	tests := []struct {
		name    string
		input   DecomposeProblemRequest
		wantErr bool
	}{
		{
			name: "decomposable problem",
			input: DecomposeProblemRequest{
				Problem: "How to improve system performance by optimizing database queries and reducing memory usage?",
			},
			wantErr: false,
		},
		{
			name: "simple problem",
			input: DecomposeProblemRequest{
				Problem: "What is the meaning of life?",
			},
			wantErr: false,
		},
		{
			name: "empty problem",
			input: DecomposeProblemRequest{
				Problem: "",
			},
			wantErr: true,
		},
		{
			name: "problem too long",
			input: DecomposeProblemRequest{
				Problem: strings.Repeat("a", MaxContentLength+1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleDecomposeProblem(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleDecomposeProblem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("DecomposeProblemResponse should not be nil")
				}
				if response.Status != "success" {
					t.Errorf("Status = %v, want success", response.Status)
				}
				// Either canDecompose is true with decomposition, or false with reason
				if response.CanDecompose {
					if response.Decomposition == nil {
						t.Error("Decomposition should not be nil when CanDecompose is true")
					}
					if response.Metadata == nil {
						t.Error("Metadata should not be nil when problem is decomposable")
					}
				} else {
					if response.ClassificationReason == "" {
						t.Error("ClassificationReason should be provided when problem is not decomposable")
					}
				}
			}
		})
	}
}

func TestDecisionHandler_HandleSensitivityAnalysis(t *testing.T) {
	store := storage.NewMemoryStorage()
	decisionMaker := reasoning.NewDecisionMaker()
	problemDecomposer := reasoning.NewProblemDecomposer()
	sensitivityAnalyzer := analysis.NewSensitivityAnalyzer()
	handler := NewDecisionHandler(store, decisionMaker, problemDecomposer, sensitivityAnalyzer)

	tests := []struct {
		name    string
		input   SensitivityAnalysisRequest
		wantErr bool
	}{
		{
			name: "valid sensitivity analysis",
			input: SensitivityAnalysisRequest{
				TargetClaim:    "Project X will succeed",
				Assumptions:    []string{"Market remains stable", "No major competitors enter", "Team stays intact"},
				BaseConfidence: 0.8,
			},
			wantErr: false,
		},
		{
			name: "minimal assumptions",
			input: SensitivityAnalysisRequest{
				TargetClaim:    "Product will sell",
				Assumptions:    []string{"Users want this"},
				BaseConfidence: 0.5,
			},
			wantErr: false,
		},
		{
			name: "missing target claim",
			input: SensitivityAnalysisRequest{
				TargetClaim:    "",
				Assumptions:    []string{"Assumption 1"},
				BaseConfidence: 0.8,
			},
			wantErr: true,
		},
		{
			name: "missing assumptions",
			input: SensitivityAnalysisRequest{
				TargetClaim:    "Claim",
				Assumptions:    []string{},
				BaseConfidence: 0.8,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence too high",
			input: SensitivityAnalysisRequest{
				TargetClaim:    "Claim",
				Assumptions:    []string{"Assumption"},
				BaseConfidence: 1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid confidence negative",
			input: SensitivityAnalysisRequest{
				TargetClaim:    "Claim",
				Assumptions:    []string{"Assumption"},
				BaseConfidence: -0.1,
			},
			wantErr: true,
		},
		{
			name: "empty assumption in list",
			input: SensitivityAnalysisRequest{
				TargetClaim:    "Claim",
				Assumptions:    []string{"Valid", ""},
				BaseConfidence: 0.8,
			},
			wantErr: true,
		},
		{
			name: "too many assumptions",
			input: SensitivityAnalysisRequest{
				TargetClaim:    "Claim",
				Assumptions:    createManyAssumptions(25),
				BaseConfidence: 0.8,
			},
			wantErr: true,
		},
		{
			name: "target claim too long",
			input: SensitivityAnalysisRequest{
				TargetClaim:    strings.Repeat("a", MaxContentLength+1),
				Assumptions:    []string{"Assumption"},
				BaseConfidence: 0.8,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleSensitivityAnalysis(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleSensitivityAnalysis() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("SensitivityAnalysisResponse should not be nil")
				}
				if response.Analysis == nil {
					t.Error("Analysis should not be nil")
				}
				if response.Status != "success" {
					t.Errorf("Status = %v, want success", response.Status)
				}
			}
		})
	}
}

func TestExtractToolSuggestions(t *testing.T) {
	tests := []struct {
		name     string
		approach string
		wantLen  int
	}{
		{
			name:     "approach with tool mentions",
			approach: "Use unified-thinking:think and unified-thinking:make-decision",
			wantLen:  2,
		},
		{
			name:     "approach with short tool names",
			approach: "Use think and make-decision tools",
			wantLen:  2,
		},
		{
			name:     "philosophical approach",
			approach: "This requires philosophical and reflective analysis",
			wantLen:  2,
		},
		{
			name:     "creative approach",
			approach: "Take a creative and divergent approach",
			wantLen:  2,
		},
		{
			name:     "no matching tools",
			approach: "Generic approach without specific tool mentions",
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := extractToolSuggestions(tt.approach)
			if len(tools) != tt.wantLen {
				t.Errorf("extractToolSuggestions() returned %d tools, want %d", len(tools), tt.wantLen)
			}
		})
	}
}

func TestValidateMakeDecisionRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *MakeDecisionRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &MakeDecisionRequest{
				Question: "Test?",
				Options: []*types.DecisionOption{
					{ID: "a", Name: "A", Scores: map[string]float64{"x": 0.5}},
				},
				Criteria: []*types.DecisionCriterion{
					{ID: "x", Name: "X", Weight: 1.0},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid UTF-8 in question",
			req: &MakeDecisionRequest{
				Question: string([]byte{0xff, 0xfe}),
				Options: []*types.DecisionOption{
					{ID: "a", Name: "A", Scores: map[string]float64{"x": 0.5}},
				},
				Criteria: []*types.DecisionCriterion{
					{ID: "x", Name: "X", Weight: 1.0},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMakeDecisionRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMakeDecisionRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDecomposeProblemRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *DecomposeProblemRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     &DecomposeProblemRequest{Problem: "Test problem"},
			wantErr: false,
		},
		{
			name:    "empty problem",
			req:     &DecomposeProblemRequest{Problem: ""},
			wantErr: true,
		},
		{
			name:    "invalid UTF-8",
			req:     &DecomposeProblemRequest{Problem: string([]byte{0xff, 0xfe})},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDecomposeProblemRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDecomposeProblemRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSensitivityAnalysisRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *SensitivityAnalysisRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &SensitivityAnalysisRequest{
				TargetClaim:    "Claim",
				Assumptions:    []string{"A"},
				BaseConfidence: 0.8,
			},
			wantErr: false,
		},
		{
			name: "invalid UTF-8 in target claim",
			req: &SensitivityAnalysisRequest{
				TargetClaim:    string([]byte{0xff, 0xfe}),
				Assumptions:    []string{"A"},
				BaseConfidence: 0.8,
			},
			wantErr: true,
		},
		{
			name: "invalid UTF-8 in assumption",
			req: &SensitivityAnalysisRequest{
				TargetClaim:    "Claim",
				Assumptions:    []string{string([]byte{0xff, 0xfe})},
				BaseConfidence: 0.8,
			},
			wantErr: true,
		},
		{
			name: "assumption too long",
			req: &SensitivityAnalysisRequest{
				TargetClaim:    "Claim",
				Assumptions:    []string{strings.Repeat("a", MaxQueryLength+1)},
				BaseConfidence: 0.8,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSensitivityAnalysisRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSensitivityAnalysisRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper functions for creating test data
func createManyOptions(n int) []*types.DecisionOption {
	options := make([]*types.DecisionOption, n)
	for i := 0; i < n; i++ {
		options[i] = &types.DecisionOption{
			ID:     string(rune('a' + i%26)),
			Name:   "Option",
			Scores: map[string]float64{"cost": 0.5},
		}
	}
	return options
}

func createManyScores(n int) map[string]float64 {
	scores := make(map[string]float64)
	for i := 0; i < n; i++ {
		scores[string(rune('a'+i%26))] = 0.5
	}
	return scores
}

func createManyCriteria(n int) []*types.DecisionCriterion {
	criteria := make([]*types.DecisionCriterion, n)
	weight := 1.0 / float64(n)
	for i := 0; i < n; i++ {
		criteria[i] = &types.DecisionCriterion{
			ID:     string(rune('a' + i%26)),
			Name:   "Criterion",
			Weight: weight,
		}
	}
	return criteria
}

func createManyAssumptions(n int) []string {
	assumptions := make([]string, n)
	for i := 0; i < n; i++ {
		assumptions[i] = "Assumption"
	}
	return assumptions
}
