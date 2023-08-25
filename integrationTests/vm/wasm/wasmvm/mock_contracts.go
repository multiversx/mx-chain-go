package wasmvm

import (
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/factory"
	"github.com/multiversx/mx-chain-go/state"
	stateFactory "github.com/multiversx/mx-chain-go/state/factory"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
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

// InitializeMockContracts -
func InitializeMockContracts(
	t *testing.T,
	net *integrationTests.TestNetwork,
	mockSCs ...testcommon.MockTestSmartContract,
) {
	InitializeMockContractsWithVMContainer(t, net, nil, mockSCs...)
}

// InitializeMockContractsWithVMContainer -
func InitializeMockContractsWithVMContainer(
	t *testing.T,
	net *integrationTests.TestNetwork,
	_ process.VirtualMachinesContainer,
	mockSCs ...testcommon.MockTestSmartContract,
) {
	InitializeMockContractsWithVMContainerAndVMTypes(t, net, nil, [][]byte{factory.WasmVirtualMachine}, mockSCs...)
}

// InitializeMockContractsWithVMContainerAndVMTypes -
func InitializeMockContractsWithVMContainerAndVMTypes(
	t *testing.T,
	net *integrationTests.TestNetwork,
	vmContainer process.VirtualMachinesContainer,
	vmKeys [][]byte,
	mockSCs ...testcommon.MockTestSmartContract,
) {
	shardToHost, shardToInstanceBuilder :=
		CreateHostAndInstanceBuilder(t, net, vmContainer, vmKeys)
	for _, mockSC := range mockSCs {
		shardID := mockSC.GetShardID()
		mockSC.Initialize(t,
			shardToHost[shardID][string(mockSC.GetVMType())],
			shardToInstanceBuilder[shardID][string(mockSC.GetVMType())], true)
	}
}

// GetAddressForNewAccountOnWalletAndNode -
func GetAddressForNewAccountOnWalletAndNode(
	t *testing.T,
	net *integrationTests.TestNetwork,
	wallet *integrationTests.TestWalletAccount,
	node *integrationTests.TestProcessorNode,
) ([]byte, state.UserAccountHandler) {
	return GetAddressForNewAccountOnWalletAndNodeWithVM(t, net, wallet, node, net.DefaultVM)
}

// GetAddressForNewAccountOnWalletAndNodeWithVM -
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
	argsAccCreation := stateFactory.ArgsAccountCreator{
		Hasher:              &hashingMocks.HasherMock{},
		Marshaller:          &marshallerMock.MarshalizerMock{},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
	accountFactory, _ := stateFactory.NewAccountCreator(argsAccCreation)

	account, _ := accountFactory.CreateAccount(address)
	userAccount := account.(state.UserAccountHandler)
	_ = userAccount.AddToBalance(MockInitialBalance)
	userAccount.SetCode(address)
	userAccount.SetCodeHash(address)
	err = node.AccntState.SaveAccount(userAccount)
	require.Nil(t, err)
	_, _ = node.AccntState.Commit()

	return address, userAccount
}

// SetCodeMetadata -
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

// GetAddressForNewAccount -
func GetAddressForNewAccount(
	t *testing.T,
	net *integrationTests.TestNetwork,
	node *integrationTests.TestProcessorNode) ([]byte, state.UserAccountHandler) {
	return GetAddressForNewAccountWithVM(t, net, node, net.DefaultVM)
}

// GetAddressForNewAccountWithVM
func GetAddressForNewAccountWithVM(
	t *testing.T,
	net *integrationTests.TestNetwork,
	node *integrationTests.TestProcessorNode,
	vmType []byte) ([]byte, state.UserAccountHandler) {
	return GetAddressForNewAccountOnWalletAndNodeWithVM(t, net, net.Wallets[node.ShardCoordinator.SelfId()], node, vmType)
}

// MakeTestWalletAddress generates a new wallet address to be used for
// testing based on the given identifier.
func MakeTestWalletAddress(identifier string) []byte {
	return makeTestAddress(WalletAddressPrefix, identifier)
}

func makeTestAddress(_ []byte, identifier string) []byte {
	numberOfTrailingDots := vmhost.AddressSize - len(vmhost.SCAddressPrefix) - len(identifier)
	leftBytes := vmhost.SCAddressPrefix
	rightBytes := []byte(identifier + strings.Repeat(".", numberOfTrailingDots))
	return append(leftBytes, rightBytes...)
}

func CreateHostAndInstanceBuilder(t *testing.T,
	net *integrationTests.TestNetwork,
	vmContainer process.VirtualMachinesContainer,
	vmKeys [][]byte) (map[uint32]map[string]vmhost.VMHost, map[uint32]map[string]*contextmock.ExecutorMock) {
	numberOfShards := uint32(net.NumShards)
	shardToWorld := make(map[uint32]*worldmock.MockWorld, numberOfShards)
	shardToInstanceBuilder := make(map[uint32]map[string]*contextmock.ExecutorMock, numberOfShards)
	shardToHost := make(map[uint32]map[string]vmhost.VMHost, numberOfShards)

	if vmContainer != nil {
		err := net.DefaultNode.BlockchainHook.SetVMContainer(vmContainer)
		require.Nil(t, err)
	}

	for shardID := uint32(0); shardID < numberOfShards; shardID++ {
		world := worldmock.NewMockWorld()
		world.SetProvidedBlockchainHook(net.DefaultNode.BlockchainHook)
		world.SelfShardID = shardID
		shardToWorld[shardID] = world
		for _, vmKey := range vmKeys {
			instanceBuilderMock, _ := contextmock.NewExecutorMockFactory(world).CreateExecutor(executor.ExecutorFactoryArgs{})
			if shardToInstanceBuilder[shardID] == nil {
				shardToInstanceBuilder[shardID] = make(map[string]*contextmock.ExecutorMock, len(vmKeys))
			}
			shardToInstanceBuilder[shardID][string(vmKey)] = instanceBuilderMock.(*contextmock.ExecutorMock)
		}
	}

	for shardID := uint32(0); shardID < numberOfShards; shardID++ {
		node := net.NodesSharded[shardID][0]
		for _, vmType := range node.VMContainer.Keys() {
			host, err := node.VMContainer.Get(vmType)
			require.NotNil(t, host)
			require.Nil(t, err)
			if _, ok := host.(vmhost.VMHost); !ok {
				continue
			}
			host.(vmhost.VMHost).Runtime().ReplaceVMExecutor(shardToInstanceBuilder[shardID][string(vmType)])
			err = node.VMContainer.Replace(vmType, host)
			require.Nil(t, err)
			if shardToHost[shardID] == nil {
				shardToHost[shardID] = make(map[string]vmhost.VMHost, len(vmKeys))
			}
			shardToHost[shardID][string(vmType)] = host.(vmhost.VMHost)
		}
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
