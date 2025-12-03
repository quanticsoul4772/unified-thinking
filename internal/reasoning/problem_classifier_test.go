package reasoning

import (
	"testing"
)

func TestNewProblemClassifier(t *testing.T) {
	pc := NewProblemClassifier()
	if pc == nil {
		t.Error("NewProblemClassifier returned nil")
	}
}

func TestClassifyProblem_MetaQuestion(t *testing.T) {
	pc := NewProblemClassifier()

	tests := []struct {
		name          string
		problem       string
		expectedType  ProblemType
		minConfidence float64
	}{
		{
			name:          "meta question about problems that resist analysis",
			problem:       "What problems resist formal analysis and decomposition?",
			expectedType:  ProblemTypeMeta,
			minConfidence: 0.8,
		},
		{
			name:          "meta question about questions that cannot be answered",
			problem:       "What questions cannot be answered through systematic reasoning?",
			expectedType:  ProblemTypeMeta,
			minConfidence: 0.8,
		},
		{
			name:          "meta question about situations that fail",
			problem:       "Which situations fail to benefit from structured analysis?",
			expectedType:  ProblemTypeMeta,
			minConfidence: 0.8,
		},
		{
			name:          "meta question about approaches that don't work",
			problem:       "What approaches are not suitable for this methodology?",
			expectedType:  ProblemTypeMeta,
			minConfidence: 0.8,
		},
		{
			name:          "meta question about limits of decomposition",
			problem:       "What are the limits of problem decomposition methods?",
			expectedType:  ProblemTypeMeta,
			minConfidence: 0.8,
		},
		{
			name:          "meta question about what cannot be reasoned",
			problem:       "Which methods don't work for informal reasoning?",
			expectedType:  ProblemTypeMeta,
			minConfidence: 0.8,
		},
		{
			name:          "meta question with won't keyword",
			problem:       "Which questions won't respond to systematic analysis?",
			expectedType:  ProblemTypeMeta,
			minConfidence: 0.8,
		},
		{
			name:          "meta question about beyond limits",
			problem:       "What problems are beyond the limits of formal reasoning?",
			expectedType:  ProblemTypeMeta,
			minConfidence: 0.8,
		},
		{
			name:          "meta question about unsuitable cases",
			problem:       "Which cases are unsuitable for this decomposition approach?",
			expectedType:  ProblemTypeMeta,
			minConfidence: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pc.ClassifyProblem(tt.problem)

			if result.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, result.Type)
			}

			if result.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %.2f, got %.2f", tt.minConfidence, result.Confidence)
			}

			if len(result.Indicators) == 0 {
				t.Error("Expected at least one indicator")
			}

			if result.Reasoning == "" {
				t.Error("Expected non-empty reasoning")
			}

			if result.Approach == "" {
				t.Error("Expected non-empty approach")
			}
		})
	}
}

func TestClassifyProblem_EmotionalProblem(t *testing.T) {
	pc := NewProblemClassifier()

	tests := []struct {
		name          string
		problem       string
		expectedType  ProblemType
		minConfidence float64
	}{
		{
			name:          "personal emotional decision",
			problem:       "Should I forgive my friend who betrayed my trust?",
			expectedType:  ProblemTypeEmotional,
			minConfidence: 0.8,
		},
		{
			name:          "grief processing",
			problem:       "How do I cope with grief after losing my father?",
			expectedType:  ProblemTypeEmotional,
			minConfidence: 0.8,
		},
		{
			name:          "emotional state with personal pronoun",
			problem:       "I feel hurt and angry about the betrayal",
			expectedType:  ProblemTypeEmotional,
			minConfidence: 0.8,
		},
		{
			name:          "relationship issue with guilt",
			problem:       "I feel guilt and shame about my actions in the relationship",
			expectedType:  ProblemTypeEmotional,
			minConfidence: 0.8,
		},
		{
			name:          "personal worry with decision",
			problem:       "Should I worry about my future if I feel anxious and sad?",
			expectedType:  ProblemTypeEmotional,
			minConfidence: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pc.ClassifyProblem(tt.problem)

			if result.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, result.Type)
			}

			if result.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %.2f, got %.2f", tt.minConfidence, result.Confidence)
			}
		})
	}
}

func TestClassifyProblem_ValueConflict(t *testing.T) {
	pc := NewProblemClassifier()

	tests := []struct {
		name          string
		problem       string
		expectedType  ProblemType
		minConfidence float64
	}{
		{
			name:          "privacy vs security",
			problem:       "How do we balance privacy vs security in our application?",
			expectedType:  ProblemTypeValueConflict,
			minConfidence: 0.8,
		},
		{
			name:          "efficiency versus fairness",
			problem:       "We need to choose between efficiency versus fairness in our hiring process",
			expectedType:  ProblemTypeValueConflict,
			minConfidence: 0.8,
		},
		{
			name:          "growth vs sustainability",
			problem:       "The company faces growth vs sustainability challenges",
			expectedType:  ProblemTypeValueConflict,
			minConfidence: 0.8,
		},
		{
			name:          "profit vs ethics",
			problem:       "There's tension between profit vs ethics in this decision",
			expectedType:  ProblemTypeValueConflict,
			minConfidence: 0.8,
		},
		{
			name:          "individual vs collective",
			problem:       "Should we prioritize individual vs collective needs?",
			expectedType:  ProblemTypeValueConflict,
			minConfidence: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pc.ClassifyProblem(tt.problem)

			if result.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, result.Type)
			}

			if result.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %.2f, got %.2f", tt.minConfidence, result.Confidence)
			}
		})
	}
}

func TestClassifyProblem_CreativeProblem(t *testing.T) {
	pc := NewProblemClassifier()

	tests := []struct {
		name          string
		problem       string
		expectedType  ProblemType
		minConfidence float64
	}{
		{
			name:          "novel solution request",
			problem:       "We need a new and novel approach to this problem that doesn't exist yet",
			expectedType:  ProblemTypeCreative,
			minConfidence: 0.8,
		},
		{
			name:          "innovative design",
			problem:       "Design an innovative and original product for this market",
			expectedType:  ProblemTypeCreative,
			minConfidence: 0.8,
		},
		{
			name:          "creative from scratch",
			problem:       "Create from scratch a creative solution for this unprecedented challenge",
			expectedType:  ProblemTypeCreative,
			minConfidence: 0.8,
		},
		{
			name:          "unprecedented situation",
			problem:       "This situation has never existed before and needs a new approach",
			expectedType:  ProblemTypeCreative,
			minConfidence: 0.8,
		},
		{
			name:          "invent something new",
			problem:       "We need to invent a novel solution for this",
			expectedType:  ProblemTypeCreative,
			minConfidence: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pc.ClassifyProblem(tt.problem)

			if result.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, result.Type)
			}

			if result.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %.2f, got %.2f", tt.minConfidence, result.Confidence)
			}
		})
	}
}

func TestClassifyProblem_TacitKnowledge(t *testing.T) {
	pc := NewProblemClassifier()

	tests := []struct {
		name          string
		problem       string
		expectedType  ProblemType
		minConfidence float64
	}{
		{
			name:          "when to bail out",
			problem:       "When should I bail out of this project?",
			expectedType:  ProblemTypeTacit,
			minConfidence: 0.7,
		},
		{
			name:          "gut feeling question",
			problem:       "When should one trust gut feeling and instinct in judgment?",
			expectedType:  ProblemTypeTacit,
			minConfidence: 0.7,
		},
		{
			name:          "expert intuition",
			problem:       "When to double down based on expert intuition?",
			expectedType:  ProblemTypeTacit,
			minConfidence: 0.7,
		},
		{
			name:          "judgment call",
			problem:       "When to cut losses requires good judgment",
			expectedType:  ProblemTypeTacit,
			minConfidence: 0.7,
		},
		{
			name:          "sense and instinct",
			problem:       "How do I develop a sense for when to pivot based on instinct?",
			expectedType:  ProblemTypeTacit,
			minConfidence: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pc.ClassifyProblem(tt.problem)

			if result.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, result.Type)
			}

			if result.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %.2f, got %.2f", tt.minConfidence, result.Confidence)
			}
		})
	}
}

func TestClassifyProblem_ChaoticProblem(t *testing.T) {
	pc := NewProblemClassifier()

	tests := []struct {
		name          string
		problem       string
		expectedType  ProblemType
		minConfidence float64
	}{
		{
			name:          "stock market prediction",
			problem:       "Can we predict the stock market movements?",
			expectedType:  ProblemTypeChaotic,
			minConfidence: 0.7,
		},
		{
			name:          "weather forecast",
			problem:       "How to forecast weather patterns in chaotic systems?",
			expectedType:  ProblemTypeChaotic,
			minConfidence: 0.7,
		},
		{
			name:          "volatile market",
			problem:       "Can we anticipate volatile market changes?",
			expectedType:  ProblemTypeChaotic,
			minConfidence: 0.7,
		},
		{
			name:          "emergent system prediction",
			problem:       "How to predict emergent behavior in complex systems?",
			expectedType:  ProblemTypeChaotic,
			minConfidence: 0.7,
		},
		{
			name:          "black swan events",
			problem:       "Can we forecast black swan events in the market?",
			expectedType:  ProblemTypeChaotic,
			minConfidence: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pc.ClassifyProblem(tt.problem)

			if result.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, result.Type)
			}

			if result.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %.2f, got %.2f", tt.minConfidence, result.Confidence)
			}
		})
	}
}

func TestClassifyProblem_Decomposable(t *testing.T) {
	pc := NewProblemClassifier()

	tests := []struct {
		name          string
		problem       string
		expectedType  ProblemType
		minConfidence float64
	}{
		{
			name:          "simple technical problem",
			problem:       "How do we implement a REST API for user management?",
			expectedType:  ProblemTypeDecomposable,
			minConfidence: 0.7,
		},
		{
			name:          "optimization problem",
			problem:       "Optimize the database queries for better performance",
			expectedType:  ProblemTypeDecomposable,
			minConfidence: 0.7,
		},
		{
			name:          "straightforward task",
			problem:       "Write unit tests for the calculator module",
			expectedType:  ProblemTypeDecomposable,
			minConfidence: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pc.ClassifyProblem(tt.problem)

			if result.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s", tt.expectedType, result.Type)
			}

			if result.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %.2f, got %.2f", tt.minConfidence, result.Confidence)
			}
		})
	}
}

func TestDetectMetaQuestion_NoMatch(t *testing.T) {
	pc := NewProblemClassifier()

	// These should not be detected as meta questions
	problems := []string{
		"How do I implement a function?",
		"What is the best database for this?",
		"Which framework should I use?",
		"What are the problems with this code?", // No resistance/limit keywords
	}

	for _, problem := range problems {
		result := pc.detectMetaQuestion(problem)
		if result != nil {
			t.Errorf("Expected nil for problem %q, got type %s", problem, result.Type)
		}
	}
}

func TestDetectEmotionalProblem_NoMatch(t *testing.T) {
	pc := NewProblemClassifier()

	// These should not be detected as emotional problems
	problems := []string{
		"How do we implement this feature?",      // No personal pronoun
		"The team should decide on architecture", // No emotional terms
		"We need better performance",             // No emotional context
	}

	for _, problem := range problems {
		result := pc.detectEmotionalProblem(problem)
		if result != nil {
			t.Errorf("Expected nil for problem %q, got type %s", problem, result.Type)
		}
	}
}

func TestDetectValueConflict_NoMatch(t *testing.T) {
	pc := NewProblemClassifier()

	// These should not be detected as value conflicts
	problems := []string{
		"How do we improve security?",          // Single value, no vs
		"Balance the load across servers",      // Not a value conflict
		"Choose between option A and option B", // No "vs" or "versus"
	}

	for _, problem := range problems {
		result := pc.detectValueConflict(problem)
		if result != nil {
			t.Errorf("Expected nil for problem %q, got type %s", problem, result.Type)
		}
	}
}

func TestDetectCreativeProblem_NoMatch(t *testing.T) {
	pc := NewProblemClassifier()

	// These should not be detected as creative problems
	problems := []string{
		"Implement the existing design",
		"Follow the established pattern",
		"Fix the bug in the code",
	}

	for _, problem := range problems {
		result := pc.detectCreativeProblem(problem)
		if result != nil {
			t.Errorf("Expected nil for problem %q, got type %s", problem, result.Type)
		}
	}
}

func TestDetectCreativeProblem_TechnicalContextsExcluded(t *testing.T) {
	pc := NewProblemClassifier()

	// Technical problems with "design" or "new" should NOT be classified as creative
	// They should fall through to decomposable for domain-aware decomposition
	tests := []struct {
		name    string
		problem string
	}{
		{
			name:    "architecture with design keyword",
			problem: "Design the microservice architecture and API interfaces for the new payment system",
		},
		{
			name:    "system design",
			problem: "Design a new system for user authentication",
		},
		{
			name:    "component design",
			problem: "Design new component interfaces for the module",
		},
		{
			name:    "debugging with new",
			problem: "Debug the new error in the crash trace",
		},
		{
			name:    "research methodology",
			problem: "Design new research methodology to analyze the data",
		},
		{
			name:    "infrastructure design",
			problem: "Design new infrastructure for the service deployment",
		},
		{
			name:    "API design",
			problem: "Design a new API interface for the integration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pc.detectCreativeProblem(tt.problem)
			if result != nil {
				t.Errorf("Expected nil (not creative) for technical problem %q, got type %s", tt.problem, result.Type)
			}
		})
	}
}

func TestClassifyProblem_TechnicalWithDesignKeyword(t *testing.T) {
	pc := NewProblemClassifier()

	// These technical problems should be classified as decomposable, not creative
	tests := []struct {
		name          string
		problem       string
		expectedType  ProblemType
		minConfidence float64
	}{
		{
			name:          "architecture design is decomposable",
			problem:       "Design the microservice architecture and API interfaces for the new payment system",
			expectedType:  ProblemTypeDecomposable,
			minConfidence: 0.7,
		},
		{
			name:          "system design is decomposable",
			problem:       "Design a new authentication system with component interfaces",
			expectedType:  ProblemTypeDecomposable,
			minConfidence: 0.7,
		},
		{
			name:          "api interface design is decomposable",
			problem:       "Design new API interfaces for the service integration",
			expectedType:  ProblemTypeDecomposable,
			minConfidence: 0.7,
		},
		{
			name:          "module design is decomposable",
			problem:       "Design new module with component architecture",
			expectedType:  ProblemTypeDecomposable,
			minConfidence: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pc.ClassifyProblem(tt.problem)

			if result.Type != tt.expectedType {
				t.Errorf("Expected type %s, got %s for problem: %q", tt.expectedType, result.Type, tt.problem)
			}

			if result.Confidence < tt.minConfidence {
				t.Errorf("Expected confidence >= %.2f, got %.2f", tt.minConfidence, result.Confidence)
			}
		})
	}
}

func TestDetectTacitKnowledge_NoMatch(t *testing.T) {
	pc := NewProblemClassifier()

	// These should not be detected as tacit knowledge problems
	problems := []string{
		"What is the syntax for a function?",
		"How to configure the database?",
		"Define the API endpoint",
	}

	for _, problem := range problems {
		result := pc.detectTacitKnowledge(problem)
		if result != nil {
			t.Errorf("Expected nil for problem %q, got type %s", problem, result.Type)
		}
	}
}

func TestClassificationResult_Structure(t *testing.T) {
	pc := NewProblemClassifier()

	result := pc.ClassifyProblem("How do we implement a feature?")

	// Verify all fields are populated
	if result.Type == "" {
		t.Error("Type should not be empty")
	}

	if result.Confidence <= 0 || result.Confidence > 1 {
		t.Errorf("Confidence should be between 0 and 1, got %.2f", result.Confidence)
	}

	if result.Reasoning == "" {
		t.Error("Reasoning should not be empty")
	}

	if len(result.Indicators) == 0 {
		t.Error("Should have at least one indicator")
	}

	if result.Approach == "" {
		t.Error("Approach should not be empty")
	}
}

func TestClassifyProblem_EdgeCases(t *testing.T) {
	pc := NewProblemClassifier()

	tests := []struct {
		name       string
		problem    string
		expectType ProblemType
	}{
		{
			name:       "empty string defaults to decomposable",
			problem:    "",
			expectType: ProblemTypeDecomposable,
		},
		{
			name:       "single word defaults to decomposable",
			problem:    "test",
			expectType: ProblemTypeDecomposable,
		},
		{
			name:       "special characters",
			problem:    "!@#$%^&*()",
			expectType: ProblemTypeDecomposable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pc.ClassifyProblem(tt.problem)
			if result.Type != tt.expectType {
				t.Errorf("Expected type %s, got %s", tt.expectType, result.Type)
			}
		})
	}
}

func TestClassifyProblem_CaseInsensitivity(t *testing.T) {
	pc := NewProblemClassifier()

	// Same problem in different cases should give same classification
	problems := []string{
		"How do we balance PRIVACY vs SECURITY?",
		"how do we balance privacy vs security?",
		"HOW DO WE BALANCE Privacy VS Security?",
	}

	var firstType ProblemType
	for i, problem := range problems {
		result := pc.ClassifyProblem(problem)
		if i == 0 {
			firstType = result.Type
		} else if result.Type != firstType {
			t.Errorf("Case sensitivity issue: got %s for %q, expected %s", result.Type, problem, firstType)
		}
	}
}
