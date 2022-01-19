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
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/process/slash/detector"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/slashMocks"
	"github.com/stretchr/testify/require"
)

// NewMultipleHeaderProposalsDetector ---
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
			expectedErr: process.ErrNilNodesCoordinator,
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
				args.SlashingCache = nil
				return args
			},
			expectedErr: process.ErrNilRoundValidatorHeadersCache,
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
		{
			args: func() *detector.MultipleHeaderProposalDetectorArgs {
				args := generateMultipleHeaderProposalDetectorArgs()
				args.HeaderSigVerifier = nil
				return args
			},
			expectedErr: process.ErrNilHeaderSigVerifier,
		},
	}

	for _, currTest := range tests {
		_, err := detector.NewMultipleHeaderProposalsDetector(currTest.args())
		require.Equal(t, currTest.expectedErr, err)
	}
}

// VerifyData ---
func TestMultipleHeaderProposalsDetector_VerifyData_NilData_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	res, err := sd.VerifyData(nil)
	require.Nil(t, res)
	require.Equal(t, process.ErrNilInterceptedData, err)
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
	args.RoundHandler = &testscommon.RoundHandlerMock{
		IndexCalled: func() int64 {
			return int64(round)
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	hData := slashMocks.CreateInterceptedHeaderData(&block.Header{Round: round + detector.MaxDeltaToCurrentRound + 1})
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
	args.SlashingCache = &slashMocks.RoundDetectorCacheStub{
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

func TestMultipleHeaderProposalsDetector_VerifyData_MultipleProposedHeadersSameRound(t *testing.T) {
	t.Parallel()

	round := uint64(2)
	shardID := uint32(1)
	pubKey := []byte("proposer1")
	h1 := &block.HeaderV2{Header: &block.Header{Round: round, ShardID: shardID, PrevRandSeed: []byte("seed1")}}
	h2 := &block.HeaderV2{Header: &block.Header{Round: round, ShardID: shardID, PrevRandSeed: []byte("seed2")}}
	h3 := &block.HeaderV2{Header: &block.Header{Round: round, ShardID: shardID, PrevRandSeed: []byte("seed3")}}
	h4 := &block.HeaderV2{Header: &block.Header{Round: round, ShardID: shardID, PrevRandSeed: []byte("seed4")}}

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
	args.SlashingCache = &slashMocks.RoundDetectorCacheStub{
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

	expectedProofTxData := &coreSlash.ProofTxData{
		Round:   round,
		ShardID: shardID,
		ProofID: coreSlash.MultipleProposalProofID,
	}
	tmp, err := sd.VerifyData(hData1)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	tmp, _ = sd.VerifyData(hData2)
	res := tmp.(coreSlash.MultipleProposalProofHandler)
	proofTxData, _ := res.GetProofTxData()
	require.Equal(t, expectedProofTxData, proofTxData)
	require.Equal(t, res.GetLevel(), coreSlash.Medium)
	require.Len(t, res.GetHeaders(), 2)
	require.Equal(t, res.GetHeaders()[0], h1)
	require.Equal(t, res.GetHeaders()[1], h2)

	tmp, err = sd.VerifyData(hData2)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)

	tmp, _ = sd.VerifyData(hData3)
	res = tmp.(coreSlash.MultipleProposalProofHandler)
	proofTxData, _ = res.GetProofTxData()
	require.Equal(t, expectedProofTxData, proofTxData)
	require.Equal(t, res.GetLevel(), coreSlash.High)
	require.Len(t, res.GetHeaders(), 3)
	require.Equal(t, res.GetHeaders()[0], h1)
	require.Equal(t, res.GetHeaders()[1], h2)
	require.Equal(t, res.GetHeaders()[2], h3)

	tmp, _ = sd.VerifyData(hData4)
	res = tmp.(coreSlash.MultipleProposalProofHandler)
	proofTxData, _ = res.GetProofTxData()
	require.Equal(t, expectedProofTxData, proofTxData)
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

// ValidateProof
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

	proof := &slashMocks.MultipleHeaderSigningProofStub{
		GetProofTxDataCalled: func() (*coreSlash.ProofTxData, error) {
			return &coreSlash.ProofTxData{
				ProofID: coreSlash.MultipleProposalProofID,
			}, nil
		},
	}
	err := sd.ValidateProof(proof)
	require.Equal(t, process.ErrCannotCastProofToMultipleProposedHeaders, err)
}

func TestMultipleHeaderProposalsDetector_ValidateProof_InvalidThreatLevel(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	proof := &slashMocks.MultipleHeaderProposalProofStub{
		GetLevelCalled: func() coreSlash.ThreatLevel {
			return coreSlash.Zero
		},
	}
	err := sd.ValidateProof(proof)
	require.Equal(t, process.ErrInvalidSlashLevel, err)
}

func TestMultipleHeaderProposalsDetector_ValidateProof_DifferentHeaders(t *testing.T) {
	t.Parallel()

	nodesCoordinatorMock := &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(randomness []byte, round uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
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
		args        func() (coreSlash.ThreatLevel, slash.HeaderList)
		expectedErr error
	}{
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 5}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 5}}
				return coreSlash.Medium, slash.HeaderList{h1, h2}
			},
			expectedErr: process.ErrHeadersNotDifferentHashes,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 5}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 5}}
				h3 := &block.HeaderV2{Header: &block.Header{Round: 5}}
				return coreSlash.High, slash.HeaderList{h1, h2, h3}
			},
			expectedErr: process.ErrHeadersNotDifferentHashes,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 5, TimeStamp: 1}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 5, TimeStamp: 2}}
				h3 := &block.HeaderV2{Header: &block.Header{Round: 5, TimeStamp: 1}}
				return coreSlash.High, slash.HeaderList{h1, h2, h3}
			},
			expectedErr: process.ErrHeadersNotDifferentHashes,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 4}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 5}}
				return coreSlash.Medium, slash.HeaderList{h1, h2}
			},
			expectedErr: process.ErrHeadersNotSameRound,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 4, TimeStamp: 1}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 4, TimeStamp: 2}}
				h3 := &block.HeaderV2{Header: &block.Header{Round: 5, TimeStamp: 3}}
				return coreSlash.High, slash.HeaderList{h1, h2, h3}
			},
			expectedErr: process.ErrHeadersNotSameRound,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 1, PrevRandSeed: []byte("h1")}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 1, PrevRandSeed: []byte("h2")}}
				return coreSlash.Medium, slash.HeaderList{h1, h2}
			},
			expectedErr: process.ErrHeadersNotSameProposer,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 4, TimeStamp: 1}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 4, TimeStamp: 2}}
				return coreSlash.Medium, slash.HeaderList{h1, h2}
			},
			expectedErr: nil,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 5, TimeStamp: 1}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 5, TimeStamp: 2}}
				h3 := &block.HeaderV2{Header: &block.Header{Round: 5, TimeStamp: 3}}
				return coreSlash.High, slash.HeaderList{h1, h2, h3}
			},
			expectedErr: nil,
		},
	}

	args := generateMultipleHeaderProposalDetectorArgs()
	args.NodesCoordinator = nodesCoordinatorMock
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	for _, currTest := range tests {
		level, headers := currTest.args()
		proof := &slashMocks.MultipleHeaderProposalProofStub{
			GetLevelCalled: func() coreSlash.ThreatLevel {
				return level
			},
			GetHeadersCalled: func() []data.HeaderHandler {
				return headers
			},
		}
		err := sd.ValidateProof(proof)
		require.Equal(t, currTest.expectedErr, err)
	}
}

// CheckProposedHeaders ---
func TestMultipleHeaderProposalsDetector_CheckProposedHeaders_NotEnoughHeaders_ExpectErr(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	err := sd.CheckProposedHeaders([]data.HeaderHandler{})
	require.Equal(t, process.ErrNotEnoughHeadersProvided, err)
}

func TestMultipleHeaderProposalsDetector_CheckProposedHeaders_NilHeaders_ExpectErr(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	header1 := data.HeaderHandler(nil)
	header2 := data.HeaderHandler(nil)
	header3 := data.HeaderHandler(nil)

	// All headers nil
	headers := []data.HeaderHandler{header1, header2, header3}
	err := sd.CheckProposedHeaders(headers)
	require.Equal(t, process.ErrNilHeaderHandler, err)

	// First header valid, second and third headers nil
	header1 = &block.Header{Round: 1, TimeStamp: 1}
	headers = []data.HeaderHandler{header1, header2, header3}
	err = sd.CheckProposedHeaders(headers)
	require.Equal(t, process.ErrNilHeaderHandler, err)

	// First and second header valid, third header nil
	header2 = &block.Header{Round: 1, TimeStamp: 2}
	headers = []data.HeaderHandler{header1, header2, header3}
	err = sd.CheckProposedHeaders(headers)
	require.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestMultipleHeaderProposalsDetector_CheckProposedHeaders_CannotGetProposer_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	errGetProposer := errors.New("cannot get proposer")
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return nil, errGetProposer
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	header1 := &block.Header{Round: 1, TimeStamp: 1}
	header2 := &block.Header{Round: 1, TimeStamp: 2}
	headers := []data.HeaderHandler{header1, header2}
	err := sd.CheckProposedHeaders(headers)
	require.Equal(t, errGetProposer, err)
}

func TestMultipleHeaderProposalsDetector_CheckProposedHeaders_CannotComputeHash_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	errMarshaller := errors.New("error marshaller")
	args.Marshaller = &testscommon.MarshalizerStub{
		MarshalCalled: func(interface{}) ([]byte, error) {
			return nil, errMarshaller
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	headers := []data.HeaderHandler{&block.Header{}, &block.Header{}}
	err := sd.CheckProposedHeaders(headers)
	require.Equal(t, errMarshaller, err)
}

func TestMultipleHeaderProposalsDetector_CheckProposedHeaders_HeadersSameHash_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	header1 := &block.Header{Round: 1}
	header2 := &block.Header{Round: 1}
	headers := []data.HeaderHandler{header1, header2}
	err := sd.CheckProposedHeaders(headers)
	require.Equal(t, process.ErrHeadersNotDifferentHashes, err)
}

func TestMultipleHeaderProposalsDetector_CheckProposedHeaders_ErrHeadersNotSameRound_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	header1 := &block.Header{Round: 1}
	header2 := &block.Header{Round: 2}
	headers := []data.HeaderHandler{header1, header2}
	err := sd.CheckProposedHeaders(headers)
	require.Equal(t, process.ErrHeadersNotSameRound, err)
}

// CheckHeaderHasSameProposerAndRound ---
func TestMultipleHeaderProposalsDetector_CheckHeaderHasSameProposerAndRound_CannotGetProposer_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	errGetProposer := errors.New("cannot get proposer")
	args.NodesCoordinator = &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return nil, errGetProposer
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	header := &block.Header{Round: 1}
	err := sd.CheckHeaderHasSameProposerAndRound(header, 1, validatorPubKey)
	require.Equal(t, errGetProposer, err)
}

func TestMultipleHeaderProposalsDetector_CheckHeaderHasSameProposerAndRound_NotSameRound_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	header := &block.Header{Round: 100}
	err := sd.CheckHeaderHasSameProposerAndRound(header, 99, validatorPubKey)
	require.Equal(t, process.ErrHeadersNotSameRound, err)
}

func TestMultipleHeaderProposalsDetector_CheckHeaderHasSameProposerAndRound_NotSameProposer_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	header := &block.Header{Round: 1}
	err := sd.CheckHeaderHasSameProposerAndRound(header, 1, []byte("proposer2"))
	require.Equal(t, process.ErrHeadersNotSameProposer, err)
}

func TestMultipleHeaderProposalsDetector_CheckHeaderHasSameProposerAndRound_LeaderDidNotSignHeader_ExpectError(t *testing.T) {
	t.Parallel()

	args := generateMultipleHeaderProposalDetectorArgs()
	errSignature := errors.New("leader did not sign this header")
	args.HeaderSigVerifier = &mock.HeaderSigVerifierStub{
		VerifyLeaderSignatureCalled: func(data.HeaderHandler) error {
			return errSignature
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(args)

	header := &block.Header{Round: 1}
	err := sd.CheckHeaderHasSameProposerAndRound(header, 1, validatorPubKey)
	require.Equal(t, errSignature, err)
}

func generateMultipleHeaderProposalDetectorArgs() *detector.MultipleHeaderProposalDetectorArgs {
	return &detector.MultipleHeaderProposalDetectorArgs{
		MultipleHeaderDetectorArgs: generateMockMultipleHeaderDetectorArgs(),
	}
}
