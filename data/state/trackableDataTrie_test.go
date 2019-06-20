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

	as := state.NewTrackableDataTrie(nil)
	assert.NotNil(t, as)

	_, err := as.RetrieveValue([]byte{65, 66, 67})
	assert.NotNil(t, err)
}

func TestTrackableDataAccountRetrieveValueFoundInDirtyShouldWork(t *testing.T) {
	t.Parallel()

	trie := &mock.TrieStub{}
	tdaw := state.NewTrackableDataTrie(trie)
	assert.NotNil(t, tdaw)

	tdaw.SetDataTrie(&mock.TrieStub{})
	tdaw.DirtyData()["ABC"] = []byte{32, 33, 34}

	val, err := tdaw.RetrieveValue([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{32, 33, 34}, val)
}

func TestTrackableDataAccountRetrieveValueFoundInOriginalShouldWork(t *testing.T) {
	t.Parallel()

	trie := &mock.TrieStub{}
	mdaw := state.NewTrackableDataTrie(trie)
	assert.NotNil(t, mdaw)

	mdaw.SetDataTrie(&mock.TrieStub{})
	mdaw.DirtyData()["ABC"] = []byte{32, 33, 34}
	mdaw.OriginalData()["ABD"] = []byte{35, 36, 37}

	val, err := mdaw.RetrieveValue([]byte{65, 66, 68})
	assert.Nil(t, err)
	assert.Equal(t, []byte{35, 36, 37}, val)
}

func TestTrackableDataAccountRetrieveValueFoundInTrieShouldWork(t *testing.T) {
	t.Parallel()

	keyExpected := []byte("key")
	value := []byte("value")
	trie := &mock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetCalled: func(key []byte) (b []byte, e error) {
			if bytes.Equal(key, keyExpected) {
				return value, nil
			}
			return nil, nil
		},
	}
	mdaw := state.NewTrackableDataTrie(trie)
	assert.NotNil(t, mdaw)

	mdaw.DirtyData()[string(keyExpected)] = value

	valRecovered, err := mdaw.RetrieveValue(keyExpected)
	assert.Nil(t, err)
	assert.Equal(t, valRecovered, value)
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
	mdaw := state.NewTrackableDataTrie(trie)
	assert.NotNil(t, mdaw)

	valRecovered, err := mdaw.RetrieveValue(keyExpected)
	assert.Equal(t, errExpected, err)
	assert.Nil(t, valRecovered)
}

func TestTrackableDataAccountSaveKeyValueShouldSaveOnlyInDirty(t *testing.T) {
	t.Parallel()

	keyExpected := []byte("key")
	value := []byte("value")
	trie := &mock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetCalled: func(key []byte) (b []byte, e error) {
			assert.Fail(t, "should not have saved directly in the trie")
			return nil, nil
		},
	}
	mdaw := state.NewTrackableDataTrie(trie)
	assert.NotNil(t, mdaw)

	mdaw.SaveKeyValue(keyExpected, value)

	//test in dirty
	assert.Equal(t, value, mdaw.DirtyData()[string(keyExpected)])
	//test in original
	assert.Nil(t, mdaw.OriginalData()[string(keyExpected)])
}

func TestTrackableDataAccountClearDataCachesValidDataShouldWork(t *testing.T) {
	t.Parallel()

	trie := &mock.TrieStub{}
	mdaw := state.NewTrackableDataTrie(trie)
	assert.NotNil(t, mdaw)

	mdaw.SetDataTrie(&mock.TrieStub{})

	assert.Equal(t, 0, len(mdaw.DirtyData()))

	//add something
	mdaw.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	assert.Equal(t, 1, len(mdaw.DirtyData()))

	//clear
	mdaw.ClearDataCaches()
	assert.Equal(t, 0, len(mdaw.DirtyData()))
}
