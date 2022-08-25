package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/heartbeat"
)

type interceptedHeartbeatDataFactory struct {
	marshalizer marshal.Marshalizer
	peerID      core.PeerID
}

// NewInterceptedHeartbeatDataFactory creates an instance of interceptedHeartbeatDataFactory
func NewInterceptedHeartbeatDataFactory(arg ArgInterceptedDataFactory) (*interceptedHeartbeatDataFactory, error) {
	if check.IfNil(arg.CoreComponents.InternalMarshalizer()) {
		return nil, process.ErrNilMarshalizer
	}
	if len(arg.PeerID) == 0 {
		return nil, process.ErrEmptyPeerID
	}

	return &interceptedHeartbeatDataFactory{
		marshalizer: arg.CoreComponents.InternalMarshalizer(),
		peerID:      arg.PeerID,
	}, nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (ihdf *interceptedHeartbeatDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	arg := heartbeat.ArgInterceptedHeartbeat{
		ArgBaseInterceptedHeartbeat: heartbeat.ArgBaseInterceptedHeartbeat{
			DataBuff:   buff,
			Marshaller: ihdf.marshalizer,
		},
		PeerId: ihdf.peerID,
	}

	return heartbeat.NewInterceptedHeartbeat(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ihdf *interceptedHeartbeatDataFactory) IsInterfaceNil() bool {
	return ihdf == nil
}
