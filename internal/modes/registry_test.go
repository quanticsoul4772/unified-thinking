package modes

import (
	"context"
	"testing"

	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("NewRegistry returned nil")
	}

	if registry.modes == nil {
		t.Error("modes map not initialized")
	}

	if len(registry.modes) != 0 {
		t.Errorf("Expected empty registry, got %d modes", len(registry.modes))
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)

	err := registry.Register(linear)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Verify mode was registered
	if len(registry.modes) != 1 {
		t.Errorf("Expected 1 registered mode, got %d", len(registry.modes))
	}

	// Try to register same mode again
	err = registry.Register(linear)
	if err == nil {
		t.Error("Expected error when registering duplicate mode, got nil")
	}
}

func TestRegistry_Register_MultipleModes(t *testing.T) {
	registry := NewRegistry()
	store := storage.NewMemoryStorage()

	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)

	// Register all modes
	if err := registry.Register(linear); err != nil {
		t.Fatalf("Register(linear) error = %v", err)
	}
	if err := registry.Register(tree); err != nil {
		t.Fatalf("Register(tree) error = %v", err)
	}
	if err := registry.Register(divergent); err != nil {
		t.Fatalf("Register(divergent) error = %v", err)
	}

	if len(registry.modes) != 3 {
		t.Errorf("Expected 3 registered modes, got %d", len(registry.modes))
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	store := storage.NewMemoryStorage()
	linear := NewLinearMode(store)

	_ = registry.Register(linear)

	tests := []struct {
		name     string
		modeName types.ThinkingMode
		wantErr  bool
		wantName types.ThinkingMode
	}{
		{
			name:     "get existing mode",
			modeName: types.ModeLinear,
			wantErr:  false,
			wantName: types.ModeLinear,
		},
		{
			name:     "get non-existent mode",
			modeName: types.ModeTree,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, err := registry.Get(tt.modeName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if mode == nil {
					t.Error("Get() returned nil mode")
				}
				if mode.Name() != tt.wantName {
					t.Errorf("Mode name = %v, want %v", mode.Name(), tt.wantName)
				}
			}
		})
	}
}

func TestRegistry_SelectBest(t *testing.T) {
	registry := NewRegistry()
	store := storage.NewMemoryStorage()

	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)

	_ = registry.Register(linear)
	_ = registry.Register(tree)
	_ = registry.Register(divergent)

	tests := []struct {
		name     string
		input    ThoughtInput
		wantMode types.ThinkingMode
		wantErr  bool
	}{
		{
			name: "select divergent for creative content",
			input: ThoughtInput{
				Content:        "Let's think creatively about this",
				ForceRebellion: true,
			},
			wantMode: types.ModeDivergent,
			wantErr:  false,
		},
		{
			name: "select tree for branching",
			input: ThoughtInput{
				Content:  "Explore multiple branches",
				BranchID: "branch-1",
			},
			wantMode: types.ModeTree,
			wantErr:  false,
		},
		{
			name: "select linear for simple thought",
			input: ThoughtInput{
				Content: "Simple linear thought",
			},
			wantMode: types.ModeLinear,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, err := registry.SelectBest(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("SelectBest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if mode == nil {
					t.Fatal("SelectBest() returned nil mode")
				}
				if mode.Name() != tt.wantMode {
					t.Errorf("Selected mode = %v, want %v", mode.Name(), tt.wantMode)
				}
			}
		})
	}
}

func TestRegistry_SelectBest_EmptyRegistry(t *testing.T) {
	registry := NewRegistry()

	input := ThoughtInput{
		Content: "Test content",
	}

	_, err := registry.SelectBest(input)
	if err == nil {
		t.Error("Expected error for empty registry, got nil")
	}
}

func TestRegistry_Available(t *testing.T) {
	registry := NewRegistry()
	store := storage.NewMemoryStorage()

	linear := NewLinearMode(store)
	tree := NewTreeMode(store)

	_ = registry.Register(linear)
	_ = registry.Register(tree)

	modeNames := registry.Available()

	if len(modeNames) != 2 {
		t.Errorf("Available() returned %d modes, want 2", len(modeNames))
	}

	// Check that both modes are in the list
	found := make(map[types.ThinkingMode]bool)
	for _, name := range modeNames {
		found[name] = true
	}

	if !found[types.ModeLinear] {
		t.Error("Linear mode not in available list")
	}
	if !found[types.ModeTree] {
		t.Error("Tree mode not in available list")
	}
}

func TestRegistry_Count(t *testing.T) {
	registry := NewRegistry()
	store := storage.NewMemoryStorage()

	if registry.Count() != 0 {
		t.Errorf("Count() = %d, want 0", registry.Count())
	}

	_ = registry.Register(NewLinearMode(store))
	if registry.Count() != 1 {
		t.Errorf("Count() = %d, want 1", registry.Count())
	}

	_ = registry.Register(NewTreeMode(store))
	if registry.Count() != 2 {
		t.Errorf("Count() = %d, want 2", registry.Count())
	}

	_ = registry.Register(NewDivergentMode(store))
	if registry.Count() != 3 {
		t.Errorf("Count() = %d, want 3", registry.Count())
	}
}

func TestRegistry_ModeMethods(t *testing.T) {
	store := storage.NewMemoryStorage()

	// Create modes
	linear := NewLinearMode(store)
	tree := NewTreeMode(store)
	divergent := NewDivergentMode(store)
	auto := NewAutoMode(linear, tree, divergent)

	// Test Name() method
	if linear.Name() != types.ModeLinear {
		t.Errorf("linear.Name() = %v, want %v", linear.Name(), types.ModeLinear)
	}
	if tree.Name() != types.ModeTree {
		t.Errorf("tree.Name() = %v, want %v", tree.Name(), types.ModeTree)
	}
	if divergent.Name() != types.ModeDivergent {
		t.Errorf("divergent.Name() = %v, want %v", divergent.Name(), types.ModeDivergent)
	}
	if auto.Name() != types.ModeAuto {
		t.Errorf("auto.Name() = %v, want %v", auto.Name(), types.ModeAuto)
	}

	// Test CanHandle() method
	ctx := context.Background()

	// Linear should handle basic content
	if !linear.CanHandle(ThoughtInput{Content: "test"}) {
		t.Error("linear should handle basic content")
	}

	// Tree should handle branch inputs
	if !tree.CanHandle(ThoughtInput{Content: "test", BranchID: "b1"}) {
		t.Error("tree should handle branch inputs")
	}

	// Divergent should handle creative/rebellion inputs
	if !divergent.CanHandle(ThoughtInput{Content: "test", ForceRebellion: true}) {
		t.Error("divergent should handle rebellion inputs")
	}

	// Auto should handle anything
	if !auto.CanHandle(ThoughtInput{Content: "test"}) {
		t.Error("auto should handle any input")
	}

	// Test ProcessThought() method
	_, err := linear.ProcessThought(ctx, ThoughtInput{Content: "test"})
	if err != nil {
		t.Errorf("linear.ProcessThought() error = %v", err)
	}

	_, err = tree.ProcessThought(ctx, ThoughtInput{Content: "test"})
	if err != nil {
		t.Errorf("tree.ProcessThought() error = %v", err)
	}

	_, err = divergent.ProcessThought(ctx, ThoughtInput{Content: "test"})
	if err != nil {
		t.Errorf("divergent.ProcessThought() error = %v", err)
	}

	_, err = auto.ProcessThought(ctx, ThoughtInput{Content: "test"})
	if err != nil {
		t.Errorf("auto.ProcessThought() error = %v", err)
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		name string
		s    string
		sub  string
		want bool
	}{
		{"exact match", "hello", "hello", true},
		{"case insensitive match", "Hello", "hello", true},
		{"substring match", "Hello World", "world", true},
		{"no match", "Hello", "goodbye", false},
		{"empty substring", "Hello", "", true},
		{"empty string", "", "hello", false},
		{"both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsIgnoreCase(tt.s, tt.sub); got != tt.want {
				t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tt.s, tt.sub, got, tt.want)
			}
		})
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"all lowercase", "hello", "hello"},
		{"all uppercase", "HELLO", "hello"},
		{"mixed case", "HeLLo", "hello"},
		{"with numbers", "Hello123", "hello123"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toLower(tt.input); got != tt.want {
				t.Errorf("toLower(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		str    string
		substr string
		want   bool
	}{
		{"substring present", "abcdef", "cd", true},
		{"substring not present", "abcdef", "gh", false},
		{"empty substring", "hello", "", true},
		{"empty string", "", "a", false},
		{"both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.str, tt.substr); got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.str, tt.substr, got, tt.want)
			}
		})
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()
	store := storage.NewMemoryStorage()

	linear := NewLinearMode(store)
	tree := NewTreeMode(store)

	// Register modes concurrently
	done := make(chan bool, 2)

	go func() {
		_ = registry.Register(linear)
		done <- true
	}()

	go func() {
		_ = registry.Register(tree)
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify both modes registered
	if registry.Count() != 2 {
		t.Errorf("Expected 2 modes after concurrent registration, got %d", registry.Count())
	}

	// Get modes concurrently
	done2 := make(chan bool, 2)

	go func() {
		_, err := registry.Get(types.ModeLinear)
		if err != nil {
			t.Errorf("Get(linear) error = %v", err)
		}
		done2 <- true
	}()

	go func() {
		_, err := registry.Get(types.ModeTree)
		if err != nil {
			t.Errorf("Get(tree) error = %v", err)
		}
		done2 <- true
	}()

	<-done2
	<-done2
}
