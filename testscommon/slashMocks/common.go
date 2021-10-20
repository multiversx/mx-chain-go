package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
)

func createInterceptedHeaderArg(header data.HeaderHandler) *interceptedBlocks.ArgInterceptedBlockHeader {
	args := &interceptedBlocks.ArgInterceptedBlockHeader{
		ShardCoordinator:        &mock.ShardCoordinatorStub{},
		Hasher:                  &hashingMocks.HasherMock{},
		Marshalizer:             &mock.MarshalizerMock{},
		HeaderSigVerifier:       &mock.HeaderSigVerifierStub{},
		HeaderIntegrityVerifier: &mock.HeaderIntegrityVerifierStub{},
		ValidityAttester:        &mock.ValidityAttesterStub{},
		EpochStartTrigger:       &mock.EpochStartTriggerStub{},
	}

	args.HdrBuff, _ = args.Marshalizer.Marshal(header)

	return args
}

// CreateInterceptedHeaderData creates a new interceptedBlocks.InterceptedHeader with given data.HeaderHandler
func CreateInterceptedHeaderData(header data.HeaderHandler) *interceptedBlocks.InterceptedHeader {
	args := createInterceptedHeaderArg(header)
	interceptedHeader, _ := interceptedBlocks.NewInterceptedHeader(args)

	return interceptedHeader
}
