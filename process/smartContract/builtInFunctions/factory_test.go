package builtInFunctions

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
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
		Accounts:             &mock.AccountsStub{},
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
	assert.Equal(t, len(container.Keys()), 15)
}
