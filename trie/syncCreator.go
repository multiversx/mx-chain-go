package trie

import (
	"context"
	"fmt"
	"time"
)

const (
	initialVersion = 1
	secondVersion  = 2
	thirdVersion   = 3
)

// TrieSyncer synchronizes the trie, asking on the network for the missing nodes
type TrieSyncer interface {
	StartSyncing(rootHash []byte, ctx context.Context) error
	NumLeaves() uint64
	NumBytes() uint64
	NumTrieNodes() uint64
	Duration() time.Duration
	IsInterfaceNil() bool
}

// CreateTrieSyncer is the method factory to create the correct trie syncer implementation
// TODO try to split this package (syncers should go in sync package, this file in the factory package)
func CreateTrieSyncer(arg ArgTrieSyncer, trieSyncerVersion int) (TrieSyncer, error) {
	switch trieSyncerVersion {
	case initialVersion:
		return NewTrieSyncer(arg)
	case secondVersion:
		return NewDoubleListTrieSyncer(arg)
	case thirdVersion:
		return NewDepthFirstTrieSyncer(arg)
	default:
		return nil, fmt.Errorf("%w, unknown value %d", ErrInvalidTrieSyncerVersion, trieSyncerVersion)
	}
}

// CheckTrieSyncerVersion can check if the syncer version has a correct value
func CheckTrieSyncerVersion(trieSyncerVersion int) error {
	isCorrectVersion := trieSyncerVersion >= initialVersion && trieSyncerVersion <= thirdVersion
	if isCorrectVersion {
		return nil
	}

	return fmt.Errorf("%w, unknown value %d", ErrInvalidTrieSyncerVersion, trieSyncerVersion)
}
