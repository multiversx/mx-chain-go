package smartContractResult

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestSmartContractResult_SaveLoad(t *testing.T) {
	smrS := SmartContractResult{
		Nonce:   uint64(1),
		Value:   big.NewInt(1),
		RcvAddr: []byte("receiver_address"),
		SndAddr: []byte("sender_address"),
		Data:    []byte("tx_data"),
		Code:    []byte("code"),
		TxHash:  []byte("txHash"),
	}

	var b bytes.Buffer
	smrS.Save(&b)

	loadSMR := SmartContractResult{}
	loadSMR.Load(&b)

	assert.Equal(t, smrS, loadSMR)
}
