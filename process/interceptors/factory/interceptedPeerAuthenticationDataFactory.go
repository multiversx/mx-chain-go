package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/heartbeat"
)

const minDurationInSec = 10

type interceptedPeerAuthenticationDataFactory struct {
	marshalizer          marshal.Marshalizer
	nodesCoordinator     heartbeat.NodesCoordinator
	signaturesHandler    heartbeat.SignaturesHandler
	peerSignatureHandler crypto.PeerSignatureHandler
	expiryTimespanInSec  int64
}

// NewInterceptedPeerAuthenticationDataFactory creates an instance of interceptedPeerAuthenticationDataFactory
func NewInterceptedPeerAuthenticationDataFactory(arg ArgInterceptedDataFactory) (*interceptedPeerAuthenticationDataFactory, error) {
	if check.IfNil(arg.CoreComponents) {
		return nil, process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(arg.CoreComponents.InternalMarshalizer()) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(arg.NodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
	}
	if check.IfNil(arg.SignaturesHandler) {
		return nil, process.ErrNilSignaturesHandler
	}
	if check.IfNil(arg.PeerSignatureHandler) {
		return nil, process.ErrNilPeerSignatureHandler
	}
	if arg.HeartbeatExpiryTimespanInSec < minDurationInSec {
		return nil, process.ErrInvalidExpiryTimespan
	}

	return &interceptedPeerAuthenticationDataFactory{
		marshalizer:          arg.CoreComponents.InternalMarshalizer(),
		nodesCoordinator:     arg.NodesCoordinator,
		signaturesHandler:    arg.SignaturesHandler,
		peerSignatureHandler: arg.PeerSignatureHandler,
		expiryTimespanInSec:  arg.HeartbeatExpiryTimespanInSec,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (ipadf *interceptedPeerAuthenticationDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	arg := heartbeat.ArgInterceptedPeerAuthentication{
		ArgBaseInterceptedHeartbeat: heartbeat.ArgBaseInterceptedHeartbeat{
			DataBuff:    buff,
			Marshalizer: ipadf.marshalizer,
		},
		NodesCoordinator:     ipadf.nodesCoordinator,
		SignaturesHandler:    ipadf.signaturesHandler,
		PeerSignatureHandler: ipadf.peerSignatureHandler,
		ExpiryTimespanInSec:  ipadf.expiryTimespanInSec,
	}

	return heartbeat.NewInterceptedPeerAuthentication(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ipadf *interceptedPeerAuthenticationDataFactory) IsInterfaceNil() bool {
	return ipadf == nil
}
