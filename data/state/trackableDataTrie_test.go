package state_test

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestTrackableDataAccountRetrieveValueNilDataTrieShouldErr(t *testing.T) {
	t.Parallel()

	as := state.NewTrackableDataTrie([]byte("identifier"), nil)
	assert.NotNil(t, as)

	_, err := as.RetrieveValue([]byte("ABC"))
	assert.NotNil(t, err)
}

func TestTrackableDataAccountRetrieveValueFoundInDirtyShouldWork(t *testing.T) {
	t.Parallel()

	stringKey := "ABC"
	identifier := []byte("identifier")
	trie := &mock.TrieStub{}
	tdaw := state.NewTrackableDataTrie(identifier, trie)
	assert.NotNil(t, tdaw)

	tdaw.SetDataTrie(&mock.TrieStub{})
	key := []byte(stringKey)
	val := []byte("123")

	trieVal := append(val, key...)
	trieVal = append(trieVal, identifier...)

	tdaw.DirtyData()[stringKey] = trieVal

	retrievedVal, err := tdaw.RetrieveValue(key)
	assert.Nil(t, err)
	assert.Equal(t, val, retrievedVal)
}

func TestTrackableDataAccountRetrieveValueFoundInOriginalShouldWork(t *testing.T) {
	t.Parallel()

	originalKeyString := "ABD"
	identifier := []byte("identifier")
	trie := &mock.TrieStub{}
	mdaw := state.NewTrackableDataTrie(identifier, trie)
	assert.NotNil(t, mdaw)

	originalKey := []byte(originalKeyString)
	dirtyVal := []byte("123")

	expectedVal := []byte("456")
	originalVal := append(expectedVal, originalKey...)
	originalVal = append(originalVal, identifier...)

	mdaw.SetDataTrie(&mock.TrieStub{})
	mdaw.DirtyData()["ABC"] = dirtyVal
	mdaw.OriginalData()[originalKeyString] = originalVal

	val, err := mdaw.RetrieveValue(originalKey)
	assert.Nil(t, err)
	assert.Equal(t, expectedVal, val)
}

func TestTrackableDataAccountRetrieveValueFoundInTrieShouldWork(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	expectedKey := []byte("key")

	expectedVal := []byte("value")
	value := append(expectedVal, expectedKey...)
	value = append(value, identifier...)

	trie := &mock.TrieStub{
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

	mdaw.DirtyData()[string(expectedKey)] = value

	valRecovered, err := mdaw.RetrieveValue(expectedKey)
	assert.Nil(t, err)
	assert.Equal(t, expectedVal, valRecovered)
}

func TestTrackableDataAccountRetrieveValueMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	keyExpected := []byte("key")
	trie := &mock.TrieStub{
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

func TestTrackableDataAccountSaveKeyValueShouldSaveOnlyInDirty(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	keyExpected := []byte("key")
	value := []byte("value")

	expectedVal := append(value, keyExpected...)
	expectedVal = append(expectedVal, identifier...)

	trie := &mock.TrieStub{
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

	mdaw.SaveKeyValue(keyExpected, value)

	//test in dirty
	assert.Equal(t, expectedVal, mdaw.DirtyData()[string(keyExpected)])
	//test in original
	assert.Nil(t, mdaw.OriginalData()[string(keyExpected)])
}

func TestTrackableDataAccountClearDataCachesValidDataShouldWork(t *testing.T) {
	t.Parallel()

	trie := &mock.TrieStub{}
	mdaw := state.NewTrackableDataTrie([]byte("identifier"), trie)
	assert.NotNil(t, mdaw)

	mdaw.SetDataTrie(&mock.TrieStub{})

	assert.Equal(t, 0, len(mdaw.DirtyData()))

	//add something
	mdaw.SaveKeyValue([]byte("ABC"), []byte("123"))
	assert.Equal(t, 1, len(mdaw.DirtyData()))

	//clear
	mdaw.ClearDataCaches()
	assert.Equal(t, 0, len(mdaw.DirtyData()))
}
