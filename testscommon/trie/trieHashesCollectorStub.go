package trie

import "github.com/multiversx/mx-chain-go/common"

// TrieHashesCollectorStub is a stub for the TrieHashesCollector interface.
type TrieHashesCollectorStub struct {
	AddDirtyHashCalled      func(hash []byte)
	GetDirtyHashesCalled    func() common.ModifiedHashes
	AddObsoleteHashesCalled func(oldRootHash []byte, oldHashes [][]byte)
	GetCollectedDataCalled  func() ([]byte, common.ModifiedHashes, common.ModifiedHashes)
	CleanCalled             func()
}

// AddDirtyHash -
func (h *TrieHashesCollectorStub) AddDirtyHash(hash []byte) {
	if h.AddDirtyHashCalled != nil {
		h.AddDirtyHashCalled(hash)
	}
}

// GetDirtyHashes -
func (h *TrieHashesCollectorStub) GetDirtyHashes() common.ModifiedHashes {
	if h.GetDirtyHashesCalled != nil {
		return h.GetDirtyHashesCalled()
	}
	return nil
}

// AddObsoleteHashes -
func (h *TrieHashesCollectorStub) AddObsoleteHashes(oldRootHash []byte, oldHashes [][]byte) {
	if h.AddObsoleteHashesCalled != nil {
		h.AddObsoleteHashesCalled(oldRootHash, oldHashes)
	}
}

// GetCollectedData -
func (h *TrieHashesCollectorStub) GetCollectedData() ([]byte, common.ModifiedHashes, common.ModifiedHashes) {
	if h.GetCollectedDataCalled != nil {
		return h.GetCollectedDataCalled()
	}
	return nil, nil, nil
}

// Clean -
func (h *TrieHashesCollectorStub) Clean() {
	if h.CleanCalled != nil {
		h.CleanCalled()
	}
}

// IsInterfaceNil -
func (h *TrieHashesCollectorStub) IsInterfaceNil() bool {
	return h == nil
}
