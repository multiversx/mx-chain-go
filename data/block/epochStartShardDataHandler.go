package block

import "github.com/ElrondNetwork/elrond-go/data"

// GetPendingMiniBlockHeaderHandlers -
func (m *EpochStartShardData) GetPendingMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler {
	if m == nil {
		return nil
	}

	pendingMbHeaderHandlers := make([]data.MiniBlockHeaderHandler, len(m.PendingMiniBlockHeaders))
	for i := range m.PendingMiniBlockHeaders {
		pendingMbHeaderHandlers[i] = &m.PendingMiniBlockHeaders[i]
	}

	return pendingMbHeaderHandlers
}

// SetShardID -
func (m *EpochStartShardData) SetShardID(shardID uint32) {
	if m == nil {
		return
	}

	m.ShardID = shardID
}

// SetEpoch -
func (m *EpochStartShardData) SetEpoch(epoch uint32) {
	if m == nil {
		return
	}

	m.Epoch = epoch
}

// SetRound -
func (m *EpochStartShardData) SetRound(round uint64) {
	if m == nil {
		return
	}

	m.Round = round
}

// SetNonce -
func (m *EpochStartShardData) SetNonce(nonce uint64) {
	if m == nil {
		return
	}

	m.Nonce = nonce
}

// SetHeaderHash -
func (m *EpochStartShardData) SetHeaderHash(hash []byte) {
	if m == nil {
		return
	}

	m.HeaderHash = hash
}

// SetRootHash -
func (m *EpochStartShardData) SetRootHash(rootHash []byte) {
	if m == nil {
		return
	}
	m.RootHash = rootHash
}

// SetFirstPendingMetaBlock -
func (m *EpochStartShardData) SetFirstPendingMetaBlock(metaBlock []byte) {
	if m == nil {
		return
	}
	m.FirstPendingMetaBlock = metaBlock
}

// SetLastFinishedMetaBlock -
func (m *EpochStartShardData) SetLastFinishedMetaBlock(lastFinishedMetaBlock []byte) {
	if m == nil {
		return
	}

	m.LastFinishedMetaBlock = lastFinishedMetaBlock
}

// SetPendingMiniBlockHeaders -
func (m *EpochStartShardData) SetPendingMiniBlockHeaders(miniBlockHeaderHandlers []data.MiniBlockHeaderHandler) {
	if m == nil {
		return
	}

	pendingMiniBlockHeaders := make([]MiniBlockHeader, len(miniBlockHeaderHandlers))
	for i := range miniBlockHeaderHandlers {
		mbHeader, ok := miniBlockHeaderHandlers[i].(*MiniBlockHeader)
		if !ok {
			m.PendingMiniBlockHeaders = nil
			return
		}
		pendingMiniBlockHeaders[i] = *mbHeader
	}

	m.PendingMiniBlockHeaders = pendingMiniBlockHeaders
}
