package handlers

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"unified-thinking/internal/reasoning"
)

func TestNewCausalHandler(t *testing.T) {
	causalReasoner := reasoning.NewCausalReasoner()
	handler := NewCausalHandler(causalReasoner)

	assert.NotNil(t, handler)
	assert.Equal(t, causalReasoner, handler.causalReasoner)
}

func TestHandleBuildCausalGraph(t *testing.T) {
	causalReasoner := reasoning.NewCausalReasoner()
	handler := NewCausalHandler(causalReasoner)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	testCases := []struct {
		name     string
		input    BuildCausalGraphRequest
		wantErr  bool
		validate func(*testing.T, *mcp.CallToolResult, *BuildCausalGraphResponse, error)
	}{
		{
			name: "basic causal graph",
			input: BuildCausalGraphRequest{
				Description: "Marketing campaign effectiveness",
				Observations: []string{
					"Increased ad spend leads to more impressions",
					"More impressions lead to more clicks",
					"More clicks lead to more conversions",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *BuildCausalGraphResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.NotNil(t, resp.Graph)
				// Graph should have ID
				if resp.Graph != nil {
					assert.NotEmpty(t, resp.Graph.ID)
				}
			},
		},
		{
			name: "complex causal relationships",
			input: BuildCausalGraphRequest{
				Description: "Software development process",
				Observations: []string{
					"Code complexity affects bug count",
					"Bug count affects development time",
					"Development time affects release date",
					"Team size affects development time",
					"Code reviews reduce bug count",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *BuildCausalGraphResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
			},
		},
		{
			name: "empty observations",
			input: BuildCausalGraphRequest{
				Description:  "Empty model",
				Observations: []string{},
			},
			wantErr: true,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *BuildCausalGraphResponse, err error) {
				// Should return an error for empty observations
				require.Error(t, err)
				assert.Contains(t, err.Error(), "at least one observation is required")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, resp, err := handler.HandleBuildCausalGraph(ctx, req, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.validate != nil {
				tc.validate(t, result, resp, err)
			}
		})
	}
}

func TestHandleSimulateIntervention(t *testing.T) {
	causalReasoner := reasoning.NewCausalReasoner()
	handler := NewCausalHandler(causalReasoner)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// First build a graph to test intervention on
	buildInput := BuildCausalGraphRequest{
		Description: "Sales process",
		Observations: []string{
			"Marketing increases awareness",
			"Awareness drives sales",
		},
	}
	_, buildResp, err := handler.HandleBuildCausalGraph(ctx, req, buildInput)
	require.NoError(t, err)
	require.NotNil(t, buildResp)
	require.NotNil(t, buildResp.Graph)

	graphID := buildResp.Graph.ID

	testCases := []struct {
		name     string
		input    SimulateInterventionRequest
		wantErr  bool
		validate func(*testing.T, *mcp.CallToolResult, *SimulateInterventionResponse, error)
	}{
		{
			name: "increase intervention",
			input: SimulateInterventionRequest{
				GraphID:          graphID,
				VariableID:       "marketing",
				InterventionType: "increase",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *SimulateInterventionResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.NotNil(t, resp.Intervention)
			},
		},
		{
			name: "decrease intervention",
			input: SimulateInterventionRequest{
				GraphID:          graphID,
				VariableID:       "awareness",
				InterventionType: "decrease",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *SimulateInterventionResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
			},
		},
		{
			name: "non-existent graph",
			input: SimulateInterventionRequest{
				GraphID:          "non-existent-graph-id",
				VariableID:       "marketing",
				InterventionType: "increase",
			},
			wantErr: true,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *SimulateInterventionResponse, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, resp, err := handler.HandleSimulateIntervention(ctx, req, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.validate != nil {
				tc.validate(t, result, resp, err)
			}
		})
	}
}

func TestHandleGenerateCounterfactual(t *testing.T) {
	causalReasoner := reasoning.NewCausalReasoner()
	handler := NewCausalHandler(causalReasoner)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Build a graph for testing
	buildInput := BuildCausalGraphRequest{
		Description: "Project timeline",
		Observations: []string{
			"Team size affects development speed",
			"Development speed affects delivery date",
		},
	}
	_, buildResp, err := handler.HandleBuildCausalGraph(ctx, req, buildInput)
	require.NoError(t, err)
	require.NotNil(t, buildResp)
	require.NotNil(t, buildResp.Graph)

	graphID := buildResp.Graph.ID

	testCases := []struct {
		name     string
		input    GenerateCounterfactualRequest
		wantErr  bool
		validate func(*testing.T, *mcp.CallToolResult, *GenerateCounterfactualResponse, error)
	}{
		{
			name: "what if scenario",
			input: GenerateCounterfactualRequest{
				GraphID:  graphID,
				Scenario: "What if we doubled the team size?",
				Changes: map[string]string{
					"team_size": "doubled",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *GenerateCounterfactualResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.NotNil(t, resp.Counterfactual)
			},
		},
		{
			name: "multiple changes",
			input: GenerateCounterfactualRequest{
				GraphID:  graphID,
				Scenario: "What if we changed both team and timeline?",
				Changes: map[string]string{
					"team_size":         "increased",
					"development_speed": "faster",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *GenerateCounterfactualResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, resp, err := handler.HandleGenerateCounterfactual(ctx, req, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.validate != nil {
				tc.validate(t, result, resp, err)
			}
		})
	}
}

func TestHandleAnalyzeCorrelationVsCausation(t *testing.T) {
	causalReasoner := reasoning.NewCausalReasoner()
	handler := NewCausalHandler(causalReasoner)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	testCases := []struct {
		name     string
		input    AnalyzeCorrelationVsCausationRequest
		wantErr  bool
		validate func(*testing.T, *mcp.CallToolResult, *AnalyzeCorrelationVsCausationResponse, error)
	}{
		{
			name: "ice cream sales correlation",
			input: AnalyzeCorrelationVsCausationRequest{
				Observation: "Ice cream sales and drowning incidents both increase in summer",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *AnalyzeCorrelationVsCausationResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.NotEmpty(t, resp.Analysis)
			},
		},
		{
			name: "direct causation example",
			input: AnalyzeCorrelationVsCausationRequest{
				Observation: "Smoking increases lung cancer risk",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *AnalyzeCorrelationVsCausationResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, resp, err := handler.HandleAnalyzeCorrelationVsCausation(ctx, req, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.validate != nil {
				tc.validate(t, result, resp, err)
			}
		})
	}
}

func TestHandleGetCausalGraph(t *testing.T) {
	causalReasoner := reasoning.NewCausalReasoner()
	handler := NewCausalHandler(causalReasoner)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// First build a graph
	buildInput := BuildCausalGraphRequest{
		Description: "Test graph",
		Observations: []string{
			"A causes B",
		},
	}
	_, buildResp, err := handler.HandleBuildCausalGraph(ctx, req, buildInput)
	require.NoError(t, err)
	require.NotNil(t, buildResp)
	require.NotNil(t, buildResp.Graph)

	graphID := buildResp.Graph.ID

	testCases := []struct {
		name     string
		input    GetCausalGraphRequest
		wantErr  bool
		validate func(*testing.T, *mcp.CallToolResult, *GetCausalGraphResponse, error)
	}{
		{
			name: "retrieve existing graph",
			input: GetCausalGraphRequest{
				GraphID: graphID,
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *GetCausalGraphResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.NotNil(t, resp.Graph)
				assert.Equal(t, graphID, resp.Graph.ID)
			},
		},
		{
			name: "non-existent graph",
			input: GetCausalGraphRequest{
				GraphID: "non-existent-id",
			},
			wantErr: true,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *GetCausalGraphResponse, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not found")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, resp, err := handler.HandleGetCausalGraph(ctx, req, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tc.validate != nil {
				tc.validate(t, result, resp, err)
			}
		})
	}
}
