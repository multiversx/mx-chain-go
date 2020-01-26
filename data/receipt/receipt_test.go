package receipt_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/stretchr/testify/assert"
)

func TestReceipt_SettersAndGetters(t *testing.T) {
	t.Parallel()

	r := receipt.Receipt{}
	assert.False(t, r.IsInterfaceNil())

	data := []byte("data")
	value := big.NewInt(37)
	addr := []byte("addr")

	r.SetData(data)
	r.SetValue(value)
	r.SetRecvAddress(addr)
	r.SetSndAddress(addr)

	assert.Equal(t, data, r.GetData())
	assert.Equal(t, value, r.GetValue())
	assert.Equal(t, addr, r.GetRecvAddress())
	assert.Equal(t, addr, r.GetSndAddress())
	assert.Equal(t, uint64(0), r.GetNonce())
	assert.Equal(t, uint64(0), r.GetGasPrice())
	assert.Equal(t, uint64(0), r.GetGasLimit())
}

func TestReceipt_SaveLoad(t *testing.T) {
	t.Parallel()

	r := receipt.Receipt{
		Value:   big.NewInt(10),
		SndAddr: []byte("sender"),
		Data:    []byte("data"),
		TxHash:  []byte("hash"),
	}

	var b bytes.Buffer
	err := r.Save(&b)
	assert.Nil(t, err)

	loadRcpt := receipt.Receipt{}
	err = loadRcpt.Load(&b)
	assert.Nil(t, err)

	assert.Equal(t, loadRcpt, r)
}
