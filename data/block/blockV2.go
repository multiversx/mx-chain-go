//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. blockV2.proto
package block

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// GetShardID returns the header shardID
func (hv2 *HeaderV2) GetShardID() uint32 {
	if hv2 == nil || hv2.Header == nil {
		return 0
	}

	return hv2.Header.ShardID
}

// GetNonce returns the header nonce
func (hv2 *HeaderV2) GetNonce() uint64 {
	if hv2 == nil || hv2.Header == nil {
		return 0
	}

	return hv2.Header.Nonce
}

// GetEpoch returns the header epoch
func (hv2 *HeaderV2) GetEpoch() uint32 {
	if hv2 == nil || hv2.Header == nil {
		return 0
	}

	return hv2.Header.Epoch
}

// GetRound returns the header round
func (hv2 *HeaderV2) GetRound() uint64 {
	if hv2 == nil || hv2.Header == nil {
		return 0
	}

	return hv2.Header.Round
}

// GetRootHash returns the header root hash
func (hv2 *HeaderV2) GetRootHash() []byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.RootHash
}

// GetPrevHash returns the header previous header hash
func (hv2 *HeaderV2) GetPrevHash() []byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.PrevHash
}

// GetPrevRandSeed returns the header previous random seed
func (hv2 *HeaderV2) GetPrevRandSeed() []byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.PrevRandSeed
}

// GetRandSeed returns the header random seed
func (hv2 *HeaderV2) GetRandSeed() []byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.RandSeed
}

// GetPubKeysBitmap returns the header public key bitmap for the aggregated signatures
func (hv2 *HeaderV2) GetPubKeysBitmap() []byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.PubKeysBitmap
}

// GetSignature returns the header aggregated signature
func (hv2 *HeaderV2) GetSignature() []byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.Signature
}

// GetLeaderSignature returns the leader signature on top of the finalized (signed) header
func (hv2 *HeaderV2) GetLeaderSignature() []byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.LeaderSignature
}

// GetChainID returns the chain ID
func (hv2 *HeaderV2) GetChainID() []byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.ChainID
}

// GetSoftwareVersion returns the header software version
func (hv2 *HeaderV2) GetSoftwareVersion() []byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.SoftwareVersion
}

// GetTimeStamp returns the header timestamp
func (hv2 *HeaderV2) GetTimeStamp() uint64 {
	if hv2 == nil || hv2.Header == nil {
		return 0
	}

	return hv2.Header.TimeStamp
}

// GetTxCount returns the number of txs included in the block
func (hv2 *HeaderV2) GetTxCount() uint32 {
	if hv2 == nil || hv2.Header == nil {
		return 0
	}

	return hv2.Header.TxCount
}

// GetReceiptsHash returns the header receipt hash
func (hv2 *HeaderV2) GetReceiptsHash() []byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.ReceiptsHash
}

// GetAccumulatedFees returns the block accumulated fees
func (hv2 *HeaderV2) GetAccumulatedFees() *big.Int {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.AccumulatedFees
}

// GetDeveloperFees returns the block developer fees
func (hv2 *HeaderV2) GetDeveloperFees() *big.Int {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.DeveloperFees
}

// GetReserved returns the reserved field
func (hv2 *HeaderV2) GetReserved() []byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.Reserved
}

// GetMetaBlockHashes returns the metaBlock hashes
func (hv2 *HeaderV2) GetMetaBlockHashes() [][]byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.MetaBlockHashes
}

// GetEpochStartMetaHash returns the epoch start metaBlock hash
func (hv2 *HeaderV2) GetEpochStartMetaHash() []byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	return hv2.Header.EpochStartMetaHash
}

// SetNonce sets header nonce
func (hv2 *HeaderV2) SetNonce(n uint64) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.Nonce = n
	return nil
}

// SetEpoch sets header epoch
func (hv2 *HeaderV2) SetEpoch(e uint32) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.Epoch = e
	return nil
}

// SetRound sets header round
func (hv2 *HeaderV2) SetRound(r uint64) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.Round = r
	return nil
}

// SetRootHash sets root hash
func (hv2 *HeaderV2) SetRootHash(rHash []byte) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.RootHash = rHash
	return nil
}

// SetPrevHash sets prev hash
func (hv2 *HeaderV2) SetPrevHash(pvHash []byte) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.PrevHash = pvHash
	return nil
}

// SetPrevRandSeed sets previous random seed
func (hv2 *HeaderV2) SetPrevRandSeed(pvRandSeed []byte) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.PrevRandSeed = pvRandSeed
	return nil
}

// SetRandSeed sets previous random seed
func (hv2 *HeaderV2) SetRandSeed(randSeed []byte) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.RandSeed = randSeed
	return nil
}

// SetPubKeysBitmap sets public key bitmap
func (hv2 *HeaderV2) SetPubKeysBitmap(pkbm []byte) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.PubKeysBitmap = pkbm
	return nil
}

// SetSignature sets header signature
func (hv2 *HeaderV2) SetSignature(sg []byte) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.Signature = sg
	return nil
}

// SetLeaderSignature will set the leader's signature
func (hv2 *HeaderV2) SetLeaderSignature(sg []byte) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.LeaderSignature = sg
	return nil
}

// SetChainID sets the chain ID on which this block is valid on
func (hv2 *HeaderV2) SetChainID(chainID []byte) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.ChainID = chainID
	return nil
}

// SetSoftwareVersion sets the software version of the header
func (hv2 *HeaderV2) SetSoftwareVersion(version []byte) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.SoftwareVersion = version
	return nil
}

// SetTimeStamp sets header timestamp
func (hv2 *HeaderV2) SetTimeStamp(ts uint64) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.TimeStamp = ts
	return nil
}

// SetAccumulatedFees sets the accumulated fees in the header
func (hv2 *HeaderV2) SetAccumulatedFees(value *big.Int) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}
	if hv2.Header.AccumulatedFees == nil {
		hv2.Header.AccumulatedFees = big.NewInt(0)
	}

	hv2.Header.AccumulatedFees.Set(value)
	return nil
}

// SetDeveloperFees sets the developer fees in the header
func (hv2 *HeaderV2) SetDeveloperFees(value *big.Int) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}
	if hv2.Header.DeveloperFees == nil {
		hv2.Header.DeveloperFees = big.NewInt(0)
	}
	hv2.Header.DeveloperFees.Set(value)
	return nil
}

// SetTxCount sets the transaction count of the block associated with this header
func (hv2 *HeaderV2) SetTxCount(txCount uint32) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.TxCount = txCount
	return nil
}

// SetShardID sets header shard ID
func (hv2 *HeaderV2) SetShardID(shId uint32) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.ShardID = shId
	return nil
}

// GetMiniBlockHeadersWithDst as a map of hashes and sender IDs
func (hv2 *HeaderV2) GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32 {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	hashDst := make(map[string]uint32)
	for _, val := range hv2.Header.MiniBlockHeaders {
		if val.ReceiverShardID == destId && val.SenderShardID != destId {
			hashDst[string(val.Hash)] = val.SenderShardID
		}
	}
	return hashDst
}

// GetOrderedCrossMiniblocksWithDst gets all cross miniblocks with the given destination shard ID, ordered in a
// chronological way, taking into consideration the round in which they were created/executed in the sender shard
func (hv2 *HeaderV2) GetOrderedCrossMiniblocksWithDst(destId uint32) []*data.MiniBlockInfo {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	miniBlocks := make([]*data.MiniBlockInfo, 0)
	for _, mb := range hv2.Header.MiniBlockHeaders {
		if mb.ReceiverShardID == destId && mb.SenderShardID != destId {
			miniBlocks = append(miniBlocks, &data.MiniBlockInfo{
				Hash:          mb.Hash,
				SenderShardID: mb.SenderShardID,
				Round:         hv2.Header.Round,
			})
		}
	}

	return miniBlocks
}

// GetMiniBlockHeadersHashes gets the miniblock hashes
func (hv2 *HeaderV2) GetMiniBlockHeadersHashes() [][]byte {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	result := make([][]byte, 0, len(hv2.Header.MiniBlockHeaders))
	for _, miniblock := range hv2.Header.MiniBlockHeaders {
		result = append(result, miniblock.Hash)
	}
	return result
}

// MapMiniBlockHashesToShards is a map of mini block hashes and sender IDs
func (hv2 *HeaderV2) MapMiniBlockHashesToShards() map[string]uint32 {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	hashDst := make(map[string]uint32)
	for _, val := range hv2.Header.MiniBlockHeaders {
		hashDst[string(val.Hash)] = val.SenderShardID
	}
	return hashDst
}

// ShallowClone returns a clone of the object
func (hv2 *HeaderV2) ShallowClone() data.HeaderHandler {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	internalHeaderCopy := *hv2.Header
	headerCopy := *hv2
	headerCopy.Header = &internalHeaderCopy

	return &headerCopy
}

// IsInterfaceNil returns true if there is no value under the interface
func (hv2 *HeaderV2) IsInterfaceNil() bool {
	return hv2 == nil
}

// IsStartOfEpochBlock verifies if the block is of type start of epoch
func (hv2 *HeaderV2) IsStartOfEpochBlock() bool {
	if hv2 == nil || hv2.Header == nil {
		return false
	}

	return len(hv2.Header.EpochStartMetaHash) > 0
}

// GetBlockBodyTypeInt32 returns the block body type as int32
func (hv2 *HeaderV2) GetBlockBodyTypeInt32() int32 {
	if hv2 == nil || hv2.Header == nil {
		return -1
	}

	return int32(hv2.Header.GetBlockBodyType())
}

// GetMiniBlockHeadersHandlers returns the miniBlock headers as an array of miniBlock header handlers
func (hv2 *HeaderV2) GetMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler {
	if hv2 == nil || hv2.Header == nil {
		return nil
	}

	mbHeaders := hv2.Header.GetMiniBlockHeaders()
	mbHeaderHandlers := make([]data.MiniBlockHeaderHandler, len(mbHeaders))

	for i := range mbHeaders {
		mbHeaderHandlers[i] = &mbHeaders[i]
	}

	return mbHeaderHandlers
}

// SetMiniBlockHeaderHandlers sets the miniBlock headers from the given miniBlock header handlers
func (hv2 *HeaderV2) SetMiniBlockHeaderHandlers(mbHeaderHandlers []data.MiniBlockHeaderHandler) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}
	if mbHeaderHandlers == nil {
		hv2.Header.MiniBlockHeaders = nil
		return nil
	}

	miniBlockHeaders := make([]MiniBlockHeader, len(mbHeaderHandlers))
	hv2.Header.MiniBlockHeaders = make([]MiniBlockHeader, len(mbHeaderHandlers))
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

	hv2.Header.MiniBlockHeaders = miniBlockHeaders

	return nil
}

// SetReceiptsHash sets the receipts hash
func (hv2 *HeaderV2) SetReceiptsHash(hash []byte) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.ReceiptsHash = hash

	return nil
}

// SetMetaBlockHashes sets the metaBlock hashes
func (hv2 *HeaderV2) SetMetaBlockHashes(hashes [][]byte) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}
	if hv2.Header == nil {
		return data.ErrNilPointerDereference
	}

	hv2.Header.MetaBlockHashes = hashes

	return nil
}

// SetScheduledRootHash sets the scheduled root hash
func (hv2 *HeaderV2) SetScheduledRootHash(rootHash []byte) error {
	if hv2 == nil {
		return data.ErrNilPointerReceiver
	}

	hv2.ScheduledRootHash = rootHash

	return nil
}
