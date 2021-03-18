package smartContractResult_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
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

	scr.SetRcvAddr(rcvAddr)
	scr.SetSndAddr(sndAddr)
	scr.SetValue(value)
	scr.SetData(data)

	assert.Equal(t, sndAddr, scr.GetSndAddr())
	assert.Equal(t, rcvAddr, scr.GetRcvAddr())
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

func TestSmartContractResult_CheckIntegrityShouldWork(t *testing.T) {
	t.Parallel()

	scr := &smartContractResult.SmartContractResult{
		Nonce:      1,
		Value:      big.NewInt(10),
		GasPrice:   1,
		GasLimit:   10,
		Data:       []byte("data"),
		RcvAddr:    []byte("rcv-address"),
		SndAddr:    []byte("snd-address"),
		PrevTxHash: []byte("prev-hash"),
	}

	err := scr.CheckIntegrity()
	assert.Nil(t, err)
}

func TestSmartContractResult_CheckIntegrityShouldErr(t *testing.T) {
	t.Parallel()

	scr := &smartContractResult.SmartContractResult{
		Nonce: 1,
		Data:  []byte("data"),
	}

	err := scr.CheckIntegrity()
	assert.Equal(t, data.ErrNilRcvAddr, err)

	scr.RcvAddr = []byte("rcv-address")

	err = scr.CheckIntegrity()
	assert.Equal(t, data.ErrNilSndAddr, err)

	scr.SndAddr = []byte("snd-address")

	err = scr.CheckIntegrity()
	assert.Equal(t, data.ErrNilValue, err)

	scr.Value = big.NewInt(-1)

	err = scr.CheckIntegrity()
	assert.Equal(t, data.ErrNegativeValue, err)

	scr.Value = big.NewInt(10)

	err = scr.CheckIntegrity()
	assert.Equal(t, data.ErrNilTxHash, err)
}
