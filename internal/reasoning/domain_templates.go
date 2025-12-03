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

// ApplyDomainTemplate applies a domain template to create subproblems
func (pd *ProblemDecomposer) ApplyDomainTemplate(problem string, domain Domain, counter int) ([]*types.Subproblem, []*types.Dependency) {
	template := GetDomainTemplate(domain)

	subproblems := make([]*types.Subproblem, len(template.Steps))
	for i, step := range template.Steps {
		subproblems[i] = &types.Subproblem{
			ID:          fmt.Sprintf("subproblem-%d-%d", counter, i+1),
			Description: step.Description,
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
			"template_version": "1.0",
		},
		CreatedAt: time.Now(),
	}

	return decomposition, nil
}
