package rating

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// BlockSigningRater defines the behaviour of a struct able to do ratings for validators
type BlockSigningRater struct {
	sharding.RatingReader
	startRating                 uint32
	maxRating                   uint32
	minRating                   uint32
	proposerIncreaseRatingStep  int32
	proposerDecreaseRatingStep  int32
	validatorIncreaseRatingStep int32
	validatorDecreaseRatingStep int32
}

//NewBlockSigningRater creates a new RaterHandler of Type BlockSigningRater
func NewBlockSigningRater(ratingsData *economics.RatingsData) (*BlockSigningRater, error) {
	if ratingsData.MinRating() > ratingsData.MaxRating() {
		return nil, process.ErrMaxRatingIsSmallerThanMinRating
	}
	if ratingsData.MaxRating() < ratingsData.StartRating() || ratingsData.MinRating() > ratingsData.StartRating() {
		return nil, process.ErrStartRatingNotBetweenMinAndMax
	}

	return &BlockSigningRater{
		startRating:                 ratingsData.StartRating(),
		minRating:                   ratingsData.MinRating(),
		maxRating:                   ratingsData.MaxRating(),
		proposerIncreaseRatingStep:  int32(ratingsData.ProposerIncreaseRatingStep()),
		proposerDecreaseRatingStep:  int32(0 - ratingsData.ProposerDecreaseRatingStep()),
		validatorIncreaseRatingStep: int32(ratingsData.ValidatorIncreaseRatingStep()),
		validatorDecreaseRatingStep: int32(0 - ratingsData.ValidatorDecreaseRatingStep()),
		RatingReader:                &NilRatingReader{},
	}, nil
}

func (bsr *BlockSigningRater) computeRating(ratingStep int32, val uint32) uint32 {
	newVal := int64(val) + int64(ratingStep)
	if newVal < int64(bsr.minRating) {
		return bsr.minRating
	}
	if newVal > int64(bsr.maxRating) {
		return bsr.maxRating
	}

	return uint32(newVal)
}

//GetRating returns the Rating for the specified public key
func (bsr *BlockSigningRater) GetRating(pk string) uint32 {
	return bsr.RatingReader.GetRating(pk)
}

//GetRatings gets all the ratings that the current rater has
func (bsr *BlockSigningRater) GetRatings(addresses []string) map[string]uint32 {
	return bsr.RatingReader.GetRatings(addresses)
}

//SetRatingReader sets the Reader that can read ratings
func (bsr *BlockSigningRater) SetRatingReader(reader sharding.RatingReader) {
	if !check.IfNil(reader) {
		bsr.RatingReader = reader
	}
}

//SetRatingReader sets the Reader that can read ratings
func (bsr *BlockSigningRater) IsInterfaceNil() bool {
	return bsr == nil
}

//GetStartRating gets the StartingRating
func (bsr *BlockSigningRater) GetStartRating() uint32 {
	return bsr.startRating
}

//ComputeIncreaseProposer computes the new rating for the increaseLeader
func (bsr *BlockSigningRater) ComputeIncreaseProposer(val uint32) uint32 {
	return bsr.computeRating(bsr.proposerIncreaseRatingStep, val)
}

//ComputeDecreaseProposer computes the new rating for the decreaseLeader
func (bsr *BlockSigningRater) ComputeDecreaseProposer(val uint32) uint32 {
	return bsr.computeRating(bsr.proposerDecreaseRatingStep, val)
}

//ComputeIncreaseValidator computes the new rating for the increaseValidator
func (bsr *BlockSigningRater) ComputeIncreaseValidator(val uint32) uint32 {
	return bsr.computeRating(bsr.validatorIncreaseRatingStep, val)
}

//ComputeDecreaseValidator computes the new rating for the decreaseValidator
func (bsr *BlockSigningRater) ComputeDecreaseValidator(val uint32) uint32 {
	return bsr.computeRating(bsr.validatorDecreaseRatingStep, val)
}
