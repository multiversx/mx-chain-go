package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/heartbeat"
)

const minDurationInSec = 10

type interceptedPeerAuthenticationDataFactory struct {
	marshalizer           marshal.Marshalizer
	nodesCoordinator      heartbeat.NodesCoordinator
	signaturesHandler     heartbeat.SignaturesHandler
	peerSignatureHandler  crypto.PeerSignatureHandler
	expiryTimespanInSec   int64
	hardforkTriggerPubKey []byte
}

// NewInterceptedPeerAuthenticationDataFactory creates an instance of interceptedPeerAuthenticationDataFactory
func NewInterceptedPeerAuthenticationDataFactory(arg ArgInterceptedDataFactory) (*interceptedPeerAuthenticationDataFactory, error) {
	err := checkArgInterceptedDataFactory(arg)
	if err != nil {
		return nil, err
	}

	return &interceptedPeerAuthenticationDataFactory{
		marshalizer:           arg.CoreComponents.InternalMarshalizer(),
		nodesCoordinator:      arg.NodesCoordinator,
		signaturesHandler:     arg.SignaturesHandler,
		peerSignatureHandler:  arg.PeerSignatureHandler,
		expiryTimespanInSec:   arg.HeartbeatExpiryTimespanInSec,
		hardforkTriggerPubKey: arg.CoreComponents.HardforkTriggerPubKey(),
	}, nil
}

func checkArgInterceptedDataFactory(args ArgInterceptedDataFactory) error {
	if check.IfNil(args.CoreComponents) {
		return process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.CoreComponents.InternalMarshalizer()) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.NodesCoordinator) {
		return process.ErrNilNodesCoordinator
	}
	if check.IfNil(args.SignaturesHandler) {
		return process.ErrNilSignaturesHandler
	}
	if check.IfNil(args.PeerSignatureHandler) {
		return process.ErrNilPeerSignatureHandler
	}
	if args.HeartbeatExpiryTimespanInSec < minDurationInSec {
		return process.ErrInvalidExpiryTimespan
	}
	if len(args.CoreComponents.HardforkTriggerPubKey()) == 0 {
		return fmt.Errorf("%w hardfork trigger public key bytes length is 0", process.ErrInvalidValue)
	}

	return nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (ipadf *interceptedPeerAuthenticationDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	arg := heartbeat.ArgInterceptedPeerAuthentication{
		ArgBaseInterceptedHeartbeat: heartbeat.ArgBaseInterceptedHeartbeat{
			DataBuff:   buff,
			Marshaller: ipadf.marshalizer,
		},
		NodesCoordinator:      ipadf.nodesCoordinator,
		SignaturesHandler:     ipadf.signaturesHandler,
		PeerSignatureHandler:  ipadf.peerSignatureHandler,
		ExpiryTimespanInSec:   ipadf.expiryTimespanInSec,
		HardforkTriggerPubKey: ipadf.hardforkTriggerPubKey,
	}

	return heartbeat.NewInterceptedPeerAuthentication(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ipadf *interceptedPeerAuthenticationDataFactory) IsInterfaceNil() bool {
	return ipadf == nil
}
