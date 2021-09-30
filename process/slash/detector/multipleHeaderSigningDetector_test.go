package detector_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	mockEpochStart "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/slash/detector"
	"github.com/ElrondNetwork/elrond-go/sharding"
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
