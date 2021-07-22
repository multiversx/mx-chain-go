package testscommon

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

var errNotImplemented = errors.New("not implemented")

// TrieStub -
type TrieStub struct {
	GetCalled                   func(key []byte) ([]byte, error)
	UpdateCalled                func(key, value []byte) error
	DeleteCalled                func(key []byte) error
	RootCalled                  func() ([]byte, error)
	CommitCalled                func() error
	RecreateCalled              func(root []byte) (temporary.Trie, error)
	GetObsoleteHashesCalled     func() [][]byte
	AppendToOldHashesCalled     func([][]byte)
	GetSerializedNodesCalled    func([]byte, uint64) ([][]byte, uint64, error)
	GetAllHashesCalled          func() ([][]byte, error)
	GetAllLeavesOnChannelCalled func(rootHash []byte) (chan core.KeyValueHolder, error)
	GetProofCalled              func(key []byte) ([][]byte, error)
	VerifyProofCalled           func(key []byte, proof [][]byte) (bool, error)
	GetStorageManagerCalled     func() temporary.StorageManager
	GetSerializedNodeCalled     func(bytes []byte) ([]byte, error)
	GetNumNodesCalled           func() temporary.NumNodesDTO
	GetOldRootCalled            func() []byte
	CloseCalled                 func() error
}

// GetStorageManager -
func (ts *TrieStub) GetStorageManager() temporary.StorageManager {
	if ts.GetStorageManagerCalled != nil {
		return ts.GetStorageManagerCalled()
	}

	return nil
}

// GetProof -
func (ts *TrieStub) GetProof(key []byte) ([][]byte, error) {
	if ts.GetProofCalled != nil {
		return ts.GetProofCalled(key)
	}

	return nil, nil
}

// VerifyProof -
func (ts *TrieStub) VerifyProof(key []byte, proof [][]byte) (bool, error) {
	if ts.VerifyProofCalled != nil {
		return ts.VerifyProofCalled(key, proof)
	}

	return false, nil
}

// GetAllLeavesOnChannel -
func (ts *TrieStub) GetAllLeavesOnChannel(rootHash []byte) (chan core.KeyValueHolder, error) {
	if ts.GetAllLeavesOnChannelCalled != nil {
		return ts.GetAllLeavesOnChannelCalled(rootHash)
	}

	ch := make(chan core.KeyValueHolder)
	close(ch)

	return ch, nil
}

// Get -
func (ts *TrieStub) Get(key []byte) ([]byte, error) {
	if ts.GetCalled != nil {
		return ts.GetCalled(key)
	}

	return nil, errNotImplemented
}

// Update -
func (ts *TrieStub) Update(key, value []byte) error {
	if ts.UpdateCalled != nil {
		return ts.UpdateCalled(key, value)
	}

	return errNotImplemented
}

// Delete -
func (ts *TrieStub) Delete(key []byte) error {
	if ts.DeleteCalled != nil {
		return ts.DeleteCalled(key)
	}

	return errNotImplemented
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
func (ts *TrieStub) Recreate(root []byte) (temporary.Trie, error) {
	if ts.RecreateCalled != nil {
		return ts.RecreateCalled(root)
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
func (ts *TrieStub) GetDirtyHashes() (temporary.ModifiedHashes, error) {
	return nil, nil
}

// SetNewHashes -
func (ts *TrieStub) SetNewHashes(_ temporary.ModifiedHashes) {
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

// GetNumNodes -
func (ts *TrieStub) GetNumNodes() temporary.NumNodesDTO {
	if ts.GetNumNodesCalled != nil {
		return ts.GetNumNodesCalled()
	}

	return temporary.NumNodesDTO{}
}

// GetOldRoot -
func (ts *TrieStub) GetOldRoot() []byte {
	if ts.GetOldRootCalled != nil {
		return ts.GetOldRootCalled()
	}

	return nil
}

// Close -
func (ts *TrieStub) Close() error {
	if ts.CloseCalled != nil {
		return ts.CloseCalled()
	}

	return nil
}
