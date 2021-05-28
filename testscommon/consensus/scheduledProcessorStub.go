package consensus

import "github.com/ElrondNetwork/elrond-go/data"

// ScheduledProcessorStub -
type ScheduledProcessorStub struct {
	StartScheduledProcessingCalled func(header data.HeaderHandler, body data.BodyHandler)
	IsProcessedOKCalled            func() bool
}

// StartScheduledProcessing -
func (sps *ScheduledProcessorStub) StartScheduledProcessing(header data.HeaderHandler, body data.BodyHandler) {
	if sps.StartScheduledProcessingCalled != nil {
		sps.StartScheduledProcessingCalled(header, body)
	}
}

// IsProcessedOK -
func (sps *ScheduledProcessorStub) IsProcessedOK() bool {
	if sps.IsProcessedOKCalled != nil {
		return sps.IsProcessedOKCalled()
	}
	return true
}

// IsInterfaceNil -
func (sps *ScheduledProcessorStub) IsInterfaceNil() bool {
	return sps == nil
}
