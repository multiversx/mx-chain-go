package leavesRetriever_test

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/keyValStorage"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	trieTest "github.com/multiversx/mx-chain-go/testscommon/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/multiversx/mx-chain-go/trie/leavesRetriever"
	"github.com/stretchr/testify/assert"
)

func TestNewLeavesRetriever(t *testing.T) {
	t.Parallel()

	t.Run("nil db", func(t *testing.T) {
		t.Parallel()

		lr, err := leavesRetriever.NewLeavesRetriever(nil, &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, 100)
		assert.Nil(t, lr)
		assert.Equal(t, leavesRetriever.ErrNilDB, err)
	})
	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		lr, err := leavesRetriever.NewLeavesRetriever(testscommon.NewMemDbMock(), nil, &hashingMocks.HasherMock{}, 100)
		assert.Nil(t, lr)
		assert.Equal(t, leavesRetriever.ErrNilMarshaller, err)
	})
	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		lr, err := leavesRetriever.NewLeavesRetriever(testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, nil, 100)
		assert.Nil(t, lr)
		assert.Equal(t, leavesRetriever.ErrNilHasher, err)
	})
	t.Run("new leaves retriever", func(t *testing.T) {
		t.Parallel()

		var lr common.TrieLeavesRetriever
		assert.True(t, check.IfNil(lr))

		lr, err := leavesRetriever.NewLeavesRetriever(testscommon.NewMemDbMock(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, 100)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(lr))
	})
}

func TestLeavesRetriever_GetLeaves(t *testing.T) {
	t.Parallel()

	tr := trieTest.GetNewTrie()
	trieTest.AddDataToTrie(tr, 25)
	rootHash, _ := tr.RootHash()
	leafParser := &trieMock.TrieLeafParserStub{
		ParseLeafCalled: func(key []byte, val []byte, version core.TrieNodeVersion) (core.KeyValueHolder, error) {
			return keyValStorage.NewKeyValStorage(key, val), nil
		},
	}
	lr, _ := leavesRetriever.NewLeavesRetriever(tr.GetStorageManager(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, 100000)
	leaves, newIteratorState, err := lr.GetLeaves(10, [][]byte{rootHash}, leafParser, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 10, len(leaves))
	assert.Equal(t, 8, len(newIteratorState))

	newLr, _ := leavesRetriever.NewLeavesRetriever(tr.GetStorageManager(), &marshallerMock.MarshalizerMock{}, &hashingMocks.HasherMock{}, 100000)
	leaves, newIteratorState, err = newLr.GetLeaves(10, newIteratorState, leafParser, context.Background())
	assert.Nil(t, err)
	assert.Equal(t, 10, len(leaves))
	assert.Equal(t, 3, len(newIteratorState))
}
