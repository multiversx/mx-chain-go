package state_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/mock"
	"github.com/ElrondNetwork/elrond-go/data/state"
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

	trie := &mock.TrieStub{}
	mdaw := state.NewTrackableDataTrie(trie)
	assert.NotNil(t, mdaw)

	mdaw.SetDataTrie(&mock.TrieStub{})
	err := mdaw.DataTrie().Update([]byte{65, 66, 69}, []byte{38, 39, 40})
	assert.Nil(t, err)
	mdaw.DirtyData()["ABC"] = []byte{32, 33, 34}
	mdaw.OriginalData()["ABD"] = []byte{35, 36, 37}

	val, err := mdaw.RetrieveValue([]byte{65, 66, 69})
	assert.Nil(t, err)
	assert.Equal(t, []byte{38, 39, 40}, val)
}

func TestTrackableDataAccountRetrieveValueMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	trie := &mock.TrieStub{}

	mdaw := state.NewTrackableDataTrie(trie)
	assert.NotNil(t, mdaw)

	mdaw.SetDataTrie(trie)

	err := mdaw.DataTrie().Update([]byte{65, 66, 69}, []byte{38, 39, 40})
	assert.Nil(t, err)
	mdaw.DirtyData()["ABC"] = []byte{32, 33, 34}
	mdaw.OriginalData()["ABD"] = []byte{35, 36, 37}

	val, err := mdaw.RetrieveValue([]byte{65, 66, 69})
	assert.NotNil(t, err)
	assert.Nil(t, val)
}

func TestTrackableDataAccountSaveKeyValueShouldSaveOnlyInDirty(t *testing.T) {
	t.Parallel()

	trie := &mock.TrieStub{}
	mdaw := state.NewTrackableDataTrie(trie)
	assert.NotNil(t, mdaw)

	mdaw.SetDataTrie(&mock.TrieStub{})
	mdaw.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})

	//test in dirty
	assert.Equal(t, []byte{32, 33, 34}, mdaw.DirtyData()["ABC"])
	//test in original
	assert.Nil(t, mdaw.OriginalData()["ABC"])
	//test in trie
	val, err := mdaw.DataTrie().Get([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Nil(t, val)
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
