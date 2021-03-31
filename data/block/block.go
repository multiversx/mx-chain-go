//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. block.proto
package block

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

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
func (h *Header) SetNonce(n uint64) {
	h.Nonce = n
}

// SetEpoch sets header epoch
func (h *Header) SetEpoch(e uint32) {
	h.Epoch = e
}

// SetRound sets header round
func (h *Header) SetRound(r uint64) {
	h.Round = r
}

// SetRootHash sets root hash
func (h *Header) SetRootHash(rHash []byte) {
	h.RootHash = rHash
}

// SetPrevHash sets prev hash
func (h *Header) SetPrevHash(pvHash []byte) {
	h.PrevHash = pvHash
}

// SetPrevRandSeed sets previous random seed
func (h *Header) SetPrevRandSeed(pvRandSeed []byte) {
	h.PrevRandSeed = pvRandSeed
}

// SetRandSeed sets previous random seed
func (h *Header) SetRandSeed(randSeed []byte) {
	h.RandSeed = randSeed
}

// SetPubKeysBitmap sets publick key bitmap
func (h *Header) SetPubKeysBitmap(pkbm []byte) {
	h.PubKeysBitmap = pkbm
}

// SetSignature sets header signature
func (h *Header) SetSignature(sg []byte) {
	h.Signature = sg
}

// SetLeaderSignature will set the leader's signature
func (h *Header) SetLeaderSignature(sg []byte) {
	h.LeaderSignature = sg
}

// SetChainID sets the chain ID on which this block is valid on
func (h *Header) SetChainID(chainID []byte) {
	h.ChainID = chainID
}

// SetSoftwareVersion sets the software version of the header
func (h *Header) SetSoftwareVersion(version []byte) {
	h.SoftwareVersion = version
}

// SetTimeStamp sets header timestamp
func (h *Header) SetTimeStamp(ts uint64) {
	h.TimeStamp = ts
}

// SetAccumulatedFees sets the accumulated fees in the header
func (h *Header) SetAccumulatedFees(value *big.Int) {
	if h.AccumulatedFees == nil {
		h.AccumulatedFees = big.NewInt(0)
	}
	h.AccumulatedFees.Set(value)
}

// SetDeveloperFees sets the developer fees in the header
func (h *Header) SetDeveloperFees(value *big.Int) {
	if h.DeveloperFees == nil {
		h.DeveloperFees = big.NewInt(0)
	}
	h.DeveloperFees.Set(value)
}

// SetTxCount sets the transaction count of the block associated with this header
func (h *Header) SetTxCount(txCount uint32) {
	h.TxCount = txCount
}

// SetShardID sets header shard ID
func (h *Header) SetShardID(shId uint32) {
	h.ShardID = shId
}

// SetMiniBlockHeaders sets the miniBlock headers
func (h *Header) SetMiniBlockHeaders(mbHeaderHandlers []data.MiniBlockHeaderHandler) {
	if mbHeaderHandlers == nil {
		h.MiniBlockHeaders = nil
		return
	}

	h.MiniBlockHeaders = make([]MiniBlockHeader, len(mbHeaderHandlers))
	for i, mb := range mbHeaderHandlers {
		mbh, ok := mb.(*MiniBlockHeader)
		if !ok || mbh == nil {
			continue
		}
		h.MiniBlockHeaders[i] = *mbh
	}
}

// GetMiniBlockHeadersWithDst as a map of hashes and sender IDs
func (h *Header) GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32 {
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
	result := make([][]byte, 0, len(h.MiniBlockHeaders))
	for _, miniblock := range h.MiniBlockHeaders {
		result = append(result, miniblock.Hash)
	}
	return result
}

// MapMiniBlockHashesToShards is a map of mini block hashes and sender IDs
func (h *Header) MapMiniBlockHashesToShards() map[string]uint32 {
	hashDst := make(map[string]uint32)
	for _, val := range h.MiniBlockHeaders {
		hashDst[string(val.Hash)] = val.SenderShardID
	}
	return hashDst
}

// Clone returns a clone of the object
func (h *Header) ShallowClone() data.HeaderHandler {
	headerCopy := *h
	return &headerCopy
}

// IntegrityAndValidity checks if data is valid
func (b *Body) IntegrityAndValidity() error {
	if b.IsInterfaceNil() {
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
	bodyCopy := *b

	return &bodyCopy
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *Body) IsInterfaceNil() bool {
	return b == nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (h *Header) IsInterfaceNil() bool {
	return h == nil
}

// IsStartOfEpochBlock verifies if the block is of type start of epoch
func (h *Header) IsStartOfEpochBlock() bool {
	return len(h.EpochStartMetaHash) > 0
}

// GetBlockBodyTypeInt32 returns the block body type as int32
func (h *Header) GetBlockBodyTypeInt32() int32 {
	return int32(h.GetBlockBodyType())
}

// GetMiniBlockHeadersHandlers returns the miniBlock headers as an array of miniBlock header handlers
func (h *Header) GetMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler {
	mbHeaders := h.GetMiniBlockHeaders()
	mbHeaderHandlers := make([]data.MiniBlockHeaderHandler, len(mbHeaders))

	for i := range mbHeaders {
		mbHeaderHandlers[i] = &mbHeaders[i]
	}

	return mbHeaderHandlers
}

// SetMiniBlockHeaderHandlers sets the miniBlock headers from the given miniBlock header handlers
func (h *Header) SetMiniBlockHeaderHandlers(mbHeaderHandlers []data.MiniBlockHeaderHandler) {
	if mbHeaderHandlers == nil {
		h.MiniBlockHeaders = nil
		return
	}

	h.MiniBlockHeaders = make([]MiniBlockHeader, len(mbHeaderHandlers))
	for i, mbHeaderHandler := range mbHeaderHandlers {
		mbHeader, ok := mbHeaderHandler.(*MiniBlockHeader)
		if !ok {
			h.MiniBlockHeaders = nil
			return
		}
		h.MiniBlockHeaders[i] = *mbHeader
	}
}

// SetReceiptsHash sets the receipts hash
func (h *Header) SetReceiptsHash(hash []byte) {
	h.ReceiptsHash = hash
}

// SetMetaBlockHashes sets the metaBlock hashes
func (h *Header) SetMetaBlockHashes(hashes [][]byte) {
	h.MetaBlockHashes = hashes
}

// Clone the underlying data
func (mb *MiniBlock) Clone() *MiniBlock {
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
	return len(mb.Reserved) > 0 && mb.Reserved[0] == byte(ScheduledBlock)
}
