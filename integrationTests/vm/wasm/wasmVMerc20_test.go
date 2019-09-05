package wasm

import (
	"encoding/hex"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/hashing/keccak"
	"io/ioutil"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/stretchr/testify/assert"
)

var wrc20wasmFile = "./test20_elrond.wasm"

func TestVmDeployWithTransferAndGasShouldDeploySCCode(t *testing.T) {
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(0)
	senderBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(50)

	scCode, err := ioutil.ReadFile(wrc20wasmFile)
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := vm.CreateTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString,
	)

	config := "./libhera.so,engine=wavm"
	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithWASMVM(t, senderNonce, senderAddressBytes, senderBalance, config)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(99999139)
	vm.TestAccount(
		t,
		accnts,
		senderAddressBytes,
		senderNonce+1,
		expectedBalance)

	fmt.Printf("%s \n", hex.EncodeToString(keccak.Keccak{}.Compute("transfer")))
	fmt.Printf("%s \n", hex.EncodeToString(keccak.Keccak{}.Compute("balance")))
}

func TestVmDeployWithTransferAndExecute(t *testing.T) {
	ownerAddressBytes := []byte("12345678901234567890123456789012")
	ownerNonce := uint64(11)
	ownerBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := uint64(100000)
	transferOnCalls := big.NewInt(500000)

	scCode, err := ioutil.ReadFile(wrc20wasmFile)
	assert.Nil(t, err)

	scCodeString := hex.EncodeToString(scCode)

	tx := vm.CreateTx(
		t,
		ownerAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		ownerNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCodeString,
	)

	config := "./libhera.so,engine=wavm"
	txProc, accnts, _ := vm.CreatePreparedTxProcessorAndAccountsWithWASMVM(t, ownerNonce, ownerAddressBytes, ownerBalance, config)

	err = txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	expectedBalance := big.NewInt(99999139)
	vm.TestAccount(
		t,
		accnts,
		ownerAddressBytes,
		ownerNonce+1,
		expectedBalance)

	scAddress, _ := hex.DecodeString("0000000000000000000061356266303466386236353138373434383437633132")

	alice := []byte("12345678901234567890123456789ALI")
	aliceNonce := uint64(0)
	_ = vm.CreateAccount(accnts, alice, aliceNonce, big.NewInt(1000000))

	bob := []byte("12345678901234567890123456789BOB")
	_ = vm.CreateAccount(accnts, bob, 0, big.NewInt(1000000))

	transferCall := "transfer" + "@" + string(bob) + "@" + "0001"

	for i := 0; i < 10000; i++ {
		tx = vm.CreateTx(
			t,
			alice,
			scAddress,
			aliceNonce,
			transferOnCalls,
			gasPrice,
			gasLimit,
			transferCall,
		)

		err = txProc.ProcessTransaction(tx, round)
		assert.Nil(t, err)

		_, err = accnts.Commit()
		assert.Nil(t, err)

		aliceNonce++
	}

}
