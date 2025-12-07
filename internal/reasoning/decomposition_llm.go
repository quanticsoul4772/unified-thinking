package reasoning

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"unified-thinking/internal/types"
)

// DecompositionGenerator interface for generating problem decomposition via LLM
type DecompositionGenerator interface {
	GenerateDecomposition(ctx context.Context, prompt string) (string, error)
}

// LLMProblemDecomposer uses LLM for generating context-aware problem decomposition
type LLMProblemDecomposer struct {
	generator DecompositionGenerator
	counter   int
}

// NewLLMProblemDecomposer creates a new LLM-enhanced problem decomposer
func NewLLMProblemDecomposer(generator DecompositionGenerator) *LLMProblemDecomposer {
	return &LLMProblemDecomposer{
		generator: generator,
	}
}

// DecomposeProblemWithDomain generates decomposition using LLM - fails if generator unavailable
func (pd *LLMProblemDecomposer) DecomposeProblemWithDomain(ctx context.Context, problem string, explicitDomain *Domain) (*types.ProblemDecomposition, error) {
	if problem == "" {
		return nil, fmt.Errorf("problem cannot be empty")
	}

	// No generator = fail fast, no fallback
	if pd.generator == nil {
		return nil, fmt.Errorf("LLM generator not configured - ANTHROPIC_API_KEY required")
	}

	// Detect domain if not explicitly provided
	domain := DomainGeneral
	domainDetected := true
	if explicitDomain != nil {
		domain = *explicitDomain
		domainDetected = false
	} else {
		domain = DetectDomain(problem)
	}

	// Build prompt for LLM
	prompt := pd.buildDecompositionPrompt(problem, domain)

	// Generate decomposition via LLM
	response, err := pd.generator.GenerateDecomposition(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM problem decomposition failed: %w", err)
	}

	// Parse response
	decomposition, err := pd.parseDecompositionFromLLM(response, problem, domain, domainDetected)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return decomposition, nil
}

// buildDecompositionPrompt creates an LLM prompt for generating problem decomposition
func (pd *LLMProblemDecomposer) buildDecompositionPrompt(problem string, domain Domain) string {
	domainContext := ""
	switch domain {
	case DomainDebugging:
		domainContext = `This appears to be a DEBUGGING/TROUBLESHOOTING problem. Consider:
- Symptom isolation and reproduction steps
- Diagnostic data collection approaches
- Root cause analysis techniques
- Fix development and verification strategies`
	case DomainProof:
		domainContext = `This appears to be a PROOF/VERIFICATION problem. Consider:
- Precise statement of what needs to be proven
- Required axioms, definitions, and lemmas
- Proof strategy (direct, contradiction, induction, etc.)
- Step-by-step verification approach`
	case DomainArchitecture:
		domainContext = `This appears to be an ARCHITECTURE/DESIGN problem. Consider:
- Requirements analysis and constraints identification
- Applicable architectural patterns
- Component design and interface definitions
- Trade-off analysis and risk assessment`
	case DomainResearch:
		domainContext = `This appears to be a RESEARCH/ANALYSIS problem. Consider:
- Research question formulation
- Literature review and prior art analysis
- Methodology design
- Data collection and analysis approaches`
	default:
		domainContext = `Analyze this problem comprehensively. Consider:
- Problem scope and definition clarity
- Information and resource requirements
- Solution approaches and alternatives
- Evaluation criteria and success metrics`
	}

	return fmt.Sprintf(`You are an expert problem decomposition specialist. Your task is to break down complex problems into actionable, well-structured subproblems.

Given this problem:
%s

Domain context:
%s

Generate a detailed, context-specific decomposition that includes:

1. For each subproblem:
   - A clear, specific description tied to THIS problem (not generic template text)
   - Realistic complexity assessment (low/medium/high) based on the actual work involved
   - Priority level (critical/high/medium/low) based on dependencies and impact
   - Specific considerations or approach hints relevant to THIS problem

2. Dependencies:
   - Identify which subproblems depend on others
   - Note any parallel execution opportunities

3. Solution path:
   - Recommended order of execution
   - Key decision points or milestones

IMPORTANT:
- Each subproblem description MUST be specific to the problem described
- Include domain-specific terminology and considerations
- Do NOT use generic placeholder text like "the given problem" or "the issue"
- Subproblems should be actionable and concrete

Return ONLY valid JSON in this format:
{
  "subproblems": [
    {
      "description": "Specific, actionable subproblem description...",
      "complexity": "low|medium|high",
      "priority": "critical|high|medium|low",
      "approach_hints": "Optional notes on how to approach this subproblem"
    }
  ],
  "dependencies": [
    {
      "from_index": 0,
      "to_index": 1,
      "type": "required|recommended|optional",
      "reason": "Why this dependency exists"
    }
  ],
  "parallel_opportunities": ["Indices of subproblems that can run in parallel, e.g., [1,2]"],
  "key_considerations": "Overall strategic notes for solving this problem"
}`, problem, domainContext)
}

// parseDecompositionFromLLM parses LLM JSON response into ProblemDecomposition
func (pd *LLMProblemDecomposer) parseDecompositionFromLLM(response string, problem string, domain Domain, domainDetected bool) (*types.ProblemDecomposition, error) {
	// Extract JSON from response (handle markdown code blocks)
	jsonStr := response

	// Remove markdown code blocks if present
	if idx := strings.Index(response, "```json\n"); idx >= 0 {
		start := idx + 8 // len("```json\n")
		if end := strings.Index(response[start:], "\n```"); end >= 0 {
			jsonStr = response[start : start+end]
		}
	} else if idx := strings.Index(response, "```json"); idx >= 0 {
		start := idx + 7
		if end := strings.Index(response[start:], "```"); end >= 0 {
			jsonStr = response[start : start+end]
		}
	} else if idx := strings.Index(response, "```\n"); idx >= 0 {
		start := idx + 4
		if end := strings.Index(response[start:], "\n```"); end >= 0 {
			jsonStr = response[start : start+end]
		}
	}

	jsonStr = strings.TrimSpace(jsonStr)

	// Parse JSON
	var parsed struct {
		Subproblems []struct {
			Description   string `json:"description"`
			Complexity    string `json:"complexity"`
			Priority      string `json:"priority"`
			ApproachHints string `json:"approach_hints,omitempty"`
		} `json:"subproblems"`
		Dependencies []struct {
			FromIndex int    `json:"from_index"`
			ToIndex   int    `json:"to_index"`
			Type      string `json:"type"`
			Reason    string `json:"reason,omitempty"`
		} `json:"dependencies"`
		ParallelOpportunities []interface{} `json:"parallel_opportunities,omitempty"`
		KeyConsiderations     string        `json:"key_considerations,omitempty"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %w (response: %s)", err, jsonStr)
	}

	// Convert to types.ProblemDecomposition
	pd.counter++
	now := time.Now()

	subproblems := make([]*types.Subproblem, len(parsed.Subproblems))
	for i, sp := range parsed.Subproblems {
		subproblems[i] = &types.Subproblem{
			ID:          fmt.Sprintf("subproblem-llm-%d-%d", now.UnixNano(), i+1),
			Description: sp.Description,
			Complexity:  sp.Complexity,
			Priority:    sp.Priority,
			Status:      "pending",
		}
	}

	dependencies := make([]*types.Dependency, len(parsed.Dependencies))
	for i, dep := range parsed.Dependencies {
		fromID := ""
		toID := ""
		if dep.FromIndex >= 0 && dep.FromIndex < len(subproblems) {
			fromID = subproblems[dep.FromIndex].ID
		}
		if dep.ToIndex >= 0 && dep.ToIndex < len(subproblems) {
			toID = subproblems[dep.ToIndex].ID
		}
		dependencies[i] = &types.Dependency{
			FromSubproblem: fromID,
			ToSubproblem:   toID,
			Type:           dep.Type,
		}
	}

	// Build solution path from dependencies
	solutionPath := make([]string, len(subproblems))
	for i, sp := range subproblems {
		solutionPath[i] = sp.ID
	}

	decomposition := &types.ProblemDecomposition{
		ID:           fmt.Sprintf("decomposition-llm-%d", now.UnixNano()),
		Problem:      problem,
		Subproblems:  subproblems,
		Dependencies: dependencies,
		SolutionPath: solutionPath,
		Metadata: map[string]interface{}{
			"source":             "llm",
			"domain":             string(domain),
			"domain_detected":    domainDetected,
			"key_considerations": parsed.KeyConsiderations,
		},
		CreatedAt: now,
	}

	return decomposition, nil
}

// HasGenerator returns true if an LLM generator is configured
func (pd *LLMProblemDecomposer) HasGenerator() bool {
	return pd.generator != nil
}
