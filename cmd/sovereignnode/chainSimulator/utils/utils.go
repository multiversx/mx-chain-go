package utils

import (
	"encoding/hex"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

const (
	maxNumOfBlocksToGenerateWhenExecutingTx = 1
	signalError                             = "signalError"
)

func getSCCode(fileName string) string {
	code, err := os.ReadFile(filepath.Clean(fileName))
	if err != nil {
		panic("Could not get SC code.")
	}

	codeEncoded := hex.EncodeToString(code)
	return codeEncoded
}

// DeployContract -
func DeployContract(
	t *testing.T,
	cs chainSimulatorIntegrationTests.ChainSimulator,
	sender []byte,
	nonce *uint64,
	receiver []byte,
	data string,
	wasmPath string,
) []byte {
	data = getSCCode(wasmPath) + "@0500@0500" + data

	tx := GenerateTransaction(sender, *nonce, receiver, big.NewInt(0), data, uint64(200000000))
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	*nonce++

	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)

	address := txResult.Logs.Events[0].Topics[0]
	require.NotNil(t, address)
	return address
}

// GenerateTransaction -
func GenerateTransaction(
	sender []byte,
	nonce uint64,
	receiver []byte,
	value *big.Int,
	data string,
	gasLimit uint64,
) *transaction.Transaction {
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

// SendTransaction -
func SendTransaction(
	t *testing.T,
	cs chainSimulatorIntegrationTests.ChainSimulator,
	sender []byte,
	nonce *uint64,
	receiver []byte,
	value *big.Int,
	data string,
	gasLimit uint64,
) *transaction.ApiTransactionResult {
	tx := GenerateTransaction(sender, *nonce, receiver, value, data, gasLimit)
	txResult, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlocksToGenerateWhenExecutingTx)
	*nonce++
	require.Nil(t, err)
	require.NotNil(t, txResult)
	require.Equal(t, transaction.TxStatusSuccess, txResult.Status)
	if txResult.Logs != nil && txResult.Logs.Events != nil && len(txResult.Logs.Events) > 0 {
		require.NotEqual(t, signalError, txResult.Logs.Events[0].Identifier)
	}

	return txResult
}
