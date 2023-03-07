package state_test

import (
	"bytes"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewTrackableDataTrie(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	trie := &trieMock.TrieStub{}
	tdaw := state.NewTrackableDataTrie(identifier, trie)

	assert.False(t, check.IfNil(tdaw))
}

func TestTrackableDataTrie_RetrieveValueNilDataTrieShouldErr(t *testing.T) {
	t.Parallel()

	as := state.NewTrackableDataTrie([]byte("identifier"), nil)
	assert.NotNil(t, as)

	_, trieDepth, err := as.RetrieveValue([]byte("ABC"))
	assert.NotNil(t, err)
	assert.Equal(t, uint32(0), trieDepth)
}

func TestTrackableDataTrie_RetrieveValueFoundInTrieShouldWork(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	expectedKey := []byte("key")

	expectedVal := []byte("value")
	value := append(expectedVal, expectedKey...)
	value = append(value, identifier...)
	expectedTrieDepth := uint32(5)

	trie := &trieMock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetCalled: func(key []byte) ([]byte, uint32, error) {
			if bytes.Equal(key, expectedKey) {
				return value, expectedTrieDepth, nil
			}
			return nil, 0, nil
		},
	}
	mdaw := state.NewTrackableDataTrie(identifier, trie)
	assert.NotNil(t, mdaw)

	valRecovered, trieDepth, err := mdaw.RetrieveValue(expectedKey)
	assert.Nil(t, err)
	assert.Equal(t, expectedVal, valRecovered)
	assert.Equal(t, expectedTrieDepth, trieDepth)
}

func TestTrackableDataTrie_RetrieveValueMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	keyExpected := []byte("key")
	trie := &trieMock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			return nil, 0, errExpected
		},
	}
	mdaw := state.NewTrackableDataTrie([]byte("identifier"), trie)
	assert.NotNil(t, mdaw)

	valRecovered, _, err := mdaw.RetrieveValue(keyExpected)
	assert.Equal(t, errExpected, err)
	assert.Nil(t, valRecovered)
}

func TestTrackableDataTrie_RetrieveValueShouldCheckDirtyDataFirst(t *testing.T) {
	t.Parallel()

	identifier := []byte("id")
	key := []byte("key")
	tail := append(key, identifier...)
	retrievedTrieVal := []byte("value")
	trieValue := append(retrievedTrieVal, tail...)
	newTrieValue := []byte("new trie value")
	expectedTrieDepth := uint32(5)

	trie := &trieMock.TrieStub{
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			return trieValue, expectedTrieDepth, nil
		},
	}
	mdaw := state.NewTrackableDataTrie([]byte("id"), trie)
	assert.NotNil(t, mdaw)

	valRecovered, trieDepth, err := mdaw.RetrieveValue(key)
	assert.Equal(t, retrievedTrieVal, valRecovered)
	assert.Nil(t, err)
	assert.Equal(t, expectedTrieDepth, trieDepth)

	_ = mdaw.SaveKeyValue(key, newTrieValue)
	valRecovered, trieDepth, err = mdaw.RetrieveValue(key)
	assert.Equal(t, newTrieValue, valRecovered)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), trieDepth)
}

func TestTrackableDataTrie_SaveKeyValueShouldSaveOnlyInDirty(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	keyExpected := []byte("key")
	value := []byte("value")

	trie := &trieMock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetCalled: func(_ []byte) ([]byte, uint32, error) {
			assert.Fail(t, "should not have saved directly in the trie")
			return nil, 0, nil
		},
	}
	mdaw := state.NewTrackableDataTrie(identifier, trie)
	assert.NotNil(t, mdaw)

	_ = mdaw.SaveKeyValue(keyExpected, value)

	// test in dirty
	retrievedVal, _, err := mdaw.RetrieveValue(keyExpected)
	assert.Nil(t, err)
	assert.Equal(t, value, retrievedVal)
}

func TestTrackableDataTrie_SetAndGetDataTrie(t *testing.T) {
	t.Parallel()

	trie := &trieMock.TrieStub{}
	mdaw := state.NewTrackableDataTrie([]byte("identifier"), trie)

	newTrie := &trieMock.TrieStub{}
	mdaw.SetDataTrie(newTrie)
	assert.Equal(t, newTrie, mdaw.DataTrie())
}

func TestTrackableDataTrie_SaveKeyValueTooBig(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	trie := &trieMock.TrieStub{}
	tdaw := state.NewTrackableDataTrie(identifier, trie)

	err := tdaw.SaveKeyValue([]byte("key"), make([]byte, core.MaxLeafSize+1))
	assert.Equal(t, err, data.ErrLeafSizeTooBig)
}
