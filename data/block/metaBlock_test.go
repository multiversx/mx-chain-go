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
		PublicKey: []byte("test"),
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
	sd := block.ShardData{
		ShardId:         uint32(10),
		HeaderHash:      []byte("header_hash"),
		TxBlockBodyHash: []byte("tx_block_body_hash"),
	}
	var b bytes.Buffer
	sd.Save(&b)

	loadSd := block.ShardData{}
	loadSd.Load(&b)

	assert.Equal(t, loadSd, sd)
}

func TestMetaBlock_SaveLoad(t *testing.T) {
	pd := block.PeerData{
		PublicKey: []byte("test"),
		Action:    block.PeerRegistrantion,
		TimeStamp: uint64(1234),
		Value:     big.NewInt(1),
	}
	sd := block.ShardData{
		ShardId:         uint32(10),
		HeaderHash:      []byte("header_hash"),
		TxBlockBodyHash: []byte("tx_block_body_hash"),
	}
	mb := block.MetaBlock{
		Nonce:         uint64(1),
		Epoch:         uint32(1),
		Round:         uint32(1),
		ShardInfo:     []block.ShardData{sd},
		PeerInfo:      []block.PeerData{pd},
		Signature:     []byte("signature"),
		PubKeysBitmap: []byte("pub_keys"),
		PreviousHash:  []byte("previous_hash"),
		StateRootHash: []byte("state_root_hash"),
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
	h := block.MetaBlock{
		Epoch: epoch,
	}

	assert.Equal(t, epoch, h.GetEpoch())
}

func TestHeader_GetNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(2)
	h := block.MetaBlock{
		Nonce: nonce,
	}

	assert.Equal(t, nonce, h.GetNonce())
}

func TestHeader_GetPrevHash(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prev hash")
	h := block.MetaBlock{
		PreviousHash: prevHash,
	}

	assert.Equal(t, prevHash, h.GetPrevHash())
}

func TestHeader_GetPubKeysBitmap(t *testing.T) {
	t.Parallel()

	pubKeysBitmap := []byte{10, 11, 12, 13}
	h := block.MetaBlock{
		PubKeysBitmap: pubKeysBitmap,
	}

	assert.Equal(t, pubKeysBitmap, h.GetPubKeysBitmap())
}

func TestHeader_GetRootHash(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	h := block.MetaBlock{
		StateRootHash: rootHash,
	}

	assert.Equal(t, rootHash, h.GetRootHash())
}

func TestHeader_GetRound(t *testing.T) {
	t.Parallel()

	round := uint32(1234)
	h := block.MetaBlock{
		Round: round,
	}

	assert.Equal(t, round, h.GetRound())
}

func TestHeader_GetSignature(t *testing.T) {
	t.Parallel()

	signature := []byte("signature")
	h := block.MetaBlock{
		Signature: signature,
	}

	assert.Equal(t, signature, h.GetSignature())
}

func TestHeader_SetEpoch(t *testing.T) {
	t.Parallel()

	epoch := uint32(10)
	h := block.MetaBlock{}
	h.SetEpoch(epoch)

	assert.Equal(t, epoch, h.GetEpoch())
}

func TestHeader_SetNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(11)
	h := block.MetaBlock{}
	h.SetNonce(nonce)

	assert.Equal(t, nonce, h.GetNonce())
}

func TestHeader_SetPrevHash(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prev hash")
	h := block.MetaBlock{}
	h.SetPrevHash(prevHash)

	assert.Equal(t, prevHash, h.GetPrevHash())
}

func TestHeader_SetPubKeysBitmap(t *testing.T) {
	t.Parallel()

	pubKeysBitmap := []byte{12, 13, 14, 15}
	h := block.MetaBlock{}
	h.SetPubKeysBitmap(pubKeysBitmap)

	assert.Equal(t, pubKeysBitmap, h.GetPubKeysBitmap())
}

func TestHeader_SetRootHash(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	h := block.MetaBlock{}
	h.SetRootHash(rootHash)

	assert.Equal(t, rootHash, h.GetRootHash())
}

func TestHeader_SetRound(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	h := block.MetaBlock{}
	h.SetRootHash(rootHash)

	assert.Equal(t, rootHash, h.GetRootHash())
}

func TestHeader_SetSignature(t *testing.T) {
	t.Parallel()

	signature := []byte("signature")
	h := block.MetaBlock{}
	h.SetSignature(signature)

	assert.Equal(t, signature, h.GetSignature())
}
