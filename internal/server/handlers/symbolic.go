package handlers

import (
	"fmt"
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/validation"
)

// SymbolicHandler handles symbolic reasoning operations
type SymbolicHandler struct {
	reasoner *validation.SymbolicReasoner
	storage  storage.Storage
}

// NewSymbolicHandler creates a new symbolic handler
func NewSymbolicHandler(reasoner *validation.SymbolicReasoner, store storage.Storage) *SymbolicHandler {
	return &SymbolicHandler{
		reasoner: reasoner,
		storage:  store,
	}
}

// ProveTheoremRequest represents a theorem proving request
type ProveTheoremRequest struct {
	Name       string   `json:"name"`
	Premises   []string `json:"premises"`
	Conclusion string   `json:"conclusion"`
}

// ProveTheoremResponse represents the response
type ProveTheoremResponse struct {
	Name       string       `json:"name"`
	Status     string       `json:"status"`
	IsValid    bool         `json:"is_valid"`
	Confidence float64      `json:"confidence"`
	Proof      *ProofOutput `json:"proof,omitempty"`
}

// ProofOutput represents a proof
type ProofOutput struct {
	Steps       []*ProofStepOutput `json:"steps"`
	Method      string             `json:"method"`
	Explanation string             `json:"explanation"`
}

// ProofStepOutput represents a proof step
type ProofStepOutput struct {
	StepNumber    int      `json:"step_number"`
	Statement     string   `json:"statement"`
	Justification string   `json:"justification"`
	Rule          string   `json:"rule"`
	Dependencies  []int    `json:"dependencies"`
}

// HandleProveTheorem attempts to prove a theorem
func (h *SymbolicHandler) HandleProveTheorem(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	var req ProveTheoremRequest
	if err := unmarshalParams(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request: " + err.Error())
	}

	if req.Conclusion == "" {
		return nil, fmt.Errorf("conclusion is required")
	}
	if len(req.Premises) == 0 {
		return nil, fmt.Errorf("premises are required")
	}

	// Create theorem
	theorem := &validation.SymbolicTheorem{
		Name:       req.Name,
		Premises:   req.Premises,
		Conclusion: req.Conclusion,
		Status:     validation.StatusUnproven,
	}

	// Prove theorem
	proof, err := h.reasoner.ProveTheorem(theorem)
	if err != nil {
		return nil, fmt.Errorf("theorem proving failed: " + err.Error())
	}

	// Build response
	resp := &ProveTheoremResponse{
		Name:       theorem.Name,
		Status:     string(theorem.Status),
		IsValid:    proof.IsValid,
		Confidence: proof.Confidence,
	}

	if proof != nil {
		steps := make([]*ProofStepOutput, len(proof.Steps))
		for i, step := range proof.Steps {
			steps[i] = &ProofStepOutput{
				StepNumber:    step.StepNumber,
				Statement:     step.Statement,
				Justification: step.Justification,
				Rule:          step.Rule,
				Dependencies:  step.Dependencies,
			}
		}

		resp.Proof = &ProofOutput{
			Steps:       steps,
			Method:      proof.Method,
			Explanation: proof.Explanation,
		}
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}

// CheckConstraintsRequest represents a constraint checking request
type CheckConstraintsRequest struct {
	Symbols     []*SymbolInput     `json:"symbols"`
	Constraints []*ConstraintInput `json:"constraints"`
}

// SymbolInput represents a symbol definition
type SymbolInput struct {
	Name   string `json:"name"`
	Type   string `json:"type"`   // "variable", "constant", "function"
	Domain string `json:"domain"` // "integer", "boolean", "real"
}

// ConstraintInput represents a constraint
type ConstraintInput struct {
	Type       string   `json:"type"`       // "equality", "inequality", etc.
	Expression string   `json:"expression"` // The constraint expression
	Symbols    []string `json:"symbols"`    // Symbols involved
}

// CheckConstraintsResponse represents the response
type CheckConstraintsResponse struct {
	IsConsistent bool              `json:"is_consistent"`
	Conflicts    []*ConflictOutput `json:"conflicts,omitempty"`
	Explanation  string            `json:"explanation"`
}

// ConflictOutput represents a constraint conflict
type ConflictOutput struct {
	Constraint1  string `json:"constraint1"`
	Constraint2  string `json:"constraint2"`
	ConflictType string `json:"conflict_type"`
	Explanation  string `json:"explanation"`
}

// HandleCheckConstraints checks constraint consistency
func (h *SymbolicHandler) HandleCheckConstraints(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	var req CheckConstraintsRequest
	if err := unmarshalParams(params, &req); err != nil {
		return nil, fmt.Errorf("invalid request: " + err.Error())
	}

	if len(req.Symbols) == 0 {
		return nil, fmt.Errorf("symbols are required")
	}
	if len(req.Constraints) == 0 {
		return nil, fmt.Errorf("constraints are required")
	}

	// Add symbols
	for _, sym := range req.Symbols {
		symbolType := validation.SymbolVariable
		if sym.Type == "constant" {
			symbolType = validation.SymbolConstant
		} else if sym.Type == "function" {
			symbolType = validation.SymbolFunction
		}

		h.reasoner.AddSymbol(sym.Name, symbolType, sym.Domain)
	}

	// Add constraints and collect IDs
	constraintIDs := make([]string, 0, len(req.Constraints))
	for _, cons := range req.Constraints {
		constraintType := validation.ConstraintType(cons.Type)

		constraint, err := h.reasoner.AddConstraint(constraintType, cons.Expression, cons.Symbols)
		if err != nil {
			return nil, fmt.Errorf("failed to add constraint: " + err.Error())
		}
		constraintIDs = append(constraintIDs, constraint.ID)
	}

	// Check consistency
	result, err := h.reasoner.CheckConstraintConsistency(constraintIDs)
	if err != nil {
		return nil, fmt.Errorf("consistency check failed: " + err.Error())
	}

	// Build response
	resp := &CheckConstraintsResponse{
		IsConsistent: result.IsConsistent,
		Explanation:  result.Explanation,
	}

	if len(result.Conflicts) > 0 {
		conflicts := make([]*ConflictOutput, len(result.Conflicts))
		for i, conf := range result.Conflicts {
			conflicts[i] = &ConflictOutput{
				Constraint1:  conf.Constraint1,
				Constraint2:  conf.Constraint2,
				ConflictType: conf.ConflictType,
				Explanation:  conf.Explanation,
			}
		}
		resp.Conflicts = conflicts
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}
