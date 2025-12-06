// Package handlers - Episodic memory handlers for MCP tools
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"unified-thinking/internal/knowledge"
	"unified-thinking/internal/memory"
)

// EpisodicMemoryHandler handles episodic memory operations
type EpisodicMemoryHandler struct {
	store         *memory.EpisodicMemoryStore
	tracker       *memory.SessionTracker
	learner       *memory.LearningEngine
	retrospective *memory.RetrospectiveAnalyzer
	kg            *knowledge.KnowledgeGraph
	extractor     *knowledge.TrajectoryExtractor
}

// NewEpisodicMemoryHandler creates a new episodic memory handler
func NewEpisodicMemoryHandler(store *memory.EpisodicMemoryStore, tracker *memory.SessionTracker, learner *memory.LearningEngine, kg *knowledge.KnowledgeGraph) *EpisodicMemoryHandler {
	var extractor *knowledge.TrajectoryExtractor
	if kg != nil {
		extractor = knowledge.NewTrajectoryExtractor(kg, false) // false = no LLM for now
	}

	return &EpisodicMemoryHandler{
		store:         store,
		tracker:       tracker,
		learner:       learner,
		retrospective: memory.NewRetrospectiveAnalyzer(store),
		kg:            kg,
		extractor:     extractor,
	}
}

// StartSessionRequest starts tracking a reasoning session
type StartSessionRequest struct {
	SessionID   string                 `json:"session_id"`
	Description string                 `json:"description"`
	Goals       []string               `json:"goals,omitempty"`
	Domain      string                 `json:"domain,omitempty"`
	Context     string                 `json:"context,omitempty"`
	Complexity  float64                `json:"complexity,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// StartSessionResponse returns session details
type StartSessionResponse struct {
	SessionID   string                   `json:"session_id"`
	ProblemID   string                   `json:"problem_id"`
	Status      string                   `json:"status"`
	Suggestions []*memory.Recommendation `json:"suggestions,omitempty"`
}

// HandleStartSession starts tracking a new reasoning session
func (h *EpisodicMemoryHandler) HandleStartSession(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	req, err := unmarshalRequest[StartSessionRequest](params)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resp, err := h.startSession(ctx, req)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}

// startSession is the typed internal implementation
func (h *EpisodicMemoryHandler) startSession(ctx context.Context, req StartSessionRequest) (*StartSessionResponse, error) {
	if req.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}

	// Create problem description
	problem := &memory.ProblemDescription{
		Description: req.Description,
		Context:     req.Context,
		Goals:       req.Goals,
		Domain:      req.Domain,
		Complexity:  req.Complexity,
	}

	// Start session
	err := h.tracker.StartSession(ctx, req.SessionID, problem)
	if err != nil {
		return nil, err
	}

	// Get recommendations from similar past trajectories
	similar, _ := h.store.RetrieveSimilarTrajectories(ctx, problem, 5)
	recCtx := &memory.RecommendationContext{
		CurrentProblem:      problem,
		SimilarTrajectories: similar,
	}
	suggestions, _ := h.store.GetRecommendations(ctx, recCtx)

	// Ensure suggestions is never nil (MCP requires arrays, not null)
	if suggestions == nil {
		suggestions = make([]*memory.Recommendation, 0, 3) // Pre-allocate typical size
	}

	return &StartSessionResponse{
		SessionID:   req.SessionID,
		ProblemID:   memory.ComputeProblemHash(problem),
		Status:      "active",
		Suggestions: suggestions,
	}, nil
}

// CompleteSessionRequest completes a reasoning session
type CompleteSessionRequest struct {
	SessionID          string   `json:"session_id"`
	Status             string   `json:"status"` // "success", "partial", "failure"
	GoalsAchieved      []string `json:"goals_achieved,omitempty"`
	GoalsFailed        []string `json:"goals_failed,omitempty"`
	Solution           string   `json:"solution,omitempty"`
	Confidence         float64  `json:"confidence,omitempty"`
	UnexpectedOutcomes []string `json:"unexpected_outcomes,omitempty"`
}

// CompleteSessionResponse returns trajectory details
type CompleteSessionResponse struct {
	TrajectoryID  string  `json:"trajectory_id"`
	SessionID     string  `json:"session_id"`
	SuccessScore  float64 `json:"success_score"`
	QualityScore  float64 `json:"quality_score"`
	PatternsFound int     `json:"patterns_found"`
	Status        string  `json:"status"`
}

// HandleCompleteSession marks a session as complete
func (h *EpisodicMemoryHandler) HandleCompleteSession(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	req, err := unmarshalRequest[CompleteSessionRequest](params)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resp, err := h.completeSession(ctx, req)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}

// completeSession is the typed internal implementation
func (h *EpisodicMemoryHandler) completeSession(ctx context.Context, req CompleteSessionRequest) (*CompleteSessionResponse, error) {
	if req.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	// Build outcome
	outcome := &memory.OutcomeDescription{
		Status:             req.Status,
		GoalsAchieved:      req.GoalsAchieved,
		GoalsFailed:        req.GoalsFailed,
		Solution:           req.Solution,
		Confidence:         req.Confidence,
		UnexpectedOutcomes: req.UnexpectedOutcomes,
	}

	// Complete session
	trajectory, err := h.tracker.CompleteSession(ctx, req.SessionID, outcome)
	if err != nil {
		return nil, err
	}

	// Extract entities to knowledge graph (automatic, always enabled when KG available)
	if h.extractor != nil {
		steps := make([]string, 0, len(trajectory.Steps))
		for _, step := range trajectory.Steps {
			// Extract tool name and mode from step
			stepDesc := fmt.Sprintf("%s (mode: %s)", step.Tool, step.Mode)
			steps = append(steps, stepDesc)
		}

		if err := h.extractor.ExtractFromTrajectory(ctx, trajectory.ID, trajectory.Problem.Description, steps); err != nil {
			log.Printf("[WARN] Failed to extract trajectory to knowledge graph: %v", err)
		} else {
			log.Printf("[DEBUG] Extracted trajectory %s to knowledge graph", trajectory.ID)
		}
	}

	// Trigger pattern learning (async would be better in production)
	patternsFound := 0
	if err := h.learner.LearnPatterns(ctx); err == nil {
		// LearnPatterns returns error only, patterns are stored internally
		// Count would require querying the pattern store, so use 0 for now
		patternsFound = 0
	}

	qualityScore := 0.5
	if trajectory.Quality != nil {
		qualityScore = trajectory.Quality.OverallQuality
	}

	return &CompleteSessionResponse{
		TrajectoryID:  trajectory.ID,
		SessionID:     trajectory.SessionID,
		SuccessScore:  trajectory.SuccessScore,
		QualityScore:  qualityScore,
		PatternsFound: patternsFound,
		Status:        "completed",
	}, nil
}

// GetRecommendationsRequest requests recommendations for current problem
type GetRecommendationsRequest struct {
	Description string   `json:"description"`
	Goals       []string `json:"goals,omitempty"`
	Domain      string   `json:"domain,omitempty"`
	Context     string   `json:"context,omitempty"`
	Complexity  float64  `json:"complexity,omitempty"`
	Limit       int      `json:"limit,omitempty"`
}

// GetRecommendationsResponse returns recommendations
type GetRecommendationsResponse struct {
	Recommendations []*memory.Recommendation    `json:"recommendations"`
	SimilarCases    int                         `json:"similar_cases"`
	LearnedPatterns []*memory.TrajectoryPattern `json:"learned_patterns,omitempty"`
	Count           int                         `json:"count"`
}

// HandleGetRecommendations provides recommendations based on similar past cases
func (h *EpisodicMemoryHandler) HandleGetRecommendations(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	req, err := unmarshalRequest[GetRecommendationsRequest](params)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resp, err := h.getRecommendations(ctx, req)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}

// getRecommendations is the typed internal implementation
func (h *EpisodicMemoryHandler) getRecommendations(ctx context.Context, req GetRecommendationsRequest) (*GetRecommendationsResponse, error) {
	if req.Description == "" {
		return nil, fmt.Errorf("description is required")
	}

	if req.Limit == 0 {
		req.Limit = 5
	}

	// Create problem description
	problem := &memory.ProblemDescription{
		Description: req.Description,
		Context:     req.Context,
		Goals:       req.Goals,
		Domain:      req.Domain,
		Complexity:  req.Complexity,
	}

	// Find similar trajectories
	similar, err := h.store.RetrieveSimilarTrajectories(ctx, problem, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve similar trajectories: %w", err)
	}

	// Get recommendations
	recCtx := &memory.RecommendationContext{
		CurrentProblem:      problem,
		SimilarTrajectories: similar,
	}
	recommendations, err := h.store.GetRecommendations(ctx, recCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// Get learned patterns (don't fail if this errors)
	patterns, _ := h.learner.GetLearnedPatterns(ctx, problem)

	// Initialize empty arrays if nil to ensure valid JSON
	if recommendations == nil {
		recommendations = make([]*memory.Recommendation, 0, 5) // Pre-allocate typical size
	}
	if patterns == nil {
		patterns = make([]*memory.TrajectoryPattern, 0, 3) // Pre-allocate typical size
	}

	return &GetRecommendationsResponse{
		Recommendations: recommendations,
		SimilarCases:    len(similar),
		LearnedPatterns: patterns,
		Count:           len(recommendations),
	}, nil
}

// SearchTrajectoriesRequest searches for past trajectories
type SearchTrajectoriesRequest struct {
	Domain      string   `json:"domain,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	MinSuccess  float64  `json:"min_success,omitempty"`
	ProblemType string   `json:"problem_type,omitempty"`
	Limit       int      `json:"limit,omitempty"`
}

// SearchTrajectoriesResponse returns matching trajectories
type SearchTrajectoriesResponse struct {
	Trajectories []*TrajectorySummary `json:"trajectories"`
	Count        int                  `json:"count"`
}

// TrajectorySummary is a summary of a trajectory
type TrajectorySummary struct {
	ID           string   `json:"id"`
	SessionID    string   `json:"session_id"`
	Problem      string   `json:"problem"`
	Domain       string   `json:"domain"`
	Strategy     string   `json:"strategy"`
	ToolsUsed    []string `json:"tools_used"`
	SuccessScore float64  `json:"success_score"`
	Duration     string   `json:"duration"`
	Tags         []string `json:"tags"`
}

// HandleSearchTrajectories searches for past reasoning trajectories
func (h *EpisodicMemoryHandler) HandleSearchTrajectories(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	req, err := unmarshalRequest[SearchTrajectoriesRequest](params)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	resp, err := h.searchTrajectories(ctx, req)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{Content: toJSONContent(resp)}, nil
}

// searchTrajectories is the typed internal implementation
func (h *EpisodicMemoryHandler) searchTrajectories(_ context.Context, req SearchTrajectoriesRequest) (*SearchTrajectoriesResponse, error) {
	if req.Limit == 0 {
		req.Limit = 10
	}

	// Get all trajectories and filter
	allTrajectories := h.store.GetAllTrajectories()

	// Filter by criteria - pre-allocate to input size
	limit := req.Limit
	if limit <= 0 {
		limit = 20 // Default limit
	}
	filtered := make([]*memory.ReasoningTrajectory, 0, limit)
	for _, traj := range allTrajectories {
		// Filter by domain
		if req.Domain != "" && traj.Domain != req.Domain {
			continue
		}

		// Filter by problem type
		if req.ProblemType != "" && traj.Problem != nil && traj.Problem.ProblemType != req.ProblemType {
			continue
		}

		// Filter by minimum success score
		if req.MinSuccess > 0 && traj.SuccessScore < req.MinSuccess {
			continue
		}

		// Filter by tags
		if len(req.Tags) > 0 {
			hasTag := false
			for _, reqTag := range req.Tags {
				for _, trajTag := range traj.Tags {
					if reqTag == trajTag {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		filtered = append(filtered, traj)
	}

	// Sort by success score (descending)
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[i].SuccessScore < filtered[j].SuccessScore {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	// Apply limit
	if req.Limit > 0 && len(filtered) > req.Limit {
		filtered = filtered[:req.Limit]
	}

	// Convert to summaries
	summaries := make([]*TrajectorySummary, len(filtered))
	for i, traj := range filtered {
		problemDesc := ""
		if traj.Problem != nil {
			problemDesc = traj.Problem.Description
		}

		strategy := ""
		toolsUsed := make([]string, 0)
		if traj.Approach != nil {
			strategy = traj.Approach.Strategy
			if traj.Approach.ToolSequence != nil {
				toolsUsed = traj.Approach.ToolSequence
			}
		}

		// Ensure tags is never nil (MCP requires arrays, not null)
		tags := traj.Tags
		if tags == nil {
			tags = make([]string, 0)
		}

		summaries[i] = &TrajectorySummary{
			ID:           traj.ID,
			SessionID:    traj.SessionID,
			Problem:      problemDesc,
			Domain:       traj.Domain,
			Strategy:     strategy,
			ToolsUsed:    toolsUsed,
			SuccessScore: traj.SuccessScore,
			Duration:     traj.Duration.String(),
			Tags:         tags,
		}
	}

	return &SearchTrajectoriesResponse{
		Trajectories: summaries,
		Count:        len(summaries),
	}, nil
}

// AnalyzeTrajectoryRequest requests retrospective analysis
type AnalyzeTrajectoryRequest struct {
	TrajectoryID string `json:"trajectory_id"`
}

// HandleAnalyzeTrajectory performs retrospective analysis of a completed session
func (h *EpisodicMemoryHandler) HandleAnalyzeTrajectory(ctx context.Context, params map[string]interface{}) (*mcp.CallToolResult, error) {
	req, err := unmarshalRequest[AnalyzeTrajectoryRequest](params)
	if err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	analysis, err := h.analyzeTrajectory(ctx, req)
	if err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{Content: toJSONContent(analysis)}, nil
}

// analyzeTrajectory is the typed internal implementation
func (h *EpisodicMemoryHandler) analyzeTrajectory(ctx context.Context, req AnalyzeTrajectoryRequest) (*memory.RetrospectiveAnalysis, error) {
	if req.TrajectoryID == "" {
		return nil, fmt.Errorf("trajectory_id is required")
	}

	analysis, err := h.retrospective.AnalyzeTrajectory(ctx, req.TrajectoryID)
	if err != nil {
		return nil, err
	}

	// Additional defensive nil checks for MCP validation (belt-and-suspenders approach)
	if analysis.Strengths == nil {
		analysis.Strengths = []string{}
	}
	if analysis.Weaknesses == nil {
		analysis.Weaknesses = []string{}
	}
	if analysis.Improvements == nil {
		analysis.Improvements = []*memory.ImprovementSuggestion{}
	}
	if analysis.LessonsLearned == nil {
		analysis.LessonsLearned = []string{}
	}

	return analysis, nil
}

// ComputeProblemHash is exported for testing
func ComputeProblemHash(problem *memory.ProblemDescription) string {
	return memory.ComputeProblemHash(problem)
}

// RegisterEpisodicMemoryTools registers episodic memory tools with the MCP server
func RegisterEpisodicMemoryTools(mcpServer *mcp.Server, handler *EpisodicMemoryHandler) {
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "start-reasoning-session",
		Description: `Start tracking a reasoning session to build episodic memory and learn from experience.

The episodic memory system enables the server to learn from past reasoning sessions, 
recognize successful patterns, and provide adaptive recommendations.

**Parameters:**
- session_id (required): Unique session identifier
- description (required): Problem description
- goals (optional): Array of goals to achieve
- domain (optional): Problem domain (e.g., "software-engineering", "science", "business")
- context (optional): Additional context about the problem
- complexity (optional): Estimated complexity 0.0-1.0
- metadata (optional): Additional metadata

**Returns:**
- session_id: Session identifier
- problem_id: Problem fingerprint hash
- status: "active"
- suggestions: Array of recommendations based on similar past problems

**Use Cases:**
1. Before complex reasoning: Get suggestions from similar past successes
2. Learning from failures: System warns about approaches that historically fail
3. Continuous improvement: Performance improves with every reasoning session

**Example:**
{
  "session_id": "debug_2024_001",
  "description": "Optimize database query performance",
  "goals": ["Reduce query time", "Improve user experience"],
  "domain": "software-engineering",
  "complexity": 0.6
}`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input StartSessionRequest) (*mcp.CallToolResult, *StartSessionResponse, error) {
		var params map[string]interface{}
		paramsBytes, err := json.Marshal(input)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal input: %w", err)
		}
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal params: %w", err)
		}

		result, err := handler.HandleStartSession(ctx, params)
		if err != nil {
			return nil, nil, err
		}

		// Unmarshal result back into response for MCP schema validation
		response := &StartSessionResponse{
			Suggestions: make([]*memory.Recommendation, 0), // Initialize array to prevent nil
		}
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				if err := json.Unmarshal([]byte(textContent.Text), response); err != nil {
					log.Printf("Warning: failed to unmarshal response: %v", err)
				}
			}
		}
		return result, response, nil
	})

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "complete-reasoning-session",
		Description: `Complete a reasoning session and store the trajectory for learning.

Marks a session as complete, calculates quality metrics, and triggers pattern learning.
The system learns which approaches work best for different problem types.

**Parameters:**
- session_id (required): Session to complete
- status (required): "success", "partial", or "failure"
- goals_achieved (optional): Array of achieved goals
- goals_failed (optional): Array of failed goals
- solution (optional): Description of solution
- confidence (optional): Confidence in solution (0.0-1.0)
- unexpected_outcomes (optional): Array of unexpected results

**Returns:**
- trajectory_id: Stored trajectory identifier
- session_id: Session identifier
- success_score: Calculated success score (0.0-1.0)
- quality_score: Overall quality score (0.0-1.0)
- patterns_found: Number of patterns updated
- status: "completed"

**Quality Metrics Calculated:**
- Efficiency: Steps taken vs optimal
- Coherence: Logical consistency
- Completeness: Goal coverage
- Innovation: Creative tool usage
- Reliability: Confidence in result

**Example:**
{
  "session_id": "debug_2024_001",
  "status": "success",
  "goals_achieved": ["Reduce query time"],
  "solution": "Added indexes and optimized queries",
  "confidence": 0.85
}`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CompleteSessionRequest) (*mcp.CallToolResult, *CompleteSessionResponse, error) {
		var params map[string]interface{}
		paramsBytes, err := json.Marshal(input)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal input: %w", err)
		}
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal params: %w", err)
		}

		result, err := handler.HandleCompleteSession(ctx, params)
		if err != nil {
			return nil, nil, err
		}

		// Properly unmarshal result for MCP validation
		response := &CompleteSessionResponse{}
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				if err := json.Unmarshal([]byte(textContent.Text), response); err != nil {
					log.Printf("warning: failed to unmarshal response for validation: %v", err)
				}
			}
		}
		return result, response, nil
	})

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "get-recommendations",
		Description: `Get adaptive recommendations based on episodic memory of similar past problems.

Retrieves recommendations from the episodic memory system based on similarity to past 
successful reasoning sessions. Includes learned patterns and historical success rates.

**Parameters:**
- description (required): Problem description
- goals (optional): Array of problem goals
- domain (optional): Problem domain
- context (optional): Additional context
- complexity (optional): Estimated complexity (0.0-1.0)
- limit (optional): Max recommendations (default: 5)

**Returns:**
- recommendations: Array of recommendations with:
  - type: "tool_sequence", "approach", "warning", or "optimization"
  - priority: Relevance score
  - suggestion: Specific advice
  - reasoning: Why this recommendation
  - success_rate: Historical success rate
- similar_cases: Count of similar past trajectories
- learned_patterns: Applicable learned patterns
- count: Number of recommendations

**Recommendation Types:**
1. **tool_sequence**: Proven tool sequences (success rate >70%)
2. **approach**: Successful reasoning strategies
3. **warning**: Approaches that historically fail (<40% success)
4. **optimization**: Performance improvements

**Example:**
{
  "description": "Need to implement user authentication",
  "domain": "security",
  "goals": ["Secure login", "Session management"],
  "limit": 3
}`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetRecommendationsRequest) (*mcp.CallToolResult, *GetRecommendationsResponse, error) {
		var params map[string]interface{}
		paramsBytes, err := json.Marshal(input)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal input: %w", err)
		}
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal params: %w", err)
		}

		result, err := handler.HandleGetRecommendations(ctx, params)
		if err != nil {
			return nil, nil, err
		}

		// Unmarshal result back into response for MCP schema validation
		response := &GetRecommendationsResponse{
			Recommendations: make([]*memory.Recommendation, 0), // Initialize arrays to prevent nil
			LearnedPatterns: make([]*memory.TrajectoryPattern, 0),
		}
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				if err := json.Unmarshal([]byte(textContent.Text), response); err != nil {
					log.Printf("Warning: failed to unmarshal response: %v", err)
				}
			}
		}
		return result, response, nil
	})

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "search-trajectories",
		Description: `Search for past reasoning trajectories to learn from experience.

Find past reasoning sessions by domain, tags, success rate, or problem type. Useful for 
understanding what worked in the past and learning from both successes and failures.

**Parameters:**
- domain (optional): Filter by domain
- tags (optional): Array of tags to filter by
- min_success (optional): Minimum success score (0.0-1.0)
- problem_type (optional): Filter by problem type
- limit (optional): Max results (default: 10)

**Returns:**
- trajectories: Array of trajectory summaries with:
  - id: Trajectory identifier
  - session_id: Original session ID
  - problem: Problem description
  - domain: Problem domain
  - strategy: Strategy used
  - tools_used: Array of tools used
  - success_score: Success score (0.0-1.0)
  - duration: Session duration
  - tags: Array of tags
- count: Number of results

**Use Cases:**
1. Review successful approaches for a domain
2. Learn from failures (min_success: 0.0-0.4)
3. Find high-performing strategies (min_success: 0.8-1.0)
4. Analyze tool usage patterns

**Example:**
{
  "domain": "software-engineering",
  "min_success": 0.7,
  "limit": 5
}`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SearchTrajectoriesRequest) (*mcp.CallToolResult, *SearchTrajectoriesResponse, error) {
		var params map[string]interface{}
		paramsBytes, err := json.Marshal(input)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal input: %w", err)
		}
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal params: %w", err)
		}

		result, err := handler.HandleSearchTrajectories(ctx, params)
		if err != nil {
			return nil, nil, err
		}

		// Unmarshal result back into response for MCP schema validation
		response := &SearchTrajectoriesResponse{
			Trajectories: make([]*TrajectorySummary, 0), // Initialize array to prevent nil
		}
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				if err := json.Unmarshal([]byte(textContent.Text), response); err != nil {
					log.Printf("Warning: failed to unmarshal response: %v", err)
				}
			}
		}
		return result, response, nil
	})

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "analyze-trajectory",
		Description: `Perform retrospective analysis of a completed reasoning session.

Provides comprehensive post-session analysis including strengths, weaknesses, actionable 
improvements, lessons learned, and comparative analysis against similar past sessions.

**Parameters:**
- trajectory_id (required): ID of trajectory to analyze (returned from complete-reasoning-session)

**Returns:**
- summary: High-level assessment with success/quality scores, duration, strategy
- strengths: What went well (efficiency, coherence, completeness, innovation, reliability)
- weaknesses: Areas for improvement with specific metrics
- improvements: Prioritized actionable suggestions with expected impact
- lessons_learned: Key takeaways for future sessions
- comparative_analysis: How this session compares to similar past sessions (percentile rank)
- detailed_metrics: Deep dive into each quality metric with explanations and suggestions

**Quality Metrics Analyzed:**
1. **Efficiency**: Steps taken vs optimal (7-10 steps baseline)
2. **Coherence**: Logical consistency (contradictions, fallacies)
3. **Completeness**: Goal achievement rate
4. **Innovation**: Use of creative/advanced tools
5. **Reliability**: Confidence in results

**Improvement Categories:**
- efficiency: Reduce unnecessary steps
- quality: Improve logical consistency
- approach: Change reasoning strategy
- tools: Use different/better tools

**Use Cases:**
1. Learn from successful sessions - understand what worked
2. Improve future performance - get specific actionable advice
3. Track progress - see percentile rank vs similar problems
4. Identify patterns - discover your reasoning strengths/weaknesses

**Example:**
{
  "trajectory_id": "traj_session_001_problem_abc_1234567890"
}

**Returns comprehensive analysis including:**
- Overall assessment: "excellent", "good", "fair", or "poor"
- Top 3-5 strengths with metrics
- Top 3-5 weaknesses with root causes
- Prioritized improvement suggestions
- Percentile rank (e.g., "better than 75% of similar sessions")
- Detailed metric breakdowns with actionable next steps`,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AnalyzeTrajectoryRequest) (*mcp.CallToolResult, *memory.RetrospectiveAnalysis, error) {
		var params map[string]interface{}
		paramsBytes, err := json.Marshal(input)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal input: %w", err)
		}
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal params: %w", err)
		}

		result, err := handler.HandleAnalyzeTrajectory(ctx, params)
		if err != nil {
			return nil, nil, err
		}

		// Properly unmarshal result for MCP validation
		response := &memory.RetrospectiveAnalysis{
			Strengths:      []string{},
			Weaknesses:     []string{},
			Improvements:   []*memory.ImprovementSuggestion{},
			LessonsLearned: []string{},
		}
		if len(result.Content) > 0 {
			if textContent, ok := result.Content[0].(*mcp.TextContent); ok {
				if err := json.Unmarshal([]byte(textContent.Text), response); err != nil {
					log.Printf("warning: failed to unmarshal response for validation: %v", err)
				}
			}
		}
		return result, response, nil
	})
}
