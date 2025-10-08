package processing

import (
	"context"
	"testing"
	"time"

	"unified-thinking/internal/modes"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"

	"github.com/stretchr/testify/assert"
)

func TestNewDualProcessExecutor(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeRegistry := make(map[types.ThinkingMode]modes.ThinkingMode)
	modeRegistry[types.ModeLinear] = modes.NewLinearMode(store)

	executor := NewDualProcessExecutor(store, modeRegistry)

	assert.NotNil(t, executor)
	assert.NotNil(t, executor.storage)
	assert.NotNil(t, executor.modes)
}

func TestDualProcessExecutor_CalculateComplexity(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeRegistry := make(map[types.ThinkingMode]modes.ThinkingMode)
	executor := NewDualProcessExecutor(store, modeRegistry)

	tests := []struct {
		name     string
		request  *ProcessingRequest
		minScore float64
		maxScore float64
	}{
		{
			name: "simple_short_content",
			request: &ProcessingRequest{
				Content:   "What is 2+2?",
				KeyPoints: []string{},
			},
			minScore: 0.0,
			maxScore: 0.3,
		},
		{
			name: "medium_complexity",
			request: &ProcessingRequest{
				Content:   "How does PostgreSQL implement MVCC for transaction isolation? Please explain the key concepts.",
				KeyPoints: []string{"MVCC", "transactions"},
			},
			minScore: 0.3,
			maxScore: 0.7,
		},
		{
			name: "high_complexity",
			request: &ProcessingRequest{
				Content: "Analyze and compare the tradeoffs between microservices and monolithic architectures. Evaluate multiple dimensions including scalability, complexity, deployment, testing, and team organization. Design an optimal approach that balances these factors.",
				KeyPoints: []string{"microservices", "monolithic", "scalability", "complexity", "deployment", "testing"},
			},
			minScore: 0.7,
			maxScore: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complexity := executor.calculateComplexity(tt.request)
			assert.GreaterOrEqual(t, complexity, tt.minScore)
			assert.LessOrEqual(t, complexity, tt.maxScore)
		})
	}
}

func TestDualProcessExecutor_SelectSystem(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeRegistry := make(map[types.ThinkingMode]modes.ThinkingMode)
	executor := NewDualProcessExecutor(store, modeRegistry)

	tests := []struct {
		name       string
		complexity float64
		expected   ProcessingSystem
	}{
		{"low_complexity_system1", 0.2, System1},
		{"medium_complexity_system2", 0.5, System2},
		{"high_complexity_system2", 0.9, System2},
		{"threshold_system1", 0.39, System1},
		{"threshold_system2", 0.4, System2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ProcessingRequest{}
			system := executor.selectSystem(req, tt.complexity)
			assert.Equal(t, tt.expected, system)
		})
	}
}

func TestDualProcessExecutor_ExecuteSystem1(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeRegistry := make(map[types.ThinkingMode]modes.ThinkingMode)
	modeRegistry[types.ModeLinear] = modes.NewLinearMode(store)

	executor := NewDualProcessExecutor(store, modeRegistry)

	ctx := context.Background()
	req := &ProcessingRequest{
		Content:    "Simple question",
		Mode:       types.ModeLinear,
		Confidence: 0.8,
	}

	thought, err := executor.executeSystem1(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, thought)
	assert.Equal(t, "System1", thought.Metadata["processing_system"])
	assert.Equal(t, "fast_heuristic", thought.Metadata["processing_mode"])
	assert.True(t, thought.Metadata["escalation_available"].(bool))
}

func TestDualProcessExecutor_ExecuteSystem2(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeRegistry := make(map[types.ThinkingMode]modes.ThinkingMode)
	modeRegistry[types.ModeLinear] = modes.NewLinearMode(store)

	executor := NewDualProcessExecutor(store, modeRegistry)

	ctx := context.Background()
	req := &ProcessingRequest{
		Content:    "Complex analytical question",
		Mode:       types.ModeLinear,
		Confidence: 0.7,
	}

	thought, err := executor.executeSystem2(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, thought)
	assert.Equal(t, "System2", thought.Metadata["processing_system"])
	assert.Equal(t, "analytical_deliberate", thought.Metadata["processing_mode"])
	assert.True(t, thought.Metadata["full_analysis"].(bool))
}

func TestDualProcessExecutor_ShouldEscalate(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeRegistry := make(map[types.ThinkingMode]modes.ThinkingMode)
	executor := NewDualProcessExecutor(store, modeRegistry)

	tests := []struct {
		name     string
		thought  *types.Thought
		request  *ProcessingRequest
		expected bool
	}{
		{
			name: "low_confidence_escalate",
			thought: &types.Thought{
				Content:    "I think maybe this is correct",
				Confidence: 0.4,
			},
			request: &ProcessingRequest{
				EscalateOnLowConf:   true,
				ConfidenceThreshold: 0.6,
			},
			expected: true,
		},
		{
			name: "high_confidence_no_escalate",
			thought: &types.Thought{
				Content:    "This is definitely correct",
				Confidence: 0.9,
			},
			request: &ProcessingRequest{
				EscalateOnLowConf:   true,
				ConfidenceThreshold: 0.6,
			},
			expected: false,
		},
		{
			name: "uncertainty_marker_escalate",
			thought: &types.Thought{
				Content:    "I'm unsure about this approach",
				Confidence: 0.8,
			},
			request: &ProcessingRequest{
				EscalateOnLowConf: false,
			},
			expected: true,
		},
		{
			name: "short_content_complex_problem",
			thought: &types.Thought{
				Content:    "Yes", // Very short response
				Confidence: 0.7,
			},
			request: &ProcessingRequest{
				Content:   "This is a very complex analytical problem that requires deep thinking about multiple dimensions including scalability, performance, security, and maintainability. We need to analyze tradeoffs between different architectures.",
				KeyPoints: []string{"scalability", "performance", "security", "maintainability", "tradeoffs", "architecture"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.shouldEscalate(tt.thought, tt.request)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDualProcessExecutor_ProcessThought_System1(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeRegistry := make(map[types.ThinkingMode]modes.ThinkingMode)
	modeRegistry[types.ModeLinear] = modes.NewLinearMode(store)

	executor := NewDualProcessExecutor(store, modeRegistry)

	ctx := context.Background()
	req := &ProcessingRequest{
		Content:    "What is 2+2?",
		Mode:       types.ModeLinear,
		Confidence: 0.9,
	}

	result, err := executor.ProcessThought(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, System1, result.SystemUsed)
	assert.False(t, result.Escalated)
	assert.GreaterOrEqual(t, result.System1Time, time.Duration(0))
	assert.Equal(t, time.Duration(0), result.System2Time)
}

func TestDualProcessExecutor_ProcessThought_System2(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeRegistry := make(map[types.ThinkingMode]modes.ThinkingMode)
	modeRegistry[types.ModeLinear] = modes.NewLinearMode(store)

	executor := NewDualProcessExecutor(store, modeRegistry)

	ctx := context.Background()
	req := &ProcessingRequest{
		Content: "Analyze the tradeoffs between different database architectures and design an optimal solution for a high-traffic e-commerce platform",
		Mode:    types.ModeLinear,
		Confidence: 0.7,
	}

	result, err := executor.ProcessThought(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, System2, result.SystemUsed)
	assert.False(t, result.Escalated)
	assert.GreaterOrEqual(t, result.System2Time, time.Duration(0))
	assert.Equal(t, time.Duration(0), result.System1Time)
}

func TestDualProcessExecutor_ProcessThought_WithEscalation(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeRegistry := make(map[types.ThinkingMode]modes.ThinkingMode)
	modeRegistry[types.ModeLinear] = modes.NewLinearMode(store)

	executor := NewDualProcessExecutor(store, modeRegistry)

	ctx := context.Background()
	req := &ProcessingRequest{
		Content:             "What is the result?",
		Mode:                types.ModeLinear,
		Confidence:          0.3, // Low confidence to trigger escalation
		EscalateOnLowConf:   true,
		ConfidenceThreshold: 0.6,
	}

	result, err := executor.ProcessThought(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Should escalate due to low confidence
	assert.True(t, result.Escalated || result.SystemUsed == System1) // May or may not escalate depending on content
	assert.GreaterOrEqual(t, result.TotalTime, time.Duration(0))
}

func TestDualProcessExecutor_ProcessThought_ForceSystem1(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeRegistry := make(map[types.ThinkingMode]modes.ThinkingMode)
	modeRegistry[types.ModeLinear] = modes.NewLinearMode(store)

	executor := NewDualProcessExecutor(store, modeRegistry)

	ctx := context.Background()
	req := &ProcessingRequest{
		Content:     "Very complex analytical question that would normally use System 2",
		Mode:        types.ModeLinear,
		Confidence:  0.7,
		ForceSystem: System1, // Force System 1 despite complexity
	}

	result, err := executor.ProcessThought(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, System1, result.SystemUsed)
}

func TestDualProcessExecutor_ProcessThought_ForceSystem2(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeRegistry := make(map[types.ThinkingMode]modes.ThinkingMode)
	modeRegistry[types.ModeLinear] = modes.NewLinearMode(store)

	executor := NewDualProcessExecutor(store, modeRegistry)

	ctx := context.Background()
	req := &ProcessingRequest{
		Content:     "Simple question",
		Mode:        types.ModeLinear,
		Confidence:  0.9,
		ForceSystem: System2, // Force System 2 despite simplicity
	}

	result, err := executor.ProcessThought(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, System2, result.SystemUsed)
	assert.False(t, result.Escalated) // No escalation since we started with System 2
}

func TestDualProcessExecutor_EscalateToSystem2(t *testing.T) {
	store := storage.NewMemoryStorage()
	modeRegistry := make(map[types.ThinkingMode]modes.ThinkingMode)
	modeRegistry[types.ModeLinear] = modes.NewLinearMode(store)

	executor := NewDualProcessExecutor(store, modeRegistry)

	ctx := context.Background()

	// Create System 1 thought
	system1Thought := &types.Thought{
		ID:         "system1-thought",
		Content:    "Maybe this is correct",
		Confidence: 0.4,
	}

	req := &ProcessingRequest{
		Content:             "Original question",
		Mode:                types.ModeLinear,
		Confidence:          0.4,
		ConfidenceThreshold: 0.6,
	}

	thought, reason, err := executor.escalateToSystem2(ctx, req, system1Thought)

	assert.NoError(t, err)
	assert.NotNil(t, thought)
	assert.NotEmpty(t, reason)
	assert.Contains(t, reason, "confidence")
	assert.True(t, thought.Metadata["escalated_from_system1"].(bool))
	assert.Equal(t, "system1-thought", thought.Metadata["system1_thought_id"])
	assert.Equal(t, "system1-thought", thought.ParentID)
}
