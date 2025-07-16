package hashesCollector_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state/hashesCollector"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestNewHashesCollector(t *testing.T) {
	t.Parallel()

	hc, err := hashesCollector.NewHashesCollector(nil)
	assert.True(t, check.IfNil(hc))
	assert.Equal(t, hashesCollector.ErrNilTrieHashesCollector, err)

	hc, err = hashesCollector.NewHashesCollector(&trie.TrieHashesCollectorStub{})
	assert.False(t, check.IfNil(hc))
	assert.Nil(t, err)
	assert.Nil(t, hc.GetOldRootHash())
}

func TestHashesCollector_AddObsoleteHashes(t *testing.T) {
	t.Parallel()

	addObsoleteHashesCalled := false
	oldRootHash := []byte("oldRootHash")
	oldHashes := [][]byte{[]byte("oldHash1"), []byte("oldHash2")}
	hc := &trie.TrieHashesCollectorStub{
		AddObsoleteHashesCalled: func(oldRootHash []byte, oldHashes [][]byte) {
			assert.Equal(t, oldRootHash, oldRootHash)
			assert.Equal(t, oldHashes, oldHashes)
			addObsoleteHashesCalled = true
		},
	}
	wrappedHc, _ := hashesCollector.NewHashesCollector(hc)

	wrappedHc.AddObsoleteHashes(oldRootHash, oldHashes)
	assert.True(t, addObsoleteHashesCalled)
	assert.Equal(t, oldRootHash, wrappedHc.GetOldRootHash())
}

func TestHashesCollector_GetCollectedData(t *testing.T) {
	t.Parallel()

	getCollectedDataCalled := false
	oldHashes := common.ModifiedHashes{"oldHash1": {}, "oldHash2": {}}
	newHashes := common.ModifiedHashes{"newHash1": {}, "newHash2": {}}
	hc := &trie.TrieHashesCollectorStub{
		GetCollectedDataCalled: func() ([]byte, common.ModifiedHashes, common.ModifiedHashes) {
			getCollectedDataCalled = true
			return []byte("oldRootHash"), oldHashes, newHashes
		},
	}
	wrappedHc, _ := hashesCollector.NewHashesCollector(hc)

	oldRootHash, collectedOldHashes, collectedNewHashes := wrappedHc.GetCollectedData()
	assert.True(t, getCollectedDataCalled)
	assert.Nil(t, oldRootHash)
	assert.Equal(t, oldHashes, collectedOldHashes)
	assert.Equal(t, newHashes, collectedNewHashes)

	wrappedHc.AddObsoleteHashes([]byte("oldRootHash1"), [][]byte{[]byte("oldHash1"), []byte("oldHash2")})
	oldRootHash, collectedOldHashes, collectedNewHashes = wrappedHc.GetCollectedData()
	assert.Equal(t, []byte("oldRootHash1"), oldRootHash)
	assert.Equal(t, oldHashes, collectedOldHashes)
	assert.Equal(t, newHashes, collectedNewHashes)
}
