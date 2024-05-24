package chainSimulator

import (
	"math/big"

	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
)

func GenerateTransaction(sender []byte, nonce uint64, receiver []byte, value *big.Int, data string, gasLimit uint64) *transaction.Transaction {
	minGasPrice := uint64(1000000000)
	txVersion := uint32(1)
	mockTxSignature := "sig"

	return &transaction.Transaction{
		Nonce:     nonce,
		Value:     value,
		SndAddr:   sender,
		RcvAddr:   receiver,
		Data:      []byte(data),
		GasLimit:  gasLimit,
		GasPrice:  minGasPrice,
		ChainID:   []byte(configs.ChainID),
		Version:   txVersion,
		Signature: []byte(mockTxSignature),
	}
}
