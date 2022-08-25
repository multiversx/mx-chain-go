package builtInFunctions

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
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
		GasSchedule:          gasScheduleNotifier,
		MapDNSAddresses:      make(map[string]struct{}),
		EnableUserNameChange: false,
		Marshalizer:          &mock.MarshalizerMock{},
		Accounts:             &stateMock.AccountsStub{},
		ShardCoordinator:     mock.NewMultiShardsCoordinatorMock(1),
		EpochNotifier:        &epochNotifier.EpochNotifierStub{},
		AutomaticCrawlerAddresses: [][]byte{
			bytes.Repeat([]byte{1}, 32),
		},
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

func TestCreateBuiltInFunctionContainer_Errors(t *testing.T) {
	t.Parallel()

	args := createMockArguments()
	args.GasSchedule = nil
	builtInFuncFactory, err := CreateBuiltInFunctionsFactory(args)
	assert.NotNil(t, err)
	assert.Nil(t, builtInFuncFactory)

	args = createMockArguments()
	args.MapDNSAddresses = nil
	builtInFuncFactory, err = CreateBuiltInFunctionsFactory(args)
	assert.Equal(t, process.ErrNilDnsAddresses, err)
	assert.Nil(t, builtInFuncFactory)

	args = createMockArguments()
	builtInFuncFactory, err = CreateBuiltInFunctionsFactory(args)
	assert.Nil(t, err)
	assert.Equal(t, len(builtInFuncFactory.BuiltInFunctionContainer().Keys()), 31)

	err = builtInFuncFactory.SetPayableHandler(&testscommon.BlockChainHookStub{})
	assert.Nil(t, err)

	assert.False(t, builtInFuncFactory.BuiltInFunctionContainer().IsInterfaceNil())
	assert.False(t, builtInFuncFactory.NFTStorageHandler().IsInterfaceNil())
	assert.False(t, builtInFuncFactory.ESDTGlobalSettingsHandler().IsInterfaceNil())
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
	t.Run("metachain takes first address", func(t *testing.T) {
		t.Parallel()

		shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
		shardCoordinator.CurrentShard = common.MetachainShardId
		addresses := [][]byte{
			bytes.Repeat([]byte{20}, 32), // bigger addresss
			bytes.Repeat([]byte{3}, 32)}  // smaller address

		allowedAddressForShard, err := GetAllowedAddress(shardCoordinator, addresses)
		assert.Nil(t, err)
		assert.Equal(t, addresses[1], allowedAddressForShard)
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
		assert.Equal(t, addresses[0], allowedAddressForShard)
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
