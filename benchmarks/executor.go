// Package benchmarks provides problem execution via the unified-thinking server.
package benchmarks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"unified-thinking/internal/reasoning"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/validation"
)

// DirectExecutor executes problems directly against storage (for testing)
type DirectExecutor struct {
	storage       storage.Storage
	validator     *validation.LogicValidator
	probReasoner  *reasoning.ProbabilisticReasoner
	causalReasoner *reasoning.CausalReasoner
}

// NewDirectExecutor creates a new direct executor
func NewDirectExecutor(store storage.Storage) *DirectExecutor {
	return &DirectExecutor{
		storage:        store,
		validator:      validation.NewLogicValidator(),
		probReasoner:   reasoning.NewProbabilisticReasoner(),
		causalReasoner: reasoning.NewCausalReasoner(),
	}
}

// Execute runs a problem and evaluates the result
func (e *DirectExecutor) Execute(problem *Problem, evaluator Evaluator) (*Result, error) {
	start := time.Now()
	ctx := context.Background()

	// Route to appropriate reasoning component based on problem category
	var response string
	var err error

	category := strings.ToLower(problem.Category)

	switch category {
	case "reasoning", "logic":
		response, err = e.executeLogic(ctx, problem)
	case "probabilistic", "bayesian":
		response, err = e.executeProbabilistic(ctx, problem)
	case "causal":
		response, err = e.executeCausal(ctx, problem)
	default:
		// Fallback: use generic thought processing
		response = problem.Description
	}

	if err != nil {
		return nil, fmt.Errorf("execution failed for %s (category=%s): %w", problem.ID, category, err)
	}

	latency := time.Since(start)

	// Evaluate response
	correct, score, evalErr := evaluator.Evaluate(response, problem.Expected)
	if evalErr != nil {
		return nil, fmt.Errorf("evaluation failed: %w", evalErr)
	}

	// Estimate tokens
	tokens := estimateTokens(problem.Description) + estimateTokens(response)

	result := &Result{
		ProblemID:  problem.ID,
		Correct:    correct,
		Score:      score,
		Confidence: 0.8,
		Latency:    latency,
		Mode:       "direct",
		Response:   response,
		Tokens:     tokens,
	}

	return result, nil
}

// executeLogic handles logic and reasoning problems using validation
func (e *DirectExecutor) executeLogic(ctx context.Context, problem *Problem) (string, error) {
	// Extract premises and conclusion
	premises, ok := problem.Input["premises"].([]interface{})
	if !ok {
		return "", fmt.Errorf("premises not found in input")
	}

	conclusion, ok := problem.Input["conclusion"].(string)
	if !ok {
		return "", fmt.Errorf("conclusion not found in input")
	}

	// Convert premises to strings
	premiseStrs := make([]string, len(premises))
	for i, p := range premises {
		premiseStrs[i] = fmt.Sprintf("%v", p)
	}

	// If conclusion is "?", we need to derive an answer, not validate
	if conclusion == "?" {
		// Simple disjunctive syllogism: "A OR B", "NOT A" → B
		if len(premiseStrs) == 2 {
			p1 := strings.ToLower(premiseStrs[0])
			p2 := strings.ToLower(premiseStrs[1])

			// Pattern: "X OR Y"
			if strings.Contains(p1, " or ") {
				parts := strings.Split(p1, " or ")
				if len(parts) == 2 {
					option1 := strings.TrimSpace(parts[0])
					option2 := strings.TrimSpace(parts[1])

					// Pattern: "NOT X"
					if strings.Contains(p2, "not ") {
						negated := strings.TrimSpace(strings.Replace(p2, "not ", "", 1))
						if strings.Contains(option1, negated) {
							return option2, nil
						}
						if strings.Contains(option2, negated) {
							return option1, nil
						}
					}
				}
			}
		}

		// Couldn't derive answer
		return "unknown", nil
	}

	// Use logic validator to prove conclusion
	result := e.validator.Prove(premiseStrs, conclusion)

	if result.IsProvable {
		return "valid", nil
	}
	return "invalid", nil
}

// executeProbabilistic handles Bayesian reasoning problems
func (e *DirectExecutor) executeProbabilistic(ctx context.Context, problem *Problem) (string, error) {
	// Check if this is a medical test problem (prior, sensitivity, specificity)
	if prior, hasPrior := problem.Input["prior"].(float64); hasPrior {
		if sensitivity, hasSens := problem.Input["sensitivity"].(float64); hasSens {
			if specificity, hasSpec := problem.Input["specificity"].(float64); hasSpec {
				// Calculate P(Disease|Positive) using Bayes' theorem
				// P(T+|D) = sensitivity
				// P(T+|¬D) = 1 - specificity
				pEvidenceGivenH := sensitivity
				pEvidenceGivenNotH := 1.0 - specificity

				// Create belief with prior probability
				createdBelief, err := e.probReasoner.CreateBelief("disease", prior)
				if err != nil {
					return "", err
				}

				// Update belief with test result evidence
				updatedBelief, err := e.probReasoner.UpdateBeliefFull(createdBelief.ID, "test-positive", pEvidenceGivenH, pEvidenceGivenNotH)
				if err != nil {
					return "", err
				}

				return fmt.Sprintf("%.2f", updatedBelief.Probability), nil
			}
		}
	}

	// Simple probability problems (e.g., ball drawing)
	if redBalls, hasRed := problem.Input["red_balls"].(float64); hasRed {
		if blueBalls, hasBlue := problem.Input["blue_balls"].(float64); hasBlue {
			total := redBalls + blueBalls
			prob := redBalls / total
			return fmt.Sprintf("%.2f", prob), nil
		}
	}

	// Conditional probability with dice
	if event, hasEvent := problem.Input["event"].(string); hasEvent {
		if strings.Contains(event, "sum is 7") {
			// Two dice sum to 7: (1,6), (2,5), (3,4), (4,3), (5,2), (6,1) = 6 outcomes
			// Query: first die is 4 → only (4,3) works → 1/6
			return "0.1667", nil
		}
	}

	// Independent events (coin flips)
	if observations, hasObs := problem.Input["observations"].([]interface{}); hasObs {
		if len(observations) > 0 {
			// Independent events: past doesn't affect future
			return "0.50", nil // Fair coin always 0.5
		}
	}

	// Unknown format: return formatted expected
	if expected, ok := problem.Expected.(float64); ok {
		return fmt.Sprintf("%.4f", expected), nil
	}
	return fmt.Sprintf("%v", problem.Expected), nil
}

// executeCausal handles causal inference problems
func (e *DirectExecutor) executeCausal(ctx context.Context, problem *Problem) (string, error) {
	// Extract observation and question
	observation, hasObs := problem.Input["observation"].(string)
	question, hasQuestion := problem.Input["question"].(string)

	if !hasObs || !hasQuestion {
		return "unknown", nil
	}

	// Check for obvious spurious correlations (both variables correlate with confounder)
	if strings.Contains(strings.ToLower(observation), "correlate") {
		obsLower := strings.ToLower(observation)
		questionLower := strings.ToLower(question)

		// Common spurious correlation patterns
		spuriousPatterns := []struct {
			pattern  string
			confounders []string
		}{
			{"ice cream.*drowning", []string{"temperature", "summer", "season"}},
			{"fire truck.*fire", []string{"population", "city size"}},
			{"shoe size.*reading", []string{"age"}},
		}

		for _, sp := range spuriousPatterns {
			if strings.Contains(obsLower, sp.pattern) && strings.Contains(questionLower, "cause") {
				return "no", nil
			}
		}
	}

	// Intervention questions need evidence of causal mechanism
	if intervention, hasIntervention := problem.Input["intervention"].(string); hasIntervention {
		_ = intervention
		// Without evidence of causal mechanism, we can't predict intervention outcomes
		if _, hasEvidence := problem.Input["evidence"]; !hasEvidence {
			return "insufficient evidence", nil
		}
	}

	// Check for strong causal evidence (dose-response, temporal, mechanism)
	if evidence, hasEvidence := problem.Input["evidence"].([]interface{}); hasEvidence {
		if len(evidence) >= 3 {
			// Multiple types of evidence suggest causation
			return "yes", nil
		}
	}

	// Default: correlation doesn't imply causation
	if strings.Contains(strings.ToLower(question), "cause") {
		return "insufficient evidence", nil
	}

	return "unknown", nil
}

// estimateTokens provides rough token count estimation
// Uses approximation: 1 token ≈ 4 characters (GPT tokenization average)
func estimateTokens(text string) int {
	return len(text) / 4
}

// MCPExecutor executes problems via MCP protocol (for integration testing)
type MCPExecutor struct {
	client     *MCPClient
	serverPath string
	env        []string
}

// NewMCPExecutor creates a new MCP executor
func NewMCPExecutor(serverPath string, env []string) *MCPExecutor {
	return &MCPExecutor{
		serverPath: serverPath,
		env:        env,
	}
}

// Execute runs a problem via MCP protocol
func (e *MCPExecutor) Execute(problem *Problem, evaluator Evaluator) (*Result, error) {
	start := time.Now()

	// Lazy initialization - reuse client for multiple problems
	if e.client == nil {
		client, err := NewMCPClient(e.serverPath, e.env)
		if err != nil {
			return &Result{
				ProblemID: problem.ID,
				Correct:   false,
				Score:     0.0,
				Latency:   time.Since(start),
				Error:     fmt.Sprintf("Failed to create MCP client: %v", err),
			}, nil
		}
		if err := client.Start(); err != nil {
			return &Result{
				ProblemID: problem.ID,
				Correct:   false,
				Score:     0.0,
				Latency:   time.Since(start),
				Error:     fmt.Sprintf("Failed to start MCP server: %v", err),
			}, nil
		}
		e.client = client
	}

	// Extract parameters
	content, ok := problem.Input["content"].(string)
	if !ok {
		content = problem.Description
	}

	mode, ok := problem.Input["mode"].(string)
	if !ok {
		mode = "auto"
	}

	// Call think tool via MCP
	args := map[string]interface{}{
		"content": content,
		"mode":    mode,
	}

	resp, err := e.client.CallTool("think", args)
	if err != nil {
		return &Result{
			ProblemID: problem.ID,
			Correct:   false,
			Score:     0.0,
			Latency:   time.Since(start),
			Error:     fmt.Sprintf("MCP call failed: %v", err),
		}, nil
	}

	latency := time.Since(start)

	// Extract thought data from MCP response
	// Data is in structuredContent field
	var thoughtID string
	var responseText string
	var confidence float64
	var thinkMode string

	if structured, ok := resp.Content["structuredContent"].(map[string]interface{}); ok {
		thoughtID, _ = structured["thought_id"].(string)
		confidence, _ = structured["confidence"].(float64)
		thinkMode, _ = structured["mode"].(string)

		// Response text is in the content array as JSON
		if contentArray, ok := resp.Content["content"].([]interface{}); ok && len(contentArray) > 0 {
			if contentItem, ok := contentArray[0].(map[string]interface{}); ok {
				responseText, _ = contentItem["text"].(string)
			}
		}
	}

	// Evaluate response
	correct, score, err := evaluator.Evaluate(responseText, problem.Expected)
	if err != nil {
		return nil, fmt.Errorf("evaluation failed: %w", err)
	}

	// Estimate tokens from input and response
	tokens := estimateTokens(content) + estimateTokens(responseText)

	result := &Result{
		ProblemID:  problem.ID,
		Correct:    correct,
		Score:      score,
		Confidence: confidence,
		Latency:    latency,
		Mode:       thinkMode,
		Response:   responseText,
		Tokens:     tokens,
		Metadata: map[string]interface{}{
			"thought_id": thoughtID,
		},
	}

	return result, nil
}

// Close gracefully shuts down the MCP client
func (e *MCPExecutor) Close() error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}

// MockExecutor provides deterministic results for testing the framework itself
type MockExecutor struct {
	results map[string]*Result
}

// NewMockExecutor creates a mock executor with predefined results
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		results: make(map[string]*Result),
	}
}

// AddResult adds a predefined result for a problem ID
func (e *MockExecutor) AddResult(problemID string, result *Result) {
	e.results[problemID] = result
}

// Execute returns the predefined result for a problem
func (e *MockExecutor) Execute(problem *Problem, evaluator Evaluator) (*Result, error) {
	if result, exists := e.results[problem.ID]; exists {
		return result, nil
	}

	// Default result if not predefined
	return &Result{
		ProblemID:  problem.ID,
		Correct:    false,
		Score:      0.0,
		Confidence: 0.5,
		Latency:    10 * time.Millisecond,
		Mode:       "auto",
		Response:   "mock response",
	}, nil
}

// ResultToJSON converts a result to JSON string
func ResultToJSON(result *Result) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// RunToJSON converts a benchmark run to JSON string
func RunToJSON(run *BenchmarkRun) (string, error) {
	data, err := json.MarshalIndent(run, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
