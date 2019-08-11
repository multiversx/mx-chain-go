package block

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node"
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

	advertiser := createMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

	nodes := createNodes(
		numOfShards,
		nodesPerShard,
		getConnectableAddress(advertiser),
	)
	displayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.node.Stop()
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
	receiversShardsID := make([]uint32, 0)
	receiversShardsID = append(receiversShardsID, senderShard)
	receiversShardsID = append(receiversShardsID, recvShards...)
	for i := 0; i < txToGenerateInEachMiniBlock; i++ {
		sendersPrivateKeys[i] = generatePrivateKeyInShardId(generateCoordinator, senderShard)

		//receivers in same shard with the sender
		sk := generatePrivateKeyInShardId(generateCoordinator, senderShard)
		receiversPrivateKeys[senderShard] = append(receiversPrivateKeys[senderShard], sk)
		//receivers in other shards
		for _, shardId := range recvShards {
			sk = generatePrivateKeyInShardId(generateCoordinator, shardId)
			receiversPrivateKeys[shardId] = append(receiversPrivateKeys[shardId], sk)
		}
	}

	fmt.Println("Step 3. Generating transactions...")
	generateAndDisseminateTxs(proposerNode.node, sendersPrivateKeys, receiversPrivateKeys, receiversShardsID, valToTransferPerTx)
	fmt.Println("Delaying for disseminating transactions...")
	time.Sleep(time.Second * 5)

	fmt.Println(makeDisplayTable(nodes))

	fmt.Println("Step 4. Minting sender addresses...")
	createMintingForSenders(nodes, senderShard, sendersPrivateKeys, valMinting)

	fmt.Println("Step 5. Proposer creates block body and header with all available transactions...")
	blockBody, blockHeader := proposeBlock(t, proposerNode, uint64(1))
	_ = proposerNode.broadcastMessenger.BroadcastBlock(blockBody, blockHeader)
	_ = proposerNode.broadcastMessenger.BroadcastHeader(blockHeader)
	miniBlocks, transactions, _ := proposerNode.blkProcessor.MarshalizedDataToBroadcast(blockHeader, blockBody)
	_ = proposerNode.broadcastMessenger.BroadcastMiniBlocks(miniBlocks)
	_ = proposerNode.broadcastMessenger.BroadcastTransactions(transactions)
	_ = proposerNode.blkProcessor.CommitBlock(proposerNode.blkc, blockHeader, blockBody)
	fmt.Println("Delaying for disseminating miniblocks and header...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	blockBody, blockHeader = proposeBlock(t, proposerNode, uint64(2))
	_ = proposerNode.broadcastMessenger.BroadcastBlock(blockBody, blockHeader)
	_ = proposerNode.broadcastMessenger.BroadcastHeader(blockHeader)
	miniBlocks, transactions, _ = proposerNode.blkProcessor.MarshalizedDataToBroadcast(blockHeader, blockBody)
	_ = proposerNode.broadcastMessenger.BroadcastMiniBlocks(miniBlocks)
	_ = proposerNode.broadcastMessenger.BroadcastTransactions(transactions)
	_ = proposerNode.blkProcessor.CommitBlock(proposerNode.blkc, blockHeader, blockBody)
	fmt.Println("Delaying for disseminating miniblocks and header...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	fmt.Println("Step 7. Nodes from proposer's shard will have to successfully process the block sent by the proposer...")
	fmt.Println(makeDisplayTable(nodes))
	for _, n := range nodes {
		isNodeInSenderShardAndNotProposer := n.shardId == senderShard && n != proposerNode
		if isNodeInSenderShardAndNotProposer {
			n.blkc.SetGenesisHeaderHash(n.headers[0].GetPrevHash())
			err := n.blkProcessor.ProcessBlock(
				n.blkc,
				n.headers[0],
				getBlockBody(testMarshalizer, testHasher, n.headers[0], n.miniblocks),
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
	_, metaHeader := proposeMetaBlock(t, metaNode, uint64(1))
	_ = metaNode.broadcastMessenger.BroadcastBlock(nil, metaHeader)
	_ = metaNode.blkProcessor.CommitBlock(metaNode.blkc, metaHeader, &block.MetaBlockBody{})
	fmt.Println("Delaying for disseminating meta header...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	_, metaHeader = proposeMetaBlock(t, metaNode, uint64(2))
	_ = metaNode.broadcastMessenger.BroadcastBlock(nil, metaHeader)
	_ = metaNode.blkProcessor.CommitBlock(metaNode.blkc, metaHeader, &block.MetaBlockBody{})
	fmt.Println("Delaying for disseminating meta header...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	_, metaHeader = proposeMetaBlock(t, metaNode, uint64(3))
	_ = metaNode.broadcastMessenger.BroadcastBlock(nil, metaHeader)
	_ = metaNode.blkProcessor.CommitBlock(metaNode.blkc, metaHeader, &block.MetaBlockBody{})
	fmt.Println("Delaying for disseminating meta header...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	_, metaHeader = proposeMetaBlock(t, metaNode, uint64(3))
	_ = metaNode.broadcastMessenger.BroadcastBlock(nil, metaHeader)
	_ = metaNode.blkProcessor.CommitBlock(metaNode.blkc, metaHeader, &block.MetaBlockBody{})
	fmt.Println("Delaying for disseminating meta header...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	fmt.Println("Step 8. Test nodes from proposer shard to have the correct balances...")
	for _, n := range nodes {
		isNodeInSenderShard := n.shardId == senderShard
		if !isNodeInSenderShard {
			continue
		}

		//test sender balances
		for _, sk := range sendersPrivateKeys {
			valTransferred := big.NewInt(0).Mul(valToTransferPerTx, big.NewInt(int64(len(receiversPrivateKeys))))
			valRemaining := big.NewInt(0).Sub(valMinting, valTransferred)
			testPrivateKeyHasBalance(t, n, sk, valRemaining)
		}
		//test receiver balances from same shard
		for _, sk := range receiversPrivateKeys[proposerNode.shardId] {
			testPrivateKeyHasBalance(t, n, sk, valToTransferPerTx)
		}
	}

	fmt.Println("Step 9. First nodes from receiver shards assemble header/body blocks and broadcast them...")
	firstReceiverNodes := make([]*testNode, 0)
	//get first nodes from receiver shards
	for _, shardId := range recvShards {
		receiverProposer := nodes[int(shardId)*nodesPerShard]
		firstReceiverNodes = append(firstReceiverNodes, receiverProposer)

		body, header := proposeBlock(t, receiverProposer, uint64(1))
		_ = receiverProposer.broadcastMessenger.BroadcastBlock(body, header)
		_ = receiverProposer.broadcastMessenger.BroadcastHeader(header)
		miniBlocks, transactions, _ := proposerNode.blkProcessor.MarshalizedDataToBroadcast(header, body)
		_ = receiverProposer.broadcastMessenger.BroadcastMiniBlocks(miniBlocks)
		_ = receiverProposer.broadcastMessenger.BroadcastTransactions(transactions)
		_ = receiverProposer.blkProcessor.CommitBlock(receiverProposer.blkc, header, body)
	}
	fmt.Println("Delaying for disseminating miniblocks and headers...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	for _, shardId := range recvShards {
		receiverProposer := nodes[int(shardId)*nodesPerShard]

		body, header := proposeBlock(t, receiverProposer, uint64(2))
		_ = receiverProposer.broadcastMessenger.BroadcastBlock(body, header)
		_ = receiverProposer.broadcastMessenger.BroadcastHeader(header)
		miniBlocks, transactions, _ := proposerNode.blkProcessor.MarshalizedDataToBroadcast(header, body)
		_ = receiverProposer.broadcastMessenger.BroadcastMiniBlocks(miniBlocks)
		_ = receiverProposer.broadcastMessenger.BroadcastTransactions(transactions)
		_ = receiverProposer.blkProcessor.CommitBlock(receiverProposer.blkc, header, body)
	}
	fmt.Println("Delaying for disseminating miniblocks and headers...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	for _, shardId := range recvShards {
		receiverProposer := nodes[int(shardId)*nodesPerShard]

		body, header := proposeBlock(t, receiverProposer, uint64(3))
		_ = receiverProposer.broadcastMessenger.BroadcastBlock(body, header)
		_ = receiverProposer.broadcastMessenger.BroadcastHeader(header)
		miniBlocks, transactions, _ := proposerNode.blkProcessor.MarshalizedDataToBroadcast(header, body)
		_ = receiverProposer.broadcastMessenger.BroadcastMiniBlocks(miniBlocks)
		_ = receiverProposer.broadcastMessenger.BroadcastTransactions(transactions)
		_ = receiverProposer.blkProcessor.CommitBlock(receiverProposer.blkc, header, body)
	}
	fmt.Println("Delaying for disseminating miniblocks and headers...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	for _, shardId := range recvShards {
		receiverProposer := nodes[int(shardId)*nodesPerShard]

		body, header := proposeBlock(t, receiverProposer, uint64(4))
		_ = receiverProposer.broadcastMessenger.BroadcastBlock(body, header)
		_ = receiverProposer.broadcastMessenger.BroadcastHeader(header)
		miniBlocks, transactions, _ := proposerNode.blkProcessor.MarshalizedDataToBroadcast(header, body)
		_ = receiverProposer.broadcastMessenger.BroadcastMiniBlocks(miniBlocks)
		_ = receiverProposer.broadcastMessenger.BroadcastTransactions(transactions)
		_ = receiverProposer.blkProcessor.CommitBlock(receiverProposer.blkc, header, body)
	}
	fmt.Println("Delaying for disseminating miniblocks and headers...")
	time.Sleep(time.Second * 5)
	fmt.Println(makeDisplayTable(nodes))

	fmt.Println("Step 10. NodesSetup from receivers shards will have to successfully process the block sent by their proposer...")
	fmt.Println(makeDisplayTable(nodes))
	for _, n := range nodes {
		if n.shardId == sharding.MetachainShardId {
			continue
		}

		isNodeInReceiverShardAndNotProposer := false
		for _, shardId := range recvShards {
			if n.shardId == shardId {
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
			if len(n.headers) > 0 {
				n.blkc.SetGenesisHeaderHash(n.headers[0].GetPrevHash())
				err := n.blkProcessor.ProcessBlock(
					n.blkc,
					n.headers[0],
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

				err = n.blkProcessor.CommitBlock(n.blkc, n.headers[0], block.Body{})
				assert.Nil(t, err)
				if err != nil {
					return
				}

				err = n.blkProcessor.ProcessBlock(
					n.blkc,
					n.headers[1],
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

				err = n.blkProcessor.CommitBlock(n.blkc, n.headers[1], block.Body{})
				assert.Nil(t, err)
				if err != nil {
					return
				}

				err = n.blkProcessor.ProcessBlock(
					n.blkc,
					n.headers[2],
					block.Body(n.miniblocks),
					func() time.Duration {
						// time 5 seconds as they have to request from leader the TXs
						return time.Second * 5
					},
				)

				assert.Nil(t, err)
				if err != nil {
					return
				}

				err = n.blkProcessor.CommitBlock(n.blkc, n.headers[2], block.Body(n.miniblocks))
				assert.Nil(t, err)
				if err != nil {
					return
				}

				err = n.blkProcessor.ProcessBlock(
					n.blkc,
					n.headers[3],
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

				err = n.blkProcessor.CommitBlock(n.blkc, n.headers[3], block.Body{})
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
			if n.shardId == shardId {
				isNodeInReceiverShardAndNotProposer = true
				break
			}
		}
		if !isNodeInReceiverShardAndNotProposer {
			continue
		}

		//test receiver balances from same shard
		for _, sk := range receiversPrivateKeys[n.shardId] {
			testPrivateKeyHasBalance(t, n, sk, valToTransferPerTx)
		}
	}

}

func getBlockBody(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	headerHandler data.HeaderHandler,
	miniblocks []*block.MiniBlock,
) block.Body {

	body := block.Body{}

	header := headerHandler.(*block.Header)
	for _, miniBlockHeader := range header.MiniBlockHeaders {
		miniBlockHash := miniBlockHeader.Hash

		for _, miniblock := range miniblocks {
			mbHash, _ := core.CalculateHash(marshalizer, hasher, miniblock)

			if bytes.Equal(mbHash, miniBlockHash) {
				body = append(body, miniblock)
				break
			}
		}
	}

	return body
}

func generateAndDisseminateTxs(
	n *node.Node,
	senders []crypto.PrivateKey,
	receiversPrivateKeys map[uint32][]crypto.PrivateKey,
	shardsIDs []uint32,
	valToTransfer *big.Int,
) {

	for i := 0; i < len(senders); i++ {
		senderKey := senders[i]
		incrementalNonce := uint64(0)
		for _, shardId := range shardsIDs {
			receiverKey := receiversPrivateKeys[shardId][i]
			tx := generateTransferTx(incrementalNonce, senderKey, receiverKey, valToTransfer)
			_, _ = n.SendTransaction(
				tx.Nonce,
				hex.EncodeToString(tx.SndAddr),
				hex.EncodeToString(tx.RcvAddr),
				tx.Value,
				0,
				0,
				tx.Data,
				tx.Signature,
			)
			incrementalNonce++
		}
	}
}

func generateTransferTx(
	nonce uint64,
	sender crypto.PrivateKey,
	receiver crypto.PrivateKey,
	valToTransfer *big.Int) *transaction.Transaction {

	tx := transaction.Transaction{
		Nonce:   nonce,
		Value:   valToTransfer,
		RcvAddr: skToPk(receiver),
		SndAddr: skToPk(sender),
		Data:    "",
	}
	txBuff, _ := testMarshalizer.Marshal(&tx)
	signer := &singlesig.SchnorrSigner{}
	tx.Signature, _ = signer.Sign(sender, txBuff)

	return &tx
}

func testPrivateKeyHasBalance(t *testing.T, n *testNode, sk crypto.PrivateKey, expectedBalance *big.Int) {
	pkBuff, _ := sk.GeneratePublic().ToByteArray()
	addr, _ := testAddressConverter.CreateAddressFromPublicKeyBytes(pkBuff)
	account, _ := n.accntState.GetExistingAccount(addr)
	assert.Equal(t, expectedBalance, account.(*state.Account).Balance)
}

func proposeBlock(t *testing.T, proposer *testNode, round uint64) (data.BodyHandler, data.HeaderHandler) {
	haveTime := func() bool { return true }

	blockBody, err := proposer.blkProcessor.CreateBlockBody(round, haveTime)
	assert.Nil(t, err)

	blockHeader, err := proposer.blkProcessor.CreateBlockHeader(blockBody, round, haveTime)
	assert.Nil(t, err)

	blockHeader.SetRound(round)
	blockHeader.SetNonce(round)
	blockHeader.SetPubKeysBitmap(make([]byte, 0))
	sig, _ := testMultiSig.AggregateSigs(nil)
	blockHeader.SetSignature(sig)
	currHdr := proposer.blkc.GetCurrentBlockHeader()
	if currHdr == nil {
		currHdr = proposer.blkc.GetGenesisHeader()
	}
	buff, _ := testMarshalizer.Marshal(currHdr)
	blockHeader.SetPrevHash(testHasher.Compute(string(buff)))
	blockHeader.SetPrevRandSeed(currHdr.GetRandSeed())
	blockHeader.SetRandSeed(sig)

	return blockBody, blockHeader
}

func proposeMetaBlock(t *testing.T, proposer *testNode, round uint64) (data.BodyHandler, data.HeaderHandler) {
	metaHeader, err := proposer.blkProcessor.CreateBlockHeader(nil, round, func() bool { return true })
	assert.Nil(t, err)

	metaHeader.SetNonce(round)
	metaHeader.SetRound(round)
	metaHeader.SetPubKeysBitmap(make([]byte, 0))
	sig, _ := testMultiSig.AggregateSigs(nil)
	metaHeader.SetSignature(sig)
	currHdr := proposer.blkc.GetCurrentBlockHeader()
	if currHdr == nil {
		currHdr = proposer.blkc.GetGenesisHeader()
	}
	buff, _ := testMarshalizer.Marshal(currHdr)
	metaHeader.SetPrevHash(testHasher.Compute(string(buff)))
	metaHeader.SetRandSeed(sig)
	metaHeader.SetPrevRandSeed(currHdr.GetRandSeed())

	return nil, metaHeader
}
