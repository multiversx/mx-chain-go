package block

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/kyber/singlesig"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestProcessSCCallsInMultiShardArchitecture(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	fmt.Println("Step 1. Setup nodes...")

	numOfShards := 2
	nodesPerShard := 1
	generalRoundNumber := uint32(1)

	senderShard := uint32(0)
	senderNonce := uint64(1)
	senderMintingValue := big.NewInt(1000)
	receiverShard := uint32(1)

	advertiser := createMessengerWithKadDht(context.Background(), "")
	advertiser.Bootstrap()

	nodes := createNodes(
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

	proposerNodeShard1 := nodes[0]

	// delay for bootstrapping and topic announcement
	fmt.Println("Delaying for node bootstrap and topic announcement...")
	time.Sleep(time.Second * 5)

	generateCoordinator, _ := sharding.NewMultiShardCoordinator(uint32(numOfShards), 0)
	senderPrivateKey := generatePrivateKeyInShardId(generateCoordinator, senderShard)
	_ = generatePrivateKeyInShardId(generateCoordinator, receiverShard)

	// Minting sender
	createMintingForSenders(nodes, senderShard, []crypto.PrivateKey{senderPrivateKey}, senderMintingValue)

	// should deploy smart contract -> we process a block containing only the sc deployment tx
	scCode := "mocked code, not taken into account"
	scDeplymentSender := skToPk(senderPrivateKey)
	fmt.Println(scDeplymentSender)
	initialValueForInternalVariable := uint64(45)

	contractTx := createTx(
		t,
		scDeplymentSender,
		vm.CreateEmptyAddress().Bytes(),
		senderNonce,
		big.NewInt(0),
		0,
		0,
		scCode,
		initialValueForInternalVariable,
	)
	senderNonce++

	signer := &singlesig.SchnorrSigner{}
	txBuff , _ := testMarshalizer.Marshal(contractTx)
	contractTx.Signature, _ = signer.Sign(senderPrivateKey, txBuff)

	nodes[0].node.SendTransaction(contractTx.Nonce, hex.EncodeToString(contractTx.SndAddr),
		hex.EncodeToString(contractTx.RcvAddr), contractTx.Value, contractTx.GasPrice, contractTx.GasLimit,
		string(contractTx.Data), contractTx.Signature)

	fmt.Println("Waiting for the SC deployment transaction to be propagated...")
	time.Sleep(time.Second * 5)

	body, header := proposeBlock(t, proposerNodeShard1, generalRoundNumber)
	err := nodes[0].blkProcessor.ProcessBlock(nodes[0].blkc, header, body, func() time.Duration {return time.Second * 2})
	assert.Nil(t, err)

	generalRoundNumber++

	// should execute smart contract + extra transfer transactions in the mix
	addValue := uint64(128)
	scAddress, _ := hex.DecodeString("b473e38c40c60c8b22eb80fffff0717a98386bb129a7750ffe8091c77b85990e")
	contractCallTx := createTx(
		t,
		scDeplymentSender,
		scAddress,
		senderNonce,
		big.NewInt(0),
		1,
		1,
		"Add",
		addValue,
	)
	senderNonce++

	txBuff , _ = testMarshalizer.Marshal(contractTx)
	contractCallTx.Signature, _ = signer.Sign(senderPrivateKey, txBuff)

	nodes[0].node.SendTransaction(contractTx.Nonce, hex.EncodeToString(contractTx.SndAddr),
		hex.EncodeToString(contractTx.RcvAddr), contractTx.Value, contractTx.GasPrice, contractTx.GasLimit,
		string(contractTx.Data), contractTx.Signature)

	body, header = proposeBlock(t, proposerNodeShard1, generalRoundNumber)
	err = nodes[0].blkProcessor.ProcessBlock(nodes[0].blkc, header, body, func() time.Duration {return time.Second * 2})
	assert.Nil(t, err)

	generalRoundNumber++
}

func createTx(
	t *testing.T,
	senderAddressBytes []byte,
	receiverAddressBytes []byte,
	senderNonce uint64,
	value *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	scCodeOrFunc string,
	scValue uint64,
) *transaction.Transaction {

	txData := fmt.Sprintf("%s@%X", scCodeOrFunc, scValue)
	tx := &transaction.Transaction{
		Nonce:    senderNonce,
		Value:    value,
		RcvAddr:  receiverAddressBytes,
		SndAddr:  senderAddressBytes,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Data:     []byte(txData),
	}
	assert.NotNil(t, tx)

	return tx
}

func createBlock(proposer *testNode, round uint32) (data.BodyHandler, data.HeaderHandler) {
	haveTime := func() bool { return true }

	blockBody, _ := proposer.blkProcessor.CreateBlockBody(round, haveTime)

	blockHeader, _ := proposer.blkProcessor.CreateBlockHeader(blockBody, round, haveTime)

	return blockBody, blockHeader
}
