package trie

import (
	"github.com/multiversx/mx-chain-go/common"
)

// StorageManagerOptions specify the options that a trie storage manager can have
type StorageManagerOptions struct {
	PruningEnabled   bool
	SnapshotsEnabled bool
}

// CreateTrieStorageManager creates a new trie storage manager based on the given type
func CreateTrieStorageManager(
	args NewTrieStorageManagerArgs,
	options StorageManagerOptions,
) (common.StorageManager, error) {
	log.Debug("trie storage manager options",
		"trie pruning status", options.PruningEnabled,
		"trie snapshot status", options.SnapshotsEnabled,
	)

	var tsm common.StorageManager
	tsm, err := NewTrieStorageManager(args)
	if err != nil {
		return nil, err
	}

	if !options.PruningEnabled {
		tsm, err = NewTrieStorageManagerWithoutPruning(tsm)
		if err != nil {
			return nil, err
		}
	}

	if !options.SnapshotsEnabled {
		tsm, err = NewTrieStorageManagerWithoutSnapshot(tsm)
		if err != nil {
			return nil, err
		}
	}

	return tsm, nil
}
