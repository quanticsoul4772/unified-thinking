// Package reasoning provides abductive reasoning capabilities.
// Abductive reasoning is "inference to the best explanation" - generating and evaluating
// hypotheses that best explain a set of observations.
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

// AbductiveReasoner performs inference to the best explanation
type AbductiveReasoner struct {
	storage storage.Storage
}

// NewAbductiveReasoner creates a new abductive reasoner
func NewAbductiveReasoner(store storage.Storage) *AbductiveReasoner {
	return &AbductiveReasoner{
		storage: store,
	}
}

// Observation represents a fact that needs explaining
type Observation struct {
	ID          string
	Description string
	Confidence  float64 // How certain we are about this observation (0-1)
	Timestamp   time.Time
	Context     string
	Metadata    map[string]interface{}
}

// Hypothesis represents a possible explanation
type Hypothesis struct {
	ID                   string
	Description          string
	Observations         []string // IDs of observations this explains
	ExplanatoryPower     float64  // How well it explains observations (0-1)
	Parsimony            float64  // Simplicity score (0-1, higher = simpler)
	PriorProbability     float64  // Prior plausibility (0-1)
	PosteriorProbability float64  // Updated probability after evidence
	Assumptions          []string // Required assumptions
	Predictions          []string // Testable predictions
	CompetingIDs         []string // IDs of competing hypotheses
	Status               HypothesisStatus
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Metadata             map[string]interface{}
}

// HypothesisStatus represents the state of a hypothesis
type HypothesisStatus string

const (
	StatusProposed  HypothesisStatus = "proposed"
	StatusEvaluated HypothesisStatus = "evaluated"
	StatusSupported HypothesisStatus = "supported"
	StatusRefuted   HypothesisStatus = "refuted"
	StatusRevised   HypothesisStatus = "revised"
)

// AbductiveInference represents the result of abductive reasoning
type AbductiveInference struct {
	ID               string
	Observations     []*Observation
	Hypotheses       []*Hypothesis
	BestHypothesis   *Hypothesis
	RankedHypotheses []*Hypothesis // Sorted by posterior probability
	InferenceMethod  string
	Confidence       float64 // Overall confidence in the inference
	Timestamp        time.Time
	Metadata         map[string]interface{}
}

// GenerateHypothesesRequest contains parameters for hypothesis generation
type GenerateHypothesesRequest struct {
	Observations    []*Observation
	MaxHypotheses   int     // Maximum number to generate
	MinParsimony    float64 // Minimum simplicity threshold
	RequireTestable bool    // Whether hypotheses must be testable
	Context         string
}

// EvaluateHypothesesRequest contains parameters for hypothesis evaluation
type EvaluateHypothesesRequest struct {
	Observations []*Observation
	Hypotheses   []*Hypothesis
	Method       EvaluationMethod
	Weights      *EvaluationWeights
}

// EvaluationMethod specifies how to evaluate hypotheses
type EvaluationMethod string

const (
	MethodBayesian    EvaluationMethod = "bayesian"    // Bayesian inference
	MethodParsimony   EvaluationMethod = "parsimony"   // Occam's Razor
	MethodExplanatory EvaluationMethod = "explanatory" // Explanatory power
	MethodCombined    EvaluationMethod = "combined"    // Weighted combination
)

// EvaluationWeights defines relative importance of evaluation criteria
type EvaluationWeights struct {
	ExplanatoryPower float64 // Default: 0.4
	Parsimony        float64 // Default: 0.3
	PriorProbability float64 // Default: 0.3
}

// DefaultEvaluationWeights returns standard weights
func DefaultEvaluationWeights() *EvaluationWeights {
	return &EvaluationWeights{
		ExplanatoryPower: 0.4,
		Parsimony:        0.3,
		PriorProbability: 0.3,
	}
}

// GenerateHypotheses creates possible explanations for observations
func (ar *AbductiveReasoner) GenerateHypotheses(ctx context.Context, req *GenerateHypothesesRequest) ([]*Hypothesis, error) {
	if len(req.Observations) == 0 {
		return nil, fmt.Errorf("no observations provided")
	}

	hypotheses := make([]*Hypothesis, 0)

	// Strategy 1: Single-cause hypothesis (one explanation for all observations)
	singleCause := ar.generateSingleCauseHypothesis(req.Observations)
	if singleCause != nil {
		hypotheses = append(hypotheses, singleCause)
	}

	// Strategy 2: Hidden variable hypothesis (confounding factor)
	hiddenVar := ar.generateHiddenVariableHypothesis(req.Observations)
	if hiddenVar != nil {
		hypotheses = append(hypotheses, hiddenVar)
	}

	// Strategy 3: Causal chain hypothesis (sequential causation)
	causalChain := ar.generateCausalChainHypothesis(req.Observations)
	if causalChain != nil {
		hypotheses = append(hypotheses, causalChain)
	}

	// Strategy 4: Multiple independent causes (different explanations for different observations)
	multipleCauses := ar.generateMultipleCauseHypotheses(req.Observations)
	hypotheses = append(hypotheses, multipleCauses...)

	// Strategy 5: Pattern-based hypotheses (if temporal patterns exist)
	patterns := ar.generatePatternHypotheses(req.Observations)
	hypotheses = append(hypotheses, patterns...)

	// Strategy 6: Feedback loop hypothesis (circular causation)
	if len(req.Observations) >= 3 {
		feedbackLoop := ar.generateFeedbackLoopHypothesis(req.Observations)
		if feedbackLoop != nil {
			hypotheses = append(hypotheses, feedbackLoop)
		}
	}

	// Filter by parsimony threshold
	if req.MinParsimony > 0 {
		filtered := make([]*Hypothesis, 0)
		for _, h := range hypotheses {
			if h.Parsimony >= req.MinParsimony {
				filtered = append(filtered, h)
			}
		}
		hypotheses = filtered
	}

	// Limit to max hypotheses
	if req.MaxHypotheses > 0 && len(hypotheses) > req.MaxHypotheses {
		hypotheses = hypotheses[:req.MaxHypotheses]
	}

	return hypotheses, nil
}

// generateHiddenVariableHypothesis creates a hypothesis about a confounding variable
func (ar *AbductiveReasoner) generateHiddenVariableHypothesis(observations []*Observation) *Hypothesis {
	obsIDs := make([]string, len(observations))
	for i, obs := range observations {
		obsIDs[i] = obs.ID
	}

	commonThemes := ar.findCommonThemes(observations)
	domain := ar.detectObservationDomain(observations)
	themesStr := strings.Join(commonThemes, " and ")

	var description string
	switch domain {
	case "science", "research":
		description = fmt.Sprintf("A hidden variable or confounding factor may be driving all observed relationships with %s, rather than direct causation between the observed variables", themesStr)
	case "economic", "financial":
		description = fmt.Sprintf("An unobserved economic factor (such as policy changes, market sentiment, or external shocks) may explain the apparent relationship between %s", themesStr)
	case "technical", "engineering":
		description = fmt.Sprintf("A shared dependency or infrastructure component may be the root cause affecting all observations related to %s", themesStr)
	default:
		description = fmt.Sprintf("A hidden confounding variable may explain the apparent relationships between %s without direct causation", themesStr)
	}

	predictions := []string{
		"Controlling for the hidden variable should weaken or eliminate the observed relationships",
		"Identifying and measuring the confounding factor would reveal spurious correlations",
		"The hidden variable should correlate with all observed variables",
	}

	return &Hypothesis{
		ID:               fmt.Sprintf("hyp-hidden-%d", time.Now().UnixNano()),
		Description:      description,
		Observations:     obsIDs,
		ExplanatoryPower: 0.6,
		Parsimony:        0.7, // Medium parsimony (introduces hidden variable)
		PriorProbability: 0.4,
		Assumptions:      []string{"Correlation is not causation", "An unmeasured variable exists that affects multiple observed variables"},
		Predictions:      predictions,
		Status:           StatusProposed,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         map[string]interface{}{"hypothesis_type": "hidden_variable"},
	}
}

// generateCausalChainHypothesis creates a hypothesis about sequential causation
func (ar *AbductiveReasoner) generateCausalChainHypothesis(observations []*Observation) *Hypothesis {
	obsIDs := make([]string, len(observations))
	for i, obs := range observations {
		obsIDs[i] = obs.ID
	}

	commonThemes := ar.findCommonThemes(observations)
	domain := ar.detectObservationDomain(observations)
	themesStr := strings.Join(commonThemes, " and ")

	var description string
	switch domain {
	case "science", "research":
		description = fmt.Sprintf("These observations may be connected through a causal chain where %s triggers a cascade of effects, with each observation representing a step in the sequence", themesStr)
	case "economic", "financial":
		description = fmt.Sprintf("A chain of economic effects starting with %s may propagate through interconnected markets or systems, producing the observed outcomes", themesStr)
	case "technical", "engineering":
		description = fmt.Sprintf("A failure or change in %s may have cascaded through dependent components, causing each observed effect in sequence", themesStr)
	default:
		description = fmt.Sprintf("The observations may be connected through a causal chain where %s initiates a sequence of downstream effects", themesStr)
	}

	predictions := []string{
		"Earlier observations in the chain should precede later ones",
		"Interrupting the chain at any point should prevent downstream effects",
		"The magnitude of effects may diminish along the chain",
	}

	return &Hypothesis{
		ID:               fmt.Sprintf("hyp-chain-%d", time.Now().UnixNano()),
		Description:      description,
		Observations:     obsIDs,
		ExplanatoryPower: 0.65,
		Parsimony:        0.6, // Lower parsimony (implies ordering)
		PriorProbability: 0.45,
		Assumptions:      []string{"Observations can be ordered causally", "Effects propagate from earlier to later in the chain"},
		Predictions:      predictions,
		Status:           StatusProposed,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         map[string]interface{}{"hypothesis_type": "causal_chain"},
	}
}

// generateFeedbackLoopHypothesis creates a hypothesis about circular causation
func (ar *AbductiveReasoner) generateFeedbackLoopHypothesis(observations []*Observation) *Hypothesis {
	obsIDs := make([]string, len(observations))
	for i, obs := range observations {
		obsIDs[i] = obs.ID
	}

	commonThemes := ar.findCommonThemes(observations)
	domain := ar.detectObservationDomain(observations)
	themesStr := strings.Join(commonThemes, " and ")

	var description string
	switch domain {
	case "science", "research":
		description = fmt.Sprintf("The observed phenomena may form a feedback loop where %s both cause and are caused by each other, creating a self-reinforcing or self-limiting cycle", themesStr)
	case "economic", "financial":
		description = fmt.Sprintf("A positive or negative feedback loop involving %s may explain the observed patterns, where effects reinforce or dampen their own causes", themesStr)
	case "technical", "engineering":
		description = fmt.Sprintf("System behavior related to %s may involve feedback mechanisms where outputs influence inputs, creating dynamic equilibrium or oscillation", themesStr)
	default:
		description = fmt.Sprintf("A feedback loop involving %s may explain these observations, where effects cyclically influence their causes", themesStr)
	}

	predictions := []string{
		"The system may exhibit oscillation or convergence to equilibrium",
		"Breaking the feedback loop should stabilize the system",
		"Small changes may be amplified (positive feedback) or dampened (negative feedback)",
	}

	return &Hypothesis{
		ID:               fmt.Sprintf("hyp-feedback-%d", time.Now().UnixNano()),
		Description:      description,
		Observations:     obsIDs,
		ExplanatoryPower: 0.55,
		Parsimony:        0.5, // Lower parsimony (complex circular causation)
		PriorProbability: 0.35,
		Assumptions:      []string{"Causal relationships can be bidirectional", "The system exhibits dynamic behavior"},
		Predictions:      predictions,
		Status:           StatusProposed,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         map[string]interface{}{"hypothesis_type": "feedback_loop"},
	}
}

// generateSingleCauseHypothesis creates a hypothesis that explains all observations with one cause
func (ar *AbductiveReasoner) generateSingleCauseHypothesis(observations []*Observation) *Hypothesis {
	obsIDs := make([]string, len(observations))
	for i, obs := range observations {
		obsIDs[i] = obs.ID
	}

	// Find common themes in observation descriptions
	commonThemes := ar.findCommonThemes(observations)

	// Generate explanatory hypothesis instead of just summarizing themes
	description := ar.generateExplanatoryDescription(commonThemes, observations, "single")

	// Generate testable predictions based on the hypothesis
	predictions := ar.generatePredictions(commonThemes, observations)

	return &Hypothesis{
		ID:               fmt.Sprintf("hyp-single-%d", time.Now().UnixNano()),
		Description:      description,
		Observations:     obsIDs,
		ExplanatoryPower: 0.7, // Default, will be refined
		Parsimony:        0.9, // High parsimony (single cause)
		PriorProbability: 0.5,
		Assumptions:      []string{"All observations share a common underlying mechanism"},
		Predictions:      predictions,
		Status:           StatusProposed,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         make(map[string]interface{}),
	}
}

// generateExplanatoryDescription creates a meaningful explanation based on themes and context
func (ar *AbductiveReasoner) generateExplanatoryDescription(themes []string, observations []*Observation, hypType string) string {
	if len(themes) == 0 {
		return "An unidentified underlying mechanism connects these observations"
	}

	themesStr := strings.Join(themes, " and ")

	// Detect domain from observations
	domain := ar.detectObservationDomain(observations)

	// Generate domain-specific explanatory hypothesis
	switch domain {
	case "science", "research":
		switch hypType {
		case "single":
			return fmt.Sprintf("These observations may be explained by a fundamental property or process related to %s that manifests in multiple measurable ways", themesStr)
		case "multi":
			return fmt.Sprintf("Multiple interacting mechanisms involving %s may each contribute to different aspects of the observed phenomena", themesStr)
		case "pattern":
			return fmt.Sprintf("A systematic relationship between %s creates a predictable pattern across these observations", themesStr)
		}
	case "economic", "financial":
		switch hypType {
		case "single":
			return fmt.Sprintf("Market dynamics or economic pressures related to %s may drive all observed outcomes through interconnected incentive structures", themesStr)
		case "multi":
			return fmt.Sprintf("Multiple economic factors involving %s interact to produce the observed market behaviors through different channels", themesStr)
		case "pattern":
			return fmt.Sprintf("Cyclical or structural economic patterns related to %s explain the temporal distribution of observations", themesStr)
		}
	case "social", "behavioral":
		switch hypType {
		case "single":
			return fmt.Sprintf("Social dynamics or behavioral patterns centered on %s may explain these observations through shared incentives or cultural factors", themesStr)
		case "multi":
			return fmt.Sprintf("Different stakeholder motivations regarding %s produce varied but related outcomes across the observations", themesStr)
		case "pattern":
			return fmt.Sprintf("Emergent social patterns involving %s create predictable outcomes when similar conditions arise", themesStr)
		}
	case "technical", "engineering":
		switch hypType {
		case "single":
			return fmt.Sprintf("A systemic technical factor related to %s may be the root cause affecting multiple components or behaviors", themesStr)
		case "multi":
			return fmt.Sprintf("Multiple technical dependencies involving %s each contribute to different failure modes or behaviors", themesStr)
		case "pattern":
			return fmt.Sprintf("Technical constraints or architectural patterns related to %s produce consistent behaviors under similar conditions", themesStr)
		}
	}

	// Generic but still explanatory fallback
	switch hypType {
	case "single":
		return fmt.Sprintf("A common underlying mechanism related to %s may explain these observations by creating similar conditions or outcomes", themesStr)
	case "multi":
		return fmt.Sprintf("Multiple independent but related factors involving %s each contribute to different aspects of the observed phenomena", themesStr)
	case "pattern":
		return fmt.Sprintf("A systematic relationship or recurring pattern involving %s connects these observations", themesStr)
	default:
		return fmt.Sprintf("The observations may be connected through mechanisms related to %s", themesStr)
	}
}

// detectObservationDomain analyzes observations to determine the domain
func (ar *AbductiveReasoner) detectObservationDomain(observations []*Observation) string {
	domainKeywords := map[string][]string{
		"science":     {"research", "experiment", "hypothesis", "data", "measure", "observe", "theory", "physics", "chemistry", "biology", "particle", "quantum", "molecular"},
		"research":    {"study", "findings", "publication", "peer", "academic", "scientist", "laboratory", "funding", "grant"},
		"economic":    {"market", "price", "cost", "revenue", "profit", "investment", "financial", "economy", "budget", "capital", "stock", "trade"},
		"financial":   {"money", "fund", "bank", "loan", "interest", "debt", "credit", "asset", "portfolio"},
		"social":      {"community", "society", "culture", "behavior", "population", "demographic", "group", "people", "public", "citizen"},
		"behavioral":  {"motivation", "incentive", "decision", "choice", "preference", "attitude", "perception", "belief"},
		"technical":   {"system", "software", "hardware", "code", "bug", "error", "performance", "architecture", "database", "server", "network"},
		"engineering": {"design", "build", "implement", "deploy", "scale", "optimize", "infrastructure", "component"},
	}

	// Count matches for each domain
	domainCounts := make(map[string]int)
	for _, obs := range observations {
		textLower := strings.ToLower(obs.Description + " " + obs.Context)
		for domain, keywords := range domainKeywords {
			for _, kw := range keywords {
				if strings.Contains(textLower, kw) {
					domainCounts[domain]++
				}
			}
		}
	}

	// Find the domain with highest count
	maxDomain := ""
	maxCount := 0
	for domain, count := range domainCounts {
		if count > maxCount {
			maxCount = count
			maxDomain = domain
		}
	}

	if maxCount >= 2 {
		return maxDomain
	}
	return ""
}

// generatePredictions creates testable predictions from themes
func (ar *AbductiveReasoner) generatePredictions(themes []string, observations []*Observation) []string {
	predictions := make([]string, 0)

	if len(themes) == 0 {
		return []string{"Additional observations following similar conditions should show related patterns"}
	}

	domain := ar.detectObservationDomain(observations)
	themesStr := strings.Join(themes, " and ")

	switch domain {
	case "science", "research":
		predictions = append(predictions,
			fmt.Sprintf("Controlled experiments varying %s should show correlated effects", themesStr),
			"Removing or altering the proposed mechanism should change the observed outcomes",
		)
	case "economic", "financial":
		predictions = append(predictions,
			fmt.Sprintf("Changes in %s-related factors should predict corresponding market movements", themesStr),
			"Historical data should show consistent patterns when similar conditions existed",
		)
	case "technical", "engineering":
		predictions = append(predictions,
			fmt.Sprintf("Modifying the %s component should affect the observed behavior", themesStr),
			"Similar systems with the same architecture should exhibit the same patterns",
		)
	default:
		predictions = append(predictions,
			fmt.Sprintf("Future observations involving %s should follow the same pattern", themesStr),
			"Intervening on the proposed cause should change subsequent observations",
		)
	}

	return predictions
}

// generateMultipleCauseHypotheses creates hypotheses with multiple independent causes
func (ar *AbductiveReasoner) generateMultipleCauseHypotheses(observations []*Observation) []*Hypothesis {
	hypotheses := make([]*Hypothesis, 0)

	// Group observations by similarity
	groups := ar.groupObservations(observations)

	if len(groups) <= 1 {
		return hypotheses
	}

	// Create hypothesis for each grouping
	obsIDs := make([]string, len(observations))
	for i, obs := range observations {
		obsIDs[i] = obs.ID
	}

	// Find common themes for explanatory description
	commonThemes := ar.findCommonThemes(observations)
	description := ar.generateExplanatoryDescription(commonThemes, observations, "multi")

	// Generate predictions
	predictions := ar.generatePredictions(commonThemes, observations)
	predictions = append(predictions, fmt.Sprintf("Isolating each of the %d causal factors should show independent effects", len(groups)))

	h := &Hypothesis{
		ID:               fmt.Sprintf("hyp-multi-%d", time.Now().UnixNano()),
		Description:      description,
		Observations:     obsIDs,
		ExplanatoryPower: 0.6,
		Parsimony:        0.5, // Lower parsimony (multiple causes)
		PriorProbability: 0.4,
		Assumptions:      []string{fmt.Sprintf("At least %d distinct causal factors are operating", len(groups)), "These factors may interact but have independent origins"},
		Predictions:      predictions,
		Status:           StatusProposed,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Metadata:         map[string]interface{}{"group_count": len(groups)},
	}

	hypotheses = append(hypotheses, h)

	return hypotheses
}

// generatePatternHypotheses identifies patterns in observations
func (ar *AbductiveReasoner) generatePatternHypotheses(observations []*Observation) []*Hypothesis {
	hypotheses := make([]*Hypothesis, 0)

	// Temporal pattern: check if observations follow a sequence
	if ar.hasTemporalPattern(observations) {
		obsIDs := make([]string, len(observations))
		for i, obs := range observations {
			obsIDs[i] = obs.ID
		}

		// Find common themes for explanatory description
		commonThemes := ar.findCommonThemes(observations)
		description := ar.generateExplanatoryDescription(commonThemes, observations, "pattern")

		// Generate pattern-specific predictions
		predictions := ar.generatePredictions(commonThemes, observations)
		predictions = append(predictions, "The temporal pattern suggests future observations will follow a similar sequence")

		h := &Hypothesis{
			ID:               fmt.Sprintf("hyp-pattern-%d", time.Now().UnixNano()),
			Description:      description,
			Observations:     obsIDs,
			ExplanatoryPower: 0.65,
			Parsimony:        0.75,
			PriorProbability: 0.45,
			Assumptions:      []string{"An underlying cyclic or sequential process governs these observations", "The temporal regularity is not coincidental"},
			Predictions:      predictions,
			Status:           StatusProposed,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			Metadata:         make(map[string]interface{}),
		}

		hypotheses = append(hypotheses, h)
	}

	return hypotheses
}

// EvaluateHypotheses ranks hypotheses by their explanatory quality
func (ar *AbductiveReasoner) EvaluateHypotheses(ctx context.Context, req *EvaluateHypothesesRequest) ([]*Hypothesis, error) {
	if len(req.Hypotheses) == 0 {
		return nil, fmt.Errorf("no hypotheses to evaluate")
	}

	weights := req.Weights
	if weights == nil {
		weights = DefaultEvaluationWeights()
	}

	// Evaluate each hypothesis
	for _, h := range req.Hypotheses {
		ar.evaluateHypothesis(h, req.Observations, req.Method, weights)
	}

	// Sort by posterior probability (highest first)
	sort.Slice(req.Hypotheses, func(i, j int) bool {
		return req.Hypotheses[i].PosteriorProbability > req.Hypotheses[j].PosteriorProbability
	})

	return req.Hypotheses, nil
}

// evaluateHypothesis calculates scores for a single hypothesis
func (ar *AbductiveReasoner) evaluateHypothesis(h *Hypothesis, observations []*Observation, method EvaluationMethod, weights *EvaluationWeights) {
	// Calculate explanatory power
	h.ExplanatoryPower = ar.calculateExplanatoryPower(h, observations)

	// Calculate parsimony (simplicity)
	h.Parsimony = ar.calculateParsimony(h)

	// Calculate posterior probability based on method
	switch method {
	case MethodBayesian:
		h.PosteriorProbability = ar.calculateBayesianProbability(h, observations)
	case MethodParsimony:
		h.PosteriorProbability = h.Parsimony
	case MethodExplanatory:
		h.PosteriorProbability = h.ExplanatoryPower
	case MethodCombined:
		h.PosteriorProbability = weights.ExplanatoryPower*h.ExplanatoryPower +
			weights.Parsimony*h.Parsimony +
			weights.PriorProbability*h.PriorProbability
	default:
		h.PosteriorProbability = weights.ExplanatoryPower*h.ExplanatoryPower +
			weights.Parsimony*h.Parsimony +
			weights.PriorProbability*h.PriorProbability
	}

	h.Status = StatusEvaluated
	h.UpdatedAt = time.Now()
}

// calculateExplanatoryPower measures how well hypothesis explains observations
func (ar *AbductiveReasoner) calculateExplanatoryPower(h *Hypothesis, observations []*Observation) float64 {
	if len(h.Observations) == 0 {
		return 0.0
	}

	// Simple heuristic: proportion of observations explained * average observation confidence
	explained := float64(len(h.Observations))
	total := float64(len(observations))
	coverage := explained / total

	// Weight by observation confidence
	totalConfidence := 0.0
	for _, obs := range observations {
		for _, obsID := range h.Observations {
			if obs.ID == obsID {
				totalConfidence += obs.Confidence
			}
		}
	}
	avgConfidence := totalConfidence / explained

	return coverage * avgConfidence
}

// calculateParsimony measures hypothesis simplicity (Occam's Razor)
func (ar *AbductiveReasoner) calculateParsimony(h *Hypothesis) float64 {
	// Parsimony inversely related to number of assumptions
	assumptionPenalty := 1.0 / (1.0 + float64(len(h.Assumptions)))

	// Simpler descriptions score higher
	descriptionComplexity := float64(len(strings.Fields(h.Description)))
	descriptionScore := 1.0 / (1.0 + math.Log(1.0+descriptionComplexity))

	return (assumptionPenalty + descriptionScore) / 2.0
}

// calculateBayesianProbability applies Bayes' theorem
func (ar *AbductiveReasoner) calculateBayesianProbability(h *Hypothesis, observations []*Observation) float64 {
	// P(H|E) = P(E|H) * P(H) / P(E)
	// Where:
	// P(H|E) = posterior probability
	// P(E|H) = likelihood (how well hypothesis predicts evidence)
	// P(H) = prior probability
	// P(E) = marginal probability of evidence

	likelihood := h.ExplanatoryPower
	prior := h.PriorProbability

	// Simplified: assume P(E) = 0.5 (evidence is equally likely under any hypothesis)
	marginal := 0.5

	posterior := (likelihood * prior) / marginal

	// Clamp to [0, 1]
	if posterior > 1.0 {
		posterior = 1.0
	}
	if posterior < 0.0 {
		posterior = 0.0
	}

	return posterior
}

// PerformAbductiveInference executes full abductive reasoning workflow
func (ar *AbductiveReasoner) PerformAbductiveInference(ctx context.Context, observations []*Observation, maxHypotheses int) (*AbductiveInference, error) {
	// Generate hypotheses
	genReq := &GenerateHypothesesRequest{
		Observations:  observations,
		MaxHypotheses: maxHypotheses,
		MinParsimony:  0.3,
	}

	hypotheses, err := ar.GenerateHypotheses(ctx, genReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate hypotheses: %w", err)
	}

	// Evaluate hypotheses
	evalReq := &EvaluateHypothesesRequest{
		Observations: observations,
		Hypotheses:   hypotheses,
		Method:       MethodCombined,
		Weights:      DefaultEvaluationWeights(),
	}

	ranked, err := ar.EvaluateHypotheses(ctx, evalReq)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate hypotheses: %w", err)
	}

	// Select best hypothesis
	var best *Hypothesis
	if len(ranked) > 0 {
		best = ranked[0]
	}

	// Calculate overall confidence
	confidence := 0.0
	if best != nil {
		confidence = best.PosteriorProbability
	}

	inference := &AbductiveInference{
		ID:               fmt.Sprintf("inference-%d", time.Now().UnixNano()),
		Observations:     observations,
		Hypotheses:       hypotheses,
		BestHypothesis:   best,
		RankedHypotheses: ranked,
		InferenceMethod:  string(MethodCombined),
		Confidence:       confidence,
		Timestamp:        time.Now(),
		Metadata:         make(map[string]interface{}),
	}

	return inference, nil
}

// Helper methods

// findCommonThemes extracts meaningful common themes from observation descriptions
// Uses bigrams and filters stop words for better theme extraction
func (ar *AbductiveReasoner) findCommonThemes(observations []*Observation) []string {
	// Track both single words and bigrams
	wordCount := make(map[string]int)
	bigramCount := make(map[string]int)

	for _, obs := range observations {
		// Clean and tokenize
		text := strings.ToLower(obs.Description)
		// Remove punctuation
		text = strings.Map(func(r rune) rune {
			if r == '.' || r == ',' || r == ':' || r == ';' || r == '!' || r == '?' {
				return ' '
			}
			return r
		}, text)

		words := strings.Fields(text)

		// Count meaningful single words
		seenWords := make(map[string]bool)
		for _, word := range words {
			// Clean the word
			word = strings.Trim(word, "\"'()[]{}*-_")
			if len(word) < 3 { // Skip very short words
				continue
			}
			if ar.isStopWord(word) {
				continue
			}
			if !seenWords[word] {
				seenWords[word] = true
				wordCount[word]++
			}
		}

		// Count bigrams (pairs of meaningful words that are adjacent in original)
		for i := 0; i < len(words)-1; i++ {
			w1 := strings.Trim(words[i], "\"'()[]{}*-_")
			w2 := strings.Trim(words[i+1], "\"'()[]{}*-_")

			// Both words should be meaningful
			if len(w1) >= 3 && len(w2) >= 3 && !ar.isStopWord(w1) && !ar.isStopWord(w2) {
				bigram := w1 + " " + w2
				bigramCount[bigram]++
			}
		}
	}

	// Calculate threshold - at least 2 observations or 1/3 of total
	threshold := len(observations) / 3
	if threshold < 2 {
		threshold = 2
	}
	if len(observations) < 3 {
		threshold = 1
	}

	themes := make([]string, 0)

	// First add bigrams that appear frequently (more specific)
	for bigram, count := range bigramCount {
		if count >= threshold {
			themes = append(themes, bigram)
		}
	}

	// Then add single words that appear frequently
	for word, count := range wordCount {
		if count >= threshold && len(word) > 3 {
			// Don't add if already covered by a bigram
			alreadyCovered := false
			for _, theme := range themes {
				if strings.Contains(theme, word) {
					alreadyCovered = true
					break
				}
			}
			if !alreadyCovered {
				themes = append(themes, word)
			}
		}
	}

	// Sort themes by specificity (longer = more specific)
	sort.Slice(themes, func(i, j int) bool {
		return len(themes[i]) > len(themes[j])
	})

	// Limit to top themes
	if len(themes) > 5 {
		themes = themes[:5]
	}

	// If still no themes, try to extract domain-specific terms
	if len(themes) == 0 {
		themes = ar.extractDomainTerms(observations)
	}

	// Final fallback - extract the most unique/specific term
	if len(themes) == 0 {
		// Get the longest word that isn't a stop word
		var longest string
		for word, count := range wordCount {
			if count >= 1 && len(word) > len(longest) {
				longest = word
			}
		}
		if longest != "" {
			themes = []string{longest}
		} else {
			themes = []string{"underlying mechanism"}
		}
	}

	return themes
}

// extractDomainTerms tries to find domain-specific terminology in observations
func (ar *AbductiveReasoner) extractDomainTerms(observations []*Observation) []string {
	// Look for capitalized words (proper nouns, acronyms)
	properNouns := make(map[string]int)

	for _, obs := range observations {
		words := strings.Fields(obs.Description)
		for _, word := range words {
			// Check for capitalized words (not at start of sentence)
			if len(word) >= 2 {
				// Check if word is capitalized
				isCapitalized := word[0] >= 'A' && word[0] <= 'Z'
				isAllCaps := strings.ToUpper(word) == word && len(word) >= 2

				if isCapitalized || isAllCaps {
					cleanWord := strings.Trim(word, "\"'()[]{}*-_.,;:!?")
					if len(cleanWord) >= 2 && !ar.isStopWord(strings.ToLower(cleanWord)) {
						properNouns[cleanWord]++
					}
				}
			}
		}
	}

	terms := make([]string, 0)
	for term, count := range properNouns {
		if count >= 1 {
			terms = append(terms, term)
		}
	}

	// Sort by frequency
	sort.Slice(terms, func(i, j int) bool {
		return properNouns[terms[i]] > properNouns[terms[j]]
	})

	if len(terms) > 3 {
		terms = terms[:3]
	}

	return terms
}

// groupObservations clusters observations by similarity
func (ar *AbductiveReasoner) groupObservations(observations []*Observation) [][]string {
	// Simple grouping: observations with similar timestamps or context
	groups := make([][]string, 0)

	// Group by time proximity (within 1 hour)
	timeGroups := make(map[int64][]string)
	for _, obs := range observations {
		hourBucket := obs.Timestamp.Unix() / 3600
		timeGroups[hourBucket] = append(timeGroups[hourBucket], obs.ID)
	}

	for _, group := range timeGroups {
		if len(group) > 0 {
			groups = append(groups, group)
		}
	}

	if len(groups) == 0 {
		// Default: treat each observation as its own group
		for _, obs := range observations {
			groups = append(groups, []string{obs.ID})
		}
	}

	return groups
}

// hasTemporalPattern checks if observations follow a time sequence
func (ar *AbductiveReasoner) hasTemporalPattern(observations []*Observation) bool {
	if len(observations) < 2 {
		return false
	}

	// Sort by timestamp
	sorted := make([]*Observation, len(observations))
	copy(sorted, observations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})

	// Check if intervals are relatively consistent
	intervals := make([]time.Duration, 0)
	for i := 1; i < len(sorted); i++ {
		interval := sorted[i].Timestamp.Sub(sorted[i-1].Timestamp)
		intervals = append(intervals, interval)
	}

	// Calculate coefficient of variation
	if len(intervals) == 0 {
		return false
	}

	avg := time.Duration(0)
	for _, interval := range intervals {
		avg += interval
	}
	avg = avg / time.Duration(len(intervals))

	variance := time.Duration(0)
	for _, interval := range intervals {
		diff := interval - avg
		variance += diff * diff / time.Duration(len(intervals))
	}

	stdDev := time.Duration(math.Sqrt(float64(variance)))
	cv := float64(stdDev) / float64(avg)

	// Pattern exists if coefficient of variation < 0.5
	return cv < 0.5
}

// expandedStopWords contains a comprehensive set of common words that should be filtered out
// when extracting meaningful themes from observations
var expandedStopWords = map[string]bool{
	// Articles and determiners
	"the": true, "a": true, "an": true, "this": true, "that": true, "these": true, "those": true,
	"my": true, "your": true, "his": true, "her": true, "its": true, "our": true, "their": true,
	"some": true, "any": true, "no": true, "every": true, "each": true, "all": true, "both": true,

	// Pronouns
	"i": true, "you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
	"me": true, "him": true, "us": true, "them": true, "who": true, "whom": true, "what": true,
	"which": true, "whose": true, "myself": true, "yourself": true, "itself": true,

	// Common verbs (to be, to have, to do, modals)
	"is": true, "are": true, "was": true, "were": true, "be": true, "been": true, "being": true,
	"am": true, "have": true, "has": true, "had": true, "having": true, "do": true, "does": true,
	"did": true, "doing": true, "done": true, "will": true, "would": true, "shall": true,
	"should": true, "can": true, "could": true, "may": true, "might": true, "must": true,

	// Prepositions
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true, "with": true,
	"by": true, "from": true, "up": true, "about": true, "into": true, "over": true, "after": true,
	"beneath": true, "under": true, "above": true, "between": true, "among": true, "through": true,
	"during": true, "before": true, "behind": true, "below": true, "against": true,

	// Conjunctions
	"and": true, "or": true, "but": true, "nor": true, "so": true, "yet": true,
	"because": true, "although": true, "unless": true, "since": true, "while": true, "if": true,
	"then": true, "than": true, "when": true, "where": true, "whether": true,

	// Adverbs
	"very": true, "really": true, "quite": true, "rather": true, "too": true, "also": true,
	"just": true, "only": true, "even": true, "still": true, "already": true, "always": true,
	"never": true, "often": true, "sometimes": true, "usually": true, "now": true,
	"here": true, "there": true, "how": true, "why": true, "well": true,
	"more": true, "most": true, "less": true, "least": true,

	// Common adjectives (non-descriptive)
	"other": true, "another": true, "such": true, "own": true, "same": true, "different": true,
	"new": true, "old": true, "good": true, "bad": true, "great": true, "small": true,
	"large": true, "big": true, "little": true, "high": true, "low": true, "long": true,
	"first": true, "last": true, "next": true, "many": true, "much": true, "few": true,

	// Numbers and quantifiers
	"one": true, "two": true, "three": true, "four": true, "five": true, "six": true,
	"seven": true, "eight": true, "nine": true, "ten": true, "several": true, "none": true,

	// Common nouns that rarely carry specific meaning
	"thing": true, "things": true, "stuff": true, "way": true, "ways": true, "time": true,
	"times": true, "year": true, "years": true, "day": true, "days": true, "part": true,
	"parts": true, "place": true, "places": true, "case": true, "cases": true, "point": true,
	"points": true, "fact": true, "facts": true, "kind": true, "kinds": true, "type": true,
	"types": true, "form": true, "forms": true, "number": true, "numbers": true, "amount": true,
	"level": true, "levels": true, "area": true, "areas": true, "side": true, "sides": true,
	"end": true, "ends": true, "matter": true, "matters": true, "issue": true, "issues": true,
	"question": true, "questions": true, "problem": true, "problems": true, "result": true,
	"results": true, "example": true, "examples": true, "reason": true, "reasons": true,
	"idea": true, "ideas": true, "situation": true, "state": true, "states": true,

	// Domain-generic terms that don't add meaning
	"research": true, "study": true, "studies": true, "data": true, "information": true,
	"system": true, "systems": true, "process": true, "processes": true, "method": true,
	"methods": true, "approach": true, "approaches": true, "analysis": true, "work": true,
	"works": true, "effect": true, "effects": true, "impact": true, "change": true,
	"changes": true, "factor": true, "factors": true, "aspect": true, "aspects": true,

	// Common verbs and actions
	"make": true, "made": true, "get": true, "got": true, "go": true, "went": true,
	"come": true, "came": true, "take": true, "took": true, "give": true, "gave": true,
	"find": true, "found": true, "know": true, "knew": true, "think": true, "thought": true,
	"see": true, "saw": true, "want": true, "use": true, "used": true, "try": true,
	"need": true, "seem": true, "help": true, "show": true, "shows": true, "shown": true,
	"lead": true, "leads": true, "led": true, "cause": true, "causes": true, "caused": true,
}

// isStopWord checks if word is a common stop word using comprehensive list
func (ar *AbductiveReasoner) isStopWord(word string) bool {
	return expandedStopWords[word]
}
