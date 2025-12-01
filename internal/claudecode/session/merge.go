package session

// MergeStrategy defines how to handle conflicts during import
type MergeStrategy string

const (
	// MergeReplace clears existing data and imports new
	MergeReplace MergeStrategy = "replace"
	// MergeMerge updates existing items, adds new ones (default)
	MergeMerge MergeStrategy = "merge"
	// MergeAppend keeps existing items, adds new with new IDs
	MergeAppend MergeStrategy = "append"
)

// IsValid checks if the merge strategy is valid
func (s MergeStrategy) IsValid() bool {
	switch s {
	case MergeReplace, MergeMerge, MergeAppend:
		return true
	default:
		return false
	}
}

// ParseMergeStrategy parses a string to MergeStrategy, defaulting to MergeMerge
func ParseMergeStrategy(s string) MergeStrategy {
	switch s {
	case "replace":
		return MergeReplace
	case "append":
		return MergeAppend
	case "merge", "":
		return MergeMerge
	default:
		return MergeMerge
	}
}

// ConflictResolution defines how to resolve specific conflicts
type ConflictResolution string

const (
	// ConflictKeepExisting keeps the existing item
	ConflictKeepExisting ConflictResolution = "keep_existing"
	// ConflictUseImported uses the imported item
	ConflictUseImported ConflictResolution = "use_imported"
	// ConflictMergeFields merges fields from both
	ConflictMergeFields ConflictResolution = "merge_fields"
)

// MergeOptions configures the merge behavior
type MergeOptions struct {
	Strategy           MergeStrategy      `json:"strategy"`
	ConflictResolution ConflictResolution `json:"conflict_resolution"`
	PreserveTimestamps bool               `json:"preserve_timestamps"`
	ValidateOnly       bool               `json:"validate_only"`
}

// DefaultMergeOptions returns the default merge options
func DefaultMergeOptions() MergeOptions {
	return MergeOptions{
		Strategy:           MergeMerge,
		ConflictResolution: ConflictUseImported,
		PreserveTimestamps: true,
		ValidateOnly:       false,
	}
}

// Conflict represents a detected conflict during merge
type Conflict struct {
	ItemType   string `json:"item_type"` // thought, branch, etc.
	ItemID     string `json:"item_id"`
	Field      string `json:"field,omitempty"`
	ExistingValue any  `json:"existing_value,omitempty"`
	ImportedValue any  `json:"imported_value,omitempty"`
	Resolution ConflictResolution `json:"resolution"`
}

// MergeReport contains details about the merge operation
type MergeReport struct {
	Strategy          MergeStrategy `json:"strategy"`
	ItemsProcessed    int           `json:"items_processed"`
	ItemsAdded        int           `json:"items_added"`
	ItemsUpdated      int           `json:"items_updated"`
	ItemsSkipped      int           `json:"items_skipped"`
	ConflictsDetected int           `json:"conflicts_detected"`
	ConflictsResolved int           `json:"conflicts_resolved"`
	Conflicts         []Conflict    `json:"conflicts,omitempty"`
	Errors            []string      `json:"errors,omitempty"`
}

// NewMergeReport creates a new merge report
func NewMergeReport(strategy MergeStrategy) *MergeReport {
	return &MergeReport{
		Strategy:  strategy,
		Conflicts: make([]Conflict, 0),
		Errors:    make([]string, 0),
	}
}

// AddConflict records a conflict in the report
func (r *MergeReport) AddConflict(conflict Conflict) {
	r.Conflicts = append(r.Conflicts, conflict)
	r.ConflictsDetected++
}

// AddError records an error in the report
func (r *MergeReport) AddError(err string) {
	r.Errors = append(r.Errors, err)
}

// Success returns true if the merge completed without errors
func (r *MergeReport) Success() bool {
	return len(r.Errors) == 0
}
