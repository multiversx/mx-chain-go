package asyncExecution

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/asyncExecution/cache"
)

var log = logger.GetOrCreate("process/asyncExecution")

const timeToSleep = time.Millisecond * 5
const timeToSleepOnError = time.Millisecond * 300
const maxRetryAttempts = 10
const maxBackoffTime = time.Second * 5

// ArgsHeadersExecutor holds all the components needed to create a new instance of *headersExecutor
type ArgsHeadersExecutor struct {
	BlocksCache      BlocksCache
	ExecutionTracker ExecutionResultsHandler
	BlockProcessor   BlockProcessor
	BlockChain       data.ChainHandler
}

type headersExecutor struct {
	blocksCache      BlocksCache
	executionTracker ExecutionResultsHandler
	blockProcessor   BlockProcessor
	blockChain       data.ChainHandler
	cancelFunc       context.CancelFunc
	mutPaused        sync.RWMutex
	isPaused         bool
}

// NewHeadersExecutor will create a new instance of *headersExecutor
func NewHeadersExecutor(args ArgsHeadersExecutor) (*headersExecutor, error) {
	if check.IfNil(args.BlocksCache) {
		return nil, ErrNilHeadersCache
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
		blocksCache:      args.BlocksCache,
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
	log.Debug("headersExecutor.PauseExecution: pausing execution")

	he.mutPaused.Lock()
	defer he.mutPaused.Unlock()

	if he.isPaused {
		return
	}

	he.isPaused = true
}

// ResumeExecution resumes the execution
func (he *headersExecutor) ResumeExecution() {
	log.Debug("headersExecutor.ResumeExecution: resuming execution")

	he.mutPaused.Lock()
	defer he.mutPaused.Unlock()

	he.isPaused = false
}

func (he *headersExecutor) start(ctx context.Context) {
	log.Debug("headersExecutor.start: starting execution")

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
			/// get pair by nonce from blockchain
			lastExecutedHeader := he.blockChain.GetLastExecutedBlockHeader()
			if lastExecutedHeader == nil {
				time.Sleep(timeToSleep)
				continue
			}

			headerBodyPair, ok := he.blocksCache.GetByNonce(lastExecutedHeader.GetNonce() + 1)
			if !ok {
				time.Sleep(timeToSleep)
				continue
			}

			if check.IfNil(headerBodyPair.Header) || check.IfNil(headerBodyPair.Body) {
				log.Debug("headersExecutor.start: popped nil header or body, continuing...")
				continue
			}

			err := he.process(headerBodyPair)
			if err != nil {
				he.handleProcessError(ctx, headerBodyPair)
			}
		}
	}
}

func (he *headersExecutor) handleProcessError(ctx context.Context, pair cache.HeaderBodyPair) {
	retryCount := 0
	backoffTime := timeToSleepOnError

	for retryCount < maxRetryAttempts {
		pairFromQueue, ok := he.blocksCache.GetLastAdded()
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
				log.Debug("headersExecutor.handleProcessError - retry succeeded",
					"nonce", pair.Header.GetNonce(),
					"retry_count", retryCount)
				return
			}
			retryCount++
			log.Warn("headersExecutor.handleProcessError - retry failed",
				"nonce", pair.Header.GetNonce(),
				"retry_count", retryCount,
				"max_retries", maxRetryAttempts,
				"err", err)

			// Exponential backoff with maximum limit
			time.Sleep(backoffTime)
			backoffTime = backoffTime * 2
			if backoffTime > maxBackoffTime {
				backoffTime = maxBackoffTime
			}
		}
	}

	log.Error("headersExecutor.handleProcessError - max retries exceeded, skipping block",
		"nonce", pair.Header.GetNonce(),
		"max_retries", maxRetryAttempts)
}

func (he *headersExecutor) process(pair cache.HeaderBodyPair) error {
	executionResult, err := he.blockProcessor.ProcessBlockProposal(pair.Header, pair.Body)
	if err != nil {
		log.Warn("headersExecutor.process process block failed",
			"nonce", pair.Header.GetNonce(),
			"prevHash", pair.Header.GetPrevHash(),
			"err", err,
		)
		return err
	}

	// Validate execution result
	if check.IfNil(executionResult) {
		log.Warn("headersExecutor.process - nil execution result received",
			"nonce", pair.Header.GetNonce())
		return ErrNilExecutionResult
	}

	// todo call check here

	err = he.executionTracker.AddExecutionResult(executionResult)
	if err != nil {
		log.Warn("headersExecutor.process add execution result failed",
			"nonce", pair.Header.GetNonce(),
			"err", err,
		)
		return err
	}

	lastExecutionResult := he.blockChain.GetLastExecutionResult()
	if lastExecutionResult != nil {
		if !bytes.Equal(lastExecutionResult.GetHeaderHash(), pair.Header.GetPrevHash()) {
			log.Error("headersExecutor.process - header hash mismatch")
			return nil
		}
	}

	he.blockChain.SetFinalBlockInfo(
		executionResult.GetHeaderNonce(),
		executionResult.GetHeaderHash(),
		executionResult.GetRootHash(),
	)

	he.blockChain.SetLastExecutedBlockHeaderAndRootHash(pair.Header, executionResult.GetHeaderHash(), executionResult.GetRootHash())
	he.blockChain.SetLastExecutionResult(executionResult)

	log.Debug("headersExecutor.process completed",
		"nonce", pair.Header.GetNonce(),
		"exec nonce", executionResult.GetHeaderNonce(),
		"exec rootHash", executionResult.GetRootHash(),
	)

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
