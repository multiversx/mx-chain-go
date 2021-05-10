package block_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestHeaderV2_GetEpochNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, uint32(0), h.GetEpoch())

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, uint32(0), h.GetEpoch())
}

func TestHeaderV2_GetEpoch(t *testing.T) {
	t.Parallel()

	epoch := uint32(1)
	h := &block.HeaderV2{
		Header: &block.Header{
			Epoch: epoch,
		},
	}

	require.Equal(t, epoch, h.GetEpoch())
}

func TestHeaderV2_GetShardNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, uint32(0), h.GetShardID())

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, uint32(0), h.GetShardID())
}

func TestHeaderV2_GetShard(t *testing.T) {
	t.Parallel()

	shardId := uint32(2)
	h := &block.HeaderV2{
		Header: &block.Header{
			ShardID: shardId,
		},
	}

	require.Equal(t, shardId, h.GetShardID())
}
func TestHeaderV2_GetNonceNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, uint64(0), h.GetNonce())

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, uint64(0), h.GetNonce())
}

func TestHeaderV2_GetNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(2)
	h := &block.HeaderV2{
		Header: &block.Header{
			Nonce: nonce,
		},
	}

	require.Equal(t, nonce, h.GetNonce())
}

func TestHeaderV2_GetPrevHashNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, []byte(nil), h.GetPrevHash())

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, []byte(nil), h.GetPrevHash())
}

func TestHeaderV2_GetPrevHash(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prev hash")
	h := &block.HeaderV2{
		Header: &block.Header{
			PrevHash: prevHash,
		},
	}

	require.Equal(t, prevHash, h.GetPrevHash())
}

func TestHeaderV2_GetPrevRandSeedNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, []byte(nil), h.GetPrevRandSeed())

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, []byte(nil), h.GetPrevRandSeed())
}

func TestHeaderV2_GetPrevRandSeed(t *testing.T) {
	t.Parallel()

	prevRandSeed := []byte("prev random seed")
	h := &block.HeaderV2{
		Header: &block.Header{
			PrevRandSeed: prevRandSeed,
		},
	}

	require.Equal(t, prevRandSeed, h.GetPrevRandSeed())
}

func TestHeaderV2_GetRandSeedNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, []byte(nil), h.GetRandSeed())

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, []byte(nil), h.GetRandSeed())
}

func TestHeaderV2_GetRandSeed(t *testing.T) {
	t.Parallel()

	randSeed := []byte("random seed")
	h := &block.HeaderV2{
		Header: &block.Header{
			RandSeed: randSeed,
		},
	}

	require.Equal(t, randSeed, h.GetRandSeed())
}
func TestHeaderV2_GetPubKeysBitmapNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, []byte(nil), h.GetPubKeysBitmap())

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, []byte(nil), h.GetPubKeysBitmap())
}

func TestHeaderV2_GetPubKeysBitmap(t *testing.T) {
	t.Parallel()

	pubKeysBitmap := []byte{10, 11, 12, 13}
	h := &block.HeaderV2{
		Header: &block.Header{
			PubKeysBitmap: pubKeysBitmap,
		},
	}

	require.Equal(t, pubKeysBitmap, h.GetPubKeysBitmap())
}

func TestHeaderV2_GetRootHashNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, []byte(nil), h.GetRootHash())

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, []byte(nil), h.GetRootHash())
}

func TestHeaderV2_GetRootHash(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	h := &block.HeaderV2{
		Header: &block.Header{
			RootHash: rootHash,
		},
	}

	require.Equal(t, rootHash, h.GetRootHash())
}

func TestHeaderV2_GetRoundNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, uint64(0), h.GetRound())

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, uint64(0), h.GetRound())
}

func TestHeaderV2_GetRound(t *testing.T) {
	t.Parallel()

	round := uint64(1234)
	h := &block.HeaderV2{
		Header: &block.Header{
			Round: round,
		},
	}

	require.Equal(t, round, h.GetRound())
}

func TestHeaderV2_GetSignatureNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, []byte(nil), h.GetSignature())

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, []byte(nil), h.GetSignature())
}

func TestHeaderV2_GetSignature(t *testing.T) {
	t.Parallel()

	signature := []byte("signature")
	h := &block.HeaderV2{
		Header: &block.Header{
			Signature: signature,
		},
	}

	require.Equal(t, signature, h.GetSignature())
}

func TestHeaderV2_GetTxCountNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, uint32(0), h.GetTxCount())

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, uint32(0), h.GetTxCount())
}

func TestHeaderV2_GetTxCount(t *testing.T) {
	t.Parallel()

	txCount := uint32(10)
	h := &block.HeaderV2{
		Header: &block.Header{
			TxCount: txCount,
		},
	}

	require.Equal(t, txCount, h.GetTxCount())
}

func TestHeaderV2_SetEpochNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, data.ErrNilPointerReceiver, h.SetEpoch(1))

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, data.ErrNilPointerDereference, h.SetEpoch(1))
}

func TestHeaderV2_SetEpoch(t *testing.T) {
	t.Parallel()

	epoch := uint32(10)
	h := &block.HeaderV2{
		Header: &block.Header{},
	}

	err := h.SetEpoch(epoch)
	require.Nil(t, err)
	require.Equal(t, epoch, h.GetEpoch())
}

func TestHeaderV2_SetNonceNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, data.ErrNilPointerReceiver, h.SetNonce(1))

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, data.ErrNilPointerDereference, h.SetNonce(1))
}

func TestHeaderV2_SetNonce(t *testing.T) {
	t.Parallel()

	nonce := uint64(11)
	h := &block.HeaderV2{
		Header: &block.Header{},
	}

	err := h.SetNonce(nonce)
	require.Nil(t, err)
	require.Equal(t, nonce, h.GetNonce())
}

func TestHeaderV2_SetPrevHashNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, data.ErrNilPointerReceiver, h.SetPrevHash([]byte("prev hash")))

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, data.ErrNilPointerDereference, h.SetPrevHash([]byte("prev hash")))
}

func TestHeaderV2_SetPrevHash(t *testing.T) {
	t.Parallel()

	prevHash := []byte("prev hash")
	h := &block.HeaderV2{
		Header: &block.Header{},
	}

	err := h.SetPrevHash(prevHash)
	require.Nil(t, err)
	require.Equal(t, prevHash, h.GetPrevHash())
}

func TestHeaderV2_SetPrevRandSeedNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, data.ErrNilPointerReceiver, h.SetPrevRandSeed([]byte("prev rand seed")))

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, data.ErrNilPointerDereference, h.SetPrevRandSeed([]byte("prev rand seed")))
}

func TestHeaderV2_SetPrevRandSeed(t *testing.T) {
	t.Parallel()

	prevRandSeed := []byte("prev random seed")
	h := &block.HeaderV2{
		Header: &block.Header{},
	}

	err := h.SetPrevRandSeed(prevRandSeed)
	require.Nil(t, err)
	require.Equal(t, prevRandSeed, h.GetPrevRandSeed())
}

func TestHeaderV2_SetRandSeedNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, data.ErrNilPointerReceiver, h.SetRandSeed([]byte("rand seed")))

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, data.ErrNilPointerDereference, h.SetRandSeed([]byte("rand seed")))
}

func TestHeaderV2_SetRandSeed(t *testing.T) {
	t.Parallel()

	randSeed := []byte("random seed")
	h := &block.HeaderV2{
		Header: &block.Header{},
	}

	err := h.SetRandSeed(randSeed)
	require.Nil(t, err)
	require.Equal(t, randSeed, h.GetRandSeed())
}

func TestHeaderV2_SetPubKeysBitmapNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, data.ErrNilPointerReceiver, h.SetPubKeysBitmap([]byte("pub key bitmap")))

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, data.ErrNilPointerDereference, h.SetPubKeysBitmap([]byte("pub key bitmap")))
}

func TestHeaderV2_SetPubKeysBitmap(t *testing.T) {
	t.Parallel()

	pubKeysBitmap := []byte{12, 13, 14, 15}
	h := &block.HeaderV2{
		Header: &block.Header{},
	}

	err := h.SetPubKeysBitmap(pubKeysBitmap)
	require.Nil(t, err)
	require.Equal(t, pubKeysBitmap, h.GetPubKeysBitmap())
}

func TestHeaderV2_SetRootHashNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, data.ErrNilPointerReceiver, h.SetRootHash([]byte("root hash")))

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, data.ErrNilPointerDereference, h.SetRootHash([]byte("root hash")))
}

func TestHeaderV2_SetRootHash(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	h := &block.HeaderV2{
		Header: &block.Header{},
	}

	err := h.SetRootHash(rootHash)
	require.Nil(t, err)
	require.Equal(t, rootHash, h.GetRootHash())
}

func TestHeaderV2_SetRoundNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, data.ErrNilPointerReceiver, h.SetRound(1))

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, data.ErrNilPointerDereference, h.SetRound(1))
}

func TestHeaderV2_SetRound(t *testing.T) {
	t.Parallel()

	rootHash := []byte("root hash")
	h := &block.HeaderV2{
		Header: &block.Header{},
	}

	err := h.SetRootHash(rootHash)
	require.Nil(t, err)
	require.Equal(t, rootHash, h.GetRootHash())
}

func TestHeaderV2_SetSignatureNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, data.ErrNilPointerReceiver, h.SetSignature([]byte("signature")))

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, data.ErrNilPointerDereference, h.SetSignature([]byte("signature")))
}

func TestHeaderV2_SetSignature(t *testing.T) {
	t.Parallel()

	signature := []byte("signature")
	h := &block.HeaderV2{
		Header: &block.Header{},
	}

	err := h.SetSignature(signature)
	require.Nil(t, err)
	require.Equal(t, signature, h.GetSignature())
}

func TestHeaderV2_SetTimeStampNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, data.ErrNilPointerReceiver, h.SetTimeStamp(100000))

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, data.ErrNilPointerDereference, h.SetTimeStamp(100000))
}

func TestHeaderV2_SetTimeStamp(t *testing.T) {
	t.Parallel()

	timeStamp := uint64(100000)
	h := &block.HeaderV2{
		Header: &block.Header{},
	}

	err := h.SetTimeStamp(timeStamp)
	require.Nil(t, err)
	require.Equal(t, timeStamp, h.GetTimeStamp())
}

func TestHeaderV2_SetTxCountNilPointerReceiverOrInnerHeader(t *testing.T) {
	t.Parallel()

	var h *block.HeaderV2
	require.Equal(t, data.ErrNilPointerReceiver, h.SetTxCount(10000))

	h = &block.HeaderV2{
		Header: nil,
	}
	require.Equal(t, data.ErrNilPointerDereference, h.SetTxCount(10000))
}

func TestHeaderV2_SetTxCount(t *testing.T) {
	t.Parallel()

	txCount := uint32(10)
	h := &block.HeaderV2{
		Header: &block.Header{},
	}

	err := h.SetTxCount(txCount)
	require.Nil(t, err)
	require.Equal(t, txCount, h.GetTxCount())
}

func TestHeaderV2_GetMiniBlockHeadersWithDstShouldWork(t *testing.T) {
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

	require.Equal(t, uint32(0), hashesWithDest2[string(hash1S0R2)])
	require.Equal(t, uint32(0), hashesWithDest2[string(hash2S0R2)])
}

func TestHeaderV2_GetOrderedCrossMiniblocksWithDstShouldWork(t *testing.T) {
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
	require.Equal(t, hash1Sh0ToSh2, miniBlocksInfo[0].Hash)
	require.Equal(t, hdr.Round, miniBlocksInfo[0].Round)
	require.Equal(t, hash2Sh0ToSh2, miniBlocksInfo[1].Hash)
	require.Equal(t, hdr.Round, miniBlocksInfo[1].Round)
}

func TestHeaderV2_SetScheduledRootHash(t *testing.T) {
	t.Parallel()

	hv2 := block.HeaderV2{
		Header: &block.Header{},
	}
	require.Nil(t, hv2.ScheduledRootHash)

	rootHash := []byte("root hash")
	err := hv2.SetScheduledRootHash(rootHash)
	require.Nil(t, err)
	require.Equal(t, rootHash, hv2.ScheduledRootHash)
}

func TestHeaderV2_ValidateHeaderVersion(t *testing.T) {
	t.Parallel()

	hv2 := block.HeaderV2{
		Header: &block.Header{},
	}

	err := hv2.ValidateHeaderVersion()
	require.Equal(t, data.ErrNilScheduledRootHash, err)

	hv2.ScheduledRootHash = make([]byte, 0)
	err = hv2.ValidateHeaderVersion()
	require.Equal(t, data.ErrNilScheduledRootHash, err)

	hv2.ScheduledRootHash = make([]byte, 32)
	err = hv2.ValidateHeaderVersion()
	require.Nil(t, err)
}
