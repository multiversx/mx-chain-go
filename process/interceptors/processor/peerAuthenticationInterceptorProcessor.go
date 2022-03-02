package processor

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgPeerAuthenticationInterceptorProcessor is the argument for the interceptor processor used for peer authentication
type ArgPeerAuthenticationInterceptorProcessor struct {
	PeerAuthenticationCacher storage.Cacher
	PeerShardMapper          process.PeerShardMapper
}

// peerAuthenticationInterceptorProcessor is the processor used when intercepting peer authentication
type peerAuthenticationInterceptorProcessor struct {
	peerAuthenticationCacher storage.Cacher
	peerShardMapper          process.PeerShardMapper
}

// NewPeerAuthenticationInterceptorProcessor creates a new peerAuthenticationInterceptorProcessor
func NewPeerAuthenticationInterceptorProcessor(args ArgPeerAuthenticationInterceptorProcessor) (*peerAuthenticationInterceptorProcessor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &peerAuthenticationInterceptorProcessor{
		peerAuthenticationCacher: args.PeerAuthenticationCacher,
		peerShardMapper:          args.PeerShardMapper,
	}, nil
}

func checkArgs(args ArgPeerAuthenticationInterceptorProcessor) error {
	if check.IfNil(args.PeerAuthenticationCacher) {
		return process.ErrNilPeerAuthenticationCacher
	}
	if check.IfNil(args.PeerShardMapper) {
		return process.ErrNilPeerShardMapper
	}

	return nil
}

// Validate checks if the intercepted data can be processed
// returns nil as proper validity checks are done at intercepted data level
func (paip *peerAuthenticationInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted peer authentication inside the peer authentication cacher
func (paip *peerAuthenticationInterceptorProcessor) Save(data process.InterceptedData, fromConnectedPeer core.PeerID, _ string) error {
	interceptedPeerAuthenticationData, ok := data.(interceptedDataMessageHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	paip.peerAuthenticationCacher.Put(fromConnectedPeer.Bytes(), interceptedPeerAuthenticationData.Message(), interceptedPeerAuthenticationData.SizeInBytes())

	return paip.updatePeerInfo(interceptedPeerAuthenticationData.Message())
}

func (paip *peerAuthenticationInterceptorProcessor) updatePeerInfo(message interface{}) error {
	peerAuthenticationData, ok := message.(heartbeat.PeerAuthentication)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	paip.peerShardMapper.UpdatePeerIDPublicKeyPair(core.PeerID(peerAuthenticationData.GetPid()), peerAuthenticationData.GetPubkey())

	return nil
}

// RegisterHandler registers a callback function to be notified of incoming peer authentication
func (paip *peerAuthenticationInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("peerAuthenticationInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (paip *peerAuthenticationInterceptorProcessor) IsInterfaceNil() bool {
	return paip == nil
}
