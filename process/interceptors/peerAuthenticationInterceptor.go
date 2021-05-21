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

// ArgPeerAuthenticationInterceptor is the argument for the peer authentication interceptor
type ArgPeerAuthenticationInterceptor struct {
	ArgSingleDataInterceptor
	Marshalizer             marshal.Marshalizer
	ValidatorChecker        process.ValidatorChecker
	AuthenticationProcessor process.PeerAuthenticationProcessor
}

type peerAuthenticationInterceptor struct {
	*baseDataInterceptor
	validatorChecker            process.ValidatorChecker
	peerAuthenticationProcessor process.PeerAuthenticationProcessor
}

// NewPeerAuthenticationInterceptor hooks a new interceptor for packed multi data containing peer authentication instances
func NewPeerAuthenticationInterceptor(arg ArgPeerAuthenticationInterceptor) (*peerAuthenticationInterceptor, error) {
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
	if check.IfNil(arg.AuthenticationProcessor) {
		return nil, process.ErrNilAuthenticationProcessor
	}

	interceptor := &peerAuthenticationInterceptor{
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
		validatorChecker:            arg.ValidatorChecker,
		peerAuthenticationProcessor: arg.AuthenticationProcessor,
	}

	return interceptor, nil
}

// ProcessReceivedMessage is the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (pai *peerAuthenticationInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, fromConnectedPeer core.PeerID) error {
	multiDataBuff, err := pai.preProcessMessage(message, fromConnectedPeer)
	if err != nil {
		return err
	}

	observerKeyFound := false
	for _, dataBuff := range multiDataBuff {
		var interceptedData process.InterceptedData
		interceptedData, err = pai.interceptedData(dataBuff, message.Peer(), fromConnectedPeer)
		if err != nil {
			pai.throttler.EndProcessing()
			return err
		}

		peerAuth, ok := interceptedData.(process.InterceptedPeerAuthentication)
		if !ok {
			//intercepted data is not of type interceptedPeerInfo
			cause := "intercepted data is not of type process.InterceptedPeerInfo"
			pai.blackListPeers(cause, nil, message.Peer(), fromConnectedPeer)
			pai.throttler.EndProcessing()

			return errors.New(cause)
		}

		var shardID uint32
		_, shardID, err = pai.validatorChecker.GetValidatorWithPublicKey(peerAuth.PublicKey())
		isObserver := err != nil
		if isObserver {
			observerKeyFound = true
			continue
		}

		peerAuth.SetComputedShardID(shardID)
		errProcess := pai.peerAuthenticationProcessor.Process(peerAuth)
		if errProcess != nil {
			pai.throttler.EndProcessing()
			pai.blackListPeers("peer info processing error", errProcess, message.Peer(), fromConnectedPeer)
			return errProcess
		}
	}
	pai.throttler.EndProcessing()
	if observerKeyFound {
		return process.ErrPeerAuthenticationForObservers
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pai *peerAuthenticationInterceptor) IsInterfaceNil() bool {
	return pai == nil
}
