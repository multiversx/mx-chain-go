package block_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/assert"
)

func TestHeader_SaveLoad(t *testing.T) {
	t.Parallel()

	mb := block.MiniBlockHeader{
		Hash:            []byte("mini block hash"),
		ReceiverShardID: uint32(0),
		SenderShardID:   uint32(10),
		TxCount:         uint32(10),
	}

	pc := block.PeerChange{
		PubKey:      []byte("peer1"),
		ShardIdDest: uint32(0),
	}

	h := block.Header{
		Nonce:              uint64(1),
		PrevHash:           []byte("previous hash"),
		PrevRandSeed:       []byte("prev random seed"),
		RandSeed:           []byte("current random seed"),
		PubKeysBitmap:      []byte("pub key bitmap"),
		ShardId:            uint32(10),
		TimeStamp:          uint64(1234),
		Round:              uint64(1),
		Epoch:              uint32(1),
		BlockBodyType:      block.TxBlock,
		Signature:          []byte("signature"),
		MiniBlockHeaders:   []block.MiniBlockHeader{mb},
		PeerChanges:        []block.PeerChange{pc},
		RootHash:           []byte("root hash"),
		MetaBlockHashes:    make([][]byte, 0),
		TxCount:            uint32(10),
		EpochStartMetaHash: []byte("epochStart"),
		LeaderSignature:    []byte("leader_sig"),
		ChainID:            []byte("chain ID"),
	}

	var b bytes.Buffer
	err := h.Save(&b)
	assert.Nil(t, err)

	loadHeader := block.Header{}
	err = loadHeader.Load(&b)
	assert.Nil(t, err)

	assert.Equal(t, loadHeader, h)
}

func TestPeerChange_SaveLoad(t *testing.T) {
	t.Parallel()

	pc := block.PeerChange{
		PubKey:      []byte("pubKey"),
		ShardIdDest: uint32(1),
	}

	var b bytes.Buffer
	err := pc.Save(&b)
	assert.Nil(t, err)

	loadPc := block.PeerChange{}
	err = loadPc.Load(&b)
	assert.Nil(t, err)

	assert.Equal(t, loadPc, pc)
}

func TestMiniBlockHeader_SaveLoad(t *testing.T) {
	t.Parallel()

	mbh := block.MiniBlockHeader{
		Hash:            []byte("mini block hash"),
		SenderShardID:   uint32(1),
		ReceiverShardID: uint32(0),
	}

	var b bytes.Buffer
	err := mbh.Save(&b)
	assert.Nil(t, err)

	loadMbh := block.MiniBlockHeader{}
	err = loadMbh.Load(&b)
	assert.Nil(t, err)

	assert.Equal(t, loadMbh, mbh)
}

func TestMiniBlock_SaveLoad(t *testing.T) {
	t.Parallel()

	mb := block.MiniBlock{
		TxHashes:        [][]byte{[]byte("tx hash1"), []byte("tx hash2")},
		ReceiverShardID: uint32(0),
		SenderShardID:   uint32(0),
	}

	var b bytes.Buffer
	err := mb.Save(&b)
	assert.Nil(t, err)

	loadMb := block.MiniBlock{}
	err = loadMb.Load(&b)
	assert.Nil(t, err)

	assert.Equal(t, loadMb, mb)
}

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
		ShardId: shardId,
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

	body := block.Body{}
	body = nil
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

	body := make(block.Body, 0)
	body = append(body, &mb0)
	body = append(body, &mb1)

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

	body := make(block.Body, 0)
	body = append(body, &mb0)

	assert.Equal(t, nil, body.IntegrityAndValidity())
}

func TestHeader_GetMiniBlockHeadersWithDstShouldWork(t *testing.T) {
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

func TestHeader_CheckChainID(t *testing.T) {
	t.Parallel()

	chainID := []byte("chainID")
	okChainID := []byte("chainID")
	wrongChainID := []byte("wrong chain ID")
	hdr := &block.Header{
		ChainID: chainID,
	}

	assert.Nil(t, hdr.CheckChainID(okChainID))
	assert.True(t, errors.Is(hdr.CheckChainID(wrongChainID), data.ErrInvalidChainID))
}
