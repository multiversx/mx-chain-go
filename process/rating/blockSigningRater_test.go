package rating_test

import (
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/stretchr/testify/assert"

	"testing"
)

const (
	validatorIncreaseRatingStep   = "validatorIncreaseRatingStep"
	validatorDecreaseRatingStep   = "validatorDecreaseRatingStep"
	proposerIncreaseRatingStepKey = "proposerIncreaseRatingStep"
	proposerDecreaseRatingStepKey = "proposerDecreaseRatingStep"
)

func createDefaultRatingsData() *economics.RatingsData {
	ratingValues := make(map[string]int32, 0)
	ratingValues[validatorIncreaseRatingStep] = 1
	ratingValues[validatorDecreaseRatingStep] = 2
	ratingValues[proposerIncreaseRatingStepKey] = 3
	ratingValues[proposerDecreaseRatingStepKey] = 4

	ratingsData, _ := economics.NewRatingsData(uint32(5), uint32(1), uint32(10), "mockRater", ratingValues)
	return ratingsData
}

func createDefaultRatingReader(ratingsMap map[string]uint32) *mock.RatingReaderMock {
	rrm := &mock.RatingReaderMock{
		RatingsMap: ratingsMap,
		GetRatingCalled: func(s string) uint32 {
			value, ok := ratingsMap[s]
			if !ok {
				return 0
			}
			return value
		},
		GetRatingsCalled: func(pks []string) map[string]uint32 {
			newMap := make(map[string]uint32)
			for k, v := range ratingsMap {
				for _, pk := range pks {
					if k == pk {
						newMap[k] = v
					}
				}
			}
			return newMap
		},
	}

	return rrm
}

func createUpdateMap(ratingPk string, updateStep string) map[string][]string {
	updatedPeers := make(map[string][]string, 0)
	peerList := make([]string, 0)
	peerList = append(peerList, ratingPk)
	updatedPeers[updateStep] = peerList
	return updatedPeers
}

func setupRater(rd *economics.RatingsData, pk string, initialRating uint32) *rating.BlockSigningRater {
	bsr, _ := rating.NewBlockSigningRater(rd)
	ratingPk := pk
	ratingsMap := make(map[string]uint32)
	ratingsMap[ratingPk] = initialRating
	rrm := createDefaultRatingReader(ratingsMap)
	bsr.SetRatingReader(rrm)

	return bsr
}

func TestBlockSigningRater_GetRatingWithUnknownPkShoudReturnError(t *testing.T) {
	rd := createDefaultRatingsData()
	bsr, _ := rating.NewBlockSigningRater(rd)

	rrm := createDefaultRatingReader(make(map[string]uint32))
	bsr.SetRatingReader(rrm)

	rt := bsr.GetRating("test")

	assert.Equal(t, uint32(0), rt)
}

func TestBlockSigningRater_GetRatingWithKnownPkShoudReturnSetRating(t *testing.T) {
	rd := createDefaultRatingsData()

	bsr, _ := rating.NewBlockSigningRater(rd)

	ratingPk := "test"
	ratingValue := uint32(5)

	ratingsMap := make(map[string]uint32)
	ratingsMap[ratingPk] = ratingValue
	rrd := createDefaultRatingReader(ratingsMap)
	bsr.SetRatingReader(rrd)
	rt := bsr.GetRating(ratingPk)

	assert.Equal(t, ratingValue, rt)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenProposed(t *testing.T) {
	pk := "test"
	initialRatingValue := uint32(5)
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)
	computedRating := bsr.ComputeRating(pk, proposerIncreaseRatingStepKey, initialRatingValue)

	expectedValue := uint32(int32(initialRatingValue) + rd.RatingOptions()[proposerIncreaseRatingStepKey])

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenValidator(t *testing.T) {
	pk := "test"
	initialRatingValue := uint32(5)
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)

	computedRating := bsr.ComputeRating(pk, validatorIncreaseRatingStep, initialRatingValue)

	expectedValue := uint32(int32(initialRatingValue) + rd.RatingOptions()[validatorIncreaseRatingStep])

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenValidatorButNotAccepted(t *testing.T) {
	pk := "test"
	initialRatingValue := uint32(5)
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)

	computedRating := bsr.ComputeRating(pk, validatorDecreaseRatingStep, initialRatingValue)

	expectedValue := uint32(int32(initialRatingValue) + rd.RatingOptions()[validatorDecreaseRatingStep])

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenProposerButNotAccepted(t *testing.T) {
	pk := "test"
	initialRatingValue := uint32(5)
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)

	computedRating := bsr.ComputeRating(pk, proposerDecreaseRatingStepKey, initialRatingValue)

	expectedValue := uint32(int32(initialRatingValue) + rd.RatingOptions()[proposerDecreaseRatingStepKey])

	assert.Equal(t, expectedValue, computedRating)
}
