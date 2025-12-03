package reasoning

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// DecisionMaker provides structured decision-making frameworks
type DecisionMaker struct {
	mu        sync.RWMutex
	counter   int
	decisions map[string]*types.Decision // Storage for created decisions
}

// NewDecisionMaker creates a new decision maker
func NewDecisionMaker() *DecisionMaker {
	return &DecisionMaker{
		decisions: make(map[string]*types.Decision),
	}
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

	// Store the decision for future retrieval and re-evaluation
	dm.decisions[decision.ID] = decision

	return decision, nil
}

// GetDecision retrieves a decision by ID
func (dm *DecisionMaker) GetDecision(decisionID string) (*types.Decision, error) {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	decision, exists := dm.decisions[decisionID]
	if !exists {
		return nil, fmt.Errorf("decision not found: %s", decisionID)
	}

	return decision, nil
}

// ListDecisions returns all stored decisions
func (dm *DecisionMaker) ListDecisions() []*types.Decision {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	decisions := make([]*types.Decision, 0, len(dm.decisions))
	for _, d := range dm.decisions {
		decisions = append(decisions, d)
	}

	return decisions
}

// RecalculateDecision re-evaluates a decision with updated scores
// This is useful when new evidence affects option scores
func (dm *DecisionMaker) RecalculateDecision(decisionID string, scoreAdjustments map[string]map[string]float64) (*types.Decision, error) {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	decision, exists := dm.decisions[decisionID]
	if !exists {
		return nil, fmt.Errorf("decision not found: %s", decisionID)
	}

	// Apply score adjustments to options
	// scoreAdjustments: map[optionID]map[criterionID]adjustment
	for optionID, criterionAdjustments := range scoreAdjustments {
		for _, option := range decision.Options {
			if option.ID == optionID {
				for criterionID, adjustment := range criterionAdjustments {
					if score, exists := option.Scores[criterionID]; exists {
						// Adjust score and clamp to [0, 1]
						newScore := score + adjustment
						if newScore > 1.0 {
							newScore = 1.0
						}
						if newScore < 0.0 {
							newScore = 0.0
						}
						option.Scores[criterionID] = newScore
					}
				}
				break
			}
		}
	}

	// Recalculate total scores for all options
	for _, option := range decision.Options {
		totalScore := 0.0
		for _, criterion := range decision.Criteria {
			score, exists := option.Scores[criterion.ID]
			if exists {
				if criterion.Maximize {
					totalScore += score * criterion.Weight
				} else {
					totalScore += (1.0 - score) * criterion.Weight
				}
			}
		}
		option.TotalScore = totalScore
	}

	// Find new best option
	bestOption := decision.Options[0]
	for _, option := range decision.Options {
		if option.TotalScore > bestOption.TotalScore {
			bestOption = option
		}
	}

	// Update recommendation and confidence
	decision.Recommendation = fmt.Sprintf("Recommended option: %s (score: %.2f)", bestOption.Name, bestOption.TotalScore)
	decision.Confidence = dm.calculateDecisionConfidence(decision.Options, bestOption)

	// Add metadata about recalculation
	if decision.Metadata == nil {
		decision.Metadata = make(map[string]interface{})
	}
	decision.Metadata["last_recalculated"] = time.Now()
	recalcCount := 0
	if count, ok := decision.Metadata["recalculation_count"].(int); ok {
		recalcCount = count
	}
	decision.Metadata["recalculation_count"] = recalcCount + 1

	return decision, nil
}

// DeleteDecision removes a decision from storage
func (dm *DecisionMaker) DeleteDecision(decisionID string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if _, exists := dm.decisions[decisionID]; !exists {
		return fmt.Errorf("decision not found: %s", decisionID)
	}

	delete(dm.decisions, decisionID)
	return nil
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

// identifySubproblems identifies component subproblems using entity-aware templates
func (pd *ProblemDecomposer) identifySubproblems(problem string) []*types.Subproblem {
	// Extract key entities from the problem statement
	entities := pd.extractProblemEntities(problem)
	entityStr := strings.Join(entities, ", ")
	if entityStr == "" {
		entityStr = "the given problem"
	}

	// Detect domain for appropriate template selection
	domain := DetectDomain(problem)

	// Generate problem-specific subproblems based on domain and entities
	var subproblems []*types.Subproblem

	switch domain {
	case DomainResearch:
		subproblems = pd.generateResearchSubproblems(problem, entities, entityStr)
	case DomainDebugging:
		subproblems = pd.generateDebuggingSubproblems(problem, entities, entityStr)
	case DomainArchitecture:
		subproblems = pd.generateArchitectureSubproblems(problem, entities, entityStr)
	case DomainProof:
		subproblems = pd.generateProofSubproblems(problem, entities, entityStr)
	default:
		subproblems = pd.generateGeneralSubproblems(problem, entities, entityStr)
	}

	return subproblems
}

// extractProblemEntities extracts meaningful noun phrases and key terms from the problem
func (pd *ProblemDecomposer) extractProblemEntities(problem string) []string {
	entities := make([]string, 0)
	seen := make(map[string]bool)

	// Common stop words to filter out
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true, "was": true,
		"were": true, "be": true, "been": true, "being": true, "have": true,
		"has": true, "had": true, "do": true, "does": true, "did": true,
		"will": true, "would": true, "could": true, "should": true, "may": true,
		"might": true, "can": true, "must": true, "shall": true, "to": true,
		"of": true, "in": true, "for": true, "on": true, "with": true, "at": true,
		"by": true, "from": true, "up": true, "about": true, "into": true, "over": true,
		"after": true, "and": true, "or": true, "but": true, "if": true, "then": true,
		"how": true, "what": true, "why": true, "when": true, "where": true, "which": true,
		"who": true, "this": true, "that": true, "these": true, "those": true,
		"it": true, "its": true, "they": true, "their": true, "them": true,
		"we": true, "our": true, "you": true, "your": true, "i": true, "my": true,
		"some": true, "any": true, "all": true, "most": true, "many": true, "much": true,
		"more": true, "less": true, "very": true, "just": true, "only": true,
	}

	// Normalize and tokenize
	words := strings.Fields(strings.ToLower(problem))

	// Build bigrams and filter for meaningful phrases
	for i := 0; i < len(words); i++ {
		word := strings.Trim(words[i], ".,?!;:\"'()[]{}") // Remove punctuation

		// Skip stop words and short words
		if len(word) < 3 || stopWords[word] {
			continue
		}

		// Check for bigram (two-word phrase) if not at end
		if i < len(words)-1 {
			nextWord := strings.Trim(words[i+1], ".,?!;:\"'()[]{}") // Remove punctuation
			if len(nextWord) >= 3 && !stopWords[nextWord] {
				bigram := word + " " + nextWord
				if !seen[bigram] {
					seen[bigram] = true
					entities = append(entities, bigram)
				}
			}
		}

		// Also add significant single words (capitalized or domain terms)
		if !seen[word] {
			seen[word] = true
			// Only add single words if they seem significant
			if isCapitalized(word) || isDomainTerm(word) {
				entities = append(entities, word)
			}
		}
	}

	// Limit to top 5 most relevant entities
	if len(entities) > 5 {
		entities = entities[:5]
	}

	return entities
}

// isCapitalized checks if a word appears to be capitalized (proper noun)
func isCapitalized(word string) bool {
	// This is called on lowercase words, check original
	return len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z'
}

// isDomainTerm checks if a word is a significant domain term
func isDomainTerm(word string) bool {
	domainTerms := map[string]bool{
		"funding": true, "research": true, "analysis": true, "system": true,
		"data": true, "model": true, "algorithm": true, "architecture": true,
		"design": true, "implementation": true, "performance": true, "security": true,
		"testing": true, "debugging": true, "optimization": true, "integration": true,
		"deployment": true, "science": true, "engineering": true, "technology": true,
		"physics": true, "chemistry": true, "biology": true, "mathematics": true,
		"economics": true, "policy": true, "strategy": true, "management": true,
		"prioritization": true, "allocation": true, "distribution": true,
	}
	return domainTerms[word]
}

// generateResearchSubproblems creates research-specific subproblems
func (pd *ProblemDecomposer) generateResearchSubproblems(problem string, entities []string, entityStr string) []*types.Subproblem {
	return []*types.Subproblem{
		{
			ID:          fmt.Sprintf("subproblem-%d-1", pd.counter),
			Description: fmt.Sprintf("Define research questions about %s", entityStr),
			Complexity:  "medium",
			Priority:    "critical",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-2", pd.counter),
			Description: fmt.Sprintf("Review existing literature and prior work on %s", entityStr),
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-3", pd.counter),
			Description: fmt.Sprintf("Design methodology to investigate %s", entityStr),
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-4", pd.counter),
			Description: fmt.Sprintf("Gather and analyze data related to %s", entityStr),
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-5", pd.counter),
			Description: fmt.Sprintf("Synthesize findings and draw conclusions about %s", entityStr),
			Complexity:  "medium",
			Priority:    "high",
			Status:      "pending",
		},
	}
}

// generateDebuggingSubproblems creates debugging-specific subproblems
func (pd *ProblemDecomposer) generateDebuggingSubproblems(problem string, entities []string, entityStr string) []*types.Subproblem {
	return []*types.Subproblem{
		{
			ID:          fmt.Sprintf("subproblem-%d-1", pd.counter),
			Description: fmt.Sprintf("Reproduce and isolate the issue in %s", entityStr),
			Complexity:  "medium",
			Priority:    "critical",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-2", pd.counter),
			Description: fmt.Sprintf("Gather diagnostic information about %s", entityStr),
			Complexity:  "medium",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-3", pd.counter),
			Description: fmt.Sprintf("Identify root cause of the %s issue", entityStr),
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-4", pd.counter),
			Description: fmt.Sprintf("Develop and implement fix for %s", entityStr),
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-5", pd.counter),
			Description: fmt.Sprintf("Verify fix and ensure no regression in %s", entityStr),
			Complexity:  "medium",
			Priority:    "high",
			Status:      "pending",
		},
	}
}

// generateArchitectureSubproblems creates architecture-specific subproblems
func (pd *ProblemDecomposer) generateArchitectureSubproblems(problem string, entities []string, entityStr string) []*types.Subproblem {
	return []*types.Subproblem{
		{
			ID:          fmt.Sprintf("subproblem-%d-1", pd.counter),
			Description: fmt.Sprintf("Analyze requirements and constraints for %s", entityStr),
			Complexity:  "medium",
			Priority:    "critical",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-2", pd.counter),
			Description: fmt.Sprintf("Identify architectural patterns applicable to %s", entityStr),
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-3", pd.counter),
			Description: fmt.Sprintf("Design component structure and interfaces for %s", entityStr),
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-4", pd.counter),
			Description: fmt.Sprintf("Evaluate trade-offs and risks in %s architecture", entityStr),
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-5", pd.counter),
			Description: fmt.Sprintf("Document and communicate the %s architecture", entityStr),
			Complexity:  "medium",
			Priority:    "medium",
			Status:      "pending",
		},
	}
}

// generateProofSubproblems creates proof/verification-specific subproblems
func (pd *ProblemDecomposer) generateProofSubproblems(problem string, entities []string, entityStr string) []*types.Subproblem {
	return []*types.Subproblem{
		{
			ID:          fmt.Sprintf("subproblem-%d-1", pd.counter),
			Description: fmt.Sprintf("State the theorem or claim about %s precisely", entityStr),
			Complexity:  "medium",
			Priority:    "critical",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-2", pd.counter),
			Description: fmt.Sprintf("Identify axioms, definitions, and lemmas needed for %s", entityStr),
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-3", pd.counter),
			Description: fmt.Sprintf("Construct proof strategy for %s", entityStr),
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-4", pd.counter),
			Description: fmt.Sprintf("Execute proof steps and verify each deduction about %s", entityStr),
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-5", pd.counter),
			Description: fmt.Sprintf("Review proof completeness and handle edge cases for %s", entityStr),
			Complexity:  "medium",
			Priority:    "high",
			Status:      "pending",
		},
	}
}

// generateGeneralSubproblems creates general problem-solving subproblems with entity context
func (pd *ProblemDecomposer) generateGeneralSubproblems(problem string, entities []string, entityStr string) []*types.Subproblem {
	return []*types.Subproblem{
		{
			ID:          fmt.Sprintf("subproblem-%d-1", pd.counter),
			Description: fmt.Sprintf("Analyze and define the scope of %s", entityStr),
			Complexity:  "low",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-2", pd.counter),
			Description: fmt.Sprintf("Gather information and resources about %s", entityStr),
			Complexity:  "medium",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-3", pd.counter),
			Description: fmt.Sprintf("Develop potential solutions for %s", entityStr),
			Complexity:  "high",
			Priority:    "high",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-4", pd.counter),
			Description: fmt.Sprintf("Evaluate and select best approach for %s", entityStr),
			Complexity:  "medium",
			Priority:    "medium",
			Status:      "pending",
		},
		{
			ID:          fmt.Sprintf("subproblem-%d-5", pd.counter),
			Description: fmt.Sprintf("Implement and test solution for %s", entityStr),
			Complexity:  "high",
			Priority:    "medium",
			Status:      "pending",
		},
	}
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
