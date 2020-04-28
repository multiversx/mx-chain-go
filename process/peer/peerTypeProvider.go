package peer

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type peerListAndShard struct {
	pType  core.PeerType
	pShard uint32
}

// PeerTypeProvider handles the computation of a peer type
type PeerTypeProvider struct {
	nodesCoordinator          process.NodesCoordinator
	epochHandler              process.EpochHandler
	validatorsProvider        process.ValidatorsProvider
	cache                     map[string]*peerListAndShard
	lastCacheUpdate           time.Time
	cacheRefreshIntervalInSec float64
	mutCache                  sync.RWMutex
}

type ArgPeerTypeProvider struct {
	NodesCoordinator          process.NodesCoordinator
	EpochHandler              process.EpochHandler
	ValidatorsProvider        process.ValidatorsProvider
	EpochStartEventNotifier   process.EpochStartEventNotifier
	CacheRefreshIntervalInSec uint32
}

// NewPeerTypeProvider will return a new instance of PeerTypeProvider
func NewPeerTypeProvider(arg ArgPeerTypeProvider) (*PeerTypeProvider, error) {
	if check.IfNil(arg.NodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
	}
	if check.IfNil(arg.EpochHandler) {
		return nil, process.ErrNilEpochHandler
	}
	if check.IfNil(arg.ValidatorsProvider) {
		return nil, process.ErrNilValidatorsProvider
	}
	if arg.EpochStartEventNotifier == nil {
		return nil, process.ErrNilEpochStartNotifier
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

	ptp.populateCacheFromValidatorInfos(ptp.epochHandler.MetaEpoch())

	arg.EpochStartEventNotifier.RegisterHandler(ptp.epochStartEventHandler())

	return ptp, nil
}

func (ptp *PeerTypeProvider) populateCache(epoch uint32) {
	if !ptp.shouldUpdateCache() {
		return
	}

	go ptp.populateCacheFromValidatorInfos(epoch)
}

func (ptp *PeerTypeProvider) populateCacheFromValidatorInfos(epoch uint32) {
	allNodes, err := ptp.validatorsProvider.GetLatestValidatorInfos()
	if err != nil {
		log.Warn("peerTypeProvider - GetLatestValidatorInfos failed", "error", err)
		return
	}

	ptp.mutCache.Lock()
	defer func() {
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
}

func (ptp *PeerTypeProvider) shouldUpdateCache() bool {
	ptp.mutCache.Lock()
	defer ptp.mutCache.Unlock()
	lastCacheUpDate := ptp.lastCacheUpdate
	elapsedTime := time.Now().Sub(lastCacheUpDate)
	if elapsedTime.Seconds() < ptp.cacheRefreshIntervalInSec {
		return false
	}
	ptp.lastCacheUpdate = time.Now()
	return true
}

func (ptp *PeerTypeProvider) aggregatePType(nodesMapEligible map[uint32][][]byte, currentPeerType core.PeerType) {
	for shardID, eligibleValidatorsInShard := range nodesMapEligible {
		for _, val := range eligibleValidatorsInShard {
			fountInTrieValidator := ptp.cache[string(val)]
			peerType := currentPeerType
			if fountInTrieValidator != nil && fountInTrieValidator.pType != currentPeerType {
				peerType = core.PeerType(fmt.Sprintf(core.CombinedPeerType, currentPeerType, fountInTrieValidator.pType))
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
	ptp.mutCache.RUnlock()
	if ok {
		return peerData.pType, peerData.pShard, nil
	}

	ptp.populateCache(ptp.epochHandler.MetaEpoch())

	return core.ObserverList, 0, nil
}

func (ptp *PeerTypeProvider) epochStartEventHandler() sharding.EpochStartActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(func(hdr data.HeaderHandler) {
		ptp.populateCache(ptp.epochHandler.MetaEpoch())
	}, func(_ data.HeaderHandler) {}, core.ConsensusOrder)

	return subscribeHandler
}

// IsInterfaceNil returns true if there is no value under the interface
func (ptp *PeerTypeProvider) IsInterfaceNil() bool {
	return ptp == nil
}
