package staking

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/require"
)

func createMetaBlockHeader(epoch uint32, round uint64, prevHash []byte) *block.MetaBlock {
	hdr := block.MetaBlock{
		Epoch:                  epoch,
		Nonce:                  round,
		Round:                  round,
		PrevHash:               prevHash,
		Signature:              []byte("signature"),
		PubKeysBitmap:          []byte("pubKeysBitmap"),
		RootHash:               []byte("roothash"),
		ShardInfo:              make([]block.ShardData, 0),
		TxCount:                1,
		PrevRandSeed:           []byte("roothash"),
		RandSeed:               []byte("roothash" + strconv.Itoa(int(round))),
		AccumulatedFeesInEpoch: big.NewInt(0),
		AccumulatedFees:        big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
	}

	shardMiniBlockHeaders := make([]block.MiniBlockHeader, 0)
	shardMiniBlockHeader := block.MiniBlockHeader{
		Hash:            []byte("mb_hash" + strconv.Itoa(int(round))),
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxCount:         1,
	}
	shardMiniBlockHeaders = append(shardMiniBlockHeaders, shardMiniBlockHeader)
	shardData := block.ShardData{
		Nonce:                 round,
		ShardID:               0,
		HeaderHash:            []byte("hdr_hash" + strconv.Itoa(int(round))),
		TxCount:               1,
		ShardMiniBlockHeaders: shardMiniBlockHeaders,
		DeveloperFees:         big.NewInt(0),
		AccumulatedFees:       big.NewInt(0),
	}
	hdr.ShardInfo = append(hdr.ShardInfo, shardData)

	return &hdr
}

func TestNewTestMetaProcessor(t *testing.T) {
	node := NewTestMetaProcessor(1, 1, 1, 1, 1)
	//metaHdr := createMetaBlockHeader(1,1)
	//headerHandler, bodyHandler, err := node.MetaBlockProcessor.CreateBlock(metaHdr, func() bool { return true })
	//assert.Nil(t, err)
	//
	//node.DisplayNodesConfig(0, 1)
	//
	//err = node.MetaBlockProcessor.ProcessBlock(headerHandler, bodyHandler, func() time.Duration { return time.Second })
	//assert.Nil(t, err)
	//
	//err = node.MetaBlockProcessor.CommitBlock(headerHandler, bodyHandler)
	node.DisplayNodesConfig(0, 1)
	newHdr := createMetaBlockHeader(1, 1, []byte(""))
	newHdr.SetPrevHash(node.GenesisHeader.Hash)
	newHdr2, newBodyHandler2, err := node.MetaBlockProcessor.CreateBlock(newHdr, func() bool { return true })

	require.Nil(t, err)
	//newHdr22 := newHdr2.(*block.MetaBlock)

	//valstat, _ := hex.DecodeString("8de5a7881cdf0edc6f37d0382f870609c4a79559b0c4dbac8260fea955db9bb9")
	//newHdr22.ValidatorStatsRootHash = valstat

	//err = node.MetaBlockProcessor.ProcessBlock(newHdr2, newBodyHandler2, func() time.Duration { return 4 * time.Second })
	//require.Nil(t, err)
	err = node.MetaBlockProcessor.CommitBlock(newHdr2, newBodyHandler2)
	require.Nil(t, err)

	currentBlockHeader := node.BlockChain.GetCurrentBlockHeader()
	if check.IfNil(currentBlockHeader) {
		currentBlockHeader = node.BlockChain.GetGenesisHeader()
	}

	marshaller := &mock.MarshalizerMock{}
	prevBlockBytes, _ := marshaller.Marshal(newHdr2)
	prevBlockBytes = sha256.NewSha256().Compute(string(prevBlockBytes))
	prevBlockHash := hex.EncodeToString(prevBlockBytes)
	fmt.Println(prevBlockHash)

	//prevHash, _ := hex.DecodeString("a9307adeffe84090fab6a0e2e6c94c4102bdf083bc1314a389e4e85500861710")
	prevRandomness := currentBlockHeader.GetRandSeed()
	newRandomness := currentBlockHeader.GetRandSeed()
	anotherHdr := createMetaBlockHeader(1, 2, prevBlockBytes)

	//	rootHash ,_ := node.ValidatorStatistics.RootHash()
	//	anotherHdr.ValidatorStatsRootHash = rootHash
	anotherHdr.PrevRandSeed = prevRandomness
	anotherHdr.RandSeed = newRandomness
	hh, bb, err := node.MetaBlockProcessor.CreateBlock(anotherHdr, func() bool { return true })
	require.Nil(t, err)

	//err = node.MetaBlockProcessor.ProcessBlock(hh,bb,func() time.Duration { return 4* time.Second })
	//require.Nil(t, err)

	err = node.MetaBlockProcessor.CommitBlock(hh, bb)
	require.Nil(t, err)

	/*
		prevHash, _ := hex.DecodeString("7a8de8d447691a793f053a7e744b28da19c42cedbef7e76caef7d4acb2ff3906")
		prevRandSeed := newHdr2.GetRandSeed()
		newHdr2 = createMetaBlockHeader(2,2, prevHash)
		newHdr2.SetPrevRandSeed(prevRandSeed)

		metablk := newHdr2.(*block.MetaBlock)
		valStats, _ := hex.DecodeString("5f4f6e8be67205b432eaf2aafb2b1aa3555cf58a936a5f93b3b89917a9a9fa42")
		metablk.ValidatorStatsRootHash = valStats
		newHdr2, newBodyHandler2, err = node.MetaBlockProcessor.CreateBlock(newHdr2, func() bool { return true })
		require.Nil(t, err)
		err = node.MetaBlockProcessor.ProcessBlock(newHdr2, newBodyHandler2, func() time.Duration { return time.Second })
		require.Nil(t, err)
		err = node.MetaBlockProcessor.CommitBlock(newHdr2, newBodyHandler2)
		require.Nil(t, err)

	*/
}
