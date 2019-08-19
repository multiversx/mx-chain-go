package mock

type InterceptorThrottlerStub struct {
	CanProcessMessageCalled      func() bool
	StartMessageProcessingCalled func()
	EndMessageProcessingCalled   func()
}

func (its InterceptorThrottlerStub) CanProcessMessage() bool {
	return its.CanProcessMessageCalled()
}

func (its InterceptorThrottlerStub) StartMessageProcessing() {
	its.StartMessageProcessingCalled()
}

func (its InterceptorThrottlerStub) EndMessageProcessing() {
	its.EndMessageProcessingCalled()
}
