package processor

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgHeartbeatInterceptorProcessor is the argument for the interceptor processor used for heartbeat
type ArgHeartbeatInterceptorProcessor struct {
	HeartbeatCacher storage.Cacher
}

// heartbeatInterceptorProcessor is the processor used when intercepting heartbeat
type heartbeatInterceptorProcessor struct {
	heartbeatCacher storage.Cacher
}

// NewHeartbeatInterceptorProcessor creates a new heartbeatInterceptorProcessor
func NewHeartbeatInterceptorProcessor(arg ArgHeartbeatInterceptorProcessor) (*heartbeatInterceptorProcessor, error) {
	if check.IfNil(arg.HeartbeatCacher) {
		return nil, process.ErrNilHeartbeatCacher
	}

	return &heartbeatInterceptorProcessor{
		heartbeatCacher: arg.HeartbeatCacher,
	}, nil
}

// Validate checks if the intercepted data can be processed
// returns nil as proper validity checks are done at intercepted data level
func (hip *heartbeatInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted heartbeat inside the heartbeat cacher
func (hip *heartbeatInterceptorProcessor) Save(data process.InterceptedData, fromConnectedPeer core.PeerID, _ string) error {
	interceptedHeartbeat, ok := data.(interceptedDataMessageHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	hip.heartbeatCacher.Put(fromConnectedPeer.Bytes(), interceptedHeartbeat.Message(), interceptedHeartbeat.SizeInBytes())
	return nil
}

// RegisterHandler registers a callback function to be notified of incoming hearbeat
func (hip *heartbeatInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("heartbeatInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (hip *heartbeatInterceptorProcessor) IsInterfaceNil() bool {
	return hip == nil
}
