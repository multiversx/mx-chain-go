package process

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessageProcessor is the struct that will handle heartbeat message verifications and conversion between
// heartbeatMessageInfo and HeartbeatDTO
type MessageProcessor struct {
	peerSignatureHandler     crypto.PeerSignatureHandler
	marshalizer              marshal.Marshalizer
	networkShardingCollector heartbeat.NetworkShardingCollector
}

// NewMessageProcessor will return a new instance of MessageProcessor
func NewMessageProcessor(
	peerSignatureHandler crypto.PeerSignatureHandler,
	marshalizer marshal.Marshalizer,
	networkShardingCollector heartbeat.NetworkShardingCollector,
) (*MessageProcessor, error) {
	if check.IfNil(peerSignatureHandler) {
		return nil, heartbeat.ErrNilPeerSignatureHandler
	}
	if check.IfNil(marshalizer) {
		return nil, heartbeat.ErrNilMarshalizer
	}
	if check.IfNil(networkShardingCollector) {
		return nil, heartbeat.ErrNilNetworkShardingCollector
	}

	return &MessageProcessor{
		peerSignatureHandler:     peerSignatureHandler,
		marshalizer:              marshalizer,
		networkShardingCollector: networkShardingCollector,
	}, nil
}

// CreateHeartbeatFromP2PMessage will return a heartbeat if all the checks pass
func (mp *MessageProcessor) CreateHeartbeatFromP2PMessage(message p2p.MessageP2P) (*data.Heartbeat, error) {
	if check.IfNil(message) {
		return nil, heartbeat.ErrNilMessage
	}
	if message.Data() == nil {
		return nil, heartbeat.ErrNilDataToProcess
	}

	hbRecv := &data.Heartbeat{}

	err := mp.marshalizer.Unmarshal(hbRecv, message.Data())
	if err != nil {
		return nil, err
	}

	err = verifyLengths(hbRecv)
	if err != nil {
		return nil, err
	}

	err = mp.peerSignatureHandler.VerifyPeerSignature(hbRecv.Pubkey, core.PeerID(hbRecv.Pid), hbRecv.Signature)
	if err != nil {
		return nil, err
	}

	mp.networkShardingCollector.UpdatePeerIdPublicKey(message.Peer(), hbRecv.Pubkey)
	//add into the last failsafe map. Useful for observers.
	mp.networkShardingCollector.UpdatePeerIdShardId(message.Peer(), hbRecv.ShardID)
	mp.networkShardingCollector.UpdatePeerIdSubType(message.Peer(), core.P2PPeerSubType(hbRecv.PeerSubType))

	return hbRecv, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mp *MessageProcessor) IsInterfaceNil() bool {
	return mp == nil
}
