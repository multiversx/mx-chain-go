package mockVM

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/stretchr/testify/assert"
)

//TODO add integration and unit tests with generating and broadcasting transaction with empty recv address

func deployAndRunSmartContract(t *testing.T, opGas uint64, txvalue *big.Int) {
	accnts := createInMemoryShardAccountsDB()

	senderPubkeyBytes := createDummyAddress().Bytes()
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	_ = createAccount(accnts, senderPubkeyBytes, senderNonce, senderBalance)

	txProcessor := createTxProcessorWithOneSCExecutorMockVM(accnts, opGas)
	assert.NotNil(t, txProcessor)

	gasPrice := uint64(1)
	initialValueForInternalVariable := uint64(45)
	scCode := "mocked code, not taken into account"
	txData := fmt.Sprintf("%s@%X", scCode, initialValueForInternalVariable)
	tx := &transaction.Transaction{
		Nonce:    senderNonce,
		Value:    big.NewInt(0),
		SndAddr:  senderPubkeyBytes,
		RcvAddr:  createEmptyAddress().Bytes(),
		Data:     []byte(txData),
		GasPrice: gasPrice,
		GasLimit: opGas,
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
		Nonce:    senderNonce + 1,
		Value:    txvalue,
		SndAddr:  senderPubkeyBytes,
		RcvAddr:  destinationAddressBytes,
		Data:     []byte(txData),
		GasLimit: opGas,
		GasPrice: gasPrice,
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

	expectedSenderBalance := big.NewInt(0).Sub(senderBalance, txvalue)
	gasFunds := big.NewInt(0).Mul(big.NewInt(0).SetUint64(opGas), big.NewInt(0).SetUint64(gasPrice))
	expectedSenderBalance.Sub(expectedSenderBalance, gasFunds)
	expectedSenderBalance.Sub(expectedSenderBalance, gasFunds)
	testDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		txvalue,
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable + addValue)})
}

func TestRunSCWithoutTransferShouldRunSCCode(t *testing.T) {
	deployAndRunSmartContract(t, 0, big.NewInt(0))
}

func TestRunSCWithTransferShouldRunSCCode(t *testing.T) {
	deployAndRunSmartContract(t, 0, big.NewInt(50))
}

func TestRunWithTransferAndGasShouldRunSCCode(t *testing.T) {
	deployAndRunSmartContract(t, 1000, big.NewInt(50))
}
