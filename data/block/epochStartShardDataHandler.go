package block

import "github.com/ElrondNetwork/elrond-go/data"

// GetPendingMiniBlockHeaderHandlers returns the pending miniBlock header handlers
func (essd *EpochStartShardData) GetPendingMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler {
	if essd == nil {
		return nil
	}

	pendingMbHeaderHandlers := make([]data.MiniBlockHeaderHandler, len(essd.PendingMiniBlockHeaders))
	for i := range essd.PendingMiniBlockHeaders {
		pendingMbHeaderHandlers[i] = &essd.PendingMiniBlockHeaders[i]
	}

	return pendingMbHeaderHandlers
}

// SetShardID sets the epoch start shardData shardID
func (essd *EpochStartShardData) SetShardID(shardID uint32) error {
	if essd == nil {
		return data.ErrNilPointerReceiver
	}

	essd.ShardID = shardID
	return nil
}

// SetEpoch sets the epoch start shardData epoch
func (essd *EpochStartShardData) SetEpoch(epoch uint32) error {
	if essd == nil {
		return data.ErrNilPointerReceiver
	}

	essd.Epoch = epoch
	return nil
}

// SetRound sets the epoch start shardData round
func (essd *EpochStartShardData) SetRound(round uint64) error {
	if essd == nil {
		return data.ErrNilPointerReceiver
	}

	essd.Round = round
	return nil
}

// SetNonce sets the epoch start shardData nonce
func (essd *EpochStartShardData) SetNonce(nonce uint64) error {
	if essd == nil {
		return data.ErrNilPointerReceiver
	}

	essd.Nonce = nonce
	return nil
}

// SetHeaderHash sets the epoch start shardData header hash
func (essd *EpochStartShardData) SetHeaderHash(hash []byte) error {
	if essd == nil {
		return data.ErrNilPointerReceiver
	}

	essd.HeaderHash = hash
	return nil
}

// SetRootHash sets the epoch start shardData root hash
func (essd *EpochStartShardData) SetRootHash(rootHash []byte) error {
	if essd == nil {
		return data.ErrNilPointerReceiver
	}
	essd.RootHash = rootHash
	return nil
}

// SetFirstPendingMetaBlock sets the epoch start shardData first pending metaBlock
func (essd *EpochStartShardData) SetFirstPendingMetaBlock(metaBlock []byte) error {
	if essd == nil {
		return data.ErrNilPointerReceiver
	}
	essd.FirstPendingMetaBlock = metaBlock
	return nil
}

// SetLastFinishedMetaBlock sets the epoch start shardData last finished metaBlock
func (essd *EpochStartShardData) SetLastFinishedMetaBlock(lastFinishedMetaBlock []byte) error {
	if essd == nil {
		return data.ErrNilPointerReceiver
	}

	essd.LastFinishedMetaBlock = lastFinishedMetaBlock
	return nil
}

// SetPendingMiniBlockHeaders sets the epoch start shardData pending miniBlock headers
func (essd *EpochStartShardData) SetPendingMiniBlockHeaders(miniBlockHeaderHandlers []data.MiniBlockHeaderHandler) error {
	if essd == nil {
		return data.ErrNilPointerReceiver
	}

	pendingMiniBlockHeaders := make([]MiniBlockHeader, len(miniBlockHeaderHandlers))
	for i := range miniBlockHeaderHandlers {
		mbHeader, ok := miniBlockHeaderHandlers[i].(*MiniBlockHeader)
		if !ok {
			return data.ErrInvalidTypeAssertion
		}
		if mbHeader == nil {
			return data.ErrNilPointerDereference
		}
		pendingMiniBlockHeaders[i] = *mbHeader
	}

	essd.PendingMiniBlockHeaders = pendingMiniBlockHeaders
	return nil
}
