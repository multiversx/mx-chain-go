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
	validatorsProvider           process.ValidatorsProvider
	cache                        map[string]*peerListAndShard
	cacheRefreshIntervalDuration time.Duration
	refreshCache                 chan uint32
	mutCache                     sync.RWMutex
	cancelFunc                   func()
}

// ArgPeerTypeProvider contains all parameters needed for creating a PeerTypeProvider
type ArgPeerTypeProvider struct {
	NodesCoordinator             process.NodesCoordinator
	StartEpoch                   uint32
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
		validatorsProvider:           arg.ValidatorsProvider,
		cache:                        make(map[string]*peerListAndShard),
		mutCache:                     sync.RWMutex{},
		cacheRefreshIntervalDuration: arg.PeerTypeRefreshIntervalInSec,
		refreshCache:                 make(chan uint32),
	}

	currentContext, cancelfunc := context.WithCancel(context.Background())
	ptp.cancelFunc = cancelfunc
	go ptp.startRefreshProcess(currentContext, arg.StartEpoch)
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
			log.Trace("epochStartEventHandler - refreshCache forced",
				"nonce", hdr.GetNonce(),
				"shard", hdr.GetShardID(),
				"round", hdr.GetRound(),
				"epoch", hdr.GetEpoch())
			go func() {
				ptp.refreshCache <- hdr.GetEpoch()
			}()
		},
		func(_ data.HeaderHandler) {},
		core.IndexerOrder,
	)

	return subscribeHandler
}

func (ptp *PeerTypeProvider) startRefreshProcess(ctx context.Context, startEpoch uint32) {
	epoch := startEpoch
	for {
		ptp.updateCache(epoch)
		select {
		case epoch = <-ptp.refreshCache:
			log.Trace("startRefreshProcess - forced refresh", "epoch", epoch)
		case <-ctx.Done():
			log.Debug("peerTypeProvider's go routine is stopping...")
			return
		case <-time.After(ptp.cacheRefreshIntervalDuration):
			log.Trace("startRefreshProcess - time after")
		}
	}
}

func (ptp *PeerTypeProvider) updateCache(epoch uint32) {
	allNodes, err := ptp.validatorsProvider.GetLatestValidatorInfos()
	if err != nil {
		log.Trace("peerTypeProvider - GetLatestValidatorInfos failed", "error", err)
	}

	newCache := ptp.createNewCache(epoch, allNodes)

	ptp.mutCache.Lock()
	ptp.cache = newCache
	ptp.mutCache.Unlock()
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

	nodesMapEligible, err := ptp.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Debug("peerTypeProvider - GetAllEligibleValidatorsPublicKeys failed", "epoch", epoch)
	}
	aggregatePType(newCache, nodesMapEligible, core.EligibleList)

	nodesMapWaiting, err := ptp.nodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	if err != nil {
		log.Debug("peerTypeProvider - GetAllWaitingValidatorsPublicKeys failed", "epoch", epoch)
	}
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
			if foundInTrieValidator != nil && shouldCombine(foundInTrieValidator.pType, currentPeerType) {
				peerType = core.PeerType(fmt.Sprintf(core.CombinedPeerType, currentPeerType, foundInTrieValidator.pType))
			}

			newCache[string(val)] = &peerListAndShard{
				pType:  peerType,
				pShard: shardID,
			}
		}
	}
}

func shouldCombine(triePeerType core.PeerType, currentPeerType core.PeerType) bool {
	notTheSame := triePeerType != currentPeerType
	notEligibleOrWaiting := triePeerType != core.EligibleList &&
		triePeerType != core.WaitingList

	return notTheSame && notEligibleOrWaiting
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
