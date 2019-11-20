package rating_test

import (
	"github.com/ElrondNetwork/elrond-go/process/economics"
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
	bsr.SetRatings(ratingsMap)

	return bsr
}

func TestBlockSigningRater_GetRatingWithUnknownPkShoudReturnDefaultRating(t *testing.T) {
	rd := createDefaultRatingsData()
	bsr, _ := rating.NewBlockSigningRater(rd)

	rt := bsr.GetRating("test")

	assert.Equal(t, rd.StartRating(), rt)
}

func TestBlockSigningRater_GetRatingWithKnownPkShoudReturnSetRating(t *testing.T) {
	rd := createDefaultRatingsData()
	bsr, _ := rating.NewBlockSigningRater(rd)

	ratingPk := "test"
	ratingValue := uint32(5)

	ratingsMap := make(map[string]uint32)
	ratingsMap[ratingPk] = ratingValue
	bsr.SetRatings(ratingsMap)
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
