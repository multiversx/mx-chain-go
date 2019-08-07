package block

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestShouldProcessBlocksInMultiShardArchitecture(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	fmt.Println("Step 1. Setup nodes...")
	numOfShards := 6
	nodesPerShard := 3

	senderShard := uint32(0)
	recvShards := []uint32{1, 2}

	valMinting := big.NewInt(100)
	valToTransferPerTx := big.NewInt(2)

	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		integrationTests.GetConnectableAddress(advertiser),
	)
	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Node.Stop()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	fmt.Println("Step 2. Generating private keys for senders and receivers...")
	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 0)
	txToGenerateInEachMiniBlock := 3

	proposerNode := nodes[0]

	//sender shard keys, receivers  keys
	sendersPrivateKeys := make([]crypto.PrivateKey, 3)
	receiversPrivateKeys := make(map[uint32][]crypto.PrivateKey)
	for i := 0; i < txToGenerateInEachMiniBlock; i++ {
		sendersPrivateKeys[i] = integrationTests.GeneratePrivateKeyInShardId(generateCoordinator, senderShard)

		//receivers in same shard with the sender
		sk := integrationTests.GeneratePrivateKeyInShardId(generateCoordinator, senderShard)
		receiversPrivateKeys[senderShard] = append(receiversPrivateKeys[senderShard], sk)
		//receivers in other shards
		for _, shardId := range recvShards {
			sk = integrationTests.GeneratePrivateKeyInShardId(generateCoordinator, shardId)
			receiversPrivateKeys[shardId] = append(receiversPrivateKeys[shardId], sk)
		}
	}

	fmt.Println("Step 3. Generating transactions...")
	integrationTests.GenerateAndDisseminateTxs(proposerNode, sendersPrivateKeys, receiversPrivateKeys, valToTransferPerTx)
	fmt.Println("Delaying for disseminating transactions...")
	time.Sleep(time.Second * 5)

	fmt.Println("Step 4. Minting sender addresses...")
	integrationTests.CreateMintingForSenders(nodes, senderShard, sendersPrivateKeys, valMinting)

	fmt.Println("Step 5. Proposer creates block body and header with all available transactions...")
	blockBody, blockHeader := integrationTests.CreateBlockBodyAndHeader(t, proposerNode, uint64(1), proposerNode.ShardCoordinator)
	_ = proposerNode.BroadcastMessenger.BroadcastBlock(blockBody, blockHeader)
	_ = proposerNode.BroadcastMessenger.BroadcastHeader(blockHeader)
	miniBlocks, transactions, _ := proposerNode.BlockProcessor.MarshalizedDataToBroadcast(blockHeader, blockBody)
	_ = proposerNode.BroadcastMessenger.BroadcastMiniBlocks(miniBlocks)
	_ = proposerNode.BroadcastMessenger.BroadcastTransactions(transactions)
	_ = proposerNode.BlockProcessor.CommitBlock(proposerNode.BlockChain, blockHeader, blockBody)
	fmt.Println("Delaying for disseminating miniblocks and header...")
	time.Sleep(time.Second * 5)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	blockBody, blockHeader = integrationTests.CreateBlockBodyAndHeader(t, proposerNode, uint64(2), proposerNode.ShardCoordinator)
	_ = proposerNode.BroadcastMessenger.BroadcastBlock(blockBody, blockHeader)
	_ = proposerNode.BroadcastMessenger.BroadcastHeader(blockHeader)
	miniBlocks, transactions, _ = proposerNode.BlockProcessor.MarshalizedDataToBroadcast(blockHeader, blockBody)
	_ = proposerNode.BroadcastMessenger.BroadcastMiniBlocks(miniBlocks)
	_ = proposerNode.BroadcastMessenger.BroadcastTransactions(transactions)
	_ = proposerNode.BlockProcessor.CommitBlock(proposerNode.BlockChain, blockHeader, blockBody)
	fmt.Println("Delaying for disseminating miniblocks and header...")
	time.Sleep(time.Second * 5)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	fmt.Println("Step 7. Nodes from proposer's shard will have to successfully process the block sent by the proposer...")
	fmt.Println(integrationTests.MakeDisplayTable(nodes))
	for _, n := range nodes {
		isNodeInSenderShardAndNotProposer := n.ShardCoordinator.SelfId() == senderShard && n != proposerNode
		if isNodeInSenderShardAndNotProposer {
			n.BlockChain.SetGenesisHeaderHash(n.Headers[0].GetPrevHash())
			err := n.BlockProcessor.ProcessBlock(
				n.BlockChain,
				n.Headers[0],
				block.Body(n.MiniBlocks),
				func() time.Duration {
					//fair enough to process a few transactions
					return time.Second * 2
				},
			)

			assert.Nil(t, err)
		}
	}

	fmt.Println("Step 7. Metachain processes the received header...")
	metaNode := nodes[len(nodes)-1]
	_, metaHeader := integrationTests.CreateBlockBodyAndHeader(t, metaNode, uint64(1), metaNode.ShardCoordinator)
	_ = metaNode.BroadcastMessenger.BroadcastBlock(nil, metaHeader)
	_ = metaNode.BlockProcessor.CommitBlock(metaNode.BlockChain, metaHeader, &block.MetaBlockBody{})
	fmt.Println("Delaying for disseminating meta header...")
	time.Sleep(time.Second * 5)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	_, metaHeader = integrationTests.CreateBlockBodyAndHeader(t, metaNode, uint64(2), metaNode.ShardCoordinator)
	_ = metaNode.BroadcastMessenger.BroadcastBlock(nil, metaHeader)
	_ = metaNode.BlockProcessor.CommitBlock(metaNode.BlockChain, metaHeader, &block.MetaBlockBody{})
	fmt.Println("Delaying for disseminating meta header...")
	time.Sleep(time.Second * 5)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	_, metaHeader = integrationTests.CreateBlockBodyAndHeader(t, metaNode, uint64(3), metaNode.ShardCoordinator)
	_ = metaNode.BroadcastMessenger.BroadcastBlock(nil, metaHeader)
	_ = metaNode.BlockProcessor.CommitBlock(metaNode.BlockChain, metaHeader, &block.MetaBlockBody{})
	fmt.Println("Delaying for disseminating meta header...")
	time.Sleep(time.Second * 5)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	_, metaHeader = integrationTests.CreateBlockBodyAndHeader(t, metaNode, uint64(3), metaNode.ShardCoordinator)
	_ = metaNode.BroadcastMessenger.BroadcastBlock(nil, metaHeader)
	_ = metaNode.BlockProcessor.CommitBlock(metaNode.BlockChain, metaHeader, &block.MetaBlockBody{})
	fmt.Println("Delaying for disseminating meta header...")
	time.Sleep(time.Second * 5)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	fmt.Println("Step 8. Test nodes from proposer shard to have the correct balances...")
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

	fmt.Println("Step 9. First nodes from receiver shards assemble header/body blocks and broadcast them...")
	firstReceiverNodes := make([]*integrationTests.TestProcessorNode, 0)
	//get first nodes from receiver shards
	for _, shardId := range recvShards {
		receiverProposer := nodes[int(shardId)*nodesPerShard]
		firstReceiverNodes = append(firstReceiverNodes, receiverProposer)

		body, header := integrationTests.CreateBlockBodyAndHeader(t, receiverProposer, uint64(1), receiverProposer.ShardCoordinator)
		_ = receiverProposer.BroadcastMessenger.BroadcastBlock(body, header)
		_ = receiverProposer.BroadcastMessenger.BroadcastHeader(header)
		miniBlocks, transactions, _ := proposerNode.BlockProcessor.MarshalizedDataToBroadcast(header, body)
		_ = receiverProposer.BroadcastMessenger.BroadcastMiniBlocks(miniBlocks)
		_ = receiverProposer.BroadcastMessenger.BroadcastTransactions(transactions)
		_ = receiverProposer.BlockProcessor.CommitBlock(receiverProposer.BlockChain, header, body)
	}
	fmt.Println("Delaying for disseminating miniblocks and headers...")
	time.Sleep(time.Second * 5)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	for _, shardId := range recvShards {
		receiverProposer := nodes[int(shardId)*nodesPerShard]

		body, header := integrationTests.CreateBlockBodyAndHeader(t, receiverProposer, uint64(2), receiverProposer.ShardCoordinator)
		_ = receiverProposer.BroadcastMessenger.BroadcastBlock(body, header)
		_ = receiverProposer.BroadcastMessenger.BroadcastHeader(header)
		miniBlocks, transactions, _ := proposerNode.BlockProcessor.MarshalizedDataToBroadcast(header, body)
		_ = receiverProposer.BroadcastMessenger.BroadcastMiniBlocks(miniBlocks)
		_ = receiverProposer.BroadcastMessenger.BroadcastTransactions(transactions)
		_ = receiverProposer.BlockProcessor.CommitBlock(receiverProposer.BlockChain, header, body)
	}
	fmt.Println("Delaying for disseminating miniblocks and headers...")
	time.Sleep(time.Second * 5)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	for _, shardId := range recvShards {
		receiverProposer := nodes[int(shardId)*nodesPerShard]

		body, header := integrationTests.CreateBlockBodyAndHeader(t, receiverProposer, uint64(3), receiverProposer.ShardCoordinator)
		_ = receiverProposer.BroadcastMessenger.BroadcastBlock(body, header)
		_ = receiverProposer.BroadcastMessenger.BroadcastHeader(header)
		miniBlocks, transactions, _ := proposerNode.BlockProcessor.MarshalizedDataToBroadcast(header, body)
		_ = receiverProposer.BroadcastMessenger.BroadcastMiniBlocks(miniBlocks)
		_ = receiverProposer.BroadcastMessenger.BroadcastTransactions(transactions)
		_ = receiverProposer.BlockProcessor.CommitBlock(receiverProposer.BlockChain, header, body)
	}
	fmt.Println("Delaying for disseminating miniblocks and headers...")
	time.Sleep(time.Second * 5)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	for _, shardId := range recvShards {
		receiverProposer := nodes[int(shardId)*nodesPerShard]

		body, header := integrationTests.CreateBlockBodyAndHeader(t, receiverProposer, uint64(4), receiverProposer.ShardCoordinator)
		_ = receiverProposer.BroadcastMessenger.BroadcastBlock(body, header)
		_ = receiverProposer.BroadcastMessenger.BroadcastHeader(header)
		miniBlocks, transactions, _ := proposerNode.BlockProcessor.MarshalizedDataToBroadcast(header, body)
		_ = receiverProposer.BroadcastMessenger.BroadcastMiniBlocks(miniBlocks)
		_ = receiverProposer.BroadcastMessenger.BroadcastTransactions(transactions)
		_ = receiverProposer.BlockProcessor.CommitBlock(receiverProposer.BlockChain, header, body)
	}
	fmt.Println("Delaying for disseminating miniblocks and headers...")
	time.Sleep(time.Second * 5)
	fmt.Println(integrationTests.MakeDisplayTable(nodes))

	fmt.Println("Step 10. NodesSetup from receivers shards will have to successfully process the block sent by their proposer...")
	fmt.Println(integrationTests.MakeDisplayTable(nodes))
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() == sharding.MetachainShardId {
			continue
		}

		isNodeInReceiverShardAndNotProposer := false
		for _, shardId := range recvShards {
			if n.ShardCoordinator.SelfId() == shardId {
				isNodeInReceiverShardAndNotProposer = true
				break
			}
		}
		for _, proposerReceiver := range firstReceiverNodes {
			if proposerReceiver == n {
				isNodeInReceiverShardAndNotProposer = false
			}
		}

		if isNodeInReceiverShardAndNotProposer {
			if len(n.Headers) > 0 {
				n.BlockChain.SetGenesisHeaderHash(n.Headers[0].GetPrevHash())
				err := n.BlockProcessor.ProcessBlock(
					n.BlockChain,
					n.Headers[0],
					block.Body{},
					func() time.Duration {
						// time 5 seconds as they have to request from leader the TXs
						return time.Second * 5
					},
				)

				assert.Nil(t, err)
				if err != nil {
					return
				}

				err = n.BlockProcessor.CommitBlock(n.BlockChain, n.Headers[0], block.Body{})
				assert.Nil(t, err)
				if err != nil {
					return
				}

				err = n.BlockProcessor.ProcessBlock(
					n.BlockChain,
					n.Headers[1],
					block.Body{},
					func() time.Duration {
						// time 5 seconds as they have to request from leader the TXs
						return time.Second * 5
					},
				)

				assert.Nil(t, err)
				if err != nil {
					return
				}

				err = n.BlockProcessor.CommitBlock(n.BlockChain, n.Headers[1], block.Body{})
				assert.Nil(t, err)
				if err != nil {
					return
				}

				err = n.BlockProcessor.ProcessBlock(
					n.BlockChain,
					n.Headers[2],
					block.Body(n.MiniBlocks),
					func() time.Duration {
						// time 5 seconds as they have to request from leader the TXs
						return time.Second * 5
					},
				)

				assert.Nil(t, err)
				if err != nil {
					return
				}

				err = n.BlockProcessor.CommitBlock(n.BlockChain, n.Headers[2], block.Body(n.MiniBlocks))
				assert.Nil(t, err)
				if err != nil {
					return
				}

				err = n.BlockProcessor.ProcessBlock(
					n.BlockChain,
					n.Headers[3],
					block.Body{},
					func() time.Duration {
						// time 5 seconds as they have to request from leader the TXs
						return time.Second * 5
					},
				)

				assert.Nil(t, err)
				if err != nil {
					return
				}

				err = n.BlockProcessor.CommitBlock(n.BlockChain, n.Headers[3], block.Body{})
				assert.Nil(t, err)
				if err != nil {
					return
				}
			}
		}
	}

	fmt.Println("Step 11. Test nodes from receiver shards to have the correct balances...")
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
