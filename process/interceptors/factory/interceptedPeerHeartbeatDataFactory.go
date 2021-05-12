package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/heartbeat"
)

type interceptedPeerHeartbeatDataFactory struct {
	marshalizer          marshal.Marshalizer
	peerSignatureHandler crypto.PeerSignatureHandler
	hasher               hashing.Hasher
}

// NewInterceptedPeerHeartbeatDataFactory creates an instance of interceptedPeerHeartbeatDataFactory
func NewInterceptedPeerHeartbeatDataFactory(
	argument *ArgInterceptedDataFactory,
) (*interceptedPeerHeartbeatDataFactory, error) {
	if argument == nil {
		return nil, process.ErrNilArgumentStruct
	}
	if check.IfNil(argument.CoreComponents) {
		return nil, process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(argument.CoreComponents.InternalMarshalizer()) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(argument.CoreComponents.Hasher()) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(argument.PeerSignatureHandler) {
		return nil, process.ErrNilPeerSignatureHandler
	}

	return &interceptedPeerHeartbeatDataFactory{
		marshalizer:          argument.CoreComponents.InternalMarshalizer(),
		hasher:               argument.CoreComponents.Hasher(),
		peerSignatureHandler: argument.PeerSignatureHandler,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (iphdf *interceptedPeerHeartbeatDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	arg := heartbeat.ArgInterceptedPeerHeartbeat{
		DataBuff:             buff,
		Marshalizer:          iphdf.marshalizer,
		PeerSignatureHandler: iphdf.peerSignatureHandler,
		Hasher:               iphdf.hasher,
	}

	return heartbeat.NewInterceptedPeerHeartbeat(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (iphdf *interceptedPeerHeartbeatDataFactory) IsInterfaceNil() bool {
	return iphdf == nil
}
