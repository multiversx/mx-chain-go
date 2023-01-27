package arwenvm

import (
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
	"github.com/multiversx/mx-chain-vm-go/executor"
	contextmock "github.com/multiversx/mx-chain-vm-go/mock/context"
	worldmock "github.com/multiversx/mx-chain-vm-go/mock/world"
	"github.com/multiversx/mx-chain-vm-go/testcommon"
	"github.com/multiversx/mx-chain-vm-go/vmhost"
	"github.com/stretchr/testify/require"
)

var MockInitialBalance = big.NewInt(10_000_000)

// WalletAddressPrefix is the prefix of any smart contract address used for testing.
var WalletAddressPrefix = []byte("..........")

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

func GetAddressForNewAccountOnWalletAndNode(
	t *testing.T,
	net *integrationTests.TestNetwork,
	wallet *integrationTests.TestWalletAccount,
	node *integrationTests.TestProcessorNode,
) ([]byte, state.UserAccountHandler) {
	walletAccount, err := node.AccntState.GetExistingAccount(wallet.Address)
	require.Nil(t, err)
	walletAccount.IncreaseNonce(1)
	wallet.Nonce++
	err = node.AccntState.SaveAccount(walletAccount)
	require.Nil(t, err)

	address := net.NewAddress(wallet)
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
	return GetAddressForNewAccountOnWalletAndNode(t, net, net.Wallets[node.ShardCoordinator.SelfId()], node)
}

// MakeTestWalletAddress generates a new wallet address to be used for
// testing based on the given identifier.
func MakeTestWalletAddress(identifier string) []byte {
	return makeTestAddress(WalletAddressPrefix, identifier)
}

func makeTestAddress(prefix []byte, identifier string) []byte {
	numberOfTrailingDots := vmhost.AddressSize - len(vmhost.SCAddressPrefix) - len(identifier)
	leftBytes := vmhost.SCAddressPrefix
	rightBytes := []byte(identifier + strings.Repeat(".", numberOfTrailingDots))
	return append(leftBytes, rightBytes...)
}

func CreateHostAndInstanceBuilder(t *testing.T, net *integrationTests.TestNetwork, vmKey []byte) (map[uint32]vmhost.VMHost, map[uint32]*contextmock.ExecutorMock) {
	numberOfShards := uint32(net.NumShards)
	shardToWorld := make(map[uint32]*worldmock.MockWorld, numberOfShards)
	shardToInstanceBuilder := make(map[uint32]*contextmock.ExecutorMock, numberOfShards)
	shardToHost := make(map[uint32]vmhost.VMHost, numberOfShards)

	for shardID := uint32(0); shardID < numberOfShards; shardID++ {
		world := worldmock.NewMockWorld()
		world.SetProvidedBlockchainHook(net.DefaultNode.BlockchainHook)
		world.SelfShardID = shardID
		shardToWorld[shardID] = world
		instanceBuilderMock, _ := contextmock.NewExecutorMockFactory(world).CreateExecutor(executor.ExecutorFactoryArgs{})
		shardToInstanceBuilder[shardID] = instanceBuilderMock.(*contextmock.ExecutorMock)
	}

	for shardID := uint32(0); shardID < numberOfShards; shardID++ {
		node := net.NodesSharded[shardID][0]
		host, err := node.VMContainer.Get(factory.ArwenVirtualMachine)
		require.NotNil(t, host)
		require.Nil(t, err)
		host.(vmhost.VMHost).Runtime().ReplaceVMExecutor(shardToInstanceBuilder[shardID])
		err = node.VMContainer.Replace(vmKey, host)
		require.Nil(t, err)
		shardToHost[shardID] = host.(vmhost.VMHost)
	}

	return shardToHost, shardToInstanceBuilder
}

// RegisterAsyncCallForMockContract is resued also in some tests before async context serialization
func RegisterAsyncCallForMockContract(host vmhost.VMHost, config interface{}, destinationAddress []byte, egldValue []byte, callData *txDataBuilder.TxDataBuilder) error {
	testConfig := config.(*testcommon.TestConfig)

	async := host.Async()
	if !testConfig.IsLegacyAsync {
		err := async.RegisterAsyncCall("testGroup", &vmhost.AsyncCall{
			Status:          vmhost.AsyncCallPending,
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
