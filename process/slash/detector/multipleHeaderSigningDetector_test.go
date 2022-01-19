package detector_test

import (
	"bytes"
	"errors"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	mockEpochStart "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/process/slash/detector"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/slashMocks"
	"github.com/stretchr/testify/require"
)

func TestNewSigningSlashingDetector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args        func() *detector.MultipleHeaderSigningDetectorArgs
		expectedErr error
	}{
		{
			args: func() *detector.MultipleHeaderSigningDetectorArgs {
				return nil
			},
			expectedErr: process.ErrNilMultipleHeaderSigningDetectorArgs,
		},
		{
			args: func() *detector.MultipleHeaderSigningDetectorArgs {
				args := generateMultipleHeaderSigningDetectorArgs()
				args.NodesCoordinator = nil
				return args
			},
			expectedErr: process.ErrNilNodesCoordinator,
		},
		{
			args: func() *detector.MultipleHeaderSigningDetectorArgs {
				args := generateMultipleHeaderSigningDetectorArgs()
				args.RoundHandler = nil
				return args
			},
			expectedErr: process.ErrNilRoundHandler,
		},
		{
			args: func() *detector.MultipleHeaderSigningDetectorArgs {
				args := generateMultipleHeaderSigningDetectorArgs()
				args.Hasher = nil
				return args
			},
			expectedErr: process.ErrNilHasher,
		},
		{
			args: func() *detector.MultipleHeaderSigningDetectorArgs {
				args := generateMultipleHeaderSigningDetectorArgs()
				args.Marshaller = nil
				return args
			},
			expectedErr: process.ErrNilMarshalizer,
		},
		{
			args: func() *detector.MultipleHeaderSigningDetectorArgs {
				args := generateMultipleHeaderSigningDetectorArgs()
				args.RoundValidatorHeadersCache = nil
				return args
			},
			expectedErr: process.ErrNilRoundValidatorHeadersCache,
		},
		{
			args: func() *detector.MultipleHeaderSigningDetectorArgs {
				args := generateMultipleHeaderSigningDetectorArgs()
				args.RoundHashCache = nil
				return args
			},
			expectedErr: process.ErrNilRoundHeadersCache,
		},
		{
			args: func() *detector.MultipleHeaderSigningDetectorArgs {
				args := generateMultipleHeaderSigningDetectorArgs()
				args.HeaderSigVerifier = nil
				return args
			},
			expectedErr: process.ErrNilHeaderSigVerifier,
		},
		{
			args: func() *detector.MultipleHeaderSigningDetectorArgs {
				return generateMultipleHeaderSigningDetectorArgs()
			},
			expectedErr: nil,
		},
	}

	for _, currTest := range tests {
		_, err := detector.NewMultipleHeaderSigningDetector(currTest.args())
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestMultipleHeaderSigningDetector_VerifyData_CannotCastData_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	sd, _ := detector.NewMultipleHeaderSigningDetector(args)

	res, err := sd.VerifyData(&testscommon.InterceptedDataStub{})

	require.Nil(t, res)
	require.Equal(t, process.ErrCannotCastInterceptedDataToHeader, err)
}

func TestMultipleHeaderSigningDetector_VerifyData_NilHeaderHandler_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	sd, _ := detector.NewMultipleHeaderSigningDetector(args)
	res, err := sd.VerifyData(&interceptedBlocks.InterceptedHeader{})

	require.Nil(t, res)
	require.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestMultipleHeaderSigningDetector_VerifyData_IrrelevantRound_ExpectError(t *testing.T) {
	t.Parallel()

	round := uint64(100)
	args := generateMultipleHeaderSigningDetectorArgs()
	args.RoundHandler = &mock.RoundHandlerMock{RoundIndex: int64(round)}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	hData := slashMocks.CreateInterceptedHeaderData(&block.Header{Round: round + detector.MaxDeltaToCurrentRound + 1, RandSeed: []byte("seed")})
	res, err := ssd.VerifyData(hData)

	require.Nil(t, res)
	require.Equal(t, process.ErrHeaderRoundNotRelevant, err)
}

func TestMultipleHeaderSigningDetector_VerifyData_CannotCacheHeaderWithoutSignature_ExpectErrorAndHeaderNotCached(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	addCache1Flag := atomic.Flag{}
	addCache2Flag := atomic.Flag{}
	args.RoundValidatorHeadersCache = &slashMocks.RoundDetectorCacheStub{
		AddCalled: func(uint64, []byte, data.HeaderInfoHandler) error {
			addCache1Flag.Set()
			return nil
		},
	}
	args.RoundHashCache = &slashMocks.HeadersCacheStub{
		AddCalled: func(uint64, []byte) error {
			addCache2Flag.Set()
			return nil
		},
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	errSetSig := errors.New("error set signature")
	headerCopy := &testscommon.HeaderHandlerStub{
		SetSignatureCalled: func(signature []byte) error {
			return errSetSig
		},
	}
	header := &testscommon.HeaderHandlerStub{
		CloneCalled: func() data.HeaderHandler {
			return headerCopy
		},
	}
	interceptedData := testscommon.InterceptedDataStub{
		HashCalled: func() []byte {
			return []byte("hash")
		},
	}
	interceptedHeader := &testscommon.InterceptedHeaderStub{
		HeaderHandlerCalled: func() data.HeaderHandler {
			return header
		},
		InterceptedDataStub: interceptedData,
	}

	proof, err := ssd.VerifyData(interceptedHeader)
	require.Nil(t, proof)
	require.Equal(t, errSetSig, err)
	require.False(t, addCache1Flag.IsSet())
	require.False(t, addCache2Flag.IsSet())
}

func TestMultipleHeaderSigningDetector_VerifyData_InvalidMarshaller_ExpectError(t *testing.T) {
	t.Parallel()

	errMarshaller := errors.New("error marshaller")
	args := generateMultipleHeaderSigningDetectorArgs()
	args.Marshaller = &mock.MarshalizerStub{
		MarshalCalled: func(_ interface{}) ([]byte, error) {
			return nil, errMarshaller
		},
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	hData := slashMocks.CreateInterceptedHeaderData(&block.Header{})
	res, err := ssd.VerifyData(hData)

	require.Nil(t, res)
	require.Equal(t, errMarshaller, err)
}

func TestMultipleHeaderSigningDetector_VerifyData_InvalidNodesCoordinator_ExpectError(t *testing.T) {
	t.Parallel()

	errNodesCoordinator := errors.New("error nodes coordinator")
	args := generateMultipleHeaderSigningDetectorArgs()
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return nil, errNodesCoordinator
		},
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	hData := slashMocks.CreateInterceptedHeaderData(&block.HeaderV2{Header: &block.Header{}})
	res, err := ssd.VerifyData(hData)

	require.Nil(t, res)
	require.Equal(t, errNodesCoordinator, err)
}

func TestMultipleHeaderSigningDetector_VerifyData_SameHeaderData_DifferentSigners_ExpectNoSlashingEvent(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	args.RoundHashCache = &slashMocks.HeadersCacheStub{
		AddCalled: func(round uint64, hash []byte) error {
			return process.ErrHeadersNotDifferentHashes
		},
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	hData1 := slashMocks.CreateInterceptedHeaderData(&block.HeaderV2{Header: &block.Header{Round: 2, TimeStamp: 5, Signature: []byte("signature")}})
	res, err := ssd.VerifyData(hData1)
	require.Nil(t, res)
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)

	hData2 := slashMocks.CreateInterceptedHeaderData(&block.HeaderV2{Header: &block.Header{Round: 2, TimeStamp: 5, LeaderSignature: []byte("leaderSignature")}})
	res, err = ssd.VerifyData(hData2)
	require.Nil(t, res)
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)

	hData3 := slashMocks.CreateInterceptedHeaderData(&block.HeaderV2{Header: &block.Header{Round: 2, TimeStamp: 5, PubKeysBitmap: []byte("bitmap")}})
	res, err = ssd.VerifyData(hData3)
	require.Nil(t, res)
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)
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

	hData1 := slashMocks.CreateInterceptedHeaderData(&block.HeaderV2{
		Header: &block.Header{
			PrevRandSeed:  []byte("rnd1"),
			Round:         2,
			PubKeysBitmap: bitmap1,
		},
	})
	hData2 := slashMocks.CreateInterceptedHeaderData(&block.HeaderV2{
		Header: &block.Header{
			PrevRandSeed:  []byte("rnd2"),
			Round:         2,
			PubKeysBitmap: bitmap2,
		},
	})
	hData3 := slashMocks.CreateInterceptedHeaderData(&block.HeaderV2{
		Header: &block.Header{
			PrevRandSeed:  []byte("rnd3"),
			Round:         2,
			PubKeysBitmap: bitmap3,
		},
	})

	args := generateMultipleHeaderSigningDetectorArgs()
	args.RoundValidatorHeadersCache = detector.NewRoundValidatorHeaderCache(3)
	args.RoundHashCache = detector.NewRoundHashCache(3)
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
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
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	// For first header(same round): v0, v1, v3 signed => no slashing event
	tmp, err := ssd.VerifyData(hData1)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	// For 2nd header(same round): v0, v2, v4 signed => v0 signed 2 headers this round(current and previous)
	tmp, err = ssd.VerifyData(hData2)
	res := tmp.(coreSlash.MultipleSigningProofHandler)
	errProof := ssd.ValidateProof(res)
	require.Nil(t, err)
	require.Nil(t, errProof)
	require.Equal(t, coreSlash.MultipleSigning, res.GetType())

	require.Len(t, res.GetPubKeys(), 1)
	require.Equal(t, pk0, res.GetPubKeys()[0])
	require.Equal(t, coreSlash.Medium, res.GetLevel(pk0))

	require.Len(t, res.GetHeaders(pk0), 2)
	require.Contains(t, res.GetHeaders(pk0), hData1.HeaderHandler())
	require.Contains(t, res.GetHeaders(pk0), hData2.HeaderHandler())

	// For 3rd header(same round): v0, v1 signed =>
	// 1. v0 signed 3 headers this round(current and previous 2 headers)
	// 2. v1 signed 2 headers this round(current and first header)
	tmp, err = ssd.VerifyData(hData3)
	res = tmp.(coreSlash.MultipleSigningProofHandler)
	errProof = ssd.ValidateProof(res)
	require.Nil(t, err)
	require.Nil(t, errProof)
	require.Equal(t, coreSlash.MultipleSigning, res.GetType())

	require.Len(t, res.GetPubKeys(), 2)
	require.Contains(t, res.GetPubKeys(), pk0)
	require.Contains(t, res.GetPubKeys(), pk1)
	require.Equal(t, coreSlash.High, res.GetLevel(pk0))
	require.Equal(t, coreSlash.Medium, res.GetLevel(pk1))

	require.Len(t, res.GetHeaders(pk0), 3)
	require.Contains(t, res.GetHeaders(pk0), hData1.HeaderHandler())
	require.Contains(t, res.GetHeaders(pk0), hData2.HeaderHandler())
	require.Contains(t, res.GetHeaders(pk0), hData3.HeaderHandler())

	require.Len(t, res.GetHeaders(pk1), 2)
	require.Contains(t, res.GetHeaders(pk1), hData1.HeaderHandler())
	require.Contains(t, res.GetHeaders(pk1), hData3.HeaderHandler())

	// 4th header(same round) == 2nd header, but validators are changed within group =>
	// no slashing, because headers do not have different hash (without signatures). This
	// should be a slashing case of multiple header proposal.
	group2 = []sharding.Validator{v4, v2, v0, v1}
	tmp, err = ssd.VerifyData(hData2)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)
}

func TestMultipleHeaderSigningDetector_VerifyData_ValidateProof_CachingSignersFailed(t *testing.T) {
	t.Parallel()

	pubKey := []byte("pubKey0")
	validator := mock.NewValidatorMock(pubKey)

	group := []sharding.Validator{validator}
	byteMap, _ := strconv.ParseInt("00000001", 2, 9)
	bitmap := []byte{byte(byteMap)}

	h1 := &block.HeaderV2{
		Header: &block.Header{
			TimeStamp:     1,
			PrevRandSeed:  []byte("rnd"),
			Round:         2,
			PubKeysBitmap: bitmap,
		},
	}
	h2 := &block.HeaderV2{
		Header: &block.Header{
			TimeStamp:     2,
			PrevRandSeed:  []byte("rnd"),
			Round:         2,
			PubKeysBitmap: bitmap,
		},
	}

	hData1 := slashMocks.CreateInterceptedHeaderData(h1)
	hData2 := slashMocks.CreateInterceptedHeaderData(h2)

	computeConsensusGroupCalledCt := 0
	errComputeConsensusGroup := errors.New("error computing consensus group")

	args := generateMultipleHeaderSigningDetectorArgs()
	args.RoundValidatorHeadersCache = detector.NewRoundValidatorHeaderCache(3)
	args.RoundHashCache = detector.NewRoundHashCache(3)
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(randomness []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			computeConsensusGroupCalledCt++
			if computeConsensusGroupCalledCt == 2 {
				return nil, errComputeConsensusGroup
			}
			if bytes.Equal(randomness, []byte("rnd")) {
				return group, nil
			}
			return nil, nil
		},
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	// Header1 signed by validator => header1 is cached
	// No slashing event
	res, err := ssd.VerifyData(hData1)
	require.Nil(t, res)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	// Header2 signed by validator => header2 is cached.
	// When trying to compute consensus group to cache signers, we got an error => header2 is removed from cache
	res, err = ssd.VerifyData(hData2)
	require.Nil(t, res)
	require.Equal(t, errComputeConsensusGroup, err)

	// Same header2 signed by validator => header2 is cached (without error, because it was removed before)
	// Validator signed two different headers => slash event
	res, err = ssd.VerifyData(hData2)
	require.Nil(t, err)

	proof := res.(coreSlash.MultipleSigningProofHandler)
	err = ssd.ValidateProof(proof)
	require.Nil(t, err)

	require.Equal(t, coreSlash.MultipleSigning, proof.GetType())
	require.Equal(t, coreSlash.Medium, proof.GetLevel(pubKey))

	require.Len(t, proof.GetPubKeys(), 1)
	require.Contains(t, proof.GetPubKeys(), pubKey)

	require.Len(t, proof.GetHeaders(pubKey), 2)
	require.Contains(t, proof.GetHeaders(pubKey), h1)
	require.Contains(t, proof.GetHeaders(pubKey), h2)
}

func TestMultipleHeaderSigningDetector_ValidateProof_NotEnoughPubKeys_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	proof := &slashMocks.MultipleHeaderSigningProofStub{}
	err := ssd.ValidateProof(proof)

	require.Equal(t, process.ErrNotEnoughPublicKeysProvided, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_InvalidProofType_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	proof1 := &slashMocks.MultipleHeaderProposalProofStub{
		GetTypeCalled: func() coreSlash.SlashingType {
			return coreSlash.MultipleSigning
		},
	}
	err := ssd.ValidateProof(proof1)
	require.Equal(t, process.ErrCannotCastProofToMultipleSignedHeaders, err)

	proof2 := &slashMocks.MultipleHeaderSigningProofStub{
		GetTypeCalled: func() coreSlash.SlashingType {
			return coreSlash.MultipleProposal
		},
	}
	err = ssd.ValidateProof(proof2)
	require.Equal(t, process.ErrInvalidSlashType, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_NotEnoughHeaders_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)
	slashRes := map[string]coreSlash.SlashingResult{
		"pubKey": {
			SlashingLevel: coreSlash.Medium,
			Headers:       slash.HeaderInfoList{},
		},
	}

	proof, _ := coreSlash.NewMultipleSigningProof(slashRes)
	err := ssd.ValidateProof(proof)
	require.Equal(t, process.ErrNotEnoughHeadersProvided, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_SignedHeadersHaveDifferentRound_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock([]byte("pubKey"))}, nil
		},
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	h1 := &block.HeaderV2{Header: &block.Header{Round: 1, PubKeysBitmap: []byte{byte(0x1)}}}
	h2 := &block.HeaderV2{Header: &block.Header{Round: 2, PubKeysBitmap: []byte{byte(0x1)}}}
	hInfo1 := slashMocks.CreateHeaderInfoData(h1)
	hInfo2 := slashMocks.CreateHeaderInfoData(h2)
	proof, _ := coreSlash.NewMultipleSigningProof(map[string]coreSlash.SlashingResult{
		"pubKey": {
			SlashingLevel: coreSlash.Medium,
			Headers:       slash.HeaderInfoList{hInfo1, hInfo2},
		},
	})

	err := ssd.ValidateProof(proof)
	require.Equal(t, process.ErrHeadersNotSameRound, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_InvalidMarshaller_ExpectError(t *testing.T) {
	t.Parallel()

	errMarshaller := errors.New("error marshaller")
	args := generateMultipleHeaderSigningDetectorArgs()
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock([]byte("pubKey"))}, nil
		},
	}
	args.Marshaller = &mock.MarshalizerStub{
		MarshalCalled: func(_ interface{}) ([]byte, error) {
			return nil, errMarshaller
		},
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	h1 := &block.HeaderV2{Header: &block.Header{Round: 1, PubKeysBitmap: []byte{byte(0x1)}}}
	h2 := &block.HeaderV2{Header: &block.Header{Round: 2, PubKeysBitmap: []byte{byte(0x1)}}}
	hInfo1 := slashMocks.CreateHeaderInfoData(h1)
	hInfo2 := slashMocks.CreateHeaderInfoData(h2)
	proof, _ := coreSlash.NewMultipleSigningProof(map[string]coreSlash.SlashingResult{
		"pubKey": {
			SlashingLevel: coreSlash.Medium,
			Headers:       slash.HeaderInfoList{hInfo1, hInfo2},
		},
	})

	err := ssd.ValidateProof(proof)
	require.Equal(t, errMarshaller, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_SignedHeadersHaveSameHash_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock([]byte("pubKey"))}, nil
		},
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	h1 := &block.HeaderV2{Header: &block.Header{Round: 1, PubKeysBitmap: []byte{byte(0x1)}}}
	h2 := &block.HeaderV2{Header: &block.Header{Round: 1, PubKeysBitmap: []byte{byte(0x1)}}}
	hInfo1 := slashMocks.CreateHeaderInfoData(h1)
	hInfo2 := slashMocks.CreateHeaderInfoData(h2)
	proof, _ := coreSlash.NewMultipleSigningProof(map[string]coreSlash.SlashingResult{
		"pubKey": {
			SlashingLevel: coreSlash.Medium,
			Headers:       slash.HeaderInfoList{hInfo1, hInfo2},
		},
	})

	err := ssd.ValidateProof(proof)
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_HeadersNotSignedByTheSameValidator_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock([]byte("pubKey2"))}, nil
		},
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	h1 := &block.HeaderV2{Header: &block.Header{Round: 1, PubKeysBitmap: []byte{byte(0x1)}}}
	h2 := &block.HeaderV2{Header: &block.Header{Round: 1, PubKeysBitmap: []byte{byte(0x2)}}}
	hInfo1 := slashMocks.CreateHeaderInfoData(h1)
	hInfo2 := slashMocks.CreateHeaderInfoData(h2)
	proof, _ := coreSlash.NewMultipleSigningProof(map[string]coreSlash.SlashingResult{
		"pubKey": {
			SlashingLevel: coreSlash.Medium,
			Headers:       slash.HeaderInfoList{hInfo1, hInfo2},
		},
	})

	err := ssd.ValidateProof(proof)
	require.Equal(t, process.ErrHeaderNotSignedByValidator, err)
}

func TestMultipleHeaderSigningDetector_ValidateProof_InvalidSlashLevel_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	h1 := &block.HeaderV2{Header: &block.Header{Round: 1}}
	h2 := &block.HeaderV2{Header: &block.Header{Round: 1}}
	hInfo1 := slashMocks.CreateHeaderInfoData(h1)
	hInfo2 := slashMocks.CreateHeaderInfoData(h2)
	proof, _ := coreSlash.NewMultipleSigningProof(map[string]coreSlash.SlashingResult{
		"pubKey": {
			SlashingLevel: coreSlash.Low,
			Headers:       slash.HeaderInfoList{hInfo1, hInfo2},
		},
	})

	err := ssd.ValidateProof(proof)
	require.Equal(t, process.ErrInvalidSlashLevel, err)
}

func TestMultipleHeaderSigningDetector_CheckSignedHeaders_NotEnoughHeaders_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	err := ssd.CheckSignedHeaders([]byte("validator"), slash.HeaderList{})
	require.Equal(t, process.ErrNotEnoughHeadersProvided, err)
}

func TestMultipleHeaderProposalsDetector_CheckSignedHeaders_NilHeaders_ExpectErr(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	sd, _ := detector.NewMultipleHeaderSigningDetector(args)

	header1 := data.HeaderHandler(nil)
	header2 := data.HeaderHandler(nil)
	header3 := data.HeaderHandler(nil)

	// All headers nil
	headers := []data.HeaderHandler{header1, header2, header3}
	err := sd.CheckSignedHeaders([]byte("validator"), headers)
	require.Equal(t, process.ErrNilHeaderHandler, err)

	// First header valid, second and third headers nil
	header1 = &block.Header{Round: 1, TimeStamp: 1, PubKeysBitmap: []byte{0x1}}
	headers = []data.HeaderHandler{header1, header2, header3}
	err = sd.CheckSignedHeaders([]byte("validator"), headers)
	require.Equal(t, process.ErrNilHeaderHandler, err)

	// First and second header valid, third header nil
	header2 = &block.Header{Round: 1, TimeStamp: 2, PubKeysBitmap: []byte{0x1}}
	headers = []data.HeaderHandler{header1, header2, header3}
	err = sd.CheckSignedHeaders([]byte("validator"), headers)
	require.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestMultipleHeaderSigningDetector_SignedHeader_CannotGetConsensusGroup_ExpectFalse(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func([]byte, uint64, uint32, uint32) ([]sharding.Validator, error) {
			return nil, errors.New("error computing consensus group")
		},
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	header := &block.Header{Round: 1}
	signedHeader := ssd.SignedHeader([]byte("validator"), header)
	require.False(t, signedHeader)
}

func TestMultipleHeaderSigningDetector_SignedHeader_ValidatorNotInConsensusGroup_ExpectFalse(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func([]byte, uint64, uint32, uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{}, nil
		},
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	header := &block.Header{Round: 1}
	signedHeader := ssd.SignedHeader([]byte("validator"), header)
	require.False(t, signedHeader)
}

func TestMultipleHeaderSigningDetector_SignedHeader_CannotVerifySignature_ExpectFalse(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderSigningDetectorArgs()
	pubKey := []byte("validator")
	validator := mock.NewValidatorMock(pubKey)
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func([]byte, uint64, uint32, uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{validator}, nil
		},
	}
	args.HeaderSigVerifier = &mock.HeaderSigVerifierStub{
		VerifySignatureCalled: func(data.HeaderHandler) error {
			return errors.New("cannot verify signature")
		},
	}
	ssd, _ := detector.NewMultipleHeaderSigningDetector(args)

	header := &block.Header{Round: 1}
	signedHeader := ssd.SignedHeader(pubKey, header)
	require.False(t, signedHeader)
}

func generateMultipleHeaderSigningDetectorArgs() *detector.MultipleHeaderSigningDetectorArgs {
	nodesCoordinator := &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			validator := mock.NewValidatorMock([]byte("validator"))
			return []sharding.Validator{validator}, nil
		},
	}

	return &detector.MultipleHeaderSigningDetectorArgs{
		NodesCoordinator:           nodesCoordinator,
		RoundHandler:               &mock.RoundHandlerMock{},
		Hasher:                     &hashingMocks.HasherMock{},
		Marshaller:                 &mock.MarshalizerMock{},
		RoundValidatorHeadersCache: &slashMocks.RoundDetectorCacheStub{},
		RoundHashCache:             &slashMocks.HeadersCacheStub{},
		HeaderSigVerifier:          &mock.HeaderSigVerifierStub{},
	}
}
