package process

import (
	"bytes"

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
	singleSigner             crypto.SingleSigner
	keygen                   crypto.KeyGenerator
	marshalizer              marshal.Marshalizer
	networkShardingCollector heartbeat.NetworkShardingCollector
}

// NewMessageProcessor will return a new instance of MessageProcessor
func NewMessageProcessor(
	singleSigner crypto.SingleSigner,
	keygen crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
	networkShardingCollector heartbeat.NetworkShardingCollector,
) (*MessageProcessor, error) {
	if check.IfNil(singleSigner) {
		return nil, heartbeat.ErrNilSingleSigner
	}
	if check.IfNil(keygen) {
		return nil, heartbeat.ErrNilKeyGenerator
	}
	if check.IfNil(marshalizer) {
		return nil, heartbeat.ErrNilMarshalizer
	}
	if check.IfNil(networkShardingCollector) {
		return nil, heartbeat.ErrNilNetworkShardingCollector
	}

	return &MessageProcessor{
		singleSigner:             singleSigner,
		keygen:                   keygen,
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

	err = mp.verifySignature(hbRecv)
	if err != nil {
		return nil, err
	}

	mp.networkShardingCollector.UpdatePeerIdPublicKey(message.Peer(), hbRecv.Pubkey)
	//add into the last failsafe map. Useful for observers.
	mp.networkShardingCollector.UpdatePeerIdShardId(message.Peer(), hbRecv.ShardID)

	return hbRecv, nil
}

func (mp *MessageProcessor) verifySignature(hbRecv *data.Heartbeat) error {
	senderPubKey, err := mp.keygen.PublicKeyFromByteArray(hbRecv.Pubkey)
	if err != nil {
		return err
	}

	pid, sig := mp.networkShardingCollector.GetPidAndSignatureFromPk(hbRecv.Pubkey)
	if pid == nil || sig == nil {
		err = mp.singleSigner.Verify(senderPubKey, hbRecv.Pid, hbRecv.Signature)
		if err != nil {
			return err
		}

		mp.networkShardingCollector.UpdatePublicKeyPIDSignature(hbRecv.Pubkey, hbRecv.Pid, hbRecv.Signature)
		return nil
	}

	if !bytes.Equal(pid, hbRecv.Pid) {
		return heartbeat.ErrPIDMissmatch
	}

	if !bytes.Equal(sig, hbRecv.Signature) {
		return heartbeat.ErrSignatureMissmatch
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mp *MessageProcessor) IsInterfaceNil() bool {
	return mp == nil
}
