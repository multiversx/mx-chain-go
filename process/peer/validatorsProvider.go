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
	lastCacheUpdate              time.Time
	lock                         sync.RWMutex
	cancelFunc                   func()
	pubkeyConverter              core.PubkeyConverter
	maxRating                    uint32
	currentEpoch                 uint32
}

// ArgValidatorsProvider contains all parameters needed for creating a validatorsProvider
type ArgValidatorsProvider struct {
	NodesCoordinator                  process.NodesCoordinator
	EpochStartEventNotifier           process.EpochStartEventNotifier
	CacheRefreshIntervalDurationInSec time.Duration
	ValidatorStatistics               process.ValidatorStatisticsProcessor
	PubKeyConverter                   core.PubkeyConverter
	StartEpoch                        uint32
	MaxRating                         uint32
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

	valProvider := &validatorsProvider{
		nodesCoordinator:             args.NodesCoordinator,
		validatorStatistics:          args.ValidatorStatistics,
		cache:                        make(map[string]*state.ValidatorApiResponse),
		cacheRefreshIntervalDuration: args.CacheRefreshIntervalDurationInSec,
		refreshCache:                 make(chan uint32),
		lock:                         sync.RWMutex{},
		cancelFunc:                   cancelfunc,
		maxRating:                    args.MaxRating,
		pubkeyConverter:              args.PubKeyConverter,
		currentEpoch:                 args.StartEpoch,
	}

	go valProvider.startRefreshProcess(currentContext)
	args.EpochStartEventNotifier.RegisterHandler(valProvider.epochStartEventHandler())

	return valProvider, nil
}

// GetLatestValidators gets the latest configuration of validators from the peerAccountsTrie
func (vp *validatorsProvider) GetLatestValidators() map[string]*state.ValidatorApiResponse {
	vp.lock.RLock()
	shouldUpdate := time.Since(vp.lastCacheUpdate) > vp.cacheRefreshIntervalDuration
	vp.lock.RUnlock()

	if shouldUpdate {
		vp.updateCache()
	}

	vp.lock.RLock()
	clonedMap := cloneMap(vp.cache)
	vp.lock.RUnlock()

	return clonedMap
}

func cloneMap(cache map[string]*state.ValidatorApiResponse) map[string]*state.ValidatorApiResponse {
	newMap := make(map[string]*state.ValidatorApiResponse)

	for k, v := range cache {
		newMap[k] = cloneValidatorAPIResponse(v)
	}

	return newMap
}

func cloneValidatorAPIResponse(v *state.ValidatorApiResponse) *state.ValidatorApiResponse {
	if v == nil {
		return nil
	}
	return &state.ValidatorApiResponse{
		TempRating:                         v.TempRating,
		NumLeaderSuccess:                   v.NumLeaderSuccess,
		NumLeaderFailure:                   v.NumLeaderFailure,
		NumValidatorSuccess:                v.NumValidatorSuccess,
		NumValidatorFailure:                v.NumValidatorFailure,
		NumValidatorIgnoredSignatures:      v.NumValidatorIgnoredSignatures,
		Rating:                             v.Rating,
		RatingModifier:                     v.RatingModifier,
		TotalNumLeaderSuccess:              v.TotalNumLeaderSuccess,
		TotalNumLeaderFailure:              v.TotalNumLeaderFailure,
		TotalNumValidatorSuccess:           v.TotalNumValidatorSuccess,
		TotalNumValidatorFailure:           v.TotalNumValidatorFailure,
		TotalNumValidatorIgnoredSignatures: v.TotalNumValidatorIgnoredSignatures,
		ShardId:                            v.ShardId,
		ValidatorStatus:                    v.ValidatorStatus,
	}
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

func (vp *validatorsProvider) startRefreshProcess(ctx context.Context) {
	for {
		vp.updateCache()
		select {
		case epoch := <-vp.refreshCache:
			vp.lock.Lock()
			vp.currentEpoch = epoch
			vp.lock.Unlock()
			log.Trace("startRefreshProcess - forced refresh", "epoch", vp.currentEpoch)
		case <-ctx.Done():
			log.Debug("validatorsProvider's go routine is stopping...")
			return
		}
	}
}

func (vp *validatorsProvider) updateCache() {
	lastFinalizedRootHash := vp.validatorStatistics.LastFinalizedRootHash()
	if len(lastFinalizedRootHash) == 0 {
		return
	}
	allNodes, err := vp.validatorStatistics.GetValidatorInfoForRootHash(lastFinalizedRootHash)
	if err != nil {
		log.Trace("validatorsProvider - GetLatestValidatorInfos failed", "error", err)
	}

	vp.lock.RLock()
	epoch := vp.currentEpoch
	vp.lock.RUnlock()

	newCache := vp.createNewCache(epoch, allNodes)

	vp.lock.Lock()
	vp.lastCacheUpdate = time.Now()
	vp.cache = newCache
	vp.lock.Unlock()
}

func (vp *validatorsProvider) createNewCache(
	epoch uint32,
	allNodes map[uint32][]*state.ValidatorInfo,
) map[string]*state.ValidatorApiResponse {
	newCache := vp.createValidatorApiResponseMapFromValidatorInfoMap(allNodes)

	nodesMapEligible, err := vp.nodesCoordinator.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Debug("validatorsProvider - GetAllEligibleValidatorsPublicKeys failed", "epoch", epoch)
	}
	vp.aggregateLists(newCache, nodesMapEligible, core.EligibleList)

	nodesMapWaiting, err := vp.nodesCoordinator.GetAllWaitingValidatorsPublicKeys(epoch)
	if err != nil {
		log.Debug("validatorsProvider - GetAllWaitingValidatorsPublicKeys failed", "epoch", epoch)
	}
	vp.aggregateLists(newCache, nodesMapWaiting, core.WaitingList)

	return newCache
}

func (vp *validatorsProvider) createValidatorApiResponseMapFromValidatorInfoMap(allNodes map[uint32][]*state.ValidatorInfo) map[string]*state.ValidatorApiResponse {
	newCache := make(map[string]*state.ValidatorApiResponse)
	for _, validatorInfosInShard := range allNodes {
		for _, validatorInfo := range validatorInfosInShard {
			strKey := vp.pubkeyConverter.Encode(validatorInfo.PublicKey)
			newCache[strKey] = &state.ValidatorApiResponse{
				NumLeaderSuccess:                   validatorInfo.LeaderSuccess,
				NumLeaderFailure:                   validatorInfo.LeaderFailure,
				NumValidatorSuccess:                validatorInfo.ValidatorSuccess,
				NumValidatorFailure:                validatorInfo.ValidatorFailure,
				NumValidatorIgnoredSignatures:      validatorInfo.ValidatorIgnoredSignatures,
				TotalNumLeaderSuccess:              validatorInfo.TotalLeaderSuccess,
				TotalNumLeaderFailure:              validatorInfo.TotalLeaderFailure,
				TotalNumValidatorSuccess:           validatorInfo.TotalValidatorSuccess,
				TotalNumValidatorFailure:           validatorInfo.TotalValidatorFailure,
				TotalNumValidatorIgnoredSignatures: validatorInfo.TotalValidatorIgnoredSignatures,
				RatingModifier:                     validatorInfo.RatingModifier,
				Rating:                             float32(validatorInfo.Rating) * 100 / float32(vp.maxRating),
				TempRating:                         float32(validatorInfo.TempRating) * 100 / float32(vp.maxRating),
				ShardId:                            validatorInfo.ShardId,
				ValidatorStatus:                    validatorInfo.List,
			}
		}
	}

	return newCache
}

func (vp *validatorsProvider) aggregateLists(
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
				newCache[encodedKey].ValidatorStatus = peerType
				log.Debug("validator from map not found in trie", "pk", encodedKey, "map", peerType)
				continue
			}

			trieList := core.PeerType(foundInTrieValidator.ValidatorStatus)
			if shouldCombine(trieList, currentList) {
				peerType = fmt.Sprintf(core.CombinedPeerType, currentList, trieList)
			}

			newCache[encodedKey].ShardId = shardID
			newCache[encodedKey].ValidatorStatus = peerType
		}
	}
}

func shouldCombine(triePeerType core.PeerType, currentPeerType core.PeerType) bool {
	// currently just "eligible (leaving)" or "waiting (leaving)" are allowed
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
