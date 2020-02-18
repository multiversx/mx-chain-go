package block

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/stretchr/testify/assert"
)

func TestConsensus_BlockWithoutTwoThirdsPlusOneSignaturesOrWrongBitmapShouldNotBeAccepted(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	maxShards := 1
	consensusGroupSize := 2
	nodesPerShard := 5
	advertiser := integrationTests.CreateMessengerWithKadDht(context.Background(), "")
	_ = advertiser.Bootstrap()

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
				_ = n.Node.Stop()
			}
		}
	}()

	fmt.Println("Delaying for nodes p2p bootstrap...")
	time.Sleep(stepDelay)

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
	time.Sleep(stepDelay)

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
	time.Sleep(stepDelay)

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
	time.Sleep(stepDelay)

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

	blockHeader := &block.Header{}

	blockHeader.SetShardID(0)
	blockHeader.SetRound(round)
	blockHeader.SetNonce(nonce)
	blockHeader.SetPubKeysBitmap(bitmap)
	currHdr := node.BlockChain.GetCurrentBlockHeader()
	if currHdr == nil {
		currHdr = node.BlockChain.GetGenesisHeader()
	}

	buff, _ := json.Marshal(currHdr)
	blockHeader.SetPrevHash(integrationTests.TestHasher.Compute(string(buff)))
	blockHeader.SetPrevRandSeed(currHdr.GetRandSeed())
	blockHeader.SetSignature([]byte("aggregate signature"))
	blockHeader.SetRandSeed([]byte("aggregate signature"))
	blockHeader.SetLeaderSignature([]byte("leader sign"))
	blockHeader.SetChainID(node.ChainID)

	blockBody, err := node.BlockProcessor.CreateBlockBody(blockHeader, haveTime)
	if err != nil {
		fmt.Println(err.Error())
	}
	blockBody, err = node.BlockProcessor.ApplyBodyToHeader(blockHeader, blockBody)
	if err != nil {
		fmt.Println(err.Error())
	}

	shardBlockBody, ok := blockBody.(block.Body)
	txHashes := make([][]byte, 0)
	if !ok {
		return blockBody, blockHeader, txHashes
	}

	for _, mb := range shardBlockBody {
		for _, hash := range mb.TxHashes {
			copiedHash := make([]byte, len(hash))
			copy(copiedHash, hash)
			txHashes = append(txHashes, copiedHash)
		}
	}

	return blockBody, blockHeader, txHashes
}
