//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. block.proto
package block

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// TODO: add access protection through wrappers
var _ = data.HeaderHandler(&Header{})
var _ = data.ShardHeaderHandler(&Header{})

// MiniBlockSlice should be used when referring to subset of mini blocks that is not
//  necessarily representing a full block body
type MiniBlockSlice []*MiniBlock

// MiniblockAndHash holds the info related to a miniblock and its hash
type MiniblockAndHash struct {
	Miniblock *MiniBlock
	Hash      []byte
}

// SetNonce sets header nonce
func (h *Header) SetNonce(n uint64) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.Nonce = n
	return nil
}

// SetEpoch sets header epoch
func (h *Header) SetEpoch(e uint32) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.Epoch = e
	return nil
}

// SetRound sets header round
func (h *Header) SetRound(r uint64) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.Round = r
	return nil
}

// SetRootHash sets root hash
func (h *Header) SetRootHash(rHash []byte) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.RootHash = rHash
	return nil
}

// SetPrevHash sets prev hash
func (h *Header) SetPrevHash(pvHash []byte) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.PrevHash = pvHash
	return nil
}

// SetPrevRandSeed sets previous random seed
func (h *Header) SetPrevRandSeed(pvRandSeed []byte) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.PrevRandSeed = pvRandSeed
	return nil
}

// SetRandSeed sets previous random seed
func (h *Header) SetRandSeed(randSeed []byte) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.RandSeed = randSeed
	return nil
}

// SetPubKeysBitmap sets public key bitmap
func (h *Header) SetPubKeysBitmap(pkbm []byte) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.PubKeysBitmap = pkbm
	return nil
}

// SetSignature sets header signature
func (h *Header) SetSignature(sg []byte) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.Signature = sg
	return nil
}

// SetLeaderSignature will set the leader's signature
func (h *Header) SetLeaderSignature(sg []byte) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.LeaderSignature = sg
	return nil
}

// SetChainID sets the chain ID on which this block is valid on
func (h *Header) SetChainID(chainID []byte) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.ChainID = chainID
	return nil
}

// SetSoftwareVersion sets the software version of the header
func (h *Header) SetSoftwareVersion(version []byte) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.SoftwareVersion = version
	return nil
}

// SetTimeStamp sets header timestamp
func (h *Header) SetTimeStamp(ts uint64) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.TimeStamp = ts
	return nil
}

// SetAccumulatedFees sets the accumulated fees in the header
func (h *Header) SetAccumulatedFees(value *big.Int) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}
	if h.AccumulatedFees == nil {
		h.AccumulatedFees = big.NewInt(0)
	}
	h.AccumulatedFees.Set(value)
	return nil
}

// SetDeveloperFees sets the developer fees in the header
func (h *Header) SetDeveloperFees(value *big.Int) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}
	if h.DeveloperFees == nil {
		h.DeveloperFees = big.NewInt(0)
	}
	h.DeveloperFees.Set(value)
	return nil
}

// SetTxCount sets the transaction count of the block associated with this header
func (h *Header) SetTxCount(txCount uint32) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.TxCount = txCount
	return nil
}

// SetShardID sets header shard ID
func (h *Header) SetShardID(shId uint32) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}

	h.ShardID = shId
	return nil
}

// GetMiniBlockHeadersWithDst as a map of hashes and sender IDs
func (h *Header) GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32 {
	if h == nil {
		return nil
	}

	hashDst := make(map[string]uint32)
	for _, val := range h.MiniBlockHeaders {
		if val.ReceiverShardID == destId && val.SenderShardID != destId {
			hashDst[string(val.Hash)] = val.SenderShardID
		}
	}
	return hashDst
}

// GetOrderedCrossMiniblocksWithDst gets all cross miniblocks with the given destination shard ID, ordered in a
// chronological way, taking into consideration the round in which they were created/executed in the sender shard
func (h *Header) GetOrderedCrossMiniblocksWithDst(destId uint32) []*data.MiniBlockInfo {
	if h == nil {
		return nil
	}
	miniBlocks := make([]*data.MiniBlockInfo, 0)

	for _, mb := range h.MiniBlockHeaders {
		if mb.ReceiverShardID == destId && mb.SenderShardID != destId {
			miniBlocks = append(miniBlocks, &data.MiniBlockInfo{
				Hash:          mb.Hash,
				SenderShardID: mb.SenderShardID,
				Round:         h.Round,
			})
		}
	}

	return miniBlocks
}

// GetMiniBlockHeadersHashes gets the miniblock hashes
func (h *Header) GetMiniBlockHeadersHashes() [][]byte {
	if h == nil {
		return nil
	}

	result := make([][]byte, 0, len(h.MiniBlockHeaders))
	for _, miniblock := range h.MiniBlockHeaders {
		result = append(result, miniblock.Hash)
	}
	return result
}

// MapMiniBlockHashesToShards is a map of mini block hashes and sender IDs
func (h *Header) MapMiniBlockHashesToShards() map[string]uint32 {
	if h == nil {
		return nil
	}

	hashDst := make(map[string]uint32)
	for _, val := range h.MiniBlockHeaders {
		hashDst[string(val.Hash)] = val.SenderShardID
	}
	return hashDst
}

// ShallowClone returns a clone of the object
func (h *Header) ShallowClone() data.HeaderHandler {
	if h == nil {
		return nil
	}

	headerCopy := *h
	return &headerCopy
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *Header) IsInterfaceNil() bool {
	return h == nil
}

// IsStartOfEpochBlock verifies if the block is of type start of epoch
func (h *Header) IsStartOfEpochBlock() bool {
	if h == nil {
		return false
	}

	return len(h.EpochStartMetaHash) > 0
}

// GetBlockBodyTypeInt32 returns the block body type as int32
func (h *Header) GetBlockBodyTypeInt32() int32 {
	if h == nil {
		return -1
	}

	return int32(h.GetBlockBodyType())
}

// GetMiniBlockHeadersHandlers returns the miniBlock headers as an array of miniBlock header handlers
func (h *Header) GetMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler {
	if h == nil {
		return nil
	}

	mbHeaders := h.GetMiniBlockHeaders()
	mbHeaderHandlers := make([]data.MiniBlockHeaderHandler, len(mbHeaders))

	for i := range mbHeaders {
		mbHeaderHandlers[i] = &mbHeaders[i]
	}

	return mbHeaderHandlers
}

// SetMiniBlockHeaderHandlers sets the miniBlock headers from the given miniBlock header handlers
func (h *Header) SetMiniBlockHeaderHandlers(mbHeaderHandlers []data.MiniBlockHeaderHandler) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}
	if mbHeaderHandlers == nil {
		h.MiniBlockHeaders = nil
		return nil
	}

	miniBlockHeaders := make([]MiniBlockHeader, len(mbHeaderHandlers))
	h.MiniBlockHeaders = make([]MiniBlockHeader, len(mbHeaderHandlers))
	for i, mbHeaderHandler := range mbHeaderHandlers {
		mbHeader, ok := mbHeaderHandler.(*MiniBlockHeader)
		if !ok {
			return data.ErrInvalidTypeAssertion
		}
		if mbHeader == nil {
			return data.ErrNilPointerDereference
		}
		miniBlockHeaders[i] = *mbHeader
	}

	h.MiniBlockHeaders = miniBlockHeaders
	return nil
}

// SetReceiptsHash sets the receipts hash
func (h *Header) SetReceiptsHash(hash []byte) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}
	h.ReceiptsHash = hash
	return nil
}

// SetMetaBlockHashes sets the metaBlock hashes
func (h *Header) SetMetaBlockHashes(hashes [][]byte) error {
	if h == nil {
		return data.ErrNilPointerReceiver
	}
	h.MetaBlockHashes = hashes
	return nil
}

// IntegrityAndValidity checks if data is valid
func (b *Body) IntegrityAndValidity() error {
	if b == nil {
		return data.ErrNilBlockBody
	}

	for i := 0; i < len(b.MiniBlocks); i++ {
		if len(b.MiniBlocks[i].TxHashes) == 0 {
			return data.ErrMiniBlockEmpty
		}
	}

	return nil
}

// Clone returns a clone of the object
func (b *Body) Clone() data.BodyHandler {
	if b == nil {
		return nil
	}

	bodyCopy := *b

	return &bodyCopy
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *Body) IsInterfaceNil() bool {
	return b == nil
}

// Clone the underlying data
func (mb *MiniBlock) Clone() *MiniBlock {
	if mb == nil {
		return nil
	}

	newMb := &MiniBlock{
		ReceiverShardID: mb.ReceiverShardID,
		SenderShardID:   mb.SenderShardID,
		Type:            mb.Type,
	}
	newMb.TxHashes = make([][]byte, len(mb.TxHashes))
	copy(newMb.TxHashes, mb.TxHashes)

	return newMb
}

// IsScheduledMiniBlock returns if the mini block is of type scheduled or not
func (mb *MiniBlock) IsScheduledMiniBlock() bool {
	if mb == nil {
		return false
	}
	return len(mb.Reserved) > 0 && mb.Reserved[0] == byte(ScheduledBlock)
}
