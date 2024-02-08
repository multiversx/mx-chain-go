package peer

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart/notifier"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/state"
)

type peerListAndShard struct {
	pType    common.PeerType
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
func (ptp *PeerTypeProvider) ComputeForPubKey(pubKey []byte) (common.PeerType, uint32, error) {
	ptp.mutCache.RLock()
	peerData, ok := ptp.cache[string(pubKey)]
	ptp.mutCache.RUnlock()

	if ok {
		return peerData.pType, peerData.pShard, nil
	}

	return common.ObserverList, 0, nil
}

// GetAllPeerTypeInfos returns all known peer type infos
func (ptp *PeerTypeProvider) GetAllPeerTypeInfos() []*state.PeerTypeInfo {
	ptp.mutCache.RLock()
	defer ptp.mutCache.RUnlock()

	peerTypeInfos := make([]*state.PeerTypeInfo, 0, len(ptp.cache))
	for pkString, peerListAndShardVal := range ptp.cache {
		peerTypeInfos = append(peerTypeInfos, &state.PeerTypeInfo{
			PublicKey:   pkString,
			PeerType:    string(peerListAndShardVal.pType),
			PeerSubType: peerListAndShardVal.pSubType,
			ShardId:     peerListAndShardVal.pShard,
		})
	}

	return peerTypeInfos
}

func (ptp *PeerTypeProvider) epochStartEventHandler() nodesCoordinator.EpochStartActionHandler {
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
		common.IndexerOrder,
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
	computePeerTypeAndShardId(newCache, nodesMapEligible, common.EligibleList)

	nodesMapWaiting, err := ptp.nodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	if err != nil {
		log.Debug("peerTypeProvider - GetAllWaitingValidatorsPublicKeys failed", "epoch", epoch)
	}
	computePeerTypeAndShardId(newCache, nodesMapWaiting, common.WaitingList)

	auctionList, err := ptp.nodesCoordinator.GetAllAuctionPublicKeys(epoch)
	if err != nil {
		log.Debug("peerTypeProvider - GetAllAuctionPublicKeys failed", "epoch", epoch)
	}
	addAuctionList(newCache, auctionList)

	return newCache
}

func computePeerTypeAndShardId(
	newCache map[string]*peerListAndShard,
	validatorsMap map[uint32][][]byte,
	currentPeerType common.PeerType,
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

func addAuctionList(newCache map[string]*peerListAndShard, auctionList [][]byte) {
	for _, blsKey := range auctionList {
		newCache[string(blsKey)] = &peerListAndShard{
			pType:  common.AuctionList,
			pShard: core.AllShardId,
		}
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ptp *PeerTypeProvider) IsInterfaceNil() bool {
	return ptp == nil
}
