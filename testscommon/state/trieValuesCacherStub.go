package state

import "github.com/multiversx/mx-chain-core-go/core"

// TrieValuesCacherStub -
type TrieValuesCacherStub struct {
	PutCalled   func(key []byte, value core.TrieData)
	GetCalled   func(key []byte) (core.TrieData, bool)
	CleanCalled func()
}

// Put -
func (tvc *TrieValuesCacherStub) Put(key []byte, value core.TrieData) {
	if tvc.PutCalled != nil {
		tvc.PutCalled(key, value)
	}
}

// Get -
func (tvc *TrieValuesCacherStub) Get(key []byte) (core.TrieData, bool) {
	if tvc.GetCalled != nil {
		return tvc.GetCalled(key)
	}

	return core.TrieData{}, false
}

// Clean -
func (tvc *TrieValuesCacherStub) Clean() {
	if tvc.CleanCalled != nil {
		tvc.CleanCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (tvc *TrieValuesCacherStub) IsInterfaceNil() bool {
	return tvc == nil
}
