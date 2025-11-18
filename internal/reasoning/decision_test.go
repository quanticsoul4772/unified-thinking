package reasoning

import (
	"testing"

	"unified-thinking/internal/types"
)

func TestNewDecisionMaker(t *testing.T) {
	dm := NewDecisionMaker()
	if dm == nil {
		t.Error("NewDecisionMaker returned nil")
	}
}

func TestCreateDecision(t *testing.T) {
	dm := NewDecisionMaker()

	tests := []struct {
		name        string
		question    string
		options     []*types.DecisionOption
		criteria    []*types.DecisionCriterion
		expectError bool
	}{
		{
			name:     "valid decision with two options",
			question: "Which technology to use?",
			options: []*types.DecisionOption{
				{ID: "opt1", Name: "React", Scores: map[string]float64{"perf": 0.8, "ease": 0.9}},
				{ID: "opt2", Name: "Vue", Scores: map[string]float64{"perf": 0.7, "ease": 0.95}},
			},
			criteria: []*types.DecisionCriterion{
				{ID: "perf", Name: "Performance", Weight: 0.6, Maximize: true},
				{ID: "ease", Name: "Ease of Use", Weight: 0.4, Maximize: true},
			},
			expectError: false,
		},
		{
			name:     "valid decision with minimize criterion",
			question: "Which vendor to choose?",
			options: []*types.DecisionOption{
				{ID: "opt1", Name: "Vendor A", Scores: map[string]float64{"cost": 0.8, "quality": 0.9}},
				{ID: "opt2", Name: "Vendor B", Scores: map[string]float64{"cost": 0.3, "quality": 0.7}},
			},
			criteria: []*types.DecisionCriterion{
				{ID: "cost", Name: "Cost", Weight: 0.5, Maximize: false}, // Minimize cost
				{ID: "quality", Name: "Quality", Weight: 0.5, Maximize: true},
			},
			expectError: false,
		},
		{
			name:        "no options error",
			question:    "Empty options?",
			options:     []*types.DecisionOption{},
			criteria:    []*types.DecisionCriterion{{ID: "c1", Name: "Crit", Weight: 1.0, Maximize: true}},
			expectError: true,
		},
		{
			name:     "no criteria error",
			question: "Empty criteria?",
			options: []*types.DecisionOption{
				{ID: "opt1", Name: "Option 1", Scores: map[string]float64{}},
			},
			criteria:    []*types.DecisionCriterion{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := dm.CreateDecision(tt.question, tt.options, tt.criteria)

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

			if decision == nil {
				t.Error("Expected decision but got nil")
				return
			}

			if decision.Question != tt.question {
				t.Errorf("Question mismatch: got %q, want %q", decision.Question, tt.question)
			}

			if decision.Confidence < 0 || decision.Confidence > 1 {
				t.Errorf("Confidence out of range: %.2f", decision.Confidence)
			}

			if decision.Recommendation == "" {
				t.Error("Expected recommendation but got empty string")
			}
		})
	}
}

func TestCalculateDecisionConfidence(t *testing.T) {
	dm := NewDecisionMaker()

	tests := []struct {
		name          string
		options       []*types.DecisionOption
		bestOption    *types.DecisionOption
		minConfidence float64
		maxConfidence float64
	}{
		{
			name: "clear winner with large margin",
			options: []*types.DecisionOption{
				{ID: "opt1", Name: "Best", TotalScore: 0.9},
				{ID: "opt2", Name: "Second", TotalScore: 0.3},
			},
			bestOption:    &types.DecisionOption{ID: "opt1", Name: "Best", TotalScore: 0.9},
			minConfidence: 0.7,
			maxConfidence: 1.0,
		},
		{
			name: "close competition",
			options: []*types.DecisionOption{
				{ID: "opt1", Name: "Best", TotalScore: 0.6},
				{ID: "opt2", Name: "Second", TotalScore: 0.55},
			},
			bestOption:    &types.DecisionOption{ID: "opt1", Name: "Best", TotalScore: 0.6},
			minConfidence: 0.5,
			maxConfidence: 0.6,
		},
		{
			name: "single option",
			options: []*types.DecisionOption{
				{ID: "opt1", Name: "Only", TotalScore: 0.8},
			},
			bestOption:    &types.DecisionOption{ID: "opt1", Name: "Only", TotalScore: 0.8},
			minConfidence: 0.5,
			maxConfidence: 0.5,
		},
		{
			name: "three options",
			options: []*types.DecisionOption{
				{ID: "opt1", Name: "Best", TotalScore: 0.8},
				{ID: "opt2", Name: "Second", TotalScore: 0.5},
				{ID: "opt3", Name: "Third", TotalScore: 0.3},
			},
			bestOption:    &types.DecisionOption{ID: "opt1", Name: "Best", TotalScore: 0.8},
			minConfidence: 0.6,
			maxConfidence: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := dm.calculateDecisionConfidence(tt.options, tt.bestOption)

			if confidence < tt.minConfidence || confidence > tt.maxConfidence {
				t.Errorf("Confidence %.2f not in expected range [%.2f, %.2f]",
					confidence, tt.minConfidence, tt.maxConfidence)
			}
		})
	}
}

func TestAddOption(t *testing.T) {
	dm := NewDecisionMaker()

	// Create initial decision
	decision := &types.Decision{
		ID:       "test-decision",
		Question: "Test question",
		Options:  []*types.DecisionOption{},
		Criteria: []*types.DecisionCriterion{
			{ID: "crit1", Name: "Criterion 1", Weight: 1.0, Maximize: true},
		},
	}

	// Add first option
	err := dm.AddOption(decision, "Option A", "Description A",
		map[string]float64{"crit1": 0.8},
		[]string{"pro1", "pro2"},
		[]string{"con1"},
	)
	if err != nil {
		t.Errorf("Failed to add first option: %v", err)
	}

	if len(decision.Options) != 1 {
		t.Errorf("Expected 1 option, got %d", len(decision.Options))
	}

	// Add second option
	err = dm.AddOption(decision, "Option B", "Description B",
		map[string]float64{"crit1": 0.6},
		[]string{"pro3"},
		[]string{"con2", "con3"},
	)
	if err != nil {
		t.Errorf("Failed to add second option: %v", err)
	}

	if len(decision.Options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(decision.Options))
	}

	// Verify option properties
	opt := decision.Options[0]
	if opt.Name != "Option A" {
		t.Errorf("Expected name 'Option A', got %q", opt.Name)
	}
	if opt.Description != "Description A" {
		t.Errorf("Expected description 'Description A', got %q", opt.Description)
	}
	if len(opt.Pros) != 2 {
		t.Errorf("Expected 2 pros, got %d", len(opt.Pros))
	}
	if len(opt.Cons) != 1 {
		t.Errorf("Expected 1 con, got %d", len(opt.Cons))
	}
}

func TestAddCriterion(t *testing.T) {
	dm := NewDecisionMaker()

	// Create initial decision
	decision := &types.Decision{
		ID:       "test-decision",
		Question: "Test question",
		Options:  []*types.DecisionOption{},
		Criteria: []*types.DecisionCriterion{},
	}

	// Add first criterion
	err := dm.AddCriterion(decision, "Performance", "System performance", 0.6, true)
	if err != nil {
		t.Errorf("Failed to add first criterion: %v", err)
	}

	if len(decision.Criteria) != 1 {
		t.Errorf("Expected 1 criterion, got %d", len(decision.Criteria))
	}

	// Add second criterion
	err = dm.AddCriterion(decision, "Cost", "Total cost of ownership", 0.4, false)
	if err != nil {
		t.Errorf("Failed to add second criterion: %v", err)
	}

	if len(decision.Criteria) != 2 {
		t.Errorf("Expected 2 criteria, got %d", len(decision.Criteria))
	}

	// Verify criterion properties
	crit := decision.Criteria[0]
	if crit.Name != "Performance" {
		t.Errorf("Expected name 'Performance', got %q", crit.Name)
	}
	if crit.Description != "System performance" {
		t.Errorf("Expected description 'System performance', got %q", crit.Description)
	}
	if crit.Weight != 0.6 {
		t.Errorf("Expected weight 0.6, got %.2f", crit.Weight)
	}
	if !crit.Maximize {
		t.Error("Expected Maximize to be true")
	}

	// Verify minimize criterion
	crit2 := decision.Criteria[1]
	if crit2.Maximize {
		t.Error("Expected Maximize to be false for cost")
	}
}

func TestNewProblemDecomposer(t *testing.T) {
	pd := NewProblemDecomposer()
	if pd == nil {
		t.Error("NewProblemDecomposer returned nil")
	}
}

func TestDecomposeProblem(t *testing.T) {
	pd := NewProblemDecomposer()

	tests := []struct {
		name            string
		problem         string
		minSubproblems  int
		minDependencies int
	}{
		{
			name:            "simple problem",
			problem:         "Implement a user authentication system",
			minSubproblems:  5,
			minDependencies: 4,
		},
		{
			name:            "complex problem",
			problem:         "Design and deploy a microservices architecture",
			minSubproblems:  5,
			minDependencies: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decomposition, err := pd.DecomposeProblem(tt.problem)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if decomposition == nil {
				t.Error("Expected decomposition but got nil")
				return
			}

			if len(decomposition.Subproblems) < tt.minSubproblems {
				t.Errorf("Expected at least %d subproblems, got %d",
					tt.minSubproblems, len(decomposition.Subproblems))
			}

			if len(decomposition.Dependencies) < tt.minDependencies {
				t.Errorf("Expected at least %d dependencies, got %d",
					tt.minDependencies, len(decomposition.Dependencies))
			}

			if len(decomposition.SolutionPath) != len(decomposition.Subproblems) {
				t.Errorf("Solution path length %d doesn't match subproblems %d",
					len(decomposition.SolutionPath), len(decomposition.Subproblems))
			}

			// Verify subproblem structure
			for _, sp := range decomposition.Subproblems {
				if sp.ID == "" {
					t.Error("Subproblem ID should not be empty")
				}
				if sp.Description == "" {
					t.Error("Subproblem description should not be empty")
				}
				if sp.Status != "pending" {
					t.Errorf("Initial status should be 'pending', got %q", sp.Status)
				}
			}
		})
	}
}

func TestUpdateSubproblemStatus(t *testing.T) {
	pd := NewProblemDecomposer()

	// Create a decomposition first
	decomposition, err := pd.DecomposeProblem("Test problem")
	if err != nil {
		t.Fatalf("Failed to create decomposition: %v", err)
	}

	// Get the first subproblem ID
	subproblemID := decomposition.Subproblems[0].ID

	tests := []struct {
		name        string
		spID        string
		status      string
		solution    string
		expectError bool
	}{
		{
			name:        "update to in_progress",
			spID:        subproblemID,
			status:      "in_progress",
			solution:    "",
			expectError: false,
		},
		{
			name:        "update to completed with solution",
			spID:        subproblemID,
			status:      "completed",
			solution:    "Implemented the solution",
			expectError: false,
		},
		{
			name:        "update non-existent subproblem",
			spID:        "nonexistent-id",
			status:      "completed",
			solution:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pd.UpdateSubproblemStatus(decomposition, tt.spID, tt.status, tt.solution)

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

			// Verify the update
			for _, sp := range decomposition.Subproblems {
				if sp.ID == tt.spID {
					if sp.Status != tt.status {
						t.Errorf("Status not updated: got %q, want %q", sp.Status, tt.status)
					}
					if tt.solution != "" && sp.Solution != tt.solution {
						t.Errorf("Solution not updated: got %q, want %q", sp.Solution, tt.solution)
					}
				}
			}
		})
	}
}

func TestDecisionWeightNormalization(t *testing.T) {
	dm := NewDecisionMaker()

	// Create decision with non-normalized weights
	options := []*types.DecisionOption{
		{ID: "opt1", Name: "A", Scores: map[string]float64{"c1": 1.0, "c2": 1.0}},
	}
	criteria := []*types.DecisionCriterion{
		{ID: "c1", Name: "Crit1", Weight: 3.0, Maximize: true}, // Non-normalized
		{ID: "c2", Name: "Crit2", Weight: 1.0, Maximize: true}, // Non-normalized
	}

	decision, err := dm.CreateDecision("Test", options, criteria)
	if err != nil {
		t.Fatalf("Failed to create decision: %v", err)
	}

	// Verify weights are normalized
	totalWeight := 0.0
	for _, c := range decision.Criteria {
		totalWeight += c.Weight
	}

	if totalWeight < 0.99 || totalWeight > 1.01 {
		t.Errorf("Weights not normalized: total = %.3f", totalWeight)
	}
}

func TestDecisionWithMinimizeCriteria(t *testing.T) {
	dm := NewDecisionMaker()

	// Test that minimize criteria properly invert scores
	options := []*types.DecisionOption{
		{ID: "opt1", Name: "Expensive", Scores: map[string]float64{"cost": 0.9}}, // High cost = bad
		{ID: "opt2", Name: "Cheap", Scores: map[string]float64{"cost": 0.1}},     // Low cost = good
	}
	criteria := []*types.DecisionCriterion{
		{ID: "cost", Name: "Cost", Weight: 1.0, Maximize: false}, // Minimize
	}

	decision, err := dm.CreateDecision("Which is better?", options, criteria)
	if err != nil {
		t.Fatalf("Failed to create decision: %v", err)
	}

	// Cheap option should win with minimize criteria
	if decision.Options[1].TotalScore <= decision.Options[0].TotalScore {
		t.Errorf("Minimize criteria not working: cheap=%.2f, expensive=%.2f",
			decision.Options[1].TotalScore, decision.Options[0].TotalScore)
	}
}

func TestDecisionConcurrency(t *testing.T) {
	dm := NewDecisionMaker()

	// Create multiple decisions concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			options := []*types.DecisionOption{
				{ID: "opt1", Name: "A", Scores: map[string]float64{"c1": 0.8}},
			}
			criteria := []*types.DecisionCriterion{
				{ID: "c1", Name: "Crit", Weight: 1.0, Maximize: true},
			}
			_, err := dm.CreateDecision("Test", options, criteria)
			if err != nil {
				t.Errorf("Concurrent decision creation failed: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
