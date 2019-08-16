package block

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

func TestShouldProcessBlocksInMultiShardArchitecture(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	fmt.Println("Setup nodes...")
	numOfShards := 6
	nodesPerShard := 3
	numMetachainNodes := 1

	idxProposers := []int{0, 3, 6, 9, 12, 15, 18}
	senderShard := uint32(0)
	recvShards := []uint32{1, 2}
	round := uint64(0)
	nonce := uint64(0)

	valMinting := big.NewInt(100)
	valToTransferPerTx := big.NewInt(2)
	gasPricePerTx := uint64(2)
	gasLimitPerTx := uint64(2)

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)
	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	fmt.Println("Generating private keys for senders and receivers...")
	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 0)
	txToGenerateInEachMiniBlock := 3

	proposerNode := nodes[0]

	//sender shard keys, receivers  keys
	sendersPrivateKeys := make([]crypto.PrivateKey, 3)
	receiversPrivateKeys := make(map[uint32][]crypto.PrivateKey)
	for i := 0; i < txToGenerateInEachMiniBlock; i++ {
		sendersPrivateKeys[i], _, _ = integrationTests.GenerateSkAndPkInShard(generateCoordinator, senderShard)

		//receivers in same shard with the sender
		sk, _, _ := integrationTests.GenerateSkAndPkInShard(generateCoordinator, senderShard)
		receiversPrivateKeys[senderShard] = append(receiversPrivateKeys[senderShard], sk)
		//receivers in other shards
		for _, shardId := range recvShards {
			sk, _, _ = integrationTests.GenerateSkAndPkInShard(generateCoordinator, shardId)
			receiversPrivateKeys[shardId] = append(receiversPrivateKeys[shardId], sk)
		}
	}

	fmt.Println("Generating transactions...")
	integrationTests.GenerateAndDisseminateTxs(proposerNode, sendersPrivateKeys, receiversPrivateKeys,
		valToTransferPerTx, gasPricePerTx, gasLimitPerTx)
	fmt.Println("Delaying for disseminating transactions...")
	time.Sleep(time.Second * 5)

	fmt.Println("Minting sender addresses...")
	integrationTests.CreateMintingForSenders(nodes, senderShard, sendersPrivateKeys, valMinting)

	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	roundsToWait := 6
	for i := 0; i < roundsToWait; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
	}

	fmt.Println("Test nodes from proposer shard to have the correct balances...")
	for _, n := range nodes {
		isNodeInSenderShard := n.ShardCoordinator.SelfId() == senderShard
		if !isNodeInSenderShard {
			continue
		}

		//test sender balances
		for _, sk := range sendersPrivateKeys {
			valTransferred := big.NewInt(0).Mul(valToTransferPerTx, big.NewInt(int64(len(receiversPrivateKeys))))
			valRemaining := big.NewInt(0).Sub(valMinting, valTransferred)
			integrationTests.TestPrivateKeyHasBalance(t, n, sk, valRemaining)
		}
		//test receiver balances from same shard
		for _, sk := range receiversPrivateKeys[proposerNode.ShardCoordinator.SelfId()] {
			integrationTests.TestPrivateKeyHasBalance(t, n, sk, valToTransferPerTx)
		}
	}

	fmt.Println("Test nodes from receiver shards to have the correct balances...")
	for _, n := range nodes {
		isNodeInReceiverShardAndNotProposer := false
		for _, shardId := range recvShards {
			if n.ShardCoordinator.SelfId() == shardId {
				isNodeInReceiverShardAndNotProposer = true
				break
			}
		}
		if !isNodeInReceiverShardAndNotProposer {
			continue
		}

		//test receiver balances from same shard
		for _, sk := range receiversPrivateKeys[n.ShardCoordinator.SelfId()] {
			integrationTests.TestPrivateKeyHasBalance(t, n, sk, valToTransferPerTx)
		}
	}
}
