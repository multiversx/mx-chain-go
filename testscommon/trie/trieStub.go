package trie

import (
	"context"
	"errors"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

var errNotImplemented = errors.New("not implemented")

// TrieStub -
type TrieStub struct {
	GetCalled                       func(key []byte) ([]byte, uint32, error)
	UpdateCalled                    func(key, value []byte) error
	UpdateWithVersionCalled         func(key, value []byte, version core.TrieNodeVersion) error
	DeleteCalled                    func(key []byte)
	RootCalled                      func() ([]byte, error)
	CommitCalled                    func() error
	RecreateCalled                  func(root []byte) (common.Trie, error)
	RecreateFromEpochCalled         func(options common.RootHashHolder) (common.Trie, error)
	GetObsoleteHashesCalled         func() [][]byte
	AppendToOldHashesCalled         func([][]byte)
	GetSerializedNodesCalled        func([]byte, uint64) ([][]byte, uint64, error)
	GetAllHashesCalled              func() ([][]byte, error)
	GetAllLeavesOnChannelCalled     func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder common.KeyBuilder, trieLeafParser common.TrieLeafParser) error
	GetProofCalled                  func(key []byte) ([][]byte, []byte, error)
	VerifyProofCalled               func(rootHash []byte, key []byte, proof [][]byte) (bool, error)
	GetStorageManagerCalled         func() common.StorageManager
	GetSerializedNodeCalled         func(bytes []byte) ([]byte, error)
	GetOldRootCalled                func() []byte
	CloseCalled                     func() error
	CollectLeavesForMigrationCalled func(args vmcommon.ArgsMigrateDataTrieLeaves) error
	IsMigratedToLatestVersionCalled func() (bool, error)
}

// GetStorageManager -
func (ts *TrieStub) GetStorageManager() common.StorageManager {
	if ts.GetStorageManagerCalled != nil {
		return ts.GetStorageManagerCalled()
	}

	return nil
}

// GetProof -
func (ts *TrieStub) GetProof(key []byte) ([][]byte, []byte, error) {
	if ts.GetProofCalled != nil {
		return ts.GetProofCalled(key)
	}

	return nil, nil, nil
}

// VerifyProof -
func (ts *TrieStub) VerifyProof(rootHash []byte, key []byte, proof [][]byte) (bool, error) {
	if ts.VerifyProofCalled != nil {
		return ts.VerifyProofCalled(rootHash, key, proof)
	}

	return false, nil
}

// GetAllLeavesOnChannel -
func (ts *TrieStub) GetAllLeavesOnChannel(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder common.KeyBuilder, trieLeafParser common.TrieLeafParser) error {
	if ts.GetAllLeavesOnChannelCalled != nil {
		return ts.GetAllLeavesOnChannelCalled(leavesChannels, ctx, rootHash, keyBuilder, trieLeafParser)
	}

	return nil
}

// Get -
func (ts *TrieStub) Get(key []byte) ([]byte, uint32, error) {
	if ts.GetCalled != nil {
		return ts.GetCalled(key)
	}

	return nil, 0, errNotImplemented
}

// Update -
func (ts *TrieStub) Update(key, value []byte) error {
	if ts.UpdateCalled != nil {
		return ts.UpdateCalled(key, value)
	}

	return errNotImplemented
}

// UpdateWithVersion -
func (ts *TrieStub) UpdateWithVersion(key []byte, value []byte, version core.TrieNodeVersion) error {
	if ts.UpdateWithVersionCalled != nil {
		return ts.UpdateWithVersionCalled(key, value, version)
	}

	return errNotImplemented
}

// CollectLeavesForMigration -
func (ts *TrieStub) CollectLeavesForMigration(args vmcommon.ArgsMigrateDataTrieLeaves) error {
	if ts.CollectLeavesForMigrationCalled != nil {
		return ts.CollectLeavesForMigrationCalled(args)
	}

	return errNotImplemented
}

// Delete -
func (ts *TrieStub) Delete(key []byte) {
	if ts.DeleteCalled != nil {
		ts.DeleteCalled(key)
	}
}

// RootHash -
func (ts *TrieStub) RootHash() ([]byte, error) {
	if ts.RootCalled != nil {
		return ts.RootCalled()
	}

	return nil, errNotImplemented
}

// Commit -
func (ts *TrieStub) Commit() error {
	if ts.CommitCalled != nil {
		return ts.CommitCalled()
	}

	return errNotImplemented
}

// Recreate -
func (ts *TrieStub) Recreate(root []byte) (common.Trie, error) {
	if ts.RecreateCalled != nil {
		return ts.RecreateCalled(root)
	}

	return nil, errNotImplemented
}

// RecreateFromEpoch -
func (ts *TrieStub) RecreateFromEpoch(options common.RootHashHolder) (common.Trie, error) {
	if ts.RecreateFromEpochCalled != nil {
		return ts.RecreateFromEpochCalled(options)
	}

	return nil, errNotImplemented
}

// String -
func (ts *TrieStub) String() string {
	return "stub trie"
}

// IsInterfaceNil returns true if there is no value under the interface
func (ts *TrieStub) IsInterfaceNil() bool {
	return ts == nil
}

// GetObsoleteHashes resets the oldHashes and oldRoot variables and returns the old hashes
func (ts *TrieStub) GetObsoleteHashes() [][]byte {
	if ts.GetObsoleteHashesCalled != nil {
		return ts.GetObsoleteHashesCalled()
	}

	return nil
}

// GetSerializedNodes -
func (ts *TrieStub) GetSerializedNodes(hash []byte, maxBuffToSend uint64) ([][]byte, uint64, error) {
	if ts.GetSerializedNodesCalled != nil {
		return ts.GetSerializedNodesCalled(hash, maxBuffToSend)
	}
	return nil, 0, nil
}

// GetDirtyHashes -
func (ts *TrieStub) GetDirtyHashes() (common.ModifiedHashes, error) {
	return nil, nil
}

// SetNewHashes -
func (ts *TrieStub) SetNewHashes(_ common.ModifiedHashes) {
}

// GetAllHashes -
func (ts *TrieStub) GetAllHashes() ([][]byte, error) {
	if ts.GetAllHashesCalled != nil {
		return ts.GetAllHashesCalled()
	}

	return nil, nil
}

// GetSerializedNode -
func (ts *TrieStub) GetSerializedNode(bytes []byte) ([]byte, error) {
	if ts.GetSerializedNodeCalled != nil {
		return ts.GetSerializedNodeCalled(bytes)
	}

	return nil, nil
}

// GetOldRoot -
func (ts *TrieStub) GetOldRoot() []byte {
	if ts.GetOldRootCalled != nil {
		return ts.GetOldRootCalled()
	}

	return nil
}

// IsMigratedToLatestVersion -
func (ts *TrieStub) IsMigratedToLatestVersion() (bool, error) {
	if ts.IsMigratedToLatestVersionCalled != nil {
		return ts.IsMigratedToLatestVersionCalled()
	}

	return false, nil
}

// Close -
func (ts *TrieStub) Close() error {
	if ts.CloseCalled != nil {
		return ts.CloseCalled()
	}

	return nil
}
