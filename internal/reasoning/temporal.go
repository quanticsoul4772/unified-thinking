// Package reasoning provides advanced reasoning capabilities.
package reasoning

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// TemporalReasoner analyzes short-term vs long-term implications
type TemporalReasoner struct {
	mu      sync.RWMutex
	counter int
}

// NewTemporalReasoner creates a new temporal reasoner
func NewTemporalReasoner() *TemporalReasoner {
	return &TemporalReasoner{}
}

// AnalyzeTemporal performs temporal analysis on a decision or situation
func (tr *TemporalReasoner) AnalyzeTemporal(situation string, timeHorizon string) (*types.TemporalAnalysis, error) {
	if situation == "" {
		return nil, fmt.Errorf("situation cannot be empty")
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()

	tr.counter++

	// Normalize time horizon
	normalizedHorizon := tr.normalizeTimeHorizon(timeHorizon)

	// Analyze short-term implications
	shortTermView := tr.analyzeShortTerm(situation)

	// Analyze long-term implications
	longTermView := tr.analyzeLongTerm(situation, normalizedHorizon)

	// Identify tradeoffs
	tradeoffs := tr.identifyTradeoffs(situation, shortTermView, longTermView)

	// Generate recommendation
	recommendation := tr.generateRecommendation(shortTermView, longTermView, tradeoffs, normalizedHorizon)

	analysis := &types.TemporalAnalysis{
		ID:             fmt.Sprintf("temporal-%d", tr.counter),
		ShortTermView:  shortTermView,
		LongTermView:   longTermView,
		TimeHorizon:    normalizedHorizon,
		Tradeoffs:      tradeoffs,
		Recommendation: recommendation,
		Metadata:       make(map[string]interface{}),
		CreatedAt:      time.Now(),
	}

	return analysis, nil
}

// normalizeTimeHorizon converts various time horizon formats to standard values
func (tr *TemporalReasoner) normalizeTimeHorizon(horizon string) string {
	horizonLower := strings.ToLower(strings.TrimSpace(horizon))

	// Map to standard horizons
	if horizonLower == "" {
		return "months" // Default
	}

	if strings.Contains(horizonLower, "day") || strings.Contains(horizonLower, "week") || strings.Contains(horizonLower, "immediate") || strings.Contains(horizonLower, "short") {
		return "days-weeks"
	}

	if strings.Contains(horizonLower, "month") || strings.Contains(horizonLower, "quarter") || strings.Contains(horizonLower, "medium") {
		return "months"
	}

	if strings.Contains(horizonLower, "year") || strings.Contains(horizonLower, "long") || strings.Contains(horizonLower, "strategic") {
		return "years"
	}

	if strings.Contains(horizonLower, "decade") || strings.Contains(horizonLower, "generation") {
		return "decades"
	}

	return "months" // Default fallback
}

// analyzeShortTerm analyzes immediate/short-term implications
func (tr *TemporalReasoner) analyzeShortTerm(situation string) string {
	situationLower := strings.ToLower(situation)

	var implications []string

	// Check for cost implications
	if strings.Contains(situationLower, "cost") || strings.Contains(situationLower, "budget") || strings.Contains(situationLower, "expense") {
		implications = append(implications, "immediate financial impact")
	}

	// Check for implementation effort
	if strings.Contains(situationLower, "implement") || strings.Contains(situationLower, "deploy") || strings.Contains(situationLower, "launch") {
		implications = append(implications, "implementation effort and disruption")
	}

	// Check for user/stakeholder impact
	if strings.Contains(situationLower, "user") || strings.Contains(situationLower, "customer") || strings.Contains(situationLower, "employee") {
		implications = append(implications, "immediate stakeholder reactions")
	}

	// Check for operational impact
	if strings.Contains(situationLower, "operation") || strings.Contains(situationLower, "process") || strings.Contains(situationLower, "workflow") {
		implications = append(implications, "operational adjustment period")
	}

	// Check for risk
	if strings.Contains(situationLower, "risk") || strings.Contains(situationLower, "challenge") || strings.Contains(situationLower, "problem") {
		implications = append(implications, "near-term risks and challenges")
	}

	// Build short-term view
	if len(implications) == 0 {
		return "Short-term: Initial implementation will require resources and attention. Immediate feedback will reveal early successes and challenges."
	}

	return fmt.Sprintf("Short-term: %s. Early indicators will emerge within weeks, requiring close monitoring and potential adjustments.", strings.Join(implications, "; "))
}

// analyzeLongTerm analyzes long-term implications
func (tr *TemporalReasoner) analyzeLongTerm(situation string, horizon string) string {
	situationLower := strings.ToLower(situation)

	var implications []string

	// Check for scalability
	if strings.Contains(situationLower, "scale") || strings.Contains(situationLower, "growth") || strings.Contains(situationLower, "expand") {
		implications = append(implications, "scalability and growth potential")
	}

	// Check for sustainability
	if strings.Contains(situationLower, "sustain") || strings.Contains(situationLower, "maintain") || strings.Contains(situationLower, "long-term") {
		implications = append(implications, "long-term sustainability")
	}

	// Check for strategic alignment
	if strings.Contains(situationLower, "strategy") || strings.Contains(situationLower, "vision") || strings.Contains(situationLower, "future") {
		implications = append(implications, "strategic alignment and competitive position")
	}

	// Check for cultural/organizational impact
	if strings.Contains(situationLower, "culture") || strings.Contains(situationLower, "organization") || strings.Contains(situationLower, "team") {
		implications = append(implications, "organizational culture and capability development")
	}

	// Check for ecosystem effects
	if strings.Contains(situationLower, "ecosystem") || strings.Contains(situationLower, "partner") || strings.Contains(situationLower, "industry") {
		implications = append(implications, "ecosystem and market dynamics")
	}

	// Check for technical debt or legacy
	if strings.Contains(situationLower, "debt") || strings.Contains(situationLower, "legacy") || strings.Contains(situationLower, "maintenance") {
		implications = append(implications, "technical debt and maintenance burden")
	}

	// Build long-term view based on horizon
	horizonPhrase := "over time"
	if horizon == "days-weeks" {
		horizonPhrase = "in the coming weeks"
	} else if horizon == "months" {
		horizonPhrase = "over the next 6-12 months"
	} else if horizon == "years" {
		horizonPhrase = "over multiple years"
	} else if horizon == "decades" {
		horizonPhrase = "over the long term"
	}

	if len(implications) == 0 {
		return fmt.Sprintf("Long-term: %s, cumulative effects will become evident. Success depends on sustained commitment and adaptation.", horizonPhrase)
	}

	return fmt.Sprintf("Long-term: %s, this will impact %s. The full benefits and costs will compound over time.", horizonPhrase, strings.Join(implications, ", "))
}

// identifyTradeoffs identifies tensions between short and long-term considerations
func (tr *TemporalReasoner) identifyTradeoffs(situation string, shortTerm, longTerm string) []string {
	situationLower := strings.ToLower(situation)
	tradeoffs := make([]string, 0)

	// Common temporal tradeoffs
	tradeoffPatterns := map[string]struct {
		shortTerm string
		longTerm  string
	}{
		"speed": {
			shortTerm: "Rapid deployment provides quick wins but may accumulate technical debt",
			longTerm:  "Careful planning delays benefits but creates sustainable foundation",
		},
		"cost": {
			shortTerm: "Lower initial investment reduces near-term risk but may increase long-term costs",
			longTerm:  "Higher upfront investment strains current budget but reduces total cost of ownership",
		},
		"simple": {
			shortTerm: "Simple solution is faster to implement but may not scale",
			longTerm:  "Comprehensive solution takes longer but provides better long-term capabilities",
		},
		"manual": {
			shortTerm: "Manual processes are quick to start but resource-intensive over time",
			longTerm:  "Automation requires upfront effort but delivers compounding efficiency gains",
		},
	}

	// Detect relevant tradeoffs
	for keyword, tradeoff := range tradeoffPatterns {
		if strings.Contains(situationLower, keyword) {
			tradeoffs = append(tradeoffs, fmt.Sprintf("Time horizon tension: %s vs %s", tradeoff.shortTerm, tradeoff.longTerm))
		}
	}

	// Generic tradeoffs if none detected
	if len(tradeoffs) == 0 {
		tradeoffs = append(tradeoffs, "Immediate results vs sustained impact: Quick wins may not lead to lasting change")
		tradeoffs = append(tradeoffs, "Resource allocation: Near-term investment vs long-term value")
		tradeoffs = append(tradeoffs, "Risk profile: Short-term certainty vs long-term potential")
	}

	return tradeoffs
}

// generateRecommendation generates a temporal recommendation
func (tr *TemporalReasoner) generateRecommendation(shortTerm, longTerm string, tradeoffs []string, horizon string) string {
	// Analyze emphasis in short vs long term
	shortTermEmphasis := tr.countImplications(shortTerm)
	longTermEmphasis := tr.countImplications(longTerm)

	var recommendation string

	if shortTermEmphasis > longTermEmphasis*2 {
		// Strong short-term focus
		recommendation = "Recommendation: Prioritize short-term execution while establishing foundations for future sustainability. "
		recommendation += "Implement quick wins first, then reinvest gains into long-term capabilities. "
		recommendation += "Monitor closely for signs that short-term optimizations are creating long-term liabilities."
	} else if longTermEmphasis > shortTermEmphasis*2 {
		// Strong long-term focus
		recommendation = "Recommendation: Take a strategic long-term view with staged implementation. "
		recommendation += "Accept higher upfront investment for sustainable outcomes. "
		recommendation += "Communicate clear milestones to maintain stakeholder confidence during the longer timeline."
	} else {
		// Balanced approach
		recommendation = "Recommendation: Balance short and long-term considerations through phased approach. "
		recommendation += "Start with high-value, low-risk initiatives to build momentum while planning for long-term evolution. "
		recommendation += "Create feedback loops to adjust strategy based on emerging results."
	}

	// Add horizon-specific guidance
	if horizon == "days-weeks" {
		recommendation += " Given the immediate timeframe, focus on execution speed and rapid feedback cycles."
	} else if horizon == "years" || horizon == "decades" {
		recommendation += " Given the extended timeframe, invest in sustainable infrastructure and adaptable systems."
	}

	return recommendation
}

// countImplications counts the number of implications mentioned
func (tr *TemporalReasoner) countImplications(text string) int {
	// Count sentences/implications (rough heuristic)
	implications := 0
	if strings.Contains(text, ";") {
		implications = strings.Count(text, ";") + 1
	} else if strings.Contains(text, ".") {
		implications = strings.Count(text, ".") + 1
	} else {
		implications = 1
	}
	return implications
}

// CompareTimeHorizons compares implications across different time horizons
func (tr *TemporalReasoner) CompareTimeHorizons(situation string) (map[string]*types.TemporalAnalysis, error) {
	if situation == "" {
		return nil, fmt.Errorf("situation cannot be empty")
	}

	horizons := []string{"days-weeks", "months", "years"}
	analyses := make(map[string]*types.TemporalAnalysis)

	for _, horizon := range horizons {
		analysis, err := tr.AnalyzeTemporal(situation, horizon)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze horizon %s: %w", horizon, err)
		}
		analyses[horizon] = analysis
	}

	return analyses, nil
}

// IdentifyOptimalTiming suggests optimal timing for a decision
func (tr *TemporalReasoner) IdentifyOptimalTiming(situation string, constraints []string) (string, error) {
	if situation == "" {
		return "", fmt.Errorf("situation cannot be empty")
	}

	tr.mu.RLock()
	defer tr.mu.RUnlock()

	situationLower := strings.ToLower(situation)
	timing := "flexible" // Default

	// Check for urgency indicators
	urgentKeywords := []string{"urgent", "immediate", "critical", "emergency", "deadline"}
	for _, keyword := range urgentKeywords {
		if strings.Contains(situationLower, keyword) {
			timing = "immediate"
			break
		}
	}

	// Check for timing constraints
	if timing == "flexible" && len(constraints) > 0 {
		for _, constraint := range constraints {
			constraintLower := strings.ToLower(constraint)
			if strings.Contains(constraintLower, "deadline") || strings.Contains(constraintLower, "date") {
				timing = "constrained"
				break
			}
		}
	}

	// Check for strategic timing opportunities
	if timing == "flexible" {
		strategicKeywords := []string{"strategic", "planned", "roadmap", "future"}
		for _, keyword := range strategicKeywords {
			if strings.Contains(situationLower, keyword) {
				timing = "strategic"
				break
			}
		}
	}

	// Generate timing recommendation
	var recommendation string
	switch timing {
	case "immediate":
		recommendation = "Act immediately. The situation requires urgent action with limited time for deliberation. Focus on rapid execution and iteration."
	case "constrained":
		recommendation = "Execute according to constraints. Work backwards from deadlines to ensure adequate preparation and quality."
	case "strategic":
		recommendation = "Time strategically. Consider market conditions, resource availability, and stakeholder readiness. Optimal timing may provide significant advantages."
	default:
		recommendation = "Timing is flexible. Consider benefits of moving quickly (momentum, first-mover advantage) vs waiting (more information, better preparation)."
	}

	return recommendation, nil
}
