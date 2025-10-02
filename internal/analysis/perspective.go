// Package analysis provides analytical reasoning capabilities.
package analysis

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// PerspectiveAnalyzer generates and analyzes multiple stakeholder perspectives
type PerspectiveAnalyzer struct {
	mu      sync.RWMutex
	counter int
}

// NewPerspectiveAnalyzer creates a new perspective analyzer
func NewPerspectiveAnalyzer() *PerspectiveAnalyzer {
	return &PerspectiveAnalyzer{}
}

// AnalyzePerspectives generates multiple stakeholder perspectives for a situation
func (pa *PerspectiveAnalyzer) AnalyzePerspectives(situation string, stakeholderHints []string) ([]*types.Perspective, error) {
	if situation == "" {
		return nil, fmt.Errorf("situation cannot be empty")
	}

	pa.mu.Lock()
	defer pa.mu.Unlock()

	perspectives := make([]*types.Perspective, 0)

	// If stakeholder hints provided, analyze from those perspectives
	if len(stakeholderHints) > 0 {
		for _, stakeholder := range stakeholderHints {
			perspective := pa.generatePerspective(situation, stakeholder)
			perspectives = append(perspectives, perspective)
		}
	} else {
		// Auto-detect relevant stakeholders from situation
		detectedStakeholders := pa.detectStakeholders(situation)
		for _, stakeholder := range detectedStakeholders {
			perspective := pa.generatePerspective(situation, stakeholder)
			perspectives = append(perspectives, perspective)
		}
	}

	// Detect conflicts between perspectives
	conflicts := pa.detectPerspectiveConflicts(perspectives)
	if len(conflicts) > 0 {
		// Add conflict metadata to perspectives
		for _, p := range perspectives {
			p.Metadata["conflicts"] = conflicts
		}
	}

	return perspectives, nil
}

// generatePerspective creates a perspective for a specific stakeholder
func (pa *PerspectiveAnalyzer) generatePerspective(situation, stakeholder string) *types.Perspective {
	pa.counter++

	// Extract key concerns based on stakeholder type
	concerns := pa.extractConcerns(situation, stakeholder)
	priorities := pa.extractPriorities(stakeholder)
	constraints := pa.extractConstraints(stakeholder)
	viewpoint := pa.synthesizeViewpoint(situation, stakeholder, concerns)

	// Confidence based on how well-defined the stakeholder is
	confidence := pa.estimateConfidence(stakeholder, situation)

	return &types.Perspective{
		ID:          fmt.Sprintf("perspective-%d", pa.counter),
		Stakeholder: stakeholder,
		Viewpoint:   viewpoint,
		Concerns:    concerns,
		Priorities:  priorities,
		Constraints: constraints,
		Confidence:  confidence,
		Metadata:    map[string]interface{}{},
		CreatedAt:   time.Now(),
	}
}

// detectStakeholders identifies relevant stakeholders from the situation description
func (pa *PerspectiveAnalyzer) detectStakeholders(situation string) []string {
	situationLower := strings.ToLower(situation)
	stakeholders := make([]string, 0)

	// Check for common stakeholder indicators
	stakeholderPatterns := map[string][]string{
		"users":      {"user", "customer", "client", "consumer"},
		"employees":  {"employee", "worker", "staff", "team"},
		"management": {"manager", "executive", "leadership", "ceo", "director"},
		"investors":  {"investor", "shareholder", "stakeholder", "board"},
		"community":  {"community", "public", "society", "citizen"},
		"regulators": {"regulator", "government", "compliance", "legal"},
		"partners":   {"partner", "supplier", "vendor", "contractor"},
	}

	for stakeholder, patterns := range stakeholderPatterns {
		for _, pattern := range patterns {
			if strings.Contains(situationLower, pattern) {
				stakeholders = append(stakeholders, stakeholder)
				break
			}
		}
	}

	// If no stakeholders detected, use generic set
	if len(stakeholders) == 0 {
		stakeholders = []string{"decision-maker", "affected-parties", "implementers"}
	}

	return stakeholders
}

// extractConcerns identifies key concerns for a stakeholder type
func (pa *PerspectiveAnalyzer) extractConcerns(situation, stakeholder string) []string {
	concerns := make([]string, 0)
	situationLower := strings.ToLower(situation)
	stakeholderLower := strings.ToLower(stakeholder)

	// Stakeholder-specific concern patterns
	concernPatterns := map[string][]string{
		"user":        {"usability", "accessibility", "privacy", "cost", "reliability"},
		"customer":    {"value", "quality", "support", "price", "experience"},
		"employee":    {"workload", "job security", "compensation", "work environment", "career growth"},
		"management":  {"profitability", "efficiency", "risk", "scalability", "market position"},
		"investor":    {"return on investment", "risk", "growth potential", "market share", "valuation"},
		"community":   {"social impact", "environmental impact", "fairness", "accessibility", "safety"},
		"regulator":   {"compliance", "safety", "fairness", "transparency", "accountability"},
		"partner":     {"reliability", "communication", "mutual benefit", "contract terms", "long-term viability"},
		"default":     {"impact", "feasibility", "risks", "benefits", "implementation"},
	}

	// Find matching patterns
	var relevantPatterns []string
	for key, patterns := range concernPatterns {
		if strings.Contains(stakeholderLower, key) {
			relevantPatterns = patterns
			break
		}
	}
	if len(relevantPatterns) == 0 {
		relevantPatterns = concernPatterns["default"]
	}

	// Extract concerns that appear relevant to the situation
	for _, concern := range relevantPatterns {
		if strings.Contains(situationLower, concern) || len(concerns) < 3 {
			concerns = append(concerns, concern)
			if len(concerns) >= 5 {
				break
			}
		}
	}

	return concerns
}

// extractPriorities determines priorities for a stakeholder type
func (pa *PerspectiveAnalyzer) extractPriorities(stakeholder string) []string {
	stakeholderLower := strings.ToLower(stakeholder)

	priorityMap := map[string][]string{
		"user":        {"ease of use", "reliability", "value for money"},
		"customer":    {"quality", "price", "customer service"},
		"employee":    {"fair compensation", "work-life balance", "job security"},
		"management":  {"profitability", "growth", "operational efficiency"},
		"investor":    {"returns", "risk mitigation", "long-term growth"},
		"community":   {"social benefit", "environmental sustainability", "equity"},
		"regulator":   {"public safety", "compliance", "consumer protection"},
		"partner":     {"mutual success", "clear communication", "reliable execution"},
	}

	for key, priorities := range priorityMap {
		if strings.Contains(stakeholderLower, key) {
			return priorities
		}
	}

	return []string{"positive outcomes", "minimal risk", "clear benefits"}
}

// extractConstraints identifies constraints for a stakeholder type
func (pa *PerspectiveAnalyzer) extractConstraints(stakeholder string) []string {
	stakeholderLower := strings.ToLower(stakeholder)

	constraintMap := map[string][]string{
		"user":        {"limited budget", "limited technical expertise", "time constraints"},
		"customer":    {"budget limitations", "alternative options available", "switching costs"},
		"employee":    {"limited authority", "resource constraints", "existing workload"},
		"management":  {"budget constraints", "timeline pressure", "stakeholder expectations"},
		"investor":    {"fiduciary duty", "portfolio diversification", "liquidity needs"},
		"community":   {"limited resources", "diverse needs", "existing infrastructure"},
		"regulator":   {"legal framework", "enforcement capacity", "political pressures"},
		"partner":     {"contractual obligations", "resource limitations", "competing priorities"},
	}

	for key, constraints := range constraintMap {
		if strings.Contains(stakeholderLower, key) {
			return constraints
		}
	}

	return []string{"practical limitations", "resource constraints", "external dependencies"}
}

// synthesizeViewpoint creates a coherent viewpoint summary
func (pa *PerspectiveAnalyzer) synthesizeViewpoint(situation, stakeholder string, concerns []string) string {
	// Create a perspective-specific interpretation of the situation
	concernsStr := "unknown concerns"
	if len(concerns) > 0 {
		concernsStr = strings.Join(concerns, ", ")
	}

	return fmt.Sprintf("From the %s perspective, this situation primarily raises concerns about %s. The key question is how to address these issues while balancing competing priorities and constraints.", stakeholder, concernsStr)
}

// estimateConfidence estimates confidence in perspective modeling
func (pa *PerspectiveAnalyzer) estimateConfidence(stakeholder, situation string) float64 {
	// Higher confidence for well-defined stakeholders
	wellDefinedStakeholders := []string{"user", "customer", "employee", "investor", "manager"}
	confidence := 0.6 // Base confidence

	stakeholderLower := strings.ToLower(stakeholder)
	for _, wd := range wellDefinedStakeholders {
		if strings.Contains(stakeholderLower, wd) {
			confidence = 0.8
			break
		}
	}

	// Increase confidence if situation mentions stakeholder explicitly
	if strings.Contains(strings.ToLower(situation), stakeholderLower) {
		confidence += 0.1
	}

	// Clamp to valid range
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// detectPerspectiveConflicts identifies conflicts between perspectives
func (pa *PerspectiveAnalyzer) detectPerspectiveConflicts(perspectives []*types.Perspective) []string {
	conflicts := make([]string, 0)

	// Check for priority conflicts
	for i := 0; i < len(perspectives); i++ {
		for j := i + 1; j < len(perspectives); j++ {
			p1 := perspectives[i]
			p2 := perspectives[j]

			// Check if priorities are opposed
			if pa.prioritiesConflict(p1.Priorities, p2.Priorities) {
				conflict := fmt.Sprintf("%s and %s have conflicting priorities", p1.Stakeholder, p2.Stakeholder)
				conflicts = append(conflicts, conflict)
			}

			// Check if concerns are opposed
			if pa.concernsConflict(p1.Concerns, p2.Concerns) {
				conflict := fmt.Sprintf("%s and %s have opposing concerns", p1.Stakeholder, p2.Stakeholder)
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts
}

// prioritiesConflict checks if two priority lists are in conflict
func (pa *PerspectiveAnalyzer) prioritiesConflict(priorities1, priorities2 []string) bool {
	// Simple heuristic: check for opposing terms
	opposingPairs := map[string]string{
		"speed":       "thoroughness",
		"cost":        "quality",
		"innovation":  "stability",
		"growth":      "sustainability",
		"flexibility": "standardization",
	}

	for _, p1 := range priorities1 {
		for _, p2 := range priorities2 {
			p1Lower := strings.ToLower(p1)
			p2Lower := strings.ToLower(p2)

			// Check for direct opposition
			for key, opposite := range opposingPairs {
				if (strings.Contains(p1Lower, key) && strings.Contains(p2Lower, opposite)) ||
					(strings.Contains(p1Lower, opposite) && strings.Contains(p2Lower, key)) {
					return true
				}
			}
		}
	}

	return false
}

// concernsConflict checks if two concern lists are in conflict
func (pa *PerspectiveAnalyzer) concernsConflict(concerns1, concerns2 []string) bool {
	// Check for mutually exclusive concerns
	mutuallyExclusive := map[string]string{
		"privacy":     "transparency",
		"security":    "accessibility",
		"control":     "autonomy",
		"efficiency":  "thoroughness",
		"automation":  "human oversight",
	}

	for _, c1 := range concerns1 {
		for _, c2 := range concerns2 {
			c1Lower := strings.ToLower(c1)
			c2Lower := strings.ToLower(c2)

			for key, exclusive := range mutuallyExclusive {
				if (strings.Contains(c1Lower, key) && strings.Contains(c2Lower, exclusive)) ||
					(strings.Contains(c1Lower, exclusive) && strings.Contains(c2Lower, key)) {
					return true
				}
			}
		}
	}

	return false
}

// ComparePerspectives compares two or more perspectives and identifies synergies and conflicts
func (pa *PerspectiveAnalyzer) ComparePerspectives(perspectives []*types.Perspective) (map[string]interface{}, error) {
	if len(perspectives) < 2 {
		return nil, fmt.Errorf("need at least 2 perspectives to compare")
	}

	pa.mu.RLock()
	defer pa.mu.RUnlock()

	result := make(map[string]interface{})

	// Find common concerns
	commonConcerns := pa.findCommonConcerns(perspectives)
	result["common_concerns"] = commonConcerns

	// Find common priorities
	commonPriorities := pa.findCommonPriorities(perspectives)
	result["common_priorities"] = commonPriorities

	// Find conflicts
	conflicts := pa.detectPerspectiveConflicts(perspectives)
	result["conflicts"] = conflicts

	// Generate synthesis
	result["synthesis"] = pa.generateSynthesis(commonConcerns, commonPriorities, conflicts)

	return result, nil
}

// findCommonConcerns identifies concerns shared by multiple perspectives
func (pa *PerspectiveAnalyzer) findCommonConcerns(perspectives []*types.Perspective) []string {
	concernCounts := make(map[string]int)

	for _, p := range perspectives {
		for _, concern := range p.Concerns {
			concernLower := strings.ToLower(concern)
			concernCounts[concernLower]++
		}
	}

	common := make([]string, 0)
	threshold := len(perspectives) / 2 // Appears in at least half
	for concern, count := range concernCounts {
		if count >= threshold {
			common = append(common, concern)
		}
	}

	return common
}

// findCommonPriorities identifies priorities shared by multiple perspectives
func (pa *PerspectiveAnalyzer) findCommonPriorities(perspectives []*types.Perspective) []string {
	priorityCounts := make(map[string]int)

	for _, p := range perspectives {
		for _, priority := range p.Priorities {
			priorityLower := strings.ToLower(priority)
			priorityCounts[priorityLower]++
		}
	}

	common := make([]string, 0)
	threshold := len(perspectives) / 2
	for priority, count := range priorityCounts {
		if count >= threshold {
			common = append(common, priority)
		}
	}

	return common
}

// generateSynthesis creates a synthesis of perspectives
func (pa *PerspectiveAnalyzer) generateSynthesis(commonConcerns, commonPriorities []string, conflicts []string) string {
	synthesis := "Analysis of multiple perspectives reveals: "

	if len(commonConcerns) > 0 {
		synthesis += fmt.Sprintf("Shared concerns include %s. ", strings.Join(commonConcerns, ", "))
	}

	if len(commonPriorities) > 0 {
		synthesis += fmt.Sprintf("Common priorities are %s. ", strings.Join(commonPriorities, ", "))
	}

	if len(conflicts) > 0 {
		synthesis += fmt.Sprintf("However, there are %d conflicts between perspectives that need resolution. ", len(conflicts))
	} else {
		synthesis += "Perspectives are largely aligned. "
	}

	synthesis += "A balanced approach should address shared concerns while navigating conflicts through compromise or phased implementation."

	return synthesis
}
