package esdtMultiTransferThroughForwarder

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt"
	multitransfer "github.com/ElrondNetwork/elrond-go/integrationTests/vm/esdt/multi-transfer"
	"github.com/ElrondNetwork/elrond-go/testscommon/txDataBuilder"
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
	owner := senderNode.OwnAccount
	forwarder := net.DeployPayableSC(owner, "../../testdata/forwarder.wasm")
	vault := net.DeployNonpayableSC(owner, "../../testdata/vaultV2.wasm")
	vaultOtherShard := net.DeployNonpayableSC(net.NodesSharded[1][0].OwnAccount, "../../testdata/vaultV2.wasm")

	// Create the fungible token
	supply := int64(1000)
	tokenID := multitransfer.IssueFungibleToken(t, net, senderNode, "FUNG1", supply)

	// Issue and create an SFT
	sftID := multitransfer.IssueNft(net, senderNode, "SFT1", true)
	multitransfer.CreateSFT(t, net, senderNode, sftID, 1, supply)

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

	// transfer to a user from another shard
	transfers := []*multitransfer.EsdtTransfer{
		{
			TokenIdentifier: tokenID,
			Nonce:           0,
			Amount:          100,
		}}
	destAddress := net.NodesSharded[1][0].OwnAccount.Address
	multiTransferThroughForwarder(
		net,
		senderNode.OwnAccount,
		forwarder,
		"multi_transfer_via_async",
		transfers,
		destAddress)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 900)
	esdt.CheckAddressHasTokens(t, destAddress, net.Nodes, []byte(tokenID), 0, 100)

	// transfer to vault, same shard
	multiTransferThroughForwarder(
		net,
		senderNode.OwnAccount,
		forwarder,
		"forward_sync_accept_funds_multi_transfer",
		transfers,
		vault)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 800)
	esdt.CheckAddressHasTokens(t, vault, net.Nodes, []byte(tokenID), 0, 100)

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
		senderNode.OwnAccount,
		forwarder,
		"forward_sync_accept_funds_multi_transfer",
		transfers,
		vault)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 700)
	esdt.CheckAddressHasTokens(t, vault, net.Nodes, []byte(tokenID), 0, 200)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 900)
	esdt.CheckAddressHasTokens(t, vault, net.Nodes, []byte(sftID), 1, 100)

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
		senderNode.OwnAccount,
		forwarder,
		"forward_transf_exec_accept_funds_multi_transfer",
		transfers,
		vaultOtherShard)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 600)
	esdt.CheckAddressHasTokens(t, vaultOtherShard, net.Nodes, []byte(tokenID), 0, 100)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 800)
	esdt.CheckAddressHasTokens(t, vaultOtherShard, net.Nodes, []byte(sftID), 1, 100)

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
		senderNode.OwnAccount,
		forwarder,
		"multi_transfer_via_async",
		transfers,
		vaultOtherShard)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 500)
	esdt.CheckAddressHasTokens(t, vaultOtherShard, net.Nodes, []byte(tokenID), 0, 200)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 700)
	esdt.CheckAddressHasTokens(t, vaultOtherShard, net.Nodes, []byte(sftID), 1, 200)
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
	forwarder := net.DeployNonpayableSC(owner, "../../testdata/execute.wasm")
	vaultOtherShard := net.DeployNonpayableSC(net.NodesSharded[1][0].OwnAccount, "../../testdata/vaultV2.wasm")

	// Issue and create SFT
	supply := int64(1000)
	sftID := multitransfer.IssueNft(net, senderNode, "SFT1", true)
	multitransfer.CreateSFT(t, net, senderNode, sftID, 1, supply)

	// Send the tokens to the forwarder SC
	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionMultiESDTNFTTransfer)
	txData.Bytes(forwarder).Int(1)
	txData.Str(sftID).Int(1).Int64(10).Str("doAsyncCall").Bytes(forwarder)
	txData.Bytes([]byte{}).Str(core.BuiltInFunctionMultiESDTNFTTransfer).Int(6).Bytes(vaultOtherShard).Int(1).Str(sftID).Int(1).Int(1).Bytes([]byte{})
	tx := net.CreateTxUint64(owner, owner.Address, 0, txData.ToBytes())
	tx.GasLimit = net.MaxGasLimit / 2
	_ = net.SignAndSendTx(owner, tx)
	net.Steps(12)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(sftID), 1, 10)
	esdt.CheckAddressHasTokens(t, vaultOtherShard, net.Nodes, []byte(sftID), 1, 0)
	esdt.CheckAddressHasTokens(t, owner.Address, net.Nodes, []byte(sftID), 1, supply-10)
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
	forwarder := net.DeployNonpayableSC(owner, "../../testdata/execute.wasm")
	vaultOtherShard := net.DeploySCWithInitArgs(net.NodesSharded[1][0].OwnAccount, "../../testdata/contract.wasm", false, []byte{10})

	// Create the fungible token
	supply := int64(1000)
	tokenID := multitransfer.IssueFungibleToken(t, net, senderNode, "FUNG1", supply)

	// Send the tokens to the forwarder SC
	txData := txDataBuilder.NewBuilder()
	txData.Func(core.BuiltInFunctionMultiESDTNFTTransfer)
	txData.Bytes(forwarder).Int(1)
	txData.Str(tokenID).Int(0).Int64(80).Str("doAsyncCall").Bytes(forwarder)
	txData.Bytes([]byte{}).Str(core.BuiltInFunctionMultiESDTNFTTransfer).Int(6).Bytes(vaultOtherShard).Int(1).Str(tokenID).Int(0).Int(42).Bytes([]byte{})
	tx := net.CreateTxUint64(owner, owner.Address, 0, txData.ToBytes())
	tx.GasLimit = 104000
	_ = net.SignAndSendTx(owner, tx)
	net.Steps(12)

	esdt.CheckAddressHasTokens(t, forwarder, net.Nodes, []byte(tokenID), 0, 80)
	esdt.CheckAddressHasTokens(t, vaultOtherShard, net.Nodes, []byte(tokenID), 0, 0)
	esdt.CheckAddressHasTokens(t, owner.Address, net.Nodes, []byte(tokenID), 0, supply-80)
}
