package relayedTx

import (
	"encoding/hex"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	vmFactory "github.com/multiversx/mx-chain-go/process/factory"
	"github.com/stretchr/testify/assert"
)

func TestRelayedTransactionV2InMultiShardEnvironmentWithSmartContractTX(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	nodes, idxProposers, players, relayer := CreateGeneralSetupForRelayTxTest()
	defer func() {
		for _, n := range nodes {
			n.Close()
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
	scCode := wasm.GetSCCode("../../vm/wasm/testdata/erc20-c-03/wrc20_wasm.wasm")
	scAddress, _ := ownerNode.BlockchainHook.NewAddress(ownerNode.OwnAccount.Address, ownerNode.OwnAccount.Nonce, vmFactory.WasmVirtualMachine)

	integrationTests.CreateAndSendTransactionWithGasLimit(
		nodes[0],
		big.NewInt(0),
		20000,
		make([]byte, 32),
		[]byte(wasm.CreateDeployTxData(scCode)+"@"+initialSupply),
		integrationTests.ChainID,
		integrationTests.MinTransactionVersion,
	)

	transferTokenVMGas := uint64(7200)
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
			_ = CreateAndSendRelayedAndUserTxV2(nodes, relayer, player, scAddress, big.NewInt(0),
				transferTokenFullGas, []byte("transferToken@"+hex.EncodeToString(receiverAddress1)+"@00"+hex.EncodeToString(sendValue.Bytes())))
			_ = CreateAndSendRelayedAndUserTxV2(nodes, relayer, player, scAddress, big.NewInt(0),
				transferTokenFullGas, []byte("transferToken@"+hex.EncodeToString(receiverAddress2)+"@00"+hex.EncodeToString(sendValue.Bytes())))
		}

		time.Sleep(integrationTests.StepDelay)
	}

	roundToPropagateMultiShard := int64(20)
	for i := int64(0); i <= roundToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
	}

	time.Sleep(time.Second)

	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)

	checkSCBalance(t, ownerNode, scAddress, receiverAddress1, finalBalance)
	checkSCBalance(t, ownerNode, scAddress, receiverAddress1, finalBalance)

	checkPlayerBalances(t, nodes, players)

	userAcc := GetUserAccount(nodes, relayer.Address)
	assert.Equal(t, 1, userAcc.GetBalance().Cmp(relayer.Balance))
}
