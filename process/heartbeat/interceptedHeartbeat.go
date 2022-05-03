package heartbeat

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process"
)

const uint32Size = 4
const uint64Size = 8

// ArgBaseInterceptedHeartbeat is the base argument used for messages
type ArgBaseInterceptedHeartbeat struct {
	DataBuff    []byte
	Marshalizer marshal.Marshalizer
}

// ArgInterceptedHeartbeat is the argument used in the intercepted heartbeat constructor
type ArgInterceptedHeartbeat struct {
	ArgBaseInterceptedHeartbeat
	PeerId core.PeerID
}

// interceptedHeartbeat is a wrapper over HeartbeatV2
type interceptedHeartbeat struct {
	heartbeat heartbeat.HeartbeatV2
	payload   heartbeat.Payload
	peerId    core.PeerID
}

// NewInterceptedHeartbeat tries to create a new intercepted heartbeat instance
func NewInterceptedHeartbeat(arg ArgInterceptedHeartbeat) (*interceptedHeartbeat, error) {
	err := checkBaseArg(arg.ArgBaseInterceptedHeartbeat)
	if err != nil {
		return nil, err
	}
	if len(arg.PeerId) == 0 {
		return nil, process.ErrEmptyPeerID
	}

	hb, payload, err := createHeartbeat(arg.Marshalizer, arg.DataBuff)
	if err != nil {
		return nil, err
	}

	intercepted := &interceptedHeartbeat{
		heartbeat: *hb,
		payload:   *payload,
		peerId:    arg.PeerId,
	}

	return intercepted, nil
}

func checkBaseArg(arg ArgBaseInterceptedHeartbeat) error {
	if len(arg.DataBuff) == 0 {
		return process.ErrNilBuffer
	}
	if check.IfNil(arg.Marshalizer) {
		return process.ErrNilMarshalizer
	}
	return nil
}

func createHeartbeat(marshalizer marshal.Marshalizer, buff []byte) (*heartbeat.HeartbeatV2, *heartbeat.Payload, error) {
	hb := &heartbeat.HeartbeatV2{}
	err := marshalizer.Unmarshal(hb, buff)
	if err != nil {
		return nil, nil, err
	}
	payload := &heartbeat.Payload{}
	err = marshalizer.Unmarshal(payload, hb.Payload)
	if err != nil {
		return nil, nil, err
	}
	return hb, payload, nil
}

// CheckValidity will check the validity of the received peer heartbeat
func (ihb *interceptedHeartbeat) CheckValidity() error {
	err := verifyPropertyMinMaxLen(payloadProperty, ihb.heartbeat.Payload)
	if err != nil {
		return err
	}
	err = verifyPropertyMinMaxLen(versionNumberProperty, []byte(ihb.heartbeat.VersionNumber))
	if err != nil {
		return err
	}
	err = verifyPropertyMaxLen(nodeDisplayNameProperty, []byte(ihb.heartbeat.NodeDisplayName))
	if err != nil {
		return err
	}
	err = verifyPropertyMaxLen(identityProperty, []byte(ihb.heartbeat.Identity))
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

// Hash always returns an empty string
func (ihb *interceptedHeartbeat) Hash() []byte {
	return []byte("")
}

// Type returns the type of this intercepted data
func (ihb *interceptedHeartbeat) Type() string {
	return interceptedHeartbeatType
}

// Identifiers returns the identifiers used in requests
func (ihb *interceptedHeartbeat) Identifiers() [][]byte {
	return [][]byte{ihb.peerId.Bytes()}
}

// String returns the most important fields as string
func (ihb *interceptedHeartbeat) String() string {
	return fmt.Sprintf("pid=%s, version=%s, name=%s, identity=%s, nonce=%d, subtype=%d, payload=%s",
		ihb.peerId.Pretty(),
		ihb.heartbeat.VersionNumber,
		ihb.heartbeat.NodeDisplayName,
		ihb.heartbeat.Identity,
		ihb.heartbeat.Nonce,
		ihb.heartbeat.PeerSubType,
		logger.DisplayByteSlice(ihb.heartbeat.Payload))
}

// Message returns the heartbeat message
func (ihb *interceptedHeartbeat) Message() interface{} {
	return ihb.heartbeat
}

// SizeInBytes returns the size in bytes held by this instance
func (ihb *interceptedHeartbeat) SizeInBytes() int {
	return len(ihb.heartbeat.Payload) +
		len(ihb.heartbeat.VersionNumber) +
		len(ihb.heartbeat.NodeDisplayName) +
		len(ihb.heartbeat.Identity) +
		uint64Size + uint32Size
}

// IsInterfaceNil returns true if there is no value under the interface
func (ihb *interceptedHeartbeat) IsInterfaceNil() bool {
	return ihb == nil
}
