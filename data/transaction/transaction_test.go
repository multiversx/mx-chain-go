package transaction_test

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_SettersAndGetters(t *testing.T) {
	t.Parallel()

	nonce := uint64(37)
	txData := []byte("data")
	value := big.NewInt(12)
	gasPrice := uint64(1)
	gasLimit := uint64(5)
	sender := []byte("sndr")
	receiver := []byte("receiver")

	tx := &transaction.Transaction{
		Nonce:    nonce,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
	}
	assert.False(t, check.IfNil(tx))

	tx.SetSndAddr(sender)
	tx.SetData(txData)
	tx.SetValue(value)
	tx.SetRcvAddr(receiver)

	assert.Equal(t, nonce, tx.GetNonce())
	assert.Equal(t, value, tx.GetValue())
	assert.Equal(t, txData, tx.GetData())
	assert.Equal(t, gasPrice, tx.GetGasPrice())
	assert.Equal(t, gasLimit, tx.GetGasLimit())
	assert.Equal(t, sender, tx.GetSndAddr())
	assert.Equal(t, receiver, tx.GetRcvAddr())
}

func TestTransaction_MarshalUnmarshalJsonShouldWork(t *testing.T) {
	t.Parallel()

	value := big.NewInt(445566)
	tx := &transaction.Transaction{
		Nonce:     112233,
		Value:     new(big.Int).Set(value),
		RcvAddr:   []byte("receiver"),
		SndAddr:   []byte("sender"),
		GasPrice:  1234,
		GasLimit:  5678,
		Data:      []byte("data"),
		Signature: []byte("signature"),
	}

	buff, err := json.Marshal(tx)
	assert.Nil(t, err)
	txRecovered := &transaction.Transaction{}
	err = json.Unmarshal(buff, txRecovered)
	assert.Nil(t, err)
	assert.Equal(t, tx, txRecovered)

	buffAsString := string(buff)
	assert.Contains(t, buffAsString, "\""+value.String()+"\"")
}

func TestTransaction_TrimsSlicePtr(t *testing.T) {
	t.Parallel()

	tx1 := transaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(10),
		RcvAddr:   []byte("rcv"),
		SndAddr:   []byte("snd"),
		GasPrice:  1,
		GasLimit:  10,
		Data:      []byte("data"),
		Signature: []byte("sign"),
	}

	tx2 := tx1
	tx2.Nonce = 2

	input := make([]*transaction.Transaction, 0, 5)
	input = append(input, &tx1)
	input = append(input, &tx2)

	assert.Equal(t, 2, len(input))
	assert.Equal(t, 5, cap(input))

	input = transaction.TrimSlicePtr(input)

	assert.Equal(t, 2, len(input))
	assert.Equal(t, 2, cap(input))
}

func TestTransaction_TrimsSliceHandler(t *testing.T) {
	t.Parallel()

	tx1 := transaction.Transaction{
		Nonce:     1,
		Value:     big.NewInt(10),
		RcvAddr:   []byte("rcv"),
		SndAddr:   []byte("snd"),
		GasPrice:  1,
		GasLimit:  10,
		Data:      []byte("data"),
		Signature: []byte("sign"),
	}

	tx2 := tx1
	tx2.Nonce = 2

	input := make([]data.TransactionHandler, 0, 5)
	input = append(input, &tx1)
	input = append(input, &tx2)

	assert.Equal(t, 2, len(input))
	assert.Equal(t, 5, cap(input))

	input = transaction.TrimSliceHandler(input)

	assert.Equal(t, 2, len(input))
	assert.Equal(t, 2, cap(input))
}
