package transaction_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_SaveLoad(t *testing.T) {
	tx := transaction.Transaction{
		Nonce:     uint64(1),
		Value:     big.NewInt(1),
		RcvAddr:   []byte("receiver_address"),
		SndAddr:   []byte("sender_address"),
		GasPrice:  uint64(10000),
		GasLimit:  uint64(1000),
		Data:      []byte("tx_data"),
		Signature: []byte("signature"),
		Challenge: []byte("challange"),
	}

	var b bytes.Buffer
	tx.Save(&b)

	loadTx := transaction.Transaction{}
	loadTx.Load(&b)

	assert.Equal(t, loadTx, tx)
}
