package types

import "sync"

// StringInterner provides string interning for common values
// TIER 3 OPTIMIZATION: Reduces memory footprint by deduplicating strings
// Use for frequently occurring values like mode names, tool names, common metadata keys
type StringInterner struct {
	mu      sync.RWMutex
	strings map[string]string // canonical string -> itself
}

var (
	// Global interners for common string types
	modeInterner     = NewStringInterner()
	toolNameInterner = NewStringInterner()
	metadataInterner = NewStringInterner()
)

// NewStringInterner creates a new string interner
func NewStringInterner() *StringInterner {
	return &StringInterner{
		strings: make(map[string]string, 100),
	}
}

// Intern returns the canonical instance of the string
// If the string hasn't been seen before, it's added to the intern pool
func (si *StringInterner) Intern(s string) string {
	if s == "" {
		return ""
	}

	// Fast path: check if already interned (read lock)
	si.mu.RLock()
	if canonical, exists := si.strings[s]; exists {
		si.mu.RUnlock()
		return canonical
	}
	si.mu.RUnlock()

	// Slow path: intern the string (write lock)
	si.mu.Lock()
	defer si.mu.Unlock()

	// Double-check after acquiring write lock
	if canonical, exists := si.strings[s]; exists {
		return canonical
	}

	// Add to intern pool
	si.strings[s] = s
	return s
}

// InternMode interns a thinking mode string
func InternMode(mode ThinkingMode) ThinkingMode {
	return ThinkingMode(modeInterner.Intern(string(mode)))
}

// InternToolName interns a tool name string
func InternToolName(toolName string) string {
	return toolNameInterner.Intern(toolName)
}

// InternMetadataKey interns a metadata key string
func InternMetadataKey(key string) string {
	return metadataInterner.Intern(key)
}

// Size returns the number of interned strings
func (si *StringInterner) Size() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return len(si.strings)
}

// Clear removes all interned strings (useful for testing)
func (si *StringInterner) Clear() {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.strings = make(map[string]string, 100)
}
