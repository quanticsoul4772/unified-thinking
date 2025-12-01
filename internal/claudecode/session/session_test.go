package session

import (
	"encoding/json"
	"testing"
	"time"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// createTestStorage creates a memory storage for testing
func createTestStorage() storage.Storage {
	cfg := storage.Config{Type: storage.StorageTypeMemory}
	store, _ := storage.NewStorage(cfg)
	return store
}

func TestNewExporter(t *testing.T) {
	store := createTestStorage()
	exporter := NewExporter(store)
	if exporter == nil {
		t.Fatal("NewExporter should not return nil")
	}
}

func TestExporterExport(t *testing.T) {
	store := createTestStorage()
	exporter := NewExporter(store)

	// Add some test data
	store.StoreThought(&types.Thought{
		ID:         "t1",
		Content:    "First thought",
		Mode:       types.ModeLinear,
		Confidence: 0.8,
	})
	store.StoreThought(&types.Thought{
		ID:         "t2",
		Content:    "Second thought",
		Mode:       types.ModeTree,
		Confidence: 0.9,
	})
	store.StoreBranch(&types.Branch{
		ID: "b1",
	})

	export, err := exporter.Export("test-session", DefaultExportOptions())
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	if export.Version != SchemaVersion {
		t.Errorf("Expected version %s, got %s", SchemaVersion, export.Version)
	}
	if export.SessionID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got %s", export.SessionID)
	}
	if export.ExportedAt.IsZero() {
		t.Error("ExportedAt should be set")
	}
}

func TestExporterExportToJSON(t *testing.T) {
	store := createTestStorage()
	exporter := NewExporter(store)

	// Add test data
	store.StoreThought(&types.Thought{
		ID:      "t1",
		Content: "Test thought",
		Mode:    types.ModeLinear,
	})

	result, err := exporter.ExportToJSON("test-session", DefaultExportOptions())
	if err != nil {
		t.Fatalf("ExportToJSON failed: %v", err)
	}

	if result.ExportData == "" {
		t.Error("ExportData should not be empty")
	}
	if result.SizeBytes == 0 {
		t.Error("SizeBytes should be > 0")
	}
	if result.ExportVersion != SchemaVersion {
		t.Errorf("Expected version %s, got %s", SchemaVersion, result.ExportVersion)
	}

	// Verify it's valid JSON
	var decoded SessionExport
	if err := json.Unmarshal([]byte(result.ExportData), &decoded); err != nil {
		t.Fatalf("Failed to unmarshal exported JSON: %v", err)
	}
}

func TestExporterExportCompressed(t *testing.T) {
	store := createTestStorage()
	exporter := NewExporter(store)

	// Add test data
	store.StoreThought(&types.Thought{
		ID:      "t1",
		Content: "Test thought",
		Mode:    types.ModeLinear,
	})

	opts := DefaultExportOptions()
	opts.Compress = true

	result, err := exporter.ExportToJSON("test-session", opts)
	if err != nil {
		t.Fatalf("ExportToJSON compressed failed: %v", err)
	}

	if !result.Compressed {
		t.Error("Result should be marked as compressed")
	}
	if result.ExportData == "" {
		t.Error("Compressed output should not be empty")
	}
}

func TestNewImporter(t *testing.T) {
	store := createTestStorage()
	importer := NewImporter(store)
	if importer == nil {
		t.Fatal("NewImporter should not return nil")
	}
}

func TestImporterImportFromJSON(t *testing.T) {
	store := createTestStorage()
	importer := NewImporter(store)

	export := &SessionExport{
		Version:    SchemaVersion,
		ExportedAt: time.Now(),
		SessionID:  "test-session",
		Thoughts: []types.Thought{
			{ID: "t1", Content: "Imported thought", Mode: types.ModeLinear},
		},
	}

	data, _ := json.Marshal(export)

	result, err := importer.ImportFromJSON(string(data), DefaultMergeOptions())
	if err != nil {
		t.Fatalf("ImportFromJSON failed: %v", err)
	}

	if result.SessionID != "test-session" {
		t.Errorf("Expected session ID 'test-session', got '%s'", result.SessionID)
	}
	if result.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", result.Status)
	}
}

func TestImporterVersionValidation(t *testing.T) {
	store := createTestStorage()
	importer := NewImporter(store)

	export := &SessionExport{
		Version:    "99.0", // Future version (different major)
		ExportedAt: time.Now(),
		SessionID:  "test-session",
	}

	data, _ := json.Marshal(export)

	result, err := importer.ImportFromJSON(string(data), DefaultMergeOptions())
	if err != nil {
		t.Fatalf("ImportFromJSON failed: %v", err)
	}

	// Should have validation errors
	if len(result.ValidationErrors) == 0 {
		t.Error("Should have validation errors for incompatible version")
	}
}

func TestMergeStrategyValues(t *testing.T) {
	tests := []struct {
		strategy MergeStrategy
		expected string
	}{
		{MergeReplace, "replace"},
		{MergeMerge, "merge"},
		{MergeAppend, "append"},
	}

	for _, tt := range tests {
		if string(tt.strategy) != tt.expected {
			t.Errorf("MergeStrategy %v: got %s, want %s", tt.strategy, string(tt.strategy), tt.expected)
		}
	}
}

func TestMergeStrategyIsValid(t *testing.T) {
	validStrategies := []MergeStrategy{MergeReplace, MergeMerge, MergeAppend}
	for _, s := range validStrategies {
		if !s.IsValid() {
			t.Errorf("MergeStrategy %s should be valid", s)
		}
	}

	invalidStrategy := MergeStrategy("invalid")
	if invalidStrategy.IsValid() {
		t.Error("Invalid MergeStrategy should not be valid")
	}
}

func TestParseMergeStrategy(t *testing.T) {
	tests := []struct {
		input    string
		expected MergeStrategy
	}{
		{"replace", MergeReplace},
		{"merge", MergeMerge},
		{"append", MergeAppend},
		{"", MergeMerge},       // default
		{"invalid", MergeMerge}, // default
	}

	for _, tt := range tests {
		result := ParseMergeStrategy(tt.input)
		if result != tt.expected {
			t.Errorf("ParseMergeStrategy(%s): got %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestImporterMergeStrategies(t *testing.T) {
	tests := []struct {
		name     string
		strategy MergeStrategy
	}{
		{"replace", MergeReplace},
		{"merge", MergeMerge},
		{"append", MergeAppend},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := createTestStorage()
			importer := NewImporter(store)

			export := &SessionExport{
				Version:    SchemaVersion,
				ExportedAt: time.Now(),
				SessionID:  "test-session",
				Thoughts: []types.Thought{
					{ID: "t1", Content: "Test thought", Mode: types.ModeLinear},
				},
			}

			opts := DefaultMergeOptions()
			opts.Strategy = tt.strategy

			result, err := importer.Import(export, opts)
			if err != nil {
				t.Fatalf("Import failed: %v", err)
			}

			if result.Status != "success" && result.Status != "partial" {
				t.Errorf("Expected status success or partial, got %s", result.Status)
			}
		})
	}
}

func TestSessionExportJSONStructure(t *testing.T) {
	export := &SessionExport{
		Version:    SchemaVersion,
		ExportedAt: time.Now(),
		SessionID:  "test-123",
		Thoughts: []types.Thought{
			{ID: "t1", Content: "Test", Mode: types.ModeLinear},
		},
		Branches: []types.Branch{
			{ID: "b1"},
		},
		Decisions: []DecisionExport{
			{
				ID:       "d1",
				Question: "What to do?",
			},
		},
		CausalGraphs: []CausalGraphExport{
			{
				ID:          "cg1",
				Description: "Test graph",
			},
		},
	}

	data, err := json.Marshal(export)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify expected JSON keys
	var raw map[string]any
	json.Unmarshal(data, &raw)

	expectedKeys := []string{"version", "exported_at", "session_id", "thoughts", "branches"}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("Missing expected key: %s", key)
		}
	}
}

func TestExportWithEmptyStorage(t *testing.T) {
	store := createTestStorage()
	exporter := NewExporter(store)

	export, err := exporter.Export("empty-session", DefaultExportOptions())
	if err != nil {
		t.Fatalf("Export should succeed with empty storage: %v", err)
	}

	if export.Thoughts == nil {
		t.Error("Thoughts should be empty slice, not nil")
	}
	if export.Branches == nil {
		t.Error("Branches should be empty slice, not nil")
	}
}

func TestImportWithEmptyData(t *testing.T) {
	store := createTestStorage()
	importer := NewImporter(store)

	_, err := importer.ImportFromJSON("", DefaultMergeOptions())
	if err == nil {
		t.Error("Should error on empty data")
	}

	_, err = importer.ImportFromJSON("invalid json", DefaultMergeOptions())
	if err == nil {
		t.Error("Should error on invalid JSON")
	}
}

func TestDefaultExportOptions(t *testing.T) {
	opts := DefaultExportOptions()

	if !opts.IncludeDecisions {
		t.Error("DefaultExportOptions should include decisions")
	}
	if !opts.IncludeCausalGraphs {
		t.Error("DefaultExportOptions should include causal graphs")
	}
	if opts.Compress {
		t.Error("DefaultExportOptions should not compress by default")
	}
}

func TestDefaultMergeOptions(t *testing.T) {
	opts := DefaultMergeOptions()

	if opts.Strategy != MergeMerge {
		t.Errorf("DefaultMergeOptions strategy should be merge, got %s", opts.Strategy)
	}
	if !opts.PreserveTimestamps {
		t.Error("DefaultMergeOptions should preserve timestamps")
	}
	if opts.ValidateOnly {
		t.Error("DefaultMergeOptions should not be validate only")
	}
}

func TestMergeReport(t *testing.T) {
	report := NewMergeReport(MergeMerge)

	if report.Strategy != MergeMerge {
		t.Errorf("Expected strategy MergeMerge, got %s", report.Strategy)
	}

	report.AddConflict(Conflict{
		ItemType: "thought",
		ItemID:   "t1",
	})

	if report.ConflictsDetected != 1 {
		t.Errorf("Expected 1 conflict, got %d", report.ConflictsDetected)
	}

	// With conflicts but no errors, Success() should return true
	if !report.Success() {
		t.Error("Report should show success when there are no errors (conflicts don't affect success)")
	}

	report.AddError("test error")
	if report.Success() {
		t.Error("Report should not show success with errors")
	}
}

func TestValidateOnlyMode(t *testing.T) {
	store := createTestStorage()
	importer := NewImporter(store)

	export := &SessionExport{
		Version:    SchemaVersion,
		ExportedAt: time.Now(),
		SessionID:  "test-session",
		Thoughts: []types.Thought{
			{ID: "t1", Content: "Test", Mode: types.ModeLinear},
		},
	}

	opts := DefaultMergeOptions()
	opts.ValidateOnly = true

	result, err := importer.Import(export, opts)
	if err != nil {
		t.Fatalf("Validate only import failed: %v", err)
	}

	// Validate only should not import any data
	if result.ImportedThoughts != 0 {
		t.Error("Validate only should not import thoughts")
	}
}
