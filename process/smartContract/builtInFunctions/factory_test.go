package builtInFunctions

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	vmcommonBuiltInFunctions "github.com/ElrondNetwork/elrond-vm-common/builtInFunctions"
	"github.com/stretchr/testify/assert"
)

func createMockArguments() ArgsCreateBuiltInFunctionContainer {
	gasMap := make(map[string]map[string]uint64)
	fillGasMapInternal(gasMap, 1)

	gasScheduleNotifier := mock.NewGasScheduleNotifierMock(gasMap)
	args := ArgsCreateBuiltInFunctionContainer{
		GasSchedule:          gasScheduleNotifier,
		MapDNSAddresses:      make(map[string]struct{}),
		EnableUserNameChange: false,
		Marshalizer:          &mock.MarshalizerMock{},
		Accounts:             &testscommon.AccountsStub{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(1),
		EpochNotifier:        &mock.EpochNotifierStub{},
	}

	return args
}

func fillGasMapInternal(gasMap map[string]map[string]uint64, value uint64) map[string]map[string]uint64 {
	gasMap[common.BaseOperationCost] = fillGasMapBaseOperationCosts(value)
	gasMap[common.BuiltInCost] = fillGasMapBuiltInCosts(value)

	return gasMap
}

func fillGasMapBaseOperationCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["StorePerByte"] = value
	gasMap["DataCopyPerByte"] = value
	gasMap["ReleasePerByte"] = value
	gasMap["PersistPerByte"] = value
	gasMap["CompilePerByte"] = value
	gasMap["AoTPreparePerByte"] = value
	gasMap["GetCode"] = value
	return gasMap
}

func fillGasMapBuiltInCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["ClaimDeveloperRewards"] = value
	gasMap["ChangeOwnerAddress"] = value
	gasMap["SaveUserName"] = value
	gasMap["SaveKeyValue"] = value
	gasMap["ESDTTransfer"] = value
	gasMap["ESDTBurn"] = value
	gasMap["ChangeOwnerAddress"] = value
	gasMap["ClaimDeveloperRewards"] = value
	gasMap["SaveUserName"] = value
	gasMap["SaveKeyValue"] = value
	gasMap["ESDTTransfer"] = value
	gasMap["ESDTBurn"] = value
	gasMap["ESDTLocalMint"] = value
	gasMap["ESDTLocalBurn"] = value
	gasMap["ESDTNFTCreate"] = value
	gasMap["ESDTNFTAddQuantity"] = value
	gasMap["ESDTNFTBurn"] = value
	gasMap["ESDTNFTTransfer"] = value
	gasMap["ESDTNFTChangeCreateOwner"] = value
	gasMap["ESDTNFTAddUri"] = value
	gasMap["ESDTNFTUpdateAttributes"] = value
	gasMap["ESDTNFTMultiTransfer"] = value

	return gasMap
}

func TestCreateBuiltInFunctionContainer_Errors(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	args.GasSchedule = nil
	container, err := CreateBuiltInFunctionContainer(args)
	assert.NotNil(t, err)
	assert.Nil(t, container)

	args = createMockArguments()
	args.MapDNSAddresses = nil
	container, err = CreateBuiltInFunctionContainer(args)
	assert.Equal(t, process.ErrNilDnsAddresses, err)
	assert.Nil(t, container)

	args = createMockArguments()
	container, err = CreateBuiltInFunctionContainer(args)
	assert.Nil(t, err)
	assert.Equal(t, len(container.Keys()), 25)

	err = vmcommonBuiltInFunctions.SetPayableHandler(container, &mock.BlockChainHookHandlerMock{})
	assert.Nil(t, err)

	assert.False(t, container.IsInterfaceNil())
}
