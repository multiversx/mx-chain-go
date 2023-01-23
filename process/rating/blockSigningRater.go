package rating

import (
	"fmt"
	"math"
	"math/big"
	"sort"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var _ sharding.PeerAccountListAndRatingHandler = (*BlockSigningRater)(nil)

var log = logger.GetOrCreate("process/rating")

const (
	maxDecreaseValue = math.MinInt32
)

// BlockSigningRater defines the behaviour of a struct able to do ratings for validators
type BlockSigningRater struct {
	startRating             uint32
	signedBlocksThreshold   float32
	maxRating               uint32
	minRating               uint32
	shardRatingsStepHandler process.RatingsStepHandler
	metaRatingsStepHandler  process.RatingsStepHandler
	ratingChances           []process.RatingChanceHandler
}

// NewBlockSigningRater creates a new RaterHandler of Type BlockSigningRater
func NewBlockSigningRater(ratingsData process.RatingsInfoHandler) (*BlockSigningRater, error) {
	err := verifyRatingsData(ratingsData)
	if err != nil {
		return nil, err
	}

	ratingChances := make([]process.RatingChanceHandler, len(ratingsData.SelectionChances()))

	for i, chance := range ratingsData.SelectionChances() {
		ratingChances[i] = &selectionChance{
			maxThreshold:     chance.GetMaxThreshold(),
			chancePercentage: chance.GetChancePercent(),
		}
	}

	sort.Slice(ratingChances, func(i, j int) bool {
		return ratingChances[i].GetMaxThreshold() < ratingChances[j].GetMaxThreshold()
	})

	if ratingChances[0].GetMaxThreshold() > 0 {
		return nil, process.ErrNilMinChanceIfZero
	}

	for i := 1; i < len(ratingChances); i++ {
		if ratingChances[i-1].GetMaxThreshold() == ratingChances[i].GetMaxThreshold() {
			return nil, process.ErrDuplicateThreshold
		}
	}

	if ratingChances[len(ratingChances)-1].GetMaxThreshold() != ratingsData.MaxRating() {
		return nil, process.ErrNoChancesForMaxThreshold
	}

	return &BlockSigningRater{
		startRating:             ratingsData.StartRating(),
		minRating:               ratingsData.MinRating(),
		maxRating:               ratingsData.MaxRating(),
		signedBlocksThreshold:   ratingsData.SignedBlocksThreshold(),
		shardRatingsStepHandler: ratingsData.ShardChainRatingsStepHandler(),
		metaRatingsStepHandler:  ratingsData.MetaChainRatingsStepHandler(),
		ratingChances:           ratingChances,
	}, nil
}

func verifyRatingsData(ratingsData process.RatingsInfoHandler) error {
	if check.IfNil(ratingsData) {
		return process.ErrNilRatingsInfoHandler
	}
	if ratingsData.MinRating() < 1 {
		return process.ErrMinRatingSmallerThanOne
	}
	if ratingsData.MinRating() > ratingsData.MaxRating() {
		return fmt.Errorf("%w: minRating: %v, maxRating: %v",
			process.ErrMaxRatingIsSmallerThanMinRating,
			ratingsData.MinRating(),
			ratingsData.MaxRating())
	}
	if ratingsData.MaxRating() < ratingsData.StartRating() || ratingsData.MinRating() > ratingsData.StartRating() {
		return fmt.Errorf("%w: minRating: %v, startRating: %v, maxRating: %v",
			process.ErrStartRatingNotBetweenMinAndMax,
			ratingsData.MinRating(),
			ratingsData.StartRating(),
			ratingsData.MaxRating())
	}
	if len(ratingsData.SelectionChances()) == 0 {
		return process.ErrNoChancesProvided
	}
	if ratingsData.SignedBlocksThreshold() < 0 || ratingsData.SignedBlocksThreshold() > 1 {
		return fmt.Errorf("%w signedBlocksThreshold: %v",
			process.ErrSignedBlocksThresholdNotBetweenZeroAndOne,
			ratingsData.SignedBlocksThreshold())
	}
	if ratingsData.MetaChainRatingsStepHandler().ConsecutiveMissedBlocksPenalty() < 1 {
		return fmt.Errorf("%w: metaChain consecutiveMissedBlocksPenalty: %v",
			process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne,
			ratingsData.MetaChainRatingsStepHandler().ConsecutiveMissedBlocksPenalty())
	}
	if ratingsData.ShardChainRatingsStepHandler().ConsecutiveMissedBlocksPenalty() < 1 {
		return fmt.Errorf("%w: shardChain consecutiveMissedBlocksPenalty: %v",
			process.ErrConsecutiveMissedBlocksPenaltyLowerThanOne,
			ratingsData.ShardChainRatingsStepHandler().ConsecutiveMissedBlocksPenalty())
	}
	if ratingsData.ShardChainRatingsStepHandler().ProposerDecreaseRatingStep() > -1 || ratingsData.ShardChainRatingsStepHandler().ValidatorDecreaseRatingStep() > -1 {
		return fmt.Errorf("%w: shardChain decrease steps - proposer: %v, validator: %v",
			process.ErrDecreaseRatingsStepMoreThanMinusOne,
			ratingsData.ShardChainRatingsStepHandler().ProposerDecreaseRatingStep(),
			ratingsData.ShardChainRatingsStepHandler().ValidatorDecreaseRatingStep())
	}
	if ratingsData.MetaChainRatingsStepHandler().ProposerDecreaseRatingStep() > -1 || ratingsData.MetaChainRatingsStepHandler().ValidatorDecreaseRatingStep() > -1 {
		return fmt.Errorf("%w: metachain decrease steps - proposer: %v, validator: %v",
			process.ErrDecreaseRatingsStepMoreThanMinusOne,
			ratingsData.MetaChainRatingsStepHandler().ProposerDecreaseRatingStep(),
			ratingsData.MetaChainRatingsStepHandler().ValidatorDecreaseRatingStep())
	}
	return nil
}

func (bsr *BlockSigningRater) computeRating(ratingStep int32, currentRating uint32) uint32 {
	log.Trace("computing rating", "rating", currentRating, "step", ratingStep)
	newVal := int64(currentRating) + int64(ratingStep)
	if newVal < int64(bsr.minRating) {
		return bsr.minRating
	}
	if newVal > int64(bsr.maxRating) {
		return bsr.maxRating
	}

	return uint32(newVal)
}

// IsInterfaceNil returns true if there is no value under the interface
func (bsr *BlockSigningRater) IsInterfaceNil() bool {
	return bsr == nil
}

// GetStartRating gets the StartingRating
func (bsr *BlockSigningRater) GetStartRating() uint32 {
	return bsr.startRating
}

// GetSignedBlocksThreshold ets the signedBlocksThreshold
func (bsr *BlockSigningRater) GetSignedBlocksThreshold() float32 {
	return bsr.signedBlocksThreshold
}

// ComputeIncreaseProposer computes the new rating for the increaseLeader
func (bsr *BlockSigningRater) ComputeIncreaseProposer(shardId uint32, currentRating uint32) uint32 {
	log.Trace("ComputeIncreaseProposer", "shardId", shardId, "currentRating", currentRating)
	var ratingStep int32
	if shardId == core.MetachainShardId {
		ratingStep = bsr.metaRatingsStepHandler.ProposerIncreaseRatingStep()
	} else {
		ratingStep = bsr.shardRatingsStepHandler.ProposerIncreaseRatingStep()
	}

	return bsr.computeRating(ratingStep, currentRating)
}

// RevertIncreaseValidator computes the new rating based on how many reverts have to be done for the validator
func (bsr *BlockSigningRater) RevertIncreaseValidator(shardId uint32, currentRating uint32, nrReverts uint32) uint32 {
	log.Trace("RevertIncreaseValidator", "shardId", shardId, "currentRating", currentRating, "nrReverts", nrReverts)
	var ratingStep int32
	if shardId == core.MetachainShardId {
		ratingStep = bsr.metaRatingsStepHandler.ValidatorIncreaseRatingStep()
	} else {
		ratingStep = bsr.shardRatingsStepHandler.ValidatorIncreaseRatingStep()
	}

	decreaseValueBigInt := big.NewInt(0).Mul(big.NewInt(int64(-ratingStep)), big.NewInt(int64(nrReverts)))
	decreaseInt := decreaseValueBigInt.Int64()

	var decreaseValue int32
	if decreaseInt < maxDecreaseValue || decreaseInt > 0 {
		decreaseValue = maxDecreaseValue
	} else {
		decreaseValue = int32(decreaseValueBigInt.Int64())
	}
	return bsr.computeRating(decreaseValue, currentRating)
}

// ComputeDecreaseProposer computes the new rating for the decreaseLeader
func (bsr *BlockSigningRater) ComputeDecreaseProposer(shardId uint32, currentRating uint32, consecutiveMisses uint32) uint32 {
	log.Trace("ComputeDecreaseProposer", "shardId", shardId, "currentRating", currentRating, "consecutiveMisses", consecutiveMisses)
	var proposerDecreaseRatingStep int32

	var consecutiveBlocksPenalty float32
	if shardId == core.MetachainShardId {
		proposerDecreaseRatingStep = bsr.metaRatingsStepHandler.ProposerDecreaseRatingStep()
		consecutiveBlocksPenalty = bsr.metaRatingsStepHandler.ConsecutiveMissedBlocksPenalty()
	} else {
		proposerDecreaseRatingStep = bsr.shardRatingsStepHandler.ProposerDecreaseRatingStep()
		consecutiveBlocksPenalty = bsr.shardRatingsStepHandler.ConsecutiveMissedBlocksPenalty()
	}

	var consecutiveMissesIncrease int32
	computedFloat := float64(proposerDecreaseRatingStep)

	for i := uint32(0); i < consecutiveMisses; i++ {
		computedFloat *= float64(consecutiveBlocksPenalty)
		if computedFloat < maxDecreaseValue {
			computedFloat = maxDecreaseValue
			break
		}
	}

	consecutiveMissesIncrease = int32(computedFloat)
	return bsr.computeRating(consecutiveMissesIncrease, currentRating)
}

// ComputeIncreaseValidator computes the new rating for the increaseValidator
func (bsr *BlockSigningRater) ComputeIncreaseValidator(shardId uint32, currentRating uint32) uint32 {
	log.Trace("ComputeIncreaseValidator", "shardId", shardId, "currentRating", currentRating)
	var ratingStep int32
	if shardId == core.MetachainShardId {
		ratingStep = bsr.metaRatingsStepHandler.ValidatorIncreaseRatingStep()
	} else {
		ratingStep = bsr.shardRatingsStepHandler.ValidatorIncreaseRatingStep()
	}
	return bsr.computeRating(ratingStep, currentRating)
}

// ComputeDecreaseValidator computes the new rating for the decreaseValidator
func (bsr *BlockSigningRater) ComputeDecreaseValidator(shardId uint32, currentRating uint32) uint32 {
	log.Trace("ComputeDecreaseValidator", "shardId", shardId, "currentRating", currentRating)
	var ratingStep int32
	if shardId == core.MetachainShardId {
		ratingStep = bsr.metaRatingsStepHandler.ValidatorDecreaseRatingStep()
	} else {
		ratingStep = bsr.shardRatingsStepHandler.ValidatorDecreaseRatingStep()
	}
	return bsr.computeRating(ratingStep, currentRating)
}

// GetChance returns the chances modifier for the current rating
func (bsr *BlockSigningRater) GetChance(currentRating uint32) uint32 {
	chance := bsr.ratingChances[0].GetChancePercentage()
	for i := 0; i < len(bsr.ratingChances); i++ {
		currentChance := bsr.ratingChances[i]
		if currentRating > currentChance.GetMaxThreshold() {
			continue
		}
		chance = currentChance.GetChancePercentage()
		break
	}
	return chance
}
