package sender

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

type commonHeartbeatSender struct {
	baseSender
	currentBlockProvider heartbeat.CurrentBlockProvider
}

func (chs *commonHeartbeatSender) generateMessageBytes(
	versionNumber string,
	nodeDisplayName string,
	identity string,
	peerSubType uint32,
) ([]byte, error) {
	if len(versionNumber) > maxSizeInBytes {
		return nil, fmt.Errorf("%w for versionNumber, received %s of size %d, max size allowed %d",
			heartbeat.ErrPropertyTooLong, versionNumber, len(versionNumber), maxSizeInBytes)
	}
	if len(nodeDisplayName) > maxSizeInBytes {
		return nil, fmt.Errorf("%w for nodeDisplayName, received %s of size %d, max size allowed %d",
			heartbeat.ErrPropertyTooLong, nodeDisplayName, len(nodeDisplayName), maxSizeInBytes)
	}
	if len(identity) > maxSizeInBytes {
		return nil, fmt.Errorf("%w for identity, received %s of size %d, max size allowed %d",
			heartbeat.ErrPropertyTooLong, identity, len(identity), maxSizeInBytes)
	}

	payload := &heartbeat.Payload{
		Timestamp:       time.Now().Unix(),
		HardforkMessage: "", // sent through peer authentication message
	}
	payloadBytes, err := chs.marshaller.Marshal(payload)
	if err != nil {
		return nil, err
	}

	nonce := uint64(0)
	currentBlock := chs.currentBlockProvider.GetCurrentBlockHeader()
	if currentBlock != nil {
		nonce = currentBlock.GetNonce()
	}

	msg := &heartbeat.HeartbeatV2{
		Payload:         payloadBytes,
		VersionNumber:   versionNumber,
		NodeDisplayName: nodeDisplayName,
		Identity:        identity,
		Nonce:           nonce,
		PeerSubType:     peerSubType,
	}

	return chs.marshaller.Marshal(msg)
}
