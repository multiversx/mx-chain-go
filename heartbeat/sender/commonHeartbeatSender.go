package sender

import (
	"fmt"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/heartbeat"
)

type commonHeartbeatSender struct {
	baseSender
	versionNumber        string
	nodeDisplayName      string
	identity             string
	peerSubType          core.P2PPeerSubType
	currentBlockProvider heartbeat.CurrentBlockProvider
	peerTypeProvider     heartbeat.PeerTypeProviderHandler
}

func (chs *commonHeartbeatSender) generateMessageBytes(
	versionNumber string,
	nodeDisplayName string,
	identity string,
	peerSubType uint32,
	pkBytes []byte,
	numProcessedTrieNodes uint64,
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
		Payload:            payloadBytes,
		VersionNumber:      versionNumber,
		NodeDisplayName:    nodeDisplayName,
		Identity:           identity,
		Nonce:              nonce,
		PeerSubType:        peerSubType,
		Pubkey:             pkBytes,
		NumTrieNodesSynced: numProcessedTrieNodes,
	}

	return chs.marshaller.Marshal(msg)
}

// GetCurrentNodeType will return the current sender type and subtype
func (chs *commonHeartbeatSender) GetCurrentNodeType() (string, core.P2PPeerSubType, error) {
	_, pk := chs.getCurrentPrivateAndPublicKeys()
	pkBytes, err := pk.ToByteArray()
	if err != nil {
		return "", 0, err
	}

	peerType := chs.computePeerList(pkBytes)

	return peerType, chs.peerSubType, nil
}

func (chs *commonHeartbeatSender) computePeerList(pubkey []byte) string {
	peerType, _, err := chs.peerTypeProvider.ComputeForPubKey(pubkey)
	if err != nil {
		log.Warn("heartbeatSender: compute peer type", "error", err)
		return string(common.ObserverList)
	}

	return string(peerType)
}
