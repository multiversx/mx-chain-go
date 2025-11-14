package asyncExecution

import (
	"context"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/queue"
)

var log = logger.GetOrCreate("process/asyncExecution")

const timeToSleep = time.Millisecond * 5

// ArgsHeadersExecutor holds all the components needed to create a new instance of *headersExecutor
type ArgsHeadersExecutor struct {
	BlocksQueue      BlocksQueue
	ExecutionTracker ExecutionResultsHandler
	BlockProcessor   BlockProcessor
	BlockChain       data.ChainHandler
}

type headersExecutor struct {
	blocksQueue      BlocksQueue
	executionTracker ExecutionResultsHandler
	blockProcessor   BlockProcessor
	blockChain       data.ChainHandler
	cancelFunc       context.CancelFunc
	mutPaused        sync.RWMutex
	isPaused         bool
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
	if check.IfNil(args.BlockChain) {
		return nil, process.ErrNilBlockChain
	}

	instance := &headersExecutor{
		blocksQueue:      args.BlocksQueue,
		executionTracker: args.ExecutionTracker,
		blockProcessor:   args.BlockProcessor,
		blockChain:       args.BlockChain,
	}

	return instance, nil
}

// StartExecution starts a goroutine to continuously process blocks from the queue
// and add their results to the execution tracker until cancelled or closed.
func (he *headersExecutor) StartExecution() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	he.cancelFunc = cancelFunc

	go he.start(ctx)
}

// PauseExecution pauses the execution
func (he *headersExecutor) PauseExecution() {
	he.mutPaused.Lock()
	defer he.mutPaused.Unlock()

	if he.isPaused {
		return
	}

	he.isPaused = true
}

// ResumeExecution resumes the execution
func (he *headersExecutor) ResumeExecution() {
	he.mutPaused.Lock()
	defer he.mutPaused.Unlock()

	he.isPaused = false
}

func (he *headersExecutor) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			he.mutPaused.RLock()
			isPaused := he.isPaused
			he.mutPaused.RUnlock()

			if isPaused {
				time.Sleep(timeToSleep)
				continue
			}

			// blocking operation
			headerBodyPair, ok := he.blocksQueue.Pop()
			if !ok {
				// close event
				return
			}

			err := he.process(headerBodyPair)
			if err != nil {
				he.handleProcessError(ctx, headerBodyPair)
			}
		}
	}
}

func (he *headersExecutor) handleProcessError(ctx context.Context, pair queue.HeaderBodyPair) {
	for {
		pairFromQueue, ok := he.blocksQueue.Peek()
		if ok && pairFromQueue.Header.GetNonce() == pair.Header.GetNonce() {
			// continue the processing (pop the next header from queue)
			return
		}
		select {
		case <-ctx.Done():
			return
		default:
			// retry with the same pair
			err := he.process(pair)
			if err == nil {
				return
			}
		}
	}
}

func (he *headersExecutor) process(pair queue.HeaderBodyPair) error {
	executionResult, err := he.blockProcessor.ProcessBlockProposal(pair.Header, pair.Body)
	if err != nil {
		log.Warn("headersExecutor.process process block failed", "err", err)
		return err
	}

	err = he.executionTracker.AddExecutionResult(executionResult)
	if err != nil {
		log.Warn("headersExecutor.process add execution result failed", "err", err)
		return nil
	}

	he.blockChain.SetFinalBlockInfo(
		executionResult.GetHeaderNonce(),
		executionResult.GetHeaderHash(),
		executionResult.GetRootHash(),
	)

	he.blockChain.SetLastExecutedBlockHeaderAndRootHash(pair.Header, executionResult.GetHeaderHash(), executionResult.GetRootHash())

	return nil
}

// Close will close the blocks execution loop
func (he *headersExecutor) Close() error {
	if he.cancelFunc != nil {
		he.cancelFunc()
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (he *headersExecutor) IsInterfaceNil() bool {
	return he == nil
}
