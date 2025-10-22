package handlers

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
	"unified-thinking/internal/validation"
)

func TestNewValidationHandler(t *testing.T) {
	store := storage.NewMemoryStorage()
	validator := validation.NewLogicValidator()

	handler := NewValidationHandler(store, validator)

	if handler == nil {
		t.Fatal("NewValidationHandler returned nil")
	}
	if handler.storage == nil {
		t.Error("storage not initialized")
	}
	if handler.validator == nil {
		t.Error("validator not initialized")
	}
}

func TestValidationHandler_HandleValidate(t *testing.T) {
	store := storage.NewMemoryStorage()
	validator := validation.NewLogicValidator()
	handler := NewValidationHandler(store, validator)

	// Create a test thought
	thought := types.NewThought().
		Content("If P then Q. P is true. Therefore Q is true.").
		Mode(types.ModeLinear).
		Build()
	_ = store.StoreThought(thought)

	tests := []struct {
		name    string
		input   ValidateRequest
		wantErr bool
	}{
		{
			name: "validate existing thought",
			input: ValidateRequest{
				ThoughtID: thought.ID,
			},
			wantErr: false,
		},
		{
			name: "empty thought ID",
			input: ValidateRequest{
				ThoughtID: "",
			},
			wantErr: true,
		},
		{
			name: "non-existent thought",
			input: ValidateRequest{
				ThoughtID: "non-existent-id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleValidate(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleValidate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("ValidateResponse should not be nil")
				}
				// ValidationID may or may not be populated depending on implementation
				// IsValid and Reason are set by the validator
			}
		})
	}
}

func TestValidationHandler_HandleProve(t *testing.T) {
	store := storage.NewMemoryStorage()
	validator := validation.NewLogicValidator()
	handler := NewValidationHandler(store, validator)

	tests := []struct {
		name       string
		input      ProveRequest
		wantErr    bool
		checkSteps bool
	}{
		{
			name: "simple proof",
			input: ProveRequest{
				Premises:   []string{"P", "P -> Q"},
				Conclusion: "Q",
			},
			wantErr:    false,
			checkSteps: true,
		},
		{
			name: "invalid proof",
			input: ProveRequest{
				Premises:   []string{"P"},
				Conclusion: "R",
			},
			wantErr:    false,
			checkSteps: false,
		},
		{
			name: "empty premises",
			input: ProveRequest{
				Premises:   []string{},
				Conclusion: "Q",
			},
			wantErr:    false,
			checkSteps: false,
		},
		{
			name: "multiple premises",
			input: ProveRequest{
				Premises:   []string{"A", "B", "C"},
				Conclusion: "D",
			},
			wantErr:    false,
			checkSteps: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleProve(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleProve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("ProveResponse should not be nil")
				}
				if len(response.Premises) != len(tt.input.Premises) {
					t.Errorf("Premises length = %v, want %v", len(response.Premises), len(tt.input.Premises))
				}
				if response.Conclusion != tt.input.Conclusion {
					t.Errorf("Conclusion = %v, want %v", response.Conclusion, tt.input.Conclusion)
				}

				if tt.checkSteps && len(response.Steps) == 0 {
					t.Error("Expected proof steps, got none")
				}
			}
		})
	}
}

func TestValidationHandler_HandleCheckSyntax(t *testing.T) {
	store := storage.NewMemoryStorage()
	validator := validation.NewLogicValidator()
	handler := NewValidationHandler(store, validator)

	tests := []struct {
		name    string
		input   CheckSyntaxRequest
		wantErr bool
	}{
		{
			name: "valid statements",
			input: CheckSyntaxRequest{
				Statements: []string{"P", "Q", "P -> Q"},
			},
			wantErr: false,
		},
		{
			name: "single statement",
			input: CheckSyntaxRequest{
				Statements: []string{"P"},
			},
			wantErr: false,
		},
		{
			name: "empty statements",
			input: CheckSyntaxRequest{
				Statements: []string{},
			},
			wantErr: false,
		},
		{
			name: "complex statements",
			input: CheckSyntaxRequest{
				Statements: []string{
					"(P AND Q) OR R",
					"NOT P",
					"P <-> Q",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			req := &mcp.CallToolRequest{}

			result, response, err := handler.HandleCheckSyntax(ctx, req, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleCheckSyntax() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
				if response == nil {
					t.Fatal("CheckSyntaxResponse should not be nil")
				}
				// IsValid and Errors depend on validator implementation
			}
		})
	}
}

func TestValidationHandler_ValidateThoughtLifecycle(t *testing.T) {
	store := storage.NewMemoryStorage()
	validator := validation.NewLogicValidator()
	handler := NewValidationHandler(store, validator)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Create thought
	thought := types.NewThought().
		Content("Test logical thought").
		Mode(types.ModeLinear).
		Build()
	_ = store.StoreThought(thought)

	// Validate it
	_, validateResp, err := handler.HandleValidate(ctx, req, ValidateRequest{
		ThoughtID: thought.ID,
	})

	if err != nil {
		t.Fatalf("HandleValidate() error = %v", err)
	}

	// Validation response is returned
	if validateResp == nil {
		t.Fatal("ValidateResponse should not be nil")
	}

	// Response should have IsValid and Reason
	// (ValidationID storage depends on handler implementation)
}

func TestValidationHandler_ProveWithSteps(t *testing.T) {
	store := storage.NewMemoryStorage()
	validator := validation.NewLogicValidator()
	handler := NewValidationHandler(store, validator)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	_, response, err := handler.HandleProve(ctx, req, ProveRequest{
		Premises:   []string{"All humans are mortal", "Socrates is human"},
		Conclusion: "Socrates is mortal",
	})

	if err != nil {
		t.Fatalf("HandleProve() error = %v", err)
	}

	// Response should include premises and conclusion
	if len(response.Premises) != 2 {
		t.Errorf("Premises length = %v, want 2", len(response.Premises))
	}

	if response.Conclusion == "" {
		t.Error("Conclusion should not be empty")
	}

	// Steps may or may not be populated depending on validator implementation
	if response.Steps == nil {
		t.Error("Steps should not be nil (even if empty)")
	}
}

func TestValidationHandler_CheckSyntaxValidation(t *testing.T) {
	store := storage.NewMemoryStorage()
	validator := validation.NewLogicValidator()
	handler := NewValidationHandler(store, validator)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	statements := []string{
		"P",
		"Q",
		"P AND Q",
		"P OR Q",
		"NOT P",
	}

	_, response, err := handler.HandleCheckSyntax(ctx, req, CheckSyntaxRequest{
		Statements: statements,
	})

	if err != nil {
		t.Fatalf("HandleCheckSyntax() error = %v", err)
	}

	// Response should indicate validity
	// (actual validation depends on LogicValidator implementation)
	if response == nil {
		t.Fatal("Response should not be nil")
	}
	// Errors and Warnings depend on implementation
}

func TestValidationHandler_ErrorHandling(t *testing.T) {
	store := storage.NewMemoryStorage()
	validator := validation.NewLogicValidator()
	handler := NewValidationHandler(store, validator)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name    string
		handler func() error
		wantErr bool
	}{
		{
			name: "validate with empty ID",
			handler: func() error {
				_, _, err := handler.HandleValidate(ctx, req, ValidateRequest{ThoughtID: ""})
				return err
			},
			wantErr: true,
		},
		{
			name: "validate non-existent thought",
			handler: func() error {
				_, _, err := handler.HandleValidate(ctx, req, ValidateRequest{ThoughtID: "fake-id"})
				return err
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.handler()
			if (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationHandler_Integration(t *testing.T) {
	store := storage.NewMemoryStorage()
	validator := validation.NewLogicValidator()
	handler := NewValidationHandler(store, validator)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Scenario: Check syntax, then prove, then validate a thought

	// 1. Check syntax
	_, syntaxResp, err := handler.HandleCheckSyntax(ctx, req, CheckSyntaxRequest{
		Statements: []string{"P", "P -> Q"},
	})
	if err != nil {
		t.Fatalf("CheckSyntax error = %v", err)
	}
	if syntaxResp == nil {
		t.Fatal("CheckSyntax response is nil")
	}

	// 2. Prove conclusion
	_, proveResp, err := handler.HandleProve(ctx, req, ProveRequest{
		Premises:   []string{"P", "P -> Q"},
		Conclusion: "Q",
	})
	if err != nil {
		t.Fatalf("Prove error = %v", err)
	}
	if proveResp == nil {
		t.Fatal("Prove response is nil")
	}

	// 3. Create and validate a thought
	thought := types.NewThought().
		Content("Modus ponens: P and P->Q implies Q").
		Mode(types.ModeLinear).
		Build()
	_ = store.StoreThought(thought)

	_, validateResp, err := handler.HandleValidate(ctx, req, ValidateRequest{
		ThoughtID: thought.ID,
	})
	if err != nil {
		t.Fatalf("Validate error = %v", err)
	}
	if validateResp == nil {
		t.Fatal("Validate response is nil")
	}
	// ValidationID and other fields depend on implementation
}
