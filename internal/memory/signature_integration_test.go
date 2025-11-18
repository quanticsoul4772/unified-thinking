package memory

import (
	"context"
	"testing"
	"time"
)

// MockSignatureStorage implements SignatureStorage for testing
type MockSignatureStorage struct {
	stored    map[string]*ContextSignature
	storeErr  error
	callCount int
}

func NewMockSignatureStorage() *MockSignatureStorage {
	return &MockSignatureStorage{
		stored: make(map[string]*ContextSignature),
	}
}

func (m *MockSignatureStorage) StoreContextSignature(trajectoryID string, sig *ContextSignature) error {
	m.callCount++
	if m.storeErr != nil {
		return m.storeErr
	}
	m.stored[trajectoryID] = sig
	return nil
}

func TestSignatureIntegration_GenerateAndStoreSignature(t *testing.T) {
	storage := NewMockSignatureStorage()
	integration := NewSignatureIntegration(storage, nil)

	trajectory := &ReasoningTrajectory{
		ID:         "traj-123",
		SessionID:  "sess-456",
		Domain:     "engineering",
		Complexity: 0.7,
		Problem: &ProblemDescription{
			Description: "How to optimize database queries for better performance",
			Context:     "Large scale application with millions of users",
			Domain:      "engineering",
			ProblemType: "optimization",
		},
		Approach: &ApproachDescription{
			ToolSequence: []string{"think", "make-decision", "validate"},
		},
		StartTime: time.Now(),
		EndTime:   time.Now().Add(5 * time.Minute),
	}

	err := integration.GenerateAndStoreSignature(trajectory)
	if err != nil {
		t.Fatalf("GenerateAndStoreSignature failed: %v", err)
	}

	if storage.callCount != 1 {
		t.Errorf("Expected 1 storage call, got %d", storage.callCount)
	}

	sig, ok := storage.stored["traj-123"]
	if !ok {
		t.Fatal("Signature not stored")
	}

	// Check fingerprint is generated
	if sig.Fingerprint == "" {
		t.Error("Expected fingerprint to be generated")
	}
	if len(sig.Fingerprint) != 64 { // SHA256 hex
		t.Errorf("Expected fingerprint length 64, got %d", len(sig.Fingerprint))
	}

	// Check key concepts extracted
	if len(sig.KeyConcepts) == 0 {
		t.Error("Expected key concepts to be extracted")
	}
	// Should contain concepts like "optimize", "database", "queries", etc.
	found := false
	for _, concept := range sig.KeyConcepts {
		if concept == "optimize" || concept == "database" || concept == "queries" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected to find relevant concepts, got: %v", sig.KeyConcepts)
	}

	// Check domain
	if sig.Domain != "engineering" {
		t.Errorf("Expected domain 'engineering', got '%s'", sig.Domain)
	}

	// Check tool sequence
	if len(sig.ToolSequence) != 3 {
		t.Errorf("Expected 3 tools, got %d", len(sig.ToolSequence))
	}

	// Check complexity
	if sig.Complexity != 0.7 {
		t.Errorf("Expected complexity 0.7, got %f", sig.Complexity)
	}
}

func TestSignatureIntegration_NilTrajectory(t *testing.T) {
	storage := NewMockSignatureStorage()
	integration := NewSignatureIntegration(storage, nil)

	err := integration.GenerateAndStoreSignature(nil)
	if err != nil {
		t.Fatalf("Expected no error for nil trajectory, got: %v", err)
	}

	if storage.callCount != 0 {
		t.Error("Expected no storage calls for nil trajectory")
	}
}

func TestSignatureIntegration_NilProblem(t *testing.T) {
	storage := NewMockSignatureStorage()
	integration := NewSignatureIntegration(storage, nil)

	trajectory := &ReasoningTrajectory{
		ID:      "traj-123",
		Problem: nil,
	}

	err := integration.GenerateAndStoreSignature(trajectory)
	if err != nil {
		t.Fatalf("Expected no error for nil problem, got: %v", err)
	}

	if storage.callCount != 0 {
		t.Error("Expected no storage calls for nil problem")
	}
}

func TestSignatureIntegration_EmptyDescription(t *testing.T) {
	storage := NewMockSignatureStorage()
	integration := NewSignatureIntegration(storage, nil)

	trajectory := &ReasoningTrajectory{
		ID: "traj-123",
		Problem: &ProblemDescription{
			Description: "",
		},
	}

	err := integration.GenerateAndStoreSignature(trajectory)
	if err != nil {
		t.Fatalf("Expected no error for empty description, got: %v", err)
	}

	if storage.callCount != 0 {
		t.Error("Expected no storage calls for empty description")
	}
}

func TestSignatureIntegration_NilStorage(t *testing.T) {
	integration := NewSignatureIntegration(nil, nil)

	trajectory := &ReasoningTrajectory{
		ID: "traj-123",
		Problem: &ProblemDescription{
			Description: "Test problem",
		},
	}

	err := integration.GenerateAndStoreSignature(trajectory)
	if err != nil {
		t.Fatalf("Expected no error for nil storage, got: %v", err)
	}
}

func TestSignatureIntegration_ExtractToolSequenceFromSteps(t *testing.T) {
	storage := NewMockSignatureStorage()
	integration := NewSignatureIntegration(storage, nil)

	trajectory := &ReasoningTrajectory{
		ID:     "traj-123",
		Domain: "engineering",
		Problem: &ProblemDescription{
			Description: "Test problem for tool extraction",
		},
		Approach: nil, // No approach, will extract from steps
		Steps: []*ReasoningStep{
			{Tool: "think", StepNumber: 1},
			{Tool: "validate", StepNumber: 2},
			{Tool: "think", StepNumber: 3}, // Duplicate
			{Tool: "make-decision", StepNumber: 4},
		},
	}

	err := integration.GenerateAndStoreSignature(trajectory)
	if err != nil {
		t.Fatalf("GenerateAndStoreSignature failed: %v", err)
	}

	sig := storage.stored["traj-123"]
	if sig == nil {
		t.Fatal("Signature not stored")
	}

	// Should have unique tools only
	if len(sig.ToolSequence) != 3 {
		t.Errorf("Expected 3 unique tools, got %d: %v", len(sig.ToolSequence), sig.ToolSequence)
	}
}

func TestSignatureIntegration_ComplexityEstimation(t *testing.T) {
	storage := NewMockSignatureStorage()
	integration := NewSignatureIntegration(storage, nil)

	trajectory := &ReasoningTrajectory{
		ID:         "traj-123",
		Complexity: 0, // Not set, should be estimated
		Problem: &ProblemDescription{
			Description: "Short problem",
			Complexity:  0, // Not set
		},
	}

	err := integration.GenerateAndStoreSignature(trajectory)
	if err != nil {
		t.Fatalf("GenerateAndStoreSignature failed: %v", err)
	}

	sig := storage.stored["traj-123"]
	if sig.Complexity <= 0 {
		t.Error("Expected complexity to be estimated")
	}
	if sig.Complexity > 1.0 {
		t.Error("Complexity should not exceed 1.0")
	}
}

func TestSimpleConceptExtractor(t *testing.T) {
	extractor := NewSimpleConceptExtractor()

	tests := []struct {
		name     string
		text     string
		expected []string
		minCount int
	}{
		{
			name:     "basic extraction",
			text:     "optimize database queries for performance",
			expected: []string{"optimize", "database", "queries", "performance"},
			minCount: 4,
		},
		{
			name:     "filters stop words",
			text:     "the quick brown fox jumps over the lazy dog",
			expected: []string{"quick", "brown", "jumps", "over", "lazy"},
			minCount: 4,
		},
		{
			name:     "removes short words",
			text:     "a an is to do of the API SQL",
			expected: []string{},
			minCount: 0,
		},
		{
			name:     "handles punctuation",
			text:     "Hello, world! How are you?",
			expected: []string{"hello", "world"},
			minCount: 2,
		},
		{
			name:     "unique concepts only",
			text:     "database database optimization database",
			expected: []string{"database", "optimization"},
			minCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			concepts := extractor.Extract(tt.text)
			if len(concepts) < tt.minCount {
				t.Errorf("Expected at least %d concepts, got %d: %v", tt.minCount, len(concepts), concepts)
			}
		})
	}
}

func TestEpisodicMemoryStore_WithSignatureIntegration(t *testing.T) {
	store := NewEpisodicMemoryStore()
	mockStorage := NewMockSignatureStorage()
	integration := NewSignatureIntegration(mockStorage, nil)
	store.SetSignatureIntegration(integration)

	trajectory := &ReasoningTrajectory{
		ID:        "traj-123",
		SessionID: "sess-456",
		Domain:    "testing",
		Problem: &ProblemDescription{
			Description: "Test storing trajectory with signature",
		},
		StartTime: time.Now(),
	}

	ctx := context.Background()
	err := store.StoreTrajectory(ctx, trajectory)
	if err != nil {
		t.Fatalf("StoreTrajectory failed: %v", err)
	}

	// Verify signature was stored
	if mockStorage.callCount != 1 {
		t.Errorf("Expected 1 signature storage call, got %d", mockStorage.callCount)
	}

	sig := mockStorage.stored["traj-123"]
	if sig == nil {
		t.Fatal("Signature not stored")
	}
	if sig.Fingerprint == "" {
		t.Error("Expected fingerprint to be generated")
	}
}
