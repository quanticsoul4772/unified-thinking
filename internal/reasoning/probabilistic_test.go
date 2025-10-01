package reasoning

import (
	"testing"

	"unified-thinking/internal/types"
)

func TestNewProbabilisticReasoner(t *testing.T) {
	pr := NewProbabilisticReasoner()
	if pr == nil {
		t.Fatal("NewProbabilisticReasoner returned nil")
	}
	if pr.beliefs == nil {
		t.Error("beliefs map not initialized")
	}
}

func TestCreateBelief(t *testing.T) {
	pr := NewProbabilisticReasoner()

	tests := []struct {
		name      string
		statement string
		priorProb float64
		wantErr   bool
	}{
		{
			name:      "valid belief",
			statement: "It will rain tomorrow",
			priorProb: 0.3,
			wantErr:   false,
		},
		{
			name:      "probability too low",
			statement: "Test",
			priorProb: -0.1,
			wantErr:   true,
		},
		{
			name:      "probability too high",
			statement: "Test",
			priorProb: 1.5,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			belief, err := pr.CreateBelief(tt.statement, tt.priorProb)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBelief() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if belief == nil {
					t.Error("CreateBelief() returned nil belief")
					return
				}
				if belief.Statement != tt.statement {
					t.Errorf("Statement = %v, want %v", belief.Statement, tt.statement)
				}
				if belief.Probability != tt.priorProb {
					t.Errorf("Probability = %v, want %v", belief.Probability, tt.priorProb)
				}
				if belief.PriorProb != tt.priorProb {
					t.Errorf("PriorProb = %v, want %v", belief.PriorProb, tt.priorProb)
				}
			}
		})
	}
}

func TestUpdateBelief(t *testing.T) {
	pr := NewProbabilisticReasoner()
	belief, _ := pr.CreateBelief("Test hypothesis", 0.5)

	tests := []struct {
		name         string
		beliefID     string
		evidenceID   string
		likelihood   float64
		evidenceProb float64
		wantErr      bool
	}{
		{
			name:         "valid update",
			beliefID:     belief.ID,
			evidenceID:   "evidence-1",
			likelihood:   0.8,
			evidenceProb: 0.5,
			wantErr:      false,
		},
		{
			name:         "belief not found",
			beliefID:     "nonexistent",
			evidenceID:   "evidence-1",
			likelihood:   0.8,
			evidenceProb: 0.5,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := pr.UpdateBelief(tt.beliefID, tt.evidenceID, tt.likelihood, tt.evidenceProb)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateBelief() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if updated == nil {
					t.Error("UpdateBelief() returned nil")
					return
				}
				if updated.Probability < 0 || updated.Probability > 1 {
					t.Errorf("Updated probability %v is out of range [0,1]", updated.Probability)
				}
			}
		})
	}
}

func TestCombineBeliefs(t *testing.T) {
	pr := NewProbabilisticReasoner()
	belief1, _ := pr.CreateBelief("Belief 1", 0.7)
	belief2, _ := pr.CreateBelief("Belief 2", 0.6)

	tests := []struct {
		name      string
		beliefIDs []string
		operation string
		wantErr   bool
	}{
		{
			name:      "combine with AND",
			beliefIDs: []string{belief1.ID, belief2.ID},
			operation: "and",
			wantErr:   false,
		},
		{
			name:      "combine with OR",
			beliefIDs: []string{belief1.ID, belief2.ID},
			operation: "or",
			wantErr:   false,
		},
		{
			name:      "unknown operation",
			beliefIDs: []string{belief1.ID, belief2.ID},
			operation: "xor",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := pr.CombineBeliefs(tt.beliefIDs, tt.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("CombineBeliefs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if result < 0 || result > 1 {
					t.Errorf("Combined probability %v is out of range [0,1]", result)
				}
			}
		})
	}
}

func TestEstimateConfidence(t *testing.T) {
	pr := NewProbabilisticReasoner()

	tests := []struct {
		name      string
		evidences []*types.Evidence
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "no evidence",
			evidences: []*types.Evidence{},
			wantMin:   0.5,
			wantMax:   0.5,
		},
		{
			name: "supporting evidence",
			evidences: []*types.Evidence{
				{OverallScore: 0.8, SupportsClaim: true},
				{OverallScore: 0.7, SupportsClaim: true},
			},
			wantMin: 0.9,
			wantMax: 1.0,
		},
		{
			name: "mixed evidence",
			evidences: []*types.Evidence{
				{OverallScore: 0.8, SupportsClaim: true},
				{OverallScore: 0.6, SupportsClaim: false},
			},
			wantMin: 0.4,
			wantMax: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := pr.EstimateConfidence(tt.evidences)
			if confidence < tt.wantMin || confidence > tt.wantMax {
				t.Errorf("EstimateConfidence() = %v, want range [%v, %v]", confidence, tt.wantMin, tt.wantMax)
			}
		})
	}
}
