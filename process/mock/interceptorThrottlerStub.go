package mock

type InterceptorThrottlerStub struct {
	CanProcessCalled      func() bool
	StartProcessingCalled func()
	EndProcessingCalled   func()
}

func (its *InterceptorThrottlerStub) CanProcess() bool {
	return its.CanProcessCalled()
}

func (its *InterceptorThrottlerStub) StartProcessing() {
	its.StartProcessingCalled()
}

func (its *InterceptorThrottlerStub) EndProcessing() {
	its.EndProcessingCalled()
}

func (its *InterceptorThrottlerStub) IsInterfaceNil() bool {
	if its == nil {
		return true
	}
	return false
}
