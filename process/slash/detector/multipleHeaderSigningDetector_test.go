package detector_test

import (
	"errors"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
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

func TestMultipleHeaderSigningDetector_VerifyData_IrrelevantRounds_ExpectError(t *testing.T) {
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

	hData := createInterceptedHeaderData(&block.Header{Round: round + detector.MaxDeltaToCurrentRound + 1, RandSeed: []byte("seed")})
	res, err := ssd.VerifyData(hData)

	require.Nil(t, res)
	require.Equal(t, process.ErrHeaderRoundNotRelevant, err)
}

func TestMultipleHeaderSigningDetector_VerifyData_InvalidMarshaller_ExpectError(t *testing.T) {
	t.Parallel()

	errMarshaller := errors.New("error marshaller")
	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerStub{
			MarshalCalled: func(_ interface{}) ([]byte, error) {
				return nil, errMarshaller
			},
		},
		detector.CacheSize)

	hData := createInterceptedHeaderData(&block.Header{})
	res, err := ssd.VerifyData(hData)

	require.Nil(t, res)
	require.Equal(t, errMarshaller, err)
}

func TestMultipleHeaderSigningDetector_VerifyData_SameHeaderData_DifferentSigners_ExpectError(t *testing.T) {
	t.Parallel()

	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		detector.CacheSize)

	hData1 := createInterceptedHeaderData(&block.Header{Round: 2, Signature: []byte("signature")})
	res, err := ssd.VerifyData(hData1)
	require.Nil(t, res)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	hData2 := createInterceptedHeaderData(&block.Header{Round: 2, LeaderSignature: []byte("leaderSignature")})
	res, err = ssd.VerifyData(hData2)
	require.Nil(t, res)
	require.Equal(t, process.ErrHeadersShouldHaveDifferentHashes, err)

	hData3 := createInterceptedHeaderData(&block.Header{Round: 2, PubKeysBitmap: []byte("bitmap")})
	res, err = ssd.VerifyData(hData3)
	require.Nil(t, res)
	require.Equal(t, process.ErrHeadersShouldHaveDifferentHashes, err)
}

func TestMultipleHeaderSigningDetector_VerifyData(t *testing.T) {
	t.Parallel()

	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		detector.CacheSize)

	hData1 := createInterceptedHeaderData(
		&block.Header{
			Round:   2,
			TxCount: 4,
		},
	)
	tmp, err := ssd.VerifyData(hData1)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	//hData2 := createInterceptedHeaderData(
	//	&block.Header{
	//		Round:   2,
	//		TxCount: 5,
	//	},
	//)
	//tmp, err = ssd.VerifyData(hData2)
	//res := tmp.(slash.MultipleSigningProofHandler)
	//require.NotNil(t, res)
	//require.Equal(t, process.ErrNoSlashingEventDetected, err)
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
	v0g1 := mock.NewValidatorMock([]byte("pubKey0"))
	v1g1 := mock.NewValidatorMock([]byte("pubKey1"))
	v2g1 := mock.NewValidatorMock([]byte("pubKey2"))
	v3g1 := mock.NewValidatorMock([]byte("pubKey3"))
	v4g1 := mock.NewValidatorMock([]byte("pubKey4"))
	v5g1 := mock.NewValidatorMock([]byte("pubKey5"))
	v6g1 := mock.NewValidatorMock([]byte("pubKey6"))
	v7g1 := mock.NewValidatorMock([]byte("pubKey7"))
	v8g1 := mock.NewValidatorMock([]byte("pubKey8"))

	group1 := []sharding.Validator{v0g1, v1g1, v2g1, v3g1, v4g1, v5g1, v6g1, v7g1, v8g1}

	v0g2 := mock.NewValidatorMock([]byte("pubKey0"))
	v1g2 := mock.NewValidatorMock([]byte("pubKey3"))
	v2g2 := mock.NewValidatorMock([]byte("pubKey1"))
	v3g2 := mock.NewValidatorMock([]byte("pubKey2"))
	v4g2 := mock.NewValidatorMock([]byte("pubKey6"))
	v5g2 := mock.NewValidatorMock([]byte("pubKey11"))
	v6g2 := mock.NewValidatorMock([]byte("pubKey13"))
	v7g2 := mock.NewValidatorMock([]byte("pubKey15"))
	v8g2 := mock.NewValidatorMock([]byte("pubKey8"))

	group2 := []sharding.Validator{v0g2, v1g2, v2g2, v3g2, v4g2, v5g2, v6g2, v7g2, v8g2}

	byte1Map1, _ := strconv.ParseInt("01011111", 2, 9)
	byte2Map1, _ := strconv.ParseInt("00000001", 2, 9)
	bitmap1 := []byte{byte(byte1Map1), byte(byte2Map1)}

	byte1Map2, _ := strconv.ParseInt("10100011", 2, 9)
	byte2Map2, _ := strconv.ParseInt("00000001", 2, 9)
	bitmap2 := []byte{byte(byte1Map2), byte(byte2Map2)}

	doubleSigners := detector.DoubleSigners(group1, group2, bitmap1, bitmap2)

	require.Len(t, doubleSigners, 3)
	require.Equal(t, []byte("pubKey0"), doubleSigners[0].PubKey())
	require.Equal(t, []byte("pubKey3"), doubleSigners[1].PubKey())
	require.Equal(t, []byte("pubKey8"), doubleSigners[2].PubKey())
}
