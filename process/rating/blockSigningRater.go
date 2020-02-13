package rating

import (
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
}

// NewBlockSigningRaterAndListIndexer creates a new PeerAccountListAndRatingHandler of Type BlockSigningRaterAndListIndexer
func NewBlockSigningRaterAndListIndexer(ratingsData *economics.RatingsData) (*BlockSigningRaterAndListIndexer, error) {
	if ratingsData.MinRating() > ratingsData.MaxRating() {
		return nil, process.ErrMaxRatingIsSmallerThanMinRating
	}
	if ratingsData.MaxRating() < ratingsData.StartRating() || ratingsData.MinRating() > ratingsData.StartRating() {
		return nil, process.ErrStartRatingNotBetweenMinAndMax
	}

	return &BlockSigningRaterAndListIndexer{
		startRating:                 ratingsData.StartRating(),
		minRating:                   ratingsData.MinRating(),
		maxRating:                   ratingsData.MaxRating(),
		proposerIncreaseRatingStep:  int32(ratingsData.ProposerIncreaseRatingStep()),
		proposerDecreaseRatingStep:  int32(0 - ratingsData.ProposerDecreaseRatingStep()),
		validatorIncreaseRatingStep: int32(ratingsData.ValidatorIncreaseRatingStep()),
		validatorDecreaseRatingStep: int32(0 - ratingsData.ValidatorDecreaseRatingStep()),
		RatingReader:                &NilRatingReader{},
		ListIndexUpdaterHandler:     &NilListIndexUpdater{},
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

// GetRatings gets all the ratings that the current rater has
func (bsr *BlockSigningRaterAndListIndexer) GetRatings(addresses []string) map[string]uint32 {
	return bsr.RatingReader.GetRatings(addresses)
}

// SetRatingReader sets the Reader that can read ratings
func (bsr *BlockSigningRaterAndListIndexer) SetRatingReader(reader sharding.RatingReader) {
	if !check.IfNil(reader) {
		bsr.RatingReader = reader
	}
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
func (bsr *BlockSigningRaterAndListIndexer) UpdateListAndIndex(pubKey string, list string, index int) error {
	return bsr.ListIndexUpdaterHandler.UpdateListAndIndex(pubKey, list, index)
}
