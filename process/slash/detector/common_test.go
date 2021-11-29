package detector_test

import (
	mockEpochStart "github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/slash/detector"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/slashMocks"
)

var validatorPubKey = []byte("validator pub key")

func generateMultipleHeaderDetectorArgs() detector.MultipleHeaderDetectorArgs {
	nodesCoordinator := &mockEpochStart.NodesCoordinatorStub{
		ComputeConsensusGroupCalled: func(_ []byte, _ uint64, _ uint32, _ uint32) ([]sharding.Validator, error) {
			validator := mock.NewValidatorMock(validatorPubKey)
			return []sharding.Validator{validator}, nil
		},
	}

	return detector.MultipleHeaderDetectorArgs{
		NodesCoordinator:  nodesCoordinator,
		RoundHandler:      &mock.RoundHandlerMock{},
		SlashingCache:     &slashMocks.RoundDetectorCacheStub{},
		Hasher:            &hashingMocks.HasherMock{},
		Marshaller:        &testscommon.MarshalizerMock{},
		HeaderSigVerifier: &mock.HeaderSigVerifierStub{},
	}
}
