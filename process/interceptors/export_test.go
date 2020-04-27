package interceptors

import "github.com/ElrondNetwork/elrond-go/process"

func (mdi *MultiDataInterceptor) Topic() string {
	return mdi.topic
}

func (mdi *MultiDataInterceptor) InterceptedDebugHandler() process.InterceptedDebugHandler {
	mdi.mutInterceptedDebugHandler.RLock()
	defer mdi.mutInterceptedDebugHandler.RUnlock()

	return mdi.interceptedDebugHandler
}

func (sdi *SingleDataInterceptor) Topic() string {
	return sdi.topic
}

func (sdi *SingleDataInterceptor) InterceptedDebugHandler() process.InterceptedDebugHandler {
	sdi.mutInterceptedDebugHandler.RLock()
	defer sdi.mutInterceptedDebugHandler.RUnlock()

	return sdi.interceptedDebugHandler
}
