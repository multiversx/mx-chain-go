package rating_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/economics"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/stretchr/testify/assert"
)

const (
	validatorIncreaseRatingStep = uint32(1)
	validatorDecreaseRatingStep = uint32(2)
	proposerIncreaseRatingStep  = uint32(2)
	proposerDecreaseRatingStep  = uint32(4)
	minRating                   = uint32(1)
	maxRating                   = uint32(100)
	startRating                 = uint32(50)
)

func createDefaultChances() []config.SelectionChance {
	chances := make([]config.SelectionChance, 0)
	chances = append(chances, config.SelectionChance{
		MaxThreshold:  10,
		ChancePercent: 0,
	})
	chances = append(chances, config.SelectionChance{
		MaxThreshold:  25,
		ChancePercent: 90,
	})
	chances = append(chances, config.SelectionChance{
		MaxThreshold:  75,
		ChancePercent: 100,
	})
	chances = append(chances, config.SelectionChance{
		MaxThreshold:  100,
		ChancePercent: 110,
	})

	return chances
}

func createDefaultRatingsData() *economics.RatingsData {
	data := config.RatingSettings{
		StartRating:                 startRating,
		MaxRating:                   maxRating,
		MinRating:                   minRating,
		ProposerIncreaseRatingStep:  proposerIncreaseRatingStep,
		ProposerDecreaseRatingStep:  proposerDecreaseRatingStep,
		ValidatorIncreaseRatingStep: validatorIncreaseRatingStep,
		ValidatorDecreaseRatingStep: validatorDecreaseRatingStep,
		SelectionChance:             createDefaultChances(),
	}

	ratingsData, _ := economics.NewRatingsData(data)
	return ratingsData
}

func createDefaultRatingReader(ratingsMap map[string]uint32) *mock.RatingReaderMock {
	rrm := &mock.RatingReaderMock{
		RatingsMap: ratingsMap,
		GetRatingCalled: func(s string) uint32 {
			value, ok := ratingsMap[s]
			if !ok {
				return startRating
			}
			return value
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

func TestBlockSigningRater_GetRatingWithNotSetRatingReaderShouldReturnStartRating(t *testing.T) {
	rd := createDefaultRatingsData()

	bsr, _ := rating.NewBlockSigningRater(rd)
	rrm := createDefaultRatingReader(make(map[string]uint32))
	bsr.SetRatingReader(rrm)

	rt := bsr.GetRating("test")

	assert.Equal(t, rd.StartRating(), rt)
}

func TestBlockSigningRater_GetRatingWithUnknownPkShoudReturnStartRating(t *testing.T) {
	rd := createDefaultRatingsData()
	bsr, _ := rating.NewBlockSigningRater(rd)

	rrm := createDefaultRatingReader(make(map[string]uint32))
	bsr.SetRatingReader(rrm)

	rt := bsr.GetRating("test")

	assert.Equal(t, startRating, rt)
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
	computedRating := bsr.ComputeIncreaseProposer(initialRatingValue)

	expectedValue := initialRatingValue + proposerIncreaseRatingStep

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenValidator(t *testing.T) {
	pk := "test"
	initialRatingValue := uint32(5)
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)

	computedRating := bsr.ComputeIncreaseValidator(initialRatingValue)

	expectedValue := initialRatingValue + validatorIncreaseRatingStep

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenValidatorButNotAccepted(t *testing.T) {
	pk := "test"
	initialRatingValue := uint32(5)
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)

	computedRating := bsr.ComputeDecreaseValidator(initialRatingValue)

	expectedValue := initialRatingValue - validatorDecreaseRatingStep

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenProposerButNotAccepted(t *testing.T) {
	pk := "test"
	initialRatingValue := uint32(5)
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)

	computedRating := bsr.ComputeDecreaseProposer(initialRatingValue)

	expectedValue := initialRatingValue - proposerDecreaseRatingStep

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldNotIncreaseAboveMaxRating(t *testing.T) {
	pk := "test"
	initialRatingValue := maxRating - 1
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)
	computedRating := bsr.ComputeIncreaseProposer(initialRatingValue)

	expectedValue := maxRating

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldNotDecreaseBelowMinRating(t *testing.T) {
	pk := "test"
	initialRatingValue := minRating + 1
	rd := createDefaultRatingsData()

	bsr := setupRater(rd, pk, initialRatingValue)
	computedRating := bsr.ComputeDecreaseProposer(initialRatingValue)

	expectedValue := minRating

	assert.Equal(t, expectedValue, computedRating)
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

	pk1ComputedRating := bsr.ComputeIncreaseProposer(ratingsMap[pk1])
	pk2ComputedRating := bsr.ComputeDecreaseProposer(ratingsMap[pk2])
	pk3ComputedRating := bsr.ComputeIncreaseValidator(ratingsMap[pk3])
	pk4ComputedRating := bsr.ComputeDecreaseValidator(ratingsMap[pk4])

	expectedPk1 := ratingsMap[pk1] + proposerIncreaseRatingStep
	expectedPk2 := ratingsMap[pk2] - proposerDecreaseRatingStep
	expectedPk3 := ratingsMap[pk3] + validatorIncreaseRatingStep
	expectedPk4 := ratingsMap[pk4] - validatorDecreaseRatingStep

	assert.Equal(t, expectedPk1, pk1ComputedRating)
	assert.Equal(t, expectedPk2, pk2ComputedRating)
	assert.Equal(t, expectedPk3, pk3ComputedRating)
	assert.Equal(t, expectedPk4, pk4ComputedRating)
}

func TestBlockSigningRater_NewBlockSigningRaterWithChancesNilShouldErr(t *testing.T) {
	data := config.RatingSettings{
		StartRating:                 startRating,
		MaxRating:                   maxRating,
		MinRating:                   minRating,
		ProposerIncreaseRatingStep:  proposerIncreaseRatingStep,
		ProposerDecreaseRatingStep:  proposerDecreaseRatingStep,
		ValidatorIncreaseRatingStep: validatorIncreaseRatingStep,
		ValidatorDecreaseRatingStep: validatorDecreaseRatingStep,
		SelectionChance:             nil,
	}

	ratingsData, _ := economics.NewRatingsData(data)

	bsr, err := rating.NewBlockSigningRater(ratingsData)

	assert.Nil(t, bsr)
	assert.Equal(t, process.ErrNoChancesProvided, err)
}

func TestBlockSigningRater_NewBlockSigningRaterWithDupplicateMaxThresholdShouldErr(t *testing.T) {
	chances := make([]config.SelectionChance, 0)
	chances = append(chances, config.SelectionChance{
		MaxThreshold:  10,
		ChancePercent: 0,
	})
	chances = append(chances, config.SelectionChance{
		MaxThreshold:  20,
		ChancePercent: 90,
	})
	chances = append(chances, config.SelectionChance{
		MaxThreshold:  20,
		ChancePercent: 100,
	})
	chances = append(chances, config.SelectionChance{
		MaxThreshold:  100,
		ChancePercent: 110,
	})

	data := config.RatingSettings{
		StartRating:                 startRating,
		MaxRating:                   maxRating,
		MinRating:                   minRating,
		ProposerIncreaseRatingStep:  proposerIncreaseRatingStep,
		ProposerDecreaseRatingStep:  proposerDecreaseRatingStep,
		ValidatorIncreaseRatingStep: validatorIncreaseRatingStep,
		ValidatorDecreaseRatingStep: validatorDecreaseRatingStep,
		SelectionChance:             chances,
	}

	ratingsData, _ := economics.NewRatingsData(data)

	bsr, err := rating.NewBlockSigningRater(ratingsData)

	assert.Nil(t, bsr)
	assert.Equal(t, process.ErrDupplicateThreshold, err)
}

func TestBlockSigningRater_NewBlockSigningRaterWithNoValueForMaxThresholdShouldErr(t *testing.T) {
	chances := make([]config.SelectionChance, 0)
	chances = append(chances, config.SelectionChance{
		MaxThreshold:  10,
		ChancePercent: 0,
	})
	chances = append(chances, config.SelectionChance{
		MaxThreshold:  20,
		ChancePercent: 90,
	})

	data := config.RatingSettings{
		StartRating:                 startRating,
		MaxRating:                   maxRating,
		MinRating:                   minRating,
		ProposerIncreaseRatingStep:  proposerIncreaseRatingStep,
		ProposerDecreaseRatingStep:  proposerDecreaseRatingStep,
		ValidatorIncreaseRatingStep: validatorIncreaseRatingStep,
		ValidatorDecreaseRatingStep: validatorDecreaseRatingStep,
		SelectionChance:             chances,
	}

	ratingsData, _ := economics.NewRatingsData(data)

	bsr, err := rating.NewBlockSigningRater(ratingsData)

	assert.Nil(t, bsr)
	assert.Equal(t, process.ErrNoChancesForMaxThreshold, err)
}

func TestBlockSigningRater_NewBlockSigningRaterWithCorrectValueShouldWork(t *testing.T) {
	ratingsData := createDefaultRatingsData()

	bsr, err := rating.NewBlockSigningRater(ratingsData)

	assert.NotNil(t, bsr)
	assert.Nil(t, err)
}

func TestBlockSigningRater_GetChancesForStartRatingdReturnStartRatingChance(t *testing.T) {
	ratingsData := createDefaultRatingsData()

	bsr, _ := rating.NewBlockSigningRater(ratingsData)

	chance := bsr.GetChance(startRating)

	chanceForStartRating := uint32(100)
	assert.Equal(t, chanceForStartRating, chance)
}

func TestBlockSigningRater_GetChancesForSetRatingShouldReturnCorrectRating(t *testing.T) {
	rd := createDefaultRatingsData()

	bsr, _ := rating.NewBlockSigningRater(rd)

	ratingValue := uint32(80)
	chances := bsr.GetChance(ratingValue)

	chancesFor80 := uint32(110)
	assert.Equal(t, chancesFor80, chances)
}
