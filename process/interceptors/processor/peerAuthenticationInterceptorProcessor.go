package processor

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/heartbeat"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

// ArgPeerAuthenticationInterceptorProcessor is the argument for the interceptor processor used for peer authentication
type ArgPeerAuthenticationInterceptorProcessor struct {
	PeerAuthenticationCacher storage.Cacher
	PeerShardMapper          process.PeerShardMapper
	Marshaller               marshal.Marshalizer
	HardforkTrigger          heartbeat.HardforkTrigger
}

// peerAuthenticationInterceptorProcessor is the processor used when intercepting peer authentication
type peerAuthenticationInterceptorProcessor struct {
	peerAuthenticationCacher storage.Cacher
	peerShardMapper          process.PeerShardMapper
	marshaller               marshal.Marshalizer
	hardforkTrigger          heartbeat.HardforkTrigger
}

// NewPeerAuthenticationInterceptorProcessor creates a new peerAuthenticationInterceptorProcessor
func NewPeerAuthenticationInterceptorProcessor(args ArgPeerAuthenticationInterceptorProcessor) (*peerAuthenticationInterceptorProcessor, error) {
	err := checkArgsPeerAuthentication(args)
	if err != nil {
		return nil, err
	}

	return &peerAuthenticationInterceptorProcessor{
		peerAuthenticationCacher: args.PeerAuthenticationCacher,
		peerShardMapper:          args.PeerShardMapper,
		marshaller:               args.Marshaller,
		hardforkTrigger:          args.HardforkTrigger,
	}, nil
}

func checkArgsPeerAuthentication(args ArgPeerAuthenticationInterceptorProcessor) error {
	if check.IfNil(args.PeerAuthenticationCacher) {
		return process.ErrNilPeerAuthenticationCacher
	}
	if check.IfNil(args.PeerShardMapper) {
		return process.ErrNilPeerShardMapper
	}
	if check.IfNil(args.Marshaller) {
		return heartbeat.ErrNilMarshaller
	}
	if check.IfNil(args.HardforkTrigger) {
		return heartbeat.ErrNilHardforkTrigger
	}

	return nil
}

// Validate checks if the intercepted data can be processed
// returns nil as proper validity checks are done at intercepted data level
func (paip *peerAuthenticationInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted peer authentication inside the peer authentication cacher
func (paip *peerAuthenticationInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, _ string) error {
	interceptedPeerAuthenticationData, ok := data.(interceptedPeerAuthenticationMessageHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	payloadBuff := interceptedPeerAuthenticationData.Payload()
	payload := &heartbeat.Payload{}
	err := paip.marshaller.Unmarshal(payload, payloadBuff)
	if err != nil {
		return err
	}

	isHardforkTrigger, err := paip.hardforkTrigger.TriggerReceived(nil, []byte(payload.HardforkMessage), interceptedPeerAuthenticationData.Pubkey())
	if isHardforkTrigger {
		return err
	}

	return paip.updatePeerInfo(interceptedPeerAuthenticationData.Message(), interceptedPeerAuthenticationData.SizeInBytes())
}

func (paip *peerAuthenticationInterceptorProcessor) updatePeerInfo(message interface{}, messageSize int) error {
	peerAuthenticationData, ok := message.(*heartbeat.PeerAuthentication)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	pidBytes := peerAuthenticationData.GetPid()
	paip.peerAuthenticationCacher.Put(peerAuthenticationData.Pubkey, message, messageSize)
	paip.peerShardMapper.UpdatePeerIDPublicKeyPair(core.PeerID(pidBytes), peerAuthenticationData.GetPubkey())

	log.Trace("PeerAuthentication message saved")

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
