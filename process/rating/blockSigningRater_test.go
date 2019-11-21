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
	minRating                     = uint32(1)
	maxRating                     = uint32(10)
)

func createDefaultRatingsData() *economics.RatingsData {
	ratingValues := make(map[string]int32, 0)
	ratingValues[validatorIncreaseRatingStep] = 1
	ratingValues[validatorDecreaseRatingStep] = -2
	ratingValues[proposerIncreaseRatingStepKey] = 3
	ratingValues[proposerDecreaseRatingStepKey] = -4

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

func setupRater(rd *economics.RatingsData, pk string, initialRating uint32) *rating.BlockSigningRater {
	bsr, _ := rating.NewBlockSigningRater(rd)
	ratingPk := pk
	ratingsMap := make(map[string]uint32)
	ratingsMap[ratingPk] = initialRating
	rrm := createDefaultRatingReader(ratingsMap)
	bsr.SetRatingReader(rrm)

	return bsr
}

func TestBlockSigningRater_GetRatingWithNotSetRatingReaderShouldReturnZero(t *testing.T) {
	rd := createDefaultRatingsData()
	bsr, _ := rating.NewBlockSigningRater(rd)

	rt := bsr.GetRating("test")

	assert.Equal(t, uint32(0), rt)
}

func TestBlockSigningRater_GetRatingWithUnknownPkShoudReturnError(t *testing.T) {
	rd := createDefaultRatingsData()
	bsr, _ := rating.NewBlockSigningRater(rd)

	rrm := createDefaultRatingReader(make(map[string]uint32))
	bsr.SetRatingReader(rrm)

	rt := bsr.GetRating("test")

	assert.Equal(t, uint32(0), rt)
}

func TestBlockSigningRater_GetRatingsWithAllKnownPeersShouldReturnRatings(t *testing.T) {
	rd := createDefaultRatingsData()
	bsr, _ := rating.NewBlockSigningRater(rd)

	pk1 := "pk1"
	pk2 := "pk2"

	pk1Rating := uint32(4)
	pk2Rating := uint32(6)

	ratingsMap := make(map[string]uint32)
	ratingsMap[pk1] = pk1Rating
	ratingsMap[pk2] = pk2Rating

	rrm := createDefaultRatingReader(ratingsMap)
	bsr.SetRatingReader(rrm)

	rt := bsr.GetRatings([]string{pk2, pk1})

	for pk, val := range rt {
		assert.Equal(t, ratingsMap[pk], val)
	}
}

func TestBlockSigningRater_GetRatingsWithNotAllKnownPeersShouldReturnRatings(t *testing.T) {
	rd := createDefaultRatingsData()
	bsr, _ := rating.NewBlockSigningRater(rd)

	pk1 := "pk1"
	pk2 := "pk2"
	pk3 := "pk3"

	pk1Rating := uint32(4)
	pk2Rating := uint32(6)

	ratingsMap := make(map[string]uint32)
	ratingsMap[pk1] = pk1Rating
	ratingsMap[pk2] = pk2Rating

	rrm := createDefaultRatingReader(ratingsMap)
	bsr.SetRatingReader(rrm)

	rt := bsr.GetRatings([]string{pk2, pk3, pk1})

	for pk, val := range rt {
		assert.Equal(t, ratingsMap[pk], val)
	}
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

func TestBlockSigningRater_UpdateRatingsShouldNotIncreaseAboveMaxRating(t *testing.T) {
	pk := "test"
	initialRatingValue := maxRating - 1
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)
	computedRating := bsr.ComputeRating(pk, proposerIncreaseRatingStepKey, initialRatingValue)

	expectedValue := maxRating

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldNotDecreaseBelowMinRating(t *testing.T) {
	pk := "test"
	initialRatingValue := minRating + 1
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)
	computedRating := bsr.ComputeRating(pk, proposerDecreaseRatingStepKey, initialRatingValue)

	expectedValue := minRating

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_GetAllRatingKeysShouldReturnAllRatingKeys(t *testing.T) {
	pk := "test"
	initialRatingValue := minRating + 1
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)

	optionKeys := bsr.GetRatingOptionKeys()

	expectedLength := len(rd.RatingOptions())

	assert.Equal(t, expectedLength, len(optionKeys))
	for _, value := range optionKeys {
		ratingOption := rd.RatingOptions()[value]
		assert.NotNil(t, ratingOption)
	}
}

func TestBlockSigningRater_UpdateRatingsWithMultiplePeersShouldReturnRatings(t *testing.T) {
	rd := createDefaultRatingsData()
	bsr, _ := rating.NewBlockSigningRater(rd)

	pk1 := "pk1"
	pk2 := "pk2"
	pk3 := "pk3"
	pk4 := "pk4"

	pk1Rating := uint32(4)
	pk2Rating := uint32(5)
	pk3Rating := uint32(6)
	pk4Rating := uint32(7)

	ratingsMap := make(map[string]uint32)
	ratingsMap[pk1] = pk1Rating
	ratingsMap[pk2] = pk2Rating
	ratingsMap[pk3] = pk3Rating
	ratingsMap[pk4] = pk4Rating

	rrm := createDefaultRatingReader(ratingsMap)
	bsr.SetRatingReader(rrm)

	pk1ComputedRating := bsr.ComputeRating(pk1, proposerIncreaseRatingStepKey, ratingsMap[pk1])
	pk2ComputedRating := bsr.ComputeRating(pk2, proposerDecreaseRatingStepKey, ratingsMap[pk2])
	pk3ComputedRating := bsr.ComputeRating(pk3, validatorIncreaseRatingStep, ratingsMap[pk3])
	pk4ComputedRating := bsr.ComputeRating(pk4, validatorDecreaseRatingStep, ratingsMap[pk4])

	expectedPk1 := uint32(int32(ratingsMap[pk1]) + rd.RatingOptions()[proposerIncreaseRatingStepKey])
	expectedPk2 := uint32(int32(ratingsMap[pk2]) + rd.RatingOptions()[proposerDecreaseRatingStepKey])
	expectedPk3 := uint32(int32(ratingsMap[pk3]) + rd.RatingOptions()[validatorIncreaseRatingStep])
	expectedPk4 := uint32(int32(ratingsMap[pk4]) + rd.RatingOptions()[validatorDecreaseRatingStep])

	assert.Equal(t, expectedPk1, pk1ComputedRating)
	assert.Equal(t, expectedPk2, pk2ComputedRating)
	assert.Equal(t, expectedPk3, pk3ComputedRating)
	assert.Equal(t, expectedPk4, pk4ComputedRating)
}
