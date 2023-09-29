package block

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
)

var _ process.HeaderConstructionValidator = (*headerValidator)(nil)

// ArgsHeaderValidator are the arguments needed to create a new header validator
type ArgsHeaderValidator struct {
	Hasher      hashing.Hasher
	Marshalizer marshal.Marshalizer
}

type headerValidator struct {
	hasher                  hashing.Hasher
	marshalizer             marshal.Marshalizer
	calculateHeaderHashFunc func(headerHandler data.HeaderHandler) ([]byte, error)
}

// NewHeaderValidator returns a new header validator
func NewHeaderValidator(args ArgsHeaderValidator) (*headerValidator, error) {
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	hv := &headerValidator{
		hasher:      args.Hasher,
		marshalizer: args.Marshalizer,
	}

	hv.calculateHeaderHashFunc = hv.calculateHeaderHash
	return hv, nil
}

// IsHeaderConstructionValid verified if header is constructed correctly on top of other
func (h *headerValidator) IsHeaderConstructionValid(currHeader, prevHeader data.HeaderHandler) error {
	if check.IfNil(prevHeader) {
		return process.ErrNilBlockHeader
	}
	if check.IfNil(currHeader) {
		return process.ErrNilBlockHeader
	}

	if prevHeader.GetRound() >= currHeader.GetRound() {
		log.Trace("round does not match",
			"shard", currHeader.GetShardID(),
			"local header round", prevHeader.GetRound(),
			"received round", currHeader.GetRound())
		return process.ErrLowerRoundInBlock
	}

	if currHeader.GetNonce() != prevHeader.GetNonce()+1 {
		log.Trace("nonce does not match",
			"shard", currHeader.GetShardID(),
			"local header nonce", prevHeader.GetNonce(),
			"received nonce", currHeader.GetNonce())
		return process.ErrWrongNonceInBlock
	}

	prevHeaderHash, err := h.calculateHeaderHashFunc(prevHeader)
	if err != nil {
		return err
	}

	if !bytes.Equal(currHeader.GetPrevHash(), prevHeaderHash) {
		log.Trace("header hash does not match",
			"shard", currHeader.GetShardID(),
			"local header hash", prevHeaderHash,
			"received header with prev hash", currHeader.GetPrevHash(),
		)
		return process.ErrBlockHashDoesNotMatch
	}

	if !bytes.Equal(currHeader.GetPrevRandSeed(), prevHeader.GetRandSeed()) {
		log.Trace("header random seed does not match",
			"shard", currHeader.GetShardID(),
			"local header random seed", prevHeader.GetRandSeed(),
			"received header with prev random seed", currHeader.GetPrevRandSeed(),
		)
		return process.ErrRandSeedDoesNotMatch
	}

	return nil
}

func (h *headerValidator) calculateHeaderHash(headerHandler data.HeaderHandler) ([]byte, error) {
	return core.CalculateHash(h.marshalizer, h.hasher, headerHandler)
}

// IsInterfaceNil returns if underlying object is true
func (h *headerValidator) IsInterfaceNil() bool {
	return h == nil
}
