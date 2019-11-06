package systemVM

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/vm/factory"
	"github.com/stretchr/testify/assert"
)

func TestStakingUnstakingAndUnboundingOnMultiShardEnvironment(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 2
	nodesPerShard := 3
	numMetachainNodes := 3

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	logger.DefaultLogger().SetLevel("DEBUG")

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	idxProposers := make([]int, numOfShards+1)
	for i := 0; i < numOfShards; i++ {
		idxProposers[i] = i * nodesPerShard
	}
	idxProposers[numOfShards] = numOfShards * nodesPerShard

	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	initialVal := big.NewInt(10000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	// verify initial values
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(node.OwnAccount.Address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				sndAcc := getAccountFromAddrBytes(helperNode.AccntState, node.OwnAccount.Address.Bytes())
				assert.Equal(t, initialVal, sndAcc.Balance)
				break
			}
		}
	}

	for _, node := range nodes {
		roothash, _ := node.AccntState.RootHash()
		fmt.Printf("shardID: %d roothash: %s \n", node.ShardCoordinator.SelfId(), hex.EncodeToString(roothash))
	}

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	///////////------- send stake tx and check sender's balance
	for _, node := range nodes {
		pubKey, _ := node.NodeKeys.Pk.ToByteArray()
		txData := "stake" + "@" + hex.EncodeToString(pubKey)
		integrationTests.CreateAndSendTransaction(node, node.EconomicsData.StakeValue(), factory.StakingSCAddress, txData)
	}

	time.Sleep(time.Second)

	nrRoundsToPropagateMultiShard := 10
	for i := 0; i < nrRoundsToPropagateMultiShard; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	time.Sleep(time.Second)

	// verify if staking was done - value taken out of accounts
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(node.OwnAccount.Address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				sndAcc := getAccountFromAddrBytes(helperNode.AccntState, node.OwnAccount.Address.Bytes())
				assert.Equal(t, big.NewInt(0).Sub(initialVal, node.EconomicsData.StakeValue()), sndAcc.Balance)
				break
			}
		}
	}

	/////////------ send unStake tx
	for _, node := range nodes {
		txData := "unStake"
		integrationTests.CreateAndSendTransaction(node, big.NewInt(0), factory.StakingSCAddress, txData)
	}

	time.Sleep(time.Second)

	for i := 0; i < nrRoundsToPropagateMultiShard; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	/////////----- wait for unbound period
	for i := uint64(0); i < nodes[0].EconomicsData.UnBoundPeriod(); i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	////////----- send unBound
	for _, node := range nodes {
		txData := "unBound"
		integrationTests.CreateAndSendTransaction(node, big.NewInt(0), factory.StakingSCAddress, txData)
	}

	time.Sleep(time.Second)

	for i := 0; i < nrRoundsToPropagateMultiShard; i++ {
		integrationTests.ProposeBlock(nodes, idxProposers, round, nonce)
		integrationTests.SyncBlock(t, nodes, idxProposers, round)
		round = integrationTests.IncrementAndPrintRound(round)
		nonce++
	}

	// verify if unbound is done - staking value back to sender
	for _, node := range nodes {
		accShardId := node.ShardCoordinator.ComputeId(node.OwnAccount.Address)

		for _, helperNode := range nodes {
			if helperNode.ShardCoordinator.SelfId() == accShardId {
				sndAcc := getAccountFromAddrBytes(helperNode.AccntState, node.OwnAccount.Address.Bytes())
				assert.Equal(t, initialVal, sndAcc.Balance)
				break
			}
		}
	}
}

func getAccountFromAddrBytes(accState state.AccountsAdapter, address []byte) *state.Account {
	addrCont, _ := integrationTests.TestAddressConverter.CreateAddressFromPublicKeyBytes(address)
	sndrAcc, _ := accState.GetExistingAccount(addrCont)

	sndAccSt, _ := sndrAcc.(*state.Account)

	return sndAccSt
}
