package sharding

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
)

// TODO: move this component somewhere else and write tests for it

type peerListAndShard struct {
	pType  core.PeerType
	pShard uint32
}

// PeerTypeProvider handles the computation of a peer type
type PeerTypeProvider struct {
	nodesCoordinator          NodesCoordinator
	epochHandler              EpochHandler
	validatorsProvider        ValidatorsProvider
	cache                     map[string]*peerListAndShard
	lastCacheUpdate           time.Time
	cacheRefreshIntervalInSec float64
	isUpdating                bool
	mutCache                  sync.RWMutex
}

type ArgPeerTypeProvider struct {
	NodesCoordinator          NodesCoordinator
	EpochHandler              EpochHandler
	ValidatorsProvider        ValidatorsProvider
	EpochStartEventNotifier   EpochStartEventNotifier
	CacheRefreshIntervalInSec uint32
}

// NewPeerTypeProvider will return a new instance of PeerTypeProvider
func NewPeerTypeProvider(arg ArgPeerTypeProvider) (*PeerTypeProvider, error) {
	if check.IfNil(arg.NodesCoordinator) {
		return nil, ErrNilNodesCoordinator
	}
	if check.IfNil(arg.EpochHandler) {
		return nil, ErrNilEpochHandler
	}
	if check.IfNil(arg.ValidatorsProvider) {
		return nil, ErrNilValidatorsProvider
	}
	if arg.EpochStartEventNotifier == nil {
		return nil, ErrNilEpochStartNotifier
	}
	if arg.CacheRefreshIntervalInSec == 0 {
		log.Warn("all values not found in the cache will trigger a cache refresh from trie",
			"cacheRefreshInterval", arg.CacheRefreshIntervalInSec)
	}

	ptp := &PeerTypeProvider{
		nodesCoordinator:          arg.NodesCoordinator,
		epochHandler:              arg.EpochHandler,
		validatorsProvider:        arg.ValidatorsProvider,
		cache:                     make(map[string]*peerListAndShard),
		mutCache:                  sync.RWMutex{},
		cacheRefreshIntervalInSec: float64(arg.CacheRefreshIntervalInSec),
	}

	ptp.populateCache(ptp.epochHandler.MetaEpoch())

	arg.EpochStartEventNotifier.RegisterHandler(ptp.epochStartEventHandler())

	return ptp, nil
}

func (ptp *PeerTypeProvider) populateCache(epoch uint32) {
	ptp.mutCache.Lock()
	if ptp.isUpdating {
		return
	}
	ptp.isUpdating = true
	ptp.mutCache.Unlock()

	allNodes, err := ptp.validatorsProvider.GetLatestValidatorInfos()
	if err != nil {
		log.Warn("peerTypeProvider - GetLatestValidatorInfos failed", "error", err)
	}

	ptp.mutCache.Lock()
	defer func() {
		ptp.isUpdating = false
		ptp.mutCache.Unlock()
	}()

	ptp.cache = make(map[string]*peerListAndShard)

	for shardId, validatorsPerShard := range allNodes {
		for _, v := range validatorsPerShard {
			ptp.cache[string(v.PublicKey)] = &peerListAndShard{
				pType:  core.PeerType(v.List),
				pShard: shardId,
			}
		}
	}

	nodesMapEligible, err := ptp.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Warn("peerTypeProvider - GetAllEligibleValidatorsPublicKeys failed", "epoch", epoch)
	}
	ptp.aggregatePType(nodesMapEligible, core.EligibleList)

	nodesMapWaiting, err := ptp.nodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	if err != nil {
		log.Warn("peerTypeProvider - GetAllWaitingValidatorsPublicKeys failed", "epoch", epoch)
	}
	ptp.aggregatePType(nodesMapWaiting, core.WaitingList)

	ptp.lastCacheUpdate = time.Now()
}

func (ptp *PeerTypeProvider) aggregatePType(nodesMapEligible map[uint32][][]byte, currentPeerType core.PeerType) {
	for shardID, eligibleValidatorsInShard := range nodesMapEligible {
		for _, val := range eligibleValidatorsInShard {
			fountInTrieValidator := ptp.cache[string(val)]
			peerType := currentPeerType
			if fountInTrieValidator != nil && fountInTrieValidator.pType != currentPeerType {
				peerType = core.PeerType(fmt.Sprintf("%s (%s)", currentPeerType, fountInTrieValidator.pType))
			}

			ptp.cache[string(val)] = &peerListAndShard{
				pType:  peerType,
				pShard: shardID,
			}
		}
	}
}

// ComputeForPubKey returns the peer type for a given public key and shard id
func (ptp *PeerTypeProvider) ComputeForPubKey(pubKey []byte) (core.PeerType, uint32, error) {
	ptp.mutCache.RLock()
	peerData, ok := ptp.cache[string(pubKey)]
	lastCacheUpDate := ptp.lastCacheUpdate
	ptp.mutCache.RUnlock()
	if ok {
		return peerData.pType, peerData.pShard, nil
	}

	elapsedTime := time.Now().Sub(lastCacheUpDate)
	if elapsedTime.Seconds() > ptp.cacheRefreshIntervalInSec {
		go ptp.populateCache(ptp.epochHandler.MetaEpoch())
	}

	return core.ObserverList, 0, nil
}

func (ptp *PeerTypeProvider) epochStartEventHandler() EpochStartActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {
		ptp.populateCache(ptp.epochHandler.MetaEpoch())
	}, func(_ data.HeaderHandler) {}, core.NodesCoordinatorOrder)

	return subscribeHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (ptp *PeerTypeProvider) IsInterfaceNil() bool {
	return ptp == nil
}
