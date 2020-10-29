package block_test

import (
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeader_GetEpoch(t *testing.T) {
	t.Parallel()

	epoch := uint32(1)
	h := block.Header{
		Epoch: epoch,
	}

	assert.Equal(t, epoch, h.GetEpoch())
}

func TestHeader_GetShard(t *testing.T) {
	t.Parallel()

	shardId := uint32(2)
	h := block.Header{
		ShardID: shardId,
	}

	assert.Equal(t, shardId, h.GetShardID())
}

func TestHeader_GetNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(2)
	h := block.Header{
		Nonce: nonce,
	}

	assert.Equal(t, nonce, h.GetNonce())
}

func TestHeader_GetPrevHash(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prev hash")
	h := block.Header{
		PrevHash: prevHash,
	}

	assert.Equal(t, prevHash, h.GetPrevHash())
}

func TestHeader_GetPrevRandSeed(t *testing.T) {
	t.Parallel()

	prevRandSeed := []byte("prev random seed")
	h := block.Header{
		PrevRandSeed: prevRandSeed,
	}

	assert.Equal(t, prevRandSeed, h.GetPrevRandSeed())
}

func TestHeader_GetRandSeed(t *testing.T) {
	t.Parallel()

	randSeed := []byte("random seed")
	h := block.Header{
		RandSeed: randSeed,
	}

	assert.Equal(t, randSeed, h.GetRandSeed())
}

func TestHeader_GetPubKeysBitmap(t *testing.T) {
	t.Parallel()

	pubKeysBitmap := []byte{10, 11, 12, 13}
	h := block.Header{
		PubKeysBitmap: pubKeysBitmap,
	}

	assert.Equal(t, pubKeysBitmap, h.GetPubKeysBitmap())
}

func TestHeader_GetRootHash(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	h := block.Header{
		RootHash: rootHash,
	}

	assert.Equal(t, rootHash, h.GetRootHash())
}

func TestHeader_GetRound(t *testing.T) {
	t.Parallel()

	round := uint64(1234)
	h := block.Header{
		Round: round,
	}

	assert.Equal(t, round, h.GetRound())
}

func TestHeader_GetSignature(t *testing.T) {
	t.Parallel()

	signature := []byte("signature")
	h := block.Header{
		Signature: signature,
	}

	assert.Equal(t, signature, h.GetSignature())
}

func TestHeader_GetTxCount(t *testing.T) {
	t.Parallel()

	txCount := uint32(10)
	h := block.Header{
		TxCount: txCount,
	}

	assert.Equal(t, txCount, h.GetTxCount())
}

func TestHeader_SetEpoch(t *testing.T) {
	t.Parallel()

	epoch := uint32(10)
	h := block.Header{}
	h.SetEpoch(epoch)

	assert.Equal(t, epoch, h.GetEpoch())
}

func TestHeader_SetNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(11)
	h := block.Header{}
	h.SetNonce(nonce)

	assert.Equal(t, nonce, h.GetNonce())
}

func TestHeader_SetPrevHash(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prev hash")
	h := block.Header{}
	h.SetPrevHash(prevHash)

	assert.Equal(t, prevHash, h.GetPrevHash())
}

func TestHeader_SetPrevRandSeed(t *testing.T) {
	t.Parallel()

	prevRandSeed := []byte("prev random seed")
	h := block.Header{}
	h.SetPrevRandSeed(prevRandSeed)

	assert.Equal(t, prevRandSeed, h.GetPrevRandSeed())
}

func TestHeader_SetRandSeed(t *testing.T) {
	t.Parallel()

	randSeed := []byte("random seed")
	h := block.Header{}
	h.SetRandSeed(randSeed)

	assert.Equal(t, randSeed, h.GetRandSeed())
}

func TestHeader_SetPubKeysBitmap(t *testing.T) {
	t.Parallel()

	pubKeysBitmap := []byte{12, 13, 14, 15}
	h := block.Header{}
	h.SetPubKeysBitmap(pubKeysBitmap)

	assert.Equal(t, pubKeysBitmap, h.GetPubKeysBitmap())
}

func TestHeader_SetRootHash(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	h := block.Header{}
	h.SetRootHash(rootHash)

	assert.Equal(t, rootHash, h.GetRootHash())
}

func TestHeader_SetRound(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	h := block.Header{}
	h.SetRootHash(rootHash)

	assert.Equal(t, rootHash, h.GetRootHash())
}

func TestHeader_SetSignature(t *testing.T) {
	t.Parallel()

	signature := []byte("signature")
	h := block.Header{}
	h.SetSignature(signature)

	assert.Equal(t, signature, h.GetSignature())
}

func TestHeader_SetTimeStamp(t *testing.T) {
	t.Parallel()

	timeStamp := uint64(100000)
	h := block.Header{}
	h.SetTimeStamp(timeStamp)

	assert.Equal(t, timeStamp, h.GetTimeStamp())
}

func TestHeader_SetTxCount(t *testing.T) {
	t.Parallel()

	txCount := uint32(10)
	h := block.Header{}
	h.SetTxCount(txCount)

	assert.Equal(t, txCount, h.GetTxCount())
}

func TestBody_IntegrityAndValidityNil(t *testing.T) {
	t.Parallel()

	var body *block.Body = nil
	assert.Equal(t, data.ErrNilBlockBody, body.IntegrityAndValidity())
}

func TestBody_IntegrityAndValidityEmptyMiniblockShouldThrowException(t *testing.T) {
	t.Parallel()

	txHash0 := []byte("txHash0")
	mb0 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        [][]byte{txHash0},
	}

	mb1 := block.MiniBlock{}

	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, &mb0)
	body.MiniBlocks = append(body.MiniBlocks, &mb1)

	assert.Equal(t, data.ErrMiniBlockEmpty, body.IntegrityAndValidity())
}

func TestBody_IntegrityAndValidityOK(t *testing.T) {
	t.Parallel()

	txHash0 := []byte("txHash0")
	mb0 := block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
		TxHashes:        [][]byte{txHash0},
	}

	body := &block.Body{}
	body.MiniBlocks = append(body.MiniBlocks, &mb0)

	assert.Equal(t, nil, body.IntegrityAndValidity())
}

func TestHeader_GetMiniBlockHeadersWithDstShouldWork(t *testing.T) {
	t.Parallel()

	hashS0R0 := []byte("hash_0_0")
	hashS0R1 := []byte("hash_0_1")
	hash1S0R2 := []byte("hash_0_2")
	hash2S0R2 := []byte("hash2_0_2")

	hdr := &block.Header{
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				SenderShardID:   0,
				ReceiverShardID: 0,
				Hash:            hashS0R0,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
				Hash:            hashS0R1,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 2,
				Hash:            hash1S0R2,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 2,
				Hash:            hash2S0R2,
			},
		},
	}

	hashesWithDest2 := hdr.GetMiniBlockHeadersWithDst(2)

	assert.Equal(t, uint32(0), hashesWithDest2[string(hash1S0R2)])
	assert.Equal(t, uint32(0), hashesWithDest2[string(hash2S0R2)])
}

func TestHeader_GetOrderedCrossMiniblocksWithDstShouldWork(t *testing.T) {
	t.Parallel()

	hashSh0ToSh0 := []byte("hash_0_0")
	hashSh0ToSh1 := []byte("hash_0_1")
	hash1Sh0ToSh2 := []byte("hash1_0_2")
	hash2Sh0ToSh2 := []byte("hash2_0_2")

	hdr := &block.Header{
		Round: 10,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{
				SenderShardID:   0,
				ReceiverShardID: 0,
				Hash:            hashSh0ToSh0,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 1,
				Hash:            hashSh0ToSh1,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 2,
				Hash:            hash1Sh0ToSh2,
			},
			{
				SenderShardID:   0,
				ReceiverShardID: 2,
				Hash:            hash2Sh0ToSh2,
			},
		},
	}

	miniBlocksInfo := hdr.GetOrderedCrossMiniblocksWithDst(2)

	require.Equal(t, 2, len(miniBlocksInfo))
	assert.Equal(t, hash1Sh0ToSh2, miniBlocksInfo[0].Hash)
	assert.Equal(t, hdr.Round, miniBlocksInfo[0].Round)
	assert.Equal(t, hash2Sh0ToSh2, miniBlocksInfo[1].Hash)
	assert.Equal(t, hdr.Round, miniBlocksInfo[1].Round)
}

func TestMiniBlock_Clone(t *testing.T) {
	t.Parallel()

	miniBlock := &block.MiniBlock{
		TxHashes:        [][]byte{[]byte("something"), []byte("something2")},
		ReceiverShardID: 1,
		SenderShardID:   2,
		Type:            0,
	}

	clonedMB := miniBlock.Clone()

	assert.True(t, reflect.DeepEqual(miniBlock, clonedMB))
}
