package esdtMultiTransferThroughForwarder

import (
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/esdt"
	wasmvm "github.com/multiversx/mx-chain-go/integrationTests/vm/wasm/wasmvm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	test "github.com/multiversx/mx-chain-vm-go/testcommon"
)

func TestESDTMultiTransferThroughForwarder_LegacyAsync_MockContracts(t *testing.T) {
	ESDTMultiTransferThroughForwarder_MockContracts(t, true)
}

func TestESDTMultiTransferThroughForwarder_NewAsync_MockContracts(t *testing.T) {
	ESDTMultiTransferThroughForwarder_MockContracts(t, false)
}

func ESDTMultiTransferThroughForwarder_MockContracts(t *testing.T, legacyAsync bool) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net, ownerShard1, ownerShard2, senderNode, forwarder, vaultShard1, vaultShard2 :=
		ESDTMultiTransferThroughForwarder_MockContracts_SetupNetwork(t)
	defer net.Close()

	ESDTMultiTransferThroughForwarder_MockContracts_Deploy(t,
		legacyAsync,
		net,
		ownerShard1,
		ownerShard2,
		forwarder,
		vaultShard1,
		vaultShard2)

	ESDTMultiTransferThroughForwarder_RunStepsAndAsserts(t,
		net,
		senderNode,
		ownerShard1,
		ownerShard2,
		forwarder,
		vaultShard1,
		vaultShard2,
	)
}

func ESDTMultiTransferThroughForwarder_MockContracts_SetupNetwork(t *testing.T) (*integrationTests.TestNetwork, *integrationTests.TestWalletAccount, *integrationTests.TestWalletAccount, *integrationTests.TestProcessorNode, []byte, []byte, []byte) {
	net := integrationTests.NewTestNetworkSized(t, 2, 1, 1)
	net.Start().Step()

	net.CreateUninitializedWallets(2)
	owner := net.CreateWalletOnShard(0, 0)
	owner2 := net.CreateWalletOnShard(1, 1)

	initialVal := uint64(1000000000)
	net.MintWalletsUint64(initialVal)

	node0shard0 := net.NodesSharded[0][0]
	node0shard1 := net.NodesSharded[1][0]

	forwarder, forwarderAccount := wasmvm.GetAddressForNewAccountOnWalletAndNode(t, net, owner, node0shard0)
	wasmvm.SetCodeMetadata(t, []byte{0, vmcommon.MetadataPayable}, node0shard0, forwarderAccount)

	vaultShard1, vaultShard1Account := wasmvm.GetAddressForNewAccountOnWalletAndNode(t, net, owner, node0shard0)
	wasmvm.SetCodeMetadata(t, []byte{0, vmcommon.MetadataPayable}, node0shard0, vaultShard1Account)

	vaultShard2, _ := wasmvm.GetAddressForNewAccountOnWalletAndNode(t, net, owner2, node0shard1)

	return net, owner, owner2, node0shard0, forwarder, vaultShard1, vaultShard2
}

func ESDTMultiTransferThroughForwarder_MockContracts_Deploy(t *testing.T, legacyAsync bool, net *integrationTests.TestNetwork, ownerShard1 *integrationTests.TestWalletAccount, ownerShard2 *integrationTests.TestWalletAccount, forwarder []byte, vaultShard1 []byte, vaultShard2 []byte) {
	testConfig := &test.TestConfig{
		IsLegacyAsync: legacyAsync,
		// used for new async
		SuccessCallback:    "callBack",
		ErrorCallback:      "callBack",
		GasProvidedToChild: 500_000,
		GasToLock:          300_000,
	}

	wasmvm.InitializeMockContractsWithVMContainer(
		t, net,
		net.NodesSharded[0][0].VMContainer,
		test.CreateMockContractOnShard(forwarder, 0).
			WithOwnerAddress(ownerShard1.Address).
			WithConfig(testConfig).
			WithMethods(
				esdt.MultiTransferViaAsyncMock,
				esdt.SyncMultiTransferMock,
				esdt.MultiTransferExecuteMock,
				esdt.EmptyCallbackMock),
		test.CreateMockContractOnShard(vaultShard1, 0).
			WithOwnerAddress(ownerShard1.Address).
			WithConfig(testConfig).
			WithMethods(esdt.AcceptFundsEchoMock),
		test.CreateMockContractOnShard(vaultShard2, 1).
			WithOwnerAddress(ownerShard2.Address).
			WithConfig(testConfig).
			WithMethods(esdt.AcceptMultiFundsEchoMock),
	)
}

func TestESDTMultiTransferWithWrongArgumentsSFT_MockContracts(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net, ownerShard1, ownerShard2, senderNode, forwarder, _, vaultShard2 :=
		ESDTMultiTransferThroughForwarder_MockContracts_SetupNetwork(t)
	defer net.Close()

	ESDTMultiTransferWithWrongArguments_MockContracts_Deploy(t, net, ownerShard1, forwarder, vaultShard2, ownerShard2)

	ESDTMultiTransferWithWrongArgumentsSFT_RunStepsAndAsserts(t, net, senderNode, ownerShard1, forwarder, vaultShard2)
}

func TestESDTMultiTransferWithWrongArgumentsFungible_MockContracts(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net, ownerShard1, ownerShard2, senderNode, forwarder, _, vaultShard2 :=
		ESDTMultiTransferThroughForwarder_MockContracts_SetupNetwork(t)
	defer net.Close()

	ESDTMultiTransferWithWrongArguments_MockContracts_Deploy(t, net, ownerShard1, forwarder, vaultShard2, ownerShard2)

	ESDTMultiTransferWithWrongArgumentsFungible_RunStepsAndAsserts(t, net, senderNode, ownerShard1, forwarder, vaultShard2)
}

func ESDTMultiTransferWithWrongArguments_MockContracts_Deploy(t *testing.T, net *integrationTests.TestNetwork, ownerShard1 *integrationTests.TestWalletAccount, forwarder []byte, vaultShard2 []byte, ownerShard2 *integrationTests.TestWalletAccount) {
	testConfig := &test.TestConfig{
		IsLegacyAsync: true,
	}

	wasmvm.InitializeMockContractsWithVMContainer(
		t, net,
		net.NodesSharded[0][0].VMContainer,
		test.CreateMockContractOnShard(forwarder, 0).
			WithOwnerAddress(ownerShard1.Address).
			WithConfig(testConfig).
			WithMethods(esdt.DoAsyncCallMock),
		test.CreateMockContractOnShard(vaultShard2, 1).
			WithOwnerAddress(ownerShard2.Address).
			WithConfig(testConfig).
			WithMethods(esdt.AcceptFundsEchoMock),
	)
}
