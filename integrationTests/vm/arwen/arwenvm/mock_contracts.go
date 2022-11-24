package arwenvm

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	"github.com/ElrondNetwork/wasm-vm/arwen"
	mock "github.com/ElrondNetwork/wasm-vm/mock/context"
	worldmock "github.com/ElrondNetwork/wasm-vm/mock/world"
	"github.com/ElrondNetwork/wasm-vm/testcommon"
	"github.com/stretchr/testify/require"
)

var MockInitialBalance = big.NewInt(10_000_000)

func InitializeMockContracts(
	t *testing.T,
	net *integrationTests.TestNetwork,
	mockSCs ...testcommon.MockTestSmartContract,
) {
	InitializeMockContractsWithVMContainer(t, net, nil, mockSCs...)
}

func InitializeMockContractsWithVMContainer(
	t *testing.T,
	net *integrationTests.TestNetwork,
	vmContainer process.VirtualMachinesContainer,
	mockSCs ...testcommon.MockTestSmartContract,
) {
	shardToHost, shardToInstanceBuilder :=
		CreateHostAndInstanceBuilder(t, net, vmContainer, factory.ArwenVirtualMachine)
	for _, mockSC := range mockSCs {
		shardID := mockSC.GetShardID()
		mockSC.Initialize(t, shardToHost[shardID], shardToInstanceBuilder[shardID], true)
	}
}

func GetAddressForNewAccountOnWalletAndNode(
	t *testing.T,
	net *integrationTests.TestNetwork,
	wallet *integrationTests.TestWalletAccount,
	node *integrationTests.TestProcessorNode,
) ([]byte, state.UserAccountHandler) {
	return GetAddressForNewAccountOnWalletAndNodeWithVM(t, net, wallet, node, net.DefaultVM)
}

func GetAddressForNewAccountOnWalletAndNodeWithVM(
	t *testing.T,
	net *integrationTests.TestNetwork,
	wallet *integrationTests.TestWalletAccount,
	node *integrationTests.TestProcessorNode,
	vmType []byte,
) ([]byte, state.UserAccountHandler) {
	walletAccount, err := node.AccntState.GetExistingAccount(wallet.Address)
	require.Nil(t, err)
	walletAccount.IncreaseNonce(1)
	wallet.Nonce++
	err = node.AccntState.SaveAccount(walletAccount)
	require.Nil(t, err)

	address := net.NewAddressWithVM(wallet, vmType)
	account, _ := state.NewUserAccount(address)
	account.Balance = MockInitialBalance
	account.SetCode(address)
	account.SetCodeHash(address)
	err = node.AccntState.SaveAccount(account)
	require.Nil(t, err)

	return address, account
}

func SetCodeMetadata(
	t *testing.T,
	codeMetadata []byte,
	node *integrationTests.TestProcessorNode,
	account state.UserAccountHandler,
) {
	account.SetCodeMetadata(codeMetadata)
	err := node.AccntState.SaveAccount(account)
	require.Nil(t, err)
}

func GetAddressForNewAccount(
	t *testing.T,
	net *integrationTests.TestNetwork,
	node *integrationTests.TestProcessorNode) ([]byte, state.UserAccountHandler) {
	return GetAddressForNewAccountWithVM(t, net, node, net.DefaultVM)
}

func GetAddressForNewAccountWithVM(
	t *testing.T,
	net *integrationTests.TestNetwork,
	node *integrationTests.TestProcessorNode,
	vmType []byte) ([]byte, state.UserAccountHandler) {
	return GetAddressForNewAccountOnWalletAndNodeWithVM(t, net, net.Wallets[node.ShardCoordinator.SelfId()], node, vmType)
}

func CreateHostAndInstanceBuilder(t *testing.T,
	net *integrationTests.TestNetwork,
	vmContainer process.VirtualMachinesContainer,
	vmKey []byte) (map[uint32]arwen.VMHost, map[uint32]*mock.ExecutorMock) {
	numberOfShards := uint32(net.NumShards)
	shardToWorld := make(map[uint32]*worldmock.MockWorld, numberOfShards)
	shardToExecutor := make(map[uint32]*mock.ExecutorMock, numberOfShards)
	shardToHost := make(map[uint32]arwen.VMHost, numberOfShards)

	net.DefaultNode.BlockchainHook.SetVMContainer(vmContainer)

	for shardID := uint32(0); shardID < numberOfShards; shardID++ {
		world := worldmock.NewMockWorld()
		world.SetProvidedBlockchainHook(net.DefaultNode.BlockchainHook)
		world.SelfShardID = shardID
		shardToWorld[shardID] = world
		mockExecutor := mock.NewExecutorMock(world)
		shardToExecutor[shardID] = mockExecutor
	}

	for shardID := uint32(0); shardID < numberOfShards; shardID++ {
		node := net.NodesSharded[shardID][0]
		for _, vmType := range vmContainer.Keys() {
			host, err := node.VMContainer.Get(vmType)
			require.NotNil(t, host)
			require.Nil(t, err)
			if _, ok := host.(arwen.VMHost); !ok {
				continue
			}
			host.(arwen.VMHost).Runtime().ReplaceVMExecutor(shardToExecutor[shardID])
			err = node.VMContainer.Replace(vmKey, host)
			require.Nil(t, err)
			shardToHost[shardID] = host.(arwen.VMHost)
		}
	}

	return shardToHost, shardToExecutor
}

// RegisterAsyncCallForMockContract is resued also in some tests before async context serialization
func RegisterAsyncCallForMockContract(host arwen.VMHost, config interface{}, destinationAddress []byte, egldValue []byte, callData *txDataBuilder.TxDataBuilder) error {
	testConfig := config.(*testcommon.TestConfig)

	async := host.Async()
	if !testConfig.IsLegacyAsync {
		err := async.RegisterAsyncCall("testGroup", &arwen.AsyncCall{
			Status:          arwen.AsyncCallPending,
			Destination:     destinationAddress,
			Data:            callData.ToBytes(),
			ValueBytes:      egldValue,
			SuccessCallback: testConfig.SuccessCallback,
			ErrorCallback:   testConfig.ErrorCallback,
			GasLimit:        testConfig.GasProvidedToChild,
			GasLocked:       testConfig.GasToLock,
			CallbackClosure: nil,
		})
		if err != nil {
			return err
		}
		return nil
	} else {
		return async.RegisterLegacyAsyncCall(destinationAddress, callData.ToBytes(), egldValue)
	}
}
