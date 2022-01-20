package slashMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	mockCoreData "github.com/ElrondNetwork/elrond-go-core/data/mock"
	"github.com/ElrondNetwork/elrond-go/process"
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
func CreateInterceptedHeaderData(header data.HeaderHandler) process.InterceptedHeader {
	args := createInterceptedHeaderArg(header)
	interceptedHeader, _ := interceptedBlocks.NewInterceptedHeader(args)

	return interceptedHeader
}

// CreateHeaderInfoData creates a new data.HeaderInfoHandler with given data.HeaderHandler
func CreateHeaderInfoData(header data.HeaderHandler) data.HeaderInfoHandler {
	hasher := hashingMocks.HasherMock{}
	marshaller := mock.MarshalizerMock{}

	headerBytes, _ := marshaller.Marshal(header)
	headerHash := hasher.Compute(string(headerBytes))

	return &mockCoreData.HeaderInfoStub{
		Header: header,
		Hash:   headerHash,
	}
}
