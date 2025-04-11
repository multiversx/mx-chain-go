package fieldsChecker

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("fieldsChecker")

const (
	// max size for signature in bytes
	sigMaxSize = 100
)

type fieldsSizeChecker struct {
	hasher                 hashing.Hasher
	chainParametersHandler sharding.ChainParametersHandler
}

// NewFieldsSizeChecker will create a new fields size checker component
func NewFieldsSizeChecker(
	chainParametersHandler sharding.ChainParametersHandler,
	hasher hashing.Hasher,
) (*fieldsSizeChecker, error) {
	if check.IfNil(chainParametersHandler) {
		return nil, errors.ErrNilChainParametersHandler
	}
	if check.IfNil(hasher) {
		return nil, core.ErrNilHasher
	}

	return &fieldsSizeChecker{
		chainParametersHandler: chainParametersHandler,
		hasher:                 hasher,
	}, nil
}

// IsProofSizeValid will check proof fields size
func (pc *fieldsSizeChecker) IsProofSizeValid(proof data.HeaderProofHandler) bool {
	epochForConsensus := common.GetEpochForConsensus(proof)

	return pc.isAggregatedSigSizeValid(proof.GetAggregatedSignature()) &&
		pc.isBitmapSizeValid(proof.GetPubKeysBitmap(), epochForConsensus, proof.GetHeaderShardId()) &&
		pc.isHeaderHashSizeValid(proof.GetHeaderHash())
}

func (pc *fieldsSizeChecker) isBitmapSizeValid(
	bitmap []byte,
	epoch uint32,
	shardID uint32,
) bool {
	consensusSize := common.ConsensusGroupSizeForShardAndEpoch(log, pc.chainParametersHandler, shardID, epoch)
	expectedBitmapSize := common.GetBitmapSize(consensusSize)

	return len(bitmap) == expectedBitmapSize
}

func (pc *fieldsSizeChecker) isHeaderHashSizeValid(headerHash []byte) bool {
	return len(headerHash) == pc.hasher.Size()
}

func (pc *fieldsSizeChecker) isAggregatedSigSizeValid(aggSig []byte) bool {
	return len(aggSig) > 0 && len(aggSig) <= sigMaxSize
}

// IsInterfaceNil -
func (pc *fieldsSizeChecker) IsInterfaceNil() bool {
	return pc == nil
}
