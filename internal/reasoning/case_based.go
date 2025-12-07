// Package reasoning provides case-based reasoning capabilities.
// CBR involves solving new problems by retrieving and adapting solutions
// from similar past cases.
package reasoning

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"unified-thinking/internal/storage"
)

// CaseBasedReasoner performs case-based reasoning (retrieve, reuse, revise, retain)
type CaseBasedReasoner struct {
	storage    storage.Storage
	cases      map[string]*Case
	caseIndex  *CaseIndex
	analogical *AnalogicalReasoner
}

// NewCaseBasedReasoner creates a new case-based reasoner
func NewCaseBasedReasoner(store storage.Storage) *CaseBasedReasoner {
	cbr := &CaseBasedReasoner{
		storage:    store,
		cases:      make(map[string]*Case),
		caseIndex:  NewCaseIndex(),
		analogical: NewAnalogicalReasoner(),
	}
	// Pre-populate with default cases
	cbr.populateDefaultCases()
	return cbr
}

// populateDefaultCases adds common problem-solution cases to the library
func (cbr *CaseBasedReasoner) populateDefaultCases() {
	defaultCases := []*Case{
		{
			ID:     "case-race-condition-fix",
			Domain: "software-engineering",
			Tags:   []string{"concurrency", "debugging", "go"},
			Problem: &ProblemDescription{
				Description: "Fix race conditions in concurrent Go code",
				Context:     "Multiple goroutines accessing shared state",
				Goals:       []string{"Eliminate data races", "Maintain performance"},
			},
			Solution: &SolutionDescription{
				Description: "Use sync.Mutex or sync.RWMutex to protect shared state",
				Approach:    "synchronization",
				Steps:       []string{"Identify shared state", "Add mutex", "Lock before access", "Defer unlock"},
			},
			Outcome: &Outcome{
				Success:       true,
				Effectiveness: 0.95,
			},
			SuccessRate: 0.95,
		},
		{
			ID:     "case-ci-test-failure",
			Domain: "devops",
			Tags:   []string{"testing", "ci-cd", "debugging"},
			Problem: &ProblemDescription{
				Description: "Debug failing CI tests",
				Context:     "Tests pass locally but fail in CI",
				Goals:       []string{"Identify root cause", "Fix tests", "Prevent recurrence"},
			},
			Solution: &SolutionDescription{
				Description: "Check for environment differences, missing dependencies, and timing issues",
				Approach:    "systematic-debugging",
				Steps:       []string{"Compare environments", "Check dependencies", "Add retries for flaky tests", "Improve test isolation"},
			},
			Outcome: &Outcome{
				Success:       true,
				Effectiveness: 0.85,
			},
			SuccessRate: 0.85,
		},
		{
			ID:     "case-performance-optimization",
			Domain: "software-engineering",
			Tags:   []string{"performance", "optimization"},
			Problem: &ProblemDescription{
				Description: "Optimize slow application performance",
				Context:     "Application response time degraded",
				Goals:       []string{"Identify bottlenecks", "Improve response time"},
			},
			Solution: &SolutionDescription{
				Description: "Profile, identify hotspots, optimize critical paths",
				Approach:    "data-driven-optimization",
				Steps:       []string{"Profile application", "Identify bottlenecks", "Optimize hot paths", "Add caching", "Measure improvement"},
			},
			Outcome: &Outcome{
				Success:       true,
				Effectiveness: 0.9,
			},
			SuccessRate: 0.9,
		},
	}

	for _, c := range defaultCases {
		c.CreatedAt = time.Now()
		c.UpdatedAt = time.Now()
		cbr.caseIndex.indexCase(c)
		cbr.cases[c.ID] = c
	}
}

// Case represents a past problem-solution pair
type Case struct {
	ID            string
	Problem       *ProblemDescription
	Solution      *SolutionDescription
	Outcome       *Outcome
	Domain        string
	Tags          []string
	Applicability float64 // How applicable this case is (0-1)
	SuccessRate   float64 // Historical success rate (0-1)
	UsageCount    int
	LastUsed      time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Metadata      map[string]interface{}
}

// ProblemDescription describes the problem in a case
type ProblemDescription struct {
	Description string
	Context     string
	Constraints []string
	Goals       []string
	Features    map[string]interface{} // Feature vector for similarity matching
}

// SolutionDescription describes the solution in a case
type SolutionDescription struct {
	Description string
	Approach    string
	Steps       []string
	Rationale   string
	Assumptions []string
	Resources   []string
}

// Outcome describes the result of applying a solution
type Outcome struct {
	Success        bool
	Effectiveness  float64 // 0-1 score
	TimeToSolve    time.Duration
	CostIncurred   float64
	LessonsLearned []string
	FailureReasons []string // If not successful
}

// CaseIndex provides fast retrieval of similar cases
type CaseIndex struct {
	byDomain   map[string][]string // domain -> case IDs
	byTags     map[string][]string // tag -> case IDs
	byFeatures map[string][]string // feature key -> case IDs
}

// NewCaseIndex creates a new case index
func NewCaseIndex() *CaseIndex {
	return &CaseIndex{
		byDomain:   make(map[string][]string),
		byTags:     make(map[string][]string),
		byFeatures: make(map[string][]string),
	}
}

// RetrieveRequest contains parameters for case retrieval
type RetrieveRequest struct {
	Problem         *ProblemDescription
	Domain          string
	MaxCases        int
	MinSimilarity   float64
	RequireSuccess  bool
	PreferRecent    bool
	WeightBySuccess bool
}

// AdaptationStrategy specifies how to adapt a solution
type AdaptationStrategy string

const (
	AdaptDirect     AdaptationStrategy = "direct"     // Use solution as-is
	AdaptSubstitute AdaptationStrategy = "substitute" // Replace specific elements
	AdaptTransform  AdaptationStrategy = "transform"  // Transform solution structure
	AdaptCombine    AdaptationStrategy = "combine"    // Combine multiple cases
)

// CBRCycle represents the 4Rs of case-based reasoning
type CBRCycle struct {
	Retrieved *RetrievalResult
	Reused    *ReuseResult
	Revised   *RevisionResult
	Retained  bool
}

// RetrievalResult contains retrieved cases and their similarities
type RetrievalResult struct {
	Cases     []*SimilarCase
	Query     *ProblemDescription
	Retrieved int
	TotalTime time.Duration
}

// SimilarCase is a case with its similarity score
type SimilarCase struct {
	Case       *Case
	Similarity float64
	Rationale  string
}

// ReuseResult contains the adapted solution
type ReuseResult struct {
	OriginalCase    *Case
	AdaptedSolution *SolutionDescription
	Strategy        AdaptationStrategy
	Confidence      float64
	Adaptations     []string // List of adaptations made
}

// RevisionResult contains the revised solution after evaluation
type RevisionResult struct {
	RevisedSolution *SolutionDescription
	Changes         []string
	Confidence      float64
}

// StoreCase adds a new case to the library
func (cbr *CaseBasedReasoner) StoreCase(ctx context.Context, c *Case) error {
	now := time.Now()

	if c.ID == "" {
		c.ID = fmt.Sprintf("case-%d", now.UnixNano())
	}

	c.CreatedAt = now
	c.UpdatedAt = now

	// Index the case
	cbr.caseIndex.indexCase(c)

	// Store in memory
	cbr.cases[c.ID] = c

	return nil
}

// Retrieve finds similar cases (Step 1 of CBR cycle)
func (cbr *CaseBasedReasoner) Retrieve(ctx context.Context, req *RetrieveRequest) (*RetrievalResult, error) {
	startTime := time.Now()

	if req.Problem == nil {
		return nil, fmt.Errorf("problem description required")
	}

	// Get candidate cases
	candidates := cbr.getCandidateCases(req)

	// Calculate similarity for each candidate
	similarCases := make([]*SimilarCase, 0)
	for _, c := range candidates {
		similarity := cbr.calculateSimilarity(req.Problem, c.Problem)

		// Apply filters
		if similarity < req.MinSimilarity {
			continue
		}
		if req.RequireSuccess && !c.Outcome.Success {
			continue
		}

		rationale := cbr.explainSimilarity(req.Problem, c.Problem, similarity)

		similarCases = append(similarCases, &SimilarCase{
			Case:       c,
			Similarity: similarity,
			Rationale:  rationale,
		})
	}

	// Sort by similarity (descending)
	sort.Slice(similarCases, func(i, j int) bool {
		// If weight by success, factor in success rate
		if req.WeightBySuccess {
			scoreI := similarCases[i].Similarity * similarCases[i].Case.SuccessRate
			scoreJ := similarCases[j].Similarity * similarCases[j].Case.SuccessRate
			return scoreI > scoreJ
		}
		return similarCases[i].Similarity > similarCases[j].Similarity
	})

	// Limit to max cases
	if req.MaxCases > 0 && len(similarCases) > req.MaxCases {
		similarCases = similarCases[:req.MaxCases]
	}

	return &RetrievalResult{
		Cases:     similarCases,
		Query:     req.Problem,
		Retrieved: len(similarCases),
		TotalTime: time.Since(startTime),
	}, nil
}

// Reuse adapts a solution from a similar case (Step 2 of CBR cycle)
func (cbr *CaseBasedReasoner) Reuse(ctx context.Context, similarCase *SimilarCase, targetProblem *ProblemDescription, strategy AdaptationStrategy) (*ReuseResult, error) {
	if similarCase == nil || targetProblem == nil {
		return nil, fmt.Errorf("similar case and target problem required")
	}

	adaptedSolution := cbr.deepCopySolution(similarCase.Case.Solution)
	adaptations := make([]string, 0)
	confidence := similarCase.Similarity

	switch strategy {
	case AdaptDirect:
		// Use solution as-is
		adaptations = append(adaptations, "Applied solution directly without modification")

	case AdaptSubstitute:
		// Find and replace specific elements
		substitutions := cbr.identifySubstitutions(similarCase.Case.Problem, targetProblem)
		for old, new := range substitutions {
			adaptedSolution.Description = strings.ReplaceAll(adaptedSolution.Description, old, new)
			adaptations = append(adaptations, fmt.Sprintf("Substituted '%s' with '%s'", old, new))
		}
		confidence *= 0.9 // Slightly reduce confidence

	case AdaptTransform:
		// Transform solution structure
		transformed := cbr.transformSolution(adaptedSolution, targetProblem)
		adaptedSolution = transformed
		adaptations = append(adaptations, "Transformed solution structure to fit new context")
		confidence *= 0.8 // More reduction due to structural changes

	case AdaptCombine:
		// This would combine multiple cases, but requires multiple similar cases
		adaptations = append(adaptations, "Single case provided, using direct adaptation")
	}

	return &ReuseResult{
		OriginalCase:    similarCase.Case,
		AdaptedSolution: adaptedSolution,
		Strategy:        strategy,
		Confidence:      confidence,
		Adaptations:     adaptations,
	}, nil
}

// Revise refines the solution based on feedback (Step 3 of CBR cycle)
func (cbr *CaseBasedReasoner) Revise(ctx context.Context, reuseResult *ReuseResult, feedback string, success bool) (*RevisionResult, error) {
	revised := cbr.deepCopySolution(reuseResult.AdaptedSolution)
	changes := make([]string, 0)
	confidence := reuseResult.Confidence

	if !success {
		// Solution failed, try to improve
		changes = append(changes, fmt.Sprintf("Solution failed: %s", feedback))

		// Reduce confidence
		confidence *= 0.5

		// Add feedback to rationale
		revised.Rationale = fmt.Sprintf("%s\n\nRevision based on failure: %s", revised.Rationale, feedback)
		changes = append(changes, "Added failure analysis to rationale")
	} else {
		// Solution succeeded, reinforce
		changes = append(changes, "Solution validated successfully")
		confidence = math.Min(1.0, confidence*1.1)
	}

	return &RevisionResult{
		RevisedSolution: revised,
		Changes:         changes,
		Confidence:      confidence,
	}, nil
}

// Retain stores a new case or updates an existing one (Step 4 of CBR cycle)
func (cbr *CaseBasedReasoner) Retain(ctx context.Context, problem *ProblemDescription, solution *SolutionDescription, outcome *Outcome, domain string) (*Case, error) {
	newCase := &Case{
		ID:            fmt.Sprintf("case-%d", time.Now().UnixNano()),
		Problem:       problem,
		Solution:      solution,
		Outcome:       outcome,
		Domain:        domain,
		Tags:          cbr.extractTags(problem),
		Applicability: 0.8, // Default
		SuccessRate:   0.0,
		UsageCount:    0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	// Calculate initial success rate
	if outcome.Success {
		newCase.SuccessRate = outcome.Effectiveness
	}

	err := cbr.StoreCase(ctx, newCase)
	if err != nil {
		return nil, fmt.Errorf("failed to store case: %w", err)
	}

	return newCase, nil
}

// PerformCBRCycle executes the full 4R cycle
func (cbr *CaseBasedReasoner) PerformCBRCycle(ctx context.Context, problem *ProblemDescription, domain string) (*CBRCycle, error) {
	// Step 1: Retrieve
	retrieveReq := &RetrieveRequest{
		Problem:         problem,
		Domain:          domain,
		MaxCases:        5,
		MinSimilarity:   0.3, // Lower threshold for better recall
		RequireSuccess:  true,
		WeightBySuccess: true,
	}

	retrieved, err := cbr.Retrieve(ctx, retrieveReq)
	if err != nil {
		return nil, fmt.Errorf("retrieve failed: %w", err)
	}

	if len(retrieved.Cases) == 0 {
		return nil, fmt.Errorf("no similar cases found")
	}

	// Step 2: Reuse (use most similar case)
	mostSimilar := retrieved.Cases[0]
	reused, err := cbr.Reuse(ctx, mostSimilar, problem, AdaptSubstitute)
	if err != nil {
		return nil, fmt.Errorf("reuse failed: %w", err)
	}

	// Steps 3 & 4 (Revise & Retain) would happen after solution is tried
	// For now, return the cycle state

	return &CBRCycle{
		Retrieved: retrieved,
		Reused:    reused,
		Revised:   nil, // Requires feedback
		Retained:  false,
	}, nil
}

// Helper methods

func (cbr *CaseBasedReasoner) getCandidateCases(req *RetrieveRequest) []*Case {
	candidates := make([]*Case, 0)

	// If domain specified, use domain index
	if req.Domain != "" {
		caseIDs := cbr.caseIndex.byDomain[req.Domain]
		for _, id := range caseIDs {
			if c, exists := cbr.cases[id]; exists {
				candidates = append(candidates, c)
			}
		}
	} else {
		// Use all cases
		for _, c := range cbr.cases {
			candidates = append(candidates, c)
		}
	}

	return candidates
}

func (cbr *CaseBasedReasoner) calculateSimilarity(p1, p2 *ProblemDescription) float64 {
	// Multi-factor similarity calculation

	// 1. Textual similarity (description)
	textSim := cbr.textSimilarity(p1.Description, p2.Description)

	// 2. Context similarity
	contextSim := cbr.textSimilarity(p1.Context, p2.Context)

	// 3. Goal similarity
	goalSim := cbr.setOverlap(p1.Goals, p2.Goals)

	// 4. Constraint similarity
	constraintSim := cbr.setOverlap(p1.Constraints, p2.Constraints)

	// 5. Feature similarity
	featureSim := cbr.featureSimilarity(p1.Features, p2.Features)

	// Weighted combination
	similarity := (textSim*0.3 + contextSim*0.2 + goalSim*0.2 + constraintSim*0.15 + featureSim*0.15)

	return similarity
}

func (cbr *CaseBasedReasoner) textSimilarity(text1, text2 string) float64 {
	// Simple word overlap similarity (Jaccard)
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))

	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, w := range words1 {
		set1[w] = true
	}
	for _, w := range words2 {
		set2[w] = true
	}

	intersection := 0
	for w := range set1 {
		if set2[w] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

func (cbr *CaseBasedReasoner) setOverlap(set1, set2 []string) float64 {
	if len(set1) == 0 && len(set2) == 0 {
		return 1.0
	}
	if len(set1) == 0 || len(set2) == 0 {
		return 0.0
	}

	map1 := make(map[string]bool)
	for _, item := range set1 {
		map1[strings.ToLower(item)] = true
	}

	intersection := 0
	for _, item := range set2 {
		if map1[strings.ToLower(item)] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection
	return float64(intersection) / float64(union)
}

func (cbr *CaseBasedReasoner) featureSimilarity(f1, f2 map[string]interface{}) float64 {
	if len(f1) == 0 && len(f2) == 0 {
		return 1.0
	}
	if len(f1) == 0 || len(f2) == 0 {
		return 0.0
	}

	// Count matching features
	matches := 0
	for key := range f1 {
		if _, exists := f2[key]; exists {
			matches++
		}
	}

	totalKeys := len(f1) + len(f2) - matches
	return float64(matches) / float64(totalKeys)
}

func (cbr *CaseBasedReasoner) explainSimilarity(p1, p2 *ProblemDescription, similarity float64) string {
	reasons := make([]string, 0)

	textSim := cbr.textSimilarity(p1.Description, p2.Description)
	if textSim > 0.7 {
		reasons = append(reasons, "highly similar problem descriptions")
	}

	goalSim := cbr.setOverlap(p1.Goals, p2.Goals)
	if goalSim > 0.5 {
		reasons = append(reasons, "overlapping goals")
	}

	if len(reasons) == 0 {
		return fmt.Sprintf("Overall similarity: %.2f", similarity)
	}

	return strings.Join(reasons, ", ")
}

func (cbr *CaseBasedReasoner) identifySubstitutions(oldProblem, newProblem *ProblemDescription) map[string]string {
	substitutions := make(map[string]string)

	// Simple heuristic: find words in old problem not in new, and vice versa
	oldWords := strings.Fields(strings.ToLower(oldProblem.Description))
	newWords := strings.Fields(strings.ToLower(newProblem.Description))

	oldSet := make(map[string]bool)
	newSet := make(map[string]bool)

	for _, w := range oldWords {
		oldSet[w] = true
	}
	for _, w := range newWords {
		newSet[w] = true
	}

	// Find unique words (potential substitutions)
	// This is a simplified version - real CBR would use more sophisticated mapping
	for w := range oldSet {
		if !newSet[w] && len(w) > 3 { // Skip short words
			for nw := range newSet {
				if !oldSet[nw] && len(nw) > 3 {
					substitutions[w] = nw
					break
				}
			}
		}
	}

	return substitutions
}

func (cbr *CaseBasedReasoner) transformSolution(sol *SolutionDescription, targetProblem *ProblemDescription) *SolutionDescription {
	transformed := cbr.deepCopySolution(sol)

	// Add context-specific adaptation note
	transformed.Rationale = fmt.Sprintf("Transformed for context: %s\n\n%s", targetProblem.Context, transformed.Rationale)

	return transformed
}

func (cbr *CaseBasedReasoner) extractTags(problem *ProblemDescription) []string {
	tags := make([]string, 0)

	// Extract from description
	words := strings.Fields(strings.ToLower(problem.Description))
	for _, w := range words {
		if len(w) > 5 { // Longer words are likely meaningful tags
			tags = append(tags, w)
		}
	}

	// Limit to 10 tags
	if len(tags) > 10 {
		tags = tags[:10]
	}

	return tags
}

func (cbr *CaseBasedReasoner) deepCopySolution(sol *SolutionDescription) *SolutionDescription {
	copied := &SolutionDescription{
		Description: sol.Description,
		Approach:    sol.Approach,
		Rationale:   sol.Rationale,
		Steps:       make([]string, len(sol.Steps)),
		Assumptions: make([]string, len(sol.Assumptions)),
		Resources:   make([]string, len(sol.Resources)),
	}

	copy(copied.Steps, sol.Steps)
	copy(copied.Assumptions, sol.Assumptions)
	copy(copied.Resources, sol.Resources)

	return copied
}

func (ci *CaseIndex) indexCase(c *Case) {
	// Index by domain
	ci.byDomain[c.Domain] = append(ci.byDomain[c.Domain], c.ID)

	// Index by tags
	for _, tag := range c.Tags {
		ci.byTags[tag] = append(ci.byTags[tag], c.ID)
	}

	// Index by features
	for key := range c.Problem.Features {
		ci.byFeatures[key] = append(ci.byFeatures[key], c.ID)
	}
}
