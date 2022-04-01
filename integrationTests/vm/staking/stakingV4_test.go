package staking

import (
	"math/big"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/stretchr/testify/assert"
)

func createMetaBlockHeader() *block.MetaBlock {
	hdr := block.MetaBlock{
		Nonce:                  1,
		Round:                  1,
		PrevHash:               []byte(""),
		Signature:              []byte("signature"),
		PubKeysBitmap:          []byte("pubKeysBitmap"),
		RootHash:               []byte("roothash"),
		ShardInfo:              make([]block.ShardData, 0),
		TxCount:                1,
		PrevRandSeed:           []byte("roothash"),
		RandSeed:               []byte("roothash"),
		AccumulatedFeesInEpoch: big.NewInt(0),
		AccumulatedFees:        big.NewInt(0),
		DevFeesInEpoch:         big.NewInt(0),
		DeveloperFees:          big.NewInt(0),
	}

	shardMiniBlockHeaders := make([]block.MiniBlockHeader, 0)
	shardMiniBlockHeader := block.MiniBlockHeader{
		Hash:            []byte("mb_hash1"),
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxCount:         1,
	}
	shardMiniBlockHeaders = append(shardMiniBlockHeaders, shardMiniBlockHeader)
	shardData := block.ShardData{
		Nonce:                 1,
		ShardID:               0,
		HeaderHash:            []byte("hdr_hash1"),
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
	metaHdr := createMetaBlockHeader()
	headerHandler, bodyHandler, err := node.MetaBlockProcessor.CreateBlock(metaHdr, func() bool { return true })
	assert.Nil(t, err)

	node.DisplayNodesConfig(0, 1)

	err = node.MetaBlockProcessor.ProcessBlock(headerHandler, bodyHandler, func() time.Duration { return time.Second })
	assert.Nil(t, err)
	node.DisplayNodesConfig(0, 1)
}
