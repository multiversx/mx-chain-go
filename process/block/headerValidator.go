package block

import (
	"bytes"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ process.HeaderConstructionValidator = (*headerValidator)(nil)

// ArgsHeaderValidator are the arguments needed to create a new header validator
type ArgsHeaderValidator struct {
	Logger      logger.Logger
	Hasher      hashing.Hasher
	Marshalizer marshal.Marshalizer
}

type headerValidator struct {
	log         logger.Logger
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

// NewHeaderValidator returns a new header validator
func NewHeaderValidator(args ArgsHeaderValidator) (*headerValidator, error) {
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}

	var log logger.Logger
	log = logger.GetOrCreate("process/block")
	if args.Logger != nil {
		log = args.Logger
	}

	return &headerValidator{
		log:         log,
		hasher:      args.Hasher,
		marshalizer: args.Marshalizer,
	}, nil
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
		h.log.Trace("round does not match",
			"shard", currHeader.GetShardID(),
			"local header round", prevHeader.GetRound(),
			"received round", currHeader.GetRound())
		return process.ErrLowerRoundInBlock
	}

	if currHeader.GetNonce() != prevHeader.GetNonce()+1 {
		h.log.Trace("nonce does not match",
			"shard", currHeader.GetShardID(),
			"local header nonce", prevHeader.GetNonce(),
			"received nonce", currHeader.GetNonce())
		return process.ErrWrongNonceInBlock
	}

	prevHeaderHash, err := core.CalculateHash(h.marshalizer, h.hasher, prevHeader)
	if err != nil {
		return err
	}

	if !bytes.Equal(currHeader.GetPrevHash(), prevHeaderHash) {
		h.log.Trace("header hash does not match",
			"shard", currHeader.GetShardID(),
			"local header hash", prevHeaderHash,
			"received header with prev hash", currHeader.GetPrevHash(),
		)
		return process.ErrBlockHashDoesNotMatch
	}

	if !bytes.Equal(currHeader.GetPrevRandSeed(), prevHeader.GetRandSeed()) {
		h.log.Trace("header random seed does not match",
			"shard", currHeader.GetShardID(),
			"local header random seed", prevHeader.GetRandSeed(),
			"received header with prev random seed", currHeader.GetPrevRandSeed(),
		)
		return process.ErrRandSeedDoesNotMatch
	}

	// TODO: check here if proof from currHeader is valid for prevHeader

	return nil
}

// IsInterfaceNil returns if underlying object is true
func (h *headerValidator) IsInterfaceNil() bool {
	return h == nil
}
