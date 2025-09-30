package storage

import "unified-thinking/internal/types"

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

	// Deep copy map
	if len(t.Metadata) > 0 {
		thoughtCopy.Metadata = make(map[string]interface{}, len(t.Metadata))
		for k, v := range t.Metadata {
			thoughtCopy.Metadata[k] = v
		}
	}

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

	// Deep copy map
	if len(ins.SupportingEvidence) > 0 {
		insightCopy.SupportingEvidence = make(map[string]interface{}, len(ins.SupportingEvidence))
		for k, v := range ins.SupportingEvidence {
			insightCopy.SupportingEvidence[k] = v
		}
	}

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

	// Deep copy map
	if len(v.ValidationData) > 0 {
		validationCopy.ValidationData = make(map[string]interface{}, len(v.ValidationData))
		for k, val := range v.ValidationData {
			validationCopy.ValidationData[k] = val
		}
	}

	return &validationCopy
}
