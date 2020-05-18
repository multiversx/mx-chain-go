package peer

import (
	"context"
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
	cacheRefreshIntervalDuration time.Duration
	refreshCache                 chan bool
	mutCache                     sync.RWMutex
	cancelFunc                   func()
}

// ArgPeerTypeProvider contains all parameters needed for creating a PeerTypeProvider
type ArgPeerTypeProvider struct {
	NodesCoordinator             process.NodesCoordinator
	EpochHandler                 process.EpochHandler
	ValidatorsProvider           process.ValidatorsProvider
	EpochStartEventNotifier      process.EpochStartEventNotifier
	PeerTypeRefreshIntervalInSec time.Duration
	Context                      context.Context
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
	if check.IfNil(arg.EpochStartEventNotifier) {
		return nil, process.ErrNilEpochStartNotifier
	}
	if arg.PeerTypeRefreshIntervalInSec <= 0 {
		return nil, process.ErrInvalidPeerTypeRefreshIntervalInSec
	}
	if arg.Context == nil {
		return nil, process.ErrNilContext
	}

	ptp := &PeerTypeProvider{
		nodesCoordinator:             arg.NodesCoordinator,
		epochHandler:                 arg.EpochHandler,
		validatorsProvider:           arg.ValidatorsProvider,
		cache:                        make(map[string]*peerListAndShard),
		mutCache:                     sync.RWMutex{},
		cacheRefreshIntervalDuration: arg.PeerTypeRefreshIntervalInSec,
		refreshCache:                 make(chan bool),
	}

	currentContext, cancelfunc := context.WithCancel(context.Background())
	ptp.cancelFunc = cancelfunc
	go ptp.startRefreshProcess(currentContext)
	arg.EpochStartEventNotifier.RegisterHandler(ptp.epochStartEventHandler())

	return ptp, nil
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

// GetAllPeerTypeInfos returns all known peer type infos
func (ptp *PeerTypeProvider) GetAllPeerTypeInfos() []*state.PeerTypeInfo {
	ptp.mutCache.RLock()
	defer ptp.mutCache.RUnlock()

	peerTypeInfos := make([]*state.PeerTypeInfo, 0, len(ptp.cache))
	for pkString, peerListAndShard := range ptp.cache {
		peerTypeInfos = append(peerTypeInfos, &state.PeerTypeInfo{
			PublicKey: pkString,
			PeerType:  string(peerListAndShard.pType),
			ShardId:   peerListAndShard.pShard,
		})
	}

	return peerTypeInfos
}

func (ptp *PeerTypeProvider) epochStartEventHandler() sharding.EpochStartActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(
		func(hdr data.HeaderHandler) {
			log.Debug("epochStartEventHandler - refreshCache forced",
				"nonce", hdr.GetNonce(),
				"shard", hdr.GetShardID(),
				"round", hdr.GetRound())
			ptp.refreshCache <- true
		},
		func(_ data.HeaderHandler) {},
		core.IndexerOrder,
	)

	return subscribeHandler
}

func (ptp *PeerTypeProvider) startRefreshProcess(ctx context.Context) {
	for {
		ptp.updateCache()
		select {
		case <-ptp.refreshCache:
			log.Debug("startRefreshProcess - forced refresh")
		case <-ctx.Done():
			log.Debug("peerTypeProvider's go routine is stopping...")
			return
		case <-time.After(ptp.cacheRefreshIntervalDuration):
			log.Debug("startRefreshProcess - time after")
		}
	}
}

func (ptp *PeerTypeProvider) updateCache() {
	log.Debug("peerTypeProvider - updateCache started")
	allNodes, err := ptp.validatorsProvider.GetLatestValidatorInfos()
	if err != nil {
		log.Debug("peerTypeProvider - GetLatestValidatorInfos failed", "error", err)
	}

	log.Debug("peerTypeProvider - create new cache - before")
	newCache := ptp.createNewCache(ptp.epochHandler.MetaEpoch(), allNodes)
	log.Debug("peerTypeProvider - create new cache - after")

	ptp.mutCache.Lock()
	ptp.cache = newCache
	ptp.mutCache.Unlock()

	log.Debug("peerTypeProvider - updateCache finished")
}

func (ptp *PeerTypeProvider) createNewCache(
	epoch uint32,
	allNodes map[uint32][]*state.ValidatorInfo,
) map[string]*peerListAndShard {
	newCache := make(map[string]*peerListAndShard)
	for shardId, validatorsPerShard := range allNodes {
		for _, v := range validatorsPerShard {
			newCache[string(v.PublicKey)] = &peerListAndShard{
				pType:  core.PeerType(v.List),
				pShard: shardId,
			}
		}
	}

	log.Debug("createNewCache - before GetAllEligibleValidatorsPublicKeys")
	nodesMapEligible, err := ptp.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Warn("peerTypeProvider - GetAllEligibleValidatorsPublicKeys failed", "epoch", epoch)
	}
	log.Debug("createNewCache - after GetAllEligibleValidatorsPublicKeys")
	aggregatePType(newCache, nodesMapEligible, core.EligibleList)

	log.Debug("createNewCache - before GetAllWaitingValidatorsPublicKeys")
	nodesMapWaiting, err := ptp.nodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	if err != nil {
		log.Warn("peerTypeProvider - GetAllWaitingValidatorsPublicKeys failed", "epoch", epoch)
	}
	log.Debug("createNewCache - after GetAllWaitingValidatorsPublicKeys")
	aggregatePType(newCache, nodesMapWaiting, core.WaitingList)

	return newCache
}

func aggregatePType(
	newCache map[string]*peerListAndShard,
	validatorsMap map[uint32][][]byte,
	currentPeerType core.PeerType,
) {
	for shardID, shardValidators := range validatorsMap {
		for _, val := range shardValidators {
			foundInTrieValidator := newCache[string(val)]
			peerType := currentPeerType
			if foundInTrieValidator != nil && foundInTrieValidator.pType != currentPeerType {
				peerType = core.PeerType(fmt.Sprintf(core.CombinedPeerType, currentPeerType, foundInTrieValidator.pType))
			}

			newCache[string(val)] = &peerListAndShard{
				pType:  peerType,
				pShard: shardID,
			}
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ptp *PeerTypeProvider) IsInterfaceNil() bool {
	return ptp == nil
}

// Close - frees up everything, cancels long running methods
func (ptp *PeerTypeProvider) Close() error {
	ptp.cancelFunc()

	return nil
}
