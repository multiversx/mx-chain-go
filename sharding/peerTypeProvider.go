package sharding

import (
	"bytes"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
)

// TODO: move this component somewhere else and write tests for it

// PeerTypeProvider handles the computation of a peer type
type PeerTypeProvider struct {
	nodesCoordinator NodesCoordinator
	epochHandler     EpochHandler
	cache            map[string]core.PeerType
	mutCache         sync.RWMutex
}

// NewPeerTypeProvider will return a new instance of PeerTypeProvider
func NewPeerTypeProvider(
	nodesCoordinator NodesCoordinator,
	epochHandler EpochHandler,
	epochStartNotifier EpochStartEventNotifier,
) (*PeerTypeProvider, error) {
	if check.IfNil(nodesCoordinator) {
		return nil, ErrNilNodesCoordinator
	}
	if check.IfNil(epochHandler) {
		return nil, ErrNilEpochHandler
	}

	ptp := &PeerTypeProvider{
		nodesCoordinator: nodesCoordinator,
		epochHandler:     epochHandler,
		cache:            make(map[string]core.PeerType),
		mutCache:         sync.RWMutex{},
	}

	err := ptp.populateCache(epochHandler.Epoch())
	if err != nil {
		return nil, err
	}

	epochStartNotifier.RegisterHandler(ptp.epochStartEventHandler())

	return ptp, nil
}

func (ptp *PeerTypeProvider) populateCache(epoch uint32) error {
	ptp.mutCache.Lock()
	defer ptp.mutCache.Unlock()

	ptp.cache = make(map[string]core.PeerType)

	nodesMapEligible, err := ptp.nodesCoordinator.GetEligiblePublicKeysPerShard(epoch)
	if err != nil {
		return err
	}

	for _, eligibleValidatorsInShard := range nodesMapEligible {
		for _, val := range eligibleValidatorsInShard {
			ptp.cache[string(val)] = core.EligibleList
		}
	}

	nodesMapWaiting, err := ptp.nodesCoordinator.GetWaitingPublicKeysPerShard(epoch)
	if err != nil {
		return err
	}

	for _, waitingValidatorsInShard := range nodesMapWaiting {
		for _, val := range waitingValidatorsInShard {
			ptp.cache[string(val)] = core.WaitingList
		}
	}

	return nil
}

// ComputeForPubKey returns the peer type for a given public key and shard id
func (ptp *PeerTypeProvider) ComputeForPubKey(pubKey []byte, shardID uint32) (core.PeerType, error) {
	ptp.mutCache.RLock()
	peerType, ok := ptp.cache[string(pubKey)]
	ptp.mutCache.RUnlock()
	if ok {
		return peerType, nil
	}

	return ptp.computeFromMaps(pubKey, shardID)
}

func (ptp *PeerTypeProvider) computeFromMaps(pubKey []byte, shardID uint32) (core.PeerType, error) {
	peerType := core.ObserverList
	nodesMapEligible, err := ptp.nodesCoordinator.GetEligiblePublicKeysPerShard(ptp.epochHandler.Epoch())
	if err != nil {
		return "", err
	}

	eligibleValidatorsInShard := nodesMapEligible[shardID]
	for _, val := range eligibleValidatorsInShard {
		if bytes.Equal(val, pubKey) {
			peerType = core.EligibleList
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
			peerType = core.WaitingList
			break
		}
	}

	ptp.mutCache.Lock()
	ptp.cache[string(pubKey)] = peerType
	ptp.mutCache.Unlock()

	return peerType, nil
}

func (ptp *PeerTypeProvider) epochStartEventHandler() epochStart.ActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {
		err := ptp.populateCache(hdr.GetEpoch())
		if err != nil {
			log.Warn("change epoch in peer type provider", "error", err.Error())
		}
	}, func(_ data.HeaderHandler) {})

	return subscribeHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (ptp *PeerTypeProvider) IsInterfaceNil() bool {
	return ptp == nil
}
