package block

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// GetShardMiniBlockHeaderHandlers returns the shard miniBlockHeaders as MiniBlockHeaderHandlers
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

// SetHeaderHash sets the header hash
func (sd *ShardData) SetHeaderHash(hash []byte) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	sd.HeaderHash = hash
	return nil
}

// SetShardMiniBlockHeaderHandlers sets the miniBlockHeaders from a list of MiniBlockHeaderHandler
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

// SetPrevRandSeed sets the prevRandSeed
func (sd *ShardData) SetPrevRandSeed(prevRandSeed []byte) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	sd.PrevRandSeed = prevRandSeed

	return nil
}

// SetPubKeysBitmap sets the pubKeysBitmap
func (sd *ShardData) SetPubKeysBitmap(pubKeysBitmap []byte) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	sd.PubKeysBitmap = pubKeysBitmap

	return nil
}

// SetSignature sets the signature
func (sd *ShardData) SetSignature(signature []byte) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	sd.Signature = signature

	return nil
}

// SetRound sets the round
func (sd *ShardData) SetRound(round uint64) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	sd.Round = round

	return nil
}

// SetPrevHash sets the prevHash
func (sd *ShardData) SetPrevHash(prevHash []byte) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	sd.PrevHash = prevHash

	return nil
}

// SetNonce sets the nonce
func (sd *ShardData) SetNonce(nonce uint64) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	sd.Nonce = nonce

	return nil
}

// SetAccumulatedFees sets the accumulatedFees
func (sd *ShardData) SetAccumulatedFees(fees *big.Int) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	if fees == nil {
		return data.ErrInvalidValue
	}
	if sd.AccumulatedFees == nil {
		sd.AccumulatedFees = big.NewInt(0)
	}

	sd.AccumulatedFees.Set(fees)

	return nil
}

// SetDeveloperFees sets the developerFees
func (sd *ShardData) SetDeveloperFees(fees *big.Int) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}
	if fees == nil {
		return data.ErrInvalidValue
	}
	if sd.DeveloperFees == nil {
		sd.DeveloperFees = big.NewInt(0)
	}

	sd.DeveloperFees.Set(fees)

	return nil
}

// SetNumPendingMiniBlocks sets the number of pending miniBlocks
func (sd *ShardData) SetNumPendingMiniBlocks(num uint32) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	sd.NumPendingMiniBlocks = num

	return nil
}

// SetLastIncludedMetaNonce sets the last included metaBlock nonce
func (sd *ShardData) SetLastIncludedMetaNonce(nonce uint64) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	sd.LastIncludedMetaNonce = nonce

	return nil
}

// SetShardID sets the shardID
func (sd *ShardData) SetShardID(shardID uint32) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	sd.ShardID = shardID

	return nil
}

// SetTxCount sets the transaction count
func (sd *ShardData) SetTxCount(txCount uint32) error {
	if sd == nil {
		return data.ErrNilPointerReceiver
	}

	sd.TxCount = txCount

	return nil
}

// ShallowClone creates and returns a shallow clone of shardData
func (sd *ShardData) ShallowClone() data.ShardDataHandler {
	if sd == nil {
		return nil
	}

	n := *sd

	return &n
}
