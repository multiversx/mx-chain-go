package fuzzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	f := New()
	assert.NotNil(t, f)
	assert.NoError(t, f.GetError())
}

func TestAddrBook(t *testing.T) {
	f := New().SetNumAcc(15)
	assert.NotNil(t, f)
	assert.NoError(t, f.GetError())

	//generate
	txs, err := f.GetTransactions(10)
	assert.NotNil(t, txs)
	assert.NoError(t, err)

	assert.Equal(t, len(f.addressBook), 15)

	for i, a := range f.addressBook {
		t.Logf("addr[%d] = %#v", i, a)
		assert.Equal(t, 32, len(a))
	}

	for i, tx := range txs {
		t.Logf("tx[%d] = %#v", i, tx)
	}
}

func TestDataSize(t *testing.T) {
	f := New().SetDataSzBounds(0, 1000)
	assert.NotNil(t, f)
	assert.NoError(t, f.GetError())

	// test out data size
}
