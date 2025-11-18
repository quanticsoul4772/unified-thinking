package memory

import "unified-thinking/internal/storage"

// SQLiteSignatureAdapter adapts SQLiteStorage to the SignatureStorage interface
type SQLiteSignatureAdapter struct {
	store *storage.SQLiteStorage
}

// NewSQLiteSignatureAdapter creates a new adapter for SQLite storage
func NewSQLiteSignatureAdapter(store *storage.SQLiteStorage) *SQLiteSignatureAdapter {
	return &SQLiteSignatureAdapter{store: store}
}

// StoreContextSignature converts memory.ContextSignature to storage.ContextSignature and stores it
func (a *SQLiteSignatureAdapter) StoreContextSignature(trajectoryID string, sig *ContextSignature) error {
	if a.store == nil || sig == nil {
		return nil
	}

	// Convert memory.ContextSignature to storage.ContextSignature
	storageSig := &storage.ContextSignature{
		TrajectoryID: trajectoryID,
		Fingerprint:  sig.Fingerprint,
		Domain:       sig.Domain,
		KeyConcepts:  sig.KeyConcepts,
		ToolSequence: sig.ToolSequence,
		Complexity:   sig.Complexity,
	}

	return a.store.StoreContextSignature(trajectoryID, storageSig)
}
