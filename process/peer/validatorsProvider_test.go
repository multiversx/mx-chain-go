package peer

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewValidatorsProvider_WithNilValidatorStatisticsShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.ValidatorStatistics = nil
	vp, err := NewValidatorsProvider(arg)
	assert.Equal(t, process.ErrNilValidatorStatistics, err)
	assert.Nil(t, vp)
}

func TestNewValidatorsProvider_WithMaxRatingZeroShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.MaxRating = uint32(0)
	vp, err := NewValidatorsProvider(arg)
	assert.Equal(t, process.ErrMaxRatingZero, err)
	assert.Nil(t, vp)
}

func TestNewValidatorsProvider_WithNilValidatorPubkeyConverterShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.PubKeyConverter = nil
	vp, err := NewValidatorsProvider(arg)

	assert.Equal(t, process.ErrNilPubkeyConverter, err)
	assert.True(t, check.IfNil(vp))
}

func TestNewValidatorsProvider_WithNilNodesCoordinatorrShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.NodesCoordinator = nil
	vp, err := NewValidatorsProvider(arg)

	assert.Equal(t, process.ErrNilNodesCoordinator, err)
	assert.True(t, check.IfNil(vp))
}

func TestNewValidatorsProvider_WithNilStartOfEpochTriggerShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.EpochStartEventNotifier = nil
	vp, err := NewValidatorsProvider(arg)

	assert.Equal(t, process.ErrNilEpochStartNotifier, err)
	assert.True(t, check.IfNil(vp))
}

func TestNewValidatorsProvider_WithNilRefresCacheIntervalInSecShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.CacheRefreshIntervalDurationInSec = 0
	vp, err := NewValidatorsProvider(arg)

	assert.Equal(t, process.ErrInvalidCacheRefreshIntervalInSec, err)
	assert.True(t, check.IfNil(vp))
}

func TestValidatorsProvider_GetLatestValidatorsSecondHashDoesNotExist(t *testing.T) {
	root := []byte("rootHash")
	e := errors.Errorf("not ok")
	initialInfo := createMockValidatorInfo()

	validatorInfos := map[uint32][]*state.ValidatorInfo{
		0: {initialInfo},
	}

	gotOk := false
	gotNil := false
	vs := &mock.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() (bytes []byte) {
			return root
		},
		GetValidatorInfoForRootHashCalled: func(rootHash []byte) (m map[uint32][]*state.ValidatorInfo, err error) {
			if bytes.Equal([]byte("rootHash"), rootHash) {
				gotOk = true
				return validatorInfos, nil
			}
			gotNil = true
			return nil, e
		},
	}

	maxRating := uint32(100)
	args := createDefaultValidatorsProviderArg()
	args.ValidatorStatistics = vs
	args.MaxRating = maxRating
	vp, _ := NewValidatorsProvider(args)
	time.Sleep(time.Millisecond)
	vinfos := vp.GetLatestValidators()
	assert.NotNil(t, vinfos)
	assert.Equal(t, 1, len(vinfos))
	time.Sleep(time.Millisecond)
	root = []byte("otherHash")
	vinfos2 := vp.GetLatestValidators()
	time.Sleep(time.Millisecond)
	assert.NotNil(t, vinfos2)
	assert.Equal(t, 1, len(vinfos2))

	assert.True(t, gotOk)
	assert.True(t, gotNil)

	assert.True(t, reflect.DeepEqual(vinfos, vinfos2))
	validatorInfoApi := vinfos[hex.EncodeToString(initialInfo.GetPublicKey())]
	assert.Equal(t, initialInfo.GetTempRating(), uint32(validatorInfoApi.GetTempRating()))
	assert.Equal(t, initialInfo.GetRating(), uint32(validatorInfoApi.GetRating()))
	assert.Equal(t, initialInfo.GetLeaderSuccess(), validatorInfoApi.GetNumLeaderSuccess())
	assert.Equal(t, initialInfo.GetLeaderFailure(), validatorInfoApi.GetNumLeaderFailure())
	assert.Equal(t, initialInfo.GetValidatorSuccess(), validatorInfoApi.GetNumValidatorSuccess())
	assert.Equal(t, initialInfo.GetValidatorFailure(), validatorInfoApi.GetNumValidatorFailure())
	assert.Equal(t, initialInfo.GetTotalLeaderSuccess(), validatorInfoApi.GetTotalNumLeaderSuccess())
	assert.Equal(t, initialInfo.GetTotalLeaderFailure(), validatorInfoApi.GetTotalNumLeaderFailure())
	assert.Equal(t, initialInfo.GetTotalValidatorSuccess(), validatorInfoApi.GetTotalNumValidatorSuccess())
	assert.Equal(t, initialInfo.GetTotalValidatorFailure(), validatorInfoApi.GetTotalNumValidatorFailure())
}

func TestValidatorsProvider_ShouldWork(t *testing.T) {
	args := createDefaultValidatorsProviderArg()
	vp, err := NewValidatorsProvider(args)

	assert.Nil(t, err)
	assert.NotNil(t, vp)
}

func createMockValidatorInfo() *state.ValidatorInfo {
	initialInfo := &state.ValidatorInfo{
		PublicKey:                  []byte("a1"),
		ShardId:                    0,
		List:                       "eligible",
		Index:                      1,
		TempRating:                 100,
		Rating:                     1000,
		RewardAddress:              []byte("rewardA1"),
		LeaderSuccess:              1,
		LeaderFailure:              2,
		ValidatorSuccess:           3,
		ValidatorFailure:           4,
		TotalLeaderSuccess:         10,
		TotalLeaderFailure:         20,
		TotalValidatorSuccess:      30,
		TotalValidatorFailure:      40,
		NumSelectedInSuccessBlocks: 5,
		AccumulatedFees:            big.NewInt(100),
	}
	return initialInfo
}

func createDefaultValidatorsProviderArg() ArgValidatorsProvider {
	return ArgValidatorsProvider{
		NodesCoordinator:                  &mock.NodesCoordinatorMock{},
		StartEpoch:                        1,
		EpochStartEventNotifier:           &mock.EpochStartNotifierStub{},
		CacheRefreshIntervalDurationInSec: 1 * time.Millisecond,
		ValidatorStatistics:               &mock.ValidatorStatisticsProcessorStub{},
		MaxRating:                         100,
		PubKeyConverter:                   mock.NewPubkeyConverterMock(32),
	}
}
