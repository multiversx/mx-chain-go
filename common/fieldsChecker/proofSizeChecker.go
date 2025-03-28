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

var log = logger.GetOrCreate("fieldschecker")

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
	return pc.isAggregatedSigSizeValid(proof.GetAggregatedSignature()) &&
		pc.isBitmapSizeValid(proof.GetPubKeysBitmap(), proof.GetHeaderEpoch(), proof.GetHeaderShardId()) &&
		pc.isHeaderHashSizeValid(proof.GetHeaderHash())
}

func (pc *fieldsSizeChecker) isBitmapSizeValid(
	bitmap []byte,
	epoch uint32,
	shardID uint32,
) bool {
	consensusSize := common.ConsensusGroupSizeForShardAndEpoch(log, pc.chainParametersHandler, shardID, epoch)
	expectedBitmapSize := common.GetBitmapSize(consensusSize)

	return len(bitmap) > 0 && len(bitmap) <= expectedBitmapSize
}

func (pc *fieldsSizeChecker) isHeaderHashSizeValid(headerHash []byte) bool {
	return len(headerHash) > 0 && len(headerHash) <= pc.hasher.Size()
}

func (pc *fieldsSizeChecker) isAggregatedSigSizeValid(aggSig []byte) bool {
	// TODO: add upper bound check
	return len(aggSig) > 0
}

// IsInterfaceNil -
func (pc *fieldsSizeChecker) IsInterfaceNil() bool {
	return pc == nil
}
