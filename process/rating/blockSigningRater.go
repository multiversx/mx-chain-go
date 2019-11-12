package rating

import "github.com/ElrondNetwork/elrond-go/process/economics"

// BlockSigningRater defines the behaviour of a struct able to do ratings for validators
type BlockSigningRater struct {
	startRating                     uint64
	maxRating                       uint64
	minRating                       uint64
	increaseRatingStep              uint64
	decreaseRatingStep              uint64
	proposerExtraIncreaseRatingStep uint64
	proposerExtraDecreaseRatingStep uint64
	ratings                         map[string]int64
}

func NewBlockSigningRater(ratingsData *economics.RatingsData) *BlockSigningRater {
	return &BlockSigningRater{
		ratings:                         make(map[string]int64),
		startRating:                     ratingsData.StartRating(),
		minRating:                       ratingsData.MinRating(),
		maxRating:                       ratingsData.MaxRating(),
		increaseRatingStep:              ratingsData.IncreaseRatingStep(),
		decreaseRatingStep:              ratingsData.DecreaseRatingStep(),
		proposerExtraIncreaseRatingStep: ratingsData.ProposerExtraIncreaseRatingStep(),
		proposerExtraDecreaseRatingStep: ratingsData.ProposerExtraDecreaseRatingStep(),
	}
}

func (rc *BlockSigningRater) UpdateRating() error {
	return nil
}
