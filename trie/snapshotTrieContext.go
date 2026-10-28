package trie

import "github.com/multiversx/mx-chain-go/common"

// snapshotTrieContext carries the trie codec and the storage used for snapshots.
// Snapshot storage provides epoch-bounded reads and uncached writes independently
// of the storage manager used for ordinary trie operations.
type snapshotTrieContext struct {
	common.TrieContext
	storage snapshotDb
}

func (stc *snapshotTrieContext) GetSnapshotStorage() snapshotDb {
	return stc.storage
}
