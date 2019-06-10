package vm

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/stretchr/testify/assert"
)

func TestVmDeployWithoutTransferShouldDeploySCCode(t *testing.T) {
	testVMDeploy(t, 0, big.NewInt(0))
}

func TestVmDeployWithTransferShouldDeploySCCode(t *testing.T) {
	testVMDeploy(t, 0, big.NewInt(50))
}

func TestVmDeployWithTransferAndGasShouldDeploySCCode(t *testing.T) {
	testVMDeploy(t, 1000, big.NewInt(50))
}

func TestVMDeployWithTransferWithInsufficientGasShouldReturnErr(t *testing.T) {
	accnts := createInMemoryShardAccountsDB()

	transferOnCalls := big.NewInt(50)

	senderPubkeyBytes := createDummyAddress().Bytes()
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	_ = createAccount(accnts, senderPubkeyBytes, senderNonce, senderBalance)

	vmOpGas := uint64(10000)

	txProcessor := createTxProcessorWithOneSCExecutorMockVM(accnts, vmOpGas)
	assert.NotNil(t, txProcessor)

	gasPrice := uint64(1)

	initialValueForInternalVariable := uint64(45)
	scCode := "mocked code, not taken into account"
	txData := fmt.Sprintf("%s@%X", scCode, initialValueForInternalVariable)
	tx := &transaction.Transaction{
		Nonce:    senderNonce,
		Value:    transferOnCalls,
		SndAddr:  senderPubkeyBytes,
		RcvAddr:  createEmptyAddress().Bytes(),
		Data:     []byte(txData),
		GasPrice: gasPrice,
		//slightly lower than required
		GasLimit: vmOpGas - 1,
	}
	assert.NotNil(t, tx)

	round := uint32(444)

	_, err := txProcessor.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	//we should now have the 2 accounts in the trie. Should get them and test all values
	senderAddress, _ := addrConv.CreateAddressFromPublicKeyBytes(senderPubkeyBytes)
	senderRecovAccount, _ := accnts.GetExistingAccount(senderAddress)
	senderRecovShardAccount := senderRecovAccount.(*state.Account)

	gasFunds := big.NewInt(0).Mul(big.NewInt(0).SetUint64(vmOpGas-1), big.NewInt(0).SetUint64(gasPrice))
	expectedSenderBalance := big.NewInt(0).Sub(senderBalance, gasFunds)

	assert.Equal(t, senderNonce+1, senderRecovShardAccount.GetNonce())
	assert.Equal(t, expectedSenderBalance, senderRecovShardAccount.Balance)

	destinationAddressBytes := computeSCDestinationAddressBytes(senderNonce, senderPubkeyBytes)
	destinationAddress, _ := addrConv.CreateAddressFromPublicKeyBytes(destinationAddressBytes)
	destinationAccount, _ := accnts.GetExistingAccount(destinationAddress)
	assert.Nil(t, destinationAccount)
}

func testVMDeploy(t *testing.T, opGas uint64, transferOnCalls *big.Int) {
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
		Value:    transferOnCalls,
		SndAddr:  senderPubkeyBytes,
		RcvAddr:  createEmptyAddress().Bytes(),
		Data:     []byte(txData),
		GasPrice: gasPrice,
		GasLimit: opGas,
	}
	assert.NotNil(t, tx)

	round := uint32(444)

	_, err := txProcessor.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	//we should now have the 2 accounts in the trie. Should get them and test all values
	senderAddress, _ := addrConv.CreateAddressFromPublicKeyBytes(senderPubkeyBytes)
	senderRecovAccount, _ := accnts.GetExistingAccount(senderAddress)
	senderRecovShardAccount := senderRecovAccount.(*state.Account)

	expectedSenderBalance := big.NewInt(0).Sub(senderBalance, transferOnCalls)
	gasFunds := big.NewInt(0).Mul(big.NewInt(0).SetUint64(opGas), big.NewInt(0).SetUint64(gasPrice))
	expectedSenderBalance.Sub(expectedSenderBalance, gasFunds)

	assert.Equal(t, senderNonce+1, senderRecovShardAccount.GetNonce())
	assert.Equal(t, expectedSenderBalance, senderRecovShardAccount.Balance)

	destinationAddressBytes := computeSCDestinationAddressBytes(senderNonce, senderPubkeyBytes)

	testDeployedContractContents(
		t,
		destinationAddressBytes,
		accnts,
		transferOnCalls,
		scCode,
		map[string]*big.Int{"a": big.NewInt(0).SetUint64(initialValueForInternalVariable)})
}
