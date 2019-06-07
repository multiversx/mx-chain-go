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

func deployFreeAndRunSmartContract(t *testing.T, runSCvalue *big.Int) {
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

	destinationAddressBytes := computeSCDestinationAddressBytes(senderNonce, senderPubkeyBytes)
	addValue := uint64(128)
	txData = fmt.Sprintf("Add@%X", addValue)

	txRun := &transaction.Transaction{
		Nonce:   senderNonce + 1,
		Value:   runSCvalue,
		SndAddr: senderPubkeyBytes,
		RcvAddr: destinationAddressBytes,
		Data:    []byte(txData),
	}

	err = txProcessor.ProcessTransaction(txRun, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	//we should now have the 2 accounts in the trie. Should get them and test all values
	senderAddress, _ := addrConv.CreateAddressFromPublicKeyBytes(senderPubkeyBytes)
	senderRecovAccount, _ := accnts.GetExistingAccount(senderAddress)
	senderRecovShardAccount := senderRecovAccount.(*state.Account)

	assert.Equal(t, senderNonce+2, senderRecovShardAccount.GetNonce())
	assert.Equal(t, senderBalance.Uint64()-runSCvalue.Uint64(), senderRecovShardAccount.Balance.Uint64())

	testDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		runSCvalue,
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable + addValue)})
}

func TestVmRUNSCCodeAddWithoutBalance(t *testing.T) {
	deployFreeAndRunSmartContract(t, big.NewInt(0))
}

func TestVmRUNSCCodeAddWithBalance(t *testing.T) {
	deployFreeAndRunSmartContract(t, big.NewInt(45))
}
