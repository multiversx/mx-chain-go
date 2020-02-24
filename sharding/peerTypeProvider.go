package sharding

import (
	"bytes"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
)

// PeerTypeProvider handles the computation of a peer type
type PeerTypeProvider struct {
	nodesCoordinator NodesCoordinator
	epochHandler     EpochHandler
}

// NewPeerTypeProvider will return a new instance of PeerTypeProvider
func NewPeerTypeProvider(
	nodesCoordinator NodesCoordinator,
	epochHandler EpochHandler,
) (*PeerTypeProvider, error) {
	if check.IfNil(nodesCoordinator) {
		return nil, ErrNilNodesCoordinator
	}
	if check.IfNil(epochHandler) {
		return nil, ErrNilEpochHandler
	}

	return &PeerTypeProvider{
		nodesCoordinator: nodesCoordinator,
		epochHandler:     epochHandler,
	}, nil
}

// ComputeForPubKey returns the peer type for a given public key and shard id
func (ptp *PeerTypeProvider) ComputeForPubKey(pubKey []byte, shardID uint32) (core.ValidatorList, error) {
	peerList := core.ObserverList
	nodesMapEligible, err := ptp.nodesCoordinator.GetEligiblePublicKeysPerShard(ptp.epochHandler.Epoch())
	if err != nil {
		return "", err
	}

	eligibleValidatorsInShard := nodesMapEligible[shardID]
	for _, validator := range eligibleValidatorsInShard {
		if bytes.Equal(validator, pubKey) {
			peerList = core.EligibleList
			break
		}
	}

	nodesMapWaiting, err := ptp.nodesCoordinator.GetWaitingPublicKeysPerShard(ptp.epochHandler.Epoch())
	if err != nil {
		return "", err
	}

	waitingValidatorsInShard := nodesMapWaiting[shardID]
	for _, peerPubKey := range waitingValidatorsInShard {
		if bytes.Equal(peerPubKey, pubKey) {
			peerList = core.WaitingList
			break
		}
	}

	return peerList, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ptp *PeerTypeProvider) IsInterfaceNil() bool {
	return ptp == nil
}
