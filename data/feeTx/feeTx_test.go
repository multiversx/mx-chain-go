package feeTx_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/feeTx"
	"github.com/stretchr/testify/assert"
)

func TestFeeTx_SaveLoad(t *testing.T) {
	smrS := feeTx.FeeTx{
		Nonce:   uint64(1),
		Value:   big.NewInt(1),
		RcvAddr: []byte("receiver_address"),
		ShardId: 10,
	}

	var b bytes.Buffer
	err := smrS.Save(&b)
	assert.Nil(t, err)

	loadSMR := feeTx.FeeTx{}
	err = loadSMR.Load(&b)
	assert.Nil(t, err)

	assert.Equal(t, smrS, loadSMR)
}

func TestFeeTx_GetRecvAddr(t *testing.T) {
	t.Parallel()

	data := []byte("data")
	scr := &feeTx.FeeTx{RcvAddr: data}

	assert.Equal(t, data, scr.RcvAddr)
}

func TestFeeTx_GetValue(t *testing.T) {
	t.Parallel()

	value := big.NewInt(10)
	scr := &feeTx.FeeTx{Value: value}

	assert.Equal(t, value, scr.Value)
}

func TestFeeTx_SetRecvAddr(t *testing.T) {
	t.Parallel()

	data := []byte("data")
	scr := &feeTx.FeeTx{}
	scr.SetRecvAddress(data)

	assert.Equal(t, data, scr.RcvAddr)
}

func TestFeeTx_SetValue(t *testing.T) {
	t.Parallel()

	value := big.NewInt(10)
	scr := &feeTx.FeeTx{}
	scr.SetValue(value)

	assert.Equal(t, value, scr.Value)
}
