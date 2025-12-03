// Package reasoning - Domain-specific decomposition templates
package reasoning

import (
	"fmt"
	"strings"
	"time"

	"unified-thinking/internal/types"
)

// Domain represents the detected problem domain
type Domain string

const (
	DomainDebugging    Domain = "debugging"
	DomainProof        Domain = "proof"
	DomainArchitecture Domain = "architecture"
	DomainResearch     Domain = "research"
	DomainGeneral      Domain = "general"
)

// DomainTemplate defines a domain-specific decomposition template
type DomainTemplate struct {
	Domain       Domain                 `json:"domain"`
	Description  string                 `json:"description"`
	Steps        []DomainStep           `json:"steps"`
	Dependencies []DomainStepDependency `json:"dependencies"`
}

// DomainStep represents a step in a domain-specific template
type DomainStep struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Complexity  string `json:"complexity"`
	Priority    string `json:"priority"`
}

// DomainStepDependency represents a dependency between steps (0-indexed)
type DomainStepDependency struct {
	FromStep int    `json:"from_step"`
	ToStep   int    `json:"to_step"`
	Type     string `json:"type"` // "required", "optional", "parallel"
}

// commonStopWords contains words to exclude from entity extraction
var commonStopWords = map[string]bool{
	// Articles and determiners
	"a": true, "an": true, "the": true, "this": true, "that": true, "these": true, "those": true,
	// Pronouns
	"i": true, "you": true, "he": true, "she": true, "it": true, "we": true, "they": true,
	"me": true, "him": true, "her": true, "us": true, "them": true, "my": true, "your": true,
	"his": true, "its": true, "our": true, "their": true, "mine": true, "yours": true,
	// Common verbs
	"is": true, "are": true, "was": true, "were": true, "be": true, "been": true, "being": true,
	"have": true, "has": true, "had": true, "do": true, "does": true, "did": true,
	"will": true, "would": true, "could": true, "should": true, "may": true, "might": true,
	"must": true, "can": true, "shall": true,
	// Prepositions
	"in": true, "on": true, "at": true, "by": true, "for": true, "with": true, "about": true,
	"against": true, "between": true, "into": true, "through": true, "during": true,
	"before": true, "after": true, "above": true, "below": true, "to": true, "from": true,
	"up": true, "down": true, "out": true, "off": true, "over": true, "under": true,
	// Conjunctions
	"and": true, "but": true, "or": true, "nor": true, "so": true, "yet": true, "both": true,
	"either": true, "neither": true, "not": true, "only": true, "than": true, "when": true,
	"where": true, "while": true, "if": true, "then": true, "because": true, "although": true,
	// Common problem-solving words (don't add value as entities)
	"problem": true, "issue": true, "solution": true, "approach": true, "method": true,
	"way": true, "need": true, "want": true, "like": true, "how": true, "what": true,
	"which": true, "who": true, "why": true, "there": true, "here": true, "now": true,
	"also": true, "just": true, "very": true, "even": true, "still": true, "already": true,
	"some": true, "any": true, "all": true, "each": true, "every": true, "many": true,
	"much": true, "more": true, "most": true, "other": true, "another": true, "such": true,
	"no": true, "same": true, "different": true, "new": true, "old": true, "good": true,
	"bad": true, "first": true, "last": true, "next": true, "own": true, "well": true,
	"make": true, "made": true, "get": true, "got": true, "take": true, "put": true,
	"use": true, "used": true, "using": true, "work": true, "working": true, "works": true,
}

// domainKeywords maps keywords to domains for detection
var domainKeywords = map[Domain][]string{
	DomainDebugging: {
		"debug", "bug", "error", "fix", "crash", "fail", "broken", "issue",
		"trace", "stack", "exception", "flaky", "intermittent", "regression",
		"troubleshoot", "investigate", "diagnose", "symptom", "root cause",
	},
	DomainProof: {
		"prove", "proof", "theorem", "lemma", "axiom", "formal", "verify",
		"mathematical", "logic", "proposition", "hypothesis", "deduction",
		"induction", "contradiction", "qed", "valid", "sound", "complete",
	},
	DomainArchitecture: {
		"architect", "design", "system", "component", "module", "service",
		"microservice", "api", "interface", "pattern", "structure", "scale",
		"refactor", "migrate", "integration", "dependency", "coupling",
		"cohesion", "layer", "tier", "deploy", "infrastructure",
	},
	DomainResearch: {
		"research", "study", "analyze", "explore", "investigate", "survey",
		"literature", "experiment", "hypothesis", "data", "finding", "result",
		"conclusion", "methodology", "approach", "compare", "evaluate", "assess",
		"benchmark", "metric", "observation", "correlation", "causation",
	},
}

// domainTemplates contains the predefined templates for each domain
var domainTemplates = map[Domain]*DomainTemplate{
	DomainDebugging: {
		Domain:      DomainDebugging,
		Description: "Systematic debugging and issue resolution workflow",
		Steps: []DomainStep{
			{
				Name:        "Reproduce the issue",
				Description: "Create a reliable reproduction case that demonstrates the bug consistently",
				Complexity:  "medium",
				Priority:    "critical",
			},
			{
				Name:        "Gather diagnostic information",
				Description: "Collect logs, stack traces, error messages, and system state at time of failure",
				Complexity:  "low",
				Priority:    "high",
			},
			{
				Name:        "Formulate hypotheses",
				Description: "Generate possible root causes based on symptoms and available evidence",
				Complexity:  "medium",
				Priority:    "high",
			},
			{
				Name:        "Isolate the root cause",
				Description: "Systematically test hypotheses using binary search, logging, or debugger",
				Complexity:  "high",
				Priority:    "critical",
			},
			{
				Name:        "Develop and test fix",
				Description: "Implement the fix and verify it resolves the issue without regressions",
				Complexity:  "medium",
				Priority:    "high",
			},
			{
				Name:        "Add regression tests",
				Description: "Create tests that prevent this issue from recurring",
				Complexity:  "low",
				Priority:    "medium",
			},
		},
		Dependencies: []DomainStepDependency{
			{FromStep: 0, ToStep: 1, Type: "required"},
			{FromStep: 1, ToStep: 2, Type: "required"},
			{FromStep: 2, ToStep: 3, Type: "required"},
			{FromStep: 3, ToStep: 4, Type: "required"},
			{FromStep: 4, ToStep: 5, Type: "required"},
		},
	},
	DomainProof: {
		Domain:      DomainProof,
		Description: "Formal proof construction and verification workflow",
		Steps: []DomainStep{
			{
				Name:        "State the proposition clearly",
				Description: "Formalize the claim to be proven with precise definitions and notation",
				Complexity:  "medium",
				Priority:    "critical",
			},
			{
				Name:        "Identify proof strategy",
				Description: "Choose approach: direct proof, contradiction, induction, construction, etc.",
				Complexity:  "high",
				Priority:    "high",
			},
			{
				Name:        "List prerequisites",
				Description: "Identify axioms, lemmas, and theorems needed as building blocks",
				Complexity:  "medium",
				Priority:    "high",
			},
			{
				Name:        "Construct proof steps",
				Description: "Build the logical chain from premises to conclusion",
				Complexity:  "high",
				Priority:    "critical",
			},
			{
				Name:        "Verify each step",
				Description: "Check that each inference follows validly from previous steps",
				Complexity:  "high",
				Priority:    "critical",
			},
			{
				Name:        "Consider edge cases",
				Description: "Test boundary conditions and special cases that might break the proof",
				Complexity:  "medium",
				Priority:    "high",
			},
			{
				Name:        "Review completeness",
				Description: "Ensure all cases are covered and no gaps exist in the argument",
				Complexity:  "medium",
				Priority:    "high",
			},
		},
		Dependencies: []DomainStepDependency{
			{FromStep: 0, ToStep: 1, Type: "required"},
			{FromStep: 0, ToStep: 2, Type: "required"},
			{FromStep: 1, ToStep: 3, Type: "required"},
			{FromStep: 2, ToStep: 3, Type: "required"},
			{FromStep: 3, ToStep: 4, Type: "required"},
			{FromStep: 4, ToStep: 5, Type: "parallel"},
			{FromStep: 4, ToStep: 6, Type: "parallel"},
		},
	},
	DomainArchitecture: {
		Domain:      DomainArchitecture,
		Description: "System architecture design and evaluation workflow",
		Steps: []DomainStep{
			{
				Name:        "Define requirements and constraints",
				Description: "Clarify functional requirements, quality attributes, and constraints",
				Complexity:  "medium",
				Priority:    "critical",
			},
			{
				Name:        "Identify key components",
				Description: "Determine major system components and their responsibilities",
				Complexity:  "high",
				Priority:    "high",
			},
			{
				Name:        "Design interfaces and contracts",
				Description: "Define APIs, data formats, and interaction patterns between components",
				Complexity:  "high",
				Priority:    "high",
			},
			{
				Name:        "Evaluate trade-offs",
				Description: "Analyze architectural trade-offs: scalability, maintainability, cost, etc.",
				Complexity:  "high",
				Priority:    "high",
			},
			{
				Name:        "Document decisions (ADRs)",
				Description: "Create Architecture Decision Records for key choices",
				Complexity:  "medium",
				Priority:    "medium",
			},
			{
				Name:        "Create implementation roadmap",
				Description: "Plan phased implementation with milestones and dependencies",
				Complexity:  "medium",
				Priority:    "medium",
			},
		},
		Dependencies: []DomainStepDependency{
			{FromStep: 0, ToStep: 1, Type: "required"},
			{FromStep: 1, ToStep: 2, Type: "required"},
			{FromStep: 1, ToStep: 3, Type: "parallel"},
			{FromStep: 2, ToStep: 3, Type: "required"},
			{FromStep: 3, ToStep: 4, Type: "required"},
			{FromStep: 4, ToStep: 5, Type: "required"},
		},
	},
	DomainResearch: {
		Domain:      DomainResearch,
		Description: "Research investigation and synthesis workflow",
		Steps: []DomainStep{
			{
				Name:        "Define research questions",
				Description: "Articulate specific questions the research should answer",
				Complexity:  "medium",
				Priority:    "critical",
			},
			{
				Name:        "Conduct literature review",
				Description: "Survey existing work, identify gaps, and establish context",
				Complexity:  "high",
				Priority:    "high",
			},
			{
				Name:        "Design methodology",
				Description: "Choose research methods appropriate for the questions",
				Complexity:  "high",
				Priority:    "high",
			},
			{
				Name:        "Gather evidence",
				Description: "Collect data, run experiments, or analyze sources",
				Complexity:  "high",
				Priority:    "high",
			},
			{
				Name:        "Analyze findings",
				Description: "Process evidence to identify patterns, correlations, and insights",
				Complexity:  "high",
				Priority:    "high",
			},
			{
				Name:        "Synthesize conclusions",
				Description: "Draw conclusions that answer the research questions",
				Complexity:  "medium",
				Priority:    "high",
			},
			{
				Name:        "Identify limitations",
				Description: "Acknowledge constraints and areas for future work",
				Complexity:  "low",
				Priority:    "medium",
			},
		},
		Dependencies: []DomainStepDependency{
			{FromStep: 0, ToStep: 1, Type: "required"},
			{FromStep: 0, ToStep: 2, Type: "required"},
			{FromStep: 1, ToStep: 2, Type: "optional"},
			{FromStep: 2, ToStep: 3, Type: "required"},
			{FromStep: 3, ToStep: 4, Type: "required"},
			{FromStep: 4, ToStep: 5, Type: "required"},
			{FromStep: 5, ToStep: 6, Type: "required"},
		},
	},
	DomainGeneral: {
		Domain:      DomainGeneral,
		Description: "General problem-solving workflow",
		Steps: []DomainStep{
			{
				Name:        "Analyze and define the problem scope",
				Description: "Understand what the problem is and what success looks like",
				Complexity:  "low",
				Priority:    "high",
			},
			{
				Name:        "Gather required information and resources",
				Description: "Collect data, context, and tools needed to solve the problem",
				Complexity:  "medium",
				Priority:    "high",
			},
			{
				Name:        "Develop potential solutions",
				Description: "Generate and explore multiple approaches to solving the problem",
				Complexity:  "high",
				Priority:    "high",
			},
			{
				Name:        "Evaluate and select best approach",
				Description: "Compare solutions and choose the most appropriate one",
				Complexity:  "medium",
				Priority:    "medium",
			},
			{
				Name:        "Implement and test solution",
				Description: "Execute the chosen approach and verify it works",
				Complexity:  "high",
				Priority:    "medium",
			},
		},
		Dependencies: []DomainStepDependency{
			{FromStep: 0, ToStep: 1, Type: "required"},
			{FromStep: 1, ToStep: 2, Type: "required"},
			{FromStep: 2, ToStep: 3, Type: "required"},
			{FromStep: 3, ToStep: 4, Type: "required"},
		},
	},
}

// DetectDomain analyzes a problem statement and returns the most likely domain
func DetectDomain(problem string) Domain {
	problemLower := strings.ToLower(problem)

	// Score each domain by counting keyword matches
	scores := make(map[Domain]int)

	for domain, keywords := range domainKeywords {
		for _, keyword := range keywords {
			if strings.Contains(problemLower, keyword) {
				scores[domain]++
			}
		}
	}

	// Find domain with highest score
	maxScore := 0
	bestDomain := DomainGeneral

	for domain, score := range scores {
		if score > maxScore {
			maxScore = score
			bestDomain = domain
		}
	}

	// Require at least 2 keyword matches to switch from general
	if maxScore < 2 {
		return DomainGeneral
	}

	return bestDomain
}

// GetDomainTemplate returns the decomposition template for a given domain
func GetDomainTemplate(domain Domain) *DomainTemplate {
	if template, ok := domainTemplates[domain]; ok {
		return template
	}
	return domainTemplates[DomainGeneral]
}

// GetAllDomains returns all available domains
func GetAllDomains() []Domain {
	return []Domain{
		DomainDebugging,
		DomainProof,
		DomainArchitecture,
		DomainResearch,
		DomainGeneral,
	}
}

// ExtractedEntities holds the entities extracted from a problem statement
type ExtractedEntities struct {
	TechnicalTerms []string `json:"technical_terms"` // OAuth 2.0, SAML, API, etc.
	Stakeholders   []string `json:"stakeholders"`    // enterprise customers, API clients, etc.
	Constraints    []string `json:"constraints"`     // backward compatibility, scalability, etc.
	KeyConcepts    []string `json:"key_concepts"`    // authentication, migration, etc.
	AllEntities    []string `json:"all_entities"`    // Combined list for easy access
	EntitySummary  string   `json:"entity_summary"`  // Brief phrase summarizing the entities
}

// ExtractProblemEntities extracts key entities from a problem statement
func ExtractProblemEntities(problem string) *ExtractedEntities {
	entities := &ExtractedEntities{
		TechnicalTerms: make([]string, 0),
		Stakeholders:   make([]string, 0),
		Constraints:    make([]string, 0),
		KeyConcepts:    make([]string, 0),
		AllEntities:    make([]string, 0),
	}

	// Extract technical terms (acronyms, versioned names, capitalized terms)
	entities.TechnicalTerms = extractTechnicalTerms(problem)

	// Extract stakeholders (users, customers, clients, teams)
	entities.Stakeholders = extractStakeholders(problem)

	// Extract constraints (compatibility, performance requirements)
	entities.Constraints = extractConstraints(problem)

	// Extract key concepts (remaining significant nouns)
	entities.KeyConcepts = extractKeyConcepts(problem, entities)

	// Combine all entities (deduplicated)
	seen := make(map[string]bool)
	for _, terms := range [][]string{entities.TechnicalTerms, entities.Stakeholders, entities.Constraints, entities.KeyConcepts} {
		for _, term := range terms {
			lower := strings.ToLower(term)
			if !seen[lower] {
				seen[lower] = true
				entities.AllEntities = append(entities.AllEntities, term)
			}
		}
	}

	// Create a summary phrase
	entities.EntitySummary = createEntitySummary(entities)

	return entities
}

// extractTechnicalTerms finds technical terms like OAuth 2.0, SAML, REST API, etc.
func extractTechnicalTerms(problem string) []string {
	terms := make([]string, 0)
	seen := make(map[string]bool)

	// Split into words while preserving some punctuation for version numbers
	words := strings.Fields(problem)

	for i := 0; i < len(words); i++ {
		word := strings.Trim(words[i], ",.;:!?()[]{}\"'")

		// Skip empty or very short words
		if len(word) < 2 {
			continue
		}

		// Pattern 1: All-caps acronyms (API, REST, SAML, SQL, etc.)
		if isAcronym(word) {
			if !seen[strings.ToUpper(word)] {
				seen[strings.ToUpper(word)] = true
				terms = append(terms, word)
			}
			continue
		}

		// Pattern 2: Versioned terms (OAuth 2.0, HTTP/2, TLS 1.3)
		if i+1 < len(words) {
			nextWord := strings.Trim(words[i+1], ",.;:!?()[]{}\"'")
			if isVersionNumber(nextWord) {
				combined := word + " " + nextWord
				if !seen[strings.ToLower(combined)] {
					seen[strings.ToLower(combined)] = true
					terms = append(terms, combined)
				}
				i++ // Skip the version number
				continue
			}
		}

		// Pattern 3: CamelCase or PascalCase terms (JavaScript, TypeScript, PostgreSQL)
		if isCamelOrPascalCase(word) && len(word) > 3 {
			if !seen[word] {
				seen[word] = true
				terms = append(terms, word)
			}
			continue
		}

		// Pattern 4: Hyphenated technical terms (micro-service, event-driven, real-time)
		if strings.Contains(word, "-") && len(word) > 5 {
			if !seen[strings.ToLower(word)] {
				seen[strings.ToLower(word)] = true
				terms = append(terms, word)
			}
			continue
		}
	}

	return terms
}

// isAcronym checks if a word is likely an acronym (2-6 uppercase letters, possibly with numbers)
func isAcronym(word string) bool {
	if len(word) < 2 || len(word) > 8 {
		return false
	}
	upperCount := 0
	for _, c := range word {
		if c >= 'A' && c <= 'Z' {
			upperCount++
		} else if c >= '0' && c <= '9' {
			// Numbers are OK in acronyms (like HTTP2)
		} else if c >= 'a' && c <= 'z' {
			return false // lowercase letters disqualify
		}
	}
	return upperCount >= 2
}

// isVersionNumber checks if a word looks like a version number (2.0, 1.3, v2, etc.)
func isVersionNumber(word string) bool {
	if len(word) < 1 || len(word) > 10 {
		return false
	}
	// Patterns: 2.0, 1.3.0, v2, v1.0
	hasDigit := false
	for _, c := range word {
		if c >= '0' && c <= '9' {
			hasDigit = true
		} else if c != '.' && c != 'v' && c != 'V' {
			return false
		}
	}
	return hasDigit
}

// isCamelOrPascalCase checks if a word has mixed case patterns typical of code
func isCamelOrPascalCase(word string) bool {
	hasUpper := false
	hasLower := false
	upperInMiddle := false

	for i, c := range word {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
			if i > 0 {
				upperInMiddle = true
			}
		} else if c >= 'a' && c <= 'z' {
			hasLower = true
		}
	}

	return hasUpper && hasLower && upperInMiddle
}

// extractStakeholders finds stakeholder references
func extractStakeholders(problem string) []string {
	stakeholders := make([]string, 0)
	seen := make(map[string]bool)
	problemLower := strings.ToLower(problem)

	// Common stakeholder patterns
	stakeholderPatterns := []string{
		"enterprise customers", "api clients", "end users", "mobile users", "web users",
		"developers", "administrators", "system administrators", "database administrators",
		"product managers", "stakeholders", "team members", "third-party", "third party",
		"internal users", "external users", "partners", "vendors", "suppliers",
		"security team", "operations team", "devops team", "support team",
		"legacy systems", "existing clients", "new clients",
	}

	for _, pattern := range stakeholderPatterns {
		if strings.Contains(problemLower, pattern) && !seen[pattern] {
			seen[pattern] = true
			stakeholders = append(stakeholders, pattern)
		}
	}

	return stakeholders
}

// extractConstraints finds constraint references
func extractConstraints(problem string) []string {
	constraints := make([]string, 0)
	seen := make(map[string]bool)
	problemLower := strings.ToLower(problem)

	// Common constraint patterns
	constraintPatterns := []string{
		"backward compatibility", "backwards compatibility", "forward compatibility",
		"high availability", "low latency", "real-time", "real time",
		"scalability", "performance", "security", "compliance",
		"maintainability", "reliability", "fault tolerance", "fault-tolerance",
		"data integrity", "consistency", "durability",
		"cost effective", "cost-effective", "budget",
		"timeline", "deadline", "legacy support", "migration",
	}

	for _, pattern := range constraintPatterns {
		if strings.Contains(problemLower, pattern) && !seen[pattern] {
			seen[pattern] = true
			constraints = append(constraints, pattern)
		}
	}

	return constraints
}

// extractKeyConcepts finds remaining key concepts (significant nouns not already captured)
func extractKeyConcepts(problem string, existing *ExtractedEntities) []string {
	concepts := make([]string, 0)
	seen := make(map[string]bool)

	// Mark existing entities as seen
	for _, term := range existing.TechnicalTerms {
		seen[strings.ToLower(term)] = true
	}
	for _, term := range existing.Stakeholders {
		seen[strings.ToLower(term)] = true
	}
	for _, term := range existing.Constraints {
		seen[strings.ToLower(term)] = true
	}

	// Domain-specific concept patterns
	conceptPatterns := []string{
		"authentication", "authorization", "access control", "single sign-on", "sso",
		"encryption", "decryption", "token", "session", "credential",
		"database", "cache", "queue", "message broker", "load balancer",
		"microservice", "monolith", "distributed system", "event-driven",
		"api gateway", "service mesh", "container", "kubernetes", "docker",
		"ci/cd", "deployment", "monitoring", "logging", "alerting",
		"data model", "schema", "migration", "integration", "synchronization",
		"workflow", "pipeline", "automation", "orchestration",
	}

	problemLower := strings.ToLower(problem)

	for _, pattern := range conceptPatterns {
		if strings.Contains(problemLower, pattern) && !seen[pattern] {
			seen[pattern] = true
			concepts = append(concepts, pattern)
		}
	}

	// Also extract significant capitalized noun phrases (likely proper nouns or product names)
	words := strings.Fields(problem)
	for i := 0; i < len(words); i++ {
		word := strings.Trim(words[i], ",.;:!?()[]{}\"'")
		wordLower := strings.ToLower(word)

		// Skip if already captured, stop word, or too short
		if seen[wordLower] || commonStopWords[wordLower] || len(word) < 3 {
			continue
		}

		// Check for capitalized words that aren't at sentence start
		if i > 0 && len(word) > 2 && word[0] >= 'A' && word[0] <= 'Z' {
			prevWord := strings.Trim(words[i-1], ",.;:!?()[]{}\"'")
			// Not after sentence-ending punctuation
			if !strings.HasSuffix(prevWord, ".") && !strings.HasSuffix(prevWord, "!") && !strings.HasSuffix(prevWord, "?") {
				seen[wordLower] = true
				concepts = append(concepts, word)
			}
		}
	}

	return concepts
}

// createEntitySummary creates a brief summary phrase of the key entities
func createEntitySummary(entities *ExtractedEntities) string {
	parts := make([]string, 0)

	// Include up to 2 technical terms
	for i, term := range entities.TechnicalTerms {
		if i >= 2 {
			break
		}
		parts = append(parts, term)
	}

	// Include up to 1 stakeholder
	for i, stake := range entities.Stakeholders {
		if i >= 1 {
			break
		}
		parts = append(parts, stake)
	}

	// Include up to 2 key concepts
	for i, concept := range entities.KeyConcepts {
		if i >= 2 {
			break
		}
		parts = append(parts, concept)
	}

	if len(parts) == 0 {
		return "the problem domain"
	}

	return strings.Join(parts, ", ")
}

// ApplyDomainTemplate applies a domain template to create subproblems
func (pd *ProblemDecomposer) ApplyDomainTemplate(problem string, domain Domain, counter int) ([]*types.Subproblem, []*types.Dependency) {
	template := GetDomainTemplate(domain)

	// Extract entities from the problem to parameterize templates
	entities := ExtractProblemEntities(problem)

	subproblems := make([]*types.Subproblem, len(template.Steps))
	for i, step := range template.Steps {
		// Parameterize the step description with extracted entities
		description := parameterizeStepDescription(step.Name, step.Description, domain, entities)

		subproblems[i] = &types.Subproblem{
			ID:          fmt.Sprintf("subproblem-%d-%d", counter, i+1),
			Description: description,
			Complexity:  step.Complexity,
			Priority:    step.Priority,
			Status:      "pending",
		}
	}

	dependencies := make([]*types.Dependency, len(template.Dependencies))
	for i, dep := range template.Dependencies {
		dependencies[i] = &types.Dependency{
			FromSubproblem: subproblems[dep.FromStep].ID,
			ToSubproblem:   subproblems[dep.ToStep].ID,
			Type:           dep.Type,
		}
	}

	return subproblems, dependencies
}

// parameterizeStepDescription creates a problem-specific description from a template step
func parameterizeStepDescription(stepName, baseDescription string, domain Domain, entities *ExtractedEntities) string {
	// If no entities extracted, return enhanced base description
	if len(entities.AllEntities) == 0 {
		return baseDescription
	}

	// Build context-aware descriptions based on step name and domain
	stepNameLower := strings.ToLower(stepName)

	// Domain-specific parameterization
	switch domain {
	case DomainArchitecture:
		return parameterizeArchitectureStep(stepNameLower, baseDescription, entities)
	case DomainDebugging:
		return parameterizeDebuggingStep(stepNameLower, baseDescription, entities)
	case DomainResearch:
		return parameterizeResearchStep(stepNameLower, baseDescription, entities)
	case DomainProof:
		return parameterizeProofStep(stepNameLower, baseDescription, entities)
	default:
		return parameterizeGeneralStep(stepNameLower, baseDescription, entities)
	}
}

// parameterizeArchitectureStep creates architecture-specific step descriptions
func parameterizeArchitectureStep(stepName, baseDescription string, entities *ExtractedEntities) string {
	techTerms := joinLimited(entities.TechnicalTerms, ", ", 3)
	stakeholders := joinLimited(entities.Stakeholders, ", ", 2)
	constraints := joinLimited(entities.Constraints, ", ", 2)
	concepts := joinLimited(entities.KeyConcepts, ", ", 2)

	switch {
	case strings.Contains(stepName, "requirements") || strings.Contains(stepName, "constraints"):
		parts := []string{baseDescription}
		if techTerms != "" {
			parts = append(parts, fmt.Sprintf("considering %s technologies", techTerms))
		}
		if stakeholders != "" {
			parts = append(parts, fmt.Sprintf("for %s", stakeholders))
		}
		if constraints != "" {
			parts = append(parts, fmt.Sprintf("with focus on %s requirements", constraints))
		}
		return strings.Join(parts, ", ")

	case strings.Contains(stepName, "components") || strings.Contains(stepName, "identify"):
		if techTerms != "" && concepts != "" {
			return fmt.Sprintf("Determine major components for %s, including %s capabilities", techTerms, concepts)
		} else if techTerms != "" {
			return fmt.Sprintf("Determine major components needed to support %s", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "interfaces") || strings.Contains(stepName, "contracts"):
		if techTerms != "" && stakeholders != "" {
			return fmt.Sprintf("Define APIs and integration points for %s to serve %s", techTerms, stakeholders)
		} else if techTerms != "" {
			return fmt.Sprintf("Define APIs, data formats, and interaction patterns for %s integration", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "trade-off") || strings.Contains(stepName, "evaluate"):
		parts := []string{"Analyze architectural trade-offs"}
		if constraints != "" {
			parts = append(parts, fmt.Sprintf("especially regarding %s", constraints))
		}
		if techTerms != "" {
			parts = append(parts, fmt.Sprintf("when integrating %s", techTerms))
		}
		return strings.Join(parts, " ")

	case strings.Contains(stepName, "document") || strings.Contains(stepName, "adr"):
		if techTerms != "" {
			return fmt.Sprintf("Create Architecture Decision Records for %s integration choices", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "roadmap") || strings.Contains(stepName, "implementation"):
		if stakeholders != "" && constraints != "" {
			return fmt.Sprintf("Plan phased implementation addressing %s needs while ensuring %s", stakeholders, constraints)
		} else if constraints != "" {
			return fmt.Sprintf("Plan phased implementation with milestones ensuring %s", constraints)
		}
		return baseDescription

	default:
		// Generic enhancement with entities
		if entities.EntitySummary != "" {
			return fmt.Sprintf("%s for %s", baseDescription, entities.EntitySummary)
		}
		return baseDescription
	}
}

// parameterizeDebuggingStep creates debugging-specific step descriptions
func parameterizeDebuggingStep(stepName, baseDescription string, entities *ExtractedEntities) string {
	techTerms := joinLimited(entities.TechnicalTerms, ", ", 3)
	concepts := joinLimited(entities.KeyConcepts, ", ", 2)

	switch {
	case strings.Contains(stepName, "reproduce"):
		if techTerms != "" {
			return fmt.Sprintf("Create a reliable reproduction case for the %s issue", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "diagnostic") || strings.Contains(stepName, "gather"):
		if techTerms != "" {
			return fmt.Sprintf("Collect logs, stack traces, and system state related to %s", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "hypothes"):
		if concepts != "" {
			return fmt.Sprintf("Generate possible root causes based on %s behavior", concepts)
		} else if techTerms != "" {
			return fmt.Sprintf("Generate hypotheses for the %s failure", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "isolate") || strings.Contains(stepName, "root cause"):
		if techTerms != "" && concepts != "" {
			return fmt.Sprintf("Systematically test hypotheses to isolate root cause in %s %s", techTerms, concepts)
		} else if techTerms != "" {
			return fmt.Sprintf("Systematically test hypotheses in %s using binary search or debugging", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "fix") || strings.Contains(stepName, "develop"):
		if techTerms != "" {
			return fmt.Sprintf("Implement the fix for %s and verify it resolves the issue", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "regression") || strings.Contains(stepName, "test"):
		if techTerms != "" {
			return fmt.Sprintf("Create regression tests for %s to prevent recurrence", techTerms)
		}
		return baseDescription

	default:
		if entities.EntitySummary != "" {
			return fmt.Sprintf("%s for %s", baseDescription, entities.EntitySummary)
		}
		return baseDescription
	}
}

// parameterizeResearchStep creates research-specific step descriptions
func parameterizeResearchStep(stepName, baseDescription string, entities *ExtractedEntities) string {
	techTerms := joinLimited(entities.TechnicalTerms, ", ", 3)
	concepts := joinLimited(entities.KeyConcepts, ", ", 2)
	constraints := joinLimited(entities.Constraints, ", ", 2)

	switch {
	case strings.Contains(stepName, "question") || strings.Contains(stepName, "define"):
		if techTerms != "" && concepts != "" {
			return fmt.Sprintf("Articulate specific research questions about %s and %s", techTerms, concepts)
		} else if techTerms != "" {
			return fmt.Sprintf("Articulate specific research questions about %s", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "literature") || strings.Contains(stepName, "review"):
		if techTerms != "" {
			return fmt.Sprintf("Survey existing work on %s, identify gaps, and establish context", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "methodology") || strings.Contains(stepName, "design"):
		if techTerms != "" && constraints != "" {
			return fmt.Sprintf("Choose research methods for studying %s considering %s", techTerms, constraints)
		} else if techTerms != "" {
			return fmt.Sprintf("Design methodology appropriate for analyzing %s", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "evidence") || strings.Contains(stepName, "gather"):
		if techTerms != "" {
			return fmt.Sprintf("Collect data, run experiments, or analyze sources related to %s", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "analyze") || strings.Contains(stepName, "findings"):
		if concepts != "" {
			return fmt.Sprintf("Process evidence to identify patterns and insights about %s", concepts)
		} else if techTerms != "" {
			return fmt.Sprintf("Analyze findings related to %s", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "synthesize") || strings.Contains(stepName, "conclusion"):
		if techTerms != "" && concepts != "" {
			return fmt.Sprintf("Draw conclusions about %s and %s that answer the research questions", techTerms, concepts)
		} else if techTerms != "" {
			return fmt.Sprintf("Synthesize conclusions about %s", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "limitation"):
		if techTerms != "" {
			return fmt.Sprintf("Acknowledge constraints and areas for future work on %s", techTerms)
		}
		return baseDescription

	default:
		if entities.EntitySummary != "" {
			return fmt.Sprintf("%s for %s", baseDescription, entities.EntitySummary)
		}
		return baseDescription
	}
}

// parameterizeProofStep creates proof-specific step descriptions
func parameterizeProofStep(stepName, baseDescription string, entities *ExtractedEntities) string {
	techTerms := joinLimited(entities.TechnicalTerms, ", ", 3)
	concepts := joinLimited(entities.KeyConcepts, ", ", 2)

	switch {
	case strings.Contains(stepName, "proposition") || strings.Contains(stepName, "state"):
		if techTerms != "" || concepts != "" {
			terms := techTerms
			if terms == "" {
				terms = concepts
			}
			return fmt.Sprintf("Formalize the claim about %s with precise definitions", terms)
		}
		return baseDescription

	case strings.Contains(stepName, "strategy") || strings.Contains(stepName, "identify"):
		if concepts != "" {
			return fmt.Sprintf("Choose proof approach for %s: direct, contradiction, or induction", concepts)
		}
		return baseDescription

	case strings.Contains(stepName, "prerequisite") || strings.Contains(stepName, "list"):
		if techTerms != "" {
			return fmt.Sprintf("Identify axioms and theorems needed for proving %s", techTerms)
		}
		return baseDescription

	default:
		if entities.EntitySummary != "" {
			return fmt.Sprintf("%s for %s", baseDescription, entities.EntitySummary)
		}
		return baseDescription
	}
}

// parameterizeGeneralStep creates general step descriptions with entity context
func parameterizeGeneralStep(stepName, baseDescription string, entities *ExtractedEntities) string {
	techTerms := joinLimited(entities.TechnicalTerms, ", ", 3)
	stakeholders := joinLimited(entities.Stakeholders, ", ", 2)
	concepts := joinLimited(entities.KeyConcepts, ", ", 2)

	switch {
	case strings.Contains(stepName, "analyze") || strings.Contains(stepName, "define") || strings.Contains(stepName, "scope"):
		parts := []string{baseDescription}
		if techTerms != "" {
			parts = append(parts, fmt.Sprintf("focusing on %s", techTerms))
		}
		if stakeholders != "" {
			parts = append(parts, fmt.Sprintf("for %s", stakeholders))
		}
		return strings.Join(parts, ", ")

	case strings.Contains(stepName, "gather") || strings.Contains(stepName, "information"):
		if techTerms != "" {
			return fmt.Sprintf("Collect data, context, and tools needed for %s", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "solution") || strings.Contains(stepName, "develop"):
		if concepts != "" && techTerms != "" {
			return fmt.Sprintf("Generate approaches for %s considering %s requirements", concepts, techTerms)
		} else if concepts != "" {
			return fmt.Sprintf("Develop potential solutions for %s", concepts)
		}
		return baseDescription

	case strings.Contains(stepName, "evaluate") || strings.Contains(stepName, "select"):
		if techTerms != "" {
			return fmt.Sprintf("Compare solutions and choose the best approach for %s", techTerms)
		}
		return baseDescription

	case strings.Contains(stepName, "implement") || strings.Contains(stepName, "test"):
		if techTerms != "" {
			return fmt.Sprintf("Execute the chosen approach for %s and verify it works", techTerms)
		}
		return baseDescription

	default:
		if entities.EntitySummary != "" {
			return fmt.Sprintf("%s for %s", baseDescription, entities.EntitySummary)
		}
		return baseDescription
	}
}

// joinLimited joins strings up to a maximum count
func joinLimited(items []string, sep string, max int) string {
	if len(items) == 0 {
		return ""
	}
	if len(items) <= max {
		return strings.Join(items, sep)
	}
	return strings.Join(items[:max], sep)
}

// DecomposeProblemWithDomain decomposes a problem using domain-specific templates
func (pd *ProblemDecomposer) DecomposeProblemWithDomain(problem string, explicitDomain *Domain) (*types.ProblemDecomposition, error) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	pd.counter++

	// Detect or use explicit domain
	var domain Domain
	if explicitDomain != nil {
		domain = *explicitDomain
	} else {
		domain = DetectDomain(problem)
	}

	// Extract entities for metadata and template parameterization
	entities := ExtractProblemEntities(problem)

	// Apply domain-specific template
	subproblems, dependencies := pd.ApplyDomainTemplate(problem, domain, pd.counter)
	solutionPath := pd.determineSolutionPath(subproblems, dependencies)

	decomposition := &types.ProblemDecomposition{
		ID:           fmt.Sprintf("decomposition-%d", pd.counter),
		Problem:      problem,
		Subproblems:  subproblems,
		Dependencies: dependencies,
		SolutionPath: solutionPath,
		Metadata: map[string]interface{}{
			"domain":           string(domain),
			"domain_detected":  explicitDomain == nil,
			"template_version": "2.0",
			"extracted_entities": map[string]interface{}{
				"technical_terms": entities.TechnicalTerms,
				"stakeholders":    entities.Stakeholders,
				"constraints":     entities.Constraints,
				"key_concepts":    entities.KeyConcepts,
				"entity_summary":  entities.EntitySummary,
			},
		},
		CreatedAt: time.Now(),
	}

	return decomposition, nil
}
