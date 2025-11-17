package handlers

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"unified-thinking/internal/analysis"
	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/types"
)

func TestNewTemporalHandler(t *testing.T) {
	perspectiveAnalyzer := &analysis.PerspectiveAnalyzer{}
	temporalReasoner := &reasoning.TemporalReasoner{}

	handler := NewTemporalHandler(perspectiveAnalyzer, temporalReasoner)

	assert.NotNil(t, handler)
	assert.Equal(t, perspectiveAnalyzer, handler.perspectiveAnalyzer)
	assert.Equal(t, temporalReasoner, handler.temporalReasoner)
}

func TestHandleAnalyzePerspectives(t *testing.T) {
	// Create mock analyzer that returns test data
	perspectiveAnalyzer := &analysis.PerspectiveAnalyzer{}
	temporalReasoner := &reasoning.TemporalReasoner{}
	handler := NewTemporalHandler(perspectiveAnalyzer, temporalReasoner)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	testCases := []struct {
		name     string
		input    AnalyzePerspectivesRequest
		wantErr  bool
		validate func(*testing.T, *mcp.CallToolResult, *AnalyzePerspectivesResponse, error)
	}{
		{
			name: "basic perspective analysis",
			input: AnalyzePerspectivesRequest{
				Situation: "Implementing a new authentication system",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *AnalyzePerspectivesResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.NotNil(t, resp.Perspectives)
				// Perspectives should be initialized as empty array, not nil
				if resp.Perspectives == nil {
					resp.Perspectives = []*types.Perspective{}
				}
			},
		},
		{
			name: "perspective analysis with stakeholder hints",
			input: AnalyzePerspectivesRequest{
				Situation:        "Migrating to cloud infrastructure",
				StakeholderHints: []string{"developers", "security team", "finance"},
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *AnalyzePerspectivesResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.NotNil(t, resp.Metadata)
			},
		},
		{
			name: "empty situation",
			input: AnalyzePerspectivesRequest{
				Situation: "",
			},
			wantErr: true, // Empty situation returns an error
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *AnalyzePerspectivesResponse, err error) {
				// Should return an error for empty situation
				require.Error(t, err)
				assert.Contains(t, err.Error(), "situation cannot be empty")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, resp, err := handler.HandleAnalyzePerspectives(ctx, req, tc.input)

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

func TestHandleAnalyzeTemporal(t *testing.T) {
	perspectiveAnalyzer := &analysis.PerspectiveAnalyzer{}
	temporalReasoner := &reasoning.TemporalReasoner{}
	handler := NewTemporalHandler(perspectiveAnalyzer, temporalReasoner)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	testCases := []struct {
		name     string
		input    AnalyzeTemporalRequest
		wantErr  bool
		validate func(*testing.T, *mcp.CallToolResult, *AnalyzeTemporalResponse, error)
	}{
		{
			name: "basic temporal analysis",
			input: AnalyzeTemporalRequest{
				Situation: "Should we refactor the codebase now or after the release?",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *AnalyzeTemporalResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.NotNil(t, resp.Analysis)
			},
		},
		{
			name: "temporal analysis with time horizon",
			input: AnalyzeTemporalRequest{
				Situation:   "Investment in new technology platform",
				TimeHorizon: "years",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *AnalyzeTemporalResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.NotNil(t, resp.Analysis)
				assert.NotNil(t, resp.Metadata)
			},
		},
		{
			name: "short-term analysis",
			input: AnalyzeTemporalRequest{
				Situation:   "Quick bug fix vs proper solution",
				TimeHorizon: "days-weeks",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *AnalyzeTemporalResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, resp, err := handler.HandleAnalyzeTemporal(ctx, req, tc.input)

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

func TestHandleCompareTimeHorizons(t *testing.T) {
	perspectiveAnalyzer := &analysis.PerspectiveAnalyzer{}
	temporalReasoner := &reasoning.TemporalReasoner{}
	handler := NewTemporalHandler(perspectiveAnalyzer, temporalReasoner)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	testCases := []struct {
		name     string
		input    CompareTimeHorizonsRequest
		wantErr  bool
		validate func(*testing.T, *mcp.CallToolResult, *CompareTimeHorizonsResponse, error)
	}{
		{
			name: "basic time horizon comparison",
			input: CompareTimeHorizonsRequest{
				Situation: "Hiring contractors vs full-time employees",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *CompareTimeHorizonsResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.NotNil(t, resp.Analyses)
				// Should have multiple time horizons
				assert.NotEmpty(t, resp.Analyses)
			},
		},
		{
			name: "technology decision comparison",
			input: CompareTimeHorizonsRequest{
				Situation: "Choosing between proven tech stack vs emerging technology",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *CompareTimeHorizonsResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, resp, err := handler.HandleCompareTimeHorizons(ctx, req, tc.input)

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

func TestHandleIdentifyOptimalTiming(t *testing.T) {
	perspectiveAnalyzer := &analysis.PerspectiveAnalyzer{}
	temporalReasoner := &reasoning.TemporalReasoner{}
	handler := NewTemporalHandler(perspectiveAnalyzer, temporalReasoner)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	testCases := []struct {
		name     string
		input    IdentifyOptimalTimingRequest
		wantErr  bool
		validate func(*testing.T, *mcp.CallToolResult, *IdentifyOptimalTimingResponse, error)
	}{
		{
			name: "basic optimal timing",
			input: IdentifyOptimalTimingRequest{
				Situation: "When to launch the new product feature",
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *IdentifyOptimalTimingResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.NotEmpty(t, resp.Recommendation)
			},
		},
		{
			name: "timing with constraints",
			input: IdentifyOptimalTimingRequest{
				Situation: "Database migration timing",
				Constraints: []string{
					"Minimal user disruption",
					"Before Q4 traffic spike",
					"After critical bug fixes",
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *IdentifyOptimalTimingResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.NotEmpty(t, resp.Recommendation)
			},
		},
		{
			name: "empty constraints",
			input: IdentifyOptimalTimingRequest{
				Situation:   "System upgrade timing",
				Constraints: []string{},
			},
			wantErr: false,
			validate: func(t *testing.T, result *mcp.CallToolResult, resp *IdentifyOptimalTimingResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, resp, err := handler.HandleIdentifyOptimalTiming(ctx, req, tc.input)

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

// TestTemporalHandlerIntegration tests the integration between temporal and perspective analysis
func TestTemporalHandlerIntegration(t *testing.T) {
	perspectiveAnalyzer := &analysis.PerspectiveAnalyzer{}
	temporalReasoner := &reasoning.TemporalReasoner{}
	handler := NewTemporalHandler(perspectiveAnalyzer, temporalReasoner)

	ctx := context.Background()

	// Test a complex scenario that uses both perspective and temporal analysis
	t.Run("integrated analysis workflow", func(t *testing.T) {
		// First analyze perspectives
		perspReq := &mcp.CallToolRequest{}
		perspInput := AnalyzePerspectivesRequest{
			Situation: "Major architectural refactoring",
			StakeholderHints: []string{"developers", "management", "users"},
		}

		perspResult, perspResp, err := handler.HandleAnalyzePerspectives(ctx, perspReq, perspInput)
		require.NoError(t, err)
		require.NotNil(t, perspResult)
		require.NotNil(t, perspResp)

		// Then analyze temporal aspects
		tempReq := &mcp.CallToolRequest{}
		tempInput := AnalyzeTemporalRequest{
			Situation: "Major architectural refactoring",
			TimeHorizon: "months",
		}

		tempResult, tempResp, err := handler.HandleAnalyzeTemporal(ctx, tempReq, tempInput)
		require.NoError(t, err)
		require.NotNil(t, tempResult)
		require.NotNil(t, tempResp)

		// Verify both analyses succeeded
		assert.Equal(t, "success", perspResp.Status)
		assert.Equal(t, "success", tempResp.Status)
	})
}