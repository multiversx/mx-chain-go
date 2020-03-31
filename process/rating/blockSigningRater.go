package rating

import (
	"sort"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// BlockSigningRaterAndListIndexer defines the behaviour of a struct able to do ratings for validators
type BlockSigningRaterAndListIndexer struct {
	sharding.RatingReader
	sharding.ListIndexUpdaterHandler
	startRating                 uint32
	maxRating                   uint32
	minRating                   uint32
	proposerIncreaseRatingStep  int32
	proposerDecreaseRatingStep  int32
	validatorIncreaseRatingStep int32
	validatorDecreaseRatingStep int32
	ratingChances               []sharding.RatingChance
}

// NewBlockSigningRaterAndListIndexer creates a new PeerAccountListAndRatingHandler of Type BlockSigningRaterAndListIndexer
func NewBlockSigningRaterAndListIndexer(ratingsData *economics.RatingsData) (*BlockSigningRaterAndListIndexer, error) {
	if ratingsData.MinRating() < 1 {
		return nil, process.ErrMinRatingSmallerThanOne
	}
	if ratingsData.MinRating() > ratingsData.MaxRating() {
		return nil, process.ErrMaxRatingIsSmallerThanMinRating
	}
	if ratingsData.MaxRating() < ratingsData.StartRating() || ratingsData.MinRating() > ratingsData.StartRating() {
		return nil, process.ErrStartRatingNotBetweenMinAndMax
	}
	if len(ratingsData.SelectionChances()) == 0 {
		return nil, process.ErrNoChancesProvided
	}

	ratingChances := make([]sharding.RatingChance, len(ratingsData.SelectionChances()))

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

	return &BlockSigningRaterAndListIndexer{
		startRating:                 ratingsData.StartRating(),
		minRating:                   ratingsData.MinRating(),
		maxRating:                   ratingsData.MaxRating(),
		proposerIncreaseRatingStep:  int32(ratingsData.ProposerIncreaseRatingStep()),
		proposerDecreaseRatingStep:  int32(0 - ratingsData.ProposerDecreaseRatingStep()),
		validatorIncreaseRatingStep: int32(ratingsData.ValidatorIncreaseRatingStep()),
		validatorDecreaseRatingStep: int32(0 - ratingsData.ValidatorDecreaseRatingStep()),
		RatingReader:                NewDisabledRatingReader(ratingsData.StartRating()),
		ratingChances:               ratingChances,
		ListIndexUpdaterHandler:     &DisabledListIndexUpdater{},
	}, nil
}

func (bsr *BlockSigningRaterAndListIndexer) computeRating(ratingStep int32, val uint32) uint32 {
	newVal := int64(val) + int64(ratingStep)
	if newVal < int64(bsr.minRating) {
		return bsr.minRating
	}
	if newVal > int64(bsr.maxRating) {
		return bsr.maxRating
	}

	return uint32(newVal)
}

// GetRating returns the Rating for the specified public key
func (bsr *BlockSigningRaterAndListIndexer) GetRating(pk string) uint32 {
	return bsr.RatingReader.GetRating(pk)
}

// SetRatingReader sets the Reader that can read ratings
func (bsr *BlockSigningRaterAndListIndexer) SetRatingReader(reader sharding.RatingReader) {
	if !check.IfNil(reader) {
		bsr.RatingReader = reader
	}
}

// UpdateRatingFromTempRating returns the TempRating for the specified public keys
func (bsr *BlockSigningRaterAndListIndexer) UpdateRatingFromTempRating(pks []string) error {
	return bsr.RatingReader.UpdateRatingFromTempRating(pks)
}

// SetListIndexUpdater sets the list index update
func (bsr *BlockSigningRaterAndListIndexer) SetListIndexUpdater(updater sharding.ListIndexUpdaterHandler) {
	if !check.IfNil(updater) {
		bsr.ListIndexUpdaterHandler = updater
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (bsr *BlockSigningRaterAndListIndexer) IsInterfaceNil() bool {
	return bsr == nil
}

// GetStartRating gets the StartingRating
func (bsr *BlockSigningRaterAndListIndexer) GetStartRating() uint32 {
	return bsr.startRating
}

// ComputeIncreaseProposer computes the new rating for the increaseLeader
func (bsr *BlockSigningRaterAndListIndexer) ComputeIncreaseProposer(val uint32) uint32 {
	return bsr.computeRating(bsr.proposerIncreaseRatingStep, val)
}

// ComputeDecreaseProposer computes the new rating for the decreaseLeader
func (bsr *BlockSigningRaterAndListIndexer) ComputeDecreaseProposer(val uint32) uint32 {
	return bsr.computeRating(bsr.proposerDecreaseRatingStep, val)
}

// ComputeIncreaseValidator computes the new rating for the increaseValidator
func (bsr *BlockSigningRaterAndListIndexer) ComputeIncreaseValidator(val uint32) uint32 {
	return bsr.computeRating(bsr.validatorIncreaseRatingStep, val)
}

// ComputeDecreaseValidator computes the new rating for the decreaseValidator
func (bsr *BlockSigningRaterAndListIndexer) ComputeDecreaseValidator(val uint32) uint32 {
	return bsr.computeRating(bsr.validatorDecreaseRatingStep, val)
}

// UpdateListAndIndex will update the list and the index for a peer
func (bsr *BlockSigningRaterAndListIndexer) UpdateListAndIndex(pubKey string, shardID uint32, list string, index uint32) error {
	return bsr.ListIndexUpdaterHandler.UpdateListAndIndex(pubKey, shardID, list, index)
}

// GetChance returns the RatingChance for the pk
func (bsr *BlockSigningRaterAndListIndexer) GetChance(currentRating uint32) uint32 {
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
