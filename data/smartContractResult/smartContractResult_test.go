package smartContractResult_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/smartContractResult"
	"github.com/stretchr/testify/assert"
)

func TestSmartContractResult_SaveLoad(t *testing.T) {
	smrS := smartContractResult.SmartContractResult{
		Nonce:   uint64(1),
		Value:   big.NewInt(1),
		RcvAddr: []byte("receiver_address"),
		SndAddr: []byte("sender_address"),
		Data:    []byte("scr_data"),
		Code:    []byte("code"),
		TxHash:  []byte("scrHash"),
	}

	var b bytes.Buffer
	smrS.Save(&b)

	loadSMR := smartContractResult.SmartContractResult{}
	loadSMR.Load(&b)

	assert.Equal(t, smrS, loadSMR)
}

func TestSmartContractResult_GetData(t *testing.T) {
	t.Parallel()

	data := []byte("data")
	scr := &smartContractResult.SmartContractResult{Data: data}

	assert.Equal(t, data, scr.Data)
}

func TestSmartContractResult_GetRecvAddr(t *testing.T) {
	t.Parallel()

	data := []byte("data")
	scr := &smartContractResult.SmartContractResult{RcvAddr: data}

	assert.Equal(t, data, scr.RcvAddr)
}

func TestSmartContractResult_GetSndAddr(t *testing.T) {
	t.Parallel()

	data := []byte("data")
	scr := &smartContractResult.SmartContractResult{SndAddr: data}

	assert.Equal(t, data, scr.SndAddr)
}

func TestSmartContractResult_GetValue(t *testing.T) {
	t.Parallel()

	value := big.NewInt(10)
	scr := &smartContractResult.SmartContractResult{Value: value}

	assert.Equal(t, value, scr.Value)
}

func TestSmartContractResult_SetData(t *testing.T) {
	t.Parallel()

	data := []byte("data")
	scr := &smartContractResult.SmartContractResult{}
	scr.SetData(data)

	assert.Equal(t, data, scr.Data)
}

func TestSmartContractResult_SetRecvAddr(t *testing.T) {
	t.Parallel()

	data := []byte("data")
	scr := &smartContractResult.SmartContractResult{}
	scr.SetRecvAddress(data)

	assert.Equal(t, data, scr.RcvAddr)
}

func TestSmartContractResult_SetSndAddr(t *testing.T) {
	t.Parallel()

	data := []byte("data")
	scr := &smartContractResult.SmartContractResult{}
	scr.SetSndAddress(data)

	assert.Equal(t, data, scr.SndAddr)
}

func TestSmartContractResult_SetValue(t *testing.T) {
	t.Parallel()

	value := big.NewInt(10)
	scr := &smartContractResult.SmartContractResult{}
	scr.SetValue(value)

	assert.Equal(t, value, scr.Value)
}
