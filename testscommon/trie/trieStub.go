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
	UpdateCalled                    func(key, value []byte)
	UpdateWithVersionCalled         func(key, value []byte, version core.TrieNodeVersion)
	DeleteCalled                    func(key []byte)
	RootCalled                      func() ([]byte, error)
	CommitCalled                    func(collector common.TrieHashesCollector) error
	RecreateCalled                  func(options common.RootHashHolder, identifier string) (common.Trie, error)
	AppendToOldHashesCalled         func([][]byte)
	GetSerializedNodesCalled        func([]byte, uint64) ([][]byte, uint64, error)
	GetAllLeavesOnChannelCalled     func(leavesChannels *common.TrieIteratorChannels, ctx context.Context, rootHash []byte, keyBuilder common.KeyBuilder, trieLeafParser common.TrieLeafParser) error
	GetProofCalled                  func(key []byte, rootHash []byte) ([][]byte, []byte, error)
	VerifyProofCalled               func(rootHash []byte, key []byte, proof [][]byte) (bool, error)
	GetStorageManagerCalled         func() common.StorageManager
	GetSerializedNodeCalled         func(bytes []byte) ([]byte, error)
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
func (ts *TrieStub) GetProof(key []byte, rootHash []byte) ([][]byte, []byte, error) {
	if ts.GetProofCalled != nil {
		return ts.GetProofCalled(key, rootHash)
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
func (ts *TrieStub) Update(key, value []byte) {
	if ts.UpdateCalled != nil {
		ts.UpdateCalled(key, value)
	}
}

// UpdateWithVersion -
func (ts *TrieStub) UpdateWithVersion(key []byte, value []byte, version core.TrieNodeVersion) {
	if ts.UpdateWithVersionCalled != nil {
		ts.UpdateWithVersionCalled(key, value, version)
	}
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
func (ts *TrieStub) Commit(hc common.TrieHashesCollector) error {
	if ts.CommitCalled != nil {
		return ts.CommitCalled(hc)
	}

	return errNotImplemented
}

// Recreate -
func (ts *TrieStub) Recreate(options common.RootHashHolder, identifier string) (common.Trie, error) {
	if ts.RecreateCalled != nil {
		return ts.RecreateCalled(options, identifier)
	}

	return nil, errNotImplemented
}

// IsInterfaceNil returns true if there is no value under the interface
func (ts *TrieStub) IsInterfaceNil() bool {
	return ts == nil
}

// GetSerializedNodes -
func (ts *TrieStub) GetSerializedNodes(hash []byte, maxBuffToSend uint64) ([][]byte, uint64, error) {
	if ts.GetSerializedNodesCalled != nil {
		return ts.GetSerializedNodesCalled(hash, maxBuffToSend)
	}
	return nil, 0, nil
}

// SetNewHashes -
func (ts *TrieStub) SetNewHashes(_ common.ModifiedHashes) {
}

// GetSerializedNode -
func (ts *TrieStub) GetSerializedNode(bytes []byte) ([]byte, error) {
	if ts.GetSerializedNodeCalled != nil {
		return ts.GetSerializedNodeCalled(bytes)
	}

	return nil, nil
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
