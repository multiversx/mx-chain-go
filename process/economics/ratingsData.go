package economics

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

// RatingsData will store information about ratingsComputation
type RatingsData struct {
	startRating   uint32
	maxRating     uint32
	minRating     uint32
	ratingType    string
	ratingOptions map[string]int32
}

func NewRatingsData(
	startRating uint32,
	minRating uint32,
	maxRating uint32,
	ratingType string,
	ratingValues map[string]int32) (*RatingsData, error) {
	if minRating > maxRating {
		return nil, process.ErrMaxRatingIsSmallerThanMinRating
	}
	if maxRating < startRating || minRating > startRating {
		return nil, process.ErrStartRatingNotBetweenMinAndMax
	}

	return &RatingsData{
		startRating:   startRating,
		maxRating:     maxRating,
		minRating:     minRating,
		ratingOptions: ratingValues,
		ratingType:    ratingType,
	}, nil
}

// StartRating will return the start rating
func (rd *RatingsData) StartRating() uint32 {
	return rd.startRating
}

// MaxRating will return the max rating
func (rd *RatingsData) MaxRating() uint32 {
	return rd.maxRating
}

// MinRating will return the min rating
func (rd *RatingsData) MinRating() uint32 {
	return rd.minRating
}

// RatingOptions will return the options for rating
func (rd *RatingsData) RatingOptions() map[string]int32 {
	return rd.ratingOptions
}

// RatingType will return the type for rating
func (rd *RatingsData) RatingType() string {
	return rd.ratingType
}
