package multitransfer

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
)

func TestESDTMultiTransferThroughForwarder(t *testing.T) {
	net := integrationTests.NewTestNetworkSized(t, 2, 1, 1)
	net.Start()
	defer net.Close()

	initialVal := uint64(1000000000)
	net.MintNodeAccountsUint64(initialVal)
	net.Step()

	senderNode := net.NodesSharded[0][0]
	owner := senderNode.OwnAccount
	forwarder := net.DeployPayableSC(owner, "../testdata/forwarder.wasm")
	vault := net.DeployNonpayableSC(owner, "../testdata/vaultV2.wasm")
	vaultOtherShard := net.DeployNonpayableSC(net.NodesSharded[1][0].OwnAccount, "../testdata/vaultV2.wasm")

	// Create the fungible token
	supply := int64(1000)
	tokenID := issueFungibleToken(t, net, senderNode, "FUNG1", supply)

	// Issuer and create an SFT
	sftID := issueNft(net, senderNode, "SFT1", true)
	createSFT(t, net, senderNode, sftID, 1, supply)

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

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, sftID, 1, supply)
	esdt.CheckAddressHasESDTTokens(t, forwarder, net.Nodes, tokenID, supply)

	// transfer to a user from another shard
	transfers := []*esdtTransfer{
		{
			tokenIdentifier: tokenID,
			nonce:           0,
			amount:          100,
		}}
	destAddress := net.NodesSharded[1][0].OwnAccount.Address
	multiTransferThroughForwarder(
		net,
		senderNode.OwnAccount,
		forwarder,
		"multi_transfer_via_async",
		transfers,
		destAddress)

	esdt.CheckAddressHasESDTTokens(t, forwarder, net.Nodes, tokenID, 900)
	esdt.CheckAddressHasESDTTokens(t, destAddress, net.Nodes, tokenID, 100)

	// transfer to vault, same shard
	multiTransferThroughForwarder(
		net,
		senderNode.OwnAccount,
		forwarder,
		"forward_sync_accept_funds_multi_transfer",
		transfers,
		vault)

	esdt.CheckAddressHasESDTTokens(t, forwarder, net.Nodes, tokenID, 800)
	esdt.CheckAddressHasESDTTokens(t, vault, net.Nodes, tokenID, 100)

	// transfer fungible and non-fungible
	// transfer to vault, same shard
	transfers = []*esdtTransfer{
		{
			tokenIdentifier: tokenID,
			nonce:           0,
			amount:          100,
		},
		{
			tokenIdentifier: sftID,
			nonce:           1,
			amount:          100,
		},
	}
	multiTransferThroughForwarder(
		net,
		senderNode.OwnAccount,
		forwarder,
		"forward_sync_accept_funds_multi_transfer",
		transfers,
		vault)

	esdt.CheckAddressHasESDTTokens(t, forwarder, net.Nodes, tokenID, 700)
	esdt.CheckAddressHasESDTTokens(t, vault, net.Nodes, tokenID, 200)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, sftID, 1, 900)
	esdt.CheckAddressHasTokens(t, vault, net.Nodes, sftID, 1, 100)

	// transfer fungible and non-fungible
	// transfer to vault, cross shard via transfer and execute
	transfers = []*esdtTransfer{
		{
			tokenIdentifier: tokenID,
			nonce:           0,
			amount:          100,
		},
		{
			tokenIdentifier: sftID,
			nonce:           1,
			amount:          100,
		},
	}
	multiTransferThroughForwarder(
		net,
		senderNode.OwnAccount,
		forwarder,
		"forward_transf_exec_accept_funds_multi_transfer",
		transfers,
		vaultOtherShard)

	esdt.CheckAddressHasESDTTokens(t, forwarder, net.Nodes, tokenID, 600)
	esdt.CheckAddressHasESDTTokens(t, vaultOtherShard, net.Nodes, tokenID, 100)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, sftID, 1, 800)
	esdt.CheckAddressHasTokens(t, vaultOtherShard, net.Nodes, sftID, 1, 100)

	// transfer to vault, cross shard, via async call
	transfers = []*esdtTransfer{
		{
			tokenIdentifier: tokenID,
			nonce:           0,
			amount:          100,
		},
		{
			tokenIdentifier: sftID,
			nonce:           1,
			amount:          100,
		},
	}
	multiTransferThroughForwarder(
		net,
		senderNode.OwnAccount,
		forwarder,
		"multi_transfer_via_async",
		transfers,
		vaultOtherShard)

	esdt.CheckAddressHasESDTTokens(t, forwarder, net.Nodes, tokenID, 500)
	esdt.CheckAddressHasESDTTokens(t, vaultOtherShard, net.Nodes, tokenID, 200)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, sftID, 1, 700)
	esdt.CheckAddressHasTokens(t, vaultOtherShard, net.Nodes, sftID, 1, 200)
}

func multiTransferThroughForwarder(
	net *integrationTests.TestNetwork,
	ownerWallet *integrationTests.TestWalletAccount,
	forwarderAddress []byte,
	function string,
	transfers []*esdtTransfer,
	destAddress []byte) {

	txData := txDataBuilder.NewBuilder()
	txData.Func(function).Bytes(destAddress)

	for _, transfer := range transfers {
		txData.Str(transfer.tokenIdentifier).Int64(transfer.nonce).Int64(transfer.amount)
	}

	tx := net.CreateTxUint64(ownerWallet, forwarderAddress, 0, txData.ToBytes())
	tx.GasLimit = net.MaxGasLimit / 2
	_ = net.SignAndSendTx(ownerWallet, tx)
	net.Steps(10)
}
