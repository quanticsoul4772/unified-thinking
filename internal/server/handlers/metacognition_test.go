package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/metacognition"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)

func TestNewMetacognitionHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	selfEvaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()

	handler := NewMetacognitionHandler(store, selfEvaluator, biasDetector, fallacyDetector)

	if handler == nil {
		t.Fatal("NewMetacognitionHandler returned nil")
	}
	if handler.storage == nil {
		t.Error("storage not initialized")
	}
	if handler.selfEvaluator == nil {
		t.Error("selfEvaluator not initialized")
	}
	if handler.biasDetector == nil {
		t.Error("biasDetector not initialized")
	}
	if handler.fallacyDetector == nil {
		t.Error("fallacyDetector not initialized")
	}
}

func TestMetacognitionHandler_HandleSelfEvaluate(t *testing.T) {
	store := storage.NewMemoryStorage()
	selfEvaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()
	handler := NewMetacognitionHandler(store, selfEvaluator, biasDetector, fallacyDetector)

	// Create test thought
	thought := types.NewThought().
		Content("This is a well-reasoned argument with clear logic.").
		Mode(types.ModeLinear).
		Confidence(0.8).
		Build()
	_ = store.StoreThought(thought)

	// Create test branch with thoughts
	branch := types.NewBranch().Build()
	_ = store.StoreBranch(branch)

	// Add thought to branch
	branchThought := types.NewThought().
		Content("Branch thought content").
		Mode(types.ModeTree).
		InBranch(branch.ID).
		Build()
	_ = store.StoreThought(branchThought)
	branch.Thoughts = append(branch.Thoughts, branchThought)

	tests := []struct {
		name    string
		input   SelfEvaluateRequest
		wantErr bool
	}{
		{
			name: "evaluate thought by ID",
			input: SelfEvaluateRequest{
				ThoughtID: thought.ID,
			},
			wantErr: false,
		},
		{
			name: "evaluate branch",
			input: SelfEvaluateRequest{
				BranchID: branch.ID,
			},
			wantErr: false,
		},
		{
			name: "missing both IDs",
			input: SelfEvaluateRequest{
				ThoughtID: "",
				BranchID:  "",
			},
			wantErr: true,
		},
		{
			name: "both IDs provided",
			input: SelfEvaluateRequest{
				ThoughtID: thought.ID,
				BranchID:  branch.ID,
			},
			wantErr: true,
		},
		{
			name: "non-existent thought",
			input: SelfEvaluateRequest{
				ThoughtID: "non-existent",
			},
			wantErr: true,
		},
		{
			name: "non-existent branch",
			input: SelfEvaluateRequest{
				BranchID: "non-existent",
			},
			wantErr: true,
		},
		{
			name: "thought ID too long",
			input: SelfEvaluateRequest{
				ThoughtID: strings.Repeat("a", MaxBranchIDLength+1),
			},
			wantErr: true,
		},
		{
			name: "branch ID too long",
			input: SelfEvaluateRequest{
				BranchID: strings.Repeat("a", MaxBranchIDLength+1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleSelfEvaluate(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleSelfEvaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("SelfEvaluateResponse should not be nil")
				}
				if response.Evaluation == nil {
					t.Error("Evaluation should not be nil")
				}
				if response.Status != "success" {
					t.Errorf("Status = %v, want success", response.Status)
				}
			}
		})
	}
}

func TestMetacognitionHandler_HandleDetectBiases(t *testing.T) {
	store := storage.NewMemoryStorage()
	selfEvaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()
	handler := NewMetacognitionHandler(store, selfEvaluator, biasDetector, fallacyDetector)

	// Create thought with potential biases
	thought := types.NewThought().
		Content("Everyone knows that this is true. Only an idiot would disagree. We should stick with tradition.").
		Mode(types.ModeLinear).
		Confidence(0.9).
		Build()
	_ = store.StoreThought(thought)

	// Create thought with confirmation bias indicators
	biasedThought := types.NewThought().
		Content("I've always believed this, so it must be true. I only looked at evidence that supports my view.").
		Mode(types.ModeLinear).
		Confidence(0.95).
		Build()
	_ = store.StoreThought(biasedThought)

	// Create branch with thoughts
	branch := types.NewBranch().Build()
	_ = store.StoreBranch(branch)

	branchThought := types.NewThought().
		Content("Ad hominem attack: You're wrong because you're stupid.").
		Mode(types.ModeTree).
		InBranch(branch.ID).
		Build()
	_ = store.StoreThought(branchThought)
	branch.Thoughts = append(branch.Thoughts, branchThought)

	tests := []struct {
		name    string
		input   DetectBiasesRequest
		wantErr bool
	}{
		{
			name: "detect biases in thought",
			input: DetectBiasesRequest{
				ThoughtID: thought.ID,
			},
			wantErr: false,
		},
		{
			name: "detect biases in biased thought",
			input: DetectBiasesRequest{
				ThoughtID: biasedThought.ID,
			},
			wantErr: false,
		},
		{
			name: "detect biases in branch",
			input: DetectBiasesRequest{
				BranchID: branch.ID,
			},
			wantErr: false,
		},
		{
			name: "missing both IDs",
			input: DetectBiasesRequest{
				ThoughtID: "",
				BranchID:  "",
			},
			wantErr: true,
		},
		{
			name: "both IDs provided",
			input: DetectBiasesRequest{
				ThoughtID: thought.ID,
				BranchID:  branch.ID,
			},
			wantErr: true,
		},
		{
			name: "non-existent thought",
			input: DetectBiasesRequest{
				ThoughtID: "non-existent",
			},
			wantErr: true,
		},
		{
			name: "non-existent branch",
			input: DetectBiasesRequest{
				BranchID: "non-existent",
			},
			wantErr: true,
		},
		{
			name: "thought ID too long",
			input: DetectBiasesRequest{
				ThoughtID: strings.Repeat("a", MaxBranchIDLength+1),
			},
			wantErr: true,
		},
		{
			name: "branch ID too long",
			input: DetectBiasesRequest{
				BranchID: strings.Repeat("a", MaxBranchIDLength+1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleDetectBiases(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleDetectBiases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("DetectBiasesResponse should not be nil")
				}
				if response.Biases == nil {
					t.Error("Biases should not be nil (even if empty)")
				}
				if response.Fallacies == nil {
					t.Error("Fallacies should not be nil (even if empty)")
				}
				if response.Combined == nil {
					t.Error("Combined should not be nil")
				}
				if response.Status != "success" {
					t.Errorf("Status = %v, want success", response.Status)
				}
				// Count should match combined length
				if response.Count != len(response.Combined) {
					t.Errorf("Count = %v, want %v", response.Count, len(response.Combined))
				}
			}
		})
	}
}

func TestValidateSelfEvaluateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *SelfEvaluateRequest
		wantErr bool
	}{
		{
			name:    "valid thought ID",
			req:     &SelfEvaluateRequest{ThoughtID: "thought_123"},
			wantErr: false,
		},
		{
			name:    "valid branch ID",
			req:     &SelfEvaluateRequest{BranchID: "branch_123"},
			wantErr: false,
		},
		{
			name:    "both missing",
			req:     &SelfEvaluateRequest{},
			wantErr: true,
		},
		{
			name:    "both provided",
			req:     &SelfEvaluateRequest{ThoughtID: "t", BranchID: "b"},
			wantErr: true,
		},
		{
			name:    "thought ID too long",
			req:     &SelfEvaluateRequest{ThoughtID: strings.Repeat("a", MaxBranchIDLength+1)},
			wantErr: true,
		},
		{
			name:    "branch ID too long",
			req:     &SelfEvaluateRequest{BranchID: strings.Repeat("a", MaxBranchIDLength+1)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSelfEvaluateRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSelfEvaluateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDetectBiasesRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *DetectBiasesRequest
		wantErr bool
	}{
		{
			name:    "valid thought ID",
			req:     &DetectBiasesRequest{ThoughtID: "thought_123"},
			wantErr: false,
		},
		{
			name:    "valid branch ID",
			req:     &DetectBiasesRequest{BranchID: "branch_123"},
			wantErr: false,
		},
		{
			name:    "both missing",
			req:     &DetectBiasesRequest{},
			wantErr: true,
		},
		{
			name:    "both provided",
			req:     &DetectBiasesRequest{ThoughtID: "t", BranchID: "b"},
			wantErr: true,
		},
		{
			name:    "thought ID too long",
			req:     &DetectBiasesRequest{ThoughtID: strings.Repeat("a", MaxBranchIDLength+1)},
			wantErr: true,
		},
		{
			name:    "branch ID too long",
			req:     &DetectBiasesRequest{BranchID: strings.Repeat("a", MaxBranchIDLength+1)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDetectBiasesRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDetectBiasesRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectedIssue_Structure(t *testing.T) {
	// Test that DetectedIssue struct works correctly
	issue := &DetectedIssue{
		Type:        "bias",
		Name:        "confirmation_bias",
		Category:    "cognitive",
		Description: "Seeking confirming evidence",
		Location:    "thought content",
		Example:     "I always believed this",
		Mitigation:  "Seek disconfirming evidence",
		Confidence:  0.9,
	}

	if issue.Type != "bias" {
		t.Errorf("Type = %v, want bias", issue.Type)
	}
	if issue.Confidence != 0.9 {
		t.Errorf("Confidence = %v, want 0.9", issue.Confidence)
	}
}

func TestMetacognitionHandler_Integration(t *testing.T) {
	store := storage.NewMemoryStorage()
	selfEvaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()
	handler := NewMetacognitionHandler(store, selfEvaluator, biasDetector, fallacyDetector)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create a thought
	thought := types.NewThought().
		Content("Logical argument with premises and conclusion.").
		Mode(types.ModeLinear).
		Confidence(0.7).
		Build()
	_ = store.StoreThought(thought)

	// Self-evaluate
	_, evalResp, err := handler.HandleSelfEvaluate(ctx, req, SelfEvaluateRequest{
		ThoughtID: thought.ID,
	})
	if err != nil {
		t.Fatalf("HandleSelfEvaluate() error = %v", err)
	}
	if evalResp == nil {
		t.Fatal("SelfEvaluateResponse should not be nil")
	}

	// Detect biases
	_, biasResp, err := handler.HandleDetectBiases(ctx, req, DetectBiasesRequest{
		ThoughtID: thought.ID,
	})
	if err != nil {
		t.Fatalf("HandleDetectBiases() error = %v", err)
	}
	if biasResp == nil {
		t.Fatal("DetectBiasesResponse should not be nil")
	}
}

func TestMetacognitionHandler_BranchWithEmptyThoughts(t *testing.T) {
	store := storage.NewMemoryStorage()
	selfEvaluator := metacognition.NewSelfEvaluator()
	biasDetector := metacognition.NewBiasDetector()
	fallacyDetector := validation.NewFallacyDetector()
	handler := NewMetacognitionHandler(store, selfEvaluator, biasDetector, fallacyDetector)

	// Create branch with empty thoughts slice
	branch := types.NewBranch().Build()
	branch.Thoughts = []*types.Thought{}
	_ = store.StoreBranch(branch)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Should handle empty thoughts gracefully
	_, response, err := handler.HandleDetectBiases(ctx, req, DetectBiasesRequest{
		BranchID: branch.ID,
	})

	if err != nil {
		t.Errorf("HandleDetectBiases() error = %v", err)
		return
	}

	if response == nil {
		t.Fatal("Response should not be nil")
	}
	if response.Status != "success" {
		t.Errorf("Status = %v, want success", response.Status)
	}
}
