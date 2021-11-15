package detector_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	mockCoreData "github.com/ElrondNetwork/elrond-go-core/data/mock"
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

func TestNewMultipleHeaderProposalsDetector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args        func() *detector.MultipleHeaderProposalDetectorArgs
		expectedErr error
	}{
		{
			args: func() *detector.MultipleHeaderProposalDetectorArgs {
				return nil
			},
			expectedErr: process.ErrNilMultipleHeaderProposalDetectorArgs,
		},
		{
			args: func() *detector.MultipleHeaderProposalDetectorArgs {
				args := generateMultipleHeaderProposalDetectorArgs()
				args.NodesCoordinator = nil
				return args
			},
			expectedErr: process.ErrNilShardCoordinator,
		},
		{
			args: func() *detector.MultipleHeaderProposalDetectorArgs {
				args := generateMultipleHeaderProposalDetectorArgs()
				args.RoundHandler = nil
				return args
			},
			expectedErr: process.ErrNilRoundHandler,
		},
		{
			args: func() *detector.MultipleHeaderProposalDetectorArgs {
				args := generateMultipleHeaderProposalDetectorArgs()
				args.Cache = nil
				return args
			},
			expectedErr: process.ErrNilRoundDetectorCache,
		},
		{
			args: func() *detector.MultipleHeaderProposalDetectorArgs {
				args := generateMultipleHeaderProposalDetectorArgs()
				args.Hasher = nil
				return args
			},
			expectedErr: process.ErrNilHasher,
		},
		{
			args: func() *detector.MultipleHeaderProposalDetectorArgs {
				args := generateMultipleHeaderProposalDetectorArgs()
				args.Marshaller = nil
				return args
			},
			expectedErr: process.ErrNilMarshalizer,
		},
	}

	for _, currTest := range tests {
		_, err := detector.NewMultipleHeaderProposalsDetector(currTest.args())
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestMultipleHeaderProposalsDetector_VerifyData_Nil_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	res, err := sd.VerifyData(nil)
	require.Nil(t, res)
	require.Equal(t, process.ErrNilInterceptedData, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_CannotCastData_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	res, err := sd.VerifyData(&testscommon.InterceptedDataStub{})
	require.Nil(t, res)
	require.Equal(t, process.ErrCannotCastInterceptedDataToHeader, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_NilHeaderHandler_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	res, err := sd.VerifyData(&interceptedBlocks.InterceptedHeader{})
	require.Nil(t, res)
	require.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_CannotGetProposer_ExpectError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("cannot get proposer")
	args := generateMultipleHeaderProposalDetectorArgs()
	nodesCoordinator := &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return nil, expectedErr
		},
	}
	args.NodesCoordinator = nodesCoordinator
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	res, err := sd.VerifyData(slashMocks.CreateInterceptedHeaderData(&block.Header{}))
	require.Nil(t, res)
	require.Equal(t, expectedErr, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_IrrelevantRound_ExpectError(t *testing.T) {
	t.Parallel()

	round := uint64(100)
	args := generateMultipleHeaderProposalDetectorArgs()
	args.RoundHandler = &mock.RoundHandlerMock{RoundIndex: int64(round)}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	hData := slashMocks.CreateInterceptedHeaderData(&block.Header{Round: round + detector.MaxDeltaToCurrentRound + 1, RandSeed: []byte("seed")})
	res, err := sd.VerifyData(hData)

	require.Nil(t, res)
	require.Equal(t, process.ErrHeaderRoundNotRelevant, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_EmptyProposerList_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{}, nil
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	res, err := sd.VerifyData(slashMocks.CreateInterceptedHeaderData(&block.Header{}))
	require.Nil(t, res)
	require.Equal(t, process.ErrEmptyConsensusGroup, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_MultipleHeaders_SameHash_ExpectNoSlashing(t *testing.T) {
	t.Parallel()

	round := uint64(1)
	pubKey := []byte("proposer1")

	args := generateMultipleHeaderProposalDetectorArgs()
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock(pubKey)}, nil
		},
	}
	args.Cache = &slashMocks.RoundDetectorCacheStub{
		AddCalled: func(r uint64, pk []byte, header data.HeaderInfoHandler) error {
			if r == round && bytes.Equal(pk, pubKey) {
				return process.ErrHeadersNotDifferentHashes
			}
			return nil
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	hData := slashMocks.CreateInterceptedHeaderData(&block.HeaderV2{Header: &block.Header{Round: round, RandSeed: []byte("seed")}})
	res, err := sd.VerifyData(hData)
	require.Nil(t, res)
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_MultipleHeaders(t *testing.T) {
	t.Parallel()

	round := uint64(2)
	pubKey := []byte("proposer1")
	h1 := &block.HeaderV2{Header: &block.Header{Round: round, PrevRandSeed: []byte("seed1")}}
	h2 := &block.HeaderV2{Header: &block.Header{Round: round, PrevRandSeed: []byte("seed2")}}
	h3 := &block.HeaderV2{Header: &block.Header{Round: round, PrevRandSeed: []byte("seed3")}}
	h4 := &block.HeaderV2{Header: &block.Header{Round: round, PrevRandSeed: []byte("seed4")}}

	hData1 := slashMocks.CreateInterceptedHeaderData(h1)
	hData2 := slashMocks.CreateInterceptedHeaderData(h2)
	hData3 := slashMocks.CreateInterceptedHeaderData(h3)
	hData4 := slashMocks.CreateInterceptedHeaderData(h4)

	hInfo1 := &mockCoreData.HeaderInfoStub{Header: h1, Hash: []byte("h1")}
	hInfo2 := &mockCoreData.HeaderInfoStub{Header: h2, Hash: []byte("h2")}
	hInfo3 := &mockCoreData.HeaderInfoStub{Header: h3, Hash: []byte("h3")}
	hInfo4 := &mockCoreData.HeaderInfoStub{Header: h4, Hash: []byte("h4")}

	getCalledCt := 0
	addCalledCt := 0
	args := generateMultipleHeaderProposalDetectorArgs()
	args.Cache = &slashMocks.RoundDetectorCacheStub{
		AddCalled: func(_ uint64, _ []byte, header data.HeaderInfoHandler) error {
			addCalledCt++
			if bytes.Equal(header.GetHash(), hData2.Hash()) && addCalledCt == 3 {
				return process.ErrHeadersNotDifferentHashes
			}
			if addCalledCt > 5 {
				return process.ErrHeadersNotDifferentHashes
			}
			return nil
		},
		GetPubKeysCalled: func(_ uint64) [][]byte {
			return [][]byte{pubKey}
		},
		GetHeadersCalled: func(r uint64, pk []byte) []data.HeaderInfoHandler {
			getCalledCt++
			switch getCalledCt {
			case 1:
				return slash.HeaderInfoList{hInfo1}
			case 2:
				return slash.HeaderInfoList{hInfo1, hInfo2}
			case 3:
				return slash.HeaderInfoList{hInfo1, hInfo2, hInfo3}
			case 4:
				return slash.HeaderInfoList{hInfo1, hInfo2, hInfo3, hInfo4}
			default:
				return nil
			}
		},
	}
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock(pubKey)}, nil
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	tmp, err := sd.VerifyData(hData1)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	tmp, _ = sd.VerifyData(hData2)
	res := tmp.(coreSlash.MultipleProposalProofHandler)
	require.Equal(t, res.GetType(), coreSlash.MultipleProposal)
	require.Equal(t, res.GetLevel(), coreSlash.Medium)
	require.Len(t, res.GetHeaders(), 2)
	require.Equal(t, res.GetHeaders()[0], h1)
	require.Equal(t, res.GetHeaders()[1], h2)

	tmp, err = sd.VerifyData(hData2)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)

	tmp, _ = sd.VerifyData(hData3)
	res = tmp.(coreSlash.MultipleProposalProofHandler)
	require.Equal(t, res.GetType(), coreSlash.MultipleProposal)
	require.Equal(t, res.GetLevel(), coreSlash.High)
	require.Len(t, res.GetHeaders(), 3)
	require.Equal(t, res.GetHeaders()[0], h1)
	require.Equal(t, res.GetHeaders()[1], h2)
	require.Equal(t, res.GetHeaders()[2], h3)

	tmp, _ = sd.VerifyData(hData4)
	res = tmp.(coreSlash.MultipleProposalProofHandler)
	require.Equal(t, res.GetType(), coreSlash.MultipleProposal)
	require.Equal(t, res.GetLevel(), coreSlash.High)
	require.Len(t, res.GetHeaders(), 4)
	require.Equal(t, res.GetHeaders()[0], h1)
	require.Equal(t, res.GetHeaders()[1], h2)
	require.Equal(t, res.GetHeaders()[2], h3)
	require.Equal(t, res.GetHeaders()[3], h4)

	tmp, err = sd.VerifyData(hData4)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)
}

func TestMultipleHeaderProposalsDetector_ValidateProof_NilProof_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	err := sd.ValidateProof(nil)
	require.Equal(t, process.ErrNilProof, err)
}

func TestMultipleHeaderProposalsDetector_ValidateProof_InvalidProofType_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	proof1 := &slashMocks.MultipleHeaderSigningProofStub{
		GetTypeCalled: func() coreSlash.SlashingType {
			return coreSlash.MultipleProposal
		},
	}
	err := sd.ValidateProof(proof1)
	require.Equal(t, process.ErrCannotCastProofToMultipleProposedHeaders, err)

	proof2 := &slashMocks.MultipleHeaderProposalProofStub{
		GetTypeCalled: func() coreSlash.SlashingType {
			return coreSlash.MultipleSigning
		},
	}
	err = sd.ValidateProof(proof2)
	require.Equal(t, process.ErrInvalidSlashType, err)
}

func TestMultipleHeaderProposalsDetector_ValidateProof_DifferentSlashLevelsAndTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args        func() (coreSlash.ThreatLevel, slash.HeaderInfoList)
		expectedErr error
	}{
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				return coreSlash.Low, slash.HeaderInfoList{}
			},
			expectedErr: process.ErrInvalidSlashLevel,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				return coreSlash.ThreatLevel(44), slash.HeaderInfoList{}
			},
			expectedErr: process.ErrInvalidSlashLevel,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				return coreSlash.Medium, slash.HeaderInfoList{}
			},
			expectedErr: process.ErrNotEnoughHeadersProvided,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				h1 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 1}})
				h2 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 2}})
				h3 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 3}})
				return coreSlash.Medium, slash.HeaderInfoList{h1, h2, h3}
			},
			expectedErr: process.ErrSlashLevelDoesNotMatchSlashType,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				h1 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 1}})
				h2 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 2}})
				return coreSlash.High, slash.HeaderInfoList{h1, h2}
			},
			expectedErr: process.ErrSlashLevelDoesNotMatchSlashType,
		},
	}

	for _, currTest := range tests {
		args := generateMultipleHeaderProposalDetectorArgs()
		sd, _ := detector.NewMultipleHeaderProposalsDetector(args)
		level, headers := currTest.args()
		proof, _ := coreSlash.NewMultipleProposalProof(
			&coreSlash.SlashingResult{
				SlashingLevel: level,
				Headers:       headers,
			},
		)

		err := sd.ValidateProof(proof)
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestMultipleHeaderProposalsDetector_ValidateProof_DifferentHeaders(t *testing.T) {
	t.Parallel()

	errGetProposer := errors.New("cannot get proposer")
	nodesCoordinatorMock := &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(randomness []byte, round uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			if round == 0 && bytes.Equal(randomness, []byte("h1")) {
				return nil, errGetProposer
			}
			if round == 1 && bytes.Equal(randomness, []byte("h1")) {
				return []sharding.Validator{mock.NewValidatorMock([]byte("proposer1"))}, nil
			}
			if round == 1 && bytes.Equal(randomness, []byte("h2")) {
				return []sharding.Validator{mock.NewValidatorMock([]byte("proposer2"))}, nil
			}

			return []sharding.Validator{mock.NewValidatorMock([]byte("proposer"))}, nil
		},
	}
	tests := []struct {
		args        func() (coreSlash.ThreatLevel, slash.HeaderInfoList)
		expectedErr error
	}{
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				h1 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 5}})
				h2 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 5}})
				return coreSlash.Medium, slash.HeaderInfoList{h1, h2}
			},
			expectedErr: process.ErrHeadersNotDifferentHashes,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				h1 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 5}})
				h2 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 5}})
				h3 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 5}})
				return coreSlash.High, slash.HeaderInfoList{h1, h2, h3}
			},
			expectedErr: process.ErrHeadersNotDifferentHashes,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				h1 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 4}})
				h2 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 5}})
				return coreSlash.Medium, slash.HeaderInfoList{h1, h2}
			},
			expectedErr: process.ErrHeadersNotSameRound,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				h1 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 4, TimeStamp: 1}})
				h2 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 4, TimeStamp: 2}})
				h3 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 5, TimeStamp: 3}})
				return coreSlash.High, slash.HeaderInfoList{h1, h2, h3}
			},
			expectedErr: process.ErrHeadersNotSameRound,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				h1 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 0, PrevRandSeed: []byte("h1"), TimeStamp: 1}})
				h2 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 0, PrevRandSeed: []byte("h1"), TimeStamp: 2}})
				return coreSlash.Medium, slash.HeaderInfoList{h1, h2}
			},
			expectedErr: errGetProposer,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				h1 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 0, PrevRandSeed: []byte("h")}})
				h2 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 0, PrevRandSeed: []byte("h1")}})
				return coreSlash.Medium, slash.HeaderInfoList{h1, h2}
			},
			expectedErr: errGetProposer,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				h1 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 1, PrevRandSeed: []byte("h1")}})
				h2 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 1, PrevRandSeed: []byte("h2")}})
				return coreSlash.Medium, slash.HeaderInfoList{h1, h2}
			},
			expectedErr: process.ErrHeadersNotSameProposer,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				h1 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 4, TimeStamp: 1}})
				h2 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 4, TimeStamp: 2}})
				return coreSlash.Medium, slash.HeaderInfoList{h1, h2}
			},
			expectedErr: nil,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderInfoList) {
				h1 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 5, TimeStamp: 1}})
				h2 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 5, TimeStamp: 2}})
				h3 := slashMocks.CreateHeaderInfoData(&block.HeaderV2{Header: &block.Header{Round: 5, TimeStamp: 3}})
				return coreSlash.High, slash.HeaderInfoList{h1, h2, h3}
			},
			expectedErr: nil,
		},
	}

	args := generateMultipleHeaderProposalDetectorArgs()
	args.NodesCoordinator = nodesCoordinatorMock
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	for _, currTest := range tests {
		level, headers := currTest.args()
		proof, _ := coreSlash.NewMultipleProposalProof(
			&coreSlash.SlashingResult{
				SlashingLevel: level,
				Headers:       headers,
			},
		)
		err := sd.ValidateProof(proof)
		require.Equal(t, currTest.expectedErr, err)
	}
}

func generateMultipleHeaderProposalDetectorArgs() *detector.MultipleHeaderProposalDetectorArgs {
	return &detector.MultipleHeaderProposalDetectorArgs{
		NodesCoordinator: &mock.NodesCoordinatorMock{},
		RoundHandler:     &mock.RoundHandlerMock{},
		Cache:            &slashMocks.RoundDetectorCacheStub{},
		Hasher:           &hashingMocks.HasherMock{},
		Marshaller:       &testscommon.MarshalizerMock{},
	}
}
