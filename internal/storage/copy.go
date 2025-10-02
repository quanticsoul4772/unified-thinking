package storage

import (
	"encoding/json"
	"log"
	"unified-thinking/internal/types"
)

// copyThought creates a deep copy of a thought to prevent external modification
func copyThought(t *types.Thought) *types.Thought {
	if t == nil {
		return nil
	}

	thoughtCopy := *t

	// Deep copy slices
	if len(t.KeyPoints) > 0 {
		thoughtCopy.KeyPoints = make([]string, len(t.KeyPoints))
		copy(thoughtCopy.KeyPoints, t.KeyPoints)
	}

	// Deep copy map - use JSON marshal/unmarshal for true deep copy
	thoughtCopy.Metadata = deepCopyMap(t.Metadata)

	return &thoughtCopy
}

// copyBranch creates a deep copy of a branch to prevent external modification
func copyBranch(b *types.Branch) *types.Branch {
	if b == nil {
		return nil
	}

	branchCopy := *b

	// Deep copy Thoughts slice (pointers to thoughts)
	if len(b.Thoughts) > 0 {
		branchCopy.Thoughts = make([]*types.Thought, len(b.Thoughts))
		for i, t := range b.Thoughts {
			branchCopy.Thoughts[i] = copyThought(t)
		}
	}

	// Deep copy Insights slice (pointers to insights)
	if len(b.Insights) > 0 {
		branchCopy.Insights = make([]*types.Insight, len(b.Insights))
		for i, ins := range b.Insights {
			branchCopy.Insights[i] = copyInsight(ins)
		}
	}

	// Deep copy CrossRefs slice (pointers to cross-refs)
	if len(b.CrossRefs) > 0 {
		branchCopy.CrossRefs = make([]*types.CrossRef, len(b.CrossRefs))
		for i, xr := range b.CrossRefs {
			branchCopy.CrossRefs[i] = copyCrossRef(xr)
		}
	}

	return &branchCopy
}

// copyInsight creates a deep copy of an insight
func copyInsight(ins *types.Insight) *types.Insight {
	if ins == nil {
		return nil
	}

	insightCopy := *ins

	// Deep copy slices
	if len(ins.Context) > 0 {
		insightCopy.Context = make([]string, len(ins.Context))
		copy(insightCopy.Context, ins.Context)
	}

	if len(ins.ParentInsights) > 0 {
		insightCopy.ParentInsights = make([]string, len(ins.ParentInsights))
		copy(insightCopy.ParentInsights, ins.ParentInsights)
	}

	// Deep copy map - use JSON marshal/unmarshal for true deep copy
	insightCopy.SupportingEvidence = deepCopyMap(ins.SupportingEvidence)

	// Deep copy Validations slice
	if len(ins.Validations) > 0 {
		insightCopy.Validations = make([]*types.Validation, len(ins.Validations))
		for i, v := range ins.Validations {
			insightCopy.Validations[i] = copyValidation(v)
		}
	}

	return &insightCopy
}

// copyCrossRef creates a deep copy of a cross-reference
func copyCrossRef(xr *types.CrossRef) *types.CrossRef {
	if xr == nil {
		return nil
	}

	xrefCopy := *xr

	// Deep copy TouchPoints slice
	if len(xr.TouchPoints) > 0 {
		xrefCopy.TouchPoints = make([]types.TouchPoint, len(xr.TouchPoints))
		copy(xrefCopy.TouchPoints, xr.TouchPoints)
	}

	return &xrefCopy
}

// copyValidation creates a deep copy of a validation
func copyValidation(v *types.Validation) *types.Validation {
	if v == nil {
		return nil
	}

	validationCopy := *v

	// Deep copy map - use JSON marshal/unmarshal for true deep copy
	validationCopy.ValidationData = deepCopyMap(v.ValidationData)

	return &validationCopy
}

// copyRelationship creates a deep copy of a relationship
func copyRelationship(r *types.Relationship) *types.Relationship {
	if r == nil {
		return nil
	}

	relationshipCopy := *r

	// Deep copy map - use JSON marshal/unmarshal for true deep copy
	relationshipCopy.Metadata = deepCopyMap(r.Metadata)

	return &relationshipCopy
}

// deepCopyMap creates a true deep copy of a map[string]interface{} using JSON marshaling
// This ensures no shared references between the original and copy
// Returns empty map (not nil) to ensure JSON serialization as {} instead of null
func deepCopyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return map[string]interface{}{}
	}

	// Use JSON marshal/unmarshal for true deep copy
	data, err := json.Marshal(m)
	if err != nil {
		log.Printf("Warning: Failed to deep copy map: %v", err)
		return make(map[string]interface{})
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		log.Printf("Warning: Failed to unmarshal map copy: %v", err)
		return make(map[string]interface{})
	}

	return result
}
