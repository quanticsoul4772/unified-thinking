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

func TestUpdateBeliefWithEvidence(t *testing.T) {
	pr := NewProbabilisticReasoner()

	tests := []struct {
		name         string
		priorProb    float64
		evidence     *types.Evidence
		expectError  bool
		posteriorMin float64
		posteriorMax float64
	}{
		{
			name:      "supporting high quality evidence",
			priorProb: 0.5,
			evidence: &types.Evidence{
				ID:            "ev1",
				OverallScore:  0.9,
				SupportsClaim: true,
			},
			expectError:  false,
			posteriorMin: 0.5,
			posteriorMax: 1.0,
		},
		{
			name:      "refuting high quality evidence",
			priorProb: 0.5,
			evidence: &types.Evidence{
				ID:            "ev2",
				OverallScore:  0.9,
				SupportsClaim: false,
			},
			expectError:  false,
			posteriorMin: 0.0,
			posteriorMax: 0.5,
		},
		{
			name:      "weak supporting evidence",
			priorProb: 0.5,
			evidence: &types.Evidence{
				ID:            "ev3",
				OverallScore:  0.3,
				SupportsClaim: true,
			},
			expectError:  false,
			posteriorMin: 0.4,
			posteriorMax: 0.7,
		},
		{
			name:      "weak refuting evidence",
			priorProb: 0.5,
			evidence: &types.Evidence{
				ID:            "ev4",
				OverallScore:  0.3,
				SupportsClaim: false,
			},
			expectError:  false,
			posteriorMin: 0.3,
			posteriorMax: 0.6,
		},
		{
			name:      "high prior with supporting evidence",
			priorProb: 0.8,
			evidence: &types.Evidence{
				ID:            "ev5",
				OverallScore:  0.7,
				SupportsClaim: true,
			},
			expectError:  false,
			posteriorMin: 0.8,
			posteriorMax: 1.0,
		},
		{
			name:      "low prior with refuting evidence",
			priorProb: 0.2,
			evidence: &types.Evidence{
				ID:            "ev6",
				OverallScore:  0.8,
				SupportsClaim: false,
			},
			expectError:  false,
			posteriorMin: 0.0,
			posteriorMax: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh belief for each test
			belief, err := pr.CreateBelief("Test hypothesis", tt.priorProb)
			if err != nil {
				t.Fatalf("Failed to create belief: %v", err)
			}

			updated, err := pr.UpdateBeliefWithEvidence(belief.ID, tt.evidence)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if updated == nil {
				t.Error("Expected updated belief but got nil")
				return
			}

			if updated.Probability < tt.posteriorMin || updated.Probability > tt.posteriorMax {
				t.Errorf("Posterior %.3f not in expected range [%.3f, %.3f]",
					updated.Probability, tt.posteriorMin, tt.posteriorMax)
			}

			// Verify evidence was recorded
			found := false
			for _, evID := range updated.Evidence {
				if evID == tt.evidence.ID {
					found = true
					break
				}
			}
			if !found {
				t.Error("Evidence ID not recorded in belief")
			}
		})
	}
}

func TestUpdateBeliefWithEvidence_NotFound(t *testing.T) {
	pr := NewProbabilisticReasoner()

	evidence := &types.Evidence{
		ID:            "ev1",
		OverallScore:  0.8,
		SupportsClaim: true,
	}

	_, err := pr.UpdateBeliefWithEvidence("nonexistent", evidence)
	if err == nil {
		t.Error("Expected error for nonexistent belief but got none")
	}
}

func TestUpdateBeliefFull(t *testing.T) {
	pr := NewProbabilisticReasoner()

	tests := []struct {
		name              string
		priorProb         float64
		likelihoodIfTrue  float64
		likelihoodIfFalse float64
		expectError       bool
		posteriorMin      float64
		posteriorMax      float64
	}{
		{
			name:              "medical test - high sensitivity, high specificity",
			priorProb:         0.01, // 1% disease prevalence
			likelihoodIfTrue:  0.99, // 99% sensitivity
			likelihoodIfFalse: 0.05, // 5% false positive
			expectError:       false,
			posteriorMin:      0.1,
			posteriorMax:      0.2, // Base rate fallacy: ~17%
		},
		{
			name:              "equal likelihoods - no information",
			priorProb:         0.5,
			likelihoodIfTrue:  0.5,
			likelihoodIfFalse: 0.5,
			expectError:       false,
			posteriorMin:      0.5,
			posteriorMax:      0.5, // No change
		},
		{
			name:              "strong evidence for hypothesis",
			priorProb:         0.5,
			likelihoodIfTrue:  0.9,
			likelihoodIfFalse: 0.1,
			expectError:       false,
			posteriorMin:      0.8,
			posteriorMax:      1.0,
		},
		{
			name:              "strong evidence against hypothesis",
			priorProb:         0.5,
			likelihoodIfTrue:  0.1,
			likelihoodIfFalse: 0.9,
			expectError:       false,
			posteriorMin:      0.0,
			posteriorMax:      0.2,
		},
		{
			name:              "invalid likelihood - too high",
			priorProb:         0.5,
			likelihoodIfTrue:  1.5,
			likelihoodIfFalse: 0.5,
			expectError:       true,
		},
		{
			name:              "invalid likelihood - negative",
			priorProb:         0.5,
			likelihoodIfTrue:  0.5,
			likelihoodIfFalse: -0.1,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			belief, err := pr.CreateBelief("Test hypothesis", tt.priorProb)
			if err != nil {
				t.Fatalf("Failed to create belief: %v", err)
			}

			updated, err := pr.UpdateBeliefFull(belief.ID, "evidence-1", tt.likelihoodIfTrue, tt.likelihoodIfFalse)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if updated == nil {
				t.Error("Expected updated belief but got nil")
				return
			}

			if updated.Probability < tt.posteriorMin || updated.Probability > tt.posteriorMax {
				t.Errorf("Posterior %.3f not in expected range [%.3f, %.3f]",
					updated.Probability, tt.posteriorMin, tt.posteriorMax)
			}
		})
	}
}

func TestGetBelief(t *testing.T) {
	pr := NewProbabilisticReasoner()

	// Create a belief
	created, err := pr.CreateBelief("Test belief", 0.6)
	if err != nil {
		t.Fatalf("Failed to create belief: %v", err)
	}

	// Retrieve it
	retrieved, err := pr.GetBelief(created.ID)
	if err != nil {
		t.Errorf("Failed to get belief: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("ID mismatch: got %q, want %q", retrieved.ID, created.ID)
	}

	if retrieved.Statement != created.Statement {
		t.Errorf("Statement mismatch: got %q, want %q", retrieved.Statement, created.Statement)
	}

	// Try non-existent
	_, err = pr.GetBelief("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent belief but got none")
	}
}

func TestCombineBeliefs_EdgeCases(t *testing.T) {
	pr := NewProbabilisticReasoner()

	// Test empty belief IDs
	_, err := pr.CombineBeliefs([]string{}, "and")
	if err == nil {
		t.Error("Expected error for empty belief IDs but got none")
	}

	// Test nonexistent belief in AND
	_, err = pr.CombineBeliefs([]string{"nonexistent"}, "and")
	if err == nil {
		t.Error("Expected error for nonexistent belief in AND but got none")
	}

	// Test nonexistent belief in OR
	_, err = pr.CombineBeliefs([]string{"nonexistent"}, "or")
	if err == nil {
		t.Error("Expected error for nonexistent belief in OR but got none")
	}
}

func TestProbabilisticReasoner_Concurrency(t *testing.T) {
	pr := NewProbabilisticReasoner()

	// Create beliefs concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			_, err := pr.CreateBelief("Concurrent belief", 0.5)
			if err != nil {
				t.Errorf("Concurrent create failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestEstimateConfidence_EdgeCases(t *testing.T) {
	pr := NewProbabilisticReasoner()

	// Test with zero score evidence
	evidences := []*types.Evidence{
		{OverallScore: 0.0, SupportsClaim: true},
		{OverallScore: 0.0, SupportsClaim: false},
	}
	confidence := pr.EstimateConfidence(evidences)
	if confidence != 0.5 {
		t.Errorf("Expected 0.5 for zero score evidence, got %.3f", confidence)
	}
}
