package vmGet

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"sync"
	"testing"

	vmData "github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVmGetShouldReturnValue(t *testing.T) {
	accnts, destinationAddressBytes, expectedValueForVar := deploySmartContract(t)

	mockVM := vm.CreateOneSCExecutorMockVM(accnts)
	vmContainer := &mock.VMContainerMock{
		GetCalled: func(key []byte) (handler vmcommon.VMExecutionHandler, e error) {
			return mockVM, nil
		}}
	argsNewSCQueryService := smartContract.ArgsNewSCQueryService{
		VmContainer: vmContainer,
		EconomicsFee: &mock.FeeHandlerStub{
			MaxGasLimitPerBlockCalled: func() uint64 {
				return uint64(math.MaxUint64)
			},
		},
		BlockChainHook:    &mock.BlockChainHookHandlerMock{},
		BlockChain:        &mock.BlockChainMock{},
		ArwenChangeLocker: &sync.RWMutex{},
	}
	service, _ := smartContract.NewSCQueryService(argsNewSCQueryService)

	functionName := "Get"
	query := process.SCQuery{
		ScAddress: destinationAddressBytes,
		FuncName:  functionName,
		Arguments: [][]byte{},
	}

	vmOutput, err := service.ExecuteQuery(&query)
	assert.Nil(t, err)

	returnData, _ := vmOutput.GetFirstReturnData(vmData.AsBigInt)
	assert.Equal(t, expectedValueForVar, returnData)
}

func deploySmartContract(t *testing.T) (state.AccountsAdapter, []byte, *big.Int) {
	vmOpGas := uint64(1)
	senderAddressBytes := []byte("12345678901234567890123456789012")
	senderNonce := uint64(11)
	senderBalance := big.NewInt(100000000)
	gasPrice := uint64(1)
	gasLimit := vmOpGas + 1000
	transferOnCalls := big.NewInt(0)

	initialValueForInternalVariable := uint64(45)
	scCode := fmt.Sprintf("aaaa@%s@0000@%X", hex.EncodeToString(factory.InternalTestingVM), initialValueForInternalVariable)

	tx := vm.CreateTx(
		senderAddressBytes,
		vm.CreateEmptyAddress(),
		senderNonce,
		transferOnCalls,
		gasPrice,
		gasLimit,
		scCode,
	)

	txProc, accnts, err := vm.CreatePreparedTxProcessorAndAccountsWithMockedVM(
		vmOpGas,
		senderNonce,
		senderAddressBytes,
		senderBalance,
		vm.ArgEnableEpoch{},
		&sync.RWMutex{},
	)
	require.Nil(t, err)

	_, err = txProc.ProcessTransaction(tx)
	assert.Nil(t, err)

	_, err = accnts.Commit()
	assert.Nil(t, err)

	destinationAddressBytes, _ := hex.DecodeString("0000000000000000ffff1a2983b179a480a60c4308da48f13b4480dbb4d33132")
	return accnts, destinationAddressBytes, big.NewInt(0).SetUint64(initialValueForInternalVariable)
}
