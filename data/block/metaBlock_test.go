package block_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/stretchr/testify/assert"
)

func TestPeerData_SaveLoad(t *testing.T) {
	pd := block.PeerData{
		PublicKey: []byte("public key"),
		Action:    block.PeerRegistrantion,
		TimeStamp: uint64(1234),
		Value:     big.NewInt(1),
	}
	var b bytes.Buffer
	pd.Save(&b)

	loadPd := block.PeerData{}
	loadPd.Load(&b)

	assert.Equal(t, loadPd, pd)
}

func TestShardData_SaveLoad(t *testing.T) {

	mbh := block.ShardMiniBlockHeader{
		Hash:            []byte("miniblock hash"),
		SenderShardId:   uint32(0),
		ReceiverShardId: uint32(1),
		TxCount:         uint32(1),
	}

	sd := block.ShardData{
		ShardId:               uint32(10),
		HeaderHash:            []byte("header_hash"),
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{mbh},
		TxCount:               uint32(1),
	}

	var b bytes.Buffer
	sd.Save(&b)

	loadSd := block.ShardData{}
	loadSd.Load(&b)

	assert.Equal(t, loadSd, sd)
}

func TestMetaBlock_SaveLoad(t *testing.T) {
	pd := block.PeerData{
		PublicKey: []byte("public key"),
		Action:    block.PeerRegistrantion,
		TimeStamp: uint64(1234),
		Value:     big.NewInt(1),
	}

	mbh := block.ShardMiniBlockHeader{
		Hash:            []byte("miniblock hash"),
		SenderShardId:   uint32(0),
		ReceiverShardId: uint32(1),
		TxCount:         uint32(1),
	}

	sd := block.ShardData{
		ShardId:               uint32(10),
		HeaderHash:            []byte("header_hash"),
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{mbh},
		TxCount:               uint32(1),
	}

	mb := block.MetaBlock{
		Nonce:         uint64(1),
		Epoch:         uint32(1),
		Round:         uint32(1),
		TimeStamp:     uint64(100000),
		ShardInfo:     []block.ShardData{sd},
		PeerInfo:      []block.PeerData{pd},
		Signature:     []byte("signature"),
		PubKeysBitmap: []byte("pub keys"),
		PrevHash:      []byte("previous hash"),
		PrevRandSeed:  []byte("previous random seed"),
		RandSeed:      []byte("random seed"),
		RootHash:      []byte("root hash"),
		TxCount:       uint32(1),
	}
	var b bytes.Buffer
	mb.Save(&b)

	loadMb := block.MetaBlock{}
	loadMb.Load(&b)

	assert.Equal(t, loadMb, mb)
}

func TestMetaBlock_GetEpoch(t *testing.T) {
	t.Parallel()

	epoch := uint32(1)
	m := block.MetaBlock{
		Epoch: epoch,
	}

	assert.Equal(t, epoch, m.GetEpoch())
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

	round := uint32(1234)
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
	m.SetEpoch(epoch)

	assert.Equal(t, epoch, m.GetEpoch())
}

func TestMetaBlock_SetNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(11)
	m := block.MetaBlock{}
	m.SetNonce(nonce)

	assert.Equal(t, nonce, m.GetNonce())
}

func TestMetaBlock_SetPrevHash(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prev hash")
	m := block.MetaBlock{}
	m.SetPrevHash(prevHash)

	assert.Equal(t, prevHash, m.GetPrevHash())
}

func TestMetaBlock_SetPubKeysBitmap(t *testing.T) {
	t.Parallel()

	pubKeysBitmap := []byte{12, 13, 14, 15}
	m := block.MetaBlock{}
	m.SetPubKeysBitmap(pubKeysBitmap)

	assert.Equal(t, pubKeysBitmap, m.GetPubKeysBitmap())
}

func TestMetaBlock_SetPrevRandSeed(t *testing.T) {
	t.Parallel()

	prevRandSeed := []byte("previous random seed")
	m := block.MetaBlock{}
	m.SetPrevRandSeed(prevRandSeed)

	assert.Equal(t, prevRandSeed, m.GetPrevRandSeed())
}

func TestMetaBlock_SetRandSeed(t *testing.T) {
	t.Parallel()

	randSeed := []byte("random seed")
	m := block.MetaBlock{}
	m.SetRandSeed(randSeed)

	assert.Equal(t, randSeed, m.GetRandSeed())
}

func TestMetaBlock_SetRootHash(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	m := block.MetaBlock{}
	m.SetRootHash(rootHash)

	assert.Equal(t, rootHash, m.GetRootHash())
}

func TestMetaBlock_SetRound(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	m := block.MetaBlock{}
	m.SetRootHash(rootHash)

	assert.Equal(t, rootHash, m.GetRootHash())
}

func TestMetaBlock_SetSignature(t *testing.T) {
	t.Parallel()

	signature := []byte("signature")
	m := block.MetaBlock{}
	m.SetSignature(signature)

	assert.Equal(t, signature, m.GetSignature())
}

func TestMetaBlock_SetTimeStamp(t *testing.T) {
	t.Parallel()

	timestamp := uint64(100000)
	m := block.MetaBlock{}
	m.SetTimeStamp(timestamp)

	assert.Equal(t, timestamp, m.GetTimeStamp())
}

func TestMetaBlock_SetTxCount(t *testing.T) {
	t.Parallel()

	txCount := uint32(100)
	m := block.MetaBlock{}
	m.SetTxCount(txCount)

	assert.Equal(t, txCount, m.GetTxCount())
}
