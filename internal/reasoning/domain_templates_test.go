// Package reasoning - Tests for domain-specific decomposition templates
package reasoning

import (
	"testing"

	"unified-thinking/internal/types"
)

// ============================================================================
// Domain Detection Tests
// ============================================================================

func TestDetectDomain_Debugging(t *testing.T) {
	tests := []struct {
		name     string
		problem  string
		expected Domain
	}{
		{
			name:     "explicit bug mention",
			problem:  "There's a bug causing an error in the payment processing module",
			expected: DomainDebugging,
		},
		{
			name:     "error investigation",
			problem:  "Investigate the error that occurs when users try to login",
			expected: DomainDebugging,
		},
		{
			name:     "flaky test",
			problem:  "The flaky test in CI keeps failing intermittently",
			expected: DomainDebugging,
		},
		{
			name:     "root cause analysis",
			problem:  "Find the root cause of the crash in the payment system",
			expected: DomainDebugging,
		},
		{
			name:     "troubleshooting",
			problem:  "Troubleshoot the connection issues with the database",
			expected: DomainDebugging,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectDomain(tt.problem)
			if result != tt.expected {
				t.Errorf("DetectDomain() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectDomain_Proof(t *testing.T) {
	tests := []struct {
		name     string
		problem  string
		expected Domain
	}{
		{
			name:     "prove theorem",
			problem:  "Prove that the algorithm is correct using mathematical induction",
			expected: DomainProof,
		},
		{
			name:     "formal verification",
			problem:  "Formally verify that the logic is sound and complete",
			expected: DomainProof,
		},
		{
			name:     "lemma proof",
			problem:  "Prove this lemma: for all x, if P(x) then Q(x)",
			expected: DomainProof,
		},
		{
			name:     "proof by contradiction",
			problem:  "Show by contradiction that no valid counter-example exists",
			expected: DomainProof,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectDomain(tt.problem)
			if result != tt.expected {
				t.Errorf("DetectDomain() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectDomain_Architecture(t *testing.T) {
	tests := []struct {
		name     string
		problem  string
		expected Domain
	}{
		{
			name:     "system design",
			problem:  "Design a system for handling high-volume message queuing",
			expected: DomainArchitecture,
		},
		{
			name:     "microservice architecture",
			problem:  "Architect the microservice structure for the payment platform",
			expected: DomainArchitecture,
		},
		{
			name:     "api design",
			problem:  "Design the API interface for the new component",
			expected: DomainArchitecture,
		},
		{
			name:     "refactoring",
			problem:  "Plan the refactor of the monolith into separate modules",
			expected: DomainArchitecture,
		},
		{
			name:     "scaling",
			problem:  "Design a scalable infrastructure for the application",
			expected: DomainArchitecture,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectDomain(tt.problem)
			if result != tt.expected {
				t.Errorf("DetectDomain() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectDomain_Research(t *testing.T) {
	tests := []struct {
		name     string
		problem  string
		expected Domain
	}{
		{
			name:     "research question",
			problem:  "Research and analyze the best practices for implementing caching strategies",
			expected: DomainResearch,
		},
		{
			name:     "literature survey",
			problem:  "Survey the literature on distributed consensus algorithms",
			expected: DomainResearch,
		},
		{
			name:     "comparative analysis",
			problem:  "Compare and evaluate different database options for our use case",
			expected: DomainResearch,
		},
		{
			name:     "benchmark study",
			problem:  "Conduct a benchmark study to analyze performance metrics",
			expected: DomainResearch,
		},
		{
			name:     "methodology exploration",
			problem:  "Explore and investigate different methodologies for testing distributed systems",
			expected: DomainResearch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectDomain(tt.problem)
			if result != tt.expected {
				t.Errorf("DetectDomain() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectDomain_General(t *testing.T) {
	tests := []struct {
		name     string
		problem  string
		expected Domain
	}{
		{
			name:     "single keyword insufficient",
			problem:  "Fix the button color",
			expected: DomainGeneral, // Only one keyword "fix", needs at least 2
		},
		{
			name:     "no domain keywords",
			problem:  "How do I implement a feature?",
			expected: DomainGeneral,
		},
		{
			name:     "ambiguous single word",
			problem:  "Create something new",
			expected: DomainGeneral,
		},
		{
			name:     "empty problem",
			problem:  "",
			expected: DomainGeneral,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectDomain(tt.problem)
			if result != tt.expected {
				t.Errorf("DetectDomain() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDetectDomain_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		problem  string
		expected Domain
	}{
		{
			name:     "uppercase debugging",
			problem:  "DEBUG the ERROR in the stack TRACE",
			expected: DomainDebugging,
		},
		{
			name:     "mixed case proof",
			problem:  "PROVE the THEOREM using mathematical INDUCTION",
			expected: DomainProof,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectDomain(tt.problem)
			if result != tt.expected {
				t.Errorf("DetectDomain() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Domain Template Tests
// ============================================================================

func TestGetDomainTemplate_AllDomains(t *testing.T) {
	domains := GetAllDomains()

	for _, domain := range domains {
		t.Run(string(domain), func(t *testing.T) {
			template := GetDomainTemplate(domain)
			if template == nil {
				t.Fatalf("GetDomainTemplate(%s) returned nil", domain)
			}

			if template.Domain != domain {
				t.Errorf("Template domain = %v, want %v", template.Domain, domain)
			}

			if template.Description == "" {
				t.Error("Template description is empty")
			}

			if len(template.Steps) == 0 {
				t.Error("Template has no steps")
			}

			// Verify each step has required fields
			for i, step := range template.Steps {
				if step.Name == "" {
					t.Errorf("Step %d has empty name", i)
				}
				if step.Description == "" {
					t.Errorf("Step %d has empty description", i)
				}
				if step.Complexity == "" {
					t.Errorf("Step %d has empty complexity", i)
				}
				if step.Priority == "" {
					t.Errorf("Step %d has empty priority", i)
				}
			}
		})
	}
}

func TestGetDomainTemplate_InvalidDomain(t *testing.T) {
	template := GetDomainTemplate(Domain("invalid"))
	if template == nil {
		t.Fatal("GetDomainTemplate for invalid domain should return general template, not nil")
	}
	if template.Domain != DomainGeneral {
		t.Errorf("GetDomainTemplate for invalid domain should return general template, got %s", template.Domain)
	}
}

func TestGetDomainTemplate_DebuggingSteps(t *testing.T) {
	template := GetDomainTemplate(DomainDebugging)

	expectedSteps := []string{
		"Reproduce the issue",
		"Gather diagnostic information",
		"Formulate hypotheses",
		"Isolate the root cause",
		"Develop and test fix",
		"Add regression tests",
	}

	if len(template.Steps) != len(expectedSteps) {
		t.Errorf("Debugging template has %d steps, expected %d", len(template.Steps), len(expectedSteps))
	}

	for i, step := range template.Steps {
		if i < len(expectedSteps) && step.Name != expectedSteps[i] {
			t.Errorf("Step %d name = %s, expected %s", i, step.Name, expectedSteps[i])
		}
	}
}

func TestGetDomainTemplate_ProofSteps(t *testing.T) {
	template := GetDomainTemplate(DomainProof)

	if len(template.Steps) != 7 {
		t.Errorf("Proof template has %d steps, expected 7", len(template.Steps))
	}

	// Verify first and last steps
	if template.Steps[0].Name != "State the proposition clearly" {
		t.Errorf("First step = %s, expected 'State the proposition clearly'", template.Steps[0].Name)
	}
	if template.Steps[6].Name != "Review completeness" {
		t.Errorf("Last step = %s, expected 'Review completeness'", template.Steps[6].Name)
	}
}

func TestGetDomainTemplate_Dependencies(t *testing.T) {
	tests := []struct {
		domain          Domain
		expectedDeps    int
		hasParallelDeps bool
	}{
		{DomainDebugging, 5, false},   // All sequential
		{DomainProof, 7, true},        // Has parallel deps (steps 5,6 after 4)
		{DomainArchitecture, 6, true}, // Has parallel deps (steps 1,3 parallel)
		{DomainResearch, 7, false},    // Has optional dep but all sequential
		{DomainGeneral, 4, false},     // All sequential
	}

	for _, tt := range tests {
		t.Run(string(tt.domain), func(t *testing.T) {
			template := GetDomainTemplate(tt.domain)

			if len(template.Dependencies) != tt.expectedDeps {
				t.Errorf("Domain %s has %d dependencies, expected %d", tt.domain, len(template.Dependencies), tt.expectedDeps)
			}

			hasParallel := false
			for _, dep := range template.Dependencies {
				if dep.Type == "parallel" {
					hasParallel = true
					break
				}
			}

			if hasParallel != tt.hasParallelDeps {
				t.Errorf("Domain %s hasParallelDeps = %v, expected %v", tt.domain, hasParallel, tt.hasParallelDeps)
			}
		})
	}
}

// ============================================================================
// GetAllDomains Tests
// ============================================================================

func TestGetAllDomains(t *testing.T) {
	domains := GetAllDomains()

	if len(domains) != 5 {
		t.Errorf("GetAllDomains() returned %d domains, expected 5", len(domains))
	}

	expected := map[Domain]bool{
		DomainDebugging:    true,
		DomainProof:        true,
		DomainArchitecture: true,
		DomainResearch:     true,
		DomainGeneral:      true,
	}

	for _, domain := range domains {
		if !expected[domain] {
			t.Errorf("Unexpected domain: %s", domain)
		}
		delete(expected, domain)
	}

	for domain := range expected {
		t.Errorf("Missing domain: %s", domain)
	}
}

// ============================================================================
// ApplyDomainTemplate Tests
// ============================================================================

func TestApplyDomainTemplate(t *testing.T) {
	pd := NewProblemDecomposer()

	tests := []struct {
		domain           Domain
		expectedSubCount int
		expectedDepCount int
	}{
		{DomainDebugging, 6, 5},
		{DomainProof, 7, 7},
		{DomainArchitecture, 6, 6},
		{DomainResearch, 7, 7},
		{DomainGeneral, 5, 4},
	}

	for _, tt := range tests {
		t.Run(string(tt.domain), func(t *testing.T) {
			subproblems, dependencies := pd.ApplyDomainTemplate("Test problem", tt.domain, 1)

			if len(subproblems) != tt.expectedSubCount {
				t.Errorf("ApplyDomainTemplate returned %d subproblems, expected %d", len(subproblems), tt.expectedSubCount)
			}

			if len(dependencies) != tt.expectedDepCount {
				t.Errorf("ApplyDomainTemplate returned %d dependencies, expected %d", len(dependencies), tt.expectedDepCount)
			}

			// Verify subproblem IDs are unique
			idsSeen := make(map[string]bool)
			for _, sub := range subproblems {
				if idsSeen[sub.ID] {
					t.Errorf("Duplicate subproblem ID: %s", sub.ID)
				}
				idsSeen[sub.ID] = true
			}

			// Verify dependencies reference valid subproblems
			for _, dep := range dependencies {
				foundFrom := false
				foundTo := false
				for _, sub := range subproblems {
					if sub.ID == dep.FromSubproblem {
						foundFrom = true
					}
					if sub.ID == dep.ToSubproblem {
						foundTo = true
					}
				}
				if !foundFrom {
					t.Errorf("Dependency references unknown FromSubproblem: %s", dep.FromSubproblem)
				}
				if !foundTo {
					t.Errorf("Dependency references unknown ToSubproblem: %s", dep.ToSubproblem)
				}
			}
		})
	}
}

// ============================================================================
// DecomposeProblemWithDomain Tests
// ============================================================================

func TestDecomposeProblemWithDomain_AutoDetect(t *testing.T) {
	pd := NewProblemDecomposer()

	tests := []struct {
		name           string
		problem        string
		expectedDomain Domain
	}{
		{
			name:           "debugging problem auto-detected",
			problem:        "Debug the error that causes the crash in production",
			expectedDomain: DomainDebugging,
		},
		{
			name:           "architecture problem auto-detected",
			problem:        "Design a scalable microservice architecture for the API",
			expectedDomain: DomainArchitecture,
		},
		{
			name:           "research problem auto-detected",
			problem:        "Research and analyze the best database options",
			expectedDomain: DomainResearch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decomposition, err := pd.DecomposeProblemWithDomain(tt.problem, nil)
			if err != nil {
				t.Fatalf("DecomposeProblemWithDomain() error = %v", err)
			}

			if decomposition == nil {
				t.Fatal("DecomposeProblemWithDomain() returned nil decomposition")
			}

			// Check metadata for domain
			domainStr, ok := decomposition.Metadata["domain"].(string)
			if !ok {
				t.Fatal("Metadata does not contain domain")
			}

			if Domain(domainStr) != tt.expectedDomain {
				t.Errorf("Detected domain = %s, expected %s", domainStr, tt.expectedDomain)
			}

			// Verify domain_detected is true (auto-detection)
			domainDetected, ok := decomposition.Metadata["domain_detected"].(bool)
			if !ok {
				t.Fatal("Metadata does not contain domain_detected")
			}
			if !domainDetected {
				t.Error("domain_detected should be true for auto-detection")
			}
		})
	}
}

func TestDecomposeProblemWithDomain_ExplicitDomain(t *testing.T) {
	pd := NewProblemDecomposer()

	// Provide explicit domain that overrides detection
	domain := DomainProof
	decomposition, err := pd.DecomposeProblemWithDomain(
		"Fix the bug in the code", // Would normally detect as debugging
		&domain,
	)

	if err != nil {
		t.Fatalf("DecomposeProblemWithDomain() error = %v", err)
	}

	// Check that explicit domain was used
	domainStr, ok := decomposition.Metadata["domain"].(string)
	if !ok {
		t.Fatal("Metadata does not contain domain")
	}

	if Domain(domainStr) != DomainProof {
		t.Errorf("Domain = %s, expected proof (explicit override)", domainStr)
	}

	// Verify domain_detected is false (explicit domain)
	domainDetected, ok := decomposition.Metadata["domain_detected"].(bool)
	if !ok {
		t.Fatal("Metadata does not contain domain_detected")
	}
	if domainDetected {
		t.Error("domain_detected should be false for explicit domain")
	}

	// Verify we got proof template steps (7 steps)
	if len(decomposition.Subproblems) != 7 {
		t.Errorf("Got %d subproblems, expected 7 for proof template", len(decomposition.Subproblems))
	}
}

func TestDecomposeProblemWithDomain_SubproblemContent(t *testing.T) {
	pd := NewProblemDecomposer()

	domain := DomainDebugging
	decomposition, err := pd.DecomposeProblemWithDomain("Test problem", &domain)

	if err != nil {
		t.Fatalf("DecomposeProblemWithDomain() error = %v", err)
	}

	// Verify subproblems have proper status
	for _, sub := range decomposition.Subproblems {
		if sub.Status != "pending" {
			t.Errorf("Subproblem %s has status %s, expected pending", sub.ID, sub.Status)
		}
	}

	// Verify solution path is populated
	if len(decomposition.SolutionPath) == 0 {
		t.Error("SolutionPath is empty")
	}

	// Verify timestamps
	if decomposition.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestDecomposeProblemWithDomain_UniqueIDs(t *testing.T) {
	pd := NewProblemDecomposer()

	// Create multiple decompositions to verify unique IDs
	var decompositions [3]*types.ProblemDecomposition
	for i := 0; i < 3; i++ {
		d, err := pd.DecomposeProblemWithDomain("Problem "+string(rune('A'+i)), nil)
		if err != nil {
			t.Fatalf("DecomposeProblemWithDomain() error = %v", err)
		}
		decompositions[i] = d
	}

	// Check all decomposition IDs are unique
	ids := make(map[string]bool)
	for _, d := range decompositions {
		if ids[d.ID] {
			t.Errorf("Duplicate decomposition ID: %s", d.ID)
		}
		ids[d.ID] = true

		// Also check subproblem IDs across decompositions
		for _, sub := range d.Subproblems {
			if ids[sub.ID] {
				t.Errorf("Duplicate subproblem ID: %s", sub.ID)
			}
			ids[sub.ID] = true
		}
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

func TestDetectDomain_MixedDomains(t *testing.T) {
	// When multiple domains have equal scores, the first encountered wins
	// (due to Go map iteration order being non-deterministic, but the algorithm
	// picks based on score, so equal scores may vary)
	problem := "Debug and design the architecture for the new research system"

	domain := DetectDomain(problem)

	// Should detect one of the domains (at least 2 keywords match required)
	validDomains := []Domain{DomainDebugging, DomainArchitecture, DomainResearch}
	found := false
	for _, valid := range validDomains {
		if domain == valid {
			found = true
			break
		}
	}

	if !found && domain != DomainGeneral {
		t.Errorf("DetectDomain() returned unexpected domain: %s", domain)
	}
}

func TestDependencyTypes(t *testing.T) {
	validTypes := map[string]bool{
		"required": true,
		"optional": true,
		"parallel": true,
	}

	domains := GetAllDomains()
	for _, domain := range domains {
		template := GetDomainTemplate(domain)
		for i, dep := range template.Dependencies {
			if !validTypes[dep.Type] {
				t.Errorf("Domain %s dependency %d has invalid type: %s", domain, i, dep.Type)
			}
		}
	}
}

func TestComplexityAndPriorityValues(t *testing.T) {
	validComplexity := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}

	validPriority := map[string]bool{
		"low":      true,
		"medium":   true,
		"high":     true,
		"critical": true,
	}

	domains := GetAllDomains()
	for _, domain := range domains {
		template := GetDomainTemplate(domain)
		for i, step := range template.Steps {
			if !validComplexity[step.Complexity] {
				t.Errorf("Domain %s step %d has invalid complexity: %s", domain, i, step.Complexity)
			}
			if !validPriority[step.Priority] {
				t.Errorf("Domain %s step %d has invalid priority: %s", domain, i, step.Priority)
			}
		}
	}
}
