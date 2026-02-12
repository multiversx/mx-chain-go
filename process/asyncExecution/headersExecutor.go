package asyncExecution

import (
	"bytes"
	"context"
	"errors"
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
	BlocksCache                 BlocksCache
	ExecutionTracker            ExecutionResultsHandler
	BlockProcessor              BlockProcessor
	BlockChain                  data.ChainHandler
	SignalProcessCompletionChan chan uint64
}

type headersExecutor struct {
	blocksCache                 BlocksCache
	executionTracker            ExecutionResultsHandler
	blockProcessor              BlockProcessor
	blockChain                  data.ChainHandler
	cancelFunc                  context.CancelFunc
	mutPaused                   sync.RWMutex
	isPaused                    bool
	signalProcessCompletionChan chan uint64
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
		blocksCache:                 args.BlocksCache,
		executionTracker:            args.ExecutionTracker,
		blockProcessor:              args.BlockProcessor,
		blockChain:                  args.BlockChain,
		signalProcessCompletionChan: args.SignalProcessCompletionChan,
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

			lastExecutedNonce, lastExecutedHeaderHash, _ := he.blockChain.GetLastExecutedBlockInfo()
			if len(lastExecutedHeaderHash) == 0 {
				time.Sleep(timeToSleep)
				continue
			}

			// check if we need to execute another block for the same nonce (replacement block)
			headerBodyPair, ok := he.blocksCache.GetByNonce(lastExecutedNonce)
			if !ok {
				// Either block not in cache (genesis or cleaned) or same hash (already executed)
				// Try to get the next block to execute
				headerBodyPair, ok = he.blocksCache.GetByNonce(lastExecutedNonce + 1)
				if !ok {
					time.Sleep(timeToSleep)
					continue
				}
			}

			if headerBodyPair.Header.GetNonce() == lastExecutedNonce && !bytes.Equal(lastExecutedHeaderHash, headerBodyPair.HeaderHash) {
				// Different block at same nonce - this is a replacement block and needs to be executed
				log.Debug("headersExecutor.start: detected replacement block at same nonce",
					"nonce", lastExecutedNonce,
					"executed_hash", lastExecutedHeaderHash,
					"replacement_hash", headerBodyPair.HeaderHash,
				)
			}

			if bytes.Equal(lastExecutedHeaderHash, headerBodyPair.HeaderHash) {
				// Already executed this block, try to get the next one
				headerBodyPair, ok = he.blocksCache.GetByNonce(lastExecutedNonce + 1)
				if !ok {
					time.Sleep(timeToSleep)
					continue
				}
			}

			err := he.process(headerBodyPair)
			if err != nil {
				if errors.Is(err, ErrContextMismatch) {
					time.Sleep(timeToSleep)
					continue
				}
				he.handleProcessError(ctx, headerBodyPair)
			}
		}
	}
}

func (he *headersExecutor) handleProcessError(ctx context.Context, pair cache.HeaderBodyPair) {
	retryCount := 0
	backoffTime := timeToSleepOnError

	for retryCount < maxRetryAttempts {
		pairFromQueue, ok := he.blocksCache.GetByNonce(pair.Header.GetNonce())
		if ok && bytes.Equal(pair.HeaderHash, pairFromQueue.HeaderHash) {
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
	ok := he.checkLastExecutionResultContext(pair.Header, pair.HeaderHash)
	if !ok {
		return ErrContextMismatch
	}

	executionResult, err := he.blockProcessor.ProcessBlockProposal(pair.Header, pair.HeaderHash, pair.Body)
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

	ok = he.checkLastExecutionResultContext(pair.Header, pair.HeaderHash)
	if !ok {
		return nil
	}

	lastCommittedBlockHash := he.blockChain.GetCurrentBlockHeaderHash()
	lastCommittedBlockHeader := he.blockChain.GetCurrentBlockHeader()
	if !check.IfNil(lastCommittedBlockHeader) &&
		executionResult.GetHeaderNonce() == lastCommittedBlockHeader.GetNonce() &&
		!bytes.Equal(executionResult.GetHeaderHash(), lastCommittedBlockHash) {
		log.Debug("headersExecutor.process - execution result header hash does not match last committed block hash",
			"nonce", pair.Header.GetNonce(),
			"exec_header_hash", executionResult.GetHeaderHash(),
			"committed_block_hash", lastCommittedBlockHash,
		)
		return nil
	}

	added, err := he.executionTracker.AddExecutionResult(executionResult)
	if err != nil {
		log.Warn("headersExecutor.process add execution result failed",
			"nonce", pair.Header.GetNonce(),
			"err", err,
		)
		return err
	}
	if !added {
		// Result was rejected because consensus already committed a different block for this nonce.
		// Skip blockchain updates as the commit flow already set the correct state.
		log.Debug("headersExecutor.process execution result not added, skipping blockchain updates",
			"nonce", pair.Header.GetNonce(),
		)
		return nil
	}

	lastExecutionResult := he.blockChain.GetLastExecutionResult()
	if !check.IfNil(lastExecutionResult) {
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

	he.signalProcessCompletion(pair.Header.GetNonce())

	log.Debug("headersExecutor.process completed",
		"nonce", pair.Header.GetNonce(),
		"exec nonce", executionResult.GetHeaderNonce(),
		"exec rootHash", executionResult.GetRootHash(),
	)

	return nil
}

func (he *headersExecutor) signalProcessCompletion(currentNonce uint64) {
	if he.signalProcessCompletionChan == nil {
		return
	}

	select {
	case he.signalProcessCompletionChan <- currentNonce:
	default:
	}
}

func (he *headersExecutor) checkLastExecutionResultContext(
	currentHeader data.HeaderHandler,
	currentHeaderHash []byte,
) bool {
	if check.IfNil(currentHeader) {
		return false
	}

	lastExecutionResult := he.blockChain.GetLastExecutionResult()
	if check.IfNil(lastExecutionResult) {
		return true
	}

	if process.IsReplacementBlockForExecution(currentHeader, currentHeaderHash, lastExecutionResult) {
		return true
	}

	if currentHeader.GetNonce() != lastExecutionResult.GetHeaderNonce()+1 {
		log.Debug("headersExecutor.process: concurrent revert event",
			"previous nonce", lastExecutionResult.GetHeaderNonce(),
			"current nonce", currentHeader.GetNonce(),
			"err", process.ErrWrongNonceInBlock,
		)
		return false
	}

	lastExecResultHeaderHash := lastExecutionResult.GetHeaderHash()

	if !bytes.Equal(lastExecResultHeaderHash, currentHeader.GetPrevHash()) {
		log.Debug("headersExecutor.process: concurrent revert event",
			"header previous hash", currentHeader.GetPrevHash(),
			"last execution result hash", lastExecResultHeaderHash,
			"err", process.ErrBlockHashDoesNotMatch,
		)
		return false
	}

	return true
}

// GetSignalProcessCompletionChan returns the channel used to signal the sync loop after execution completes
func (he *headersExecutor) GetSignalProcessCompletionChan() chan uint64 {
	return he.signalProcessCompletionChan
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
