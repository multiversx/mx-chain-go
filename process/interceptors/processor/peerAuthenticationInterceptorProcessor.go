package processor

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/heartbeat"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgPeerAuthenticationInterceptorProcessor is the argument for the interceptor processor used for peer authentication
type ArgPeerAuthenticationInterceptorProcessor struct {
	PeerAuthenticationCacher storage.Cacher
}

// PeerAuthenticationInterceptorProcessor is the processor used when intercepting peer authentication
type PeerAuthenticationInterceptorProcessor struct {
	peerAuthenticationCacher storage.Cacher
}

// NewPeerAuthenticationInterceptorProcessor creates a new PeerAuthenticationInterceptorProcessor
func NewPeerAuthenticationInterceptorProcessor(arg ArgPeerAuthenticationInterceptorProcessor) (*PeerAuthenticationInterceptorProcessor, error) {
	if check.IfNil(arg.PeerAuthenticationCacher) {
		return nil, process.ErrNilPeerAuthenticationCacher
	}

	return &PeerAuthenticationInterceptorProcessor{
		peerAuthenticationCacher: arg.PeerAuthenticationCacher,
	}, nil
}

// Validate checks if the intercepted data can be processed
// returns nil as proper validity checks are done at intercepted data level
func (paip *PeerAuthenticationInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted peer authentication inside the peer authentication cacher
func (paip *PeerAuthenticationInterceptorProcessor) Save(data process.InterceptedData, fromConnectedPeer core.PeerID, _ string) error {
	interceptedPeerAuthenticationData, ok := data.(*heartbeat.InterceptedPeerAuthentication)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	paip.peerAuthenticationCacher.Put(fromConnectedPeer.Bytes(), interceptedPeerAuthenticationData, interceptedPeerAuthenticationData.SizeInBytes())
	return nil
}

// RegisterHandler registers a callback function to be notified of incoming peer authentication
func (paip *PeerAuthenticationInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("PeerAuthenticationInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (paip *PeerAuthenticationInterceptorProcessor) IsInterfaceNil() bool {
	return paip == nil
}
