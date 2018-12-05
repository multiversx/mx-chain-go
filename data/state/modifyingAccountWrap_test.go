package state_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state/mock"
	"github.com/stretchr/testify/assert"
)

func TestModifyingDataAccountWrapInvalidValsShouldErr(t *testing.T) {
	t.Parallel()

	_, err := state.NewModifyingDataAccountWrap(nil)
	assert.NotNil(t, err)
}

func TestModifyingDataAccountRetrieveValueNilDataTrieShouldErr(t *testing.T) {
	t.Parallel()

	as, err := state.NewModifyingDataAccountWrap(mock.NewSimpleAccountWrapMock())
	assert.Nil(t, err)

	_, err = as.RetrieveValue([]byte{65, 66, 67})
	assert.NotNil(t, err)
}

func TestModifyingDataAccountRetrieveValueFoundInDirtyShouldWork(t *testing.T) {
	t.Parallel()

	mdaw, err := state.NewModifyingDataAccountWrap(mock.NewSimpleAccountWrapMock())
	assert.Nil(t, err)

	mdaw.SetDataTrie(mock.NewMockTrie())
	mdaw.DirtyData()["ABC"] = []byte{32, 33, 34}

	val, err := mdaw.RetrieveValue([]byte{65, 66, 67})
	assert.Nil(t, err)
	assert.Equal(t, []byte{32, 33, 34}, val)
}

func TestModifyingDataAccountRetrieveValueFoundInOriginalShouldWork(t *testing.T) {
	t.Parallel()

	mdaw, err := state.NewModifyingDataAccountWrap(mock.NewSimpleAccountWrapMock())
	assert.Nil(t, err)

	mdaw.SetDataTrie(mock.NewMockTrie())
	mdaw.DirtyData()["ABC"] = []byte{32, 33, 34}
	mdaw.OriginalData()["ABD"] = []byte{35, 36, 37}

	val, err := mdaw.RetrieveValue([]byte{65, 66, 68})
	assert.Nil(t, err)
	assert.Equal(t, []byte{35, 36, 37}, val)
}

func TestModifyingDataAccountRetrieveValueFoundInTrieShouldWork(t *testing.T) {
	t.Parallel()

	mdaw, err := state.NewModifyingDataAccountWrap(mock.NewSimpleAccountWrapMock())
	assert.Nil(t, err)

	mdaw.SetDataTrie(mock.NewMockTrie())
	err = mdaw.DataTrie().Update([]byte{65, 66, 69}, []byte{38, 39, 40})
	assert.Nil(t, err)
	mdaw.DirtyData()["ABC"] = []byte{32, 33, 34}
	mdaw.OriginalData()["ABD"] = []byte{35, 36, 37}

	val, err := mdaw.RetrieveValue([]byte{65, 66, 69})
	assert.Nil(t, err)
	assert.Equal(t, []byte{38, 39, 40}, val)
}

func TestModifyingDataAccountRetrieveValueMalfunctionTrieShouldErr(t *testing.T) {
	t.Parallel()

	mdaw, err := state.NewModifyingDataAccountWrap(mock.NewSimpleAccountWrapMock())
	assert.Nil(t, err)

	trie := mock.NewMockTrie()
	trie.FailGet = true
	mdaw.SetDataTrie(trie)

	err = mdaw.DataTrie().Update([]byte{65, 66, 69}, []byte{38, 39, 40})
	assert.Nil(t, err)
	mdaw.DirtyData()["ABC"] = []byte{32, 33, 34}
	mdaw.OriginalData()["ABD"] = []byte{35, 36, 37}

	val, err := mdaw.RetrieveValue([]byte{65, 66, 69})
	assert.NotNil(t, err)
	assert.Nil(t, val)
}

func TestModifyingDataAccountSaveKeyValueShouldSaveOnlyInDirty(t *testing.T) {
	t.Parallel()

	mdaw, err := state.NewModifyingDataAccountWrap(mock.NewSimpleAccountWrapMock())
	assert.Nil(t, err)

	mdaw.SetDataTrie(mock.NewMockTrie())
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

func TestModifyingDataAccountClearDataCachesValidDataShouldWork(t *testing.T) {
	t.Parallel()

	mdaw, err := state.NewModifyingDataAccountWrap(mock.NewSimpleAccountWrapMock())
	assert.Nil(t, err)

	mdaw.SetDataTrie(mock.NewMockTrie())

	assert.Equal(t, 0, len(mdaw.DirtyData()))

	//add something
	mdaw.SaveKeyValue([]byte{65, 66, 67}, []byte{32, 33, 34})
	assert.Equal(t, 1, len(mdaw.DirtyData()))

	//clear
	mdaw.ClearDataCaches()
	assert.Equal(t, 0, len(mdaw.DirtyData()))
}
