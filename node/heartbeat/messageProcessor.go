package heartbeat

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessageProcessor is the struct that will handle heartbeat message verifications and conversion between
// heartbeatMessageInfo and HeartbeatDTO
type MessageProcessor struct {
	singleSigner           crypto.SingleSigner
	keygen                 crypto.KeyGenerator
	marshalizer            marshal.Marshalizer
	networkShardingUpdater NetworkShardingUpdater
}

// NewMessageProcessor will return a new instance of MessageProcessor
func NewMessageProcessor(
	singleSigner crypto.SingleSigner,
	keygen crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
	networkShardingUpdater NetworkShardingUpdater,
) (*MessageProcessor, error) {
	if check.IfNil(singleSigner) {
		return nil, ErrNilSingleSigner
	}
	if check.IfNil(keygen) {
		return nil, ErrNilKeyGenerator
	}
	if check.IfNil(marshalizer) {
		return nil, ErrNilMarshalizer
	}
	if check.IfNil(networkShardingUpdater) {
		return nil, ErrNilNetworkShardingUpdater
	}

	return &MessageProcessor{
		singleSigner:           singleSigner,
		keygen:                 keygen,
		marshalizer:            marshalizer,
		networkShardingUpdater: networkShardingUpdater,
	}, nil
}

// CreateHeartbeatFromP2pMessage will return a heartbeat if all the checks pass
func (mp *MessageProcessor) CreateHeartbeatFromP2pMessage(message p2p.MessageP2P) (*Heartbeat, error) {
	if message == nil || message.IsInterfaceNil() {
		return nil, ErrNilMessage
	}
	if message.Data() == nil {
		return nil, ErrNilDataToProcess
	}

	hbRecv := &Heartbeat{}

	err := mp.marshalizer.Unmarshal(hbRecv, message.Data())
	if err != nil {
		return nil, err
	}

	err = verifyLengths(hbRecv)
	if err != nil {
		return nil, err
	}

	err = mp.verifySignature(hbRecv)
	if err != nil {
		return nil, err
	}

	mp.networkShardingUpdater.UpdatePeerIdPublicKey(message.Peer(), hbRecv.Pubkey)

	return hbRecv, nil
}

func (mp *MessageProcessor) verifySignature(hbRecv *Heartbeat) error {
	senderPubKey, err := mp.keygen.PublicKeyFromByteArray(hbRecv.Pubkey)
	if err != nil {
		return err
	}

	copiedHeartbeat := *hbRecv
	copiedHeartbeat.Signature = nil
	buffCopiedHeartbeat, err := mp.marshalizer.Marshal(copiedHeartbeat)
	if err != nil {
		return err
	}

	return mp.singleSigner.Verify(senderPubKey, buffCopiedHeartbeat, hbRecv.Signature)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mp *MessageProcessor) IsInterfaceNil() bool {
	if mp == nil {
		return true
	}
	return false
}
