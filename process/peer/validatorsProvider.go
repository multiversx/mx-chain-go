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

var _ process.ValidatorsProvider = (*validatorsProvider)(nil)

// validatorsProvider is the main interface for validators' provider
type validatorsProvider struct {
	nodesCoordinator             process.NodesCoordinator
	validatorStatistics          process.ValidatorStatisticsProcessor
	cache                        map[string]*state.ValidatorApiResponse
	cacheRefreshIntervalDuration time.Duration
	refreshCache                 chan uint32
	mutCache                     sync.RWMutex
	cancelFunc                   func()
	maxRating                    uint32
	pubkeyConverter              state.PubkeyConverter
}

// ArgValidatorsProvider contains all parameters needed for creating a validatorsProvider
type ArgValidatorsProvider struct {
	NodesCoordinator                  process.NodesCoordinator
	StartEpoch                        uint32
	EpochStartEventNotifier           process.EpochStartEventNotifier
	CacheRefreshIntervalDurationInSec time.Duration
	ValidatorStatistics               process.ValidatorStatisticsProcessor
	MaxRating                         uint32
	PubKeyConverter                   state.PubkeyConverter
}

// NewValidatorsProvider instantiates a new validatorsProvider structure responsible of keeping account of
//  the latest information about the validators
func NewValidatorsProvider(
	args ArgValidatorsProvider,
) (*validatorsProvider, error) {
	if check.IfNil(args.ValidatorStatistics) {
		return nil, process.ErrNilValidatorStatistics
	}
	if check.IfNil(args.PubKeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.NodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
	}
	if check.IfNil(args.EpochStartEventNotifier) {
		return nil, process.ErrNilEpochStartNotifier
	}
	if args.MaxRating == 0 {
		return nil, process.ErrMaxRatingZero
	}
	if args.CacheRefreshIntervalDurationInSec <= 0 {
		return nil, process.ErrInvalidCacheRefreshIntervalInSec
	}

	currentContext, cancelfunc := context.WithCancel(context.Background())

	validatorsProvider := &validatorsProvider{
		nodesCoordinator:             args.NodesCoordinator,
		validatorStatistics:          args.ValidatorStatistics,
		cache:                        make(map[string]*state.ValidatorApiResponse),
		cacheRefreshIntervalDuration: args.CacheRefreshIntervalDurationInSec,
		refreshCache:                 make(chan uint32),
		mutCache:                     sync.RWMutex{},
		cancelFunc:                   cancelfunc,
		maxRating:                    args.MaxRating,
		pubkeyConverter:              args.PubKeyConverter,
	}

	go validatorsProvider.startRefreshProcess(currentContext, args.StartEpoch)
	args.EpochStartEventNotifier.RegisterHandler(validatorsProvider.epochStartEventHandler())

	return validatorsProvider, nil
}

// GetLatestValidators gets the latest configuration of validators from the peerAccountsTrie
func (vp *validatorsProvider) GetLatestValidators() map[string]*state.ValidatorApiResponse {
	vp.mutCache.Lock()
	defer vp.mutCache.Unlock()

	return vp.cache
}

func (vp *validatorsProvider) epochStartEventHandler() sharding.EpochStartActionHandler {
	subscribeHandler := notifier.NewHandlerForEpochStart(
		func(hdr data.HeaderHandler) {
			log.Trace("epochStartEventHandler - refreshCache forced",
				"nonce", hdr.GetNonce(),
				"shard", hdr.GetShardID(),
				"round", hdr.GetRound(),
				"epoch", hdr.GetEpoch())
			go func() {
				vp.refreshCache <- hdr.GetEpoch()
			}()
		},
		func(_ data.HeaderHandler) {},
		core.IndexerOrder,
	)

	return subscribeHandler
}

func (vp *validatorsProvider) startRefreshProcess(ctx context.Context, startEpoch uint32) {
	epoch := startEpoch
	for {
		vp.updateCache(epoch)
		select {
		case epoch = <-vp.refreshCache:
			log.Trace("startRefreshProcess - forced refresh", "epoch", epoch)
		case <-ctx.Done():
			log.Debug("validatorsProvider's go routine is stopping...")
			return
		case <-time.After(vp.cacheRefreshIntervalDuration):
			log.Trace("startRefreshProcess - time after")
		}
	}
}

func (vp *validatorsProvider) updateCache(epoch uint32) {
	lastFinalizedRootHash := vp.validatorStatistics.LastFinalizedRootHash()
	allNodes, err := vp.validatorStatistics.GetValidatorInfoForRootHash(lastFinalizedRootHash)
	if err != nil {
		log.Trace("validatorsProvider - GetLatestValidatorInfos failed", "error", err)
	}

	newCache := vp.createNewCache(epoch, allNodes)

	vp.mutCache.Lock()
	vp.cache = newCache
	vp.mutCache.Unlock()
}

func (vp *validatorsProvider) createNewCache(
	epoch uint32,
	allNodes map[uint32][]*state.ValidatorInfo,
) map[string]*state.ValidatorApiResponse {
	newCache := vp.createValidatorApiResponseMapFromValidatorInfoMap(allNodes)

	nodesMapEligible, err := vp.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Debug("peerTypeProvider - GetAllEligibleValidatorsPublicKeys failed", "epoch", epoch)
	}
	vp.aggregatePType(newCache, nodesMapEligible, core.EligibleList)

	nodesMapWaiting, err := vp.nodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	if err != nil {
		log.Debug("peerTypeProvider - GetAllWaitingValidatorsPublicKeys failed", "epoch", epoch)
	}
	vp.aggregatePType(newCache, nodesMapWaiting, core.WaitingList)

	return newCache
}

func (vp *validatorsProvider) createValidatorApiResponseMapFromValidatorInfoMap(allNodes map[uint32][]*state.ValidatorInfo) map[string]*state.ValidatorApiResponse {
	newCache := make(map[string]*state.ValidatorApiResponse)
	inactiveList := string(core.InactiveList)
	for _, validatorInfosInShard := range allNodes {
		for _, validatorInfo := range validatorInfosInShard {
			// do not display inactive validators
			if validatorInfo.List == inactiveList {
				continue
			}

			strKey := vp.pubkeyConverter.Encode(validatorInfo.PublicKey)
			newCache[strKey] = &state.ValidatorApiResponse{
				NumLeaderSuccess:         validatorInfo.LeaderSuccess,
				NumLeaderFailure:         validatorInfo.LeaderFailure,
				NumValidatorSuccess:      validatorInfo.ValidatorSuccess,
				NumValidatorFailure:      validatorInfo.ValidatorFailure,
				TotalNumLeaderSuccess:    validatorInfo.TotalLeaderSuccess,
				TotalNumLeaderFailure:    validatorInfo.TotalLeaderFailure,
				TotalNumValidatorSuccess: validatorInfo.TotalValidatorSuccess,
				TotalNumValidatorFailure: validatorInfo.TotalValidatorFailure,
				RatingModifier:           validatorInfo.RatingModifier,
				Rating:                   float32(validatorInfo.Rating) * 100 / float32(vp.maxRating),
				TempRating:               float32(validatorInfo.TempRating) * 100 / float32(vp.maxRating),
				ShardId:                  validatorInfo.ShardId,
				List:                     validatorInfo.List,
			}
		}
	}

	return newCache
}

func (vp *validatorsProvider) aggregatePType(
	newCache map[string]*state.ValidatorApiResponse,
	validatorsMap map[uint32][][]byte,
	currentList core.PeerType,
) {
	for shardID, shardValidators := range validatorsMap {
		for _, val := range shardValidators {
			encodedKey := vp.pubkeyConverter.Encode(val)
			foundInTrieValidator, ok := newCache[encodedKey]
			peerType := string(currentList)

			if !ok || foundInTrieValidator == nil {
				newCache[encodedKey] = &state.ValidatorApiResponse{}
				newCache[encodedKey].ShardId = shardID
				newCache[encodedKey].List = peerType
				continue
			}

			trieList := core.PeerType(foundInTrieValidator.List)
			if shouldCombine(trieList, currentList) {
				peerType = fmt.Sprintf(core.CombinedPeerType, currentList, trieList)
			}

			newCache[encodedKey].ShardId = shardID
			newCache[encodedKey].List = peerType
		}
	}
}

func shouldCombine(triePeerType core.PeerType, currentPeerType core.PeerType) bool {
	// currently just "eligible (leaving)" or "waiting (leaving)" are allowed to combine
	isLeaving := triePeerType == core.LeavingList
	isEligibleOrWaiting := currentPeerType == core.EligibleList ||
		currentPeerType == core.WaitingList

	return isLeaving && isEligibleOrWaiting
}

// IsInterfaceNil returns true if there is no value under the interface
func (vp *validatorsProvider) IsInterfaceNil() bool {
	return vp == nil
}

// Close - frees up everything, cancels long running methods
func (vp *validatorsProvider) Close() error {
	vp.cancelFunc()

	return nil
}
