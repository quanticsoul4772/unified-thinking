package contextbridge

import (
	"unified-thinking/internal/storage"
)

// StorageAdapter adapts storage.SQLiteStorage to SignatureStorage interface
type StorageAdapter struct {
	sqlite *storage.SQLiteStorage
}

// NewStorageAdapter creates a new storage adapter
func NewStorageAdapter(sqlite *storage.SQLiteStorage) *StorageAdapter {
	return &StorageAdapter{sqlite: sqlite}
}

// FindCandidatesWithSignatures implements SignatureStorage interface
func (a *StorageAdapter) FindCandidatesWithSignatures(domain string, fingerprintPrefix string, limit int) ([]*CandidateWithSignature, error) {
	storageCandidates, err := a.sqlite.FindCandidatesWithSignatures(domain, fingerprintPrefix, limit)
	if err != nil {
		return nil, err
	}

	// Convert storage types to contextbridge types
	candidates := make([]*CandidateWithSignature, len(storageCandidates))
	for i, sc := range storageCandidates {
		var sig *Signature
		if sc.Signature != nil {
			sig = &Signature{
				Fingerprint:  sc.Signature.Fingerprint,
				Domain:       sc.Signature.Domain,
				KeyConcepts:  sc.Signature.KeyConcepts,
				ToolSequence: sc.Signature.ToolSequence,
				Complexity:   sc.Signature.Complexity,
			}
		}

		candidates[i] = &CandidateWithSignature{
			TrajectoryID: sc.TrajectoryID,
			SessionID:    sc.SessionID,
			Description:  sc.Description,
			SuccessScore: sc.SuccessScore,
			QualityScore: sc.QualityScore,
			Signature:    sig,
		}
	}

	return candidates, nil
}

// StoreContextSignature stores a signature using the adapted storage
func (a *StorageAdapter) StoreContextSignature(trajectoryID string, sig *Signature) error {
	storageSig := &storage.ContextSignature{
		TrajectoryID: trajectoryID,
		Fingerprint:  sig.Fingerprint,
		Domain:       sig.Domain,
		KeyConcepts:  sig.KeyConcepts,
		ToolSequence: sig.ToolSequence,
		Complexity:   sig.Complexity,
	}

	return a.sqlite.StoreContextSignature(trajectoryID, storageSig)
}
