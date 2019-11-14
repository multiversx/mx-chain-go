package rating

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
)

// BlockSigningRater defines the behaviour of a struct able to do ratings for validators
type BlockSigningRater struct {
	startRating   int64
	maxRating     int64
	minRating     int64
	ratings       map[string]int64
	ratingOptions map[string]int64
}

//NewBlockSigningRater creates a new Rater of Type BlockSigningRater
func NewBlockSigningRater(ratingsData *economics.RatingsData) (*BlockSigningRater, error) {
	if ratingsData.MinRating() > ratingsData.MaxRating() {
		return nil, process.ErrMaxRatingIsSmallerThanMinRating
	}
	if ratingsData.MaxRating() < ratingsData.StartRating() || ratingsData.MinRating() > ratingsData.StartRating() {
		return nil, process.ErrStartRatingNotBetweenMinAndMax
	}
	return &BlockSigningRater{
		ratings:       make(map[string]int64),
		ratingOptions: ratingsData.RatingOptions(),
		startRating:   ratingsData.StartRating(),
		minRating:     ratingsData.MinRating(),
		maxRating:     ratingsData.MaxRating(),
	}, nil
}

//UpdateRatings does an update on the ratings based on the new updatevalues
func (bsr *BlockSigningRater) UpdateRatings(updateValues map[string][]string) {
	for k, v := range updateValues {
		for _, pk := range v {
			bsr.updateRating(k, pk)
		}
	}
}

func (bsr *BlockSigningRater) updateRating(ratingKey string, pk string) {
	val, ok := bsr.ratings[pk]
	if !ok {
		val = bsr.startRating
	}
	val += bsr.ratingOptions[ratingKey]
	if val < bsr.minRating {
		val = bsr.minRating
	}
	if val > bsr.maxRating {
		val = bsr.maxRating
	}

	bsr.ratings[pk] = val
}

//GetRating returns the Rating for the specified public key
func (bsr *BlockSigningRater) GetRating(pk string) int64 {
	_, ok := bsr.ratings[pk]
	if !ok {
		bsr.ratings[pk] = bsr.startRating
	}
	return bsr.ratings[pk]
}

//GetRatings gets all the ratings that the current rater has
func (bsr *BlockSigningRater) GetRatings() map[string]int64 {
	return bsr.ratings
}
