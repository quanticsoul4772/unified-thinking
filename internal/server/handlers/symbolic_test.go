package handlers

import (
	"context"
	"testing"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/validation"
)

// TestSymbolicHandler_NewSymbolicHandler tests handler creation
func TestSymbolicHandler_NewSymbolicHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	reasoner := validation.NewSymbolicReasoner()

	handler := NewSymbolicHandler(reasoner, store)

	if handler == nil {
		t.Fatal("NewSymbolicHandler() should return a handler")
	}

	if handler.reasoner == nil {
		t.Error("Handler should have a reasoner")
	}

	if handler.storage == nil {
		t.Error("Handler should have storage")
	}
}

// TestSymbolicHandler_HandleProveTheorem tests theorem proving
func TestSymbolicHandler_HandleProveTheorem(t *testing.T) {
	store := storage.NewMemoryStorage()
	reasoner := validation.NewSymbolicReasoner()
	handler := NewSymbolicHandler(reasoner, store)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid theorem with modus ponens",
			params: map[string]interface{}{
				"name":       "Modus Ponens Test",
				"premises":   []string{"P", "P -> Q"},
				"conclusion": "Q",
			},
			wantErr: false,
		},
		{
			name: "valid theorem with multiple premises",
			params: map[string]interface{}{
				"name":       "Multiple Premises",
				"premises":   []string{"A", "B", "A & B -> C"},
				"conclusion": "C",
			},
			wantErr: false,
		},
		{
			name: "theorem without name",
			params: map[string]interface{}{
				"premises":   []string{"P", "P -> Q"},
				"conclusion": "Q",
			},
			wantErr: false,
		},
		{
			name: "missing conclusion",
			params: map[string]interface{}{
				"premises": []string{"P", "Q"},
			},
			wantErr: true,
		},
		{
			name: "empty conclusion",
			params: map[string]interface{}{
				"premises":   []string{"P"},
				"conclusion": "",
			},
			wantErr: true,
		},
		{
			name: "missing premises",
			params: map[string]interface{}{
				"conclusion": "Q",
			},
			wantErr: true,
		},
		{
			name: "empty premises",
			params: map[string]interface{}{
				"premises":   []string{},
				"conclusion": "Q",
			},
			wantErr: true,
		},
		{
			name: "invalid params structure",
			params: map[string]interface{}{
				"premises": "not an array",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := handler.HandleProveTheorem(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleProveTheorem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("HandleProveTheorem() should return result on success")
			}
		})
	}
}

// TestSymbolicHandler_HandleCheckConstraints tests constraint checking
func TestSymbolicHandler_HandleCheckConstraints(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid constraint system",
			params: map[string]interface{}{
				"symbols": []map[string]interface{}{
					{"name": "x", "type": "variable", "domain": "integer"},
					{"name": "y", "type": "variable", "domain": "integer"},
				},
				"constraints": []map[string]interface{}{
					{"type": "equality", "expression": "x = 5", "symbols": []string{"x"}},
					{"type": "inequality", "expression": "y > 0", "symbols": []string{"y"}},
				},
			},
			wantErr: false,
		},
		{
			name: "constraint system with constant",
			params: map[string]interface{}{
				"symbols": []map[string]interface{}{
					{"name": "PI", "type": "constant", "domain": "real"},
					{"name": "r", "type": "variable", "domain": "real"},
				},
				"constraints": []map[string]interface{}{
					{"type": "equality", "expression": "PI = 3.14", "symbols": []string{"PI"}},
					{"type": "inequality", "expression": "r > 0", "symbols": []string{"r"}},
				},
			},
			wantErr: false,
		},
		{
			name: "conflicting constraints",
			params: map[string]interface{}{
				"symbols": []map[string]interface{}{
					{"name": "x", "type": "variable", "domain": "integer"},
				},
				"constraints": []map[string]interface{}{
					{"type": "equality", "expression": "x = 5", "symbols": []string{"x"}},
					{"type": "inequality", "expression": "x > 10", "symbols": []string{"x"}},
				},
			},
			wantErr: false, // Should not error, but will report inconsistency
		},
		{
			name:    "missing symbols",
			params:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "empty symbols",
			params: map[string]interface{}{
				"symbols":     []map[string]interface{}{},
				"constraints": []map[string]interface{}{},
			},
			wantErr: true,
		},
		{
			name: "missing constraints",
			params: map[string]interface{}{
				"symbols": []map[string]interface{}{
					{"name": "x", "type": "variable", "domain": "integer"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty constraints",
			params: map[string]interface{}{
				"symbols": []map[string]interface{}{
					{"name": "x", "type": "variable", "domain": "integer"},
				},
				"constraints": []map[string]interface{}{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh handler for each test to avoid state pollution
			store := storage.NewMemoryStorage()
			reasoner := validation.NewSymbolicReasoner()
			handler := NewSymbolicHandler(reasoner, store)

			ctx := context.Background()
			result, err := handler.HandleCheckConstraints(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleCheckConstraints() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == nil {
				t.Error("HandleCheckConstraints() should return result on success")
			}
		})
	}
}

// TestSymbolicHandler_TheoremProvingWithComplexLogic tests complex theorems
func TestSymbolicHandler_TheoremProvingWithComplexLogic(t *testing.T) {
	store := storage.NewMemoryStorage()
	reasoner := validation.NewSymbolicReasoner()
	handler := NewSymbolicHandler(reasoner, store)

	tests := []struct {
		name       string
		premises   []string
		conclusion string
	}{
		{
			name:       "simple implication",
			premises:   []string{"P", "P -> Q"},
			conclusion: "Q",
		},
		{
			name:       "transitive implication",
			premises:   []string{"P -> Q", "Q -> R", "P"},
			conclusion: "R",
		},
		{
			name:       "conjunction",
			premises:   []string{"P", "Q"},
			conclusion: "P & Q",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			params := map[string]interface{}{
				"name":       tt.name,
				"premises":   tt.premises,
				"conclusion": tt.conclusion,
			}

			result, err := handler.HandleProveTheorem(ctx, params)

			if err != nil {
				t.Errorf("HandleProveTheorem() unexpected error: %v", err)
			}

			if result == nil {
				t.Error("HandleProveTheorem() should return result")
			}
		})
	}
}

// TestSymbolicHandler_ConstraintConsistencyChecking tests constraint consistency
func TestSymbolicHandler_ConstraintConsistencyChecking(t *testing.T) {
	store := storage.NewMemoryStorage()
	reasoner := validation.NewSymbolicReasoner()
	handler := NewSymbolicHandler(reasoner, store)

	ctx := context.Background()

	// Test consistent constraints
	params := map[string]interface{}{
		"symbols": []map[string]interface{}{
			{"name": "x", "type": "variable", "domain": "integer"},
			{"name": "y", "type": "variable", "domain": "integer"},
		},
		"constraints": []map[string]interface{}{
			{"type": "inequality", "expression": "x > 5", "symbols": []string{"x"}},
			{"type": "inequality", "expression": "y < 10", "symbols": []string{"y"}},
		},
	}

	result, err := handler.HandleCheckConstraints(ctx, params)

	if err != nil {
		t.Errorf("HandleCheckConstraints() unexpected error: %v", err)
	}

	if result == nil {
		t.Error("HandleCheckConstraints() should return result")
	}
}

// TestSymbolicHandler_HandleCheckConstraints_MoreCases tests additional constraint scenarios
func TestSymbolicHandler_HandleCheckConstraints_MoreCases(t *testing.T) {
	store := storage.NewMemoryStorage()
	reasoner := validation.NewSymbolicReasoner()
	handler := NewSymbolicHandler(reasoner, store)

	ctx := context.Background()

	// Test with function symbol
	params := map[string]interface{}{
		"symbols": []map[string]interface{}{
			{"name": "f", "type": "function", "domain": "real"},
			{"name": "x", "type": "variable", "domain": "real"},
		},
		"constraints": []map[string]interface{}{
			{"type": "equality", "expression": "f(x) = x^2", "symbols": []string{"f", "x"}},
		},
	}

	result, err := handler.HandleCheckConstraints(ctx, params)
	if err != nil {
		t.Errorf("HandleCheckConstraints() unexpected error: %v", err)
	}
	if result == nil {
		t.Error("HandleCheckConstraints() should return result")
	}

	// Test with boolean domain
	params2 := map[string]interface{}{
		"symbols": []map[string]interface{}{
			{"name": "p", "type": "variable", "domain": "boolean"},
			{"name": "q", "type": "variable", "domain": "boolean"},
		},
		"constraints": []map[string]interface{}{
			{"type": "implication", "expression": "p -> q", "symbols": []string{"p", "q"}},
		},
	}

	result2, err := handler.HandleCheckConstraints(ctx, params2)
	if err != nil {
		t.Errorf("HandleCheckConstraints() with boolean domain error: %v", err)
	}
	if result2 == nil {
		t.Error("HandleCheckConstraints() with boolean domain should return result")
	}
}

// TestSymbolicHandler_HandleProveTheorem_InvalidProof tests invalid proofs
func TestSymbolicHandler_HandleProveTheorem_InvalidProof(t *testing.T) {
	store := storage.NewMemoryStorage()
	reasoner := validation.NewSymbolicReasoner()
	handler := NewSymbolicHandler(reasoner, store)

	ctx := context.Background()

	// Invalid proof - conclusion doesn't follow
	params := map[string]interface{}{
		"name":       "Invalid proof",
		"premises":   []string{"P"},
		"conclusion": "Q",
	}

	result, err := handler.HandleProveTheorem(ctx, params)
	if err != nil {
		t.Errorf("HandleProveTheorem() unexpected error: %v", err)
	}
	// Should return result even if proof is invalid
	if result == nil {
		t.Error("HandleProveTheorem() should return result even for invalid proof")
	}
}
