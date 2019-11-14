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
	ratingValues := make(map[string]int64, 0)
	ratingValues[validatorIncreaseRatingStep] = 1
	ratingValues[validatorDecreaseRatingStep] = 2
	ratingValues[proposerIncreaseRatingStepKey] = 3
	ratingValues[proposerDecreaseRatingStepKey] = 4

	ratingsData, _ := economics.NewRatingsData(int64(5), int64(1), int64(10), "mockRater", ratingValues)
	return ratingsData
}

func createUpdateMap(ratingPk string, updateStep string) map[string][]string {
	updatedPeers := make(map[string][]string, 0)
	peerList := make([]string, 0)
	peerList = append(peerList, ratingPk)
	updatedPeers[updateStep] = peerList
	return updatedPeers
}

func setupRater(rd *economics.RatingsData, pk string, initialRating int64) *rating.BlockSigningRater {
	bsr, _ := rating.NewBlockSigningRater(rd)
	ratingPk := pk
	ratingsMap := make(map[string]int64)
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
	ratingValue := int64(5)

	ratingsMap := make(map[string]int64)
	ratingsMap[ratingPk] = ratingValue
	bsr.SetRatings(ratingsMap)
	rt := bsr.GetRating(ratingPk)

	assert.Equal(t, ratingValue, rt)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenProposed(t *testing.T) {
	pk := "test"
	initialRatingValue := int64(5)
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)
	updatedPeers := createUpdateMap(pk, proposerIncreaseRatingStepKey)
	bsr.UpdateRatings(updatedPeers)

	rt := bsr.GetRating(pk)

	expectedValue := initialRatingValue + rd.RatingOptions()[proposerIncreaseRatingStepKey]

	assert.Equal(t, expectedValue, rt)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenValidator(t *testing.T) {
	pk := "test"
	initialRatingValue := int64(5)
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)
	updatedPeers := createUpdateMap(pk, validatorIncreaseRatingStep)
	bsr.UpdateRatings(updatedPeers)
	rt := bsr.GetRating(pk)

	expectedValue := initialRatingValue + rd.RatingOptions()[validatorIncreaseRatingStep]

	assert.Equal(t, expectedValue, rt)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenValidatorButNotAccepted(t *testing.T) {
	pk := "test"
	initialRatingValue := int64(5)
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)
	updatedPeers := createUpdateMap(pk, validatorDecreaseRatingStep)
	bsr.UpdateRatings(updatedPeers)
	rt := bsr.GetRating(pk)

	expectedValue := initialRatingValue + rd.RatingOptions()[validatorDecreaseRatingStep]

	assert.Equal(t, expectedValue, rt)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenProposerButNotAccepted(t *testing.T) {
	pk := "test"
	initialRatingValue := int64(5)
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)
	updatedPeers := createUpdateMap(pk, proposerDecreaseRatingStepKey)
	bsr.UpdateRatings(updatedPeers)
	rt := bsr.GetRating(pk)

	expectedValue := initialRatingValue + rd.RatingOptions()[proposerDecreaseRatingStepKey]

	assert.Equal(t, expectedValue, rt)
}
