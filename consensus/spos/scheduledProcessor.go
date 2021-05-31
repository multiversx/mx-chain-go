package spos

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
)

type processingStatus int

const (
	processingNotStarted processingStatus = 0
	processingError      processingStatus = 1
	inProgress           processingStatus = 2
	processingOK         processingStatus = 3

	processingNotStartedString = "processing not started"
	processingErrorString      = "processing error"
	inProgressString           = "processing in progress"
	processingOKString         = "processing OK"
	unexpectedString           = "processing OK"

	processingCheckStep = 1 * time.Millisecond
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
	default:
		return unexpectedString
	}
}

// ScheduledProcessorArgs holds the arguments required to instantiate the pipelineExecution
type ScheduledProcessorArgs struct {
	SyncTimer                  ntp.SyncTimer
	Processor                  process.ScheduledBlockProcessor
	ProcessingTimeMilliSeconds uint32
}

type scheduledProcessor struct {
	syncTimer      ntp.SyncTimer
	processingTime time.Duration
	processor      process.ScheduledBlockProcessor

	startTime time.Time
	status    processingStatus
	sync.RWMutex

	oneInstance sync.RWMutex
}

// NewScheduledProcessor creates a new processor for scheduled transactions
func NewScheduledProcessor(args ScheduledProcessorArgs) (*scheduledProcessor, error) {
	if check.IfNil(args.SyncTimer) {
		return nil, process.ErrNilSyncTimer
	}
	if check.IfNil(args.Processor) {
		return nil, process.ErrNilBlockProcessor
	}

	return &scheduledProcessor{
		syncTimer:      args.SyncTimer,
		processingTime: time.Duration(args.ProcessingTimeMilliSeconds) * time.Millisecond,
		processor:      args.Processor,
		status:         processingNotStarted,
	}, nil
}

func (sp *scheduledProcessor) computeRemainingProcessingTime() time.Duration {
	currTime := sp.syncTimer.CurrentTime()

	sp.RLock()
	elapsedTime := currTime.Sub(sp.startTime)
	sp.RUnlock()

	if elapsedTime < 0 {
		return 0
	}

	remainingTime := sp.processingTime - elapsedTime
	if remainingTime < 0 {
		return 0
	}

	return remainingTime
}

// IsProcessedOK returns true if the scheduled processing was finalized without error
// Function is blocking until the allotted time for processing is finished
func (sp *scheduledProcessor) IsProcessedOK() bool {
	status := sp.getStatus()

loop:
	for status == inProgress {
		remainingExecutionTime := sp.computeRemainingProcessingTime()
		select {
		case <-time.After(remainingExecutionTime):
			status = sp.getStatus()
			break loop
		case <-time.After(processingCheckStep):
			status = sp.getStatus()
		}
	}

	processedOK := processingOK == status
	log.Debug("scheduledProcessor.IsProcessedOK", "status", status.String())
	return processedOK
}

// StartScheduledProcessing starts the scheduled processing
func (sp *scheduledProcessor) StartScheduledProcessing(header data.HeaderHandler, body data.BodyHandler) {
	if !header.HasScheduledSupport() {
		log.Debug("scheduled processing not supported")
		sp.setStatus(processingOK)
		return
	}

	log.Debug("scheduledProcessor.StartScheduledProcessing - scheduled processing has been started")
	sp.setStatus(inProgress)

	sp.oneInstance.Lock()
	go func() {
		defer sp.oneInstance.Unlock()
		errSchExec := sp.processScheduledMiniBlocks(header, body)
		if errSchExec != nil {
			log.Error("scheduledProcessor.processScheduledMiniBlocks",
				"err", errSchExec.Error())
			sp.setStatus(processingError)
			return
		}
		log.Debug("scheduledProcessor.StartScheduledProcessing - scheduled processing has finished OK")
		sp.setStatus(processingOK)
	}()
}

func (sp *scheduledProcessor) getStatus() processingStatus {
	sp.RLock()
	defer sp.RUnlock()

	return sp.status
}

func (sp *scheduledProcessor) setStatus(status processingStatus) {
	sp.Lock()
	defer sp.Unlock()

	sp.status = status
}

func (sp *scheduledProcessor) getStartTime() time.Time {
	sp.RLock()
	defer sp.RUnlock()

	return sp.startTime
}

func (sp *scheduledProcessor) processScheduledMiniBlocks(header data.HeaderHandler, body data.BodyHandler) error {
	sp.Lock()
	sp.startTime = sp.syncTimer.CurrentTime()
	sp.Unlock()

	haveTime := func() time.Duration {
		return sp.computeRemainingProcessingTime()
	}

	return sp.processor.ProcessScheduledBlock(header, body, haveTime)
}

// IsInterfaceNil returns true if there is no value under the interface
func (sp *scheduledProcessor) IsInterfaceNil() bool {
	return sp == nil
}
