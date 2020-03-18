package mock

import "sync/atomic"

// InterceptorThrottlerStub -
type InterceptorThrottlerStub struct {
	CanProcessCalled     func() bool
	startProcessingCount int32
	endProcessingCount   int32
}

// CanProcess -
func (its *InterceptorThrottlerStub) CanProcess() bool {
	return its.CanProcessCalled()
}

// StartProcessing -
func (its *InterceptorThrottlerStub) StartProcessing() {
	atomic.AddInt32(&its.startProcessingCount, 1)
}

// EndProcessing -
func (its *InterceptorThrottlerStub) EndProcessing() {
	atomic.AddInt32(&its.endProcessingCount, 1)
}

// StartProcessingCount -
func (its *InterceptorThrottlerStub) StartProcessingCount() int32 {
	return atomic.LoadInt32(&its.startProcessingCount)
}

// EndProcessingCount -
func (its *InterceptorThrottlerStub) EndProcessingCount() int32 {
	return atomic.LoadInt32(&its.endProcessingCount)
}

// IsInterfaceNil -
func (its *InterceptorThrottlerStub) IsInterfaceNil() bool {
	return its == nil
}
