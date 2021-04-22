package txRouting

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	ed25519SingleSig "github.com/ElrondNetwork/elrond-go/crypto/signing/ed25519/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoutingOfTransactionsInShards(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 5
	nodesPerShard := 1
	numMetachainNodes := 1

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
	)
	mintValue := big.NewInt(1000000000000000000)
	integrationTests.MintAllNodes(nodes, mintValue)
	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	for i := 0; i < numOfShards; i++ {
		txs := generateTransactionsInAllConfigurations(nodes, uint32(numOfShards))
		require.Equal(t, numOfShards, len(txs))

		for shardID, listTxs := range txs {
			integrationTests.WhiteListTxs(nodes, listTxs)
			dispatchNode := getNodeOnShard(shardID, nodes)
			_, err := dispatchNode.Node.SendBulkTransactions(listTxs)
			require.Nil(t, err)
		}
	}

	fmt.Println("waiting for txs to be disseminated")
	time.Sleep(integrationTests.StepDelay)

	//expectedNumTxs is computed in the following manner:
	//- if the node sends all numOfShards*numOfShards txs, then it will have in the pool
	//  (numOfShards + numOfShards - 1 -> both sender and destination) txs related to its shard
	//- if the node will have to receive all those generated txs, it will receive only those txs
	//  where the node is the sender, so numOfShards in total
	//- since all shards emmit numOfShards*numOfShards, expectedNumTxs is computed as:
	expectedNumTxs := numOfShards * numOfShards
	checkTransactionsInPool(t, nodes, expectedNumTxs)
}

func generateTransactionsInAllConfigurations(nodes []*integrationTests.TestProcessorNode, numOfShards uint32) map[uint32][]*transaction.Transaction {
	txs := make(map[uint32][]*transaction.Transaction)

	for i := uint32(0); i < numOfShards; i++ {
		senderNode := getNodeOnShard(i, nodes)

		for j := uint32(0); j < numOfShards; j++ {
			destNode := getNodeOnShard(j, nodes)

			tx := generateTx(senderNode.OwnAccount.SkTxSign, destNode.OwnAccount.PkTxSign, senderNode.OwnAccount.Nonce)
			senderNode.OwnAccount.Nonce++
			txs[i] = append(txs[i], tx)
		}
	}

	return txs
}

func getNodeOnShard(shardId uint32, nodes []*integrationTests.TestProcessorNode) *integrationTests.TestProcessorNode {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() == shardId {
			return n
		}
	}

	return nil
}

func checkTransactionsInPool(t *testing.T, nodes []*integrationTests.TestProcessorNode, numExpected int) {
	for _, n := range nodes {
		if n.ShardCoordinator.SelfId() == core.MetachainShardId {
			assert.Equal(t, int32(0), n.CounterTxRecv)
			continue
		}

		assert.Equal(t, int32(numExpected), n.CounterTxRecv)
		checkTransactionsInPoolAreForTheNode(t, n)
	}
}

func checkTransactionsInPoolAreForTheNode(t *testing.T, n *integrationTests.TestProcessorNode) {
	n.ReceivedTransactions.Range(func(key, value interface{}) bool {
		tx, ok := value.(*transaction.Transaction)
		if !ok {
			assert.Fail(t, "found a non transaction object, key: "+hex.EncodeToString(key.([]byte)))
			return false
		}

		senderBuff := tx.SndAddr
		recvBuff := tx.RcvAddr

		senderShardId := n.ShardCoordinator.ComputeId(senderBuff)
		recvShardId := n.ShardCoordinator.ComputeId(recvBuff)
		isForCurrentShard := senderShardId == n.ShardCoordinator.SelfId() || recvShardId == n.ShardCoordinator.SelfId()
		if !isForCurrentShard {
			assert.Fail(t, fmt.Sprintf("found a transaction wrongly dispatched, key: %s, tx: %v",
				hex.EncodeToString(key.([]byte)),
				tx,
			),
			)
			return false
		}

		return true
	})
}

func generateTx(sender crypto.PrivateKey, receiver crypto.PublicKey, nonce uint64) *transaction.Transaction {
	receiverBytes, _ := receiver.ToByteArray()
	senderBytes, _ := sender.GeneratePublic().ToByteArray()

	tx := &transaction.Transaction{
		Nonce:     nonce,
		Value:     big.NewInt(10),
		RcvAddr:   receiverBytes,
		SndAddr:   senderBytes,
		GasPrice:  integrationTests.MinTxGasPrice,
		GasLimit:  integrationTests.MinTxGasLimit,
		Data:      []byte(""),
		ChainID:   integrationTests.ChainID,
		Signature: nil,
		Version:   integrationTests.MinTransactionVersion,
	}
	marshalizedTxBeforeSigning, _ := tx.GetDataForSigning(integrationTests.TestAddressPubkeyConverter, integrationTests.TestTxSignMarshalizer)
	signer := ed25519SingleSig.Ed25519Signer{}

	signature, _ := signer.Sign(sender, marshalizedTxBeforeSigning)
	tx.Signature = signature
	return tx
}
