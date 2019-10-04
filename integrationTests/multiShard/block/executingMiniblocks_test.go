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
	receiversPublicKeys := make(map[uint32][]crypto.PublicKey)
	for i := 0; i < txToGenerateInEachMiniBlock; i++ {
		sendersPrivateKeys[i], _, _ = integrationTests.GenerateSkAndPkInShard(generateCoordinator, senderShard)
		//receivers in same shard with the sender
		_, pk, _ := integrationTests.GenerateSkAndPkInShard(generateCoordinator, senderShard)
		receiversPublicKeys[senderShard] = append(receiversPublicKeys[senderShard], pk)
		//receivers in other shards
		for _, shardId := range recvShards {
			_, pk, _ = integrationTests.GenerateSkAndPkInShard(generateCoordinator, shardId)
			receiversPublicKeys[shardId] = append(receiversPublicKeys[shardId], pk)
		}
	}

	fmt.Println("Minting sender addresses...")
	integrationTests.CreateMintingForSenders(nodes, senderShard, sendersPrivateKeys, valMinting)

	fmt.Println("Generating transactions...")
	integrationTests.GenerateAndDisseminateTxs(
		proposerNode,
		sendersPrivateKeys,
		receiversPublicKeys,
		valToTransferPerTx,
		integrationTests.MinTxGasPrice,
		integrationTests.MinTxGasLimit,
	)
	fmt.Println("Delaying for disseminating transactions...")
	time.Sleep(time.Second * 5)

	round = integrationTests.IncrementAndPrintRound(round)
	nonce++
	roundsToWait := 6
	for i := 0; i < roundsToWait; i++ {
		round, nonce = integrationTests.ProposeAndSyncOneBlock(t, nodes, idxProposers, round, nonce)
	}

	gasPricePerTxBigInt := big.NewInt(0).SetUint64(integrationTests.MinTxGasPrice)
	gasLimitPerTxBigInt := big.NewInt(0).SetUint64(integrationTests.MinTxGasLimit)
	gasValue := big.NewInt(0).Mul(gasPricePerTxBigInt, gasLimitPerTxBigInt)
	totalValuePerTx := big.NewInt(0).Add(gasValue, valToTransferPerTx)
	fmt.Println("Test nodes from proposer shard to have the correct balances...")
	for _, n := range nodes {
		isNodeInSenderShard := n.ShardCoordinator.SelfId() == senderShard
		if !isNodeInSenderShard {
			continue
		}

		//test sender balances
		for _, sk := range sendersPrivateKeys {
			valTransferred := big.NewInt(0).Mul(totalValuePerTx, big.NewInt(int64(len(receiversPublicKeys))))
			valRemaining := big.NewInt(0).Sub(valMinting, valTransferred)
			integrationTests.TestPrivateKeyHasBalance(t, n, sk, valRemaining)
		}
		//test receiver balances from same shard
		for _, pk := range receiversPublicKeys[proposerNode.ShardCoordinator.SelfId()] {
			integrationTests.TestPublicKeyHasBalance(t, n, pk, valToTransferPerTx)
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
		for _, pk := range receiversPublicKeys[n.ShardCoordinator.SelfId()] {
			integrationTests.TestPublicKeyHasBalance(t, n, pk, valToTransferPerTx)
		}
	}
}
