package builtInFunctions

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/guardianMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/stretchr/testify/assert"
)

func createMockArguments() ArgsCreateBuiltInFunctionContainer {
	gasMap := make(map[string]map[string]uint64)
	fillGasMapInternal(gasMap, 1)

	gasScheduleNotifier := testscommon.NewGasScheduleNotifierMock(gasMap)
	args := ArgsCreateBuiltInFunctionContainer{
		GasSchedule:          gasScheduleNotifier,
		MapDNSAddresses:      make(map[string]struct{}),
		MapDNSV2Addresses:    make(map[string]struct{}),
		EnableUserNameChange: false,
		Marshalizer:          &mock.MarshalizerMock{},
		Accounts:             &stateMock.AccountsStub{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(1),
		EpochNotifier:        &epochNotifier.EpochNotifierStub{},
		EnableEpochsHandler:  &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		AutomaticCrawlerAddresses: [][]byte{
			bytes.Repeat([]byte{1}, 32),
		},
		MaxNumNodesInTransferRole: 100,
		GuardedAccountHandler:     &guardianMocks.GuardedAccountHandlerStub{},
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
	gasMap["ESDTModifyRoyalties"] = value
	gasMap["ESDTModifyCreator"] = value
	gasMap["ESDTNFTRecreate"] = value
	gasMap["ESDTNFTUpdate"] = value
	gasMap["ESDTNFTSetNewURIs"] = value
	gasMap["SetGuardian"] = value
	gasMap["GuardAccount"] = value
	gasMap["TrieLoadPerNode"] = value
	gasMap["TrieStorePerNode"] = value

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
		assert.Equal(t, 42, len(builtInFuncFactory.BuiltInFunctionContainer().Keys()))

		err = builtInFuncFactory.SetPayableHandler(&testscommon.BlockChainHookStub{})
		assert.Nil(t, err)

		assert.False(t, builtInFuncFactory.BuiltInFunctionContainer().IsInterfaceNil())
		assert.False(t, builtInFuncFactory.NFTStorageHandler().IsInterfaceNil())
		assert.False(t, builtInFuncFactory.ESDTGlobalSettingsHandler().IsInterfaceNil())
	})
}

func TestCreateBuiltInFunctionContainerGetAllowedAddress_Errors(t *testing.T) {
	t.Parallel()

	t.Run("nil shardCoordinator", func(t *testing.T) {
		t.Parallel()

		_, addresses := GetMockShardCoordinatorAndAddresses(1)
		allowedAddressForShard, err := GetAllowedAddress(nil, addresses)
		assert.Nil(t, allowedAddressForShard)
		assert.Equal(t, process.ErrNilShardCoordinator, err)
	})
	t.Run("nil addresses", func(t *testing.T) {
		t.Parallel()

		shardCoordinator, _ := GetMockShardCoordinatorAndAddresses(1)
		allowedAddressForShard, err := GetAllowedAddress(shardCoordinator, nil)
		assert.Nil(t, allowedAddressForShard)
		assert.True(t, errors.Is(err, process.ErrNilCrawlerAllowedAddress))
		assert.True(t, strings.Contains(err.Error(), "provided count is 0"))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		shardCoordinator, addresses := GetMockShardCoordinatorAndAddresses(1)
		allowedAddressForShard, err := GetAllowedAddress(shardCoordinator, addresses)
		assert.NotNil(t, allowedAddressForShard)
		assert.Nil(t, err)
	})
	t.Run("existing address for shard 1", func(t *testing.T) {
		t.Parallel()

		shardCoordinator, _ := GetMockShardCoordinatorAndAddresses(1)
		addresses := [][]byte{
			bytes.Repeat([]byte{1}, 32), // shard 1
		}

		allowedAddressForShard, err := GetAllowedAddress(shardCoordinator, addresses)
		assert.NotNil(t, allowedAddressForShard)
		assert.Nil(t, err)
	})
	t.Run("no address for shard 1", func(t *testing.T) {
		t.Parallel()

		shardCoordinator, _ := GetMockShardCoordinatorAndAddresses(1)
		addresses := [][]byte{
			bytes.Repeat([]byte{2}, 32), // shard 2
			bytes.Repeat([]byte{3}, 32)} // shard 0

		allowedAddressForShard, err := GetAllowedAddress(shardCoordinator, addresses)
		assert.Nil(t, allowedAddressForShard)
		assert.True(t, errors.Is(err, process.ErrNilCrawlerAllowedAddress))
		expectedMessage := fmt.Sprintf("for shard %d, provided count is %d", shardCoordinator.SelfId(), len(addresses))
		assert.True(t, strings.Contains(err.Error(), expectedMessage))
	})
	t.Run("metachain takes core.SystemAccountAddress", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		shardCoordinator.CurrentShard = common.MetachainShardId
		addresses := [][]byte{
			bytes.Repeat([]byte{20}, 32), // bigger addresss
			bytes.Repeat([]byte{3}, 32)}  // smaller address

		allowedAddressForShard, err := GetAllowedAddress(shardCoordinator, addresses)
		assert.Nil(t, err)
		assert.Equal(t, core.SystemAccountAddress, allowedAddressForShard)
	})
	t.Run("every shard gets an allowedCrawlerAddress", func(t *testing.T) {
		t.Parallel()

		nrShards := uint32(3)
		shardCoordinator := mock.NewMultiShardsCoordinatorMock(nrShards)
		addresses := make([][]byte, nrShards)
		for i := byte(0); i < byte(nrShards); i++ {
			addresses[i] = bytes.Repeat([]byte{i + byte(nrShards)}, 32)
		}
		shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
			lastByte := address[len(address)-1]
			return uint32(lastByte) % nrShards
		}

		currentShardId := uint32(0)
		shardCoordinator.CurrentShard = currentShardId
		allowedAddressForShard, _ := GetAllowedAddress(shardCoordinator, addresses)
		assert.Equal(t, addresses[currentShardId], allowedAddressForShard)

		currentShardId = uint32(1)
		shardCoordinator.CurrentShard = currentShardId
		allowedAddressForShard, _ = GetAllowedAddress(shardCoordinator, addresses)
		assert.Equal(t, addresses[currentShardId], allowedAddressForShard)

		currentShardId = uint32(2)
		shardCoordinator.CurrentShard = currentShardId
		allowedAddressForShard, _ = GetAllowedAddress(shardCoordinator, addresses)
		assert.Equal(t, addresses[currentShardId], allowedAddressForShard)

		currentShardId = common.MetachainShardId
		shardCoordinator.CurrentShard = currentShardId
		allowedAddressForShard, _ = GetAllowedAddress(shardCoordinator, addresses)
		assert.Equal(t, core.SystemAccountAddress, allowedAddressForShard)
	})

}

func GetMockShardCoordinatorAndAddresses(currentShardId uint32) (sharding.Coordinator, [][]byte) {
	nrShards := uint32(3)
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(nrShards)
	addresses := make([][]byte, nrShards)
	for i := byte(0); i < byte(nrShards); i++ {
		addresses[i] = bytes.Repeat([]byte{i + byte(nrShards)}, 32)
	}
	shardCoordinator.CurrentShard = currentShardId
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		lastByte := address[len(address)-1]
		return uint32(lastByte) % nrShards
	}

	return shardCoordinator, addresses
}
