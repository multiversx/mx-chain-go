package block

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing/kv2/singlesig"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/node"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

func TestShouldProcessBlocksInMultiShardArchitecture(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 6
	startingPort := 36000
	nodesPerShard := 3

	senderShard := uint32(0)
	recvShards := []uint32{1, 2}

	advertiser := createMessengerWithKadDht(context.Background(), startingPort, "")
	advertiser.Bootstrap()
	startingPort++

	nodes := createNodes(
		startingPort,
		numOfShards,
		nodesPerShard,
		getConnectableAddress(advertiser),
	)
	displayAndStartNodes(nodes)

	defer func() {
		advertiser.Close()
		for _, n := range nodes {
			n.node.Stop()
		}
	}()

	// delay for bootstrapping and topic announcement
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 0)
	txToGenerateInEachMiniBlock := 3

	senderNode := nodes[0]

	//sender shard keys, receivers  keys
	sendersPrivateKeys := make([]crypto.PrivateKey, 3)
	receiversPrivateKeys := make(map[uint32][]crypto.PrivateKey)
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

	generateAndDisseminateTxs(senderNode.node, sendersPrivateKeys, receiversPrivateKeys)

}

func generateAndDisseminateTxs(
	n *node.Node,
	senders []crypto.PrivateKey,
	receiversPrivateKeys map[uint32][]crypto.PrivateKey,
) {

	for i := 0; i < len(senders); i++ {
		senderKey := senders[i]
		incrementalNonce := uint64(0)
		for _, privateKeys := range receiversPrivateKeys {
			receiverKey := privateKeys[i]
			tx := generateTransferTxWithValueOne(senderKey, receiverKey)
			n.SendTransaction(
				tx.Nonce+incrementalNonce,
				hex.EncodeToString(tx.SndAddr),
				hex.EncodeToString(tx.RcvAddr),
				tx.Value,
				string(tx.Data),
				tx.Signature,
			)
			incrementalNonce++
		}
	}
}

func generateTransferTxWithValueOne(sender crypto.PrivateKey, receiver crypto.PrivateKey) *transaction.Transaction {
	tx := transaction.Transaction{
		Nonce:   0,
		Value:   big.NewInt(1),
		RcvAddr: skToPk(receiver),
		SndAddr: skToPk(sender),
		Data:    make([]byte, 0),
	}
	txBuff, _ := testMarshalizer.Marshal(&tx)
	signer := &singlesig.SchnorrSigner{}
	tx.Signature, _ = signer.Sign(sender, txBuff)

	return &tx
}

func skToPk(sk crypto.PrivateKey) []byte {
	pkBuff, _ := sk.GeneratePublic().ToByteArray()
	return pkBuff
}
