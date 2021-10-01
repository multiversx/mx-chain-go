package detector_test

import (
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	mockEpochStart "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/slash/detector"
	"github.com/ElrondNetwork/elrond-go/sharding"
	mock2 "github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewSigningSlashingDetector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args        func() (sharding.NodesCoordinator, process.RoundHandler, hashing.Hasher, marshal.Marshalizer, uint64)
		expectedErr error
	}{
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, hashing.Hasher, marshal.Marshalizer, uint64) {
				return nil, &mock.RoundHandlerMock{}, &mock.HasherMock{}, &mock.MarshalizerMock{}, detector.CacheSize
			},
			expectedErr: process.ErrNilShardCoordinator,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, hashing.Hasher, marshal.Marshalizer, uint64) {
				return &mock.NodesCoordinatorMock{}, nil, &mock.HasherMock{}, &mock.MarshalizerMock{}, detector.CacheSize
			},
			expectedErr: process.ErrNilRoundHandler,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, hashing.Hasher, marshal.Marshalizer, uint64) {
				return &mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, nil, &mock.MarshalizerMock{}, detector.CacheSize
			},
			expectedErr: process.ErrNilHasher,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, hashing.Hasher, marshal.Marshalizer, uint64) {
				return &mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, &mock.HasherMock{}, nil, detector.CacheSize
			},
			expectedErr: process.ErrNilMarshalizer,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, hashing.Hasher, marshal.Marshalizer, uint64) {
				return &mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, &mock.HasherMock{}, &mock.MarshalizerMock{}, detector.CacheSize
			},
			expectedErr: nil,
		},
	}

	for _, currTest := range tests {
		_, err := detector.NewSigningSlashingDetector(currTest.args())
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestMultipleHeaderSigningDetector_VerifyData_CannotCastData_ExpectError(t *testing.T) {
	t.Parallel()

	sd, _ := detector.NewSigningSlashingDetector(
		&mock.NodesCoordinatorMock{},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		detector.CacheSize)

	res, err := sd.VerifyData(&testscommon.InterceptedDataStub{})

	require.Nil(t, res)
	require.Equal(t, process.ErrCannotCastInterceptedDataToHeader, err)
}

func TestMultipleHeaderSigningDetector_VerifyData_DifferentRelevantAndIrrelevantRounds(t *testing.T) {
	t.Parallel()

	round := uint64(100)
	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{},
		&mock.RoundHandlerMock{
			RoundIndex: int64(round),
		},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		detector.CacheSize)

	hData := createInterceptedHeaderData(round+detector.MaxDeltaToCurrentRound+1, []byte("seed"))
	res, err := ssd.VerifyData(hData)

	require.Nil(t, res)
	require.Equal(t, process.ErrHeaderRoundNotRelevant, err)
}

func TestMultipleHeaderSigningDetector_DoubleSigners_EmptyValidatorLists_ExpectNoDoubleSigners(t *testing.T) {
	var group1 []sharding.Validator
	var group2 []sharding.Validator
	var bitmap1 []byte
	var bitmap2 []byte

	doubleSigners := detector.DoubleSigners(group1, group2, bitmap1, bitmap2)
	require.Len(t, doubleSigners, 0)

	validator := mock2.NewValidatorMock([]byte("pubKey1"), 0, 0)
	group1 = []sharding.Validator{validator}
	byte1, _ := strconv.ParseInt("00000001", 2, 8)
	bitmap1 = []byte{byte(byte1)}

	doubleSigners = detector.DoubleSigners(group1, group2, bitmap1, bitmap2)
	require.Len(t, doubleSigners, 0)

	group1 = []sharding.Validator{}
	group2 = []sharding.Validator{validator}
	bitmap2 = []byte{byte(byte1)}

	doubleSigners = detector.DoubleSigners(group1, group2, bitmap1, bitmap2)
	require.Len(t, doubleSigners, 0)
}

func TestMultipleHeaderSigningDetector_DoubleSigners_ExpectThreeDoubleSigners(t *testing.T) {
	v1 := mock2.NewValidatorMock([]byte("pubKey1"), 0, 0)
	v2 := mock2.NewValidatorMock([]byte("pubKey2"), 0, 1)
	v3 := mock2.NewValidatorMock([]byte("pubKey3"), 0, 4)
	v4 := mock2.NewValidatorMock([]byte("pubKey4"), 0, 5)
	v5 := mock2.NewValidatorMock([]byte("pubKey6"), 0, 8)
	group1 := []sharding.Validator{v1, v2, v3, v4, v5}

	v6 := mock2.NewValidatorMock([]byte("pubKey1"), 0, 11)
	v7 := mock2.NewValidatorMock([]byte("pubKey2"), 0, 2)
	v8 := mock2.NewValidatorMock([]byte("pubKey3"), 0, 4)
	v9 := mock2.NewValidatorMock([]byte("pubKey4"), 0, 5)
	v10 := mock2.NewValidatorMock([]byte("pubKey5"), 0, 6)
	v11 := mock2.NewValidatorMock([]byte("pubKey6"), 0, 8)
	v12 := mock2.NewValidatorMock([]byte("pubKey7"), 0, 10)
	group2 := []sharding.Validator{v6, v7, v8, v9, v10, v11, v12}

	byte1Map1, _ := strconv.ParseInt("00110011", 2, 8)
	byte2Map1, _ := strconv.ParseInt("00000001", 2, 8)
	bitmap1 := []byte{byte(byte1Map1), byte(byte2Map1)}

	byte1Map2, _ := strconv.ParseInt("00010001", 2, 8)
	byte2Map2, _ := strconv.ParseInt("00001101", 2, 8)
	bitmap2 := []byte{byte(byte1Map2), byte(byte2Map2)}

	doubleSigners := detector.DoubleSigners(group1, group2, bitmap1, bitmap2)

	require.Len(t, doubleSigners, 3)
	require.Equal(t, []byte("pubKey1"), doubleSigners[0].PubKey())
	require.Equal(t, []byte("pubKey3"), doubleSigners[1].PubKey())
	require.Equal(t, []byte("pubKey6"), doubleSigners[2].PubKey())
}
