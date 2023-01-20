package interceptors

import "github.com/multiversx/mx-chain-go/process"

// Topic -
func (mdi *MultiDataInterceptor) Topic() string {
	return mdi.topic
}

// InterceptedDebugHandler -
func (mdi *MultiDataInterceptor) InterceptedDebugHandler() process.InterceptedDebugger {
	mdi.mutDebugHandler.RLock()
	defer mdi.mutDebugHandler.RUnlock()

	return mdi.debugHandler
}

// Topic -
func (sdi *SingleDataInterceptor) Topic() string {
	return sdi.topic
}

// InterceptedDebugHandler -
func (sdi *SingleDataInterceptor) InterceptedDebugHandler() process.InterceptedDebugger {
	sdi.mutDebugHandler.RLock()
	defer sdi.mutDebugHandler.RUnlock()

	return sdi.debugHandler
}

// ChunksProcessor
func (mdi *MultiDataInterceptor) ChunksProcessor() process.InterceptedChunksProcessor {
	mdi.mutChunksProcessor.RLock()
	defer mdi.mutChunksProcessor.RUnlock()

	return mdi.chunksProcessor
}
