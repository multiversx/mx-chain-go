package softfork

import (
	"encoding/hex"
	"io/ioutil"
	"math/big"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("integrationtests/singleshard/block/softfork")

func TestScDeploy(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()
	advertiserAddr := integrationTests.GetConnectableAddress(advertiser)
	builtinEnableEpoch := uint32(0)
	deployEnableEpoch := uint32(1)
	relayedTxEnableEpoch := uint32(0)
	penalizedTooMuchGasEnableEpoch := uint32(0)
	roundsPerEpoch := uint64(10)

	shardNode := integrationTests.NewTestProcessorNodeSoftFork(
		1,
		0,
		0,
		advertiserAddr,
		builtinEnableEpoch,
		deployEnableEpoch,
		relayedTxEnableEpoch,
		penalizedTooMuchGasEnableEpoch,
	)
	shardNode.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)

	metaNode := integrationTests.NewTestProcessorNodeSoftFork(
		1,
		core.MetachainShardId,
		0,
		advertiserAddr,
		builtinEnableEpoch,
		deployEnableEpoch,
		relayedTxEnableEpoch,
		penalizedTooMuchGasEnableEpoch,
	)
	metaNode.EpochStartTrigger.SetRoundsPerEpoch(roundsPerEpoch)

	nodes := []*integrationTests.TestProcessorNode{
		shardNode,
		metaNode,
	}

	idxProposers := []int{0, 1}

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for _, n := range nodes {
		_ = n.Messenger.Bootstrap()
	}

	log.Info("delaying for nodes p2p bootstrap...")
	time.Sleep(integrationTests.P2pBootstrapDelay)

	round := uint64(1)
	nonce := uint64(1)
	numRounds := roundsPerEpoch + 5

	integrationTests.CreateMintingForSenders(nodes, 0, []crypto.PrivateKey{shardNode.OwnAccount.SkTxSign}, big.NewInt(1000000000))

	accnt, _ := shardNode.AccntState.GetExistingAccount(shardNode.OwnAccount.Address)
	userAccnt := accnt.(state.UserAccountHandler)
	balance := userAccnt.GetBalance()
	log.Info("balance", "value", balance.String())

	deployedFailedAddress := deploySc(t, nodes)

	for i := uint64(0); i < numRounds; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++

		time.Sleep(time.Second)
	}

	log.Info("resulted sc address (failed)", "address", integrationTests.TestAddressPubkeyConverter.Encode(deployedFailedAddress))
	assert.False(t, scAccountExists(shardNode, deployedFailedAddress))

	deploySucceeded := deploySc(t, nodes)
	for i := uint64(0); i < 5; i++ {
		integrationTests.UpdateRound(nodes, round)
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++

		time.Sleep(time.Second)
	}

	log.Info("resulted sc address (success)", "address", integrationTests.TestAddressPubkeyConverter.Encode(deploySucceeded))
	assert.True(t, scAccountExists(shardNode, deploySucceeded))
}

func deploySc(t *testing.T, nodes []*integrationTests.TestProcessorNode) []byte {
	scCode, err := ioutil.ReadFile("./testdata/answer.wasm")
	require.Nil(t, err)

	node := nodes[0]
	scAddress, err := node.BlockchainHook.NewAddress(node.OwnAccount.Address, node.OwnAccount.Nonce, factory.ArwenVirtualMachine)
	require.Nil(t, err)

	integrationTests.DeployScTx(nodes, 0, hex.EncodeToString(scCode), factory.ArwenVirtualMachine, "001000000000")

	return scAddress
}

func scAccountExists(node *integrationTests.TestProcessorNode, address []byte) bool {
	accnt, _ := node.AccntState.GetExistingAccount(address)

	return !check.IfNil(accnt)
}
