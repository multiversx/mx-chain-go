package mock

import "sync/atomic"

type InterceptorThrottlerStub struct {
	CanProcessCalled     func() bool
	startProcessingCount int32
	endProcessingCount   int32
}

func (its *InterceptorThrottlerStub) CanProcess() bool {
	return its.CanProcessCalled()
}

func (its *InterceptorThrottlerStub) StartProcessing() {
	atomic.AddInt32(&its.startProcessingCount, 1)
}

func (its *InterceptorThrottlerStub) EndProcessing() {
	atomic.AddInt32(&its.endProcessingCount, 1)
}

func (it *InterceptorThrottlerStub) StartProcessingCount() int32 {
	return atomic.LoadInt32(&it.startProcessingCount)
}

func (it *InterceptorThrottlerStub) EndProcessingCount() int32 {
	return atomic.LoadInt32(&it.endProcessingCount)
}

func (its *InterceptorThrottlerStub) IsInterfaceNil() bool {
	if its == nil {
		return true
	}
	return false
}
