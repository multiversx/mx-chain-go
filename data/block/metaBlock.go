package block

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// PeerAction type represents the possible events that a node can trigger for the metachain to notarize
type PeerAction uint8

// Constants mapping the actions that a node can take
const (
	PeerRegistration PeerAction = iota + 1
	PeerUnstaking
	PeerDeregistration
	PeerJailed
	PeerUnJailed
	PeerSlashed
	PeerReStake
)

func (pa PeerAction) String() string {
	switch pa {
	case PeerRegistration:
		return "PeerRegistration"
	case PeerUnstaking:
		return "PeerUnstaking"
	case PeerDeregistration:
		return "PeerDeregistration"
	case PeerJailed:
		return "PeerJailed"
	case PeerUnJailed:
		return "PeerUnjailed"
	case PeerSlashed:
		return "PeerSlashed"
	case PeerReStake:
		return "PeerReStake"
	default:
		return fmt.Sprintf("Unknown type (%d)", pa)
	}
}

// PeerData holds information about actions taken by a peer:
//  - a peer can register with an amount to become a validator
//  - a peer can choose to deregister and get back the deposited value
type PeerData struct {
	Address     []byte
	PublicKey   []byte
	Action      PeerAction
	TimeStamp   uint64
	ValueChange *big.Int
}

// ShardMiniBlockHeader holds data for one shard miniblock header
type ShardMiniBlockHeader struct {
	Hash            []byte
	ReceiverShardID uint32
	SenderShardID   uint32
	TxCount         uint32
}

// ShardData holds the block information sent by the shards to the metachain
type ShardData struct {
	HeaderHash            []byte
	ShardMiniBlockHeaders []ShardMiniBlockHeader
	PrevRandSeed          []byte
	PubKeysBitmap         []byte
	Signature             []byte
	Round                 uint64
	PrevHash              []byte
	Nonce                 uint64
	NumPendingMiniBlocks  uint32
	ShardID               uint32
	TxCount               uint32
	AccumulatedFees       *big.Int
}

// EpochStartShardData hold the last finalized headers hash and state root hash
type EpochStartShardData struct {
	ShardId                 uint32
	HeaderHash              []byte
	RootHash                []byte
	FirstPendingMetaBlock   []byte
	LastFinishedMetaBlock   []byte
	PendingMiniBlockHeaders []ShardMiniBlockHeader
}

// EpochStart holds the block information for end-of-epoch
type EpochStart struct {
	LastFinalizedHeaders []EpochStartShardData
}

// MetaBlock holds the data that will be saved to the metachain each round
type MetaBlock struct {
	Nonce                  uint64
	Round                  uint64
	TimeStamp              uint64
	ShardInfo              []ShardData
	PeerInfo               []PeerData
	Signature              []byte
	LeaderSignature        []byte
	PubKeysBitmap          []byte
	PrevHash               []byte
	PrevRandSeed           []byte
	RandSeed               []byte
	RootHash               []byte
	ValidatorStatsRootHash []byte
	MiniBlockHeaders       []MiniBlockHeader
	ReceiptsHash           []byte
	EpochStart             EpochStart
	ChainID                []byte
	Epoch                  uint32
	TxCount                uint32
	AccumulatedFees        *big.Int
	AccumulatedFeesInEpoch *big.Int
}

// GetShardID returns the metachain shard id
func (m *MetaBlock) GetShardID() uint32 {
	return sharding.MetachainShardId
}

// GetNonce return header nonce
func (m *MetaBlock) GetNonce() uint64 {
	return m.Nonce
}

// GetEpoch return header epoch
func (m *MetaBlock) GetEpoch() uint32 {
	return m.Epoch
}

// GetRound return round from header
func (m *MetaBlock) GetRound() uint64 {
	return m.Round
}

// GetTimeStamp returns the time stamp
func (m *MetaBlock) GetTimeStamp() uint64 {
	return m.TimeStamp
}

// GetRootHash returns the roothash from header
func (m *MetaBlock) GetRootHash() []byte {
	return m.RootHash
}

// GetValidatorStatsRootHash returns the root hash for the validator statistics trie at this current block
func (m *MetaBlock) GetValidatorStatsRootHash() []byte {
	return m.ValidatorStatsRootHash
}

// GetPrevHash returns previous block header hash
func (m *MetaBlock) GetPrevHash() []byte {
	return m.PrevHash
}

// GetPrevRandSeed gets the previous random seed
func (m *MetaBlock) GetPrevRandSeed() []byte {
	return m.PrevRandSeed
}

// GetRandSeed gets the current random seed
func (m *MetaBlock) GetRandSeed() []byte {
	return m.RandSeed
}

// GetPubKeysBitmap return signers bitmap
func (m *MetaBlock) GetPubKeysBitmap() []byte {
	return m.PubKeysBitmap
}

// GetSignature return signed data
func (m *MetaBlock) GetSignature() []byte {
	return m.Signature
}

// GetLeaderSignature returns the signature of the leader
func (m *MetaBlock) GetLeaderSignature() []byte {
	return m.LeaderSignature
}

// GetChainID gets the chain ID on which this block is valid on
func (m *MetaBlock) GetChainID() []byte {
	return m.ChainID
}

// GetTxCount returns transaction count in the current meta block
func (m *MetaBlock) GetTxCount() uint32 {
	return m.TxCount
}

// GetReceiptsHash returns the hash of the receipts and intra-shard smart contract results
func (m *MetaBlock) GetReceiptsHash() []byte {
	return m.ReceiptsHash
}

// GetAccumulatedFees returns the accumulated fees in the header
func (m *MetaBlock) GetAccumulatedFees() *big.Int {
	return big.NewInt(0).Set(m.AccumulatedFees)
}

// SetAccumulatedFees sets the accumulated fees in the header
func (m *MetaBlock) SetAccumulatedFees(value *big.Int) {
	m.AccumulatedFees.Set(value)
}

// SetShardID sets header shard ID
func (m *MetaBlock) SetShardID(_ uint32) {
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

// SetTimeStamp sets header timestamp
func (m *MetaBlock) SetTimeStamp(ts uint64) {
	m.TimeStamp = ts
}

// SetTxCount sets the transaction count of the current meta block
func (m *MetaBlock) SetTxCount(txCount uint32) {
	m.TxCount = txCount
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
		if val.ReceiverShardID == destId && val.SenderShardID != destId {
			hashDst[string(val.Hash)] = val.SenderShardID
		}
	}

	return hashDst
}

// IsInterfaceNil returns true if there is no value under the interface
func (m *MetaBlock) IsInterfaceNil() bool {
	if m == nil {
		return true
	}
	return false
}

// ItemsInHeader gets the number of items(hashes) added in block header
func (m *MetaBlock) ItemsInHeader() uint32 {
	itemsInHeader := len(m.ShardInfo)
	for i := 0; i < len(m.ShardInfo); i++ {
		itemsInHeader += len(m.ShardInfo[i].ShardMiniBlockHeaders)
	}

	itemsInHeader += len(m.PeerInfo)
	itemsInHeader += len(m.MiniBlockHeaders)

	return uint32(itemsInHeader)
}

// IsStartOfEpochBlock verifies if the block is of type start of epoch
func (m *MetaBlock) IsStartOfEpochBlock() bool {
	return len(m.EpochStart.LastFinalizedHeaders) > 0
}

// ItemsInBody gets the number of items(hashes) added in block body
func (m *MetaBlock) ItemsInBody() uint32 {
	return m.TxCount
}

// Clone will return a clone of the object
func (m *MetaBlock) Clone() data.HeaderHandler {
	metaBlockCopy := *m
	return &metaBlockCopy
}

// CheckChainID returns nil if the header's chain ID matches the one provided
// otherwise, it will error
func (m *MetaBlock) CheckChainID(reference []byte) error {
	if !bytes.Equal(m.ChainID, reference) {
		return fmt.Errorf(
			"%w, expected: %s, got %s",
			data.ErrInvalidChainID,
			hex.EncodeToString(reference),
			hex.EncodeToString(m.ChainID),
		)
	}

	return nil
}
