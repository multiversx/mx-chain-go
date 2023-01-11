package consensus

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
)

// ScheduledProcessorStub -
type ScheduledProcessorStub struct {
	StartScheduledProcessingCalled func(header data.HeaderHandler, body data.BodyHandler, startTime time.Time)
	IsProcessedOKCalled            func() bool
}

// StartScheduledProcessing -
func (sps *ScheduledProcessorStub) StartScheduledProcessing(header data.HeaderHandler, body data.BodyHandler, startTime time.Time) {
	if sps.StartScheduledProcessingCalled != nil {
		sps.StartScheduledProcessingCalled(header, body, startTime)
	}
}

// ForceStopScheduledExecutionBlocking -
func (sps *ScheduledProcessorStub) ForceStopScheduledExecutionBlocking() {
}

// IsProcessedOKWithTimeout -
func (sps *ScheduledProcessorStub) IsProcessedOKWithTimeout() bool {
	if sps.IsProcessedOKCalled != nil {
		return sps.IsProcessedOKCalled()
	}
	return true
}

// IsInterfaceNil -
func (sps *ScheduledProcessorStub) IsInterfaceNil() bool {
	return sps == nil
}
