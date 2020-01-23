package track_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/track"
	"github.com/stretchr/testify/assert"
)

func CreateBlockProcessorMockArguments() track.ArgBlockProcessor {
	shardCoordinatorMock := mock.NewMultipleShardsCoordinatorMock()
	argsHeaderValidator := block.ArgsHeaderValidator{
		Hasher:      &mock.HasherMock{},
		Marshalizer: &mock.MarshalizerMock{},
	}
	headerValidator, _ := block.NewHeaderValidator(argsHeaderValidator)

	arguments := track.ArgBlockProcessor{
		HeaderValidator:               headerValidator,
		RequestHandler:                &mock.RequestHandlerStub{},
		ShardCoordinator:              shardCoordinatorMock,
		BlockTracker:                  &track.BlockTrackerHandlerMock{},
		CrossNotarizer:                &track.BlockNotarizerHandlerMock{},
		CrossNotarizedHeadersNotifier: &track.BlockNotifierHandlerMock{},
		SelfNotarizedHeadersNotifier:  &track.BlockNotifierHandlerMock{},
	}

	return arguments
}

func TestNewBlockProcessor_ShouldErrNilHeaderValidator(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.HeaderValidator = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, process.ErrNilHeaderValidator, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilRequestHandler(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.RequestHandler = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, process.ErrNilRequestHandler, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilShardCoordinator(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.ShardCoordinator = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, process.ErrNilShardCoordinator, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilBlockTrackerHandler(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.BlockTracker = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrNilBlockTrackerHandler, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrNilCrossNotarizer(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.CrossNotarizer = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrNilCrossNotarizer, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrCrossNotarizedHeadersNotifier(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.CrossNotarizedHeadersNotifier = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrCrossNotarizedHeadersNotifier, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldErrSelfNotarizedHeadersNotifier(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	blockProcessorArguments.SelfNotarizedHeadersNotifier = nil

	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Equal(t, track.ErrSelfNotarizedHeadersNotifier, err)
	assert.Nil(t, bp)
}

func TestNewBlockProcessor_ShouldWork(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	assert.Nil(t, err)
	assert.NotNil(t, bp)
}

func TestProcessReceivedHeader_ShouldWork(t *testing.T) {
	blockProcessorArguments := CreateBlockProcessorMockArguments()
	bp, err := track.NewBlockProcessor(blockProcessorArguments)

	bp.ProcessReceivedHeader(header)
	assert.Nil(t, err)
	assert.NotNil(t, bp)
}
