package heartbeat

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
)

// argBaseInterceptedHeartbeat is the base argument used for messages
type argBaseInterceptedHeartbeat struct {
	DataBuff    []byte
	Marshalizer marshal.Marshalizer
	Hasher      hashing.Hasher
}

// ArgInterceptedHeartbeat is the argument used in the intercepted heartbeat constructor
type ArgInterceptedHeartbeat struct {
	argBaseInterceptedHeartbeat
}

type interceptedHeartbeat struct {
	heartbeat heartbeat.HeartbeatV2
	hash      []byte
}

// NewInterceptedHeartbeat tries to create a new intercepted heartbeat instance
func NewInterceptedHeartbeat(arg ArgInterceptedHeartbeat) (*interceptedHeartbeat, error) {
	err := checkBaseArg(arg.argBaseInterceptedHeartbeat)
	if err != nil {
		return nil, err
	}

	hb, err := createHeartbeat(arg.Marshalizer, arg.DataBuff)
	if err != nil {
		return nil, err
	}

	intercepted := &interceptedHeartbeat{
		heartbeat: *hb,
	}
	intercepted.hash = arg.Hasher.Compute(string(arg.DataBuff))

	return intercepted, nil
}

func checkBaseArg(arg argBaseInterceptedHeartbeat) error {
	if len(arg.DataBuff) == 0 {
		return process.ErrNilBuffer
	}
	if check.IfNil(arg.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(arg.Hasher) {
		return process.ErrNilHasher
	}
	return nil
}

func createHeartbeat(marshalizer marshal.Marshalizer, buff []byte) (*heartbeat.HeartbeatV2, error) {
	hb := &heartbeat.HeartbeatV2{}
	err := marshalizer.Unmarshal(hb, buff)
	if err != nil {
		return nil, err
	}
	return hb, nil
}

// CheckValidity will check the validity of the received peer heartbeat
func (ihb *interceptedHeartbeat) CheckValidity() error {
	err := verifyPropertyLen(payloadProperty, ihb.heartbeat.Payload)
	if err != nil {
		return err
	}
	err = verifyPropertyLen(versionNumberProperty, []byte(ihb.heartbeat.VersionNumber))
	if err != nil {
		return err
	}
	err = verifyPropertyLen(nodeDisplayNameProperty, []byte(ihb.heartbeat.NodeDisplayName))
	if err != nil {
		return err
	}
	err = verifyPropertyLen(identityProperty, []byte(ihb.heartbeat.Identity))
	if err != nil {
		return err
	}
	if ihb.heartbeat.PeerSubType != uint32(core.RegularPeer) && ihb.heartbeat.PeerSubType != uint32(core.FullHistoryObserver) {
		return process.ErrInvalidPeerSubType
	}
	return nil
}

// IsForCurrentShard always returns true
func (ihb *interceptedHeartbeat) IsForCurrentShard() bool {
	return true
}

// Hash returns the hash of this intercepted heartbeat
func (ihb *interceptedHeartbeat) Hash() []byte {
	return ihb.hash
}

// Type returns the type of this intercepted data
func (ihb *interceptedHeartbeat) Type() string {
	return interceptedHeartbeatType
}

// Identifiers returns the identifiers used in requests
func (ihb *interceptedHeartbeat) Identifiers() [][]byte {
	return [][]byte{ihb.hash}
}

// String returns the most important fields as string
func (ihb *interceptedHeartbeat) String() string {
	return fmt.Sprintf("version=%s, name=%s, identity=%s, nonce=%d, subtype=%d, payload=%s",
		ihb.heartbeat.VersionNumber,
		ihb.heartbeat.NodeDisplayName,
		ihb.heartbeat.Identity,
		ihb.heartbeat.Nonce,
		ihb.heartbeat.PeerSubType,
		logger.DisplayByteSlice(ihb.heartbeat.Payload))
}

// IsInterfaceNil returns true if there is no value under the interface
func (ihb *interceptedHeartbeat) IsInterfaceNil() bool {
	return ihb == nil
}
