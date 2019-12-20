package block_test

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestPeerData_SaveLoad(t *testing.T) {
	pd := block.PeerData{
		PublicKey:   []byte("public key"),
		Action:      block.PeerRegistrantion,
		TimeStamp:   uint64(1234),
		ValueChange: big.NewInt(1),
		Address:     []byte("address"),
	}
	var b bytes.Buffer
	_ = pd.Save(&b)

	loadPd := block.PeerData{}
	_ = loadPd.Load(&b)

	assert.Equal(t, loadPd, pd)
}

func TestShardData_SaveLoad(t *testing.T) {

	mbh := block.ShardMiniBlockHeader{
		Hash:            []byte("miniblock hash"),
		SenderShardID:   uint32(0),
		ReceiverShardID: uint32(1),
		TxCount:         uint32(1),
	}

	sd := block.ShardData{
		ShardID:               uint32(10),
		HeaderHash:            []byte("header_hash"),
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{mbh},
		PubKeysBitmap:         []byte{1},
		Signature:             []byte{1},
		PrevRandSeed:          []byte{1},
		TxCount:               uint32(1),
	}

	var b bytes.Buffer
	_ = sd.Save(&b)

	loadSd := block.ShardData{}
	_ = loadSd.Load(&b)

	assert.Equal(t, loadSd, sd)
}

func TestEpochStart_SaveLoad(t *testing.T) {

	mbh := block.ShardMiniBlockHeader{
		Hash:            []byte("miniblock hash"),
		SenderShardID:   uint32(0),
		ReceiverShardID: uint32(1),
		TxCount:         uint32(1),
	}

	lastFinalHdr := block.EpochStartShardData{
		ShardId:                 0,
		HeaderHash:              []byte("headerhash"),
		RootHash:                []byte("roothash"),
		FirstPendingMetaBlock:   []byte("firstPending"),
		LastFinishedMetaBlock:   []byte("lastfinished"),
		PendingMiniBlockHeaders: []block.ShardMiniBlockHeader{mbh},
	}

	epochStart := block.EpochStart{
		LastFinalizedHeaders: []block.EpochStartShardData{lastFinalHdr},
	}

	var b bytes.Buffer
	_ = epochStart.Save(&b)

	loadEpoch := block.EpochStart{}
	_ = loadEpoch.Load(&b)

	assert.Equal(t, loadEpoch, epochStart)
}

func TestMetaBlock_SaveLoad(t *testing.T) {
	pd := block.PeerData{
		Address:     []byte("address"),
		PublicKey:   []byte("public key"),
		Action:      block.PeerRegistrantion,
		TimeStamp:   uint64(1234),
		ValueChange: big.NewInt(1),
	}

	mbh := block.ShardMiniBlockHeader{
		Hash:            []byte("miniblock hash"),
		SenderShardID:   uint32(0),
		ReceiverShardID: uint32(1),
		TxCount:         uint32(1),
	}

	sd := block.ShardData{
		ShardID:               uint32(10),
		HeaderHash:            []byte("header_hash"),
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{mbh},
		TxCount:               uint32(1),
		PubKeysBitmap:         []byte("pkb"),
		Signature:             []byte("sig"),
		PrevRandSeed:          []byte("rand"),
	}

	mbHdr := block.MiniBlockHeader{
		Hash:            []byte("mini block hash"),
		ReceiverShardID: uint32(0),
		SenderShardID:   uint32(10),
		TxCount:         uint32(10),
	}

	lastFinalHdr := block.EpochStartShardData{
		ShardId:                 0,
		HeaderHash:              []byte("headerhash"),
		RootHash:                []byte("roothash"),
		FirstPendingMetaBlock:   []byte("firstPending"),
		LastFinishedMetaBlock:   []byte("lastfinished"),
		PendingMiniBlockHeaders: []block.ShardMiniBlockHeader{mbh},
	}

	mb := block.MetaBlock{
		Nonce:                  uint64(1),
		Epoch:                  uint32(1),
		Round:                  uint64(1),
		TimeStamp:              uint64(100000),
		ShardInfo:              []block.ShardData{sd},
		PeerInfo:               []block.PeerData{pd},
		Signature:              []byte("signature"),
		PubKeysBitmap:          []byte("pub keys"),
		PrevHash:               []byte("previous hash"),
		PrevRandSeed:           []byte("previous random seed"),
		RandSeed:               []byte("random seed"),
		RootHash:               []byte("root hash"),
		TxCount:                uint32(1),
		ValidatorStatsRootHash: []byte("rootHash"),
		MiniBlockHeaders:       []block.MiniBlockHeader{mbHdr},
		LeaderSignature:        []byte("leader_sign"),
		ChainID:                []byte("chain ID"),
		EpochStart: block.EpochStart{
			LastFinalizedHeaders: []block.EpochStartShardData{lastFinalHdr},
		},
	}
	var b bytes.Buffer
	err := mb.Save(&b)
	assert.Nil(t, err)

	loadMb := block.MetaBlock{}
	err = loadMb.Load(&b)
	assert.Nil(t, err)

	assert.Equal(t, mb, loadMb)
}

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

	assert.Equal(t, sharding.MetachainShardId, m.GetShardID())
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

func TestMetaBlock_GetMiniBlockHeadersWithDst(t *testing.T) {
	t.Parallel()

	metaHdr := &block.MetaBlock{Round: 15}
	metaHdr.ShardInfo = make([]block.ShardData, 0)

	shardMBHeader := make([]block.ShardMiniBlockHeader, 0)
	shMBHdr1 := block.ShardMiniBlockHeader{SenderShardID: 0, ReceiverShardID: 1, Hash: []byte("hash1")}
	shMBHdr2 := block.ShardMiniBlockHeader{SenderShardID: 0, ReceiverShardID: 1, Hash: []byte("hash2")}
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

func TestMetaBlock_IsChainIDValid(t *testing.T) {
	t.Parallel()

	chainID := []byte("chainID")
	okChainID := []byte("chainID")
	wrongChainID := []byte("wrong chain ID")
	metablock := &block.MetaBlock{
		ChainID: chainID,
	}

	assert.Nil(t, metablock.CheckChainID(okChainID))
	assert.True(t, errors.Is(metablock.CheckChainID(wrongChainID), data.ErrInvalidChainID))
}
