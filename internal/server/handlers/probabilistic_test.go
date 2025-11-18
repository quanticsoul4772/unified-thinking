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

func TestNewProbabilisticHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	probabilisticReasoner := reasoning.NewProbabilisticReasoner()
	evidenceAnalyzer := analysis.NewEvidenceAnalyzer()
	contradictionDetector := analysis.NewContradictionDetector()

	handler := NewProbabilisticHandler(store, probabilisticReasoner, evidenceAnalyzer, contradictionDetector)

	if handler == nil {
		t.Fatal("NewProbabilisticHandler returned nil")
	}
	if handler.storage == nil {
		t.Error("storage not initialized")
	}
	if handler.probabilisticReasoner == nil {
		t.Error("probabilisticReasoner not initialized")
	}
	if handler.evidenceAnalyzer == nil {
		t.Error("evidenceAnalyzer not initialized")
	}
	if handler.contradictionDetector == nil {
		t.Error("contradictionDetector not initialized")
	}
}

func TestProbabilisticHandler_HandleProbabilisticReasoning_Create(t *testing.T) {
	store := storage.NewMemoryStorage()
	probabilisticReasoner := reasoning.NewProbabilisticReasoner()
	evidenceAnalyzer := analysis.NewEvidenceAnalyzer()
	contradictionDetector := analysis.NewContradictionDetector()
	handler := NewProbabilisticHandler(store, probabilisticReasoner, evidenceAnalyzer, contradictionDetector)

	tests := []struct {
		name    string
		input   ProbabilisticReasoningRequest
		wantErr bool
	}{
		{
			name: "valid create operation",
			input: ProbabilisticReasoningRequest{
				Operation: "create",
				Statement: "It will rain tomorrow",
				PriorProb: 0.3,
			},
			wantErr: false,
		},
		{
			name: "create with zero prior",
			input: ProbabilisticReasoningRequest{
				Operation: "create",
				Statement: "Unlikely event",
				PriorProb: 0.0,
			},
			wantErr: false,
		},
		{
			name: "create with max prior",
			input: ProbabilisticReasoningRequest{
				Operation: "create",
				Statement: "Certain event",
				PriorProb: 1.0,
			},
			wantErr: false,
		},
		{
			name: "missing statement",
			input: ProbabilisticReasoningRequest{
				Operation: "create",
				Statement: "",
				PriorProb: 0.3,
			},
			wantErr: true,
		},
		{
			name: "invalid prior too high",
			input: ProbabilisticReasoningRequest{
				Operation: "create",
				Statement: "Test",
				PriorProb: 1.5,
			},
			wantErr: true,
		},
		{
			name: "invalid prior negative",
			input: ProbabilisticReasoningRequest{
				Operation: "create",
				Statement: "Test",
				PriorProb: -0.1,
			},
			wantErr: true,
		},
		{
			name: "statement too long",
			input: ProbabilisticReasoningRequest{
				Operation: "create",
				Statement: strings.Repeat("a", MaxContentLength+1),
				PriorProb: 0.3,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleProbabilisticReasoning(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleProbabilisticReasoning() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("ProbabilisticReasoningResponse should not be nil")
				}
				if response.Belief == nil {
					t.Error("Belief should not be nil for create operation")
				}
				if response.Operation != "create" {
					t.Errorf("Operation = %v, want create", response.Operation)
				}
				if response.Status != "success" {
					t.Errorf("Status = %v, want success", response.Status)
				}
			}
		})
	}
}

func TestProbabilisticHandler_HandleProbabilisticReasoning_Update(t *testing.T) {
	store := storage.NewMemoryStorage()
	probabilisticReasoner := reasoning.NewProbabilisticReasoner()
	evidenceAnalyzer := analysis.NewEvidenceAnalyzer()
	contradictionDetector := analysis.NewContradictionDetector()
	handler := NewProbabilisticHandler(store, probabilisticReasoner, evidenceAnalyzer, contradictionDetector)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// First create a belief
	_, createResp, err := handler.HandleProbabilisticReasoning(ctx, req, ProbabilisticReasoningRequest{
		Operation: "create",
		Statement: "It will rain",
		PriorProb: 0.3,
	})
	if err != nil {
		t.Fatalf("Failed to create belief: %v", err)
	}
	beliefID := createResp.Belief.ID

	tests := []struct {
		name    string
		input   ProbabilisticReasoningRequest
		wantErr bool
	}{
		{
			name: "valid update",
			input: ProbabilisticReasoningRequest{
				Operation:    "update",
				BeliefID:     beliefID,
				EvidenceID:   "evidence_1",
				Likelihood:   0.8,
				EvidenceProb: 0.5,
			},
			wantErr: false,
		},
		{
			name: "missing belief ID",
			input: ProbabilisticReasoningRequest{
				Operation:    "update",
				BeliefID:     "",
				EvidenceID:   "evidence_1",
				Likelihood:   0.8,
				EvidenceProb: 0.5,
			},
			wantErr: true,
		},
		{
			name: "missing evidence ID",
			input: ProbabilisticReasoningRequest{
				Operation:    "update",
				BeliefID:     beliefID,
				EvidenceID:   "",
				Likelihood:   0.8,
				EvidenceProb: 0.5,
			},
			wantErr: true,
		},
		{
			name: "invalid likelihood",
			input: ProbabilisticReasoningRequest{
				Operation:    "update",
				BeliefID:     beliefID,
				EvidenceID:   "evidence_1",
				Likelihood:   1.5,
				EvidenceProb: 0.5,
			},
			wantErr: true,
		},
		{
			name: "invalid evidence prob zero",
			input: ProbabilisticReasoningRequest{
				Operation:    "update",
				BeliefID:     beliefID,
				EvidenceID:   "evidence_1",
				Likelihood:   0.8,
				EvidenceProb: 0.0,
			},
			wantErr: true,
		},
		{
			name: "invalid evidence prob too high",
			input: ProbabilisticReasoningRequest{
				Operation:    "update",
				BeliefID:     beliefID,
				EvidenceID:   "evidence_1",
				Likelihood:   0.8,
				EvidenceProb: 1.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, response, err := handler.HandleProbabilisticReasoning(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleProbabilisticReasoning() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("Response should not be nil")
				}
				if response.Operation != "update" {
					t.Errorf("Operation = %v, want update", response.Operation)
				}
			}
		})
	}
}

func TestProbabilisticHandler_HandleProbabilisticReasoning_Get(t *testing.T) {
	store := storage.NewMemoryStorage()
	probabilisticReasoner := reasoning.NewProbabilisticReasoner()
	evidenceAnalyzer := analysis.NewEvidenceAnalyzer()
	contradictionDetector := analysis.NewContradictionDetector()
	handler := NewProbabilisticHandler(store, probabilisticReasoner, evidenceAnalyzer, contradictionDetector)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create a belief first
	_, createResp, _ := handler.HandleProbabilisticReasoning(ctx, req, ProbabilisticReasoningRequest{
		Operation: "create",
		Statement: "Test belief",
		PriorProb: 0.5,
	})
	beliefID := createResp.Belief.ID

	tests := []struct {
		name    string
		input   ProbabilisticReasoningRequest
		wantErr bool
	}{
		{
			name: "valid get",
			input: ProbabilisticReasoningRequest{
				Operation: "get",
				BeliefID:  beliefID,
			},
			wantErr: false,
		},
		{
			name: "missing belief ID",
			input: ProbabilisticReasoningRequest{
				Operation: "get",
				BeliefID:  "",
			},
			wantErr: true,
		},
		{
			name: "non-existent belief",
			input: ProbabilisticReasoningRequest{
				Operation: "get",
				BeliefID:  "non-existent",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, response, err := handler.HandleProbabilisticReasoning(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleProbabilisticReasoning() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("Response should not be nil")
				}
				if response.Operation != "get" {
					t.Errorf("Operation = %v, want get", response.Operation)
				}
			}
		})
	}
}

func TestProbabilisticHandler_HandleProbabilisticReasoning_Combine(t *testing.T) {
	store := storage.NewMemoryStorage()
	probabilisticReasoner := reasoning.NewProbabilisticReasoner()
	evidenceAnalyzer := analysis.NewEvidenceAnalyzer()
	contradictionDetector := analysis.NewContradictionDetector()
	handler := NewProbabilisticHandler(store, probabilisticReasoner, evidenceAnalyzer, contradictionDetector)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create multiple beliefs
	_, resp1, _ := handler.HandleProbabilisticReasoning(ctx, req, ProbabilisticReasoningRequest{
		Operation: "create",
		Statement: "Belief 1",
		PriorProb: 0.5,
	})
	_, resp2, _ := handler.HandleProbabilisticReasoning(ctx, req, ProbabilisticReasoningRequest{
		Operation: "create",
		Statement: "Belief 2",
		PriorProb: 0.7,
	})

	tests := []struct {
		name    string
		input   ProbabilisticReasoningRequest
		wantErr bool
	}{
		{
			name: "combine with AND",
			input: ProbabilisticReasoningRequest{
				Operation: "combine",
				BeliefIDs: []string{resp1.Belief.ID, resp2.Belief.ID},
				CombineOp: "and",
			},
			wantErr: false,
		},
		{
			name: "combine with OR",
			input: ProbabilisticReasoningRequest{
				Operation: "combine",
				BeliefIDs: []string{resp1.Belief.ID, resp2.Belief.ID},
				CombineOp: "or",
			},
			wantErr: false,
		},
		{
			name: "missing belief IDs",
			input: ProbabilisticReasoningRequest{
				Operation: "combine",
				BeliefIDs: []string{},
				CombineOp: "and",
			},
			wantErr: true,
		},
		{
			name: "invalid combine operation",
			input: ProbabilisticReasoningRequest{
				Operation: "combine",
				BeliefIDs: []string{resp1.Belief.ID},
				CombineOp: "invalid",
			},
			wantErr: true,
		},
		{
			name: "too many belief IDs",
			input: ProbabilisticReasoningRequest{
				Operation: "combine",
				BeliefIDs: createManyBeliefIDs(55),
				CombineOp: "and",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, response, err := handler.HandleProbabilisticReasoning(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleProbabilisticReasoning() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("Response should not be nil")
				}
				if response.Operation != "combine" {
					t.Errorf("Operation = %v, want combine", response.Operation)
				}
			}
		})
	}
}

func TestProbabilisticHandler_HandleProbabilisticReasoning_InvalidOperation(t *testing.T) {
	store := storage.NewMemoryStorage()
	probabilisticReasoner := reasoning.NewProbabilisticReasoner()
	evidenceAnalyzer := analysis.NewEvidenceAnalyzer()
	contradictionDetector := analysis.NewContradictionDetector()
	handler := NewProbabilisticHandler(store, probabilisticReasoner, evidenceAnalyzer, contradictionDetector)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, _, err := handler.HandleProbabilisticReasoning(ctx, req, ProbabilisticReasoningRequest{
		Operation: "invalid_operation",
	})

	if err == nil {
		t.Error("Expected error for invalid operation")
	}
}

func TestProbabilisticHandler_HandleAssessEvidence(t *testing.T) {
	store := storage.NewMemoryStorage()
	probabilisticReasoner := reasoning.NewProbabilisticReasoner()
	evidenceAnalyzer := analysis.NewEvidenceAnalyzer()
	contradictionDetector := analysis.NewContradictionDetector()
	handler := NewProbabilisticHandler(store, probabilisticReasoner, evidenceAnalyzer, contradictionDetector)

	tests := []struct {
		name    string
		input   AssessEvidenceRequest
		wantErr bool
	}{
		{
			name: "valid supporting evidence",
			input: AssessEvidenceRequest{
				Content:       "Experimental data shows positive correlation",
				Source:        "Scientific Journal",
				ClaimID:       "claim_1",
				SupportsClaim: true,
			},
			wantErr: false,
		},
		{
			name: "valid contradicting evidence",
			input: AssessEvidenceRequest{
				Content:       "Study found no effect",
				Source:        "Research Paper",
				SupportsClaim: false,
			},
			wantErr: false,
		},
		{
			name: "missing content",
			input: AssessEvidenceRequest{
				Content:       "",
				Source:        "Source",
				SupportsClaim: true,
			},
			wantErr: true,
		},
		{
			name: "missing source",
			input: AssessEvidenceRequest{
				Content:       "Content",
				Source:        "",
				SupportsClaim: true,
			},
			wantErr: true,
		},
		{
			name: "content too long",
			input: AssessEvidenceRequest{
				Content:       strings.Repeat("a", MaxContentLength+1),
				Source:        "Source",
				SupportsClaim: true,
			},
			wantErr: true,
		},
		{
			name: "source too long",
			input: AssessEvidenceRequest{
				Content:       "Content",
				Source:        strings.Repeat("a", MaxQueryLength+1),
				SupportsClaim: true,
			},
			wantErr: true,
		},
		{
			name: "claim ID too long",
			input: AssessEvidenceRequest{
				Content:       "Content",
				Source:        "Source",
				ClaimID:       strings.Repeat("a", MaxBranchIDLength+1),
				SupportsClaim: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleAssessEvidence(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleAssessEvidence() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("AssessEvidenceResponse should not be nil")
				}
				if response.Evidence == nil {
					t.Error("Evidence should not be nil")
				}
				if response.Status != "success" {
					t.Errorf("Status = %v, want success", response.Status)
				}
			}
		})
	}
}

func TestProbabilisticHandler_HandleDetectContradictions(t *testing.T) {
	store := storage.NewMemoryStorage()
	probabilisticReasoner := reasoning.NewProbabilisticReasoner()
	evidenceAnalyzer := analysis.NewEvidenceAnalyzer()
	contradictionDetector := analysis.NewContradictionDetector()
	handler := NewProbabilisticHandler(store, probabilisticReasoner, evidenceAnalyzer, contradictionDetector)

	// Create test thoughts
	thought1 := types.NewThought().
		Content("X is true").
		Mode(types.ModeLinear).
		Build()
	_ = store.StoreThought(thought1)

	thought2 := types.NewThought().
		Content("X is false").
		Mode(types.ModeLinear).
		Build()
	_ = store.StoreThought(thought2)

	// Create test branch
	branch := types.NewBranch().Build()
	branch.Thoughts = []*types.Thought{thought1, thought2}
	_ = store.StoreBranch(branch)

	tests := []struct {
		name    string
		input   DetectContradictionsRequest
		wantErr bool
	}{
		{
			name: "detect by thought IDs",
			input: DetectContradictionsRequest{
				ThoughtIDs: []string{thought1.ID, thought2.ID},
			},
			wantErr: false,
		},
		{
			name: "detect in branch",
			input: DetectContradictionsRequest{
				BranchID: branch.ID,
			},
			wantErr: false,
		},
		{
			name: "detect by mode",
			input: DetectContradictionsRequest{
				Mode: "linear",
			},
			wantErr: false,
		},
		{
			name: "detect all (empty request)",
			input: DetectContradictionsRequest{},
			wantErr: false,
		},
		{
			name: "non-existent thought ID",
			input: DetectContradictionsRequest{
				ThoughtIDs: []string{"non-existent"},
			},
			wantErr: true,
		},
		{
			name: "non-existent branch",
			input: DetectContradictionsRequest{
				BranchID: "non-existent",
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			input: DetectContradictionsRequest{
				Mode: "invalid",
			},
			wantErr: true,
		},
		{
			name: "too many thought IDs",
			input: DetectContradictionsRequest{
				ThoughtIDs: createManyThoughtIDs(105),
			},
			wantErr: true,
		},
		{
			name: "branch ID too long",
			input: DetectContradictionsRequest{
				BranchID: strings.Repeat("a", MaxBranchIDLength+1),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleDetectContradictions(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleDetectContradictions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("DetectContradictionsResponse should not be nil")
				}
				if response.Contradictions == nil {
					t.Error("Contradictions should not be nil (even if empty)")
				}
				if response.Count != len(response.Contradictions) {
					t.Errorf("Count = %v, want %v", response.Count, len(response.Contradictions))
				}
				if response.Status != "success" {
					t.Errorf("Status = %v, want success", response.Status)
				}
			}
		})
	}
}

func TestValidateProbabilisticReasoningRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *ProbabilisticReasoningRequest
		wantErr bool
	}{
		{
			name: "valid create",
			req: &ProbabilisticReasoningRequest{
				Operation: "create",
				Statement: "Test",
				PriorProb: 0.5,
			},
			wantErr: false,
		},
		{
			name: "invalid UTF-8 statement",
			req: &ProbabilisticReasoningRequest{
				Operation: "create",
				Statement: string([]byte{0xff, 0xfe}),
				PriorProb: 0.5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProbabilisticReasoningRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProbabilisticReasoningRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAssessEvidenceRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *AssessEvidenceRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &AssessEvidenceRequest{
				Content: "Content",
				Source:  "Source",
			},
			wantErr: false,
		},
		{
			name: "invalid UTF-8 content",
			req: &AssessEvidenceRequest{
				Content: string([]byte{0xff, 0xfe}),
				Source:  "Source",
			},
			wantErr: true,
		},
		{
			name: "invalid UTF-8 source",
			req: &AssessEvidenceRequest{
				Content: "Content",
				Source:  string([]byte{0xff, 0xfe}),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAssessEvidenceRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAssessEvidenceRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDetectContradictionsRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *DetectContradictionsRequest
		wantErr bool
	}{
		{
			name:    "empty request (all thoughts)",
			req:     &DetectContradictionsRequest{},
			wantErr: false,
		},
		{
			name:    "valid mode",
			req:     &DetectContradictionsRequest{Mode: "tree"},
			wantErr: false,
		},
		{
			name:    "invalid mode",
			req:     &DetectContradictionsRequest{Mode: "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDetectContradictionsRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDetectContradictionsRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper functions
func createManyBeliefIDs(n int) []string {
	ids := make([]string, n)
	for i := 0; i < n; i++ {
		ids[i] = "belief"
	}
	return ids
}

func createManyThoughtIDs(n int) []string {
	ids := make([]string, n)
	for i := 0; i < n; i++ {
		ids[i] = "thought"
	}
	return ids
}
