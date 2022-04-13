package arwenvm

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/arwen-wasm-vm/v1_4/arwen"
	contextmock "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/mock/context"
	worldmock "github.com/ElrondNetwork/arwen-wasm-vm/v1_4/mock/world"
	"github.com/ElrondNetwork/arwen-wasm-vm/v1_4/testcommon"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/stretchr/testify/require"
)

var MockInitialBalance = big.NewInt(10_000_000)

func InitializeMockContracts(
	t *testing.T,
	net *integrationTests.TestNetwork,
	mockSCs ...testcommon.MockTestSmartContract,
) {
	shardToHost, shardToInstanceBuilder :=
		CreateHostAndInstanceBuilder(t, net, factory.ArwenVirtualMachine)
	for _, mockSC := range mockSCs {
		shardID := mockSC.GetShardID()
		mockSC.Initialize(t, shardToHost[shardID], shardToInstanceBuilder[shardID], true)
	}
}

func GetAddressForNewAccount(
	t *testing.T,
	net *integrationTests.TestNetwork,
	shardID uint32) []byte {
	address := net.NewAddress(net.Wallets[shardID])
	account, _ := state.NewUserAccount(address)
	account.Balance = MockInitialBalance
	account.SetCode(address)
	account.SetCodeHash(address)
	err := net.NodesSharded[shardID][0].AccntState.SaveAccount(account)
	require.Nil(t, err)
	return address
}

func CreateHostAndInstanceBuilder(t *testing.T, net *integrationTests.TestNetwork, vmKey []byte) (map[uint32]arwen.VMHost, map[uint32]*contextmock.InstanceBuilderMock) {
	numberOfShards := uint32(net.NumShards)
	shardToWorld := make(map[uint32]*worldmock.MockWorld, numberOfShards)
	shardToInstanceBuilder := make(map[uint32]*contextmock.InstanceBuilderMock, numberOfShards)
	shardToHost := make(map[uint32]arwen.VMHost, numberOfShards)

	for shardID := uint32(0); shardID < numberOfShards; shardID++ {
		world := worldmock.NewMockWorld()
		world.SetProvidedBlockchainHook(net.DefaultNode.BlockchainHook)
		world.SelfShardID = shardID
		shardToWorld[shardID] = world
		instanceBuilderMock := contextmock.NewInstanceBuilderMock(world)
		shardToInstanceBuilder[shardID] = instanceBuilderMock
	}

	for shardID := uint32(0); shardID < numberOfShards; shardID++ {
		node := net.NodesSharded[shardID][0]
		host := testcommon.DefaultTestArwen(t, shardToWorld[shardID])
		host.Runtime().ReplaceInstanceBuilder(shardToInstanceBuilder[shardID])
		err := node.VMContainer.Replace(vmKey, host)
		require.Nil(t, err)
		shardToHost[shardID] = host
	}

	return shardToHost, shardToInstanceBuilder
}
