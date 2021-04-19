package consensusNotAchieved

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	testBlock "github.com/ElrondNetwork/elrond-go/integrationTests/singleShard/block"
	"github.com/stretchr/testify/assert"
)

var log = logger.GetOrCreate("consensusNotAchieved")

func TestConsensus_BlockWithoutTwoThirdsPlusOneSignaturesOrWrongBitmapShouldNotBeAccepted(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxShards := 1
	consensusGroupSize := 2
	nodesPerShard := 5
	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap(0)

	singleSigner := &mock.SignerMock{
		VerifyStub: func(public crypto.PublicKey, msg []byte, sig []byte) error {
			return nil
		},
		SignStub: func(private crypto.PrivateKey, msg []byte) ([]byte, error) {
			return nil, nil
		},
	}
	keyGen := &mock.KeyGenMock{}

	// create map of shards - testNodeProcessors for metachain and shard chain
	nodesMap := integrationTests.CreateNodesWithNodesCoordinatorAndHeaderSigVerifier(
		nodesPerShard,
		nodesPerShard,
		maxShards,
		consensusGroupSize,
		consensusGroupSize,
		integrationTests.GetConnectableAddress(advertiser),
		singleSigner,
		keyGen,
	)

	for _, nodes := range nodesMap {
		integrationTests.DisplayAndStartNodes(nodes)
		integrationTests.SetEconomicsParameters(nodes, integrationTests.MaxGasLimitPerBlock, integrationTests.MinTxGasPrice, integrationTests.MinTxGasLimit)
	}

	defer func() {
		_ = advertiser.Close()
		for _, nodes := range nodesMap {
			for _, n := range nodes {
				_ = n.Messenger.Close()
			}
		}
	}()

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(testBlock.StepDelay)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	bitMapNotEnough := []byte{1}
	for _, nodes := range nodesMap {
		integrationTests.UpdateRound(nodes, round)
	}
	body, hdr, _ := proposeBlock(nodesMap[0][0], round, nonce, bitMapNotEnough)
	assert.NotNil(t, body)
	assert.NotNil(t, hdr)

	nodesMap[0][0].BroadcastBlock(body, hdr)
	time.Sleep(testBlock.StepDelay)

	// the block should have not pass the interceptor
	assert.Equal(t, int32(0), nodesMap[0][1].CounterHdrRecv)

	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	for _, nodes := range nodesMap {
		integrationTests.UpdateRound(nodes, round)
	}
	bitMapTooBig := []byte{1, 0, 1, 0, 1} // only one byte was needed, so this block should not pass
	body, hdr, _ = proposeBlock(nodesMap[0][0], round, nonce, bitMapTooBig)
	assert.NotNil(t, body)
	assert.NotNil(t, hdr)

	nodesMap[0][0].BroadcastBlock(body, hdr)
	time.Sleep(testBlock.StepDelay)

	// this block should have not passed the interceptor
	assert.Equal(t, int32(0), nodesMap[0][1].CounterHdrRecv)

	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	for _, nodes := range nodesMap {
		integrationTests.UpdateRound(nodes, round)
	}
	bitMapEnough := []byte{11} // 11 = 0b0000 1011 so 3 signatures
	body, hdr, _ = proposeBlock(nodesMap[0][0], round, nonce, bitMapEnough)
	assert.NotNil(t, body)
	assert.NotNil(t, hdr)

	nodesMap[0][0].BroadcastBlock(body, hdr)
	time.Sleep(testBlock.StepDelay)

	// this block should have passed the interceptor
	assert.Equal(t, int32(1), nodesMap[0][1].CounterHdrRecv)
}

func proposeBlock(node *integrationTests.TestProcessorNode, round uint64, nonce uint64, bitmap []byte) (data.BodyHandler, data.HeaderHandler, [][]byte) {
	startTime := time.Now()
	maxTime := time.Second * 2

	haveTime := func() bool {
		elapsedTime := time.Since(startTime)
		remainingTime := maxTime - elapsedTime
		return remainingTime > 0
	}

	blockHeader := node.BlockProcessor.CreateNewHeader(round, nonce)

	err := blockHeader.SetShardID(0)
	if err != nil {
		log.Error("blockHeader.SetShardID", "error", err.Error())
	}

	err = blockHeader.SetPubKeysBitmap(bitmap)
	if err != nil {
		log.Error("blockHeader.SetPubKeysBitmap", "error", err.Error())
	}

	currHdr := node.BlockChain.GetCurrentBlockHeader()
	if check.IfNil(currHdr) {
		currHdr = node.BlockChain.GetGenesisHeader()
	}

	buff, _ := json.Marshal(currHdr)
	err = blockHeader.SetPrevHash(integrationTests.TestHasher.Compute(string(buff)))
	if err != nil {
		log.Error("blockHeader.SetPrevHash", "error", err.Error())
	}

	err = blockHeader.SetPrevRandSeed(currHdr.GetRandSeed())
	if err != nil {
		log.Error("blockHeader.SetPrevRandSeed", "error", err.Error())
	}

	err = blockHeader.SetSignature([]byte("aggregate signature"))
	if err != nil {
		log.Error("blockHeader.SetSignature", "error", err.Error())
	}

	err = blockHeader.SetRandSeed([]byte("aggregate signature"))
	if err != nil {
		log.Error("blockHeader.SetRandSeed", "error", err.Error())
	}

	err = blockHeader.SetLeaderSignature([]byte("leader sign"))
	if err != nil {
		log.Error("blockHeader.SetLeaderSignature", "error", err.Error())
	}

	err = blockHeader.SetChainID(node.ChainID)
	if err != nil {
		log.Error("blockHeader.SetChainID", "error", err.Error())
	}

	blockHeader, blockBody, err := node.BlockProcessor.CreateBlock(blockHeader, haveTime)
	if err != nil {
		fmt.Println(err.Error())
	}

	shardBlockBody, ok := blockBody.(*block.Body)
	txHashes := make([][]byte, 0)
	if !ok {
		return blockBody, blockHeader, txHashes
	}

	for _, mb := range shardBlockBody.MiniBlocks {
		for _, hash := range mb.TxHashes {
			copiedHash := make([]byte, len(hash))
			copy(copiedHash, hash)
			txHashes = append(txHashes, copiedHash)
		}
	}

	return blockBody, blockHeader, txHashes
}
