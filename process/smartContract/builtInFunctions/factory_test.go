package builtInFunctions

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		Accounts:             &mock.AccountsStub{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(1),
	}

	return args
}

func fillGasMapInternal(gasMap map[string]map[string]uint64, value uint64) map[string]map[string]uint64 {
	gasMap[core.BaseOperationCost] = fillGasMapBaseOperationCosts(value)
	gasMap[core.BuiltInCost] = fillGasMapBuiltInCosts(value)

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

	return gasMap
}

func TestCreateBuiltInFunctionContainer_Errors(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	args.GasSchedule = nil
	factory, err := NewBuiltInFunctionsFactory(args)
	assert.NotNil(t, err)
	assert.Nil(t, factory)

	args = createMockArguments()
	args.MapDNSAddresses = nil
	factory, err = NewBuiltInFunctionsFactory(args)
	assert.Equal(t, process.ErrNilDnsAddresses, err)
	assert.Nil(t, factory)

	args = createMockArguments()
	factory, err = NewBuiltInFunctionsFactory(args)
	assert.Nil(t, err)
	container, err := factory.CreateBuiltInFunctionContainer()
	assert.Nil(t, err)
	assert.Equal(t, len(container.Keys()), 20)

	err = SetPayableHandler(container, &mock.BlockChainHookHandlerMock{})
	assert.Nil(t, err)

	assert.False(t, factory.IsInterfaceNil())
}

func TestBuiltInFuncFactory_GasScheduleChange(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	factory, _ := NewBuiltInFunctionsFactory(args)

	baseOpCosts := map[string]uint64{
		"StorePerByte":      100,
		"ReleasePerByte":    100,
		"DataCopyPerByte":   100,
		"PersistPerByte":    100,
		"CompilePerByte":    100,
		"AoTPreparePerByte": 100,
	}

	builtInCosts := map[string]uint64{
		"ChangeOwnerAddress":       100,
		"ClaimDeveloperRewards":    100,
		"SaveUserName":             100,
		"SaveKeyValue":             100,
		"ESDTTransfer":             100,
		"ESDTBurn":                 100,
		"ESDTLocalMint":            100,
		"ESDTLocalBurn":            100,
		"ESDTNFTCreate":            100,
		"ESDTNFTAddQuantity":       100,
		"ESDTNFTBurn":              100,
		"ESDTNFTTransfer":          100,
		"ESDTNFTChangeCreateOwner": 100,
	}
	gasMap := map[string]map[string]uint64{
		core.BaseOperationCost: baseOpCosts,
		core.BuiltInCost:       builtInCosts,
	}
	factory.GasScheduleChange(gasMap)

	require.Equal(t, builtInCosts["ChangeOwnerAddress"], factory.gasConfig.BuiltInCost.ChangeOwnerAddress)
}
