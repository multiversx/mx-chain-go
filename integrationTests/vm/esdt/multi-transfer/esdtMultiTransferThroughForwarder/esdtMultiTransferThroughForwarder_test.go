package esdtMultiTransferThroughForwarder

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/esdt"
	multitransfer "github.com/multiversx/mx-chain-go/integrationTests/vm/esdt/multi-transfer"
	"github.com/multiversx/mx-chain-go/testscommon/txDataBuilder"
)

func TestESDTMultiTransferThroughForwarder(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net := integrationTests.NewTestNetworkSized(t, 2, 1, 1)
	net.Start()
	defer net.Close()

	initialVal := uint64(1000000000)
	net.MintNodeAccountsUint64(initialVal)
	net.Step()

	senderNode := net.NodesSharded[0][0]
	destAccount := net.NodesSharded[1][0].OwnAccount

	owner := senderNode.OwnAccount
	forwarder := net.DeployPayableSC(owner, "../../testdata/forwarder.wasm")
	vault := net.DeployNonpayableSC(owner, "../../testdata/vaultV2.wasm")
	vaultOtherShard := net.DeployNonpayableSC(net.NodesSharded[1][0].OwnAccount, "../../testdata/vaultV2.wasm")

	ESDTMultiTransferThroughForwarder_RunStepsAndAsserts(
		t,
		net,
		senderNode,
		senderNode.OwnAccount,
		destAccount,
		forwarder,
		vault,
		vaultOtherShard,
	)
}

func ESDTMultiTransferThroughForwarder_RunStepsAndAsserts(
	t *testing.T,
	net *integrationTests.TestNetwork,
	ownerShard1Node *integrationTests.TestProcessorNode,
	ownerWallet *integrationTests.TestWalletAccount,
	ownerShard2Wallet *integrationTests.TestWalletAccount,
	forwarder []byte,
	vaultShard1 []byte,
	vaultShard2 []byte,
) {
	// Create the fungible token
	supply := int64(1000)
	tokenID := multitransfer.IssueFungibleTokenWithIssuerAddress(t, net, ownerShard1Node, ownerWallet, "FUNG1", supply)

	// Issue and create an SFT
	sftID := multitransfer.IssueNftWithIssuerAddress(net, ownerShard1Node, ownerWallet, "SFT1", true)
	multitransfer.CreateSFT(t, net, ownerShard1Node, ownerWallet, sftID, 1, supply)

	// Send the tokens to the forwarder SC
	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionMultiESDTNFTTransfer)
	txData.Bytes(forwarder).Int(2)
	txData.Str(tokenID).Int(0).Int64(supply)
	txData.Str(sftID).Int(1).Int64(supply)

	tx := net.CreateTxUint64(ownerWallet, ownerWallet.Address, 0, txData.ToBytes())
	tx.GasLimit = net.MaxGasLimit / 2
	_ = net.SignAndSendTx(ownerWallet, tx)
	net.Steps(4)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, supply)
	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, supply)

	// transfer to a user from another shard
	transfers := []*multitransfer.EsdtTransfer{
		{
			TokenIdentifier: tokenID,
			Nonce:           0,
			Amount:          100,
		}}

	multiTransferThroughForwarder(
		net,
		ownerWallet,
		forwarder,
		"multi_transfer_via_async",
		transfers,
		ownerShard2Wallet.Address)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 900)
	esdt.CheckAddressHasTokens(t, ownerShard2Wallet.Address, net.Nodes, []byte(tokenID), 0, 100)

	// transfer to vault, same shard
	multiTransferThroughForwarder(
		net,
		ownerWallet,
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
		ownerWallet,
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
		ownerWallet,
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
		ownerWallet,
		forwarder,
		"multi_transfer_via_async",
		transfers,
		vaultShard2)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 500)
	esdt.CheckAddressHasTokens(t, vaultShard2, net.Nodes, []byte(tokenID), 0, 200)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 700)
	esdt.CheckAddressHasTokens(t, vaultShard2, net.Nodes, []byte(sftID), 1, 200)
}

func multiTransferThroughForwarder(
	net *integrationTests.TestNetwork,
	ownerWallet *integrationTests.TestWalletAccount,
	forwarderAddress []byte,
	function string,
	transfers []*multitransfer.EsdtTransfer,
	destAddress []byte) {

	txData := txDataBuilder.NewBuilder()
	txData.Func(function).Bytes(destAddress)

	for _, transfer := range transfers {
		txData.Str(transfer.TokenIdentifier).Int64(transfer.Nonce).Int64(transfer.Amount)
	}

	tx := net.CreateTxUint64(ownerWallet, forwarderAddress, 0, txData.ToBytes())
	tx.GasLimit = net.MaxGasLimit / 2
	_ = net.SignAndSendTx(ownerWallet, tx)
	net.Steps(10)
}

func TestESDTMultiTransferWithWrongArgumentsSFT(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net := integrationTests.NewTestNetworkSized(t, 2, 1, 1)
	net.Start()
	defer net.Close()

	initialVal := uint64(1000000000)
	net.MintNodeAccountsUint64(initialVal)
	net.Step()

	senderNode := net.NodesSharded[0][0]
	owner := senderNode.OwnAccount
	forwarder := net.DeployNonpayableSC(owner, "../../testdata/execute/output/execute.wasm")
	vaultOtherShard := net.DeployNonpayableSC(net.NodesSharded[1][0].OwnAccount, "../../testdata/vault.wasm")

	ESDTMultiTransferWithWrongArgumentsSFT_RunStepsAndAsserts(t, net, senderNode, senderNode.OwnAccount, forwarder, vaultOtherShard)
}

func ESDTMultiTransferWithWrongArgumentsSFT_RunStepsAndAsserts(
	t *testing.T,
	net *integrationTests.TestNetwork,
	ownerShard1Node *integrationTests.TestProcessorNode,
	ownerWallet *integrationTests.TestWalletAccount,
	forwarder []byte,
	vaultShard2 []byte) {
	// Issue and create SFT
	supply := int64(1000)
	sftID := multitransfer.IssueNftWithIssuerAddress(net, ownerShard1Node, ownerWallet, "SFT1", true)
	multitransfer.CreateSFT(t, net, ownerShard1Node, ownerWallet, sftID, 1, supply)

	// Send the tokens to the forwarder SC
	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionMultiESDTNFTTransfer)
	txData.Bytes(forwarder).Int(1)
	txData.Str(sftID).Int(1).Int64(10).Str("doAsyncCall").Bytes(forwarder)
	txData.Bytes([]byte{}).Str(core.BuiltInFunctionMultiESDTNFTTransfer).Int(6).Bytes(vaultShard2).Int(1).Str(sftID).Int(1).Int(1).Bytes([]byte{})
	tx := net.CreateTxUint64(ownerWallet, ownerWallet.Address, 0, txData.ToBytes())
	tx.GasLimit = net.MaxGasLimit / 2
	_ = net.SignAndSendTx(ownerWallet, tx)
	net.Steps(12)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 10)
	esdt.CheckAddressHasTokens(t, vaultShard2, net.Nodes, []byte(sftID), 1, 0)
	esdt.CheckAddressHasTokens(t, ownerWallet.Address, net.Nodes, []byte(sftID), 1, supply-10)
}

func TestESDTMultiTransferWithWrongArgumentsFungible(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	net := integrationTests.NewTestNetworkSized(t, 2, 1, 1)
	net.Start()
	defer net.Close()

	initialVal := uint64(1000000000)
	net.MintNodeAccountsUint64(initialVal)
	net.Step()

	senderNode := net.NodesSharded[0][0]
	owner := senderNode.OwnAccount
	forwarder := net.DeployNonpayableSC(owner, "../../testdata/execute/output/execute.wasm")
	vaultOtherShard := net.DeploySCWithInitArgs(net.NodesSharded[1][0].OwnAccount, "../../testdata/contract.wasm", false, []byte{10})

	ESDTMultiTransferWithWrongArgumentsFungible_RunStepsAndAsserts(t, net, senderNode, senderNode.OwnAccount, forwarder, vaultOtherShard)
}

func ESDTMultiTransferWithWrongArgumentsFungible_RunStepsAndAsserts(
	t *testing.T,
	net *integrationTests.TestNetwork,
	senderNode *integrationTests.TestProcessorNode,
	ownerWallet *integrationTests.TestWalletAccount,
	forwarder []byte, vaultOtherShard []byte) {
	// Create the fungible token
	supply := int64(1000)
	tokenID := multitransfer.IssueFungibleTokenWithIssuerAddress(t, net, senderNode, ownerWallet, "FUNG1", supply)

	// Send the tokens to the forwarder SC
	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionMultiESDTNFTTransfer)
	txData.Bytes(forwarder).Int(1)
	txData.Str(tokenID).Int(0).Int64(80).Str("doAsyncCall").Bytes(forwarder)
	txData.Bytes([]byte{}).Str(core.BuiltInFunctionMultiESDTNFTTransfer).Int(6).Bytes(vaultOtherShard).Int(1).Str(tokenID).Int(0).Int(42).Bytes([]byte{})
	tx := net.CreateTxUint64(ownerWallet, ownerWallet.Address, 0, txData.ToBytes())
	tx.GasLimit = 104000
	_ = net.SignAndSendTx(ownerWallet, tx)
	net.Steps(12)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 80)
	esdt.CheckAddressHasTokens(t, vaultOtherShard, net.Nodes, []byte(tokenID), 0, 0)
	esdt.CheckAddressHasTokens(t, ownerWallet.Address, net.Nodes, []byte(tokenID), 0, supply-80)
}
