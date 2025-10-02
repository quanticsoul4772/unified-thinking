package reasoning

import (
	"fmt"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// DecisionMaker provides structured decision-making frameworks
type DecisionMaker struct {
	mu      sync.RWMutex
	counter int
}

// NewDecisionMaker creates a new decision maker
func NewDecisionMaker() *DecisionMaker {
	return &DecisionMaker{}
}

// CreateDecision creates a structured decision framework
func (dm *DecisionMaker) CreateDecision(question string, options []*types.DecisionOption, criteria []*types.DecisionCriterion) (*types.Decision, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	dm.counter++

	if len(options) == 0 {
		return nil, fmt.Errorf("at least one option is required")
	}
	if len(criteria) == 0 {
		return nil, fmt.Errorf("at least one criterion is required")
	}

	// Normalize criterion weights to sum to 1.0
	totalWeight := 0.0
	for _, c := range criteria {
		totalWeight += c.Weight
	}
	if totalWeight > 0 {
		for _, c := range criteria {
			c.Weight = c.Weight / totalWeight
		}
	}

	// Calculate total scores for each option
	for _, option := range options {
		totalScore := 0.0
		for _, criterion := range criteria {
			score, exists := option.Scores[criterion.ID]
			if exists {
				// Apply weight and consider maximize/minimize
				if criterion.Maximize {
					totalScore += score * criterion.Weight
				} else {
					// For minimize criteria, invert the score
					totalScore += (1.0 - score) * criterion.Weight
				}
			}
		}
		option.TotalScore = totalScore
	}

	// Find best option
	bestOption := options[0]
	for _, option := range options {
		if option.TotalScore > bestOption.TotalScore {
			bestOption = option
		}
	}

	// Calculate confidence based on margin of victory
	confidence := dm.calculateDecisionConfidence(options, bestOption)

	decision := &types.Decision{
		ID:             fmt.Sprintf("decision-%d", dm.counter),
		Question:       question,
		Options:        options,
		Criteria:       criteria,
		Recommendation: fmt.Sprintf("Recommended option: %s (score: %.2f)", bestOption.Name, bestOption.TotalScore),
		Confidence:     confidence,
		Metadata:       map[string]interface{}{},
		CreatedAt:      time.Now(),
	}

	return decision, nil
}

// calculateDecisionConfidence estimates confidence based on score separation
func (dm *DecisionMaker) calculateDecisionConfidence(options []*types.DecisionOption, bestOption *types.DecisionOption) float64 {
	if len(options) <= 1 {
		return 0.5 // Low confidence with only one option
	}

	// Find second-best option
	var secondBest *types.DecisionOption
	for _, option := range options {
		if option.ID != bestOption.ID {
			if secondBest == nil || option.TotalScore > secondBest.TotalScore {
				secondBest = option
			}
		}
	}

	if secondBest == nil {
		return 0.5
	}

	// Confidence based on margin: larger margin = higher confidence
	margin := bestOption.TotalScore - secondBest.TotalScore

	// Convert margin to confidence (0.5 to 1.0 range)
	confidence := 0.5 + (margin * 0.5)

	// Clamp to valid range
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.5 {
		confidence = 0.5
	}

	return confidence
}

// AddOption adds a new option to a decision
func (dm *DecisionMaker) AddOption(decision *types.Decision, name, description string, scores map[string]float64, pros, cons []string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	optionID := fmt.Sprintf("option-%d-%d", dm.counter, len(decision.Options)+1)

	option := &types.DecisionOption{
		ID:          optionID,
		Name:        name,
		Description: description,
		Scores:      scores,
		Pros:        pros,
		Cons:        cons,
	}

	decision.Options = append(decision.Options, option)
	return nil
}

// AddCriterion adds a new criterion to a decision
func (dm *DecisionMaker) AddCriterion(decision *types.Decision, name, description string, weight float64, maximize bool) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	criterionID := fmt.Sprintf("criterion-%d-%d", dm.counter, len(decision.Criteria)+1)

	criterion := &types.DecisionCriterion{
		ID:          criterionID,
		Name:        name,
		Description: description,
		Weight:      weight,
		Maximize:    maximize,
	}

	decision.Criteria = append(decision.Criteria, criterion)
	return nil
}

// ProblemDecomposer breaks down complex problems into subproblems
type ProblemDecomposer struct {
	mu      sync.RWMutex
	counter int
}

// NewProblemDecomposer creates a new problem decomposer
func NewProblemDecomposer() *ProblemDecomposer {
	return &ProblemDecomposer{}
}

// DecomposeProblem breaks down a problem into manageable subproblems
func (pd *ProblemDecomposer) DecomposeProblem(problem string) (*types.ProblemDecomposition, error) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	pd.counter++

	// Simple heuristic decomposition based on problem structure
	subproblems := pd.identifySubproblems(problem)
	dependencies := pd.identifyDependencies(subproblems)
	solutionPath := pd.determineSolutionPath(subproblems, dependencies)

	decomposition := &types.ProblemDecomposition{
		ID:           fmt.Sprintf("decomposition-%d", pd.counter),
		Problem:      problem,
		Subproblems:  subproblems,
		Dependencies: dependencies,
		SolutionPath: solutionPath,
		Metadata:     map[string]interface{}{},
		CreatedAt:    time.Now(),
	}

	return decomposition, nil
}

// identifySubproblems identifies component subproblems (heuristic approach)
func (pd *ProblemDecomposer) identifySubproblems(problem string) []*types.Subproblem {
	// In a real implementation, this would use NLP or more sophisticated analysis
	// For now, we provide a simple template-based approach

	subproblems := []*types.Subproblem{
		{
			ID:          fmt.Sprintf("subproblem-%d-1", pd.counter),
			Description: "Analyze and define the problem scope",
			Complexity:  "low",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-2", pd.counter),
			Description: "Gather required information and resources",
			Complexity:  "medium",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-3", pd.counter),
			Description: "Develop potential solutions",
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-4", pd.counter),
			Description: "Evaluate and select best approach",
			Complexity:  "medium",
			Priority:    "medium",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-5", pd.counter),
			Description: "Implement and test solution",
			Complexity:  "high",
			Priority:    "medium",
			Status:      "pending",
		},
	}

	return subproblems
}

// identifyDependencies identifies dependencies between subproblems
func (pd *ProblemDecomposer) identifyDependencies(subproblems []*types.Subproblem) []*types.Dependency {
	dependencies := make([]*types.Dependency, 0)

	// Sequential dependencies for our template approach
	for i := 0; i < len(subproblems)-1; i++ {
		dependency := &types.Dependency{
			FromSubproblem: subproblems[i].ID,
			ToSubproblem:   subproblems[i+1].ID,
			Type:           "required",
		}
		dependencies = append(dependencies, dependency)
	}

	return dependencies
}

// determineSolutionPath determines optimal order to solve subproblems
func (pd *ProblemDecomposer) determineSolutionPath(subproblems []*types.Subproblem, dependencies []*types.Dependency) []string {
	// Simple topological sort based on dependencies
	// For our template approach, it's sequential
	path := make([]string, len(subproblems))
	for i, sp := range subproblems {
		path[i] = sp.ID
	}
	return path
}

// UpdateSubproblemStatus updates the status of a subproblem
func (pd *ProblemDecomposer) UpdateSubproblemStatus(decomposition *types.ProblemDecomposition, subproblemID, status, solution string) error {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	for _, sp := range decomposition.Subproblems {
		if sp.ID == subproblemID {
			sp.Status = status
			if solution != "" {
				sp.Solution = solution
			}
			return nil
		}
	}
	return fmt.Errorf("subproblem not found: %s", subproblemID)
}
