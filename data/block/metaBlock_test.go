package block_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaBlock_GetEpoch(t *testing.T) {
	t.Parallel()

	epoch := uint32(1)
	m := block.MetaBlock{
		Epoch: epoch,
	}

	assert.Equal(t, epoch, m.GetEpoch())
}

func TestMetaBlock_GetShard(t *testing.T) {
	t.Parallel()

	m := block.MetaBlock{}

	assert.Equal(t, core.MetachainShardId, m.GetShardID())
}

func TestMetaBlock_GetNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(2)
	m := block.MetaBlock{
		Nonce: nonce,
	}

	assert.Equal(t, nonce, m.GetNonce())
}

func TestMetaBlock_GetPrevHash(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prev hash")
	m := block.MetaBlock{
		PrevHash: prevHash,
	}

	assert.Equal(t, prevHash, m.GetPrevHash())
}

func TestMetaBlock_GetPubKeysBitmap(t *testing.T) {
	t.Parallel()

	pubKeysBitmap := []byte{10, 11, 12, 13}
	m := block.MetaBlock{
		PubKeysBitmap: pubKeysBitmap,
	}

	assert.Equal(t, pubKeysBitmap, m.GetPubKeysBitmap())
}

func TestMetaBlock_GetPrevRandSeed(t *testing.T) {
	t.Parallel()

	prevRandSeed := []byte("previous random seed")
	m := block.MetaBlock{
		PrevRandSeed: prevRandSeed,
	}

	assert.Equal(t, prevRandSeed, m.GetPrevRandSeed())
}

func TestMetaBlock_GetRandSeed(t *testing.T) {
	t.Parallel()

	randSeed := []byte("random seed")
	m := block.MetaBlock{
		RandSeed: randSeed,
	}

	assert.Equal(t, randSeed, m.GetRandSeed())
}

func TestMetaBlock_GetRootHash(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	m := block.MetaBlock{
		RootHash: rootHash,
	}

	assert.Equal(t, rootHash, m.GetRootHash())
}

func TestMetaBlock_GetRound(t *testing.T) {
	t.Parallel()

	round := uint64(1234)
	m := block.MetaBlock{
		Round: round,
	}

	assert.Equal(t, round, m.GetRound())
}

func TestMetaBlock_GetTimestamp(t *testing.T) {
	t.Parallel()

	timestamp := uint64(1000000)
	m := block.MetaBlock{
		TimeStamp: timestamp,
	}

	assert.Equal(t, timestamp, m.GetTimeStamp())
}

func TestMetaBlock_GetSignature(t *testing.T) {
	t.Parallel()

	signature := []byte("signature")
	m := block.MetaBlock{
		Signature: signature,
	}

	assert.Equal(t, signature, m.GetSignature())
}

func TestMetaBlock_GetTxCount(t *testing.T) {
	t.Parallel()

	txCount := uint32(100)
	m := block.MetaBlock{
		TxCount: txCount,
	}

	assert.Equal(t, txCount, m.GetTxCount())
}

func TestMetaBlock_SetEpoch(t *testing.T) {
	t.Parallel()

	epoch := uint32(10)
	m := block.MetaBlock{}
	err := m.SetEpoch(epoch)

	assert.Nil(t, err)
	assert.Equal(t, epoch, m.GetEpoch())
}

func TestMetaBlock_SetNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(11)
	m := block.MetaBlock{}
	err := m.SetNonce(nonce)

	assert.Nil(t, err)
	assert.Equal(t, nonce, m.GetNonce())
}

func TestMetaBlock_SetPrevHash(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prev hash")
	m := block.MetaBlock{}
	err := m.SetPrevHash(prevHash)

	assert.Nil(t, err)
	assert.Equal(t, prevHash, m.GetPrevHash())
}

func TestMetaBlock_SetPubKeysBitmap(t *testing.T) {
	t.Parallel()

	pubKeysBitmap := []byte{12, 13, 14, 15}
	m := block.MetaBlock{}
	err := m.SetPubKeysBitmap(pubKeysBitmap)

	assert.Nil(t, err)
	assert.Equal(t, pubKeysBitmap, m.GetPubKeysBitmap())
}

func TestMetaBlock_SetPrevRandSeed(t *testing.T) {
	t.Parallel()

	prevRandSeed := []byte("previous random seed")
	m := block.MetaBlock{}
	err := m.SetPrevRandSeed(prevRandSeed)

	assert.Nil(t, err)
	assert.Equal(t, prevRandSeed, m.GetPrevRandSeed())
}

func TestMetaBlock_SetRandSeed(t *testing.T) {
	t.Parallel()

	randSeed := []byte("random seed")
	m := block.MetaBlock{}
	err := m.SetRandSeed(randSeed)

	assert.Nil(t, err)
	assert.Equal(t, randSeed, m.GetRandSeed())
}

func TestMetaBlock_SetRootHash(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	m := block.MetaBlock{}
	err := m.SetRootHash(rootHash)

	assert.Nil(t, err)
	assert.Equal(t, rootHash, m.GetRootHash())
}

func TestMetaBlock_SetRound(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	m := block.MetaBlock{}
	err := m.SetRootHash(rootHash)

	assert.Nil(t, err)
	assert.Equal(t, rootHash, m.GetRootHash())
}

func TestMetaBlock_SetSignature(t *testing.T) {
	t.Parallel()

	signature := []byte("signature")
	m := block.MetaBlock{}
	err := m.SetSignature(signature)

	assert.Nil(t, err)
	assert.Equal(t, signature, m.GetSignature())
}

func TestMetaBlock_SetTimeStamp(t *testing.T) {
	t.Parallel()

	timestamp := uint64(100000)
	m := block.MetaBlock{}
	err := m.SetTimeStamp(timestamp)

	assert.Nil(t, err)
	assert.Equal(t, timestamp, m.GetTimeStamp())
}

func TestMetaBlock_SetTxCount(t *testing.T) {
	t.Parallel()

	txCount := uint32(100)
	m := block.MetaBlock{}
	err := m.SetTxCount(txCount)

	assert.Nil(t, err)
	assert.Equal(t, txCount, m.GetTxCount())
}

func TestMetaBlock_GetMiniBlockHeadersWithDst(t *testing.T) {
	t.Parallel()

	metaHdr := &block.MetaBlock{Round: 15}
	metaHdr.ShardInfo = make([]block.ShardData, 0)

	shardMBHeader := make([]block.MiniBlockHeader, 0)
	shMBHdr1 := block.MiniBlockHeader{SenderShardID: 0, ReceiverShardID: 1, Hash: []byte("hash1")}
	shMBHdr2 := block.MiniBlockHeader{SenderShardID: 0, ReceiverShardID: 1, Hash: []byte("hash2")}
	shardMBHeader = append(shardMBHeader, shMBHdr1, shMBHdr2)

	shData1 := block.ShardData{ShardID: 0, HeaderHash: []byte("sh"), ShardMiniBlockHeaders: shardMBHeader}
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shData1)

	shData2 := block.ShardData{ShardID: 1, HeaderHash: []byte("sh"), ShardMiniBlockHeaders: shardMBHeader}
	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shData2)

	mbDst0 := metaHdr.GetMiniBlockHeadersWithDst(0)
	assert.Equal(t, 0, len(mbDst0))
	mbDst1 := metaHdr.GetMiniBlockHeadersWithDst(1)
	assert.Equal(t, len(shardMBHeader), len(mbDst1))
}

func TestMetaBlock_GetOrderedCrossMiniblocksWithDstShouldWork(t *testing.T) {
	t.Parallel()

	metaHdr := &block.MetaBlock{Round: 6}
	metaHdr.ShardInfo = make([]block.ShardData, 0)

	shardMBHeader1 := make([]block.MiniBlockHeader, 0)
	shMBHdr1 := block.MiniBlockHeader{SenderShardID: 0, ReceiverShardID: 1, Hash: []byte("hash1")}
	shardMBHeader1 = append(shardMBHeader1, shMBHdr1)
	shData1 := block.ShardData{Round: 11, ShardID: 0, HeaderHash: []byte("sh1"), ShardMiniBlockHeaders: shardMBHeader1}

	shardMBHeader2 := make([]block.MiniBlockHeader, 0)
	shMBHdr2 := block.MiniBlockHeader{SenderShardID: 0, ReceiverShardID: 1, Hash: []byte("hash2")}
	shardMBHeader2 = append(shardMBHeader2, shMBHdr2)
	shData2 := block.ShardData{Round: 9, ShardID: 0, HeaderHash: []byte("sh2"), ShardMiniBlockHeaders: shardMBHeader2}

	shardMBHeader3 := make([]block.MiniBlockHeader, 0)
	shMBHdr3 := block.MiniBlockHeader{SenderShardID: 2, ReceiverShardID: 1, Hash: []byte("hash3")}
	shardMBHeader3 = append(shardMBHeader3, shMBHdr3)
	shData3 := block.ShardData{Round: 10, ShardID: 2, HeaderHash: []byte("sh3"), ShardMiniBlockHeaders: shardMBHeader3}

	shardMBHeader4 := make([]block.MiniBlockHeader, 0)
	shMBHdr4 := block.MiniBlockHeader{SenderShardID: 2, ReceiverShardID: 1, Hash: []byte("hash4")}
	shardMBHeader4 = append(shardMBHeader4, shMBHdr4)
	shData4 := block.ShardData{Round: 8, ShardID: 2, HeaderHash: []byte("sh4"), ShardMiniBlockHeaders: shardMBHeader4}

	shardMBHeader5 := make([]block.MiniBlockHeader, 0)
	shMBHdr5 := block.MiniBlockHeader{SenderShardID: 1, ReceiverShardID: 2, Hash: []byte("hash5")}
	shardMBHeader5 = append(shardMBHeader5, shMBHdr5)
	shData5 := block.ShardData{Round: 7, ShardID: 1, HeaderHash: []byte("sh5"), ShardMiniBlockHeaders: shardMBHeader5}

	metaHdr.ShardInfo = append(metaHdr.ShardInfo, shData1, shData2, shData3, shData4, shData5)

	metaHdr.MiniBlockHeaders = append(metaHdr.MiniBlockHeaders, block.MiniBlockHeader{
		Hash:            []byte("hash6"),
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 1,
	})

	metaHdr.MiniBlockHeaders = append(metaHdr.MiniBlockHeaders, block.MiniBlockHeader{
		Hash:            []byte("hash7"),
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: core.AllShardId,
	})

	metaHdr.MiniBlockHeaders = append(metaHdr.MiniBlockHeaders, block.MiniBlockHeader{
		Hash:            []byte("hash8"),
		SenderShardID:   core.MetachainShardId,
		ReceiverShardID: 2,
	})

	miniBlocksInfo := metaHdr.GetOrderedCrossMiniblocksWithDst(1)
	require.Equal(t, 6, len(miniBlocksInfo))
	assert.Equal(t, miniBlocksInfo[0].Hash, []byte("hash6"))
	assert.Equal(t, miniBlocksInfo[0].Round, uint64(6))
	assert.Equal(t, miniBlocksInfo[1].Hash, []byte("hash7"))
	assert.Equal(t, miniBlocksInfo[1].Round, uint64(6))
	assert.Equal(t, miniBlocksInfo[2].Hash, []byte("hash4"))
	assert.Equal(t, miniBlocksInfo[2].Round, uint64(8))
	assert.Equal(t, miniBlocksInfo[3].Hash, []byte("hash2"))
	assert.Equal(t, miniBlocksInfo[3].Round, uint64(9))
	assert.Equal(t, miniBlocksInfo[4].Hash, []byte("hash3"))
	assert.Equal(t, miniBlocksInfo[4].Round, uint64(10))
	assert.Equal(t, miniBlocksInfo[5].Hash, []byte("hash1"))
	assert.Equal(t, miniBlocksInfo[5].Round, uint64(11))

	miniBlocksInfo = metaHdr.GetOrderedCrossMiniblocksWithDst(2)
	require.Equal(t, 3, len(miniBlocksInfo))
	assert.Equal(t, miniBlocksInfo[0].Hash, []byte("hash7"))
	assert.Equal(t, miniBlocksInfo[0].Round, uint64(6))
	assert.Equal(t, miniBlocksInfo[1].Hash, []byte("hash8"))
	assert.Equal(t, miniBlocksInfo[1].Round, uint64(6))
	assert.Equal(t, miniBlocksInfo[2].Hash, []byte("hash5"))
	assert.Equal(t, miniBlocksInfo[2].Round, uint64(7))
}
