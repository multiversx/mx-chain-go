package asyncExecution

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/queue"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process/asyncExecution")

// ArgsHeadersExecutor holds all the components needed to create a new instance of *headersExecutor
type ArgsHeadersExecutor struct {
	BlocksQueue      BlocksQueue
	ExecutionTracker ExecutionResultsHandler
	BlockProcessor   BlockProcessor
}

type headersExecutor struct {
	blocksQueue      BlocksQueue
	executionTracker ExecutionResultsHandler
	blockProcessor   BlockProcessor
	cancelFunc       context.CancelFunc
}

// NewHeadersExecutor will create a new instance of *headersExecutor
func NewHeadersExecutor(args ArgsHeadersExecutor) (*headersExecutor, error) {
	if check.IfNil(args.BlocksQueue) {
		return nil, ErrNilHeadersQueue
	}
	if check.IfNil(args.ExecutionTracker) {
		return nil, ErrNilExecutionTracker
	}
	if check.IfNil(args.BlockProcessor) {
		return nil, ErrNilBlockProcessor
	}

	return &headersExecutor{
		blocksQueue:      args.BlocksQueue,
		executionTracker: args.ExecutionTracker,
		blockProcessor:   args.BlockProcessor,
	}, nil
}

// StartExecution starts a goroutine to continuously process blocks from the queue
// and add their results to the execution tracker until cancelled or closed.
func (he *headersExecutor) StartExecution() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	he.cancelFunc = cancelFunc

	go he.start(ctx)
}

func (he *headersExecutor) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// blocking operation
			headerBodyPair, ok := he.blocksQueue.Pop()
			if !ok {
				// close event
				return
			}

			he.process(headerBodyPair)
		}
	}
}

func (he *headersExecutor) process(pair queue.HeaderBodyPair) {
	executionResult, err := he.blockProcessor.ProcessBlock(pair.Header, pair.Body)
	if err != nil {
		log.Warn("headersExecutor.process process block failed", "err", err)
	}

	err = he.executionTracker.AddExecutionResult(executionResult)
	if err != nil {
		log.Warn("headersExecutor.process add execution result failed", "err", err)
	}
}

// Close will close the blocks execution look
func (he *headersExecutor) Close() error {
	he.blocksQueue.Close()
	if he.cancelFunc != nil {
		he.cancelFunc()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (he *headersExecutor) IsInterfaceNil() bool {
	return he == nil
}
