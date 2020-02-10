package sync

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type nilSyncTries struct {
	emptyTries map[string]data.Trie
}

// NewNilSyncTries returns a nilSyncTries structure which implements EpochStartTrieSyncer interface
func NewNilSyncTries() *nilSyncTries {
	return &nilSyncTries{
		emptyTries: make(map[string]data.Trie),
	}
}

// SyncTriesFrom returns nil as this is an empty implementation
func (n *nilSyncTries) SyncTriesFrom(_ *block.MetaBlock, _ time.Duration) error {
	return nil
}

// GetTries returns the empty tries
func (n *nilSyncTries) GetTries() (map[string]data.Trie, error) {
	return n.emptyTries, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (n *nilSyncTries) IsInterfaceNil() bool {
	return n == nil
}
