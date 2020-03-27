package rating

import (
	"fmt"
	"math"
	"sort"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var log = logger.GetOrCreate("process/rating")

// BlockSigningRater defines the behaviour of a struct able to do ratings for validators
type BlockSigningRater struct {
	sharding.RatingReader
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
	if ratingsData.MinRating() < 1 {
		return nil, process.ErrMinRatingSmallerThanOne
	}
	if ratingsData.MinRating() > ratingsData.MaxRating() {
		return nil, fmt.Errorf("%w: minRating: %v, maxRating: %v",
			process.ErrMaxRatingIsSmallerThanMinRating,
			ratingsData.MinRating(),
			ratingsData.MaxRating())
	}
	if ratingsData.MaxRating() < ratingsData.StartRating() || ratingsData.MinRating() > ratingsData.StartRating() {
		return nil, fmt.Errorf("%w: minRating: %v, startRating: %v, maxRating: %v",
			process.ErrStartRatingNotBetweenMinAndMax,
			ratingsData.MinRating(),
			ratingsData.StartRating(),
			ratingsData.MaxRating())
	}
	if len(ratingsData.SelectionChances()) == 0 {
		return nil, process.ErrNoChancesProvided
	}
	if ratingsData.SignedBlocksThreshold() < 0 || ratingsData.SignedBlocksThreshold() > 1 {
		return nil, fmt.Errorf("%w signedBlocksThreshold: %v",
			process.ErrSignedBlocksThresholdNotBetweenZeroAndOne,
			ratingsData.SignedBlocksThreshold())
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
		RatingReader:            NewNilRatingReader(ratingsData.StartRating()),
		ratingChances:           ratingChances,
	}, nil
}

func (bsr *BlockSigningRater) computeRating(ratingStep int32, currentRating uint32) uint32 {
	log.Debug("computing rating", "rating", currentRating, "step", ratingStep)
	newVal := int64(currentRating) + int64(ratingStep)
	if newVal < int64(bsr.minRating) {
		return bsr.minRating
	}
	if newVal > int64(bsr.maxRating) {
		return bsr.maxRating
	}

	return uint32(newVal)
}

// GetRating returns the Rating for the specified public key
func (bsr *BlockSigningRater) GetRating(pk string) uint32 {
	return bsr.RatingReader.GetRating(pk)
}

// SetRatingReader sets the Reader that can read ratings
func (bsr *BlockSigningRater) SetRatingReader(reader sharding.RatingReader) {
	if !check.IfNil(reader) {
		bsr.RatingReader = reader
	}
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

// ComputeIncreaseProposer computes the new rating for the increaseLeader
func (bsr *BlockSigningRater) RevertIncreaseValidator(shardId uint32, currentRating uint32, nrReverts uint32) uint32 {
	log.Trace("RevertIncreaseValidator", "shardId", shardId, "currentRating", currentRating, "nrReverts", nrReverts)
	var ratingStep int32
	if shardId == core.MetachainShardId {
		ratingStep = bsr.metaRatingsStepHandler.ValidatorIncreaseRatingStep()
	} else {
		ratingStep = bsr.shardRatingsStepHandler.ValidatorIncreaseRatingStep()
	}

	decreaseValue := -int64(ratingStep) * int64(nrReverts)
	// overflow
	if decreaseValue > 0 || decreaseValue < math.MinInt32 {
		decreaseValue = math.MinInt32
	}
	return bsr.computeRating(int32(decreaseValue), currentRating)
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
	computedIncrease := float64(proposerDecreaseRatingStep) * math.Pow(float64(consecutiveBlocksPenalty), float64(consecutiveMisses))
	var consecutiveMissesIncrease int32
	if computedIncrease < math.MinInt32 || computedIncrease >= 0 {
		consecutiveMissesIncrease = math.MinInt32
	} else {
		consecutiveMissesIncrease = int32(computedIncrease)
	}

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

// GetChance returns the chances modifier for the pk
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
