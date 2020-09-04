package scTxs

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/relayedTx"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	vmFactory "github.com/ElrondNetwork/elrond-go/process/factory"
)

func TestRelayedTransactionInMultiShardEnvironmentWithSmartContractTX(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer, advertiser := relayedTx.CreateGeneralSetupForRelayTxTest()
	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	sendValue := big.NewInt(5)
	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	receiverAddress1 := []byte("12345678901234567890123456789012")
	receiverAddress2 := []byte("12345678901234567890123456789011")

	ownerNode := nodes[0]
	initialSupply := "00" + hex.EncodeToString(big.NewInt(100000000000).Bytes())
	scCode := arwen.GetSCCode("../../../vm/arwen/testdata/erc20-c-03/wrc20_arwen.wasm")
	scAddress, _ := ownerNode.BlockchainHook.NewAddress(ownerNode.OwnAccount.Address, ownerNode.OwnAccount.Nonce, vmFactory.ArwenVirtualMachine)

	integrationTests.CreateAndSendTransactionWithGasLimit(
		nodes[0],
		big.NewInt(0),
		20000,
		make([]byte, 32),
		[]byte(arwen.CreateDeployTxData(scCode)+"@"+initialSupply),
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)

	transferTokenVMGas := uint64(6960)
	transferTokenBaseGas := ownerNode.EconomicsData.ComputeGasLimit(&transaction.Transaction{Data: []byte("transferToken@" + hex.EncodeToString(receiverAddress1) + "@00" + hex.EncodeToString(sendValue.Bytes()))})
	transferTokenFullGas := transferTokenBaseGas + transferTokenVMGas

	initialTokenSupply := big.NewInt(1000000000)
	initialPlusForGas := uint64(1000)
	for _, player := range players {
		integrationTests.CreateAndSendTransactionWithGasLimit(
			ownerNode,
			big.NewInt(0),
			transferTokenFullGas+initialPlusForGas,
			scAddress,
			[]byte("transferToken@"+hex.EncodeToString(player.Address)+"@00"+hex.EncodeToString(initialTokenSupply.Bytes())),
			integrationTests.ChainID,
			integrationTests.MinTransactionVersion,
		)
	}
	time.Sleep(time.Second)

	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		for _, player := range players {
			_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, big.NewInt(0),
				transferTokenFullGas, []byte("transferToken@"+hex.EncodeToString(receiverAddress1)+"@00"+hex.EncodeToString(sendValue.Bytes())))
			_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, scAddress, big.NewInt(0),
				transferTokenFullGas, []byte("transferToken@"+hex.EncodeToString(receiverAddress2)+"@00"+hex.EncodeToString(sendValue.Bytes())))
		}

		time.Sleep(time.Second)
	}

	roundToPropagateMultiShard := int64(20)
	for i := int64(0); i <= roundToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)

	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)

	relayedTx.CheckSCBalance(t, ownerNode, scAddress, receiverAddress1, finalBalance)
	relayedTx.CheckSCBalance(t, ownerNode, scAddress, receiverAddress1, finalBalance)

	players = append(players, relayer)
	relayedTx.CheckPlayerBalances(t, nodes, players)
}
