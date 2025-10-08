// Package reasoning provides advanced reasoning capabilities.
//
// The causal reasoning module implements Pearl's causal inference framework,
// including proper graph surgery for interventions (do-calculus) to correctly
// distinguish correlation from causation.
package reasoning

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// CausalReasoner performs causal inference and counterfactual reasoning
type CausalReasoner struct {
	mu      sync.RWMutex
	graphs  map[string]*types.CausalGraph
	counter int
}

// NewCausalReasoner creates a new causal reasoner
func NewCausalReasoner() *CausalReasoner {
	return &CausalReasoner{
		graphs: make(map[string]*types.CausalGraph),
	}
}

// BuildCausalGraph constructs a causal graph from observations
func (cr *CausalReasoner) BuildCausalGraph(description string, observations []string) (*types.CausalGraph, error) {
	if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}
	if len(observations) == 0 {
		return nil, fmt.Errorf("at least one observation is required")
	}

	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.counter++

	// Extract variables from observations
	variables := cr.extractVariables(observations)

	// Identify causal links between variables
	links := cr.identifyCausalLinks(variables, observations)

	graph := &types.CausalGraph{
		ID:          fmt.Sprintf("causal-graph-%d", cr.counter),
		Description: description,
		Variables:   variables,
		Links:       links,
		Metadata:    map[string]interface{}{},
		CreatedAt:   time.Now(),
	}

	cr.graphs[graph.ID] = graph

	return graph, nil
}

// extractVariables identifies variables from observations
func (cr *CausalReasoner) extractVariables(observations []string) []*types.CausalVariable {
	variableMap := make(map[string]*types.CausalVariable)
	varCounter := 0

	// Patterns that indicate variables
	causalPatterns := []struct {
		pattern string
		before  string
		after   string
	}{
		{"increases", "cause", "effect"},
		{"decreases", "cause", "effect"},
		{"causes", "cause", "effect"},
		{"leads to", "cause", "effect"},
		{"results in", "cause", "effect"},
		{"affects", "cause", "effect"},
		{"influences", "cause", "effect"},
		{"depends on", "effect", "cause"},
	}

	for _, obs := range observations {
		obsLower := strings.ToLower(obs)

		// Extract variables based on causal patterns
		for _, pattern := range causalPatterns {
			if strings.Contains(obsLower, pattern.pattern) {
				parts := strings.Split(obsLower, pattern.pattern)
				if len(parts) == 2 {
					// Extract cause variable
					causeVar := cr.cleanVariableName(parts[0])
					if causeVar != "" && variableMap[causeVar] == nil {
						varCounter++
						variableMap[causeVar] = &types.CausalVariable{
							ID:         fmt.Sprintf("var-%d", varCounter),
							Name:       causeVar,
							Type:       cr.inferVariableType(causeVar, obs),
							Observable: true,
							Metadata:   map[string]interface{}{},
						}
					}

					// Extract effect variable
					effectVar := cr.cleanVariableName(parts[1])
					if effectVar != "" && variableMap[effectVar] == nil {
						varCounter++
						variableMap[effectVar] = &types.CausalVariable{
							ID:         fmt.Sprintf("var-%d", varCounter),
							Name:       effectVar,
							Type:       cr.inferVariableType(effectVar, obs),
							Observable: true,
							Metadata:   map[string]interface{}{},
						}
					}
				}
			}
		}
	}

	// Convert map to slice
	variables := make([]*types.CausalVariable, 0, len(variableMap))
	for _, v := range variableMap {
		variables = append(variables, v)
	}

	return variables
}

// cleanVariableName extracts and cleans a variable name
func (cr *CausalReasoner) cleanVariableName(text string) string {
	text = strings.TrimSpace(text)

	// Remove common prefixes/suffixes
	text = strings.TrimPrefix(text, "when ")
	text = strings.TrimPrefix(text, "if ")
	text = strings.TrimPrefix(text, "the ")
	text = strings.TrimSuffix(text, ".")
	text = strings.TrimSuffix(text, ",")

	// Take first significant phrase (before commas, periods)
	for _, sep := range []string{",", ".", ":", ";"} {
		if idx := strings.Index(text, sep); idx > 0 {
			text = text[:idx]
		}
	}

	text = strings.TrimSpace(text)

	// Must be substantial enough
	if len(text) < 3 || len(strings.Fields(text)) == 0 {
		return ""
	}

	return text
}

// inferVariableType infers the type of variable from context
func (cr *CausalReasoner) inferVariableType(varName, context string) string {
	contextLower := strings.ToLower(context)
	varLower := strings.ToLower(varName)

	// Check for binary indicators
	binaryIndicators := []string{"yes", "no", "true", "false", "present", "absent", "exists", "occurs"}
	for _, indicator := range binaryIndicators {
		if strings.Contains(contextLower, indicator) {
			return "binary"
		}
	}

	// Check for continuous indicators
	continuousIndicators := []string{"amount", "level", "rate", "degree", "temperature", "pressure", "speed", "cost", "price"}
	for _, indicator := range continuousIndicators {
		if strings.Contains(varLower, indicator) {
			return "continuous"
		}
	}

	// Check for categorical indicators
	categoricalIndicators := []string{"type", "kind", "category", "class", "color", "status"}
	for _, indicator := range categoricalIndicators {
		if strings.Contains(varLower, indicator) {
			return "categorical"
		}
	}

	// Default to continuous for numeric contexts
	return "continuous"
}

// identifyCausalLinks identifies causal relationships between variables
func (cr *CausalReasoner) identifyCausalLinks(variables []*types.CausalVariable, observations []string) []*types.CausalLink {
	links := make([]*types.CausalLink, 0)
	linkCounter := 0

	// Map variable names to IDs for lookup
	nameToVar := make(map[string]*types.CausalVariable)
	for _, v := range variables {
		nameToVar[strings.ToLower(v.Name)] = v
	}

	// Analyze each observation for causal relationships
	for _, obs := range observations {
		obsLower := strings.ToLower(obs)

		// Look for causal patterns
		if link := cr.extractCausalLink(obsLower, nameToVar, &linkCounter); link != nil {
			links = append(links, link)
		}
	}

	return links
}

// extractCausalLink extracts a causal link from an observation
func (cr *CausalReasoner) extractCausalLink(obs string, nameToVar map[string]*types.CausalVariable, counter *int) *types.CausalLink {
	// Patterns with their link types
	patterns := []struct {
		keywords []string
		linkType string
	}{
		{[]string{"increases", "raises", "boosts", "enhances"}, "positive"},
		{[]string{"decreases", "reduces", "lowers", "diminishes"}, "negative"},
		{[]string{"causes", "leads to", "results in", "produces"}, "positive"},
		{[]string{"prevents", "blocks", "inhibits"}, "negative"},
	}

	for _, pattern := range patterns {
		for _, keyword := range pattern.keywords {
			if strings.Contains(obs, keyword) {
				// Find variables on either side of the keyword
				parts := strings.Split(obs, keyword)
				if len(parts) != 2 {
					continue
				}

				// Find matching variables
				var fromVar, toVar *types.CausalVariable
				for varName, v := range nameToVar {
					if strings.Contains(parts[0], varName) {
						fromVar = v
					}
					if strings.Contains(parts[1], varName) {
						toVar = v
					}
				}

				if fromVar != nil && toVar != nil {
					*counter++
					strength := cr.estimateLinkStrength(obs, keyword)
					confidence := cr.estimateLinkConfidence(obs)

					return &types.CausalLink{
						ID:         fmt.Sprintf("link-%d", *counter),
						From:       fromVar.ID,
						To:         toVar.ID,
						Strength:   strength,
						Type:       pattern.linkType,
						Confidence: confidence,
						Evidence:   []string{obs},
						Metadata:   map[string]interface{}{},
					}
				}
			}
		}
	}

	return nil
}

// estimateLinkStrength estimates the strength of a causal link
func (cr *CausalReasoner) estimateLinkStrength(obs, keyword string) float64 {
	// Strong indicators
	strongWords := []string{"strongly", "significantly", "greatly", "substantially", "dramatically"}
	for _, word := range strongWords {
		if strings.Contains(strings.ToLower(obs), word) {
			return 0.9
		}
	}

	// Moderate indicators
	moderateWords := []string{"moderately", "somewhat", "partially"}
	for _, word := range moderateWords {
		if strings.Contains(strings.ToLower(obs), word) {
			return 0.6
		}
	}

	// Weak indicators
	weakWords := []string{"slightly", "marginally", "weakly", "may"}
	for _, word := range weakWords {
		if strings.Contains(strings.ToLower(obs), word) {
			return 0.3
		}
	}

	// Default strength
	return 0.7
}

// estimateLinkConfidence estimates confidence in a causal link
func (cr *CausalReasoner) estimateLinkConfidence(obs string) float64 {
	// High confidence indicators
	highConfWords := []string{"proven", "demonstrated", "established", "confirmed", "definitely"}
	for _, word := range highConfWords {
		if strings.Contains(strings.ToLower(obs), word) {
			return 0.9
		}
	}

	// Low confidence indicators
	lowConfWords := []string{"possibly", "might", "perhaps", "uncertain", "unclear", "suspected"}
	for _, word := range lowConfWords {
		if strings.Contains(strings.ToLower(obs), word) {
			return 0.5
		}
	}

	// Default confidence
	return 0.7
}

// performGraphSurgery applies Pearl's graph surgery for interventions
// When we intervene on a variable (do(X=x)), we break all incoming causal links to X
// This represents setting X to a fixed value regardless of its natural causes
func (cr *CausalReasoner) performGraphSurgery(graph *types.CausalGraph, interventionVarID string) *types.CausalGraph {
	// Create a deep copy of the graph to avoid modifying the original
	surgicalGraph := &types.CausalGraph{
		ID:          graph.ID + "-surgical",
		Description: graph.Description + " (with graph surgery)",
		Variables:   make([]*types.CausalVariable, len(graph.Variables)),
		Links:       make([]*types.CausalLink, 0),
		Metadata:    make(map[string]interface{}),
		CreatedAt:   graph.CreatedAt,
	}

	// Deep copy variables
	for i, v := range graph.Variables {
		surgicalGraph.Variables[i] = &types.CausalVariable{
			ID:         v.ID,
			Name:       v.Name,
			Type:       v.Type,
			Observable: v.Observable,
			Metadata:   v.Metadata,
		}
	}

	// Deep copy metadata
	for k, v := range graph.Metadata {
		surgicalGraph.Metadata[k] = v
	}

	// CRITICAL: Copy all links EXCEPT those pointing TO the intervention variable
	// This is the core of Pearl's do-calculus - we break the causal mechanisms
	// that normally determine the intervention variable's value
	removedCount := 0
	for _, link := range graph.Links {
		if link.To != interventionVarID {
			// Keep this link - it's not pointing to the intervention variable
			surgicalGraph.Links = append(surgicalGraph.Links, &types.CausalLink{
				ID:         link.ID,
				From:       link.From,
				To:         link.To,
				Strength:   link.Strength,
				Type:       link.Type,
				Confidence: link.Confidence,
				Evidence:   link.Evidence,
				Metadata:   link.Metadata,
			})
		} else {
			// Remove this link - it points to the intervention variable
			removedCount++
		}
	}

	// Record the surgery in metadata
	surgicalGraph.Metadata["graph_surgery"] = map[string]interface{}{
		"intervention_variable": interventionVarID,
		"removed_edge_count":    removedCount,
		"surgery_type":          "do-calculus",
		"description":           fmt.Sprintf("Removed %d incoming edges to variable %s", removedCount, interventionVarID),
	}

	return surgicalGraph
}

// SimulateIntervention simulates the effects of an intervention using Pearl's do-calculus
func (cr *CausalReasoner) SimulateIntervention(graphID, variableID, interventionType string) (*types.CausalIntervention, error) {
	cr.mu.RLock()
	graph, exists := cr.graphs[graphID]
	cr.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("graph not found: %s", graphID)
	}

	// Find the target variable
	var targetVar *types.CausalVariable
	for _, v := range graph.Variables {
		if v.ID == variableID || strings.EqualFold(v.Name, variableID) {
			targetVar = v
			break
		}
	}

	if targetVar == nil {
		return nil, fmt.Errorf("variable not found: %s", variableID)
	}

	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.counter++

	// CRITICAL: Apply graph surgery for intervention (Pearl's do-calculus)
	// When we do(X=x), we remove all incoming edges to X, breaking its natural causes
	// This isolates X from its parents while preserving its effects on descendants
	surgicalGraph := cr.performGraphSurgery(graph, targetVar.ID)

	// Trace downstream effects using the surgically modified graph
	effects := cr.traceDownstreamEffects(surgicalGraph, targetVar.ID, interventionType, 1)

	// Calculate overall confidence
	overallConfidence := cr.calculateInterventionConfidence(effects)

	intervention := &types.CausalIntervention{
		ID:               fmt.Sprintf("intervention-%d", cr.counter),
		GraphID:          graphID,
		Variable:         targetVar.Name,
		InterventionType: interventionType,
		PredictedEffects: effects,
		Confidence:       overallConfidence,
		Metadata: map[string]interface{}{
			"graph_surgery_applied": true,
			"intervention_note":     "Applied Pearl's do-calculus: removed incoming edges to intervention variable",
		},
		CreatedAt: time.Now(),
	}

	return intervention, nil
}

// traceDownstreamEffects traces the causal effects downstream
func (cr *CausalReasoner) traceDownstreamEffects(graph *types.CausalGraph, startVarID, interventionType string, pathLength int) []*types.PredictedEffect {
	effects := make([]*types.PredictedEffect, 0)
	visited := make(map[string]bool)

	// Find outgoing links from start variable
	for _, link := range graph.Links {
		if link.From == startVarID && !visited[link.To] {
			visited[link.To] = true

			// Find target variable
			var targetVar *types.CausalVariable
			for _, v := range graph.Variables {
				if v.ID == link.To {
					targetVar = v
					break
				}
			}

			if targetVar == nil {
				continue
			}

			// Determine effect direction
			effectDirection := cr.determineEffectDirection(interventionType, link.Type)

			// Calculate probability
			probability := link.Confidence * link.Strength

			// Generate explanation
			explanation := fmt.Sprintf("Via %s causal link from intervention variable", link.Type)

			effect := &types.PredictedEffect{
				Variable:    targetVar.Name,
				Effect:      effectDirection,
				Magnitude:   link.Strength,
				Probability: probability,
				Explanation: explanation,
				PathLength:  pathLength,
			}

			effects = append(effects, effect)

			// Recursively trace further effects (limit depth)
			if pathLength < 3 {
				furtherEffects := cr.traceDownstreamEffects(graph, link.To, effectDirection, pathLength+1)
				effects = append(effects, furtherEffects...)
			}
		}
	}

	return effects
}

// determineEffectDirection determines the direction of effect
func (cr *CausalReasoner) determineEffectDirection(interventionType, linkType string) string {
	if interventionType == "increase" {
		if linkType == "positive" {
			return "increase"
		}
		return "decrease"
	} else if interventionType == "decrease" {
		if linkType == "positive" {
			return "decrease"
		}
		return "increase"
	}
	return "change"
}

// calculateInterventionConfidence calculates overall intervention confidence
func (cr *CausalReasoner) calculateInterventionConfidence(effects []*types.PredictedEffect) float64 {
	if len(effects) == 0 {
		return 0.5
	}

	totalProb := 0.0
	for _, effect := range effects {
		// Weight by inverse path length (closer effects more certain)
		weight := 1.0 / float64(effect.PathLength)
		totalProb += effect.Probability * weight
	}

	avgConfidence := totalProb / float64(len(effects))

	// Clamp to valid range
	if avgConfidence > 1.0 {
		avgConfidence = 1.0
	}
	if avgConfidence < 0.0 {
		avgConfidence = 0.0
	}

	return avgConfidence
}

// GenerateCounterfactual generates a counterfactual scenario
func (cr *CausalReasoner) GenerateCounterfactual(graphID, scenario string, changes map[string]string) (*types.Counterfactual, error) {
	cr.mu.RLock()
	graph, exists := cr.graphs[graphID]
	cr.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("graph not found: %s", graphID)
	}

	if scenario == "" {
		return nil, fmt.Errorf("scenario description required")
	}

	if len(changes) == 0 {
		return nil, fmt.Errorf("at least one change required")
	}

	cr.mu.Lock()
	defer cr.mu.Unlock()

	cr.counter++

	// Predict outcomes based on changes
	outcomes := make(map[string]string)

	for varName, changeValue := range changes {
		// Find variable
		var sourceVar *types.CausalVariable
		for _, v := range graph.Variables {
			if strings.EqualFold(v.Name, varName) {
				sourceVar = v
				break
			}
		}

		if sourceVar == nil {
			continue
		}

		// Apply graph surgery for this counterfactual intervention
		// This correctly models what would happen if we forcibly set the variable
		surgicalGraph := cr.performGraphSurgery(graph, sourceVar.ID)

		// Trace effects using the surgically modified graph
		effects := cr.traceDownstreamEffects(surgicalGraph, sourceVar.ID, changeValue, 1)

		for _, effect := range effects {
			outcomes[effect.Variable] = effect.Effect
		}
	}

	// Estimate plausibility
	plausibility := cr.estimateCounterfactualPlausibility(changes, outcomes)

	counterfactual := &types.Counterfactual{
		ID:           fmt.Sprintf("counterfactual-%d", cr.counter),
		GraphID:      graphID,
		Scenario:     scenario,
		Changes:      changes,
		Outcomes:     outcomes,
		Plausibility: plausibility,
		Metadata:     map[string]interface{}{},
		CreatedAt:    time.Now(),
	}

	return counterfactual, nil
}

// estimateCounterfactualPlausibility estimates how plausible a counterfactual is
func (cr *CausalReasoner) estimateCounterfactualPlausibility(changes, outcomes map[string]string) float64 {
	// Base plausibility
	plausibility := 0.7

	// Reduce plausibility for multiple simultaneous changes
	if len(changes) > 3 {
		plausibility *= 0.8
	}

	// Reduce plausibility for very long causal chains
	if len(outcomes) > 5 {
		plausibility *= 0.9
	}

	return plausibility
}

// GetGraph retrieves a causal graph by ID
func (cr *CausalReasoner) GetGraph(graphID string) (*types.CausalGraph, error) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	graph, exists := cr.graphs[graphID]
	if !exists {
		return nil, fmt.Errorf("graph not found: %s", graphID)
	}

	return graph, nil
}

// AnalyzeCorrelationVsCausation distinguishes correlation from causation
func (cr *CausalReasoner) AnalyzeCorrelationVsCausation(observation string) (string, error) {
	if observation == "" {
		return "", fmt.Errorf("observation cannot be empty")
	}

	obsLower := strings.ToLower(observation)

	// Check for correlation-only indicators
	correlationIndicators := []string{
		"correlated with",
		"associated with",
		"related to",
		"linked to",
		"connected to",
	}

	// Check for causal indicators
	causalIndicators := []string{
		"causes",
		"leads to",
		"results in",
		"produces",
		"triggers",
		"creates",
	}

	hasCorrelation := false
	hasCausation := false

	for _, indicator := range correlationIndicators {
		if strings.Contains(obsLower, indicator) {
			hasCorrelation = true
			break
		}
	}

	for _, indicator := range causalIndicators {
		if strings.Contains(obsLower, indicator) {
			hasCausation = true
			break
		}
	}

	if hasCausation {
		return "This observation suggests a causal relationship. However, verify: (1) temporal precedence - does cause precede effect? (2) no confounding variables, (3) mechanism of action is plausible.", nil
	}

	if hasCorrelation {
		return "This observation indicates correlation, not necessarily causation. Consider: (1) could a third variable explain both? (2) is reverse causation possible? (3) is this merely coincidental?", nil
	}

	return "The relationship type is unclear. To establish causation, need: (1) temporal ordering, (2) elimination of confounders, (3) demonstration of mechanism, (4) ideally, experimental intervention evidence.", nil
}
