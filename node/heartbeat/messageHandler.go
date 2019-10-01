package heartbeat

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MessageHandler is the struct that will handle heartbeat message verifications and conversion between
// heartbeatMessageInfo and HeartbeatDTO
type MessageHandler struct {
	singleSigner         crypto.SingleSigner
	keygen               crypto.KeyGenerator
	marshalizer          marshal.Marshalizer
	mutHeartbeatMessages *sync.RWMutex
}

// NewMessageHandler will return a new instance of MessageHandler
func NewMessageHandler(
	singleSigner crypto.SingleSigner,
	keygen crypto.KeyGenerator,
	marshalizer marshal.Marshalizer,
) *MessageHandler {
	return &MessageHandler{
		singleSigner:         singleSigner,
		keygen:               keygen,
		marshalizer:          marshalizer,
		mutHeartbeatMessages: &sync.RWMutex{},
	}
}

// CreateHeartbeatFromP2pMessage will return a heartbeat if all the checks pass
func (hmh *MessageHandler) CreateHeartbeatFromP2pMessage(message p2p.MessageP2P) (*Heartbeat, error) {
	if message == nil || message.IsInterfaceNil() {
		return nil, ErrNilMessage
	}
	if message.Data() == nil {
		return nil, ErrNilDataToProcess
	}

	hbRecv := &Heartbeat{}

	err := hmh.marshalizer.Unmarshal(hbRecv, message.Data())
	if err != nil {
		return nil, err
	}

	err = hmh.verifySignature(hbRecv)
	if err != nil {
		return nil, err
	}

	return hbRecv, nil
}

func (hmh *MessageHandler) verifySignature(hbRecv *Heartbeat) error {
	senderPubKey, err := hmh.keygen.PublicKeyFromByteArray(hbRecv.Pubkey)
	if err != nil {
		return err
	}

	copiedHeartbeat := *hbRecv
	copiedHeartbeat.Signature = nil
	buffCopiedHeartbeat, err := hmh.marshalizer.Marshal(copiedHeartbeat)
	if err != nil {
		return err
	}

	return hmh.singleSigner.Verify(senderPubKey, buffCopiedHeartbeat, hbRecv.Signature)
}

func (hmh *MessageHandler) convertToExportedStruct(v *heartbeatMessageInfo) HeartbeatDTO {
	return HeartbeatDTO{
		TimeStamp:          v.timeStamp,
		MaxInactiveTime:    v.maxInactiveTime,
		IsActive:           v.isActive,
		ReceivedShardID:    v.receivedShardID,
		ComputedShardID:    v.computedShardID,
		TotalUpTime:        v.totalUpTime,
		TotalDownTime:      v.totalDownTime,
		VersionNumber:      v.versionNumber,
		IsValidator:        v.isValidator,
		NodeDisplayName:    v.nodeDisplayName,
		LastUptimeDowntime: v.lastUptimeDowntime,
	}
}

func (hmh *MessageHandler) convertFromExportedStruct(hbDTO HeartbeatDTO, maxDuration time.Duration) heartbeatMessageInfo {
	hbmi := heartbeatMessageInfo{
		maxDurationPeerUnresponsive: maxDuration,
		maxInactiveTime:             hbDTO.MaxInactiveTime,
		timeStamp:                   hbDTO.TimeStamp,
		isActive:                    hbDTO.IsActive,
		totalUpTime:                 hbDTO.TotalUpTime,
		totalDownTime:               hbDTO.TotalDownTime,
		receivedShardID:             hbDTO.ReceivedShardID,
		computedShardID:             hbDTO.ComputedShardID,
		versionNumber:               hbDTO.VersionNumber,
		nodeDisplayName:             hbDTO.NodeDisplayName,
		isValidator:                 hbDTO.IsValidator,
		lastUptimeDowntime:          hbDTO.LastUptimeDowntime,
	}

	return hbmi
}
