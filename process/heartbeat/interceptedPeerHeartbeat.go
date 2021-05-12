package heartbeat

import (
	"fmt"
	"sync"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/heartbeat"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

const (
	minSizeInBytes    = 1
	maxSizeInBytes    = 128
	interceptedType   = "intercepted peer heartbeat"
	publicKeyProperty = "public key"
	signatureProperty = "signature"
	payloadProperty   = "payload"
	peerIdProperty    = "peer id"
)

// ArgInterceptedPeerHeartbeat is the argument used in the intercepted peer heartbeat constructor
type ArgInterceptedPeerHeartbeat struct {
	DataBuff             []byte
	Marshalizer          marshal.Marshalizer
	PeerSignatureHandler crypto.PeerSignatureHandler
	Hasher               hashing.Hasher
}

type interceptedPeerHeartbeat struct {
	peerHeartbeat        heartbeat.PeerHeartbeat
	peerId               core.PeerID
	hash                 []byte
	mutComputedShardID   sync.RWMutex
	computedShardID      uint32
	peerSignatureHandler crypto.PeerSignatureHandler
}

// NewInterceptedPeerHeartbeat tries to create a new intercepted peer heartbeat instance
func NewInterceptedPeerHeartbeat(arg ArgInterceptedPeerHeartbeat) (*interceptedPeerHeartbeat, error) {
	if len(arg.DataBuff) == 0 {
		return nil, process.ErrNilBuffer
	}
	if check.IfNil(arg.Marshalizer) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(arg.PeerSignatureHandler) {
		return nil, process.ErrNilPeerSignatureHandler
	}
	if check.IfNil(arg.Hasher) {
		return nil, process.ErrNilHasher
	}

	peerHeartbeat, err := createPeerHeartbeat(arg.Marshalizer, arg.DataBuff)
	if err != nil {
		return nil, err
	}

	intercepted := &interceptedPeerHeartbeat{
		peerHeartbeat:        *peerHeartbeat,
		peerSignatureHandler: arg.PeerSignatureHandler,
	}

	intercepted.processFields(arg.Hasher, arg.DataBuff)

	return intercepted, nil
}

func createPeerHeartbeat(marshalizer marshal.Marshalizer, buff []byte) (*heartbeat.PeerHeartbeat, error) {
	peerHeartbeat := &heartbeat.PeerHeartbeat{}
	err := marshalizer.Unmarshal(peerHeartbeat, buff)
	if err != nil {
		return nil, err
	}

	return peerHeartbeat, nil
}

func (iph *interceptedPeerHeartbeat) processFields(hasher hashing.Hasher, buff []byte) {
	iph.hash = hasher.Compute(string(buff))
	iph.peerId = core.PeerID(iph.peerHeartbeat.Pid)
}

// PublicKey returns the public key as byte slice
func (iph *interceptedPeerHeartbeat) PublicKey() []byte {
	return iph.peerHeartbeat.Pubkey
}

// SetComputedShardID sets the computed shard ID. Concurrency safe.
func (iph *interceptedPeerHeartbeat) SetComputedShardID(shardId uint32) {
	iph.mutComputedShardID.Lock()
	iph.computedShardID = shardId
	iph.mutComputedShardID.Unlock()
}

// CheckValidity will check the validity of the received peer heartbeat
// However, this will not perform the signature validation since that can be an attack vector
func (iph *interceptedPeerHeartbeat) CheckValidity() error {
	err := VerifyHeartbeatProperyLen(publicKeyProperty, iph.peerHeartbeat.Pubkey)
	if err != nil {
		return err
	}
	err = VerifyHeartbeatProperyLen(signatureProperty, iph.peerHeartbeat.Signature)
	if err != nil {
		return err
	}
	err = VerifyHeartbeatProperyLen(payloadProperty, iph.peerHeartbeat.Payload)
	if err != nil {
		return err
	}
	err = VerifyHeartbeatProperyLen(peerIdProperty, iph.peerId.Bytes())
	if err != nil {
		return err
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
	return interceptedType
}

// Identifiers returns the identifiers used in requests
func (iph *interceptedPeerHeartbeat) Identifiers() [][]byte {
	return [][]byte{iph.peerHeartbeat.Pubkey, iph.peerHeartbeat.Pid}
}

// String returns the transaction's most important fields as string
func (iph *interceptedPeerHeartbeat) String() string {
	iph.mutComputedShardID.RLock()
	defer iph.mutComputedShardID.RUnlock()

	return fmt.Sprintf("pk=%s, pid=%s, sig=%s, payload=%s, received shardID=%d, computed shardID=%d",
		logger.DisplayByteSlice(iph.peerHeartbeat.Pubkey),
		iph.peerId.Pretty(),
		logger.DisplayByteSlice(iph.peerHeartbeat.Signature),
		logger.DisplayByteSlice(iph.peerHeartbeat.Payload),
		iph.peerHeartbeat.ShardID,
		iph.computedShardID,
	)
}

// VerifyHeartbeatProperyLen returns an error if the provided value is longer than accepted by the network
func VerifyHeartbeatProperyLen(property string, value []byte) error {
	if len(value) > maxSizeInBytes {
		return fmt.Errorf("%w for %s", process.ErrPropertyTooLong, property)
	}
	if len(value) < minSizeInBytes {
		return fmt.Errorf("%w for %s", process.ErrPropertyTooShort, property)
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (iph *interceptedPeerHeartbeat) IsInterfaceNil() bool {
	return iph == nil
}
