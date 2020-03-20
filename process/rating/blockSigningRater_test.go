package rating_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/rating"
	"github.com/stretchr/testify/assert"
)

const (
	validatorIncreaseRatingStep     = int32(1)
	validatorDecreaseRatingStep     = int32(-2)
	proposerIncreaseRatingStep      = int32(2)
	proposerDecreaseRatingStep      = int32(-4)
	metaValidatorIncreaseRatingStep = int32(3)
	metaValidatorDecreaseRatingStep = int32(-4)
	metaProposerIncreaseRatingStep  = int32(5)
	metaProposerDecreaseRatingStep  = int32(-10)
	minRating                       = uint32(1)
	maxRating                       = uint32(100)
	startRating                     = uint32(50)
)

func createDefaultChances() []process.SelectionChance {
	chances := []process.SelectionChance{
		&rating.SelectionChance{MaxThreshold: 0, ChancePercent: 50},
		&rating.SelectionChance{MaxThreshold: 10, ChancePercent: 0},
		&rating.SelectionChance{MaxThreshold: 25, ChancePercent: 90},
		&rating.SelectionChance{MaxThreshold: 75, ChancePercent: 100},
		&rating.SelectionChance{MaxThreshold: 100, ChancePercent: 110},
	}

	return chances
}

func createDefaultRatingsData() *mock.RatingsInfoMock {
	ratingsData := &mock.RatingsInfoMock{
		StartRatingProperty: startRating,
		MaxRatingProperty:   maxRating,
		MinRatingProperty:   minRating,
		MetaRatingsStepDataProperty: &mock.RatingStepMock{
			ProposerIncreaseRatingStepProperty:  metaProposerIncreaseRatingStep,
			ProposerDecreaseRatingStepProperty:  metaProposerDecreaseRatingStep,
			ValidatorIncreaseRatingStepProperty: metaValidatorIncreaseRatingStep,
			ValidatorDecreaseRatingStepProperty: metaValidatorDecreaseRatingStep,
		},
		ShardRatingsStepDataProperty: &mock.RatingStepMock{
			ProposerIncreaseRatingStepProperty:  proposerIncreaseRatingStep,
			ProposerDecreaseRatingStepProperty:  proposerDecreaseRatingStep,
			ValidatorIncreaseRatingStepProperty: validatorIncreaseRatingStep,
			ValidatorDecreaseRatingStepProperty: validatorDecreaseRatingStep,
		},
		SelectionChancesProperty: createDefaultChances(),
	}

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

func setupRater(rd process.RatingsInfoHandler, pk string, initialRating uint32) *rating.BlockSigningRater {
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
	shardId := uint32(0)

	bsr := setupRater(rd, pk, initialRatingValue)
	computedRating := bsr.ComputeIncreaseProposer(shardId, initialRatingValue)

	expectedValue := uint32(int32(initialRatingValue) + proposerIncreaseRatingStep)

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenValidator(t *testing.T) {
	pk := "test"
	initialRatingValue := uint32(5)
	rd := createDefaultRatingsData()
	shardId := uint32(0)

	bsr := setupRater(rd, pk, initialRatingValue)
	computedRating := bsr.ComputeIncreaseValidator(shardId, initialRatingValue)

	expectedValue := uint32(int32(initialRatingValue) + validatorIncreaseRatingStep)

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenValidatorButNotAccepted(t *testing.T) {
	pk := "test"
	initialRatingValue := uint32(5)
	rd := createDefaultRatingsData()
	shardId := uint32(0)

	bsr := setupRater(rd, pk, initialRatingValue)
	computedRating := bsr.ComputeDecreaseValidator(shardId, initialRatingValue)

	expectedValue := uint32(int32(initialRatingValue) + validatorDecreaseRatingStep)

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldUpdateRatingWhenProposerButNotAccepted(t *testing.T) {
	pk := "test"
	initialRatingValue := uint32(5)
	rd := createDefaultRatingsData()
	shardId := uint32(0)

	bsr := setupRater(rd, pk, initialRatingValue)
	computedRating := bsr.ComputeDecreaseProposer(shardId, initialRatingValue)

	expectedValue := uint32(int32(initialRatingValue) + proposerDecreaseRatingStep)

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldNotIncreaseAboveMaxRating(t *testing.T) {
	pk := "test"
	initialRatingValue := maxRating - 1
	rd := createDefaultRatingsData()
	shardId := uint32(0)

	bsr := setupRater(rd, pk, initialRatingValue)
	computedRating := bsr.ComputeIncreaseProposer(shardId, initialRatingValue)

	expectedValue := maxRating

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsShouldNotDecreaseBelowMinRating(t *testing.T) {
	pk := "test"
	initialRatingValue := minRating + 1
	rd := createDefaultRatingsData()
	shardId := uint32(0)

	bsr := setupRater(rd, pk, initialRatingValue)
	computedRating := bsr.ComputeDecreaseProposer(shardId, initialRatingValue)

	expectedValue := minRating

	assert.Equal(t, expectedValue, computedRating)
}

func TestBlockSigningRater_UpdateRatingsWithMultiplePeersShouldReturnRatings(t *testing.T) {
	rd := createDefaultRatingsData()
	bsr, _ := rating.NewBlockSigningRater(rd)
	shardId := uint32(0)

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

	pk1ComputedRating := bsr.ComputeIncreaseProposer(shardId, ratingsMap[pk1])
	pk2ComputedRating := bsr.ComputeDecreaseProposer(shardId, ratingsMap[pk2])
	pk3ComputedRating := bsr.ComputeIncreaseValidator(shardId, ratingsMap[pk3])
	pk4ComputedRating := bsr.ComputeDecreaseValidator(shardId, ratingsMap[pk4])

	expectedPk1 := uint32(int32(ratingsMap[pk1]) + proposerIncreaseRatingStep)
	expectedPk2 := uint32(int32(ratingsMap[pk2]) + proposerDecreaseRatingStep)
	expectedPk3 := uint32(int32(ratingsMap[pk3]) + validatorIncreaseRatingStep)
	expectedPk4 := uint32(int32(ratingsMap[pk4]) + validatorDecreaseRatingStep)

	assert.Equal(t, expectedPk1, pk1ComputedRating)
	assert.Equal(t, expectedPk2, pk2ComputedRating)
	assert.Equal(t, expectedPk3, pk3ComputedRating)
	assert.Equal(t, expectedPk4, pk4ComputedRating)
}

func TestBlockSigningRater_UpdateRatingsOnMetaWithMultiplePeersShouldReturnRatings(t *testing.T) {
	rd := createDefaultRatingsData()
	bsr, _ := rating.NewBlockSigningRater(rd)

	pk1 := "pk1"
	pk2 := "pk2"
	pk3 := "pk3"
	pk4 := "pk4"

	pk1Rating := uint32(14)
	pk2Rating := uint32(15)
	pk3Rating := uint32(16)
	pk4Rating := uint32(17)

	ratingsMap := make(map[string]uint32)
	ratingsMap[pk1] = pk1Rating
	ratingsMap[pk2] = pk2Rating
	ratingsMap[pk3] = pk3Rating
	ratingsMap[pk4] = pk4Rating

	rrm := createDefaultRatingReader(ratingsMap)
	bsr.SetRatingReader(rrm)

	pk1ComputedRating := bsr.ComputeIncreaseProposer(core.MetachainShardId, ratingsMap[pk1])
	pk2ComputedRating := bsr.ComputeDecreaseProposer(core.MetachainShardId, ratingsMap[pk2])
	pk3ComputedRating := bsr.ComputeIncreaseValidator(core.MetachainShardId, ratingsMap[pk3])
	pk4ComputedRating := bsr.ComputeDecreaseValidator(core.MetachainShardId, ratingsMap[pk4])

	expectedPk1 := uint32(int32(ratingsMap[pk1]) + metaProposerIncreaseRatingStep)
	expectedPk2 := uint32(int32(ratingsMap[pk2]) + metaProposerDecreaseRatingStep)
	expectedPk3 := uint32(int32(ratingsMap[pk3]) + metaValidatorIncreaseRatingStep)
	expectedPk4 := uint32(int32(ratingsMap[pk4]) + metaValidatorDecreaseRatingStep)

	assert.Equal(t, expectedPk1, pk1ComputedRating)
	assert.Equal(t, expectedPk2, pk2ComputedRating)
	assert.Equal(t, expectedPk3, pk3ComputedRating)
	assert.Equal(t, expectedPk4, pk4ComputedRating)
}

func TestBlockSigningRater_NewBlockSigningRaterWithChancesNilShouldErr(t *testing.T) {
	ratingsData := createDefaultRatingsData()
	ratingsData.SelectionChancesProperty = nil

	bsr, err := rating.NewBlockSigningRater(ratingsData)

	assert.Nil(t, bsr)
	assert.Equal(t, process.ErrNoChancesProvided, err)
}

func TestBlockSigningRater_NewBlockSigningRaterWithDupplicateMaxThresholdShouldErr(t *testing.T) {
	chances := []process.SelectionChance{
		&rating.SelectionChance{MaxThreshold: 0, ChancePercent: 50},
		&rating.SelectionChance{MaxThreshold: 10, ChancePercent: 0},
		&rating.SelectionChance{MaxThreshold: 20, ChancePercent: 90},
		&rating.SelectionChance{MaxThreshold: 20, ChancePercent: 100},
		&rating.SelectionChance{MaxThreshold: 100, ChancePercent: 110},
	}
	ratingsData := createDefaultRatingsData()
	ratingsData.SelectionChancesProperty = chances

	bsr, err := rating.NewBlockSigningRater(ratingsData)

	assert.Nil(t, bsr)
	assert.Equal(t, process.ErrDuplicateThreshold, err)
}

func TestBlockSigningRater_NewBlockSigningRaterWithZeroMinRatingShouldErr(t *testing.T) {
	ratingsData := createDefaultRatingsData()
	ratingsData.MinRatingProperty = 0

	bsr, err := rating.NewBlockSigningRater(ratingsData)

	assert.Nil(t, bsr)
	assert.Equal(t, process.ErrMinRatingSmallerThanOne, err)
}

func TestBlockSigningRater_NewBlockSigningRaterWithNonExistingMaxThresholdZeroShouldErr(t *testing.T) {
	chances := []process.SelectionChance{
		&rating.SelectionChance{MaxThreshold: 10, ChancePercent: 0},
		&rating.SelectionChance{MaxThreshold: 20, ChancePercent: 90},
		&rating.SelectionChance{MaxThreshold: 20, ChancePercent: 100},
		&rating.SelectionChance{MaxThreshold: 100, ChancePercent: 110},
	}

	ratingsData := createDefaultRatingsData()
	ratingsData.SelectionChancesProperty = chances

	bsr, err := rating.NewBlockSigningRater(ratingsData)

	assert.Nil(t, bsr)
	assert.Equal(t, process.ErrNilMinChanceIfZero, err)
}

func TestBlockSigningRater_NewBlockSigningRaterWithNoValueForMaxThresholdShouldErr(t *testing.T) {
	chances := []process.SelectionChance{
		&rating.SelectionChance{MaxThreshold: 0, ChancePercent: 5},
		&rating.SelectionChance{MaxThreshold: 10, ChancePercent: 0},
		&rating.SelectionChance{MaxThreshold: 20, ChancePercent: 90},
	}
	ratingsData := createDefaultRatingsData()
	ratingsData.SelectionChancesProperty = chances

	bsr, err := rating.NewBlockSigningRater(ratingsData)

	assert.Nil(t, bsr)
	assert.Equal(t, process.ErrNoChancesForMaxThreshold, err)
}

func TestBlockSigningRater_NewBlockSigningRaterWithCorrectValueShouldWork(t *testing.T) {
	ratingsData := createDefaultRatingsData()

	bsr, err := rating.NewBlockSigningRater(ratingsData)

	assert.NotNil(t, bsr)
	assert.Nil(t, err)

	testValue := int32(50)
	assert.Equal(t, ratingsData.StartRating(), bsr.GetStartRating())
	assert.Equal(t, uint32(testValue+ratingsData.ShardChainRatingsStepHandler().ValidatorIncreaseRatingStep()), bsr.ComputeIncreaseValidator(0, uint32(testValue)))
	assert.Equal(t, uint32(testValue+ratingsData.ShardChainRatingsStepHandler().ValidatorDecreaseRatingStep()), bsr.ComputeDecreaseValidator(0, uint32(testValue)))
	assert.Equal(t, uint32(testValue+ratingsData.ShardChainRatingsStepHandler().ProposerIncreaseRatingStep()), bsr.ComputeIncreaseProposer(0, uint32(testValue)))
	assert.Equal(t, uint32(testValue+ratingsData.ShardChainRatingsStepHandler().ProposerDecreaseRatingStep()), bsr.ComputeDecreaseProposer(0, uint32(testValue)))

	assert.Equal(t, ratingsData.StartRating(), bsr.GetStartRating())
	assert.Equal(t, uint32(testValue+ratingsData.MetaChainRatingsStepHandler().ValidatorIncreaseRatingStep()), bsr.ComputeIncreaseValidator(core.MetachainShardId, uint32(testValue)))
	assert.Equal(t, uint32(testValue+ratingsData.MetaChainRatingsStepHandler().ValidatorDecreaseRatingStep()), bsr.ComputeDecreaseValidator(core.MetachainShardId, uint32(testValue)))
	assert.Equal(t, uint32(testValue+ratingsData.MetaChainRatingsStepHandler().ProposerIncreaseRatingStep()), bsr.ComputeIncreaseProposer(core.MetachainShardId, uint32(testValue)))
	assert.Equal(t, uint32(testValue+ratingsData.MetaChainRatingsStepHandler().ProposerDecreaseRatingStep()), bsr.ComputeDecreaseProposer(core.MetachainShardId, uint32(testValue)))

	assert.Equal(t, ratingsData.SelectionChances()[0].GetChancePercent(), bsr.GetChance(uint32(0)))
	assert.Equal(t, ratingsData.SelectionChances()[1].GetChancePercent(), bsr.GetChance(uint32(9)))
	assert.Equal(t, ratingsData.SelectionChances()[2].GetChancePercent(), bsr.GetChance(uint32(20)))
	assert.Equal(t, ratingsData.SelectionChances()[3].GetChancePercent(), bsr.GetChance(uint32(50)))
	assert.Equal(t, ratingsData.SelectionChances()[4].GetChancePercent(), bsr.GetChance(uint32(100)))
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
