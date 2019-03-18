package block_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)

func TestHeader_SaveLoad(t *testing.T) {
	t.Parallel()

	mb := block.MiniBlockHeader{
		Hash:            []byte("mini block hash"),
		ReceiverShardID: uint32(0),
		SenderShardID:   uint32(10),
	}

	pc := block.PeerChange{
		PubKey:      []byte("peer1"),
		ShardIdDest: uint32(0),
	}

	h := block.Header{
		Nonce:            uint64(1),
		PrevHash:         []byte("previous hash"),
		PrevRandSeed:     []byte("prev random seed"),
		RandSeed:         []byte("current random seed"),
		PubKeysBitmap:    []byte("pub key bitmap"),
		ShardId:          uint32(10),
		TimeStamp:        uint64(1234),
		Round:            uint32(1),
		Epoch:            uint32(1),
		BlockBodyType:    block.TxBlock,
		Signature:        []byte("signature"),
		MiniBlockHeaders: []block.MiniBlockHeader{mb},
		PeerChanges:      []block.PeerChange{pc},
		RootHash:         []byte("root hash"),
	}

	var b bytes.Buffer
	h.Save(&b)

	loadHeader := block.Header{}
	loadHeader.Load(&b)

	assert.Equal(t, loadHeader, h)
}

func TestPeerChange_SaveLoad(t *testing.T) {
	t.Parallel()

	pc := block.PeerChange{
		PubKey:      []byte("pubKey"),
		ShardIdDest: uint32(1),
	}

	var b bytes.Buffer
	pc.Save(&b)

	loadPc := block.PeerChange{}
	loadPc.Load(&b)

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
	mbh.Save(&b)

	loadMbh := block.MiniBlockHeader{}
	loadMbh.Load(&b)

	assert.Equal(t, loadMbh, mbh)
}

func TestMiniBlock_SaveLoad(t *testing.T) {
	t.Parallel()

	mb := block.MiniBlock{
		TxHashes: [][]byte{[]byte("tx hash1"), []byte("tx hash2")},
		ShardID:  uint32(0),
	}

	var b bytes.Buffer
	mb.Save(&b)

	loadMb := block.MiniBlock{}
	loadMb.Load(&b)

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

	round := uint32(1234)
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

	assert.Equal(t, timeStamp, h.GetTimestamp())
}


func Test_test(t *testing.T){
	var list []string
	list = nil

	for i:= range list {
		fmt.Println("entered ", i)
	}
}