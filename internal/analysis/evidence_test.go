package analysis

import (
	"testing"

	"unified-thinking/internal/types"
)

func TestNewEvidenceAnalyzer(t *testing.T) {
	ea := NewEvidenceAnalyzer()
	if ea == nil {
		t.Fatal("NewEvidenceAnalyzer returned nil")
	}
}

func TestAssessEvidence(t *testing.T) {
	ea := NewEvidenceAnalyzer()

	tests := []struct {
		name          string
		content       string
		source        string
		claimID       string
		supportsClaim bool
		wantQuality   types.EvidenceQuality
	}{
		{
			name:          "strong evidence",
			content:       "A peer-reviewed study shows that this approach is effective based on statistical analysis of 1000 participants",
			source:        "Journal of Research",
			claimID:       "claim-1",
			supportsClaim: true,
			wantQuality:   types.EvidenceStrong,
		},
		{
			name:          "weak evidence",
			content:       "I heard that this might work",
			source:        "Unknown",
			claimID:       "claim-1",
			supportsClaim: true,
			wantQuality:   types.EvidenceWeak,
		},
		{
			name:          "anecdotal evidence",
			content:       "This is just an anecdote and my opinion that maybe this could help",
			source:        "Personal anecdote",
			claimID:       "claim-1",
			supportsClaim: true,
			wantQuality:   types.EvidenceAnecdotal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evidence, err := ea.AssessEvidence(tt.content, tt.source, tt.claimID, tt.supportsClaim)
			if err != nil {
				t.Errorf("AssessEvidence() error = %v", err)
				return
			}
			if evidence == nil {
				t.Fatal("AssessEvidence() returned nil evidence")
			}
			if evidence.Quality != tt.wantQuality {
				t.Errorf("Quality = %v, want %v", evidence.Quality, tt.wantQuality)
			}
			if evidence.OverallScore < 0 || evidence.OverallScore > 1 {
				t.Errorf("OverallScore %v is out of range [0,1]", evidence.OverallScore)
			}
		})
	}
}

func TestAggregateEvidence(t *testing.T) {
	ea := NewEvidenceAnalyzer()

	// Create test evidence
	evidence1, _ := ea.AssessEvidence("Strong research data", "Journal", "claim-1", true)
	evidence2, _ := ea.AssessEvidence("Supporting study", "University", "claim-1", true)
	evidence3, _ := ea.AssessEvidence("Contradicting evidence", "Source", "claim-1", false)

	tests := []struct {
		name             string
		evidences        []*types.Evidence
		wantSupportCount int
		wantRefuteCount  int
	}{
		{
			name:             "no evidence",
			evidences:        []*types.Evidence{},
			wantSupportCount: 0,
			wantRefuteCount:  0,
		},
		{
			name:             "all supporting",
			evidences:        []*types.Evidence{evidence1, evidence2},
			wantSupportCount: 2,
			wantRefuteCount:  0,
		},
		{
			name:             "mixed evidence",
			evidences:        []*types.Evidence{evidence1, evidence2, evidence3},
			wantSupportCount: 2,
			wantRefuteCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agg := ea.AggregateEvidence(tt.evidences)
			if agg == nil {
				t.Fatal("AggregateEvidence() returned nil")
			}
			if agg.SupportingCount != tt.wantSupportCount {
				t.Errorf("SupportingCount = %v, want %v", agg.SupportingCount, tt.wantSupportCount)
			}
			if agg.RefutingCount != tt.wantRefuteCount {
				t.Errorf("RefutingCount = %v, want %v", agg.RefutingCount, tt.wantRefuteCount)
			}
		})
	}
}
