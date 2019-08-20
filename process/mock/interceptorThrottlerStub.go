package mock

type InterceptorThrottlerStub struct {
	CanProcessCalled     func() bool
	StartToProcessCalled func()
	EndProcessCalled     func()
}

func (its InterceptorThrottlerStub) CanProcess() bool {
	return its.CanProcessCalled()
}

func (its InterceptorThrottlerStub) StartToProcess() {
	its.StartToProcessCalled()
}

func (its InterceptorThrottlerStub) EndProcess() {
	its.EndProcessCalled()
}
