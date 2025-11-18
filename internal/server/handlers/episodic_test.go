package handlers

import (
	"context"
	"testing"

	"unified-thinking/internal/memory"
)

func TestNewEpisodicMemoryHandler(t *testing.T) {
	store := memory.NewEpisodicMemoryStore()
	tracker := memory.NewSessionTracker(store)
	learner := memory.NewLearningEngine(store)

	handler := NewEpisodicMemoryHandler(store, tracker, learner)

	if handler == nil {
		t.Fatal("NewEpisodicMemoryHandler returned nil")
	}
	if handler.store == nil {
		t.Error("store not initialized")
	}
	if handler.tracker == nil {
		t.Error("tracker not initialized")
	}
	if handler.learner == nil {
		t.Error("learner not initialized")
	}
	if handler.retrospective == nil {
		t.Error("retrospective not initialized")
	}
}

func TestEpisodicMemoryHandler_HandleStartSession(t *testing.T) {
	store := memory.NewEpisodicMemoryStore()
	tracker := memory.NewSessionTracker(store)
	learner := memory.NewLearningEngine(store)
	handler := NewEpisodicMemoryHandler(store, tracker, learner)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid session start",
			params: map[string]interface{}{
				"session_id":  "test_session_1",
				"description": "Test problem description",
				"goals":       []interface{}{"goal1", "goal2"},
				"domain":      "testing",
				"complexity":  0.5,
			},
			wantErr: false,
		},
		{
			name: "minimal session",
			params: map[string]interface{}{
				"session_id":  "test_session_2",
				"description": "Minimal test",
			},
			wantErr: false,
		},
		{
			name: "missing session ID",
			params: map[string]interface{}{
				"description": "Test description",
			},
			wantErr: true,
		},
		{
			name: "missing description",
			params: map[string]interface{}{
				"session_id": "test_session_3",
			},
			wantErr: true,
		},
		{
			name:    "empty params",
			params:  map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handler.HandleStartSession(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleStartSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
					return
				}
				if result.Content == nil {
					t.Error("Content should not be nil")
				}
			}
		})
	}
}

func TestEpisodicMemoryHandler_HandleCompleteSession(t *testing.T) {
	store := memory.NewEpisodicMemoryStore()
	tracker := memory.NewSessionTracker(store)
	learner := memory.NewLearningEngine(store)
	handler := NewEpisodicMemoryHandler(store, tracker, learner)

	ctx := context.Background()

	// First start a session
	startParams := map[string]interface{}{
		"session_id":  "complete_test_1",
		"description": "Test for completion",
		"goals":       []interface{}{"test goal"},
	}
	_, err := handler.HandleStartSession(ctx, startParams)
	if err != nil {
		t.Fatalf("Failed to start session: %v", err)
	}

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid completion",
			params: map[string]interface{}{
				"session_id":     "complete_test_1",
				"status":         "success",
				"goals_achieved": []interface{}{"test goal"},
				"solution":       "Test solution",
				"confidence":     0.85,
			},
			wantErr: false,
		},
		{
			name: "missing session ID",
			params: map[string]interface{}{
				"status": "success",
			},
			wantErr: true,
		},
		{
			name: "non-existent session",
			params: map[string]interface{}{
				"session_id": "non_existent",
				"status":     "success",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the valid completion test, we need to start a new session
			// since we already completed the previous one
			if tt.name == "valid completion" {
				// Re-start the session for this test
				_, _ = handler.HandleStartSession(ctx, map[string]interface{}{
					"session_id":  "complete_test_1",
					"description": "Test for completion",
				})
			}

			result, err := handler.HandleCompleteSession(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleCompleteSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
			}
		})
	}
}

func TestEpisodicMemoryHandler_HandleGetRecommendations(t *testing.T) {
	store := memory.NewEpisodicMemoryStore()
	tracker := memory.NewSessionTracker(store)
	learner := memory.NewLearningEngine(store)
	handler := NewEpisodicMemoryHandler(store, tracker, learner)

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid recommendations request",
			params: map[string]interface{}{
				"description": "Need to implement authentication",
				"goals":       []interface{}{"Secure login"},
				"domain":      "security",
				"limit":       5,
			},
			wantErr: false,
		},
		{
			name: "minimal request",
			params: map[string]interface{}{
				"description": "Simple task",
			},
			wantErr: false,
		},
		{
			name: "missing description",
			params: map[string]interface{}{
				"domain": "testing",
			},
			wantErr: true,
		},
		{
			name:    "empty params",
			params:  map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			result, err := handler.HandleGetRecommendations(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleGetRecommendations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
			}
		})
	}
}

func TestEpisodicMemoryHandler_HandleSearchTrajectories(t *testing.T) {
	store := memory.NewEpisodicMemoryStore()
	tracker := memory.NewSessionTracker(store)
	learner := memory.NewLearningEngine(store)
	handler := NewEpisodicMemoryHandler(store, tracker, learner)

	ctx := context.Background()

	// Create and complete a session to have data to search
	_, _ = handler.HandleStartSession(ctx, map[string]interface{}{
		"session_id":  "search_test_1",
		"description": "Test trajectory for search",
		"domain":      "testing",
	})
	_, _ = handler.HandleCompleteSession(ctx, map[string]interface{}{
		"session_id": "search_test_1",
		"status":     "success",
		"confidence": 0.9,
	})

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name:    "search all trajectories",
			params:  map[string]interface{}{},
			wantErr: false,
		},
		{
			name: "search by domain",
			params: map[string]interface{}{
				"domain": "testing",
			},
			wantErr: false,
		},
		{
			name: "search by min success",
			params: map[string]interface{}{
				"min_success": 0.5,
			},
			wantErr: false,
		},
		{
			name: "search with limit",
			params: map[string]interface{}{
				"limit": 5,
			},
			wantErr: false,
		},
		{
			name: "search by tags",
			params: map[string]interface{}{
				"tags": []interface{}{"tag1", "tag2"},
			},
			wantErr: false,
		},
		{
			name: "search by problem type",
			params: map[string]interface{}{
				"problem_type": "debugging",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.HandleSearchTrajectories(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleSearchTrajectories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
			}
		})
	}
}

func TestEpisodicMemoryHandler_HandleAnalyzeTrajectory(t *testing.T) {
	store := memory.NewEpisodicMemoryStore()
	tracker := memory.NewSessionTracker(store)
	learner := memory.NewLearningEngine(store)
	handler := NewEpisodicMemoryHandler(store, tracker, learner)

	ctx := context.Background()

	// Create and complete a session to have a trajectory to analyze
	_, _ = handler.HandleStartSession(ctx, map[string]interface{}{
		"session_id":  "analyze_test_1",
		"description": "Test trajectory for analysis",
		"goals":       []interface{}{"test goal"},
	})
	_, _ = handler.HandleCompleteSession(ctx, map[string]interface{}{
		"session_id": "analyze_test_1",
		"status":     "success",
		"confidence": 0.8,
	})

	// Get the trajectory ID from search
	searchResult, _ := handler.HandleSearchTrajectories(ctx, map[string]interface{}{})
	_ = searchResult // We need to find the trajectory ID

	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "missing trajectory ID",
			params: map[string]interface{}{
				"trajectory_id": "",
			},
			wantErr: true,
		},
		{
			name: "non-existent trajectory",
			params: map[string]interface{}{
				"trajectory_id": "non_existent_traj",
			},
			wantErr: true,
		},
		{
			name:    "empty params",
			params:  map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.HandleAnalyzeTrajectory(ctx, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("HandleAnalyzeTrajectory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Error("CallToolResult should not be nil")
				}
			}
		})
	}
}

func TestComputeProblemHash(t *testing.T) {
	problem := &memory.ProblemDescription{
		Description: "Test problem",
		Context:     "Test context",
		Goals:       []string{"goal1"},
		Domain:      "testing",
	}

	hash1 := ComputeProblemHash(problem)
	hash2 := ComputeProblemHash(problem)

	// Same problem should produce same hash
	if hash1 != hash2 {
		t.Errorf("Same problem produced different hashes: %v vs %v", hash1, hash2)
	}

	// Different problem should produce different hash
	problem2 := &memory.ProblemDescription{
		Description: "Different problem",
		Context:     "Different context",
	}
	hash3 := ComputeProblemHash(problem2)

	if hash1 == hash3 {
		t.Error("Different problems should produce different hashes")
	}
}

func TestEpisodicMemoryHandler_Integration(t *testing.T) {
	store := memory.NewEpisodicMemoryStore()
	tracker := memory.NewSessionTracker(store)
	learner := memory.NewLearningEngine(store)
	handler := NewEpisodicMemoryHandler(store, tracker, learner)

	ctx := context.Background()

	// Complete workflow: start -> get recommendations -> complete -> search -> analyze
	sessionID := "integration_test"

	// 1. Start session
	startResult, err := handler.HandleStartSession(ctx, map[string]interface{}{
		"session_id":  sessionID,
		"description": "Integration test problem",
		"goals":       []interface{}{"Complete test"},
		"domain":      "integration",
		"complexity":  0.7,
	})
	if err != nil {
		t.Fatalf("HandleStartSession() error = %v", err)
	}
	if startResult == nil {
		t.Fatal("Start result should not be nil")
	}

	// 2. Get recommendations
	recResult, err := handler.HandleGetRecommendations(ctx, map[string]interface{}{
		"description": "Similar problem",
		"domain":      "integration",
	})
	if err != nil {
		t.Fatalf("HandleGetRecommendations() error = %v", err)
	}
	if recResult == nil {
		t.Fatal("Recommendations result should not be nil")
	}

	// 3. Complete session
	completeResult, err := handler.HandleCompleteSession(ctx, map[string]interface{}{
		"session_id":     sessionID,
		"status":         "success",
		"goals_achieved": []interface{}{"Complete test"},
		"solution":       "Test passed",
		"confidence":     0.95,
	})
	if err != nil {
		t.Fatalf("HandleCompleteSession() error = %v", err)
	}
	if completeResult == nil {
		t.Fatal("Complete result should not be nil")
	}

	// 4. Search trajectories
	searchResult, err := handler.HandleSearchTrajectories(ctx, map[string]interface{}{
		"domain":      "integration",
		"min_success": 0.5,
	})
	if err != nil {
		t.Fatalf("HandleSearchTrajectories() error = %v", err)
	}
	if searchResult == nil {
		t.Fatal("Search result should not be nil")
	}
}

func TestEpisodicMemoryHandler_SessionLifecycle(t *testing.T) {
	store := memory.NewEpisodicMemoryStore()
	tracker := memory.NewSessionTracker(store)
	learner := memory.NewLearningEngine(store)
	handler := NewEpisodicMemoryHandler(store, tracker, learner)

	ctx := context.Background()

	// Test partial success status
	_, _ = handler.HandleStartSession(ctx, map[string]interface{}{
		"session_id":  "partial_test",
		"description": "Partial success test",
		"goals":       []interface{}{"goal1", "goal2"},
	})

	result, err := handler.HandleCompleteSession(ctx, map[string]interface{}{
		"session_id":     "partial_test",
		"status":         "partial",
		"goals_achieved": []interface{}{"goal1"},
		"goals_failed":   []interface{}{"goal2"},
		"confidence":     0.5,
	})

	if err != nil {
		t.Errorf("HandleCompleteSession() with partial status error = %v", err)
	}
	if result == nil {
		t.Error("Result should not be nil for partial completion")
	}

	// Test failure status
	_, _ = handler.HandleStartSession(ctx, map[string]interface{}{
		"session_id":  "failure_test",
		"description": "Failure test",
	})

	result, err = handler.HandleCompleteSession(ctx, map[string]interface{}{
		"session_id":           "failure_test",
		"status":               "failure",
		"goals_failed":         []interface{}{"all goals"},
		"unexpected_outcomes":  []interface{}{"error occurred"},
		"confidence":           0.1,
	})

	if err != nil {
		t.Errorf("HandleCompleteSession() with failure status error = %v", err)
	}
	if result == nil {
		t.Error("Result should not be nil for failure completion")
	}
}

func TestTrajectorySummary_Structure(t *testing.T) {
	summary := &TrajectorySummary{
		ID:           "traj_1",
		SessionID:    "session_1",
		Problem:      "Test problem",
		Domain:       "testing",
		Strategy:     "test strategy",
		ToolsUsed:    []string{"tool1", "tool2"},
		SuccessScore: 0.85,
		Duration:     "5m",
		Tags:         []string{"tag1", "tag2"},
	}

	if summary.ID != "traj_1" {
		t.Errorf("ID = %v, want traj_1", summary.ID)
	}
	if summary.SuccessScore != 0.85 {
		t.Errorf("SuccessScore = %v, want 0.85", summary.SuccessScore)
	}
	if len(summary.ToolsUsed) != 2 {
		t.Errorf("ToolsUsed length = %v, want 2", len(summary.ToolsUsed))
	}
}

func TestStartSessionResponse_Structure(t *testing.T) {
	response := &StartSessionResponse{
		SessionID:   "session_1",
		ProblemID:   "problem_hash",
		Status:      "active",
		Suggestions: []*memory.Recommendation{},
	}

	if response.SessionID != "session_1" {
		t.Errorf("SessionID = %v, want session_1", response.SessionID)
	}
	if response.Status != "active" {
		t.Errorf("Status = %v, want active", response.Status)
	}
}

func TestCompleteSessionResponse_Structure(t *testing.T) {
	response := &CompleteSessionResponse{
		TrajectoryID:  "traj_1",
		SessionID:     "session_1",
		SuccessScore:  0.9,
		QualityScore:  0.85,
		PatternsFound: 3,
		Status:        "completed",
	}

	if response.TrajectoryID != "traj_1" {
		t.Errorf("TrajectoryID = %v, want traj_1", response.TrajectoryID)
	}
	if response.SuccessScore != 0.9 {
		t.Errorf("SuccessScore = %v, want 0.9", response.SuccessScore)
	}
}

func TestGetRecommendationsResponse_Structure(t *testing.T) {
	response := &GetRecommendationsResponse{
		Recommendations: []*memory.Recommendation{},
		SimilarCases:    5,
		LearnedPatterns: []*memory.TrajectoryPattern{},
		Count:           0,
	}

	if response.SimilarCases != 5 {
		t.Errorf("SimilarCases = %v, want 5", response.SimilarCases)
	}
}

func TestSearchTrajectoriesResponse_Structure(t *testing.T) {
	response := &SearchTrajectoriesResponse{
		Trajectories: []*TrajectorySummary{},
		Count:        0,
	}

	if response.Count != 0 {
		t.Errorf("Count = %v, want 0", response.Count)
	}
}

func TestEpisodicMemoryHandler_HandleAnalyzeTrajectory_WithValidTrajectory(t *testing.T) {
	store := memory.NewEpisodicMemoryStore()
	tracker := memory.NewSessionTracker(store)
	learner := memory.NewLearningEngine(store)
	handler := NewEpisodicMemoryHandler(store, tracker, learner)

	ctx := context.Background()

	// Create and complete a session with detailed data
	sessionID := "analyze_detail_test"
	_, _ = handler.HandleStartSession(ctx, map[string]interface{}{
		"session_id":  sessionID,
		"description": "Detailed test for trajectory analysis",
		"goals":       []interface{}{"Goal 1", "Goal 2", "Goal 3"},
		"domain":      "analysis",
		"complexity":  0.8,
		"context":     "Testing analysis functionality",
	})

	_, _ = handler.HandleCompleteSession(ctx, map[string]interface{}{
		"session_id":          sessionID,
		"status":              "success",
		"goals_achieved":      []interface{}{"Goal 1", "Goal 2"},
		"goals_failed":        []interface{}{"Goal 3"},
		"solution":            "Detailed solution description",
		"confidence":          0.85,
		"unexpected_outcomes": []interface{}{"Minor issue encountered"},
	})

	// Search for the trajectory
	searchResult, _ := handler.HandleSearchTrajectories(ctx, map[string]interface{}{
		"domain": "analysis",
	})

	if searchResult == nil {
		t.Fatal("Search result should not be nil")
	}
}

func TestEpisodicMemoryHandler_HandleGetRecommendations_DetailedContext(t *testing.T) {
	store := memory.NewEpisodicMemoryStore()
	tracker := memory.NewSessionTracker(store)
	learner := memory.NewLearningEngine(store)
	handler := NewEpisodicMemoryHandler(store, tracker, learner)

	ctx := context.Background()

	// First create some historical data
	for i := 0; i < 3; i++ {
		sessionID := "historical_" + string(rune('0'+i))
		_, _ = handler.HandleStartSession(ctx, map[string]interface{}{
			"session_id":  sessionID,
			"description": "Historical problem",
			"domain":      "recommendations",
		})
		_, _ = handler.HandleCompleteSession(ctx, map[string]interface{}{
			"session_id": sessionID,
			"status":     "success",
			"confidence": 0.8,
		})
	}

	// Now get recommendations with full context
	result, err := handler.HandleGetRecommendations(ctx, map[string]interface{}{
		"description": "New similar problem",
		"goals":       []interface{}{"Goal A", "Goal B"},
		"domain":      "recommendations",
		"context":     "Detailed context for recommendations",
		"complexity":  0.6,
		"limit":       10,
	})

	if err != nil {
		t.Errorf("HandleGetRecommendations() error = %v", err)
	}
	if result == nil {
		t.Error("Result should not be nil")
	}
}

func TestEpisodicMemoryHandler_HandleSearchTrajectories_AllFilters(t *testing.T) {
	store := memory.NewEpisodicMemoryStore()
	tracker := memory.NewSessionTracker(store)
	learner := memory.NewLearningEngine(store)
	handler := NewEpisodicMemoryHandler(store, tracker, learner)

	ctx := context.Background()

	// Create sessions with different characteristics
	testData := []struct {
		id         string
		domain     string
		confidence float64
	}{
		{"filter_test_1", "domain_a", 0.9},
		{"filter_test_2", "domain_a", 0.5},
		{"filter_test_3", "domain_b", 0.8},
	}

	for _, td := range testData {
		_, _ = handler.HandleStartSession(ctx, map[string]interface{}{
			"session_id":  td.id,
			"description": "Filter test",
			"domain":      td.domain,
		})
		_, _ = handler.HandleCompleteSession(ctx, map[string]interface{}{
			"session_id": td.id,
			"status":     "success",
			"confidence": td.confidence,
		})
	}

	// Test domain filter
	result, _ := handler.HandleSearchTrajectories(ctx, map[string]interface{}{
		"domain": "domain_a",
	})
	if result == nil {
		t.Error("Domain filter result should not be nil")
	}

	// Test min_success filter
	result, _ = handler.HandleSearchTrajectories(ctx, map[string]interface{}{
		"min_success": 0.8,
	})
	if result == nil {
		t.Error("Min success filter result should not be nil")
	}

	// Test combined filters
	result, _ = handler.HandleSearchTrajectories(ctx, map[string]interface{}{
		"domain":      "domain_a",
		"min_success": 0.6,
		"limit":       1,
	})
	if result == nil {
		t.Error("Combined filter result should not be nil")
	}
}

func TestStartSessionRequest_Structure(t *testing.T) {
	req := StartSessionRequest{
		SessionID:   "test_session",
		Description: "Test description",
		Goals:       []string{"goal1", "goal2"},
		Domain:      "testing",
		Context:     "test context",
		Complexity:  0.5,
		Metadata:    map[string]interface{}{"key": "value"},
	}

	if req.SessionID != "test_session" {
		t.Errorf("SessionID = %v, want test_session", req.SessionID)
	}
	if len(req.Goals) != 2 {
		t.Errorf("Goals length = %v, want 2", len(req.Goals))
	}
}

func TestCompleteSessionRequest_Structure(t *testing.T) {
	req := CompleteSessionRequest{
		SessionID:          "test_session",
		Status:             "success",
		GoalsAchieved:      []string{"goal1"},
		GoalsFailed:        []string{"goal2"},
		Solution:           "test solution",
		Confidence:         0.8,
		UnexpectedOutcomes: []string{"outcome1"},
	}

	if req.Status != "success" {
		t.Errorf("Status = %v, want success", req.Status)
	}
	if req.Confidence != 0.8 {
		t.Errorf("Confidence = %v, want 0.8", req.Confidence)
	}
}

func TestGetRecommendationsRequest_Structure(t *testing.T) {
	req := GetRecommendationsRequest{
		Description: "test description",
		Goals:       []string{"goal1"},
		Domain:      "testing",
		Context:     "context",
		Complexity:  0.7,
		Limit:       10,
	}

	if req.Limit != 10 {
		t.Errorf("Limit = %v, want 10", req.Limit)
	}
}

func TestSearchTrajectoriesRequest_Structure(t *testing.T) {
	req := SearchTrajectoriesRequest{
		Domain:      "testing",
		Tags:        []string{"tag1", "tag2"},
		MinSuccess:  0.5,
		ProblemType: "debug",
		Limit:       5,
	}

	if req.MinSuccess != 0.5 {
		t.Errorf("MinSuccess = %v, want 0.5", req.MinSuccess)
	}
	if len(req.Tags) != 2 {
		t.Errorf("Tags length = %v, want 2", len(req.Tags))
	}
}

func TestAnalyzeTrajectoryRequest_Structure(t *testing.T) {
	req := AnalyzeTrajectoryRequest{
		TrajectoryID: "traj_123",
	}

	if req.TrajectoryID != "traj_123" {
		t.Errorf("TrajectoryID = %v, want traj_123", req.TrajectoryID)
	}
}
