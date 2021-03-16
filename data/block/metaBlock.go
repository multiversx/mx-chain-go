//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. metaBlock.proto
package block

import (
	"math"
	"math/big"
	"sort"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// don't break the interface
var _ = data.HeaderHandler(&MetaBlock{})
var _ = data.MetaHeaderHandler(&MetaBlock{})

// GetShardID returns the metachain shard id
func (m *MetaBlock) GetShardID() uint32 {
	return core.MetachainShardId
}

// SetNonce sets header nonce
func (m *MetaBlock) SetNonce(n uint64) {
	m.Nonce = n
}

// SetEpoch sets header epoch
func (m *MetaBlock) SetEpoch(e uint32) {
	m.Epoch = e
}

// SetRound sets header round
func (m *MetaBlock) SetRound(r uint64) {
	m.Round = r
}

// SetRootHash sets root hash
func (m *MetaBlock) SetRootHash(rHash []byte) {
	m.RootHash = rHash
}

// SetValidatorStatsRootHash set's the root hash for the validator statistics trie
func (m *MetaBlock) SetValidatorStatsRootHash(rHash []byte) {
	m.ValidatorStatsRootHash = rHash
}

// SetPrevHash sets prev hash
func (m *MetaBlock) SetPrevHash(pvHash []byte) {
	m.PrevHash = pvHash
}

// SetPrevRandSeed sets the previous randomness seed
func (m *MetaBlock) SetPrevRandSeed(pvRandSeed []byte) {
	m.PrevRandSeed = pvRandSeed
}

// SetRandSeed sets the current random seed
func (m *MetaBlock) SetRandSeed(randSeed []byte) {
	m.RandSeed = randSeed
}

// SetPubKeysBitmap sets publick key bitmap
func (m *MetaBlock) SetPubKeysBitmap(pkbm []byte) {
	m.PubKeysBitmap = pkbm
}

// SetSignature set header signature
func (m *MetaBlock) SetSignature(sg []byte) {
	m.Signature = sg
}

// SetLeaderSignature will set the leader's signature
func (m *MetaBlock) SetLeaderSignature(sg []byte) {
	m.LeaderSignature = sg
}

// SetChainID sets the chain ID on which this block is valid on
func (m *MetaBlock) SetChainID(chainID []byte) {
	m.ChainID = chainID
}

// SetSoftwareVersion sets the software version of the block
func (m *MetaBlock) SetSoftwareVersion(version []byte) {
	m.SoftwareVersion = version
}

// SetAccumulatedFees sets the accumulated fees in the header
func (m *MetaBlock) SetAccumulatedFees(value *big.Int) {
	if m.AccumulatedFees  == nil{
		m.AccumulatedFees = big.NewInt(0)
	}
	m.AccumulatedFees.Set(value)
}

// SetAccumulatedFeesInEpoch sets the epoch accumulated fees in the header
func (m *MetaBlock) SetAccumulatedFeesInEpoch(value *big.Int) {
	if m.AccumulatedFeesInEpoch  == nil{
		m.AccumulatedFeesInEpoch = big.NewInt(0)
	}
	m.AccumulatedFeesInEpoch.Set(value)
}

// SetDeveloperFees sets the developer fees in the header
func (m *MetaBlock) SetDeveloperFees(value *big.Int) {
	if m.DeveloperFees == nil {
		m.DeveloperFees = big.NewInt(0)
	}
	m.DeveloperFees.Set(value)
}

// SetDeveloperFees sets the developer fees in the header
func (m *MetaBlock) SetDevFeesInEpoch(value *big.Int) {
	if m.DevFeesInEpoch == nil{
		m.DevFeesInEpoch = big.NewInt(0)
	}
	m.DevFeesInEpoch.Set(value)
}

// SetTimeStamp sets header timestamp
func (m *MetaBlock) SetTimeStamp(ts uint64) {
	m.TimeStamp = ts
}

// SetTxCount sets the transaction count of the current meta block
func (m *MetaBlock) SetTxCount(txCount uint32) {
	m.TxCount = txCount
}

// SetShardID sets header shard ID
func (m *MetaBlock) SetShardID(_ uint32) {
}

// GetMiniBlockHeadersWithDst as a map of hashes and sender IDs
func (m *MetaBlock) GetMiniBlockHeadersWithDst(destId uint32) map[string]uint32 {
	hashDst := make(map[string]uint32)
	for i := 0; i < len(m.ShardInfo); i++ {
		if m.ShardInfo[i].ShardID == destId {
			continue
		}

		for _, val := range m.ShardInfo[i].ShardMiniBlockHeaders {
			if val.ReceiverShardID == destId && val.SenderShardID != destId {
				hashDst[string(val.Hash)] = val.SenderShardID
			}
		}
	}

	for _, val := range m.MiniBlockHeaders {
		isDestinationShard := (val.ReceiverShardID == destId ||
			val.ReceiverShardID == core.AllShardId) &&
			val.SenderShardID != destId
		if isDestinationShard {
			hashDst[string(val.Hash)] = val.SenderShardID
		}
	}

	return hashDst
}

// GetOrderedCrossMiniblocksWithDst gets all cross miniblocks with the given destination shard ID, ordered in a
// chronological way, taking into consideration the round in which they were created/executed in the sender shard
func (m *MetaBlock) GetOrderedCrossMiniblocksWithDst(destId uint32) []*data.MiniBlockInfo {
	miniBlocks := make([]*data.MiniBlockInfo, 0)

	for i := 0; i < len(m.ShardInfo); i++ {
		if m.ShardInfo[i].ShardID == destId {
			continue
		}

		for _, mb := range m.ShardInfo[i].ShardMiniBlockHeaders {
			if mb.ReceiverShardID == destId && mb.SenderShardID != destId {
				miniBlocks = append(miniBlocks, &data.MiniBlockInfo{
					Hash:          mb.Hash,
					SenderShardID: mb.SenderShardID,
					Round:         m.ShardInfo[i].Round,
				})
			}
		}
	}

	for _, mb := range m.MiniBlockHeaders {
		isDestinationShard := (mb.ReceiverShardID == destId ||
			mb.ReceiverShardID == core.AllShardId) &&
			mb.SenderShardID != destId
		if isDestinationShard {
			miniBlocks = append(miniBlocks, &data.MiniBlockInfo{
				Hash:          mb.Hash,
				SenderShardID: mb.SenderShardID,
				Round:         m.Round,
			})
		}
	}

	sort.Slice(miniBlocks, func(i, j int) bool {
		return miniBlocks[i].Round < miniBlocks[j].Round
	})

	return miniBlocks
}

// GetMiniBlockHeadersHashes gets the miniblock hashes
func (m *MetaBlock) GetMiniBlockHeadersHashes() [][]byte {
	result := make([][]byte, 0, len(m.MiniBlockHeaders))
	for _, miniblock := range m.MiniBlockHeaders {
		result = append(result, miniblock.Hash)
	}
	return result
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *MetaBlock) IsInterfaceNil() bool {
	return m == nil
}

// IsStartOfEpochBlock verifies if the block is of type start of epoch
func (m *MetaBlock) IsStartOfEpochBlock() bool {
	return len(m.EpochStart.LastFinalizedHeaders) > 0
}

// Clone will return a clone of the object
func (m *MetaBlock) ShallowClone() data.HeaderHandler {
	metaBlockCopy := *m
	return &metaBlockCopy
}

// GetEpochStartMetaHash returns the hash of the epoch start meta block
func (m *MetaBlock) GetEpochStartMetaHash() []byte {
	return nil
}

// GetMetaBlockHashes() - the metaBlock does not hold any metaBlock hashes
func (m *MetaBlock) GetMetaBlockHashes() [][]byte {
	return nil
}

// GetBlockBodyType - currently there is no block body type for metaBlock so return MaxInt32
func (m *MetaBlock) GetBlockBodyTypeInt32() int32 {
	return math.MaxInt32
}

// GetMiniBlockHeadersHandlers returns the miniBlock headers as an array of miniBlock header handlers
func (m *MetaBlock) GetMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler {
	mbHeaders := m.GetMiniBlockHeaders()
	mbHeaderHandlers := make([]data.MiniBlockHeaderHandler, len(mbHeaders))

	for i := range mbHeaders {
		mbHeaderHandlers[i] = &mbHeaders[i]
	}

	return mbHeaderHandlers
}

// SetMiniBlockHeaderHandlers sets the miniBlock headers from the given miniBlock header handlers
func (m *MetaBlock) SetMiniBlockHeaderHandlers(mbHeaderHandlers []data.MiniBlockHeaderHandler) {
	if mbHeaderHandlers == nil {
		m.MiniBlockHeaders = nil
		return
	}

	m.MiniBlockHeaders = make([]MiniBlockHeader, len(mbHeaderHandlers))
	for i := range mbHeaderHandlers {
		mbHeader, ok := mbHeaderHandlers[i].(*MiniBlockHeader)
		if !ok {
			m.MiniBlockHeaders = nil
			return
		}
		m.MiniBlockHeaders[i] = *mbHeader
	}
}

// SetReceiptsHash sets the receipts hash
func (m *MetaBlock) SetReceiptsHash(hash []byte) {
	m.ReceiptsHash = hash
}

// SetMetaBlockHashes - for metaBlock does nothing
func (m *MetaBlock) SetMetaBlockHashes(_ [][]byte) {
}

// GetShardInfoHandlers - gets the shardInfo as an array of ShardDataHandler
func (m *MetaBlock) GetShardInfoHandlers() []data.ShardDataHandler {
	if m.ShardInfo == nil {
		return nil
	}

	shardInfoHandlers := make([]data.ShardDataHandler, len(m.ShardInfo))
	for i := range m.ShardInfo {
		shardInfoHandlers[i] = &m.ShardInfo[i]
	}

	return shardInfoHandlers
}

// GetEpochStartHandler -
func (m *MetaBlock) GetEpochStartHandler() data.EpochStartHandler {
	if m == nil {
		return nil
	}
	return &m.EpochStart
}

// SetShardInfoHandlers -
func (m *MetaBlock) SetShardInfoHandlers(shardInfo []data.ShardDataHandler){
	if shardInfo == nil{
		m.ShardInfo = nil
		return
	}

	m.ShardInfo = make([]ShardData, len(shardInfo))
	for i := range shardInfo {
		shData, ok := shardInfo[i].(*ShardData)
		if !ok {
			m.ShardInfo = nil
			return
		}
		m.ShardInfo[i] = *shData
	}
}
