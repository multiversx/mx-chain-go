package testscommon

import "time"

// SizeSyncStatisticsHandlerStub -
type SizeSyncStatisticsHandlerStub struct {
	AddNumBytesReceivedCalled func(bytes uint64)
	AddProcessingTimeCalled   func(duration time.Duration)
	IncrementIterationCalled  func()
	ProcessingTimeCalled      func() time.Duration
	ResetCalled               func()
	AddNumReceivedCalled      func(value int)
	AddNumLargeCalled         func(value int)
	SetNumMissingCalled       func(rootHash []byte, value int)
	NumBytesReceivedField     uint64
	NumTriesField             int
	NumIterationsField        int
	NumReceivedField          int
	NumLargeField             int
	NumMissingField           int
}

// Reset -
func (stub *SizeSyncStatisticsHandlerStub) Reset() {
	if stub.ResetCalled != nil {
		stub.ResetCalled()
	}
}

// AddNumReceived -
func (stub *SizeSyncStatisticsHandlerStub) AddNumReceived(value int) {
	if stub.AddNumReceivedCalled != nil {
		stub.AddNumReceivedCalled(value)
	}
}

// AddNumLarge -
func (stub *SizeSyncStatisticsHandlerStub) AddNumLarge(value int) {
	if stub.AddNumLargeCalled != nil {
		stub.AddNumLargeCalled(value)
	}
}

// SetNumMissing -
func (stub *SizeSyncStatisticsHandlerStub) SetNumMissing(rootHash []byte, value int) {
	if stub.SetNumMissingCalled != nil {
		stub.SetNumMissingCalled(rootHash, value)
	}
}

// NumReceived -
func (stub *SizeSyncStatisticsHandlerStub) NumReceived() int {
	return stub.NumReceivedField
}

// NumLarge -
func (stub *SizeSyncStatisticsHandlerStub) NumLarge() int {
	return stub.NumLargeField
}

// NumMissing -
func (stub *SizeSyncStatisticsHandlerStub) NumMissing() int {
	return stub.NumMissingField
}

// AddNumBytesReceived -
func (stub *SizeSyncStatisticsHandlerStub) AddNumBytesReceived(bytes uint64) {
	if stub.AddNumBytesReceivedCalled != nil {
		stub.AddNumBytesReceivedCalled(bytes)
	}
}

// NumBytesReceived -
func (stub *SizeSyncStatisticsHandlerStub) NumBytesReceived() uint64 {
	return stub.NumBytesReceivedField
}

// NumTries -
func (stub *SizeSyncStatisticsHandlerStub) NumTries() int {
	return stub.NumTriesField
}

// AddProcessingTime -
func (stub *SizeSyncStatisticsHandlerStub) AddProcessingTime(duration time.Duration) {
	if stub.AddProcessingTimeCalled != nil {
		stub.AddProcessingTimeCalled(duration)
	}
}

// IncrementIteration -
func (stub *SizeSyncStatisticsHandlerStub) IncrementIteration() {
	if stub.IncrementIterationCalled != nil {
		stub.IncrementIterationCalled()
	}
}

// ProcessingTime -
func (stub *SizeSyncStatisticsHandlerStub) ProcessingTime() time.Duration {
	if stub.ProcessingTimeCalled != nil {
		return stub.ProcessingTimeCalled()
	}
	return time.Second
}

// NumIterations -
func (stub *SizeSyncStatisticsHandlerStub) NumIterations() int {
	return stub.NumIterationsField
}

// IsInterfaceNil -
func (stub *SizeSyncStatisticsHandlerStub) IsInterfaceNil() bool {
	return stub == nil
}
