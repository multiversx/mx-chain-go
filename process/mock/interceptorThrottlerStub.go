package mock

import "sync/atomic"

type InterceptorThrottlerStub struct {
	CanProcessCalled     func() bool
	startProcessingCount int32
	endProcessingCount   int32
}

func (it *InterceptorThrottlerStub) CanProcess() bool {
	return it.CanProcessCalled()
}

func (it *InterceptorThrottlerStub) StartProcessing() {
	atomic.AddInt32(&it.startProcessingCount, 1)
}

func (it *InterceptorThrottlerStub) EndProcessing() {
	atomic.AddInt32(&it.endProcessingCount, 1)
}

func (it *InterceptorThrottlerStub) StartProcessingCount() int32 {
	return atomic.LoadInt32(&it.startProcessingCount)
}

func (it *InterceptorThrottlerStub) EndProcessingCount() int32 {
	return atomic.LoadInt32(&it.endProcessingCount)
}

func (it *InterceptorThrottlerStub) IsInterfaceNil() bool {
	if it == nil {
		return true
	}
	return false
}
