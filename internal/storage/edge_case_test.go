package storage_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

// TestMemoryStorage_ExtremelyLargeContent tests handling of very large thought content
func TestMemoryStorage_ExtremelyLargeContent(t *testing.T) {
	store := storage.NewMemoryStorage()

	// Create thought with 1MB of content
	largeContent := strings.Repeat("a", 1024*1024)

	thought := &types.Thought{
		Content:    largeContent,
		Mode:       types.ModeLinear,
		Confidence: 0.8,
		Timestamp:  time.Now(),
	}

	err := store.StoreThought(thought)
	if err != nil {
		t.Fatalf("Failed to store large thought: %v", err)
	}

	// Retrieve and verify
	retrieved, err := store.GetThought(thought.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve large thought: %v", err)
	}

	if len(retrieved.Content) != len(largeContent) {
		t.Errorf("Content length mismatch: expected %d, got %d", len(largeContent), len(retrieved.Content))
	}
}

// TestMemoryStorage_UnicodeContent tests handling of Unicode edge cases
func TestMemoryStorage_UnicodeContent(t *testing.T) {
	store := storage.NewMemoryStorage()

	tests := []struct {
		name    string
		content string
	}{
		{"emoji", "🤔💭🧠 thinking with emojis"},
		{"japanese", "日本語のテスト内容"},
		{"mixed", "Mixed 日本語 and English 🎌"},
		{"rtl", "العربية نص"},               // Arabic RTL text
		{"zero width", "test\u200Bcontent"}, // Zero-width space
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thought := &types.Thought{
				Content:    tt.content,
				Mode:       types.ModeLinear,
				Confidence: 0.8,
				Timestamp:  time.Now(),
			}

			err := store.StoreThought(thought)
			if err != nil {
				t.Fatalf("Failed to store %s thought: %v", tt.name, err)
			}

			retrieved, err := store.GetThought(thought.ID)
			if err != nil {
				t.Fatalf("Failed to retrieve %s thought: %v", tt.name, err)
			}

			if retrieved.Content != tt.content {
				t.Errorf("Content mismatch for %s:\nExpected: %s\nGot: %s",
					tt.name, tt.content, retrieved.Content)
			}
		})
	}
}

// TestMemoryStorage_ConcurrentWritesSameThought tests concurrent updates to same thought
func TestMemoryStorage_ConcurrentWritesSameThought(t *testing.T) {
	store := storage.NewMemoryStorage()

	thought := &types.Thought{
		ID:         "concurrent-test",
		Content:    "Initial",
		Mode:       types.ModeLinear,
		Confidence: 0.5,
		Timestamp:  time.Now(),
	}

	// Store initial thought
	if err := store.StoreThought(thought); err != nil {
		t.Fatalf("Initial store failed: %v", err)
	}

	// Launch 50 concurrent updates
	const numUpdates = 50
	var wg sync.WaitGroup
	wg.Add(numUpdates)

	for i := 0; i < numUpdates; i++ {
		go func(idx int) {
			defer wg.Done()

			updated := &types.Thought{
				ID:         "concurrent-test",
				Content:    "Updated" + string(rune(idx)),
				Mode:       types.ModeLinear,
				Confidence: 0.5 + float64(idx)/100.0,
				Timestamp:  time.Now(),
			}

			_ = store.StoreThought(updated)
		}(i)
	}

	wg.Wait()

	// Final thought should exist (which update won is non-deterministic)
	final, err := store.GetThought("concurrent-test")
	if err != nil {
		t.Fatalf("Failed to retrieve after concurrent updates: %v", err)
	}

	if final.ID != "concurrent-test" {
		t.Error("Thought ID changed during concurrent updates")
	}
}

// TestMemoryStorage_ManyThoughts tests behavior with large number of thoughts
func TestMemoryStorage_ManyThoughts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large data test in short mode")
	}

	store := storage.NewMemoryStorage()

	// Store 10,000 thoughts
	const numThoughts = 10000

	for i := 0; i < numThoughts; i++ {
		thought := &types.Thought{
			Content:    "Thought " + string(rune(i)),
			Mode:       types.ModeLinear,
			Confidence: 0.5,
			Timestamp:  time.Now(),
		}

		if err := store.StoreThought(thought); err != nil {
			t.Fatalf("Failed to store thought %d: %v", i, err)
		}
	}

	// Retrieve thoughts via search (limited to MaxSearchResults = 1000)
	results := store.SearchThoughts("", "", 10000, 0)

	// Storage may have MaxSearchResults limit
	if len(results) < 1000 {
		t.Errorf("Expected at least 1000 thoughts (limit), got %d", len(results))
	}

	// Verify we can retrieve a specific thought
	thought, err := store.GetThought(results[0].ID)
	if err != nil {
		t.Errorf("Failed to retrieve specific thought: %v", err)
	}
	if thought == nil {
		t.Error("Retrieved thought is nil")
	}
}

// TestMemoryStorage_EmptyContent tests handling of empty thought content
func TestMemoryStorage_EmptyContent(t *testing.T) {
	store := storage.NewMemoryStorage()

	thought := &types.Thought{
		Content:    "", // Empty content
		Mode:       types.ModeLinear,
		Confidence: 0.5,
		Timestamp:  time.Now(),
	}

	err := store.StoreThought(thought)
	// Empty content should be allowed (validation happens at handler level)
	if err != nil {
		t.Errorf("Empty content should be allowed at storage level: %v", err)
	}
}

// TestMemoryStorage_NilMetadata tests handling of nil metadata
func TestMemoryStorage_NilMetadata(t *testing.T) {
	store := storage.NewMemoryStorage()

	thought := &types.Thought{
		Content:    "Test",
		Mode:       types.ModeLinear,
		Confidence: 0.8,
		Timestamp:  time.Now(),
		Metadata:   nil, // Nil metadata
	}

	err := store.StoreThought(thought)
	if err != nil {
		t.Fatalf("Failed to store thought with nil metadata: %v", err)
	}

	retrieved, err := store.GetThought(thought.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve thought: %v", err)
	}

	// Metadata should be initialized to empty map
	if retrieved.Metadata == nil {
		t.Error("Expected metadata to be initialized to empty map, got nil")
	}
}

// TestMemoryStorage_SearchWithOffsetAndLimit tests pagination edge cases
func TestMemoryStorage_SearchWithOffsetAndLimit(t *testing.T) {
	store := storage.NewMemoryStorage()

	// Create 20 thoughts
	for i := 0; i < 20; i++ {
		thought := &types.Thought{
			Content:    "Searchable content",
			Mode:       types.ModeLinear,
			Confidence: 0.5,
			Timestamp:  time.Now(),
		}
		store.StoreThought(thought)
	}

	tests := []struct {
		name          string
		query         string
		limit         int
		offset        int
		expectedCount int
	}{
		{"first page", "Searchable", 10, 0, 10},
		{"second page", "Searchable", 10, 10, 10},
		{"offset beyond results", "Searchable", 10, 100, 0},
		{"limit zero", "Searchable", 0, 0, 20},      // Should return all
		{"negative limit", "Searchable", -1, 0, 20}, // Should default to max
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := store.SearchThoughts(tt.query, "", tt.limit, tt.offset)

			if len(results) != tt.expectedCount {
				t.Errorf("Expected %d results, got %d", tt.expectedCount, len(results))
			}
		})
	}
}

// TestMemoryStorage_GetNonexistentThought tests error handling for missing thought
func TestMemoryStorage_GetNonexistentThought(t *testing.T) {
	store := storage.NewMemoryStorage()

	_, err := store.GetThought("nonexistent-id")
	if err == nil {
		t.Error("Expected error when retrieving nonexistent thought")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// TestMemoryStorage_GetNonexistentBranch tests error handling for missing branch
func TestMemoryStorage_GetNonexistentBranch(t *testing.T) {
	store := storage.NewMemoryStorage()

	_, err := store.GetBranch("nonexistent-branch")
	if err == nil {
		t.Error("Expected error when retrieving nonexistent branch")
	}
}

// TestMemoryStorage_UpdateBranchPriority_Nonexistent tests updating nonexistent branch
func TestMemoryStorage_UpdateBranchPriority_Nonexistent(t *testing.T) {
	store := storage.NewMemoryStorage()

	err := store.UpdateBranchPriority("nonexistent", 0.9)
	if err == nil {
		t.Error("Expected error when updating nonexistent branch priority")
	}
}

// TestMemoryStorage_AppendToNonexistentBranch tests appending to missing branch
func TestMemoryStorage_AppendToNonexistentBranch(t *testing.T) {
	store := storage.NewMemoryStorage()

	thought := &types.Thought{
		Content:    "Test",
		Mode:       types.ModeTree,
		Confidence: 0.8,
		Timestamp:  time.Now(),
	}

	err := store.AppendThoughtToBranch("nonexistent-branch", thought)
	if err == nil {
		t.Error("Expected error when appending to nonexistent branch")
	}
}
