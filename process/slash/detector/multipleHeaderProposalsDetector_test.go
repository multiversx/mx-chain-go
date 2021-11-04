package detector_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
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
		args        func() (sharding.NodesCoordinator, process.RoundHandler, detector.RoundDetectorCache, hashing.Hasher, marshal.Marshalizer)
		expectedErr error
	}{
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, detector.RoundDetectorCache, hashing.Hasher, marshal.Marshalizer) {
				return nil, &mock.RoundHandlerMock{}, &slashMocks.RoundDetectorCacheStub{}, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{}
			},
			expectedErr: process.ErrNilShardCoordinator,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, detector.RoundDetectorCache, hashing.Hasher, marshal.Marshalizer) {
				return &mock.NodesCoordinatorMock{}, nil, &slashMocks.RoundDetectorCacheStub{}, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{}
			},
			expectedErr: process.ErrNilRoundHandler,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, detector.RoundDetectorCache, hashing.Hasher, marshal.Marshalizer) {
				return &mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, nil, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{}
			},
			expectedErr: process.ErrNilRoundDetectorCache,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, detector.RoundDetectorCache, hashing.Hasher, marshal.Marshalizer) {
				return &mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, &slashMocks.RoundDetectorCacheStub{}, nil, &testscommon.MarshalizerMock{}
			},
			expectedErr: process.ErrNilHasher,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, detector.RoundDetectorCache, hashing.Hasher, marshal.Marshalizer) {
				return &mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, &slashMocks.RoundDetectorCacheStub{}, &hashingMocks.HasherMock{}, nil
			},
			expectedErr: process.ErrNilMarshalizer,
		},
	}

	for _, currTest := range tests {
		_, err := detector.NewMultipleHeaderProposalsDetector(currTest.args())
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestMultipleHeaderProposalsDetector_VerifyData_CannotCastData_ExpectError(t *testing.T) {
	t.Parallel()

	sd, _ := detector.NewMultipleHeaderProposalsDetector(&mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, &slashMocks.RoundDetectorCacheStub{}, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})
	res, err := sd.VerifyData(&testscommon.InterceptedDataStub{})

	require.Nil(t, res)
	require.Equal(t, process.ErrCannotCastInterceptedDataToHeader, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_NilHeaderHandler_ExpectError(t *testing.T) {
	t.Parallel()

	sd, _ := detector.NewMultipleHeaderProposalsDetector(&mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, &slashMocks.RoundDetectorCacheStub{}, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})
	res, err := sd.VerifyData(&interceptedBlocks.InterceptedHeader{})

	require.Nil(t, res)
	require.Equal(t, process.ErrNilHeaderHandler, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_CannotGetProposer_ExpectError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("cannot get proposer")
	nodesCoordinator := &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return nil, expectedErr
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(nodesCoordinator, &mock.RoundHandlerMock{}, &slashMocks.RoundDetectorCacheStub{}, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})

	res, err := sd.VerifyData(slashMocks.CreateInterceptedHeaderData(&block.Header{}))

	require.Nil(t, res)
	require.Equal(t, expectedErr, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_IrrelevantRound_ExpectError(t *testing.T) {
	t.Parallel()

	round := uint64(100)
	sd, _ := detector.NewMultipleHeaderProposalsDetector(
		&mockEpochStart.NodesCoordinatorStub{},
		&mock.RoundHandlerMock{
			RoundIndex: int64(round),
		},
		&slashMocks.RoundDetectorCacheStub{}, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})

	hData := slashMocks.CreateInterceptedHeaderData(&block.Header{Round: round + detector.MaxDeltaToCurrentRound + 1, RandSeed: []byte("seed")})
	res, err := sd.VerifyData(hData)

	require.Nil(t, res)
	require.Equal(t, process.ErrHeaderRoundNotRelevant, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_EmptyProposerList_ExpectError(t *testing.T) {
	t.Parallel()

	nodesCoordinator := mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{}, nil
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(&nodesCoordinator, &mock.RoundHandlerMock{}, &slashMocks.RoundDetectorCacheStub{}, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})

	res, err := sd.VerifyData(slashMocks.CreateInterceptedHeaderData(&block.Header{}))
	require.Nil(t, res)
	require.Equal(t, process.ErrEmptyConsensusGroup, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_MultipleHeaders_SameHash_ExpectNoSlashing(t *testing.T) {
	t.Parallel()

	round := uint64(1)
	pubKey := []byte("proposer1")
	cache := slashMocks.RoundDetectorCacheStub{
		AddCalled: func(r uint64, pk []byte, header data.HeaderInfoHandler) error {
			if r == round && bytes.Equal(pk, pubKey) {
				return process.ErrHeadersNotDifferentHashes
			}
			return nil
		},
	}
	nodesCoordinator := &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock(pubKey)}, nil
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(nodesCoordinator, &mock.RoundHandlerMock{}, &cache, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})

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

	getCalledCt := 0
	addCalledCt := 0

	cache := slashMocks.RoundDetectorCacheStub{
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
		GetDataCalled: func(r uint64, pk []byte) []data.HeaderHandler {
			getCalledCt++
			switch getCalledCt {
			case 1:
				return slash.HeaderList{h1}
			case 2:
				return slash.HeaderList{h1, h2}
			case 3:
				return slash.HeaderList{h1, h2, h3}
			case 4:
				return slash.HeaderList{h1, h2, h3, h4}
			default:
				return nil
			}
		},
	}
	nodesCoordinator := &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock(pubKey)}, nil
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(nodesCoordinator, &mock.RoundHandlerMock{}, &cache, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})

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

func TestMultipleHeaderProposalsDetector_ValidateProof_InvalidProofType_ExpectError(t *testing.T) {
	t.Parallel()

	sd, _ := detector.NewMultipleHeaderProposalsDetector(&mockEpochStart.NodesCoordinatorStub{}, &mock.RoundHandlerMock{}, &slashMocks.RoundDetectorCacheStub{}, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})

	proof1, _ := coreSlash.NewMultipleSigningProof(map[string]coreSlash.SlashingResult{})
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

func TestMultipleHeaderProposalsDetector_ValidateProof_MultipleProposalProof_DifferentSlashLevelsAndTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args        func() (coreSlash.ThreatLevel, slash.HeaderList)
		expectedErr error
	}{
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				return coreSlash.Low, slash.HeaderList{}
			},
			expectedErr: process.ErrInvalidSlashLevel,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				return coreSlash.ThreatLevel(44), slash.HeaderList{}
			},
			expectedErr: process.ErrInvalidSlashLevel,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				return coreSlash.Medium, slash.HeaderList{}
			},
			expectedErr: process.ErrNotEnoughHeadersProvided,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 1}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 2}}
				h3 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 3}}

				return coreSlash.Medium, slash.HeaderList{h1, h2, h3}
			},
			expectedErr: process.ErrSlashLevelDoesNotMatchSlashType,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 1}}
				h2 := &block.HeaderV2{Header: &block.Header{Round: 2, Nonce: 2}}
				return coreSlash.High, slash.HeaderList{h1, h2}
			},
			expectedErr: process.ErrSlashLevelDoesNotMatchSlashType,
		},
	}

	for _, currTest := range tests {
		sd, _ := detector.NewMultipleHeaderProposalsDetector(&mockEpochStart.NodesCoordinatorStub{}, &mock.RoundHandlerMock{}, &slashMocks.RoundDetectorCacheStub{}, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})
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

// TODO: AICI E FOARTE DELICATA SITUATIA, TREBUIE MOCKUIT SI HASHERUL
func TestMultipleHeaderProposalsDetector_ValidateProof_MultipleProposalProof_DifferentHeaders(t *testing.T) {
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
		args        func() (coreSlash.ThreatLevel, slash.HeaderList)
		expectedErr error
	}{
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 5}} //, []byte("h1"), )
				h2 := &block.HeaderV2{Header: &block.Header{Round: 5}} //, []byte("h1"))
				return coreSlash.Medium, slash.HeaderList{h1, h2}
			},
			expectedErr: process.ErrHeadersNotDifferentHashes,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 5}} //, []byte("h1"))
				h2 := &block.HeaderV2{Header: &block.Header{Round: 5}} //, []byte("h2"))
				h3 := &block.HeaderV2{Header: &block.Header{Round: 5}} //, []byte("h2"))
				return coreSlash.High, slash.HeaderList{h1, h2, h3}
			},
			expectedErr: process.ErrHeadersNotDifferentHashes,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 4}} //, []byte("h1"), )
				h2 := &block.HeaderV2{Header: &block.Header{Round: 5}} //, []byte("h2"))
				return coreSlash.Medium, slash.HeaderList{h1, h2}
			},
			expectedErr: process.ErrHeadersNotSameRound,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 4}} //, []byte("h1"))
				h2 := &block.HeaderV2{Header: &block.Header{Round: 4}} //, []byte("h2"))
				h3 := &block.HeaderV2{Header: &block.Header{Round: 5}} //, []byte("h3"))
				return coreSlash.High, slash.HeaderList{h1, h2, h3}
			},
			expectedErr: process.ErrHeadersNotSameRound,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 0, PrevRandSeed: []byte("h1")}} //, []byte("h1"), )
				h2 := &block.HeaderV2{Header: &block.Header{Round: 0, PrevRandSeed: []byte("h1")}} //, []byte("h2"))
				return coreSlash.Medium, slash.HeaderList{h1, h2}
			},
			expectedErr: errGetProposer,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 0, PrevRandSeed: []byte("h")}}  //, []byte("h"), )
				h2 := &block.HeaderV2{Header: &block.Header{Round: 0, PrevRandSeed: []byte("h1")}} //, []byte("h1"))
				return coreSlash.Medium, slash.HeaderList{h1, h2}
			},
			expectedErr: errGetProposer,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 1, PrevRandSeed: []byte("h1")}} //, []byte("h1"), )
				h2 := &block.HeaderV2{Header: &block.Header{Round: 1, PrevRandSeed: []byte("h2")}} //, []byte("h2"))
				return coreSlash.Medium, slash.HeaderList{h1, h2}
			},
			expectedErr: process.ErrHeadersNotSameProposer,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 4}} //, []byte("h1"), )
				h2 := &block.HeaderV2{Header: &block.Header{Round: 4}} //, []byte("h2"))
				return coreSlash.Medium, slash.HeaderList{h1, h2}
			},
			expectedErr: nil,
		},
		{
			args: func() (coreSlash.ThreatLevel, slash.HeaderList) {
				h1 := &block.HeaderV2{Header: &block.Header{Round: 5}} //, []byte("h1"), )
				h2 := &block.HeaderV2{Header: &block.Header{Round: 5}} //, []byte("h2"))
				h3 := &block.HeaderV2{Header: &block.Header{Round: 5}} //, []byte("h3"))
				return coreSlash.High, slash.HeaderList{h1, h2, h3}
			},
			expectedErr: nil,
		},
	}

	sd, _ := detector.NewMultipleHeaderProposalsDetector(nodesCoordinatorMock, &mock.RoundHandlerMock{}, &slashMocks.RoundDetectorCacheStub{}, &hashingMocks.HasherMock{}, &testscommon.MarshalizerMock{})

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
