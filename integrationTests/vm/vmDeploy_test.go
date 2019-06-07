package vm

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/stretchr/testify/assert"
)

//TODO add integration and unit tests with generating and broadcasting transaction with empty recv address

func TestVmDeployWithoutTransferShouldDeploySCCode(t *testing.T) {
	accnts := createInMemoryShardAccountsDB()

	senderPubkeyBytes := createDummyAddress().Bytes()
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	_ = createAccount(accnts, senderPubkeyBytes, senderNonce, senderBalance)

	txProcessor := createTxProcessorWithOneSCExecutorMockVM(accnts)
	assert.NotNil(t, txProcessor)

	tx := &transaction.Transaction{
		Nonce:   senderNonce,
		Value:   big.NewInt(0),
		SndAddr: senderPubkeyBytes,
		RcvAddr: createEmptyAddress().Bytes(),
	}
	assert.NotNil(t, tx)

	//round := uint32(444)

	//err := txProcessor.ProcessTransaction(tx, round)
	//assert.Nil(t, err)
}
