package processor

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/heartbeat"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgHeartbeatInterceptorProcessor is the argument for the interceptor processor used for heartbeat
type ArgHeartbeatInterceptorProcessor struct {
	HeartbeatCacher storage.Cacher
}

// HeartbeatInterceptorProcessor is the processor used when intercepting heartbeat
type HeartbeatInterceptorProcessor struct {
	heartbeatCacher storage.Cacher
}

// NewHeartbeatInterceptorProcessor creates a new HeartbeatInterceptorProcessor
func NewHeartbeatInterceptorProcessor(arg ArgHeartbeatInterceptorProcessor) (*HeartbeatInterceptorProcessor, error) {
	if check.IfNil(arg.HeartbeatCacher) {
		return nil, process.ErrNilHeartbeatCacher
	}

	return &HeartbeatInterceptorProcessor{
		heartbeatCacher: arg.HeartbeatCacher,
	}, nil
}

// Validate checks if the intercepted data can be processed
// returns nil as proper validity checks are done at intercepted data level
func (hip *HeartbeatInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted heartbeat inside the heartbeat cacher
func (hip *HeartbeatInterceptorProcessor) Save(data process.InterceptedData, fromConnectedPeer core.PeerID, _ string) error {
	interceptedHeartbeat, ok := data.(*heartbeat.InterceptedHeartbeat)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	hip.heartbeatCacher.Put(fromConnectedPeer.Bytes(), interceptedHeartbeat, interceptedHeartbeat.SizeInBytes())
	return nil
}

// RegisterHandler registers a callback function to be notified of incoming hearbeat
func (hip *HeartbeatInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("HeartbeatInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (hip *HeartbeatInterceptorProcessor) IsInterfaceNil() bool {
	return hip == nil
}
