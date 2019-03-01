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
		ShardId: uint32(10),
		HeaderHash: []byte("header_hash"),
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
		ShardId: uint32(10),
		HeaderHash: []byte("header_hash"),
		TxBlockBodyHash: []byte("tx_block_body_hash"),
	}
	mb := block.MetaBlock{
		Nonce: uint64(1),
		Epoch: uint32(1),
		Round: uint32(1),
		ShardInfo: []block.ShardData{sd},
		PeerInfo: []block.PeerData{pd},
		Signature: []byte("signature"),
		PubKeysBitmap: []byte("pub_keys"),
		PreviousHash: []byte("previous_hash"),
		StateRootHash: []byte("state_root_hash"),
	}
	var b bytes.Buffer
	mb.Save(&b)

	loadMb := block.MetaBlock{}
	loadMb.Load(&b)

	assert.Equal(t, loadMb, mb)
}
