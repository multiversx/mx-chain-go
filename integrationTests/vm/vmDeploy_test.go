package vm

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
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

	initialValueForInternalVariable := uint64(45)
	scCode := "mocked code, not taken into account"
	txData := fmt.Sprintf("%s@%X", scCode, initialValueForInternalVariable)
	tx := &transaction.Transaction{
		Nonce:   senderNonce,
		Value:   big.NewInt(0),
		SndAddr: senderPubkeyBytes,
		RcvAddr: createEmptyAddress().Bytes(),
		Data:    []byte(txData),
	}
	assert.NotNil(t, tx)

	round := uint32(444)

	err := txProcessor.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	//we should now have the 2 accounts in the trie. Should get them and test all values
	senderAddress, _ := addrConv.CreateAddressFromPublicKeyBytes(senderPubkeyBytes)
	senderRecovAccount, _ := accnts.GetExistingAccount(senderAddress)
	senderRecovShardAccount := senderRecovAccount.(*state.Account)

	assert.Equal(t, senderNonce+1, senderRecovShardAccount.GetNonce())
	assert.Equal(t, senderBalance, senderRecovShardAccount.Balance)

	destinationAddressBytes := computeSCDestinationAddressBytes(senderNonce, senderPubkeyBytes)

	testDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		big.NewInt(0),
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})
}
