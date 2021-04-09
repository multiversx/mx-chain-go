package block

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// GetShardMiniBlockHeaderHandlers - returns the shard miniBlockHeaders as MiniBlockHeaderHandlers
func (sd *ShardData) GetShardMiniBlockHeaderHandlers() []data.MiniBlockHeaderHandler {
	if sd == nil || sd.ShardMiniBlockHeaders == nil {
		return nil
	}

	miniBlockHeaderHandlers := make([]data.MiniBlockHeaderHandler, len(sd.ShardMiniBlockHeaders))
	for i := range sd.ShardMiniBlockHeaders {
		miniBlockHeaderHandlers[i] = &sd.ShardMiniBlockHeaders[i]
	}
	return miniBlockHeaderHandlers
}

// SetHeaderHash - setter for header hash
func (sd *ShardData) SetHeaderHash(hash []byte) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	sd.HeaderHash = hash
	return nil
}

// SetShardMiniBlockHeaderHandlers - setter for miniBlockHeaders from a list of MiniBlockHeaderHandler
func (sd *ShardData) SetShardMiniBlockHeaderHandlers(mbHeaderHandlers []data.MiniBlockHeaderHandler) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	if mbHeaderHandlers == nil {
		sd.ShardMiniBlockHeaders = nil
		return nil
	}

	miniBlockHeaders := make([]MiniBlockHeader, len(mbHeaderHandlers))
	for i, mbh := range mbHeaderHandlers {
		mbHeader, ok := mbh.(*MiniBlockHeader)
		if !ok {
			return data.ErrInvalidTypeAssertion
		}
		if mbHeader == nil {
			return data.ErrNilPointerDereference
		}

		miniBlockHeaders[i] = *mbHeader
	}

	sd.ShardMiniBlockHeaders = miniBlockHeaders
	return nil
}

// SetPrevRandSeed - setter for prevRandSeed
func (sd *ShardData) SetPrevRandSeed(prevRandSeed []byte) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	sd.PrevRandSeed = prevRandSeed
	return nil
}

// SetPubKeysBitmap - setter for pubKeysBitmap
func (sd *ShardData) SetPubKeysBitmap(pubKeysBitmap []byte) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	sd.PubKeysBitmap = pubKeysBitmap
	return nil
}

// SetSignature - setter for signature
func (sd *ShardData) SetSignature(signature []byte) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	sd.Signature = signature
	return nil
}

// SetRound - setter for round
func (sd *ShardData) SetRound(round uint64) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	sd.Round = round
	return nil
}

// SetPrevHash - setter for prevHash
func (sd *ShardData) SetPrevHash(prevHash []byte) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	sd.PrevHash = prevHash
	return nil
}

// SetNonce - setter for nonce
func (sd *ShardData) SetNonce(nonce uint64) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	sd.Nonce = nonce
	return nil
}

// SetAccumulatedFees - setter for accumulatedFees
func (sd *ShardData) SetAccumulatedFees(fees *big.Int) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	if sd.AccumulatedFees == nil {
		sd.AccumulatedFees = big.NewInt(0)
	}
	sd.AccumulatedFees.Set(fees)
	return nil
}

// SetDeveloperFees - setter for developerFees
func (sd *ShardData) SetDeveloperFees(fees *big.Int) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	if sd.DeveloperFees == nil {
		sd.DeveloperFees = big.NewInt(0)
	}
	sd.DeveloperFees.Set(fees)
	return nil
}

// SetNumPendingMiniBlocks - setter for number of pending miniBlocks
func (sd *ShardData) SetNumPendingMiniBlocks(num uint32) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	sd.NumPendingMiniBlocks = num
	return nil
}

// SetLastIncludedMetaNonce - setter for the last included metaBlock nonce
func (sd *ShardData) SetLastIncludedMetaNonce(nonce uint64) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	sd.LastIncludedMetaNonce = nonce
	return nil
}

// SetShardID - setter for the shardID
func (sd *ShardData) SetShardID(shardID uint32) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	sd.ShardID = shardID
	return nil
}

// SetTxCount - setter for the transaction count
func (sd *ShardData) SetTxCount(txCount uint32) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	sd.TxCount = txCount
	return nil
}

// ShallowClone - clones the shardData and returns the clone
func (sd *ShardData) ShallowClone() data.ShardDataHandler {
	if sd == nil {
		return nil
	}
	n := *sd
	return &n
}
