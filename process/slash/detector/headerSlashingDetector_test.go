package detector_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	mock2 "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/process/slash/detector"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewHeaderSlashingDetector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args        func() sharding.NodesCoordinator
		expectedErr error
	}{
		{
			args: func() sharding.NodesCoordinator {
				return nil
			},
			expectedErr: process.ErrNilShardCoordinator,
		},
		{
			args: func() sharding.NodesCoordinator {
				return &mock.NodesCoordinatorMock{}
			},
			expectedErr: nil,
		},
	}

	for _, currTest := range tests {
		_, err := detector.NewHeaderSlashingDetector(currTest.args())
		require.Equal(t, currTest.expectedErr, err)
	}
}

func TestHeaderSlashingDetector_VerifyData_CannotCastData_ExpectError(t *testing.T) {
	sd, _ := detector.NewHeaderSlashingDetector(&mock.NodesCoordinatorMock{})

	res, err := sd.VerifyData(&testscommon.InterceptedDataStub{})

	require.Nil(t, res)
	require.Equal(t, process.ErrCannotCastInterceptedDataToHeader, err)
}

func TestHeaderSlashingDetector_VerifyData_CannotGetProposer_ExpectError(t *testing.T) {
	expectedErr := errors.New("cannot get proposer")
	sd, _ := detector.NewHeaderSlashingDetector(&mock2.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return nil, expectedErr
		},
	})

	res, err := sd.VerifyData(&interceptedBlocks.InterceptedHeader{})

	require.Nil(t, res)
	require.Equal(t, expectedErr, err)
}

func TestHeaderSlashingDetector_VerifyData_NoSlashing(t *testing.T) {
	sd, _ := detector.NewHeaderSlashingDetector(&mock2.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock([]byte("proposer1"))}, nil
		},
	})

	hData := createInterceptedHeaderData([]byte("seed"))
	res, _ := sd.VerifyData(hData)

	require.Equal(t, res.GetType(), slash.None)
	require.Equal(t, res.GetData1(), hData)
	require.Nil(t, res.GetData2())

	res, _ = sd.VerifyData(hData)

	require.Equal(t, res.GetType(), slash.None)
	require.Equal(t, res.GetData1(), hData)
	require.Nil(t, res.GetData2())

	res, _ = sd.VerifyData(hData)

	require.Equal(t, res.GetType(), slash.None)
	require.Equal(t, res.GetData1(), hData)
	require.Nil(t, res.GetData2())
}

func TestHeaderSlashingDetector_VerifyData_DoubleProposal_MultipleProposal(t *testing.T) {
	sd, _ := detector.NewHeaderSlashingDetector(&mock2.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock([]byte("proposer1"))}, nil
		},
	})

	hData1 := createInterceptedHeaderData([]byte("seed1"))
	res, _ := sd.VerifyData(hData1)

	require.Equal(t, res.GetType(), slash.None)
	require.Equal(t, res.GetData1(), hData1)
	require.Nil(t, res.GetData2())

	hData2 := createInterceptedHeaderData([]byte("seed2"))
	res, _ = sd.VerifyData(hData2)

	require.Equal(t, res.GetType(), slash.DoubleProposal)
	require.Equal(t, res.GetData1(), hData2)
	require.Equal(t, res.GetData2(), hData1)

	hData3 := createInterceptedHeaderData([]byte("seed3"))
	res, _ = sd.VerifyData(hData3)

	require.Equal(t, res.GetType(), slash.MultipleProposal)
	require.Equal(t, res.GetData1(), hData3)
	require.Equal(t, res.GetData2(), hData1)
}

func TestHeaderSlashingDetector_VerifyData_MultipleProposal(t *testing.T) {
	sd, _ := detector.NewHeaderSlashingDetector(&mock2.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			return []sharding.Validator{mock.NewValidatorMock([]byte("proposer1"))}, nil
		},
	})

	hData1 := createInterceptedHeaderData([]byte("seed1"))
	res, _ := sd.VerifyData(hData1)

	require.Equal(t, res.GetType(), slash.None)
	require.Equal(t, res.GetData1(), hData1)
	require.Nil(t, res.GetData2())

	hData2 := createInterceptedHeaderData([]byte("seed2"))
	res, _ = sd.VerifyData(hData2)

	require.Equal(t, res.GetType(), slash.DoubleProposal)
	require.Equal(t, res.GetData1(), hData2)
	require.Equal(t, res.GetData2(), hData1)
}

func createInterceptedHeaderArg(randSeed []byte) *interceptedBlocks.ArgInterceptedBlockHeader {
	args := &interceptedBlocks.ArgInterceptedBlockHeader{
		ShardCoordinator:        &mock.ShardCoordinatorStub{},
		Hasher:                  mock.HasherMock{},
		Marshalizer:             &mock.MarshalizerMock{},
		HeaderSigVerifier:       &mock.HeaderSigVerifierStub{},
		HeaderIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		ValidityAttester:        &mock.ValidityAttesterStub{},
		EpochStartTrigger:       &mock.EpochStartTriggerStub{},
	}

	hdr := createBlockHeaderData(randSeed)
	args.HdrBuff, _ = args.Marshalizer.Marshal(hdr)

	return args
}

func createBlockHeaderData(randSeed []byte) *block.Header {
	return &block.Header{
		RandSeed: randSeed,
		ShardID:  1,
		Round:    2,
		Epoch:    3,
	}
}

func createInterceptedHeaderData(randSeed []byte) *interceptedBlocks.InterceptedHeader {
	args := createInterceptedHeaderArg(randSeed)
	interceptedHeader, _ := interceptedBlocks.NewInterceptedHeader(args)

	return interceptedHeader
}
