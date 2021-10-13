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
		args        func() (sharding.NodesCoordinator, process.RoundHandler, detector.RoundDetectorCache)
		expectedErr error
	}{
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, detector.RoundDetectorCache) {
				return nil, &mock.RoundHandlerMock{}, &mockSlash.RoundDetectorCacheStub{}
			},
			expectedErr: process.ErrNilShardCoordinator,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, detector.RoundDetectorCache) {
				return &mock.NodesCoordinatorMock{}, nil, &mockSlash.RoundDetectorCacheStub{}
			},
			expectedErr: process.ErrNilRoundHandler,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, detector.RoundDetectorCache) {
				return &mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, nil
			},
			expectedErr: process.ErrNilRoundDetectorCache,
		},
		{
			args: func() (sharding.NodesCoordinator, process.RoundHandler, detector.RoundDetectorCache) {
				return &mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, &mockSlash.RoundDetectorCacheStub{}
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

	sd, _ := detector.NewMultipleHeaderProposalsDetector(&mock.NodesCoordinatorMock{}, &mock.RoundHandlerMock{}, &mockSlash.RoundDetectorCacheStub{})
	res, err := sd.VerifyData(&testscommon.InterceptedDataStub{})

	require.Nil(t, res)
	require.Equal(t, process.ErrCannotCastInterceptedDataToHeader, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_CannotGetProposer_ExpectError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("cannot get proposer")
	nodesCoordinator := &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return nil, expectedErr
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(nodesCoordinator, &mock.RoundHandlerMock{}, &mockSlash.RoundDetectorCacheStub{})

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
		&mockSlash.RoundDetectorCacheStub{})

	hData := testscommon.CreateInterceptedHeaderData(&block.Header{Round: round + detector.MaxDeltaToCurrentRound + 1, RandSeed: []byte("seed")})
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
	sd, _ := detector.NewMultipleHeaderProposalsDetector(&nodesCoordinator, &mock.RoundHandlerMock{}, &mockSlash.RoundDetectorCacheStub{})

	res, err := sd.VerifyData(&interceptedBlocks.InterceptedHeader{})
	require.Nil(t, res)
	require.Equal(t, process.ErrEmptyConsensusGroup, err)
}

func TestMultipleHeaderProposalsDetector_VerifyData_MultipleHeaders_SameHash_ExpectNoSlashing(t *testing.T) {
	t.Parallel()

	round := uint64(2)
	pubKey := []byte("proposer1")
	cache := mockSlash.RoundDetectorCacheStub{
		ContainsCalled: func(r uint64, pk []byte, data process.InterceptedData) bool {
			return r == round && bytes.Equal(pk, pubKey)
		},
	}
	nodesCoordinator := &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock(pubKey)}, nil
		},
	}
	sd, _ := detector.NewMultipleHeaderProposalsDetector(nodesCoordinator, &mock.RoundHandlerMock{}, &cache)

	hData := testscommon.CreateInterceptedHeaderData(&block.Header{Round: round, RandSeed: []byte("seed")})
	res, err := sd.VerifyData(hData)
	require.Nil(t, res)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	//3211111111111113213211111111111111111111111111111111111111111111111111111111111111111111111111111 poate aici verific si daca merge cu alt header
}

func TestMultipleHeaderProposalsDetector_VerifyData_MultipleHeaders(t *testing.T) {
	t.Parallel()

	round := uint64(2)
	pubKey := []byte("proposer1")
	hData1 := testscommon.CreateInterceptedHeaderData(&block.Header{Round: round, PrevRandSeed: []byte("seed1")})
	hData2 := testscommon.CreateInterceptedHeaderData(&block.Header{Round: round, PrevRandSeed: []byte("seed2")})
	hData3 := testscommon.CreateInterceptedHeaderData(&block.Header{Round: round, PrevRandSeed: []byte("seed3")})
	hData4 := testscommon.CreateInterceptedHeaderData(&block.Header{Round: round, PrevRandSeed: []byte("seed4")})
	getCalledCt := 0
	containsCalledCt := 0

	cache := mockSlash.RoundDetectorCacheStub{
		ContainsCalled: func(_ uint64, _ []byte, data process.InterceptedData) bool {
			containsCalledCt++
			if bytes.Equal(data.Hash(), hData2.Hash()) && containsCalledCt == 3 {
				return true
			}
			if containsCalledCt > 5 {
				return true
			}
			return false
		},
		GetPubKeysCalled: func(_ uint64) [][]byte {
			return [][]byte{pubKey}
		},
		GetDataCalled: func(r uint64, pk []byte) []process.InterceptedData {
			getCalledCt++
			switch getCalledCt {
			case 1:
				return []process.InterceptedData{hData1}
			case 2:
				return []process.InterceptedData{hData1, hData2}
			case 3:
				return []process.InterceptedData{hData1, hData2, hData3}
			case 4:
				return []process.InterceptedData{hData1, hData2, hData3, hData4}
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
	sd, _ := detector.NewMultipleHeaderProposalsDetector(nodesCoordinator, &mock.RoundHandlerMock{}, &cache)

	tmp, err := sd.VerifyData(hData1)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	tmp, _ = sd.VerifyData(hData2)
	res := tmp.(slash.MultipleProposalProofHandler)
	require.Equal(t, res.GetType(), slash.MultipleProposal)
	require.Equal(t, res.GetLevel(), slash.Medium)
	require.Len(t, res.GetHeaders(), 2)
	require.Equal(t, res.GetHeaders()[0], hData1)
	require.Equal(t, res.GetHeaders()[1], hData2)

	tmp, err = sd.VerifyData(hData2)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)

	tmp, _ = sd.VerifyData(hData3)
	res = tmp.(slash.MultipleProposalProofHandler)
	require.Equal(t, res.GetType(), slash.MultipleProposal)
	require.Equal(t, res.GetLevel(), slash.High)
	require.Len(t, res.GetHeaders(), 3)
	require.Equal(t, res.GetHeaders()[0], hData1)
	require.Equal(t, res.GetHeaders()[1], hData2)
	require.Equal(t, res.GetHeaders()[2], hData3)

	tmp, _ = sd.VerifyData(hData4)
	res = tmp.(slash.MultipleProposalProofHandler)
	require.Equal(t, res.GetType(), slash.MultipleProposal)
	require.Equal(t, res.GetLevel(), slash.High)
	require.Len(t, res.GetHeaders(), 4)
	require.Equal(t, res.GetHeaders()[0], hData1)
	require.Equal(t, res.GetHeaders()[1], hData2)
	require.Equal(t, res.GetHeaders()[2], hData3)
	require.Equal(t, res.GetHeaders()[3], hData4)

	tmp, err = sd.VerifyData(hData4)
	require.Nil(t, tmp)
	require.Equal(t, process.ErrNoSlashingEventDetected, err)
}

func TestMultipleHeaderProposalsDetector_ValidateProof_InvalidProofType_ExpectError(t *testing.T) {
	t.Parallel()

	sd, _ := detector.NewMultipleHeaderProposalsDetector(&mockEpochStart.NodesCoordinatorStub{}, &mock.RoundHandlerMock{}, &mockSlash.RoundDetectorCacheStub{})

	proof1, _ := slash.NewMultipleSigningProof(map[string]slash.SlashingResult{})
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
		args        func() (slash.ThreatLevel, []process.InterceptedData)
		expectedErr error
	}{
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.Low, []process.InterceptedData{}
			},
			expectedErr: process.ErrInvalidSlashLevel,
		},
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.ThreatLevel(44), []process.InterceptedData{}
			},
			expectedErr: process.ErrInvalidSlashLevel,
		},
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.Medium, []process.InterceptedData{}
			},
			expectedErr: process.ErrNotEnoughHeadersProvided,
		},
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.Medium, []process.InterceptedData{
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("h1")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("h2")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("h3")}),
				}
			},
			expectedErr: process.ErrSlashLevelDoesNotMatchSlashType,
		},
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.High, []process.InterceptedData{
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("h1")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 2, RandSeed: []byte("h2")}),
				}
			},
			expectedErr: process.ErrSlashLevelDoesNotMatchSlashType,
		},
	}

	for _, currTest := range tests {
		sd, _ := detector.NewMultipleHeaderProposalsDetector(&mockEpochStart.NodesCoordinatorStub{}, &mock.RoundHandlerMock{}, &mockSlash.RoundDetectorCacheStub{})
		level, data := currTest.args()
		proof, _ := slash.NewMultipleProposalProof(
			&slash.SlashingResult{
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
		args        func() (slash.ThreatLevel, []process.InterceptedData)
		expectedErr error
	}{
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.Medium, []process.InterceptedData{
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 5, PrevRandSeed: []byte("h1")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 5, PrevRandSeed: []byte("h1")}),
				}
			},
			expectedErr: process.ErrHeadersShouldHaveDifferentHashes,
		},
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.High, []process.InterceptedData{
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 5, PrevRandSeed: []byte("h1")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 5, PrevRandSeed: []byte("h2")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 5, PrevRandSeed: []byte("h2")}),
				}
			},
			expectedErr: process.ErrHeadersShouldHaveDifferentHashes,
		},
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.Medium, []process.InterceptedData{
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 4, PrevRandSeed: []byte("h1")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 5, PrevRandSeed: []byte("h1")}),
				}
			},
			expectedErr: process.ErrHeadersDoNotHaveSameRound,
		},
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.High, []process.InterceptedData{
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 4, PrevRandSeed: []byte("h1")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 4, PrevRandSeed: []byte("h2")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 5, PrevRandSeed: []byte("h3")}),
				}
			},
			expectedErr: process.ErrHeadersDoNotHaveSameRound,
		},
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.Medium, []process.InterceptedData{
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 0, PrevRandSeed: []byte("h1")}), // round ==0 && rndSeed == h1 => mock returns err
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 0, PrevRandSeed: []byte("h2")}),
				}
			},
			expectedErr: errGetProposer,
		},
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.Medium, []process.InterceptedData{
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 0, PrevRandSeed: []byte("h")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 0, PrevRandSeed: []byte("h1")}),
				}
			},
			expectedErr: errGetProposer,
		},
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.Medium, []process.InterceptedData{
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 1, PrevRandSeed: []byte("h1")}), // round == 1 && rndSeed == h1 => mock returns proposer1
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 1, PrevRandSeed: []byte("h2")}), // round == 1 && rndSeed == h2 => mock returns proposer2
				}
			},
			expectedErr: process.ErrHeadersDoNotHaveSameProposer,
		},
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.Medium, []process.InterceptedData{
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 4, PrevRandSeed: []byte("h1")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 4, PrevRandSeed: []byte("h2")}),
				}
			},
			expectedErr: nil,
		},
		{
			args: func() (slash.ThreatLevel, []process.InterceptedData) {
				return slash.High, []process.InterceptedData{
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 5, PrevRandSeed: []byte("h1")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 5, PrevRandSeed: []byte("h2")}),
					testscommon.CreateInterceptedHeaderData(&block.Header{Round: 5, PrevRandSeed: []byte("h3")}),
				}
			},
			expectedErr: nil,
		},
	}

	sd, _ := detector.NewMultipleHeaderProposalsDetector(nodesCoordinatorMock, &mock.RoundHandlerMock{}, &mockSlash.RoundDetectorCacheStub{})

	for _, currTest := range tests {
		level, data := currTest.args()
		proof, _ := slash.NewMultipleProposalProof(
			&slash.SlashingResult{
				SlashingLevel: level,
				Data:          data,
			},
		)
		err := sd.ValidateProof(proof)
		require.Equal(t, currTest.expectedErr, err)
	}
}
