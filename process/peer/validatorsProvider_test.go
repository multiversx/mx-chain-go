package peer

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"reflect"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewValidatorsProvider_WithNilValidatorStatisticsShouldErr(t *testing.T) {
	maxRating := uint32(10)
	vp, err := NewValidatorsProvider(
		nil,
		maxRating,
		mock.NewPubkeyConverterMock(32),
	)
	assert.Equal(t, process.ErrNilValidatorStatistics, err)
	assert.Nil(t, vp)
}

func TestNewValidatorsProvider_WithMaxRatingZeroShouldErr(t *testing.T) {
	maxRating := uint32(0)
	vp, err := NewValidatorsProvider(
		&mock.ValidatorStatisticsProcessorStub{},
		maxRating,
		mock.NewPubkeyConverterMock(32),
	)
	assert.Equal(t, process.ErrMaxRatingZero, err)
	assert.Nil(t, vp)
}

func TestNewValidatorsProvider_WithNilValidatorPubkeyConverterShouldErr(t *testing.T) {
	maxRating := uint32(10)
	vp, err := NewValidatorsProvider(
		&mock.ValidatorStatisticsProcessorStub{},
		maxRating,
		nil,
	)

	assert.Equal(t, process.ErrNilPubkeyConverter, err)
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
		RootHashCalled: func() (bytes []byte, err error) {
			return root, nil
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
	vp, _ := NewValidatorsProvider(
		vs,
		maxRating,
		mock.NewPubkeyConverterMock(32),
	)
	vinfos := vp.GetLatestValidators()
	assert.NotNil(t, vinfos)
	assert.Equal(t, 1, len(vinfos))

	root = []byte("otherHash")
	vinfos2 := vp.GetLatestValidators()
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

func TestValidatorsProvider_ShouldWork(t *testing.T) {
	vs := &mock.ValidatorStatisticsProcessorStub{}
	maxRating := uint32(7)
	vp, err := NewValidatorsProvider(
		vs,
		maxRating,
		mock.NewPubkeyConverterMock(32),
	)

	assert.Nil(t, err)
	assert.NotNil(t, vp)
}
