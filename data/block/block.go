//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. block.proto
package block

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// MiniBlockSlice should be used when referring to subset of mini blocks that is not
//  necessarily representing a full block body
type MiniBlockSlice []*MiniBlock

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

// SetValidatorStatsRootHash set's the root hash for the validator statistics trie
func (h *Header) SetValidatorStatsRootHash(_ []byte) {
}

// GetValidatorStatsRootHash set's the root hash for the validator statistics trie
func (h *Header) GetValidatorStatsRootHash() []byte {
	return []byte{}
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
	h.AccumulatedFees.Set(value)
}

// SetDeveloperFees sets the developer fees in the header
func (h *Header) SetDeveloperFees(value *big.Int) {
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
func (h *Header) Clone() data.HeaderHandler {
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
