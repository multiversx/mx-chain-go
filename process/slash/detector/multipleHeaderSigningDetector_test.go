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
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/process/slash/detector"
	mockSlash "github.com/ElrondNetwork/elrond-go/process/slash/mock"
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

func TestMultipleHeaderSigningDetector_VerifyData_InvalidNodesCoordinator_ExpectError(t *testing.T) {
	t.Parallel()

	errNodesCoordinator := errors.New("error nodes coordinator")
	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{
			ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
				return nil, errNodesCoordinator
			},
		},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		detector.CacheSize)

	hData := createInterceptedHeaderData(&block.Header{})
	res, err := ssd.VerifyData(hData)

	require.Nil(t, res)
	require.Equal(t, errNodesCoordinator, err)
}

func TestMultipleHeaderSigningDetector_VerifyData_SameHeaderData_DifferentSigners_ExpectNoSlashingEvent(t *testing.T) {
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

func TestMultipleHeaderSigningDetector_VerifyData_ValidateProof(t *testing.T) {
	t.Parallel()

	pk0 := []byte("pubKey0")
	pk1 := []byte("pubKey1")
	pk2 := []byte("pubKey2")
	pk3 := []byte("pubKey3")
	pk4 := []byte("pubKey4")
	v0 := mock.NewValidatorMock(pk0)
	v1 := mock.NewValidatorMock(pk1)
	v2 := mock.NewValidatorMock(pk2)
	v3 := mock.NewValidatorMock(pk3)
	v4 := mock.NewValidatorMock(pk4)

	group1 := []sharding.Validator{v0, v1, v3}
	byteMap1, _ := strconv.ParseInt("00000111", 2, 9)
	bitmap1 := []byte{byte(byteMap1)}

	group2 := []sharding.Validator{v0, v2, v4}
	byteMap2, _ := strconv.ParseInt("00000111", 2, 9)
	bitmap2 := []byte{byte(byteMap2)}

	group3 := []sharding.Validator{v4, v3, v2, v1, v0}
	byteMap3, _ := strconv.ParseInt("00011000", 2, 9)
	bitmap3 := []byte{byte(byteMap3)}

	hData1 := createInterceptedHeaderData(
		&block.Header{
			PrevRandSeed:  []byte("rnd1"),
			Round:         2,
			PubKeysBitmap: bitmap1,
		},
	)
	hData2 := createInterceptedHeaderData(
		&block.Header{
			PrevRandSeed:  []byte("rnd2"),
			Round:         2,
			PubKeysBitmap: bitmap2,
		},
	)
	hData3 := createInterceptedHeaderData(
		&block.Header{
			PrevRandSeed:  []byte("rnd3"),
			Round:         2,
			PubKeysBitmap: bitmap3,
		},
	)

	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{
			ComputeConsensusGroupCalled: func(randomness []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
				switch string(randomness) {
				case "rnd1":
					return group1, nil
				case "rnd2":
					return group2, nil
				case "rnd3":
					return group3, nil
				default:
					return nil, nil
				}
			},
		},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		detector.CacheSize)

	// For first header(same round): v0, v1, v3 signed => no slashing event
	tmp, err := ssd.VerifyData(hData1)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	// For 2nd header(same round): v0, v2, v4 signed => v0 signed 2 headers this round(current and previous)
	tmp, err = ssd.VerifyData(hData2)
	res := tmp.(slash.MultipleSigningProofHandler)
	errProof := ssd.ValidateProof(res)
	require.Nil(t, err)
	require.Nil(t, errProof)
	require.Equal(t, slash.MultipleSigning, res.GetType())

	require.Len(t, res.GetPubKeys(), 1)
	require.Equal(t, pk0, res.GetPubKeys()[0])
	require.Equal(t, slash.Level1, res.GetLevel(pk0))

	require.Len(t, res.GetHeaders(pk0), 2)
	require.Equal(t, []byte("rnd1"), res.GetHeaders(pk0)[0].HeaderHandler().GetPrevRandSeed())
	require.Equal(t, []byte("rnd2"), res.GetHeaders(pk0)[1].HeaderHandler().GetPrevRandSeed())

	// For 3rd header(same round): v0, v1 signed =>
	// 1. v0 signed 3 headers this round(current and previous 2 headers)
	// 2. v1 signed 2 headers this round(current and first header)
	tmp, err = ssd.VerifyData(hData3)
	res = tmp.(slash.MultipleSigningProofHandler)
	errProof = ssd.ValidateProof(res)
	require.Nil(t, err)
	require.Nil(t, errProof)
	require.Equal(t, slash.MultipleSigning, res.GetType())

	require.Len(t, res.GetPubKeys(), 2)
	require.Equal(t, pk0, res.GetPubKeys()[0])
	require.Equal(t, pk1, res.GetPubKeys()[1])
	require.Equal(t, slash.Level2, res.GetLevel(pk0))
	require.Equal(t, slash.Level1, res.GetLevel(pk1))

	require.Len(t, res.GetHeaders(pk0), 3)
	require.Equal(t, []byte("rnd1"), res.GetHeaders(pk0)[0].HeaderHandler().GetPrevRandSeed())
	require.Equal(t, []byte("rnd2"), res.GetHeaders(pk0)[1].HeaderHandler().GetPrevRandSeed())
	require.Equal(t, []byte("rnd3"), res.GetHeaders(pk0)[2].HeaderHandler().GetPrevRandSeed())

	require.Len(t, res.GetHeaders(pk1), 2)
	require.Equal(t, []byte("rnd1"), res.GetHeaders(pk1)[0].HeaderHandler().GetPrevRandSeed())
	require.Equal(t, []byte("rnd3"), res.GetHeaders(pk1)[1].HeaderHandler().GetPrevRandSeed())

	// 4th header(same round) == 2nd header, but validators are changed within group =>
	// no slashing, because headers do not have different hash (without signatures). This
	// should be a slashing case of multiple header proposal.
	group2 = []sharding.Validator{v4, v2, v0, v1}
	tmp, err = ssd.VerifyData(hData2)
	require.Nil(t, tmp)
	require.Error(t, process.ErrHeadersShouldHaveDifferentHashes, hData2)
}

func TestMultipleHeaderSigningDetector_ValidateProof_InvalidProofType_ExpectError(t *testing.T) {
	t.Parallel()

	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		detector.CacheSize)

	err := ssd.ValidateProof(&mockSlash.MultipleHeaderProposalProofStub{})
	require.Equal(t, process.ErrCannotCastProofToMultipleSignedHeaders, err)

	err = ssd.ValidateProof(&mockSlash.MultipleHeaderSigningProofStub{
		GetTypeCalled: func() slash.SlashingType {
			return slash.MultipleProposal
		},
	})
	require.Equal(t, process.ErrInvalidSlashType, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_NotEnoughHeaders_ExpectError(t *testing.T) {
	t.Parallel()

	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		detector.CacheSize)
	proof, _ := slash.NewMultipleSigningProof(map[string]slash.SlashingData{
		"pubKey": {
			SlashingLevel: slash.Level1,
			Data:          []process.InterceptedData{},
		},
	})

	err := ssd.ValidateProof(proof)
	require.Equal(t, process.ErrNotEnoughHeadersProvided, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_SignedHeadersHaveDifferentRound_ExpectError(t *testing.T) {
	t.Parallel()

	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{
			ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
				return []sharding.Validator{mock.NewValidatorMock([]byte("pubKey"))}, nil
			},
		},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		detector.CacheSize)

	h1 := createInterceptedHeaderData(&block.Header{Round: 1, PubKeysBitmap: []byte{byte(0x1)}})
	h2 := createInterceptedHeaderData(&block.Header{Round: 2, PubKeysBitmap: []byte{byte(0x1)}})
	proof, _ := slash.NewMultipleSigningProof(map[string]slash.SlashingData{
		"pubKey": {
			SlashingLevel: slash.Level1,
			Data:          []process.InterceptedData{h1, h2},
		},
	})

	err := ssd.ValidateProof(proof)
	require.Equal(t, process.ErrHeadersDoNotHaveSameRound, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_InvalidMarshaller_ExpectError(t *testing.T) {
	t.Parallel()

	errMarshaller := errors.New("error marshaller")
	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{
			ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
				return []sharding.Validator{mock.NewValidatorMock([]byte("pubKey"))}, nil
			},
		},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerStub{
			MarshalCalled: func(_ interface{}) ([]byte, error) {
				return nil, errMarshaller
			},
		},
		detector.CacheSize)

	h1 := createInterceptedHeaderData(&block.Header{Round: 1, PubKeysBitmap: []byte{byte(0x1)}})
	h2 := createInterceptedHeaderData(&block.Header{Round: 2, PubKeysBitmap: []byte{byte(0x1)}})
	proof, _ := slash.NewMultipleSigningProof(map[string]slash.SlashingData{
		"pubKey": {
			SlashingLevel: slash.Level1,
			Data:          []process.InterceptedData{h1, h2},
		},
	})

	err := ssd.ValidateProof(proof)
	require.Equal(t, errMarshaller, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_SignedHeadersHaveSameHash_ExpectError(t *testing.T) {
	t.Parallel()

	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{
			ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
				return []sharding.Validator{mock.NewValidatorMock([]byte("pubKey"))}, nil
			},
		},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		detector.CacheSize)

	h1 := createInterceptedHeaderData(&block.Header{Round: 1, PubKeysBitmap: []byte{byte(0x1)}})
	h2 := createInterceptedHeaderData(&block.Header{Round: 1, PubKeysBitmap: []byte{byte(0x1)}})
	proof, _ := slash.NewMultipleSigningProof(map[string]slash.SlashingData{
		"pubKey": {
			SlashingLevel: slash.Level1,
			Data:          []process.InterceptedData{h1, h2},
		},
	})

	err := ssd.ValidateProof(proof)
	require.Equal(t, process.ErrProposedHeadersDoNotHaveDifferentHashes, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_HeadersNotSignedByTheSameValidator_ExpectError(t *testing.T) {
	t.Parallel()

	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{
			ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
				return []sharding.Validator{mock.NewValidatorMock([]byte("pubKey2"))}, nil
			},
		},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		detector.CacheSize)

	h1 := createInterceptedHeaderData(&block.Header{Round: 1, PubKeysBitmap: []byte{byte(0x1)}})
	h2 := createInterceptedHeaderData(&block.Header{Round: 1, PubKeysBitmap: []byte{byte(0x2)}})
	proof, _ := slash.NewMultipleSigningProof(map[string]slash.SlashingData{
		"pubKey": {
			SlashingLevel: slash.Level1,
			Data:          []process.InterceptedData{h1, h2},
		},
	})

	err := ssd.ValidateProof(proof)
	require.Equal(t, process.ErrHeaderNotSignedByValidator, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_InvalidSlashLevel_ExpectError(t *testing.T) {
	t.Parallel()

	ssd, _ := detector.NewSigningSlashingDetector(
		&mockEpochStart.NodesCoordinatorStub{},
		&mock.RoundHandlerMock{},
		&mock.HasherMock{},
		&mock.MarshalizerMock{},
		detector.CacheSize)

	h1 := createInterceptedHeaderData(&block.Header{Round: 1})
	h2 := createInterceptedHeaderData(&block.Header{Round: 1})
	proof, _ := slash.NewMultipleSigningProof(map[string]slash.SlashingData{
		"pubKey": {
			SlashingLevel: slash.Level0,
			Data:          []process.InterceptedData{h1, h2},
		},
	})

	err := ssd.ValidateProof(proof)
	require.Equal(t, process.ErrInvalidSlashLevel, err)
}

func TestMultipleHeaderSigningDetector_IsIndexSetInBitmap(t *testing.T) {
	byte1Map1, _ := strconv.ParseInt("11001101", 2, 9)
	byte2Map1, _ := strconv.ParseInt("00000101", 2, 9)
	bitmap := []byte{byte(byte1Map1), byte(byte2Map1)}

	//Byte 1
	require.True(t, detector.IsIndexSetInBitmap(0, bitmap))
	require.False(t, detector.IsIndexSetInBitmap(1, bitmap))
	require.True(t, detector.IsIndexSetInBitmap(2, bitmap))
	require.True(t, detector.IsIndexSetInBitmap(3, bitmap))
	require.False(t, detector.IsIndexSetInBitmap(4, bitmap))
	require.False(t, detector.IsIndexSetInBitmap(5, bitmap))
	require.True(t, detector.IsIndexSetInBitmap(6, bitmap))
	require.True(t, detector.IsIndexSetInBitmap(7, bitmap))
	// Byte 2
	require.True(t, detector.IsIndexSetInBitmap(8, bitmap))
	require.False(t, detector.IsIndexSetInBitmap(9, bitmap))
	require.True(t, detector.IsIndexSetInBitmap(10, bitmap))

	for i := uint32(11); i <= 100; i++ {
		require.False(t, detector.IsIndexSetInBitmap(i, bitmap))
	}
}
