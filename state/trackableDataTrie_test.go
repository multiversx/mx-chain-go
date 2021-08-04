package state_test

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/state"
	testscommonTrie "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewTrackableDataTrie(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	trie := &testscommonTrie.TrieStub{}
	tdaw := state.NewTrackableDataTrie(identifier, trie)

	assert.False(t, check.IfNil(tdaw))
}

func TestTrackableDataTrie_RetrieveValueNilDataTrieShouldErr(t *testing.T) {
	t.Parallel()

	as := state.NewTrackableDataTrie([]byte("identifier"), nil)
	assert.NotNil(t, as)

	_, err := as.RetrieveValue([]byte("ABC"))
	assert.NotNil(t, err)
}

func TestTrackableDataTrie_RetrieveValueFoundInDirtyShouldWork(t *testing.T) {
	t.Parallel()

	stringKey := "ABC"
	identifier := []byte("identifier")
	trie := &testscommonTrie.TrieStub{}
	tdaw := state.NewTrackableDataTrie(identifier, trie)
	assert.NotNil(t, tdaw)

	tdaw.SetDataTrie(&testscommonTrie.TrieStub{})
	key := []byte(stringKey)
	val := []byte("123")

	trieVal := append(val, key...)
	trieVal = append(trieVal, identifier...)

	tdaw.DirtyData()[stringKey] = trieVal

	retrievedVal, err := tdaw.RetrieveValue(key)
	assert.Nil(t, err)
	assert.Equal(t, val, retrievedVal)
}

func TestTrackableDataTrie_RetrieveValueFoundInTrieShouldWork(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	expectedKey := []byte("key")

	expectedVal := []byte("value")
	value := append(expectedVal, expectedKey...)
	value = append(value, identifier...)

	trie := &testscommonTrie.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetCalled: func(key []byte) (b []byte, e error) {
			if bytes.Equal(key, expectedKey) {
				return value, nil
			}
			return nil, nil
		},
	}
	mdaw := state.NewTrackableDataTrie(identifier, trie)
	assert.NotNil(t, mdaw)

	valRecovered, err := mdaw.RetrieveValue(expectedKey)
	assert.Nil(t, err)
	assert.Equal(t, expectedVal, valRecovered)
}

func TestTrackableDataTrie_RetrieveValueMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	keyExpected := []byte("key")
	trie := &testscommonTrie.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetCalled: func(key []byte) (b []byte, e error) {
			return nil, errExpected
		},
	}
	mdaw := state.NewTrackableDataTrie([]byte("identifier"), trie)
	assert.NotNil(t, mdaw)

	valRecovered, err := mdaw.RetrieveValue(keyExpected)
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

	trie := &testscommonTrie.TrieStub{
		GetCalled: func(key []byte) (b []byte, e error) {
			return trieValue, nil
		},
	}
	mdaw := state.NewTrackableDataTrie([]byte("id"), trie)
	assert.NotNil(t, mdaw)

	valRecovered, err := mdaw.RetrieveValue(key)
	assert.Equal(t, retrievedTrieVal, valRecovered)
	assert.Nil(t, err)

	_ = mdaw.SaveKeyValue(key, newTrieValue)
	valRecovered, err = mdaw.RetrieveValue(key)
	assert.Equal(t, newTrieValue, valRecovered)
	assert.Nil(t, err)
}

func TestTrackableDataTrie_SaveKeyValueShouldSaveOnlyInDirty(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	keyExpected := []byte("key")
	value := []byte("value")

	expectedVal := append(value, keyExpected...)
	expectedVal = append(expectedVal, identifier...)

	trie := &testscommonTrie.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetCalled: func(key []byte) (b []byte, e error) {
			assert.Fail(t, "should not have saved directly in the trie")
			return nil, nil
		},
	}
	mdaw := state.NewTrackableDataTrie(identifier, trie)
	assert.NotNil(t, mdaw)

	_ = mdaw.SaveKeyValue(keyExpected, value)

	//test in dirty
	assert.Equal(t, expectedVal, mdaw.DirtyData()[string(keyExpected)])
}

func TestTrackableDataTrie_ClearDataCachesValidDataShouldWork(t *testing.T) {
	t.Parallel()

	trie := &testscommonTrie.TrieStub{}
	mdaw := state.NewTrackableDataTrie([]byte("identifier"), trie)
	assert.NotNil(t, mdaw)

	mdaw.SetDataTrie(&testscommonTrie.TrieStub{})

	assert.Equal(t, 0, len(mdaw.DirtyData()))

	//add something
	_ = mdaw.SaveKeyValue([]byte("ABC"), []byte("123"))
	assert.Equal(t, 1, len(mdaw.DirtyData()))

	//clear
	mdaw.ClearDataCaches()
	assert.Equal(t, 0, len(mdaw.DirtyData()))
}

func TestTrackableDataTrie_SetAndGetDataTrie(t *testing.T) {
	t.Parallel()

	trie := &testscommonTrie.TrieStub{}
	mdaw := state.NewTrackableDataTrie([]byte("identifier"), trie)

	newTrie := &testscommonTrie.TrieStub{}
	mdaw.SetDataTrie(newTrie)
	assert.Equal(t, newTrie, mdaw.DataTrie())
}

func TestTrackableDataTrie_SaveKeyValueTooBig(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	trie := &testscommonTrie.TrieStub{}
	tdaw := state.NewTrackableDataTrie(identifier, trie)

	err := tdaw.SaveKeyValue([]byte("key"), make([]byte, core.MaxLeafSize+1))
	assert.Equal(t, err, data.ErrLeafSizeTooBig)
}
