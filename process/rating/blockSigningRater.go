package rating

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// BlockSigningRater defines the behaviour of a struct able to do ratings for validators
type BlockSigningRater struct {
	sharding.RatingReader
	startRating      uint32
	maxRating        uint32
	minRating        uint32
	ratingOptions    map[string]int32
	ratingOptionKeys []string
}

//NewBlockSigningRater creates a new Rater of Type BlockSigningRater
func NewBlockSigningRater(ratingsData *economics.RatingsData) (*BlockSigningRater, error) {
	if ratingsData.MinRating() > ratingsData.MaxRating() {
		return nil, process.ErrMaxRatingIsSmallerThanMinRating
	}
	if ratingsData.MaxRating() < ratingsData.StartRating() || ratingsData.MinRating() > ratingsData.StartRating() {
		return nil, process.ErrStartRatingNotBetweenMinAndMax
	}

	ratingOptionKeys := make([]string, 0)
	for key := range ratingsData.RatingOptions() {
		ratingOptionKeys = append(ratingOptionKeys, key)
	}

	return &BlockSigningRater{
		ratingOptions:    ratingsData.RatingOptions(),
		startRating:      ratingsData.StartRating(),
		minRating:        ratingsData.MinRating(),
		maxRating:        ratingsData.MaxRating(),
		ratingOptionKeys: ratingOptionKeys,
	}, nil
}

func (bsr *BlockSigningRater) ComputeRating(pk, ratingKey string, val uint32) uint32 {
	newVal := int64(val) + int64(bsr.ratingOptions[ratingKey])
	if newVal < int64(bsr.minRating) {
		return bsr.minRating
	}
	if newVal > int64(bsr.maxRating) {
		return bsr.maxRating
	}

	return uint32(newVal)
}

//GetRatingOptionKeys gets all the ratings  keys for the options
func (bsr *BlockSigningRater) GetRatingOptionKeys() []string {
	return bsr.ratingOptionKeys
}

//GetRating returns the Rating for the specified public key
func (bsr *BlockSigningRater) GetRating(pk string) uint32 {
	if bsr.RatingReader == nil {
		return 0
	}
	return bsr.RatingReader.GetRating(pk)
}

//GetRatings gets all the ratings that the current rater has
func (bsr *BlockSigningRater) GetRatings(addresses []string) map[string]uint32 {
	if bsr.RatingReader == nil {
		return map[string]uint32{}
	}
	return bsr.RatingReader.GetRatings(addresses)
}

//SetRatingReader sets the Reader that can read ratings
func (bsr *BlockSigningRater) SetRatingReader(reader sharding.RatingReader) {
	bsr.RatingReader = reader
}
