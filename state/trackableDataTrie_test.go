package state_test

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	trieMock "github.com/ElrondNetwork/elrond-go/testscommon/trie"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewTrackableDataTrie(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	trie := &trieMock.TrieStub{}
	tdaw, err := state.NewTrackableDataTrie(identifier, trie, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})
	assert.Nil(t, err)
	assert.False(t, check.IfNil(tdaw))
}

func TestNewTrackableDataTrie_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	trie := &trieMock.TrieStub{}
	tdaw, err := state.NewTrackableDataTrie(identifier, trie, nil, &testscommon.MarshalizerMock{})
	assert.Equal(t, state.ErrNilHasher, err)
	assert.True(t, check.IfNil(tdaw))
}

func TestNewTrackableDataTrie_NilMarshallerShouldErr(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	trie := &trieMock.TrieStub{}
	tdaw, err := state.NewTrackableDataTrie(identifier, trie, &hashingMocks.HasherMock{}, nil)
	assert.Equal(t, state.ErrNilMarshalizer, err)
	assert.True(t, check.IfNil(tdaw))
}

func TestTrackableDataTrie_RetrieveValueNilDataTrieShouldErr(t *testing.T) {
	t.Parallel()

	as, err := state.NewTrackableDataTrie([]byte("identifier"), nil, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})
	assert.Nil(t, err)
	assert.NotNil(t, as)

	_, err = as.RetrieveValue([]byte("ABC"))
	assert.NotNil(t, err)
}

func TestTrackableDataTrie_RetrieveValueFoundInTrieShouldWork(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	expectedKey := []byte("key")

	expectedVal := []byte("value")
	value := append(expectedVal, expectedKey...)
	value = append(value, identifier...)

	trie := &trieMock.TrieStub{
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
	mdaw, _ := state.NewTrackableDataTrie(identifier, trie, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})
	assert.NotNil(t, mdaw)

	valRecovered, err := mdaw.RetrieveValue(expectedKey)
	assert.Nil(t, err)
	assert.Equal(t, expectedVal, valRecovered)
}

func TestTrackableDataTrie_RetrieveValueMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	errExpected := errors.New("expected err")
	keyExpected := []byte("key")
	trie := &trieMock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetCalled: func(key []byte) (b []byte, e error) {
			return nil, errExpected
		},
	}
	mdaw, _ := state.NewTrackableDataTrie([]byte("identifier"), trie, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})
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

	trie := &trieMock.TrieStub{
		GetCalled: func(key []byte) (b []byte, e error) {
			return trieValue, nil
		},
	}
	mdaw, _ := state.NewTrackableDataTrie([]byte("id"), trie, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})
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

	trie := &trieMock.TrieStub{
		UpdateCalled: func(key, value []byte) error {
			return nil
		},
		GetCalled: func(key []byte) (b []byte, e error) {
			assert.Fail(t, "should not have saved directly in the trie")
			return nil, nil
		},
	}
	mdaw, _ := state.NewTrackableDataTrie(identifier, trie, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})
	assert.NotNil(t, mdaw)

	_ = mdaw.SaveKeyValue(keyExpected, value)

	// test in dirty
	retrievedVal, err := mdaw.RetrieveValue(keyExpected)
	assert.Nil(t, err)
	assert.Equal(t, value, retrievedVal)
}

func TestTrackableDataTrie_SetAndGetDataTrie(t *testing.T) {
	t.Parallel()

	trie := &trieMock.TrieStub{}
	mdaw, _ := state.NewTrackableDataTrie([]byte("identifier"), trie, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})

	newTrie := &trieMock.TrieStub{}
	mdaw.SetDataTrie(newTrie)
	assert.Equal(t, newTrie, mdaw.DataTrie())
}

func TestTrackableDataTrie_SaveKeyValueTooBig(t *testing.T) {
	t.Parallel()

	identifier := []byte("identifier")
	trie := &trieMock.TrieStub{}
	tdaw, _ := state.NewTrackableDataTrie(identifier, trie, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})

	err := tdaw.SaveKeyValue([]byte("key"), make([]byte, core.MaxLeafSize+1))
	assert.Equal(t, err, data.ErrLeafSizeTooBig)
}
