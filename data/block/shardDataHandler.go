package block

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// GetShardMiniBlockHeaderHandlers - returns the shard miniBlockHeaders as MiniBlockHeaderHandlers
func (m *ShardData) GetShardMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler {
	if m.ShardMiniBlockHeaders == nil {
		return nil
	}

	miniBlockHeaderHandlers := make([]data.MiniBlockHeaderHandler, len(m.ShardMiniBlockHeaders))
	for i, mbhh := range m.ShardMiniBlockHeaders {
		miniBlockHeaderHandlers[i] = &mbhh
	}
	return miniBlockHeaderHandlers
}

// SetHeaderHash - setter for header hash
func (m *ShardData) SetHeaderHash(hash []byte) {
	m.HeaderHash = hash
}

// SetShardMiniBlockHeaderHandlers - setter for miniBlockHeaders from a list of MiniBlockHeaderHandler
func (m *ShardData) SetShardMiniBlockHeaderHandlers(mbHeaderHandlers []data.MiniBlockHeaderHandler) {
	if mbHeaderHandlers == nil {
		m.ShardMiniBlockHeaders = nil
		return
	}

	miniBlockHeaderHandlers := make([]MiniBlockHeader, len(mbHeaderHandlers))
	for i, mbh := range mbHeaderHandlers {
		mbHeader, ok := mbh.(*MiniBlockHeader)
		if !ok {
			m.ShardMiniBlockHeaders = nil
			return
		}
		miniBlockHeaderHandlers[i] = *mbHeader
	}
}

// SetPrevRandSeed - setter for prevRandSeed
func (m *ShardData) SetPrevRandSeed(prevRandSeed []byte) {
	m.PrevRandSeed = prevRandSeed
}

// SetPubKeysBitmap - setter for pubKeysBitmap
func (m *ShardData) SetPubKeysBitmap(pubKeysBitmap []byte) {
	m.PubKeysBitmap = pubKeysBitmap
}

// SetSignature - setter for signature
func (m *ShardData) SetSignature(signature []byte) {
	m.Signature = signature
}

// SetRound - setter for round
func (m *ShardData) SetRound(round uint64) {
	m.Round = round
}

// SetPrevHash - setter for prevHash
func (m *ShardData) SetPrevHash(prevHash []byte) {
	m.PrevHash = prevHash
}

// SetNonce - setter for nonce
func (m *ShardData) SetNonce(nonce uint64) {
	m.Nonce = nonce
}

// SetAccumulatedFees - setter for accumulatedFees
func (m *ShardData) SetAccumulatedFees(fees *big.Int) {
	m.AccumulatedFees = fees
}

// SetDeveloperFees - setter for developerFees
func (m *ShardData) SetDeveloperFees(fees *big.Int) {
	m.DeveloperFees = fees
}

// SetNumPendingMiniBlocks - setter for number of pending miniBlocks
func (m *ShardData) SetNumPendingMiniBlocks(num uint32) {
	m.NumPendingMiniBlocks = num
}

// SetLastIncludedMetaNonce - setter for the last included metaBlock nonce
func (m *ShardData) SetLastIncludedMetaNonce(nonce uint64) {
	m.LastIncludedMetaNonce = nonce
}

// SetShardID - setter for the shardID
func (m *ShardData) SetShardID(shardID uint32) {
	m.ShardID = shardID
}

// SetTxCount - setter for the transaction count
func (m *ShardData) SetTxCount(txCount uint32) {
	m.TxCount = txCount
}

// Clone - clones the shardData and returns the clone
func (m *ShardData) ShallowClone() data.ShardDataHandler {
	n := *m
	return &n
}
