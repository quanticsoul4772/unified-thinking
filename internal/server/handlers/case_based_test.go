package handlers

import (
	"context"
	"testing"

	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/storage"
)

// TestCaseBasedHandler_NewCaseBasedHandler tests handler creation
func TestCaseBasedHandler_NewCaseBasedHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	reasoner := reasoning.NewCaseBasedReasoner(store)

	handler := NewCaseBasedHandler(reasoner, store)

	if handler == nil {
		t.Fatal("NewCaseBasedHandler() should return a handler")
	}

	if handler.reasoner == nil {
		t.Error("Handler should have a reasoner")
	}

	if handler.storage == nil {
		t.Error("Handler should have storage")
	}
}

// TestCaseBasedHandler_HandleRetrieveCases tests case retrieval
func TestCaseBasedHandler_HandleRetrieveCases(t *testing.T) {
	store := storage.NewMemoryStorage()
	reasoner := reasoning.NewCaseBasedReasoner(store)
	handler := NewCaseBasedHandler(reasoner, store)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid request with full problem",
			params: map[string]interface{}{
				"problem": map[string]interface{}{
					"description": "Need to optimize database queries",
					"context":     "Web application with slow performance",
					"goals":       []string{"Reduce query time", "Improve user experience"},
					"constraints": []string{"Cannot change database schema"},
					"features": map[string]interface{}{
						"database":  "PostgreSQL",
						"framework": "Django",
					},
				},
				"domain":         "software-engineering",
				"max_cases":      5,
				"min_similarity": 0.7,
			},
			wantErr: false,
		},
		{
			name: "valid request with minimal problem",
			params: map[string]interface{}{
				"problem": map[string]interface{}{
					"description": "Simple problem description",
				},
			},
			wantErr: false,
		},
		{
			name: "valid request with default values",
			params: map[string]interface{}{
				"problem": map[string]interface{}{
					"description": "Problem without domain or limits",
				},
			},
			wantErr: false,
		},
		{
			name:    "missing problem",
			params:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "problem with empty description",
			params: map[string]interface{}{
				"problem": map[string]interface{}{
					"description": "",
				},
			},
			wantErr: true,
		},
		{
			name: "problem is null",
			params: map[string]interface{}{
				"problem": nil,
			},
			wantErr: true,
		},
		{
			name: "invalid params structure",
			params: map[string]interface{}{
				"problem": "not a map",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handler.HandleRetrieveCases(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleRetrieveCases() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("HandleRetrieveCases() should return result on success")
			}
		})
	}
}

// TestCaseBasedHandler_HandlePerformCBRCycle tests full CBR cycle
func TestCaseBasedHandler_HandlePerformCBRCycle(t *testing.T) {
	store := storage.NewMemoryStorage()
	reasoner := reasoning.NewCaseBasedReasoner(store)
	handler := NewCaseBasedHandler(reasoner, store)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid CBR cycle request - expects error when no cases",
			params: map[string]interface{}{
				"problem": map[string]interface{}{
					"description": "Need to implement authentication",
					"context":     "Web application needs secure login",
					"goals":       []string{"Secure authentication", "User-friendly login"},
					"constraints": []string{"Must comply with GDPR"},
				},
				"domain": "security",
			},
			wantErr: true, // Will error because no cases in storage
		},
		{
			name: "valid request without domain - expects error when no cases",
			params: map[string]interface{}{
				"problem": map[string]interface{}{
					"description": "Generic problem",
				},
			},
			wantErr: true, // Will error because no cases in storage
		},
		{
			name: "valid request with features - expects error when no cases",
			params: map[string]interface{}{
				"problem": map[string]interface{}{
					"description": "Feature-rich problem",
					"features": map[string]interface{}{
						"complexity": "high",
						"priority":   1,
					},
				},
			},
			wantErr: true, // Will error because no cases in storage
		},
		{
			name:    "missing problem",
			params:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "empty problem description",
			params: map[string]interface{}{
				"problem": map[string]interface{}{
					"description": "",
				},
			},
			wantErr: true,
		},
		{
			name: "null problem",
			params: map[string]interface{}{
				"problem": nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handler.HandlePerformCBRCycle(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandlePerformCBRCycle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("HandlePerformCBRCycle() should return result on success")
			}
		})
	}
}

// TestCaseBasedHandler_RetrieveCasesWithStoredCases tests retrieval with actual cases
func TestCaseBasedHandler_RetrieveCasesWithStoredCases(t *testing.T) {
	store := storage.NewMemoryStorage()
	reasoner := reasoning.NewCaseBasedReasoner(store)
	handler := NewCaseBasedHandler(reasoner, store)

	// Store a test case first
	testCase := &reasoning.Case{
		ID:     "case-1",
		Domain: "software-engineering",
		Problem: &reasoning.ProblemDescription{
			Description: "Database optimization problem",
			Context:     "Slow queries in production",
			Goals:       []string{"Improve performance"},
		},
		Solution: &reasoning.SolutionDescription{
			Description: "Add indexes and optimize queries",
			Approach:    "Query optimization",
			Steps:       []string{"Analyze slow queries", "Add indexes", "Test performance"},
		},
		SuccessRate: 0.9,
	}

	ctx := context.Background()
	err := reasoner.StoreCase(ctx, testCase)
	if err != nil {
		t.Fatalf("Failed to store test case: %v", err)
	}

	// Now retrieve similar cases
	params := map[string]interface{}{
		"problem": map[string]interface{}{
			"description": "Need to speed up database",
			"context":     "Performance issues",
		},
		"domain": "software-engineering",
	}

	result, err := handler.HandleRetrieveCases(ctx, params)

	if err != nil {
		t.Errorf("HandleRetrieveCases() unexpected error: %v", err)
	}

	if result == nil {
		t.Error("HandleRetrieveCases() should return result")
	}
}

// Note: TestCaseBasedHandler_PerformCBRCycleWithStoredCases is omitted
// because it requires testing the internal CBR reasoner logic which is covered
// in reasoning/case_based_test.go. The handler tests focus on parameter validation
// and error handling rather than the full CBR cycle logic.
