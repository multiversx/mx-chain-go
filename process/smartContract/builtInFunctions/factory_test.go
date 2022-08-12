package builtInFunctions

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	"github.com/stretchr/testify/assert"
)

func createMockArguments() ArgsCreateBuiltInFunctionContainer {
	gasMap := make(map[string]map[string]uint64)
	fillGasMapInternal(gasMap, 1)

	gasScheduleNotifier := testscommon.NewGasScheduleNotifierMock(gasMap)
	args := ArgsCreateBuiltInFunctionContainer{
		GasSchedule:               gasScheduleNotifier,
		MapDNSAddresses:           make(map[string]struct{}),
		EnableUserNameChange:      false,
		Marshalizer:               &mock.MarshalizerMock{},
		Accounts:                  &stateMock.AccountsStub{},
		ShardCoordinator:          mock.NewMultiShardsCoordinatorMock(1),
		EpochNotifier:             &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandler:       &testscommon.EnableEpochsHandlerStub{},
		AutomaticCrawlerAddress:   bytes.Repeat([]byte{1}, 32),
		MaxNumNodesInTransferRole: 100,
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

func TestCreateBuiltInFunctionContainer(t *testing.T) {
	t.Parallel()

	t.Run("nil gas schedule should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.GasSchedule = nil
		builtInFuncFactory, err := CreateBuiltInFunctionsFactory(args)
		assert.Equal(t, process.ErrNilGasSchedule, err)
		assert.Nil(t, builtInFuncFactory)
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.Marshalizer = nil
		builtInFuncFactory, err := CreateBuiltInFunctionsFactory(args)
		assert.Equal(t, process.ErrNilMarshalizer, err)
		assert.Nil(t, builtInFuncFactory)
	})
	t.Run("nil accounts should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.Accounts = nil
		builtInFuncFactory, err := CreateBuiltInFunctionsFactory(args)
		assert.Equal(t, process.ErrNilAccountsAdapter, err)
		assert.Nil(t, builtInFuncFactory)
	})
	t.Run("nil map dns addresses should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.MapDNSAddresses = nil
		builtInFuncFactory, err := CreateBuiltInFunctionsFactory(args)
		assert.Equal(t, process.ErrNilDnsAddresses, err)
		assert.Nil(t, builtInFuncFactory)
	})
	t.Run("nil shard coordinator should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.ShardCoordinator = nil
		builtInFuncFactory, err := CreateBuiltInFunctionsFactory(args)
		assert.Equal(t, process.ErrNilShardCoordinator, err)
		assert.Nil(t, builtInFuncFactory)
	})
	t.Run("nil epoch notifier should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.EpochNotifier = nil
		builtInFuncFactory, err := CreateBuiltInFunctionsFactory(args)
		assert.Equal(t, process.ErrNilEpochNotifier, err)
		assert.Nil(t, builtInFuncFactory)
	})
	t.Run("nil epochs handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		args.EnableEpochsHandler = nil
		builtInFuncFactory, err := CreateBuiltInFunctionsFactory(args)
		assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
		assert.Nil(t, builtInFuncFactory)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArguments()
		builtInFuncFactory, err := CreateBuiltInFunctionsFactory(args)
		assert.Nil(t, err)
		assert.Equal(t, len(builtInFuncFactory.BuiltInFunctionContainer().Keys()), 31)

		err = builtInFuncFactory.SetPayableHandler(&testscommon.BlockChainHookStub{})
		assert.Nil(t, err)

		assert.False(t, builtInFuncFactory.BuiltInFunctionContainer().IsInterfaceNil())
		assert.False(t, builtInFuncFactory.NFTStorageHandler().IsInterfaceNil())
		assert.False(t, builtInFuncFactory.ESDTGlobalSettingsHandler().IsInterfaceNil())
	})
}
