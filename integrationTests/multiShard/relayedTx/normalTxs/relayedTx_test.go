package normalTxs

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/multiShard/relayedTx"
	"github.com/stretchr/testify/assert"
)

func TestRelayedTransactionInMultiShardEnvironmentWithNormalTx(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	//TODO - remove next lines
	logger.SetLogLevel("*:DEBUG")

	f, err := core.CreateFile(
		core.ArgCreateFileArgument{
			Directory:     "",
			Prefix:        "log",
			FileExtension: "log",
		})
	if err != nil {
		panic(err)
	}
	logger.AddLogObserver(f, &logger.PlainFormatter{})
	defer f.Close()

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

	nrRoundsToTest := int64(5)
	for i := int64(0); i < nrRoundsToTest; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)

		displayBalances(nodes, relayer, players, receiverAddress1, receiverAddress2)

		for _, player := range players {
			_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress1, sendValue, integrationTests.MinTxGasLimit, []byte(""))
			_ = relayedTx.CreateAndSendRelayedAndUserTx(nodes, relayer, player, receiverAddress2, sendValue, integrationTests.MinTxGasLimit, []byte(""))
		}

		time.Sleep(time.Second)
	}

	roundToPropagateMultiShard := int64(20)
	for i := int64(0); i <= roundToPropagateMultiShard; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
		integrationTests.AddSelfNotarizedHeaderByMetachain(nodes)
		displayBalances(nodes, relayer, players, receiverAddress1, receiverAddress2)
	}

	time.Sleep(time.Second)
	receiver1 := relayedTx.GetUserAccount(nodes, receiverAddress1)
	receiver2 := relayedTx.GetUserAccount(nodes, receiverAddress2)

	finalBalance := big.NewInt(0).Mul(big.NewInt(int64(len(players))), big.NewInt(nrRoundsToTest))
	finalBalance.Mul(finalBalance, sendValue)
	assert.Equal(t, finalBalance.String(), receiver1.GetBalance().String())
	assert.Equal(t, finalBalance.String(), receiver2.GetBalance().String())

	players = append(players, relayer)
	relayedTx.CheckPlayerBalances(t, nodes, players)
}

//TODO remove this
func displayBalances(
	nodes []*integrationTests.TestProcessorNode,
	relayer *integrationTests.TestWalletAccount,
	players []*integrationTests.TestWalletAccount,
	receiver1 []byte,
	receiver2 []byte,
) {
	hdr := []string{"account", "values"}
	lines := make([]*display.LineData, 0)

	relayerAccnt := relayedTx.GetUserAccount(nodes, relayer.Address)
	lines = append(lines, display.NewLineData(false, []string{"relayer", fmt.Sprintf("nonce: %d", relayerAccnt.GetNonce())}))
	lines = append(lines, display.NewLineData(true, []string{"relayer", fmt.Sprintf("balance: %s", relayerAccnt.GetBalance().String())}))

	for i, p := range players {
		playerAccnt := relayedTx.GetUserAccount(nodes, p.Address)
		nonce := uint64(0)
		balance := big.NewInt(0)
		if playerAccnt != nil {
			nonce = playerAccnt.GetNonce()
			balance = playerAccnt.GetBalance()
		}

		lines = append(lines, display.NewLineData(false, []string{fmt.Sprintf("player %d", i+1), fmt.Sprintf("nonce: %d", nonce)}))
		lines = append(lines, display.NewLineData(true, []string{fmt.Sprintf("player %d", i+1), fmt.Sprintf("balance: %s", balance.String())}))
	}

	rcv1 := relayedTx.GetUserAccount(nodes, receiver1)
	bal := big.NewInt(0)
	if rcv1 != nil {
		bal = rcv1.GetBalance()
	}
	lines = append(lines, display.NewLineData(true, []string{"receiver 1", fmt.Sprintf("balance: %s", bal)}))

	rcv2 := relayedTx.GetUserAccount(nodes, receiver2)
	if rcv2 != nil {
		bal = rcv2.GetBalance()
	}
	lines = append(lines, display.NewLineData(false, []string{"receiver 2", fmt.Sprintf("balance: %s", bal)}))

	tbl, _ := display.CreateTableString(hdr, lines)

	log := logger.GetOrCreate("integration")
	log.Info("\n" + tbl)
}
