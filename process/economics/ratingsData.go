package economics

import (
	"github.com/ElrondNetwork/elrond-go/process"
)

// RatingsData will store information about ratingsComputation
type RatingsData struct {
	startRating   int64
	maxRating     int64
	minRating     int64
	ratingType    string
	ratingOptions map[string]int64
}

func NewRatingsData(
	startRating int64,
	minRating int64,
	maxRating int64,
	ratingType string,
	ratingValues map[string]int64) (*RatingsData, error) {
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
func (rd *RatingsData) StartRating() int64 {
	return rd.startRating
}

// MaxRating will return the max rating
func (rd *RatingsData) MaxRating() int64 {
	return rd.maxRating
}

// MinRating will return the min rating
func (rd *RatingsData) MinRating() int64 {
	return rd.minRating
}

// RatingOptions will return the options for rating
func (rd *RatingsData) RatingOptions() map[string]int64 {
	return rd.ratingOptions
}

// RatingType will return the type for rating
func (rd *RatingsData) RatingType() string {
	return rd.ratingType
}
