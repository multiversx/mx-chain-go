package wasmvm

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-vm-common-go/txDataBuilder"
	"github.com/multiversx/mx-chain-vm-go/mock/contracts"
	"github.com/multiversx/mx-chain-vm-go/testcommon"
	test "github.com/multiversx/mx-chain-vm-go/testcommon"
	"github.com/stretchr/testify/require"
)

var LegacyAsyncCallType = []byte{0}
var NewAsyncCallType = []byte{1}

func TestMockContract_AsyncLegacy_InShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}
	testConfig := &testcommon.TestConfig{
		GasProvided:     2000,
		GasUsedByParent: 400,
	}

	transferEGLD := big.NewInt(42)

	net := integrationTests.NewTestNetworkSized(t, 1, 1, 1)
	net.Start()
	net.Step()

	net.CreateWallets(1)
	net.MintWalletsUint64(100000000000)
	owner := net.Wallets[0]

	parentAddress, _ := GetAddressForNewAccount(t, net, net.NodesSharded[0][0])

	InitializeMockContracts(
		t, net,
		test.CreateMockContract(parentAddress).
			WithConfig(testConfig).
			WithMethods(contracts.WasteGasParentMock),
	)

	txData := txDataBuilder.NewBuilder().Func("wasteGas").ToBytes()
	tx := net.CreateTx(owner, parentAddress, transferEGLD, txData)
	tx.GasLimit = testConfig.GasProvided

	_ = net.SignAndSendTx(owner, tx)

	net.Steps(2)

	parentHandler := net.GetAccountHandler(parentAddress)
	expectedEgld := big.NewInt(0)
	expectedEgld.Add(MockInitialBalance, transferEGLD)
	require.Equal(t, expectedEgld, parentHandler.GetBalance())
}

func TestMockContract_AsyncLegacy_CrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testMockContract_CrossShard(t, LegacyAsyncCallType)
}

func TestMockContract_NewAsync_CrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	testMockContract_CrossShard(t, NewAsyncCallType)
}

func testMockContract_CrossShard(t *testing.T, asyncCallType []byte) {
	transferEGLD := big.NewInt(42)

	numberOfShards := 2
	net := integrationTests.NewTestNetworkSized(t, numberOfShards, 1, 1)
	net.Start()
	net.Step()

	net.CreateWallets(2)
	net.MintWalletsUint64(100000000000)
	ownerOfParent := net.Wallets[0]

	parentAddress, _ := GetAddressForNewAccount(t, net, net.NodesSharded[0][0])
	childAddress, _ := GetAddressForNewAccount(t, net, net.NodesSharded[1][0])

	thirdPartyAddress := MakeTestWalletAddress("thirdPartyAddress")
	vaultAddress := MakeTestWalletAddress("vaultAddress")

	testConfig := &testcommon.TestConfig{
		ParentBalance: 20,
		ChildBalance:  10,

		GasProvided:        2_000_000,
		GasProvidedToChild: 1_000_000,
		GasUsedByParent:    400,

		ChildAddress:              childAddress,
		ThirdPartyAddress:         thirdPartyAddress,
		VaultAddress:              vaultAddress,
		TransferFromParentToChild: 8,

		SuccessCallback: "myCallBack",
		ErrorCallback:   "myCallBack",

		IsLegacyAsync: bytes.Equal(asyncCallType, LegacyAsyncCallType),
	}

	InitializeMockContracts(
		t, net,
		test.CreateMockContractOnShard(parentAddress, 0).
			WithBalance(testConfig.ParentBalance).
			WithConfig(testConfig).
			WithMethods(contracts.PerformAsyncCallParentMock, contracts.CallBackParentMock),
		test.CreateMockContractOnShard(childAddress, 1).
			WithBalance(testConfig.ChildBalance).
			WithConfig(testConfig).
			WithMethods(contracts.TransferToThirdPartyAsyncChildMock),
	)

	txData := txDataBuilder.
		NewBuilder().
		Func("performAsyncCall").
		Bytes([]byte{0}).
		Bytes(asyncCallType).
		ToBytes()
	tx := net.CreateTx(ownerOfParent, parentAddress, transferEGLD, txData)
	tx.GasLimit = testConfig.GasProvided

	_ = net.SignAndSendTx(ownerOfParent, tx)

	net.Steps(16)

	parentHandler, err := net.NodesSharded[0][0].BlockchainHook.GetUserAccount(parentAddress)
	require.Nil(t, err)

	parentValueA, _, err := parentHandler.AccountDataHandler().RetrieveValue(test.ParentKeyA)
	require.Nil(t, err)
	require.Equal(t, test.ParentDataA, parentValueA)

	parentValueB, _, err := parentHandler.AccountDataHandler().RetrieveValue(test.ParentKeyB)
	require.Nil(t, err)
	require.Equal(t, test.ParentDataB, parentValueB)

	callbackValue, _, err := parentHandler.AccountDataHandler().RetrieveValue(test.CallbackKey)
	require.Nil(t, err)
	require.Equal(t, test.CallbackData, callbackValue)

	originalCallerParent, _, err := parentHandler.AccountDataHandler().RetrieveValue(test.OriginalCallerParent)
	require.Nil(t, err)
	require.Equal(t, ownerOfParent.Address, originalCallerParent)

	originalCallerCallback, _, err := parentHandler.AccountDataHandler().RetrieveValue(test.OriginalCallerCallback)
	require.Nil(t, err)
	require.Equal(t, ownerOfParent.Address, originalCallerCallback)

	childHandler, err := net.NodesSharded[1][0].BlockchainHook.GetUserAccount(childAddress)
	require.Nil(t, err)

	childValue, _, err := childHandler.AccountDataHandler().RetrieveValue(test.ChildKey)
	require.Nil(t, err)
	require.Equal(t, test.ChildData, childValue)

	originalCallerChild, _, err := childHandler.AccountDataHandler().RetrieveValue(test.OriginalCallerChild)
	require.Nil(t, err)
	require.Equal(t, ownerOfParent.Address, originalCallerChild)
}

func TestMockContract_NewAsync_BackTransfer_CrossShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numberOfShards := 2
	net := integrationTests.NewTestNetworkSized(t, numberOfShards, 1, 1)
	net.Start()
	net.Step()

	transferEGLD := big.NewInt(42)
	net.CreateWallets(3)
	net.MintWalletsUint64(100000000000)
	ownerOfParent := net.Wallets[0]

	parentAddress, _ := GetAddressForNewAccount(t, net, net.NodesSharded[0][0])
	childAddress, _ := GetAddressForNewAccount(t, net, net.NodesSharded[0][0])
	nephewAddress, _ := GetAddressForNewAccount(t, net, net.NodesSharded[1][0])

	thirdPartyAddress := MakeTestWalletAddress("thirdPartyAddress")
	vaultAddress := MakeTestWalletAddress("vaultAddress")

	testConfig := &testcommon.TestConfig{
		ParentBalance: 20,
		ChildBalance:  10,

		GasProvided:        2_000_000,
		GasProvidedToChild: 1_000_000,
		GasUsedByParent:    400,

		ParentAddress:             parentAddress,
		ChildAddress:              childAddress,
		NephewAddress:             nephewAddress,
		ThirdPartyAddress:         thirdPartyAddress,
		VaultAddress:              vaultAddress,
		TransferFromParentToChild: 8,

		SuccessCallback: "myCallback",
		ErrorCallback:   "myCallback",

		ESDTTokensToTransfer: 5,
	}

	InitializeMockContracts(
		t, net,
		test.CreateMockContractOnShard(parentAddress, 0).
			WithBalance(testConfig.ParentBalance).
			WithConfig(testConfig).
			WithCodeMetadata([]byte{0, 0}).
			WithMethods(contracts.BackTransfer_ParentCallsChild),
		test.CreateMockContractOnShard(childAddress, 0).
			WithBalance(testConfig.ChildBalance).
			WithConfig(testConfig).
			WithMethods(
				contracts.BackTransfer_ChildMakesAsync,
				contracts.BackTransfer_ChildCallback,
			),
		test.CreateMockContractOnShard(nephewAddress, 1).
			WithBalance(testConfig.ChildBalance).
			WithConfig(testConfig).
			WithMethods(contracts.WasteGasChildMock),
	)

	txData := txDataBuilder.
		NewBuilder().
		Func("callChild").
		ToBytes()
	tx := net.CreateTx(ownerOfParent, parentAddress, transferEGLD, txData)
	tx.GasLimit = testConfig.GasProvided

	_ = net.SignAndSendTx(ownerOfParent, tx)

	net.Steps(16)

	parentHandler, err := net.NodesSharded[0][0].BlockchainHook.GetUserAccount(parentAddress)
	require.Nil(t, err)

	expectedEgld := big.NewInt(0)
	expectedEgld.Add(MockInitialBalance, big.NewInt(testConfig.TransferFromChildToParent))
	require.True(t, parentHandler.GetBalance().Cmp(expectedEgld) > 0)

	esdtData, err := net.NodesSharded[0][0].BlockchainHook.GetESDTToken(parentAddress, EsdtTokenIdentifier, 0)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(int64(InitialEsdt+testConfig.ESDTTokensToTransfer)), esdtData.Value)
}
