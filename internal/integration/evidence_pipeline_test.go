package integration

import (
	"math"
	"testing"

	"unified-thinking/internal/analysis"
	"unified-thinking/internal/reasoning"
)

func newTestPipeline() *EvidencePipeline {
	return NewEvidencePipeline(
		reasoning.NewProbabilisticReasoner(),
		reasoning.NewCausalReasoner(),
		reasoning.NewDecisionMaker(),
		analysis.NewEvidenceAnalyzer(),
	)
}

func TestProcessEvidenceCreatesBeliefImpact(t *testing.T) {
	pipeline := newTestPipeline()

	result, err := pipeline.ProcessEvidence(
		"Peer-reviewed study demonstrates statistically significant improvement with data 42.",
		"University Research Journal",
		"claim-success",
		true,
	)
	if err != nil {
		t.Fatalf("ProcessEvidence returned error: %v", err)
	}

	if result.Status != "success" {
		t.Fatalf("expected success status, got %q", result.Status)
	}

	if len(result.UpdatedBeliefs) == 0 {
		t.Fatal("expected at least one updated belief")
	}

	belief := result.UpdatedBeliefs[0]
	if len(belief.Evidence) == 0 {
		t.Fatal("expected belief to track supporting evidence")
	}

	if math.Abs(belief.Probability-belief.PriorProb) < 1e-6 {
		t.Fatalf("expected probability update, prior=%f posterior=%f", belief.PriorProb, belief.Probability)
	}

	impact := pipeline.GetEvidenceImpact(result.EvidenceID)
	beliefs, ok := impact["beliefs"].([]string)
	if !ok {
		t.Fatalf("expected beliefs impact slice, got %#v", impact["beliefs"])
	}
	if len(beliefs) != 1 {
		t.Fatalf("expected single linked belief, got %d", len(beliefs))
	}
}

func TestProcessEvidenceRefutingEvidenceLowersProbability(t *testing.T) {
	pipeline := newTestPipeline()

	result, err := pipeline.ProcessEvidence(
		"Anecdotal report suggests the claim might be incorrect.",
		"Opinion Blog",
		"claim-refute",
		false,
	)
	if err != nil {
		t.Fatalf("ProcessEvidence returned error: %v", err)
	}

	if len(result.UpdatedBeliefs) == 0 {
		t.Fatal("expected updated belief for refuting evidence")
	}

	belief := result.UpdatedBeliefs[0]
	if belief.Probability >= belief.PriorProb {
		t.Fatalf("expected posterior to decrease, prior=%f posterior=%f", belief.PriorProb, belief.Probability)
	}
}

func TestUpdateCausalGraphsAdjustsConfidence(t *testing.T) {
	prob := reasoning.NewProbabilisticReasoner()
	causal := reasoning.NewCausalReasoner()
	decision := reasoning.NewDecisionMaker()
	analyzer := analysis.NewEvidenceAnalyzer()
	pipeline := NewEvidencePipeline(prob, causal, decision, analyzer)

	graph, err := causal.BuildCausalGraph(
		"Marketing performance",
		[]string{"Increased marketing spend increases qualified leads"},
	)
	if err != nil {
		t.Fatalf("BuildCausalGraph error: %v", err)
	}

	if len(graph.Links) == 0 {
		t.Fatal("expected causal link to be generated")
	}

	evidence, err := analyzer.AssessEvidence(
		"Peer-reviewed study with detailed data confirms marketing spend increases leads.",
		"University Research", "claim-graph", true,
	)
	if err != nil {
		t.Fatalf("AssessEvidence error: %v", err)
	}

	pipeline.LinkEvidenceToCausalGraph(evidence.ID, graph.ID)

	originalConfidence := graph.Links[0].Confidence
	graphs, err := pipeline.updateCausalGraphs(evidence)
	if err != nil {
		t.Fatalf("updateCausalGraphs error: %v", err)
	}

	if len(graphs) != 1 {
		t.Fatalf("expected one updated graph, got %d", len(graphs))
	}

	updatedLink := graphs[0].Links[0]
	if updatedLink.Confidence <= originalConfidence {
		t.Fatalf("expected confidence increase, before=%f after=%f", originalConfidence, updatedLink.Confidence)
	}

	if len(updatedLink.Evidence) < 2 {
		t.Fatalf("expected evidence to be recorded on link, got %v", updatedLink.Evidence)
	}
}

func TestLinkEvidenceDeduplication(t *testing.T) {
	pipeline := newTestPipeline()

	pipeline.LinkEvidenceToBelief("evidence-dedup", "belief-1")
	pipeline.LinkEvidenceToBelief("evidence-dedup", "belief-1")

	pipeline.LinkEvidenceToCausalGraph("evidence-dedup", "graph-1")
	pipeline.LinkEvidenceToCausalGraph("evidence-dedup", "graph-1")

	pipeline.LinkEvidenceToDecision("evidence-dedup", "decision-1")
	pipeline.LinkEvidenceToDecision("evidence-dedup", "decision-1")

	impact := pipeline.GetEvidenceImpact("evidence-dedup")

	if beliefs := impact["beliefs"].([]string); len(beliefs) != 1 {
		t.Fatalf("expected one belief link, got %v", beliefs)
	}

	if graphs := impact["causal_graphs"].([]string); len(graphs) != 1 {
		t.Fatalf("expected one graph link, got %v", graphs)
	}

	if decisions := impact["decisions"].([]string); len(decisions) != 1 {
		t.Fatalf("expected one decision link, got %v", decisions)
	}
}
