package detector_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	mockEpochStart "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/process/slash/detector"
	mockSlash "github.com/ElrondNetwork/elrond-go/process/slash/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewMultipleHeaderProposalsDetector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args        func() (sharding.NodesCoordinator, process.RoundHandler, uint64)
		expectedErr error
	}{
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, uint64) {
				return nil, &mock.RoundHandlerMock{}, detector.CacheSize
			},
			expectedErr: process.ErrNilShardCoordinator,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, uint64) {
				return &mock.NodesCoordinatorMock{}, nil, detector.CacheSize
			},
			expectedErr: process.ErrNilRoundHandler,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, uint64) {
				return &mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, detector.CacheSize
			},
			expectedErr: nil,
		},
	}

	for _, currTest := range tests {
		_, err := detector.NewMultipleHeaderProposalsDetector(currTest.args())
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestMultipleHeaderProposalsDetector_VerifyData_CannotCastData_ExpectError(t *testing.T) {
	t.Parallel()

	sd, _ := detector.NewMultipleHeaderProposalsDetector(&mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, detector.CacheSize)
	res, err := sd.VerifyData(&testscommon.InterceptedDataStub{})

	require.Nil(t, res)
	require.Equal(t, process.ErrCannotCastInterceptedDataToHeader, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_CannotGetProposer_ExpectError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("cannot get proposer")
	sd, _ := detector.NewMultipleHeaderProposalsDetector(&mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return nil, expectedErr
		},
	}, &mock.RoundHandlerMock{}, detector.CacheSize)

	res, err := sd.VerifyData(&interceptedBlocks.InterceptedHeader{})

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
		detector.CacheSize)

	hData := createInterceptedHeaderData(&block.Header{Round: round + detector.MaxDeltaToCurrentRound + 1, RandSeed: []byte("seed")})
	res, err := sd.VerifyData(hData)

	require.Nil(t, res)
	require.Equal(t, process.ErrHeaderRoundNotRelevant, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_EmptyProposerList_ExpectError(t *testing.T) {
	t.Parallel()

	sd, _ := detector.NewMultipleHeaderProposalsDetector(
		&mockEpochStart.NodesCoordinatorStub{
			ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
				return []sharding.Validator{}, nil
			},
		}, &mock.RoundHandlerMock{}, detector.CacheSize)

	res, err := sd.VerifyData(&interceptedBlocks.InterceptedHeader{})

	require.Nil(t, res)
	require.Equal(t, process.ErrEmptyConsensusGroup, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_MultipleHeaders_SameHash_ExpectNoSlashing(t *testing.T) {
	t.Parallel()

	sd, _ := detector.NewMultipleHeaderProposalsDetector(&mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock([]byte("proposer1"))}, nil
		},
	}, &mock.RoundHandlerMock{}, detector.CacheSize)

	hData := createInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("seed")})
	res, err := sd.VerifyData(hData)
	require.Nil(t, res)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	res, err = sd.VerifyData(hData)
	require.Nil(t, res)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	res, err = sd.VerifyData(hData)
	require.Nil(t, res)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_MultipleHeaders_DifferentHashes_ExpectMultipleProposalSlashing(t *testing.T) {
	t.Parallel()

	sd, _ := detector.NewMultipleHeaderProposalsDetector(&mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock([]byte("proposer1"))}, nil
		},
	}, &mock.RoundHandlerMock{}, detector.CacheSize)

	hData1 := createInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("seed1")})
	tmp, err := sd.VerifyData(hData1)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	hData2 := createInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("seed2")})
	tmp, _ = sd.VerifyData(hData2)
	res := tmp.(slash.MultipleProposalProofHandler)
	require.Equal(t, res.GetType(), slash.MultipleProposal)
	require.Equal(t, res.GetLevel(), slash.Level1)
	require.Len(t, res.GetHeaders(), 2)
	require.Equal(t, res.GetHeaders()[0], hData1)
	require.Equal(t, res.GetHeaders()[1], hData2)

	hData3 := createInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("seed3")})
	tmp, _ = sd.VerifyData(hData3)
	res = tmp.(slash.MultipleProposalProofHandler)
	require.Equal(t, res.GetType(), slash.MultipleProposal)
	require.Equal(t, res.GetLevel(), slash.Level2)
	require.Len(t, res.GetHeaders(), 3)
	require.Equal(t, res.GetHeaders()[0], hData1)
	require.Equal(t, res.GetHeaders()[1], hData2)
	require.Equal(t, res.GetHeaders()[2], hData3)

	hData4 := createInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("seed4")})
	tmp, _ = sd.VerifyData(hData4)
	res = tmp.(slash.MultipleProposalProofHandler)
	require.Equal(t, res.GetType(), slash.MultipleProposal)
	require.Equal(t, res.GetLevel(), slash.Level2)
	require.Len(t, res.GetHeaders(), 4)
	require.Equal(t, res.GetHeaders()[0], hData1)
	require.Equal(t, res.GetHeaders()[1], hData2)
	require.Equal(t, res.GetHeaders()[2], hData3)
	require.Equal(t, res.GetHeaders()[3], hData4)
}

func TestMultipleHeaderProposalsDetector_ValidateProof_InvalidProofType_ExpectError(t *testing.T) {
	t.Parallel()

	sd, _ := detector.NewMultipleHeaderProposalsDetector(&mockEpochStart.NodesCoordinatorStub{}, &mock.RoundHandlerMock{}, detector.CacheSize)

	proof1, _ := slash.NewMultipleSigningProof(map[string]slash.SlashingData{})
	err := sd.ValidateProof(proof1)
	require.Equal(t, process.ErrCannotCastProofToMultipleProposedHeaders, err)

	proof2 := &mockSlash.MultipleHeaderProposalProofStub{
		GetTypeCalled: func() slash.SlashingType {
			return slash.MultipleSigning
		},
	}
	err = sd.ValidateProof(proof2)
	require.Equal(t, process.ErrInvalidSlashType, err)
}

func TestMultipleHeaderProposalsDetector_ValidateProof_MultipleProposalProof_DifferentSlashLevelsAndTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args        func() (slash.SlashingLevel, []process.InterceptedData)
		expectedErr error
	}{
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level0, []process.InterceptedData{}
			},
			expectedErr: process.ErrInvalidSlashLevel,
		},
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return 4444, []process.InterceptedData{}
			},
			expectedErr: process.ErrInvalidSlashLevel,
		},
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level1, []process.InterceptedData{}
			},
			expectedErr: process.ErrNotEnoughHeadersProvided,
		},
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level1, []process.InterceptedData{
					createInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("h1")}),
					createInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("h2")}),
					createInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("h3")}),
				}
			},
			expectedErr: process.ErrSlashLevelDoesNotMatchSlashType,
		},
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level2, []process.InterceptedData{
					createInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("h1")}),
					createInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("h2")}),
				}
			},
			expectedErr: process.ErrSlashLevelDoesNotMatchSlashType,
		},
	}

	for _, currTest := range tests {
		sd, _ := detector.NewMultipleHeaderProposalsDetector(&mockEpochStart.NodesCoordinatorStub{}, &mock.RoundHandlerMock{}, detector.CacheSize)
		level, data := currTest.args()
		proof, _ := slash.NewMultipleProposalProof(
			slash.SlashingData{
				SlashingLevel: level,
				Data:          data,
			},
		)

		err := sd.ValidateProof(proof)
		require.Equal(t, currTest.expectedErr, err)
	}
}

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
		args        func() (slash.SlashingLevel, []process.InterceptedData)
		expectedErr error
	}{
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level1, []process.InterceptedData{
					createInterceptedHeaderData(&block.Header{Round: 5, RandSeed: []byte("h1")}),
					createInterceptedHeaderData(&block.Header{Round: 5, RandSeed: []byte("h1")}),
				}
			},
			expectedErr: process.ErrProposedHeadersDoNotHaveDifferentHashes,
		},
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level2, []process.InterceptedData{
					createInterceptedHeaderData(&block.Header{Round: 5, RandSeed: []byte("h1")}),
					createInterceptedHeaderData(&block.Header{Round: 5, RandSeed: []byte("h2")}),
					createInterceptedHeaderData(&block.Header{Round: 5, RandSeed: []byte("h2")}),
				}
			},
			expectedErr: process.ErrProposedHeadersDoNotHaveDifferentHashes,
		},
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level1, []process.InterceptedData{
					createInterceptedHeaderData(&block.Header{Round: 4, RandSeed: []byte("h1")}),
					createInterceptedHeaderData(&block.Header{Round: 5, RandSeed: []byte("h1")}),
				}
			},
			expectedErr: process.ErrHeadersDoNotHaveSameRound,
		},
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level2, []process.InterceptedData{
					createInterceptedHeaderData(&block.Header{Round: 4, RandSeed: []byte("h1")}),
					createInterceptedHeaderData(&block.Header{Round: 4, RandSeed: []byte("h2")}),
					createInterceptedHeaderData(&block.Header{Round: 5, RandSeed: []byte("h3")}),
				}
			},
			expectedErr: process.ErrHeadersDoNotHaveSameRound,
		},
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level1, []process.InterceptedData{
					createInterceptedHeaderData(&block.Header{Round: 0, RandSeed: []byte("h1")}), // round ==0 && rndSeed == h1 => mock returns err
					createInterceptedHeaderData(&block.Header{Round: 0, RandSeed: []byte("h2")}),
				}
			},
			expectedErr: errGetProposer,
		},
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level1, []process.InterceptedData{
					createInterceptedHeaderData(&block.Header{Round: 0, RandSeed: []byte("h")}),
					createInterceptedHeaderData(&block.Header{Round: 0, RandSeed: []byte("h1")}),
				}
			},
			expectedErr: errGetProposer,
		},
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level1, []process.InterceptedData{
					createInterceptedHeaderData(&block.Header{Round: 1, RandSeed: []byte("h1")}), // round == 1 && rndSeed == h1 => mock returns proposer1
					createInterceptedHeaderData(&block.Header{Round: 1, RandSeed: []byte("h2")}), // round == 1 && rndSeed == h2 => mock returns proposer2
				}
			},
			expectedErr: process.ErrHeadersDoNotHaveSameProposer,
		},
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level1, []process.InterceptedData{
					createInterceptedHeaderData(&block.Header{Round: 4, RandSeed: []byte("h1")}),
					createInterceptedHeaderData(&block.Header{Round: 4, RandSeed: []byte("h2")}),
				}
			},
			expectedErr: nil,
		},
		{
			args: func() (slash.SlashingLevel, []process.InterceptedData) {
				return slash.Level2, []process.InterceptedData{
					createInterceptedHeaderData(&block.Header{Round: 5, RandSeed: []byte("h1")}),
					createInterceptedHeaderData(&block.Header{Round: 5, RandSeed: []byte("h2")}),
					createInterceptedHeaderData(&block.Header{Round: 5, RandSeed: []byte("h3")}),
				}
			},
			expectedErr: nil,
		},
	}

	sd, _ := detector.NewMultipleHeaderProposalsDetector(nodesCoordinatorMock, &mock.RoundHandlerMock{}, detector.CacheSize)

	for _, currTest := range tests {
		level, data := currTest.args()
		proof, _ := slash.NewMultipleProposalProof(
			slash.SlashingData{
				SlashingLevel: level,
				Data:          data,
			},
		)
		err := sd.ValidateProof(proof)
		require.Equal(t, currTest.expectedErr, err)
	}
}

func createInterceptedHeaderArg(header *block.Header) *interceptedBlocks.ArgInterceptedBlockHeader {
	args := &interceptedBlocks.ArgInterceptedBlockHeader{
		ShardCoordinator:        &mock.ShardCoordinatorStub{},
		Hasher:                  &mock.HasherMock{},
		Marshalizer:             &mock.MarshalizerMock{},
		HeaderSigVerifier:       &mock.HeaderSigVerifierStub{},
		HeaderIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		ValidityAttester:        &mock.ValidityAttesterStub{},
		EpochStartTrigger:       &mock.EpochStartTriggerStub{},
	}

	args.HdrBuff, _ = args.Marshalizer.Marshal(header)

	return args
}

func createInterceptedHeaderData(header *block.Header) *interceptedBlocks.InterceptedHeader {
	args := createInterceptedHeaderArg(header)
	interceptedHeader, _ := interceptedBlocks.NewInterceptedHeader(args)

	return interceptedHeader
}
