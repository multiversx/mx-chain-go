package smartContractResult_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/stretchr/testify/assert"
)

func TestSmartContractResult_SettersAndGetters(t *testing.T) {
	t.Parallel()

	nonce := uint64(5)
	gasPrice := uint64(1)
	gasLimit := uint64(10)
	scr := smartContractResult.SmartContractResult{
		Nonce:    nonce,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}

	rcvAddr := []byte("rcv address")
	sndAddr := []byte("snd address")
	value := big.NewInt(37)
	data := []byte("unStake")

	scr.SetRecvAddress(rcvAddr)
	scr.SetSndAddress(sndAddr)
	scr.SetValue(value)
	scr.SetData(data)

	assert.Equal(t, sndAddr, scr.GetSndAddress())
	assert.Equal(t, rcvAddr, scr.GetRecvAddress())
	assert.Equal(t, value, scr.GetValue())
	assert.Equal(t, data, scr.GetData())
	assert.Equal(t, gasLimit, scr.GetGasLimit())
	assert.Equal(t, gasPrice, scr.GetGasPrice())
	assert.Equal(t, nonce, scr.GetNonce())
}

func TestTrimSlicePtr(t *testing.T) {
	t.Parallel()

	scrSlice := make([]*smartContractResult.SmartContractResult, 0, 5)
	scr1 := &smartContractResult.SmartContractResult{Nonce: 3}
	scr2 := &smartContractResult.SmartContractResult{Nonce: 5}

	scrSlice = append(scrSlice, scr1)
	scrSlice = append(scrSlice, scr2)

	assert.Equal(t, 2, len(scrSlice))
	assert.Equal(t, 5, cap(scrSlice))

	scrSlice = smartContractResult.TrimSlicePtr(scrSlice)

	assert.Equal(t, 2, len(scrSlice))
	assert.Equal(t, 2, len(scrSlice))
}
