package analysis

import (
	"testing"

	"unified-thinking/internal/types"
)

func TestContradictionDetector_detectContradictoryClaims(t *testing.T) {
	cd := NewContradictionDetector()

	tests := []struct {
		name      string
		thoughts  []*types.Thought
		wantCount int
	}{
		{
			name: "contradictory claims about same subject",
			thoughts: []*types.Thought{
				{
					ID:      "t1",
					Content: "The system is secure",
				},
				{
					ID:      "t2",
					Content: "The system is not secure",
				},
			},
			wantCount: 1,
		},
		{
			name: "non-contradictory claims about different subjects",
			thoughts: []*types.Thought{
				{
					ID:      "t1",
					Content: "The system is secure",
				},
				{
					ID:      "t2",
					Content: "The network is fast",
				},
			},
			wantCount: 0,
		},
		{
			name: "contradictory claims with different verbs",
			thoughts: []*types.Thought{
				{
					ID:      "t1",
					Content: "The performance increases",
				},
				{
					ID:      "t2",
					Content: "The performance decreases",
				},
			},
			wantCount: 1,
		},
		{
			name: "no common subjects",
			thoughts: []*types.Thought{
				{
					ID:      "t1",
					Content: "User authentication works",
				},
				{
					ID:      "t2",
					Content: "Database optimization is needed",
				},
			},
			wantCount: 0,
		},
		{
			name: "same subject with consistent claims",
			thoughts: []*types.Thought{
				{
					ID:      "t1",
					Content: "The system is secure",
				},
				{
					ID:      "t2",
					Content: "The system has good security",
				},
			},
			wantCount: 0,
		},
		{
			name: "multiple thoughts with multiple contradictions",
			thoughts: []*types.Thought{
				{
					ID:      "t1",
					Content: "The system is secure",
				},
				{
					ID:      "t2",
					Content: "The system is not secure",
				},
				{
					ID:      "t3",
					Content: "The system will fail",
				},
				{
					ID:      "t4",
					Content: "The system will not fail",
				},
			},
			wantCount: 2, // t1 vs t2, t3 vs t4
		},
		{
			name: "single thought - no contradictions",
			thoughts: []*types.Thought{
				{
					ID:      "t1",
					Content: "The system is secure",
				},
			},
			wantCount: 0,
		},
		{
			name:      "empty thoughts",
			thoughts:  []*types.Thought{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cd.DetectContradictions(tt.thoughts)

			if err != nil {
				t.Errorf("DetectContradictions() returned error: %v", err)
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("DetectContradictions() returned %d contradictions, want %d", len(result), tt.wantCount)
			}

			// Verify contradiction details if we expect any
			if tt.wantCount > 0 && len(result) > 0 {
				for _, contradiction := range result {
					if contradiction.ContradictoryAt == "" {
						t.Error("Contradiction type should not be empty")
					}
					if contradiction.ThoughtID1 == "" || contradiction.ThoughtID2 == "" {
						t.Error("Thought IDs should not be empty")
					}
				}
			}
		})
	}
}

func TestContradictionDetector_extractSubjects(t *testing.T) {
	cd := NewContradictionDetector()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "simple subjects with articles",
			content:  "The system works well. A user can login. An admin manages settings.",
			expected: []string{"system", "user", "admin"},
		},
		{
			name:     "no articles",
			content:  "System works. User login. Admin manages.",
			expected: []string{},
		},
		{
			name:     "short subjects filtered out",
			content:  "The a is. An it works.",
			expected: []string{},
		},
		{
			name:     "subjects with punctuation",
			content:  "The system works. A user logs in!",
			expected: []string{"system", "user"},
		},
		{
			name:     "mixed case",
			content:  "THE System works. a User logs in.",
			expected: []string{"System", "User"},
		},
		{
			name:     "empty content",
			content:  "",
			expected: []string{},
		},
		{
			name:     "single word",
			content:  "test",
			expected: []string{},
		},
		{
			name:     "complex sentence",
			content:  "The enterprise security system prevents unauthorized access while allowing legitimate users to authenticate quickly.",
			expected: []string{"enterprise", "security", "system"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cd.extractSubjects(tt.content)

			if len(result) != len(tt.expected) {
				t.Errorf("extractSubjects() length = %v, want %v", len(result), len(tt.expected))
				return
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("extractSubjects() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestContradictionDetector_hasContradictoryPredicates(t *testing.T) {
	cd := NewContradictionDetector()

	tests := []struct {
		name      string
		content1  string
		content2  string
		subject   string
		wantFound bool
	}{
		{
			name:      "is vs is not contradiction",
			content1:  "The system is secure",
			content2:  "The system is not secure",
			subject:   "system",
			wantFound: true,
		},
		{
			name:      "can vs cannot contradiction",
			content1:  "The system can handle load",
			content2:  "The system cannot handle load",
			subject:   "system",
			wantFound: true,
		},
		{
			name:      "will vs will not contradiction",
			content1:  "The system will fail",
			content2:  "The system will not fail",
			subject:   "system",
			wantFound: true,
		},
		{
			name:      "increases vs decreases contradiction",
			content1:  "Performance increases with optimization",
			content2:  "Performance decreases with load",
			subject:   "Performance",
			wantFound: true,
		},
		{
			name:      "supports vs opposes contradiction",
			content1:  "The evidence supports the hypothesis",
			content2:  "The evidence opposes the hypothesis",
			subject:   "evidence",
			wantFound: true,
		},
		{
			name:      "no contradiction",
			content1:  "The system is secure",
			content2:  "The system is efficient",
			subject:   "system",
			wantFound: false,
		},
		{
			name:      "different subjects",
			content1:  "System A is secure",
			content2:  "System B is not secure",
			subject:   "System",
			wantFound: false,
		},
		{
			name:      "complex sentences",
			content1:  "The authentication system will prevent unauthorized access",
			content2:  "The authentication system will not prevent unauthorized access",
			subject:   "authentication",
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cd.hasContradictoryPredicates(tt.content1, tt.content2, tt.subject)

			if result != tt.wantFound {
				t.Errorf("hasContradictoryPredicates() = %v, want %v", result, tt.wantFound)
			}
		})
	}
}

func TestContradictionDetector_hasContradictoryVerbs(t *testing.T) {
	cd := NewContradictionDetector()

	tests := []struct {
		name      string
		s1        string
		s2        string
		wantFound bool
	}{
		{
			name:      "is vs is not",
			s1:        "The system is working",
			s2:        "The system is not working",
			wantFound: true,
		},
		{
			name:      "can vs cannot",
			s1:        "Users can access",
			s2:        "Users cannot access",
			wantFound: true,
		},
		{
			name:      "will vs will not",
			s1:        "It will work",
			s2:        "It will not work",
			wantFound: true,
		},
		{
			name:      "should vs should not",
			s1:        "You should update",
			s2:        "You should not update",
			wantFound: true,
		},
		{
			name:      "does vs does not",
			s1:        "It does work",
			s2:        "It does not work",
			wantFound: true,
		},
		{
			name:      "has vs has not",
			s1:        "It has features",
			s2:        "It has not features",
			wantFound: true,
		},
		{
			name:      "increases vs decreases",
			s1:        "Performance increases",
			s2:        "Performance decreases",
			wantFound: true,
		},
		{
			name:      "improves vs worsens",
			s1:        "It improves speed",
			s2:        "It worsens speed",
			wantFound: true,
		},
		{
			name:      "supports vs opposes",
			s1:        "Data supports claim",
			s2:        "Data opposes claim",
			wantFound: true,
		},
		{
			name:      "no contradiction",
			s1:        "It works well",
			s2:        "It performs adequately",
			wantFound: false,
		},
		{
			name:      "same direction",
			s1:        "It improves performance",
			s2:        "It enhances performance",
			wantFound: false,
		},
		{
			name:      "empty strings",
			s1:        "",
			s2:        "",
			wantFound: false,
		},
		{
			name:      "partial match",
			s1:        "The system is operational",
			s2:        "System is not the same",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cd.hasContradictoryVerbs(tt.s1, tt.s2)

			if result != tt.wantFound {
				t.Errorf("hasContradictoryVerbs() = %v, want %v", result, tt.wantFound)
			}
		})
	}
}
