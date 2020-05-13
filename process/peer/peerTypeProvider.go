package peer

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
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
	nodesCoordinator             process.NodesCoordinator
	epochHandler                 process.EpochHandler
	validatorsProvider           process.ValidatorsProvider
	cache                        map[string]*peerListAndShard
	isUpdating                   bool
	lastCacheUpdate              time.Time
	cacheRefreshIntervalDuration time.Duration
	mutCache                     sync.RWMutex
}

type ArgPeerTypeProvider struct {
	NodesCoordinator             process.NodesCoordinator
	EpochHandler                 process.EpochHandler
	ValidatorsProvider           process.ValidatorsProvider
	EpochStartEventNotifier      process.EpochStartEventNotifier
	CacheRefreshIntervalDuration time.Duration
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
	if arg.CacheRefreshIntervalDuration <= 0 {
		return nil, process.ErrInvalidCacheRefreshIntervalDuration
	}

	ptp := &PeerTypeProvider{
		nodesCoordinator:             arg.NodesCoordinator,
		epochHandler:                 arg.EpochHandler,
		validatorsProvider:           arg.ValidatorsProvider,
		cache:                        make(map[string]*peerListAndShard),
		mutCache:                     sync.RWMutex{},
		cacheRefreshIntervalDuration: arg.CacheRefreshIntervalDuration,
	}

	ptp.populateCacheFromValidatorInfos(ptp.epochHandler.MetaEpoch())

	arg.EpochStartEventNotifier.RegisterHandler(ptp.epochStartEventHandler())

	return ptp, nil
}

func (ptp *PeerTypeProvider) populateCache(epoch uint32) {
	ptp.mutCache.RLock()
	lastCacheUpDate := ptp.lastCacheUpdate
	elapsedTime := time.Now().Sub(lastCacheUpDate)
	shouldUpdate := elapsedTime > ptp.cacheRefreshIntervalDuration
	ptp.mutCache.RUnlock()

	if shouldUpdate {
		go ptp.populateCacheFromValidatorInfos(epoch)
	}
}

func (ptp *PeerTypeProvider) populateCacheFromValidatorInfos(epoch uint32) {
	ptp.mutCache.Lock()
	if ptp.isUpdating {
		ptp.mutCache.Unlock()
		return
	}

	ptp.isUpdating = true
	ptp.mutCache.Unlock()

	defer func() {
		ptp.mutCache.Lock()
		ptp.isUpdating = false
		ptp.mutCache.Unlock()
	}()

	allNodes, err := ptp.validatorsProvider.GetLatestValidatorInfos()
	if err != nil {
		log.Warn("peerTypeProvider - GetLatestValidatorInfos failed", "error", err)
		return
	}

	newCache := ptp.createNewCache(epoch, allNodes, err)

	ptp.mutCache.Lock()
	ptp.lastCacheUpdate = time.Now()
	ptp.cache = newCache
	ptp.mutCache.Unlock()
}

func (ptp *PeerTypeProvider) createNewCache(epoch uint32, allNodes map[uint32][]*state.ValidatorInfo, err error) map[string]*peerListAndShard {
	newCache := make(map[string]*peerListAndShard)
	for shardId, validatorsPerShard := range allNodes {
		for _, v := range validatorsPerShard {
			newCache[string(v.PublicKey)] = &peerListAndShard{
				pType:  core.PeerType(v.List),
				pShard: shardId,
			}
		}
	}

	nodesMapEligible, err := ptp.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Warn("peerTypeProvider - GetAllEligibleValidatorsPublicKeys failed", "epoch", epoch)
	}
	ptp.aggregatePType(newCache, nodesMapEligible, core.EligibleList)

	nodesMapWaiting, err := ptp.nodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	if err != nil {
		log.Warn("peerTypeProvider - GetAllWaitingValidatorsPublicKeys failed", "epoch", epoch)
	}
	ptp.aggregatePType(newCache, nodesMapWaiting, core.WaitingList)
	return newCache
}

func (ptp *PeerTypeProvider) aggregatePType(newCache map[string]*peerListAndShard, nodesMapEligible map[uint32][][]byte, currentPeerType core.PeerType) {
	for shardID, eligibleValidatorsInShard := range nodesMapEligible {
		for _, val := range eligibleValidatorsInShard {
			fountInTrieValidator := newCache[string(val)]
			peerType := currentPeerType
			if fountInTrieValidator != nil && fountInTrieValidator.pType != currentPeerType {
				peerType = core.PeerType(fmt.Sprintf(core.CombinedPeerType, currentPeerType, fountInTrieValidator.pType))
			}

			newCache[string(val)] = &peerListAndShard{
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
