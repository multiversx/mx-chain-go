package mock

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/data"
)

var errNotImplemented = errors.New("not implemented")

type TrieStub struct {
	GetCalled               func(key []byte) ([]byte, error)
	UpdateCalled            func(key, value []byte) error
	DeleteCalled            func(key []byte) error
	RootCalled              func() ([]byte, error)
	ProveCalled             func(key []byte) ([][]byte, error)
	VerifyProofCalled       func(proofs [][]byte, key []byte) (bool, error)
	CommitCalled            func() error
	RecreateCalled          func(root []byte) (data.Trie, error)
	DeepCloneCalled         func() (data.Trie, error)
	CancelPruneCalled       func(rootHash []byte)
	PruneCalled             func(rootHash []byte) error
	ResetOldHashesCalled    func() [][]byte
	AppendToOldHashesCalled func([][]byte)
}

func (ts *TrieStub) Get(key []byte) ([]byte, error) {
	if ts.GetCalled != nil {
		return ts.GetCalled(key)
	}

	return nil, errNotImplemented
}

func (ts *TrieStub) Update(key, value []byte) error {
	if ts.UpdateCalled != nil {
		return ts.UpdateCalled(key, value)
	}

	return errNotImplemented
}

func (ts *TrieStub) Delete(key []byte) error {
	if ts.DeleteCalled != nil {
		return ts.DeleteCalled(key)
	}

	return errNotImplemented
}

func (ts *TrieStub) Root() ([]byte, error) {
	if ts.RootCalled != nil {
		return ts.RootCalled()
	}

	return nil, errNotImplemented
}

func (ts *TrieStub) Prove(key []byte) ([][]byte, error) {
	if ts.ProveCalled != nil {
		return ts.ProveCalled(key)
	}

	return nil, errNotImplemented
}

func (ts *TrieStub) VerifyProof(proofs [][]byte, key []byte) (bool, error) {
	if ts.VerifyProofCalled != nil {
		return ts.VerifyProofCalled(proofs, key)
	}

	return false, errNotImplemented
}

func (ts *TrieStub) Commit() error {
	if ts != nil {
		return ts.CommitCalled()
	}

	return errNotImplemented
}

func (ts *TrieStub) Recreate(root []byte) (data.Trie, error) {
	if ts.RecreateCalled != nil {
		return ts.RecreateCalled(root)
	}

	return nil, errNotImplemented
}

func (ts *TrieStub) String() string {
	return "stub trie"
}

func (ts *TrieStub) DeepClone() (data.Trie, error) {
	return ts.DeepCloneCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ts *TrieStub) IsInterfaceNil() bool {
	if ts == nil {
		return true
	}
	return false
}

// CancelPrune invalidates the hashes that correspond to the given root hash from the eviction waiting list
func (ts *TrieStub) CancelPrune(rootHash []byte) {
	if ts.CancelPruneCalled != nil {
		ts.CancelPruneCalled(rootHash)
	}
}

// Prune removes from the database all the old hashes that correspond to the given root hash
func (ts *TrieStub) Prune(rootHash []byte) error {
	if ts.PruneCalled != nil {
		return ts.PruneCalled(rootHash)
	}

	return errNotImplemented
}

// ResetOldHashes resets the oldHashes and oldRoot variables and returns the old hashes
func (ts *TrieStub) ResetOldHashes() [][]byte {
	if ts.ResetOldHashesCalled != nil {
		return ts.ResetOldHashesCalled()
	}

	return nil
}

// AppendToOldHashes appends the given hashes to the trie's oldHashes variable
func (ts *TrieStub) AppendToOldHashes(hashes [][]byte) {
	if ts.AppendToOldHashesCalled != nil {
		ts.AppendToOldHashesCalled(hashes)
	}
}
