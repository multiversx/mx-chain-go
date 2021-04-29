package interceptors

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/debug/resolver"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ArgPeerHeartbeatInterceptor is the argument for the peer heartbeat interceptor
type ArgPeerHeartbeatInterceptor struct {
	ArgSingleDataInterceptor
	Marshalizer         marshal.Marshalizer
	ValidatorChecker    process.ValidatorChecker
	ProcessingThrottler process.InterceptorThrottler
	HeartbeatProcessor  process.PeerHeartbeatProcessor
}

type peerHeartbeatInterceptor struct {
	*baseDataInterceptor
	validatorChecker                 process.ValidatorChecker
	peerHeartbeatProcessingThrottler process.InterceptorThrottler
	peerHeartbeatProcessor           process.PeerHeartbeatProcessor
}

// NewPeerHeartbeatInterceptor hooks a new interceptor for packed multi data containing peer heartbeat instances
func NewPeerHeartbeatInterceptor(arg ArgPeerHeartbeatInterceptor) (*peerHeartbeatInterceptor, error) {
	err := checkArguments(arg.ArgSingleDataInterceptor)
	if err != nil {
		return nil, err
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(arg.ValidatorChecker) {
		return nil, process.ErrNilValidatorChecker
	}
	if check.IfNil(arg.ProcessingThrottler) {
		return nil, process.ErrNilProcessingThrottler
	}
	if check.IfNil(arg.HeartbeatProcessor) {
		return nil, process.ErrNilHeartbeatProcessor
	}

	interceptor := &peerHeartbeatInterceptor{
		baseDataInterceptor: &baseDataInterceptor{
			throttler:        arg.Throttler,
			antifloodHandler: arg.AntifloodHandler,
			topic:            arg.Topic,
			currentPeerId:    arg.CurrentPeerId,
			processor:        arg.Processor,
			debugHandler:     resolver.NewDisabledInterceptorResolver(),
			marshalizer:      arg.Marshalizer,
			factory:          arg.DataFactory,
		},
		validatorChecker:                 arg.ValidatorChecker,
		peerHeartbeatProcessingThrottler: arg.ProcessingThrottler,
		peerHeartbeatProcessor:           arg.HeartbeatProcessor,
	}

	return interceptor, nil
}

// ProcessReceivedMessage is the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (phi *peerHeartbeatInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	multiDataBuff, err := phi.preProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	peerHbs := make([]process.InterceptedPeerHeartbeat, 0, len(multiDataBuff))
	for _, dataBuff := range multiDataBuff {
		var interceptedData process.InterceptedData
		interceptedData, err = phi.interceptedData(dataBuff, message.Peer(), fromConnectedPeer)
		if err != nil {
			phi.throttler.EndProcessing()
			return err
		}

		peerHb, ok := interceptedData.(process.InterceptedPeerHeartbeat)
		if !ok {
			//intercepted data is not of type interceptedPeerInfo
			cause := "intercepted data is not of type process.InterceptedPeerInfo"
			phi.blackListPeers(cause, nil, message.Peer(), fromConnectedPeer)
			phi.throttler.EndProcessing()

			return errors.New(cause)
		}

		var shardID uint32
		_, shardID, err = phi.validatorChecker.GetValidatorWithPublicKey(peerHb.PublicKey())
		isSkippableObserver := err != nil && !phi.peerHeartbeatProcessingThrottler.CanProcess()
		if isSkippableObserver {
			continue
		}

		//validators will always get the chance to get their message processed despite the number of observers messages
		//being broadcast
		phi.peerHeartbeatProcessingThrottler.StartProcessing()
		peerHb.SetShardID(shardID)
		peerHbs = append(peerHbs, peerHb)
	}

	go func(originator, fromConnectedPeer core.PeerID) {
		for _, pi := range peerHbs {
			errProcess := phi.peerHeartbeatProcessor.Process(pi)
			if errProcess != nil {
				phi.blackListPeers("peer info processing error", errProcess, originator, fromConnectedPeer)
			}

			phi.peerHeartbeatProcessingThrottler.EndProcessing()
		}
		phi.throttler.EndProcessing()
	}(message.Peer(), fromConnectedPeer)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (phi *peerHeartbeatInterceptor) IsInterfaceNil() bool {
	return phi == nil
}
