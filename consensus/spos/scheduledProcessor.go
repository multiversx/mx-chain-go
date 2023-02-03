package spos

import (
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/process"
)

type processingStatus int

const (
	processingNotStarted processingStatus = 0
	processingError      processingStatus = 1
	inProgress           processingStatus = 2
	processingOK         processingStatus = 3
	stopped              processingStatus = 4

	processingNotStartedString = "processing not started"
	processingErrorString      = "processing error"
	inProgressString           = "processing in progress"
	processingOKString         = "processing OK"
	stoppedString              = "processing stopped"
	unexpectedString           = "unexpected"

	processingCheckStep = 10 * time.Millisecond
)

// String returns human readable processing status
func (ps processingStatus) String() string {
	switch ps {
	case processingNotStarted:
		return processingNotStartedString
	case processingError:
		return processingErrorString
	case inProgress:
		return inProgressString
	case processingOK:
		return processingOKString
	case stopped:
		return stoppedString
	default:
		return unexpectedString
	}
}

// ScheduledProcessorWrapperArgs holds the arguments required to instantiate the pipelineExecution
type ScheduledProcessorWrapperArgs struct {
	SyncTimer                ntp.SyncTimer
	Processor                process.ScheduledBlockProcessor
	RoundTimeDurationHandler process.RoundTimeDurationHandler
}

type scheduledProcessorWrapper struct {
	syncTimer                ntp.SyncTimer
	processor                process.ScheduledBlockProcessor
	roundTimeDurationHandler process.RoundTimeDurationHandler
	startTime                time.Time

	status processingStatus
	sync.RWMutex

	stopExecution chan struct{}
	oneInstance   sync.RWMutex
}

// NewScheduledProcessorWrapper creates a new processor for scheduled transactions
func NewScheduledProcessorWrapper(args ScheduledProcessorWrapperArgs) (*scheduledProcessorWrapper, error) {
	if check.IfNil(args.SyncTimer) {
		return nil, process.ErrNilSyncTimer
	}
	if check.IfNil(args.Processor) {
		return nil, process.ErrNilBlockProcessor
	}
	if check.IfNil(args.RoundTimeDurationHandler) {
		return nil, process.ErrNilRoundTimeDurationHandler
	}

	return &scheduledProcessorWrapper{
		syncTimer:                args.SyncTimer,
		processor:                args.Processor,
		roundTimeDurationHandler: args.RoundTimeDurationHandler,
		status:                   processingNotStarted,
		stopExecution:            make(chan struct{}, 1),
	}, nil
}

// ForceStopScheduledExecutionBlocking forces the execution to stop
func (sp *scheduledProcessorWrapper) ForceStopScheduledExecutionBlocking() {
	status := sp.getStatus()
	if status != inProgress {
		return
	}

	sp.stopExecution <- struct{}{}
	for status == inProgress {
		time.Sleep(processingCheckStep)
		status = sp.getStatus()
	}
	sp.setStatus(stopped)
	sp.emptyStopChannel()
}

func (sp *scheduledProcessorWrapper) waitForProcessingResultWithTimeout() processingStatus {
	status := sp.getStatus()
	for status == inProgress {
		remainingExecutionTime := sp.computeRemainingProcessingTime()
		select {
		case <-time.After(remainingExecutionTime):
			status = sp.getStatus()
			return status
		case <-time.After(processingCheckStep):
			status = sp.getStatus()
		}
	}

	return status
}

// IsProcessedOKWithTimeout returns true if the scheduled processing was finalized without error
// Function is blocking until the allotted time for processing is finished
func (sp *scheduledProcessorWrapper) IsProcessedOKWithTimeout() bool {
	status := sp.waitForProcessingResultWithTimeout()

	processedOK := status == processingOK
	log.Debug("scheduledProcessorWrapper.IsProcessedOKWithTimeout", "status", status.String())
	return processedOK
}

// StartScheduledProcessing starts the scheduled processing
func (sp *scheduledProcessorWrapper) StartScheduledProcessing(header data.HeaderHandler, body data.BodyHandler, startTime time.Time) {
	if !header.HasScheduledSupport() {
		log.Debug("scheduled processing not supported")
		sp.setStatus(processingOK)
		return
	}

	sp.oneInstance.Lock()

	sp.setStatus(inProgress)
	log.Debug("scheduledProcessorWrapper.StartScheduledProcessing - scheduled processing has been started")

	go func() {
		defer sp.oneInstance.Unlock()
		errSchExec := sp.processScheduledMiniBlocks(header, body, startTime)
		if errSchExec != nil {
			log.Debug("scheduledProcessorWrapper.processScheduledMiniBlocks",
				"err", errSchExec.Error())
			sp.setStatus(processingError)
			return
		}
		log.Debug("scheduledProcessorWrapper.StartScheduledProcessing - scheduled processing has finished OK")
		sp.setStatus(processingOK)
	}()
}

// IsInterfaceNil returns true if there is no value under the interface
func (sp *scheduledProcessorWrapper) IsInterfaceNil() bool {
	return sp == nil
}

func (sp *scheduledProcessorWrapper) emptyStopChannel() {
	for {
		select {
		case <-sp.stopExecution:
		default:
			return
		}
	}
}

func (sp *scheduledProcessorWrapper) computeRemainingProcessingTimeStoppable() time.Duration {
	select {
	case <-sp.stopExecution:
		return 0
	default:
		return sp.computeRemainingProcessingTime()
	}
}

func (sp *scheduledProcessorWrapper) computeRemainingProcessingTime() time.Duration {
	currTime := sp.syncTimer.CurrentTime()

	sp.RLock()
	elapsedTime := currTime.Sub(sp.startTime)
	sp.RUnlock()

	if elapsedTime < 0 {
		return 0
	}

	return sp.roundTimeDurationHandler.TimeDuration() - elapsedTime
}

func (sp *scheduledProcessorWrapper) getStatus() processingStatus {
	sp.RLock()
	defer sp.RUnlock()

	return sp.status
}

func (sp *scheduledProcessorWrapper) setStatus(status processingStatus) {
	sp.Lock()
	defer sp.Unlock()

	sp.status = status
}

func (sp *scheduledProcessorWrapper) processScheduledMiniBlocks(header data.HeaderHandler, body data.BodyHandler, startTime time.Time) error {
	sp.Lock()
	sp.startTime = startTime
	sp.Unlock()

	haveTime := func() time.Duration {
		return sp.computeRemainingProcessingTimeStoppable()
	}

	return sp.processor.ProcessScheduledBlock(header, body, haveTime)
}
