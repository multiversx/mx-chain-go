package esdtMultiTransferThroughForwarder

import (
	"testing"

	test "github.com/ElrondNetwork/arwen-wasm-vm/v1_5/testcommon"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	arwenvm "github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen/arwenvm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt/localFuncs"
	multitransfer "github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt/multi-transfer"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

func TestESDTMultiTransferThroughForwarder_MockContracts(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net := integrationTests.NewTestNetworkSized(t, 2, 1, 1)
	net.Start()
	defer net.Close()

	initialVal := uint64(1000000000)
	net.MintNodeAccountsUint64(initialVal)
	net.Step()

	net.CreateUninitializedWallets(2)
	owner := net.CreateWalletOnShard(0, 0)
	owner2 := net.CreateWalletOnShard(1, 1)
	net.MintWalletsUint64(initialVal)

	node0shard0 := net.NodesSharded[0][0]
	node0shard1 := net.NodesSharded[1][0]

	forwarder, forwarderAccount := arwenvm.GetAddressForNewAccountOnWalletAndNode(t, net, owner, node0shard0)
	arwenvm.SetCodeMetadata(t, []byte{0, vmcommon.MetadataPayable}, node0shard0, forwarderAccount)

	vaultShard1, vaultShard1Account := arwenvm.GetAddressForNewAccountOnWalletAndNode(t, net, owner, node0shard0)
	arwenvm.SetCodeMetadata(t, []byte{0, vmcommon.MetadataPayable}, node0shard0, vaultShard1Account)

	vaultShard2, _ := arwenvm.GetAddressForNewAccountOnWalletAndNode(t, net, owner2, node0shard1)

	// Create the fungible token
	supply := int64(1000)
	tokenID := multitransfer.IssueFungibleTokenWithIssuerAddress(t, net, node0shard0, owner, "FUNG1", supply)

	// Issue and create an SFT
	sftID := multitransfer.IssueNftWithIssuerAddress(net, node0shard0, owner, "SFT1", true)
	multitransfer.CreateSFT(t, net, node0shard0, owner, sftID, 1, supply)

	// Send the tokens to the forwarder SC
	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionMultiESDTNFTTransfer)
	txData.Bytes(forwarder).Int(2)
	txData.Str(tokenID).Int(0).Int64(supply)
	txData.Str(sftID).Int(1).Int64(supply)

	tx := net.CreateTxUint64(owner, owner.Address, 0, txData.ToBytes())
	tx.GasLimit = net.MaxGasLimit / 2
	_ = net.SignAndSendTx(owner, tx)
	net.Steps(4)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, supply)
	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, supply)

	testConfig := &test.TestConfig{
		IsLegacyAsync: true,
	}

	arwenvm.InitializeMockContracts(
		t, net,
		test.CreateMockContractOnShard(forwarder, 0).
			WithOwnerAddress(owner.Address).
			WithBalance(int64(initialVal)).
			WithConfig(testConfig).
			WithMethods(
				localFuncs.MultiTransferViaAsyncMock,
				localFuncs.SyncMultiTransferMock,
				localFuncs.MultiTransferExecuteMock,
				localFuncs.EmptyCallbackMock),
		test.CreateMockContractOnShard(vaultShard1, 0).
			WithOwnerAddress(owner.Address).
			WithBalance(int64(initialVal)).
			WithConfig(testConfig).
			WithMethods(localFuncs.AcceptFundsEchoMock),
		test.CreateMockContractOnShard(vaultShard2, 1).
			WithOwnerAddress(owner2.Address).
			WithBalance(int64(initialVal)).
			WithConfig(testConfig).
			WithMethods(localFuncs.AcceptMultiFundsEchoMock),
	)

	// transfer to a user from another shard
	transfers := []*multitransfer.EsdtTransfer{
		{
			TokenIdentifier: tokenID,
			Nonce:           0,
			Amount:          100,
		}}

	multiTransferThroughForwarder(
		net,
		owner,
		forwarder,
		"multi_transfer_via_async",
		transfers,
		owner2.Address)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 900)
	esdt.CheckAddressHasTokens(t, owner2.Address, net.Nodes, []byte(tokenID), 0, 100)

	// transfer to vault, same shard
	multiTransferThroughForwarder(
		net,
		owner,
		forwarder,
		"forward_sync_accept_funds_multi_transfer",
		transfers,
		vaultShard1)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 800)
	esdt.CheckAddressHasTokens(t, vaultShard1, net.Nodes, []byte(tokenID), 0, 100)

	// transfer fungible and non-fungible
	// transfer to vault, same shard
	transfers = []*multitransfer.EsdtTransfer{
		{
			TokenIdentifier: tokenID,
			Nonce:           0,
			Amount:          100,
		},
		{
			TokenIdentifier: sftID,
			Nonce:           1,
			Amount:          100,
		},
	}
	multiTransferThroughForwarder(
		net,
		owner,
		forwarder,
		"forward_sync_accept_funds_multi_transfer",
		transfers,
		vaultShard1)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 700)
	esdt.CheckAddressHasTokens(t, vaultShard1, net.Nodes, []byte(tokenID), 0, 200)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 900)
	esdt.CheckAddressHasTokens(t, vaultShard1, net.Nodes, []byte(sftID), 1, 100)

	// transfer fungible and non-fungible
	// transfer to vault, cross shard via transfer and execute
	transfers = []*multitransfer.EsdtTransfer{
		{
			TokenIdentifier: tokenID,
			Nonce:           0,
			Amount:          100,
		},
		{
			TokenIdentifier: sftID,
			Nonce:           1,
			Amount:          100,
		},
	}
	multiTransferThroughForwarder(
		net,
		owner,
		forwarder,
		"forward_transf_exec_accept_funds_multi_transfer",
		transfers,
		vaultShard2)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 600)
	esdt.CheckAddressHasTokens(t, vaultShard2, net.Nodes, []byte(tokenID), 0, 100)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 800)
	esdt.CheckAddressHasTokens(t, vaultShard2, net.Nodes, []byte(sftID), 1, 100)

	// transfer to vault, cross shard, via async call
	transfers = []*multitransfer.EsdtTransfer{
		{
			TokenIdentifier: tokenID,
			Nonce:           0,
			Amount:          100,
		},
		{
			TokenIdentifier: sftID,
			Nonce:           1,
			Amount:          100,
		},
	}
	multiTransferThroughForwarder(
		net,
		owner,
		forwarder,
		"multi_transfer_via_async",
		transfers,
		vaultShard2)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 500)
	esdt.CheckAddressHasTokens(t, vaultShard2, net.Nodes, []byte(tokenID), 0, 200)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 700)
	esdt.CheckAddressHasTokens(t, vaultShard2, net.Nodes, []byte(sftID), 1, 200)
}
