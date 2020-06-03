package disabled

import "github.com/ElrondNetwork/elrond-go/data"

type trie struct {
}

// NewTrie creates a disabled implementation of a trie
func NewTrie() *trie {
	return &trie{}
}

// Get -
func (t *trie) Get(_ []byte) ([]byte, error) {
	return nil, nil
}

// Update -
func (t *trie) Update(_, _ []byte) error {
	return nil
}

// Delete -
func (t *trie) Delete(_ []byte) error {
	return nil
}

// Root -
func (t *trie) Root() ([]byte, error) {
	return nil, nil
}

// Commit -
func (t *trie) Commit() error {
	return nil
}

// Recreate -
func (t *trie) Recreate(_ []byte) (data.Trie, error) {
	return nil, nil
}

// String -
func (t *trie) String() string {
	return ""
}

// CancelPrune -
func (t *trie) CancelPrune(_ []byte, _ data.TriePruningIdentifier) {
}

// Prune -
func (t *trie) Prune(_ []byte, _ data.TriePruningIdentifier) {
}

// TakeSnapshot -
func (t *trie) TakeSnapshot(_ []byte) {
}

// SetCheckpoint -
func (t *trie) SetCheckpoint(_ []byte) {
}

// ResetOldHashes -
func (t *trie) ResetOldHashes() [][]byte {
	return nil
}

// AppendToOldHashes -
func (t *trie) AppendToOldHashes(_ [][]byte) {
}

// GetDirtyHashes -
func (t *trie) GetDirtyHashes() (data.ModifiedHashes, error) {
	return nil, nil
}

// SetNewHashes -
func (t *trie) SetNewHashes(_ data.ModifiedHashes) {
}

// Database -
func (t *trie) Database() data.DBWriteCacher {
	return nil
}

// GetSerializedNodes -
func (t *trie) GetSerializedNodes(_ []byte, _ uint64) ([][]byte, uint64, error) {
	return nil, 0, nil
}

// GetAllLeaves -
func (t *trie) GetAllLeaves() (map[string][]byte, error) {
	return nil, nil
}

// IsPruningEnabled -
func (t *trie) IsPruningEnabled() bool {
	return false
}

// EnterSnapshotMode -
func (t *trie) EnterSnapshotMode() {
}

// ExitSnapshotMode -
func (t *trie) ExitSnapshotMode() {
}

// ClosePersister -
func (t *trie) ClosePersister() error {
	return nil
}

// IsInterfaceNil -
func (t *trie) IsInterfaceNil() bool {
	return t == nil
}
