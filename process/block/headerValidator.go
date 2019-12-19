package block

import (
	"bytes"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ArgsHeaderValidator are the arguments needed to create a new header validator
type ArgsHeaderValidator struct {
	Hasher      hashing.Hasher
	Marshalizer marshal.Marshalizer
}

type headerValidator struct {
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

	return &headerValidator{
		hasher:      args.Hasher,
		marshalizer: args.Marshalizer,
	}, nil
}

// IsHeaderConstructionValid verified if header is constructed correctly on top of other
func (h *headerValidator) IsHeaderConstructionValid(currHdr, prevHdr data.HeaderHandler) error {
	if prevHdr == nil || prevHdr.IsInterfaceNil() {
		return process.ErrNilBlockHeader
	}
	if currHdr == nil || currHdr.IsInterfaceNil() {
		return process.ErrNilBlockHeader
	}

	// special case with genesis nonce - 0
	if currHdr.GetNonce() == 0 {
		if prevHdr.GetNonce() != 0 {
			return process.ErrWrongNonceInBlock
		}
		// block with nonce 0 was already saved
		if prevHdr.GetRootHash() != nil {
			return process.ErrRootStateDoesNotMatch
		}
		return nil
	}

	//TODO: add verification if rand seed was correctly computed add other verification
	//TODO: check here if the 2 header blocks were correctly signed and the consensus group was correctly elected
	if prevHdr.GetRound() >= currHdr.GetRound() {
		log.Debug(fmt.Sprintf("round does not match in shard %d: local block round is %d and node received block with round %d\n",
			currHdr.GetShardID(), prevHdr.GetRound(), currHdr.GetRound()))
		return process.ErrLowerRoundInBlock
	}

	if currHdr.GetNonce() != prevHdr.GetNonce()+1 {
		log.Debug(fmt.Sprintf("nonce does not match in shard %d: local block nonce is %d and node received block with nonce %d\n",
			currHdr.GetShardID(), prevHdr.GetNonce(), currHdr.GetNonce()))
		return process.ErrWrongNonceInBlock
	}

	prevHeaderHash, err := core.CalculateHash(h.marshalizer, h.hasher, prevHdr)
	if err != nil {
		return err
	}

	if !bytes.Equal(currHdr.GetPrevHash(), prevHeaderHash) {
		log.Debug(fmt.Sprintf("block hash does not match in shard %d: local block hash is %s and node received block with previous hash %s\n",
			currHdr.GetShardID(), core.ToB64(prevHeaderHash), core.ToB64(currHdr.GetPrevHash())))
		return process.ErrBlockHashDoesNotMatch
	}

	if !bytes.Equal(currHdr.GetPrevRandSeed(), prevHdr.GetRandSeed()) {
		log.Debug(fmt.Sprintf("random seed does not match in shard %d: local block random seed is %s and node received block with previous random seed %s\n",
			currHdr.GetShardID(), core.ToB64(prevHdr.GetRandSeed()), core.ToB64(currHdr.GetPrevRandSeed())))
		return process.ErrRandSeedDoesNotMatch
	}

	return nil
}

// IsInterfaceNil returns if underlying object is true
func (h *headerValidator) IsInterfaceNil() bool {
	return h == nil
}
