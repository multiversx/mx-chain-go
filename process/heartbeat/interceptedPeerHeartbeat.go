package heartbeat

import (
	"fmt"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
)

type interceptedPeerHeartbeat struct {
	peerHeartbeat heartbeat.PeerHeartbeat
	peerId        core.PeerID
	hash          []byte
}

func (iph *interceptedPeerHeartbeat) PublicKey() []byte {
	panic("implement me")
}

func (iph *interceptedPeerHeartbeat) SetShardID(shardId uint32) {
	panic("implement me")
}

// CheckValidity will check the validity of the received peer heartbeat
// However, this will not perform the signature validation since that can be an attack vector
func (iph *interceptedPeerHeartbeat) CheckValidity() error {
	if len(iph.peerHeartbeat.Pubkey) == 0 {
		return process.ErrNilOrEmptyPublicKeyBytes
	}
	if len(iph.peerHeartbeat.Payload) == 0 {
		return process.ErrNilOrEmptyPayloadBytes
	}
	if len(iph.peerHeartbeat.Signature) == 0 {
		return process.ErrNilOrEmptySignatureBytes
	}
	if len(iph.peerHeartbeat.Pid) == 0 {
		return process.ErrNilOrEmptyPeerId
	}

	return nil
}

// IsForCurrentShard always returns true
func (iph *interceptedPeerHeartbeat) IsForCurrentShard() bool {
	return true
}

// Hash returns the hash of this intercepted peer heartbeat
func (iph *interceptedPeerHeartbeat) Hash() []byte {
	return iph.hash
}

// Type returns the type of this intercepted data
func (iph *interceptedPeerHeartbeat) Type() string {
	return "intercepted peer heartbeat"
}

// Identifiers returns the identifiers used in requests
func (iph *interceptedPeerHeartbeat) Identifiers() [][]byte {
	return [][]byte{iph.peerHeartbeat.Pubkey, iph.peerHeartbeat.Pid}
}

// String returns the transaction's most important fields as string
func (iph *interceptedPeerHeartbeat) String() string {
	return fmt.Sprintf("pk=%s, pid=%s, sig=%s, payload=%s, shardID=%d",
		logger.DisplayByteSlice(iph.peerHeartbeat.Pubkey),
		iph.peerId.Pretty(),
		logger.DisplayByteSlice(iph.peerHeartbeat.Signature),
		logger.DisplayByteSlice(iph.peerHeartbeat.Payload),
		iph.peerHeartbeat.ShardID,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (iph *interceptedPeerHeartbeat) IsInterfaceNil() bool {
	return iph == nil
}
