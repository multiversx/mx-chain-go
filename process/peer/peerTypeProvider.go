package peer

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart/notifier"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type peerListAndShard struct {
	pType    core.PeerType
	pSubType core.P2PPeerSubType
	pShard   uint32
}

// PeerTypeProvider handles the computation of a peer type
type PeerTypeProvider struct {
	nodesCoordinator process.NodesCoordinator
	cache            map[string]*peerListAndShard
	mutCache         sync.RWMutex
}

// ArgPeerTypeProvider contains all parameters needed for creating a PeerTypeProvider
type ArgPeerTypeProvider struct {
	NodesCoordinator        process.NodesCoordinator
	StartEpoch              uint32
	EpochStartEventNotifier process.EpochStartEventNotifier
}

// NewPeerTypeProvider will return a new instance of PeerTypeProvider
func NewPeerTypeProvider(arg ArgPeerTypeProvider) (*PeerTypeProvider, error) {
	if check.IfNil(arg.NodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
	}
	if check.IfNil(arg.EpochStartEventNotifier) {
		return nil, process.ErrNilEpochStartNotifier
	}

	ptp := &PeerTypeProvider{
		nodesCoordinator: arg.NodesCoordinator,
		cache:            make(map[string]*peerListAndShard),
		mutCache:         sync.RWMutex{},
	}

	ptp.updateCache(arg.StartEpoch)

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
			PublicKey:   pkString,
			PeerType:    string(peerListAndShard.pType),
			PeerSubType: peerListAndShard.pSubType,
			ShardId:     peerListAndShard.pShard,
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
			ptp.updateCache(hdr.GetEpoch())
		},
		func(_ data.HeaderHandler) {},
		core.IndexerOrder,
	)

	return subscribeHandler
}

func (ptp *PeerTypeProvider) updateCache(epoch uint32) {
	newCache := ptp.createNewCache(epoch)

	ptp.mutCache.Lock()
	ptp.cache = newCache
	ptp.mutCache.Unlock()
}

func (ptp *PeerTypeProvider) createNewCache(
	epoch uint32,
) map[string]*peerListAndShard {
	newCache := make(map[string]*peerListAndShard)

	nodesMapEligible, err := ptp.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Debug("peerTypeProvider - GetAllEligibleValidatorsPublicKeys failed", "epoch", epoch)
	}
	computePeerTypeAndShardId(newCache, nodesMapEligible, core.EligibleList)

	nodesMapWaiting, err := ptp.nodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	if err != nil {
		log.Debug("peerTypeProvider - GetAllWaitingValidatorsPublicKeys failed", "epoch", epoch)
	}
	computePeerTypeAndShardId(newCache, nodesMapWaiting, core.WaitingList)

	return newCache
}

func computePeerTypeAndShardId(
	newCache map[string]*peerListAndShard,
	validatorsMap map[uint32][][]byte,
	currentPeerType core.PeerType,
) {
	for shardID, shardValidators := range validatorsMap {
		for _, val := range shardValidators {
			newCache[string(val)] = &peerListAndShard{
				pType:  currentPeerType,
				pShard: shardID,
			}
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ptp *PeerTypeProvider) IsInterfaceNil() bool {
	return ptp == nil
}
