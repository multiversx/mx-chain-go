package mockVM

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/stretchr/testify/assert"
)

func TestVmGetShouldReturnValue(t *testing.T) {
	accnts, destinationAddressBytes, expectedValueForVar := deploySmartContract(t)

	mockVM := vm.CreateOneSCExecutorMockVM(accnts)
	scgd, _ := smartContract.NewSCDataGetter(mockVM)

	functionName := "Get"
	returnedVals, err := scgd.Get(destinationAddressBytes, functionName)

	assert.Nil(t, err)
	assert.Equal(t, expectedValueForVar.Bytes(), returnedVals)
}

func deploySmartContract(t *testing.T) (state.AccountsAdapter, []byte, *big.Int) {
	vmOpGas := uint64(0)
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	round := uint64(444)
	gasPrice := uint64(1)
	gasLimit := vmOpGas
	transferOnCalls := big.NewInt(0)

	initialValueForInternalVariable := uint64(45)
	scCode := fmt.Sprintf("aaaa@01@%X", initialValueForInternalVariable)

	tx := vm.CreateTx(
		t,
		senderAddressBytes,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
	)

	txProc, accnts := vm.CreatePreparedTxProcessorAndAccountsWithMockedVM(t, vmOpGas, senderNonce, senderAddressBytes, senderBalance)

	err := txProc.ProcessTransaction(tx, round)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	destinationAddressBytes, _ := hex.DecodeString("00000000000000003031597e658cd20edf45a1004da5bf04f8b6518744847c32")
	return accnts, destinationAddressBytes, big.NewInt(0).SetUint64(initialValueForInternalVariable)
}
