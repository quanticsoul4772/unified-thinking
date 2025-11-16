// Package server provides the tool executor implementation for workflow orchestration.
package server

import (
	"context"
	"fmt"

	"unified-thinking/internal/modes"
	"unified-thinking/internal/orchestration"
	"unified-thinking/internal/server/handlers"
)

// NewServerToolExecutor creates a new executor with access to server handlers
func NewServerToolExecutor(srv *UnifiedServer) orchestration.ToolExecutor {
	return &serverToolExecutor{
		server: srv,
	}
}

// serverToolExecutor implements orchestration.ToolExecutor using UnifiedServer handlers
type serverToolExecutor struct {
	server *UnifiedServer
}

// ExecuteTool executes the specified tool with the given input
func (e *serverToolExecutor) ExecuteTool(ctx context.Context, toolName string, input map[string]interface{}) (interface{}, error) {
	// Convert generic input to specific request types and call appropriate handler
	switch toolName {
	case "think":
		thoughtInput := modes.ThoughtInput{
			Content:        getStringField(input, "content"),
			Type:           getStringField(input, "type"),
			Confidence:     getFloatField(input, "confidence"),
			BranchID:       getStringField(input, "branch_id"),
			ParentID:       getStringField(input, "parent_id"),
			KeyPoints:      getStringSliceField(input, "key_points"),
			ForceRebellion: getBoolField(input, "force_rebellion"),
		}
		mode := getStringField(input, "mode")
		return e.server.ProcessThought(ctx, mode, thoughtInput)

	case "build-causal-graph":
		// Support both old field names (problem, context) and new ones (description, observations)
		description := getStringField(input, "description")
		if description == "" {
			description = getStringField(input, "problem")
		}
		observations := getStringSliceField(input, "observations")
		if len(observations) == 0 {
			// Fallback to context as a single observation
			context := getStringField(input, "context")
			if context != "" {
				observations = []string{context}
			}
		}
		return e.server.BuildCausalGraph(ctx, description, observations)

	case "probabilistic-reasoning":
		req := handlers.ProbabilisticReasoningRequest{
			Operation:    getStringField(input, "operation"),
			Statement:    getStringField(input, "statement"),
			PriorProb:    getFloatField(input, "prior_prob"),
			BeliefID:     getStringField(input, "belief_id"),
			EvidenceID:   getStringField(input, "evidence_id"),
			Likelihood:   getFloatField(input, "likelihood"),
			EvidenceProb: getFloatField(input, "evidence_prob"),
			BeliefIDs:    getStringSliceField(input, "belief_ids"),
			CombineOp:    getStringField(input, "combine_op"),
		}
		return e.server.ProbabilisticReasoning(ctx, req)

	case "assess-evidence":
		// Create a simple response for now
		claim := getStringField(input, "claim")
		evidence := getStringSliceField(input, "evidence")
		return map[string]interface{}{
			"claim":    claim,
			"evidence": evidence,
			"strength": 0.7,
			"quality":  "medium",
		}, nil

	case "detect-biases":
		// Create a simple response for now
		content := getStringField(input, "content")
		return map[string]interface{}{
			"content": content,
			"biases":  []string{},
		}, nil

	case "synthesize-insights":
		// Create a simple response for now
		inputs := getStringSliceField(input, "inputs")
		return map[string]interface{}{
			"inputs":   inputs,
			"insights": []string{"Combined analysis of inputs"},
		}, nil

	case "decompose-problem":
		req := handlers.DecomposeProblemRequest{
			Problem: getStringField(input, "problem"),
		}
		return e.server.DecomposeProblem(ctx, req)

	case "sensitivity-analysis":
		req := handlers.SensitivityAnalysisRequest{
			TargetClaim:    getStringField(input, "target_claim"),
			Assumptions:    getStringSliceField(input, "assumptions"),
			BaseConfidence: getFloatField(input, "base_confidence"),
		}
		return e.server.SensitivityAnalysis(ctx, req)

	case "analyze-perspectives":
		req := handlers.AnalyzePerspectivesRequest{
			Situation:        getStringField(input, "situation"),
			StakeholderHints: getStringSliceField(input, "stakeholder_hints"),
		}
		return e.server.AnalyzePerspectives(ctx, req)

	case "make-decision":
		// Create a simple response for now
		situation := getStringField(input, "situation")
		options := getStringSliceField(input, "options")
		return map[string]interface{}{
			"situation": situation,
			"options":   options,
			"selected":  options[0],
			"score":     0.8,
		}, nil

	default:
		return nil, fmt.Errorf("tool %s not supported in orchestrator", toolName)
	}
}

// Helper functions to extract fields from map[string]interface{}

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloatField(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case float32:
			return float64(val)
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0.0
}

func getBoolField(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func getStringSliceField(m map[string]interface{}, key string) []string {
	if v, ok := m[key]; ok {
		if slice, ok := v.([]interface{}); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
		if slice, ok := v.([]string); ok {
			return slice
		}
	}
	return nil
}

// getCriteria is reserved for future use in decision criterion extraction
func _getCriteria(m map[string]interface{}) []map[string]interface{} {
	if v, ok := m["criteria"]; ok {
		if criteria, ok := v.([]interface{}); ok {
			result := make([]map[string]interface{}, 0, len(criteria))
			for _, c := range criteria {
				if cMap, ok := c.(map[string]interface{}); ok {
					result = append(result, cMap)
				}
			}
			return result
		}
		if criteria, ok := v.([]map[string]interface{}); ok {
			return criteria
		}
	}
	return nil
}
