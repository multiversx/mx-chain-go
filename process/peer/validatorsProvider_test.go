package peer

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewValidatorsProvider_WithNilValidatorStatisticsShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.ValidatorStatistics = nil
	vp, err := NewValidatorsProvider(arg)
	assert.Equal(t, process.ErrNilValidatorStatistics, err)
	assert.Nil(t, vp)
}

func TestNewValidatorsProvider_WithMaxRatingZeroShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.MaxRating = uint32(0)
	vp, err := NewValidatorsProvider(arg)
	assert.Equal(t, process.ErrMaxRatingZero, err)
	assert.Nil(t, vp)
}

func TestNewValidatorsProvider_WithNilValidatorPubkeyConverterShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.PubKeyConverter = nil
	vp, err := NewValidatorsProvider(arg)

	assert.Equal(t, process.ErrNilPubkeyConverter, err)
	assert.True(t, check.IfNil(vp))
}

func TestNewValidatorsProvider_WithNilNodesCoordinatorrShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.NodesCoordinator = nil
	vp, err := NewValidatorsProvider(arg)

	assert.Equal(t, process.ErrNilNodesCoordinator, err)
	assert.True(t, check.IfNil(vp))
}

func TestNewValidatorsProvider_WithNilStartOfEpochTriggerShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.EpochStartEventNotifier = nil
	vp, err := NewValidatorsProvider(arg)

	assert.Equal(t, process.ErrNilEpochStartNotifier, err)
	assert.True(t, check.IfNil(vp))
}

func TestNewValidatorsProvider_WithNilRefresCacheIntervalInSecShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.CacheRefreshIntervalDurationInSec = 0
	vp, err := NewValidatorsProvider(arg)

	assert.Equal(t, process.ErrInvalidCacheRefreshIntervalInSec, err)
	assert.True(t, check.IfNil(vp))
}

func TestValidatorsProvider_GetLatestValidatorsSecondHashDoesNotExist(t *testing.T) {
	mut := sync.Mutex{}
	root := []byte("rootHash")
	e := errors.Errorf("not ok")
	initialInfo := createMockValidatorInfo()

	validatorInfos := map[uint32][]*state.ValidatorInfo{
		0: {initialInfo},
	}

	gotOk := false
	gotNil := false
	vs := &mock.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() (bytes []byte) {
			mut.Lock()
			defer mut.Unlock()
			return root
		},
		GetValidatorInfoForRootHashCalled: func(rootHash []byte) (m map[uint32][]*state.ValidatorInfo, err error) {
			mut.Lock()
			defer mut.Unlock()
			if bytes.Equal([]byte("rootHash"), rootHash) {
				gotOk = true
				return validatorInfos, nil
			}
			gotNil = true
			return nil, e
		},
	}

	nc := &mock.NodesCoordinatorMock{GetAllEligibleValidatorsPublicKeysCalled: func() (map[uint32][][]byte, error) {
		return map[uint32][][]byte{0: {initialInfo.PublicKey}}, nil
	}}

	maxRating := uint32(100)
	args := createDefaultValidatorsProviderArg()
	args.ValidatorStatistics = vs
	args.NodesCoordinator = nc
	args.MaxRating = maxRating
	vp, _ := NewValidatorsProvider(args)
	time.Sleep(time.Millisecond)
	vinfos := vp.GetLatestValidators()
	assert.NotNil(t, vinfos)
	assert.Equal(t, 1, len(vinfos))
	time.Sleep(time.Millisecond)
	mut.Lock()
	root = []byte("otherHash")
	mut.Unlock()
	vinfos2 := vp.GetLatestValidators()
	time.Sleep(time.Millisecond)
	assert.NotNil(t, vinfos2)
	assert.Equal(t, 1, len(vinfos2))
	mut.Lock()
	assert.True(t, gotOk)
	assert.True(t, gotNil)
	mut.Unlock()
	validatorInfoApi := vinfos[hex.EncodeToString(initialInfo.GetPublicKey())]
	assert.Equal(t, initialInfo.GetTempRating(), uint32(validatorInfoApi.GetTempRating()))
	assert.Equal(t, initialInfo.GetRating(), uint32(validatorInfoApi.GetRating()))
	assert.Equal(t, initialInfo.GetLeaderSuccess(), validatorInfoApi.GetNumLeaderSuccess())
	assert.Equal(t, initialInfo.GetLeaderFailure(), validatorInfoApi.GetNumLeaderFailure())
	assert.Equal(t, initialInfo.GetValidatorSuccess(), validatorInfoApi.GetNumValidatorSuccess())
	assert.Equal(t, initialInfo.GetValidatorFailure(), validatorInfoApi.GetNumValidatorFailure())
	assert.Equal(t, initialInfo.GetTotalLeaderSuccess(), validatorInfoApi.GetTotalNumLeaderSuccess())
	assert.Equal(t, initialInfo.GetTotalLeaderFailure(), validatorInfoApi.GetTotalNumLeaderFailure())
	assert.Equal(t, initialInfo.GetTotalValidatorSuccess(), validatorInfoApi.GetTotalNumValidatorSuccess())
	assert.Equal(t, initialInfo.GetTotalValidatorFailure(), validatorInfoApi.GetTotalNumValidatorFailure())
}

func TestValidatorsProvider_ShouldWork(t *testing.T) {
	args := createDefaultValidatorsProviderArg()
	vp, err := NewValidatorsProvider(args)

	assert.Nil(t, err)
	assert.NotNil(t, vp)
}

func TestValidatorsProvider_CallsPopulateAndRegister(t *testing.T) {
	numRegisterHandlerCalled := int32(0)
	numPopulateCacheCalled := int32(0)

	arg := createDefaultValidatorsProviderArg()
	arg.CacheRefreshIntervalDurationInSec = 10 * time.Millisecond
	arg.EpochStartEventNotifier = &mock.EpochStartNotifierStub{
		RegisterHandlerCalled: func(handler epochStart.ActionHandler) {
			atomic.AddInt32(&numRegisterHandlerCalled, 1)
		},
	}

	arg.ValidatorStatistics = &mock.ValidatorStatisticsProcessorStub{
		GetValidatorInfoForRootHashCalled: func(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error) {
			atomic.AddInt32(&numPopulateCacheCalled, 1)
			return nil, nil
		},
		LastFinalizedRootHashCalled: func() []byte {
			return []byte("rootHash")
		},
	}

	_, _ = NewValidatorsProvider(arg)

	time.Sleep(time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&numPopulateCacheCalled))
	assert.Equal(t, int32(1), atomic.LoadInt32(&numRegisterHandlerCalled))
}

func TestValidatorsProvider_UpdateCache_WithError(t *testing.T) {
	expectedErr := errors.New("expectedError")
	arg := createDefaultValidatorsProviderArg()

	validatorProc := &mock.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() []byte {
			return []byte("rootHash")
		},
	}
	validatorProc.GetValidatorInfoForRootHashCalled = func(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error) {
		return nil, expectedErr
	}

	pk := []byte("pk")
	nodesCoordinator := mock.NewNodesCoordinatorMock()
	nodesCoordinator.GetAllEligibleValidatorsPublicKeysCalled = func() (map[uint32][][]byte, error) {
		return map[uint32][][]byte{
			0: {pk},
		}, nil
	}

	vsp := validatorsProvider{
		nodesCoordinator:             nodesCoordinator,
		validatorStatistics:          validatorProc,
		cache:                        nil,
		cacheRefreshIntervalDuration: arg.CacheRefreshIntervalDurationInSec,
		refreshCache:                 nil,
		lock:                         sync.RWMutex{},
		pubkeyConverter:              mock.NewPubkeyConverterMock(32),
	}

	vsp.updateCache()

	assert.NotNil(t, vsp.GetCache())
	assert.Equal(t, 1, len(vsp.GetCache()))
}

func TestValidatorsProvider_Cancel_startRefreshProcess(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()

	arg.CacheRefreshIntervalDurationInSec = 1 * time.Millisecond
	vsp := validatorsProvider{
		nodesCoordinator:             arg.NodesCoordinator,
		validatorStatistics:          arg.ValidatorStatistics,
		cache:                        make(map[string]*state.ValidatorApiResponse),
		cacheRefreshIntervalDuration: arg.CacheRefreshIntervalDurationInSec,
		refreshCache:                 make(chan uint32),
		lock:                         sync.RWMutex{},
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	mutFinished := sync.Mutex{}
	finished := false
	go func() {
		vsp.startRefreshProcess(ctx)
		mutFinished.Lock()
		finished = true
		mutFinished.Unlock()
	}()

	time.Sleep(5 * time.Millisecond)
	mutFinished.Lock()
	currentFinished := finished
	mutFinished.Unlock()
	assert.False(t, currentFinished)

	cancelFunc()

	time.Sleep(5 * time.Millisecond)
	mutFinished.Lock()
	currentFinished = finished
	mutFinished.Unlock()
	assert.True(t, currentFinished)
}

func TestValidatorsProvider_UpdateCache(t *testing.T) {
	pk := []byte("pk1")
	initialShardId := uint32(1)
	initialList := string(core.EligibleList)
	validatorsMap := make(map[uint32][]*state.ValidatorInfo)
	validatorsMap[initialShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pk,
			List:      initialList,
			ShardId:   initialShardId,
		},
	}
	arg := createDefaultValidatorsProviderArg()
	validatorProc := &mock.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() []byte {
			return []byte("rootHash")
		},
	}
	validatorProc.GetValidatorInfoForRootHashCalled = func(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error) {
		return validatorsMap, nil
	}

	vsp := validatorsProvider{
		nodesCoordinator:             arg.NodesCoordinator,
		validatorStatistics:          validatorProc,
		cache:                        nil,
		cacheRefreshIntervalDuration: arg.CacheRefreshIntervalDurationInSec,
		refreshCache:                 nil,
		pubkeyConverter:              mock.NewPubkeyConverterMock(32),
		lock:                         sync.RWMutex{},
	}

	vsp.updateCache()

	assert.NotNil(t, vsp.cache)
	assert.Equal(t, len(validatorsMap[initialShardId]), len(vsp.cache))
	encodedKey := arg.PubKeyConverter.Encode(pk)
	assert.NotNil(t, vsp.cache[encodedKey])
	assert.Equal(t, initialList, vsp.cache[encodedKey].ValidatorStatus)
	assert.Equal(t, initialShardId, vsp.cache[encodedKey].ShardId)
}

func TestValidatorsProvider_aggregatePType_equal(t *testing.T) {
	pubKeyConverter := mock.NewPubkeyConverterMock(32)
	pkInactive := []byte("pk1")
	trieInctiveShardId := uint32(0)
	inactiveList := string(core.InactiveList)
	pkEligible := []byte("pk2")
	trieEligibleShardId := uint32(1)
	eligibleList := string(core.EligibleList)
	pkLeaving := []byte("pk3")
	trieLeavingShardId := uint32(2)
	leavingList := string(core.LeavingList)

	encodedEligible := pubKeyConverter.Encode(pkEligible)
	encondedInactive := pubKeyConverter.Encode(pkInactive)
	encodedLeaving := pubKeyConverter.Encode(pkLeaving)
	cache := make(map[string]*state.ValidatorApiResponse)
	cache[encondedInactive] = &state.ValidatorApiResponse{ValidatorStatus: inactiveList, ShardId: trieInctiveShardId}
	cache[encodedEligible] = &state.ValidatorApiResponse{ValidatorStatus: eligibleList, ShardId: trieEligibleShardId}
	cache[encodedLeaving] = &state.ValidatorApiResponse{ValidatorStatus: leavingList, ShardId: trieLeavingShardId}

	nodesCoordinatorEligibleShardId := uint32(0)
	nodesCoordinatorLeavingShardId := core.MetachainShardId

	validatorsMap := map[uint32][][]byte{
		nodesCoordinatorEligibleShardId: {pkEligible},
		nodesCoordinatorLeavingShardId:  {pkLeaving},
	}

	vp := validatorsProvider{
		pubkeyConverter: pubKeyConverter,
	}

	vp.aggregateLists(cache, validatorsMap, core.EligibleList)

	assert.Equal(t, trieInctiveShardId, cache[encondedInactive].ShardId)
	assert.Equal(t, inactiveList, cache[encondedInactive].ValidatorStatus)

	assert.Equal(t, nodesCoordinatorEligibleShardId, cache[encodedEligible].ShardId)
	assert.Equal(t, eligibleList, cache[encodedEligible].ValidatorStatus)

	aggregatedList := "eligible (leaving)"
	assert.Equal(t, nodesCoordinatorLeavingShardId, cache[encodedLeaving].ShardId)
	assert.Equal(t, aggregatedList, cache[encodedLeaving].ValidatorStatus)
}

func TestValidatorsProvider_createCache(t *testing.T) {
	pkEligible := []byte("pk1")
	eligibleList := string(core.EligibleList)
	pkWaiting := []byte("pk2")
	waitingList := string(core.WaitingList)
	pkLeaving := []byte("pk3")
	leavingList := string(core.LeavingList)
	pkInactive := []byte("pk4")
	inactiveList := string(core.InactiveList)
	pkNew := []byte("pk5")
	newList := string(core.NewList)

	validatorsMap := make(map[uint32][]*state.ValidatorInfo)
	eligibleShardId := uint32(0)
	waitingShardId := uint32(1)
	leavingShardId := uint32(2)
	inactiveShardId := uint32(3)
	newShardId := core.MetachainShardId
	validatorsMap[eligibleShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pkEligible,
			ShardId:   eligibleShardId,
			List:      eligibleList,
		},
	}
	validatorsMap[waitingShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pkWaiting,
			ShardId:   waitingShardId,
			List:      waitingList,
		},
	}
	validatorsMap[leavingShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pkLeaving,
			ShardId:   leavingShardId,
			List:      leavingList,
		},
	}
	validatorsMap[inactiveShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pkInactive,
			ShardId:   inactiveShardId,
			List:      inactiveList,
		},
	}
	validatorsMap[newShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pkNew,
			ShardId:   newShardId,
			List:      newList,
		},
	}
	arg := createDefaultValidatorsProviderArg()
	pubKeyConverter := mock.NewPubkeyConverterMock(32)
	vsp := validatorsProvider{
		nodesCoordinator:             arg.NodesCoordinator,
		validatorStatistics:          arg.ValidatorStatistics,
		cache:                        nil,
		cacheRefreshIntervalDuration: arg.CacheRefreshIntervalDurationInSec,
		pubkeyConverter:              pubKeyConverter,
		lock:                         sync.RWMutex{},
	}

	cache := vsp.createNewCache(0, validatorsMap)

	assert.NotNil(t, cache)

	encodedPkEligible := pubKeyConverter.Encode(pkEligible)
	assert.NotNil(t, cache[encodedPkEligible])
	assert.Equal(t, eligibleList, cache[encodedPkEligible].ValidatorStatus)
	assert.Equal(t, eligibleShardId, cache[encodedPkEligible].ShardId)

	encodedPkWaiting := pubKeyConverter.Encode(pkWaiting)
	assert.NotNil(t, cache[encodedPkWaiting])
	assert.Equal(t, waitingList, cache[encodedPkWaiting].ValidatorStatus)
	assert.Equal(t, waitingShardId, cache[encodedPkWaiting].ShardId)

	encodedPkLeaving := pubKeyConverter.Encode(pkLeaving)
	assert.NotNil(t, cache[encodedPkLeaving])
	assert.Equal(t, leavingList, cache[encodedPkLeaving].ValidatorStatus)
	assert.Equal(t, leavingShardId, cache[encodedPkLeaving].ShardId)

	encodedPkNew := pubKeyConverter.Encode(pkNew)
	assert.NotNil(t, cache[encodedPkNew])
	assert.Equal(t, newList, cache[encodedPkNew].ValidatorStatus)
	assert.Equal(t, newShardId, cache[encodedPkNew].ShardId)
}

func TestValidatorsProvider_createCache_combined(t *testing.T) {
	pkEligibleInTrie := []byte("pk1")
	eligibleList := string(core.EligibleList)
	pkInactive := []byte("pk2")
	inactiveList := string(core.InactiveList)
	pkLeavingInTrie := []byte("pk3")
	leavingList := string(core.LeavingList)

	validatorsMap := make(map[uint32][]*state.ValidatorInfo)
	eligibleShardId := uint32(0)
	inactiveShardId := uint32(1)
	leavingShardId := uint32(2)
	validatorsMap[eligibleShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pkEligibleInTrie,
			ShardId:   eligibleShardId,
			List:      eligibleList,
		},
	}
	validatorsMap[inactiveShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pkInactive,
			ShardId:   inactiveShardId,
			List:      inactiveList,
		},
	}
	validatorsMap[leavingShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pkLeavingInTrie,
			ShardId:   leavingShardId,
			List:      leavingList,
		},
	}
	arg := createDefaultValidatorsProviderArg()
	nodesCoordinator := mock.NewNodesCoordinatorMock()
	nodesCoordinatorEligibleShardId := uint32(5)
	nodesCoordinatorLeavingShardId := uint32(6)
	nodesCoordinator.GetAllEligibleValidatorsPublicKeysCalled = func() (map[uint32][][]byte, error) {
		return map[uint32][][]byte{
			nodesCoordinatorEligibleShardId: {pkEligibleInTrie},
			nodesCoordinatorLeavingShardId:  {pkLeavingInTrie},
		}, nil
	}

	vsp := validatorsProvider{
		nodesCoordinator:             nodesCoordinator,
		validatorStatistics:          arg.ValidatorStatistics,
		pubkeyConverter:              arg.PubKeyConverter,
		cache:                        nil,
		cacheRefreshIntervalDuration: arg.CacheRefreshIntervalDurationInSec,
		lock:                         sync.RWMutex{},
	}

	cache := vsp.createNewCache(0, validatorsMap)

	encodedPkEligible := arg.PubKeyConverter.Encode(pkEligibleInTrie)
	assert.NotNil(t, cache[encodedPkEligible])
	assert.Equal(t, eligibleList, cache[encodedPkEligible].ValidatorStatus)
	assert.Equal(t, nodesCoordinatorEligibleShardId, cache[encodedPkEligible].ShardId)

	encodedPkLeavingInTrie := arg.PubKeyConverter.Encode(pkLeavingInTrie)
	computedPeerType := fmt.Sprintf(core.CombinedPeerType, core.EligibleList, core.LeavingList)
	assert.NotNil(t, cache[encodedPkLeavingInTrie])
	assert.Equal(t, computedPeerType, cache[encodedPkLeavingInTrie].ValidatorStatus)
	assert.Equal(t, nodesCoordinatorLeavingShardId, cache[encodedPkLeavingInTrie].ShardId)
}

func TestValidatorsProvider_CallsPopulateOnlyAfterTimeout(t *testing.T) {
	zeroNumner := int32(0)
	populateCacheCalled := &zeroNumner

	arg := createDefaultValidatorsProviderArg()
	arg.CacheRefreshIntervalDurationInSec = time.Millisecond * 10
	validatorStatisticsProcessor := &mock.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() []byte {
			return []byte("rootHash")
		},
	}
	validatorStatisticsProcessor.GetValidatorInfoForRootHashCalled = func(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error) {
		atomic.AddInt32(populateCacheCalled, 1)
		return nil, nil
	}

	arg.ValidatorStatistics = validatorStatisticsProcessor
	vsp, _ := NewValidatorsProvider(arg)

	// allow previous call to through
	time.Sleep(time.Millisecond)

	// inside refreshInterval of 10 milis
	atomic.StoreInt32(populateCacheCalled, 0)
	_ = vsp.GetLatestValidators()
	time.Sleep(time.Millisecond)
	assert.Equal(t, int32(0), atomic.LoadInt32(populateCacheCalled))
	_ = vsp.GetLatestValidators()
	time.Sleep(time.Millisecond)
	assert.Equal(t, int32(0), atomic.LoadInt32(populateCacheCalled))

	// outside of refreshInterval
	time.Sleep(arg.CacheRefreshIntervalDurationInSec)
	_ = vsp.GetLatestValidators()
	//allow call to go through
	time.Sleep(time.Millisecond)
	assert.True(t, atomic.LoadInt32(populateCacheCalled) > 0)
}

func TestValidatorsProvider_CallsUpdateCacheOnEpochChange(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	callNumber := 0
	epochStartNotifier := &mock.EpochStartNotifierStub{}
	arg.EpochStartEventNotifier = epochStartNotifier
	arg.CacheRefreshIntervalDurationInSec = 5 * time.Millisecond
	pkEligibleInTrie := []byte("pk1")

	validatorStatisticsProcessor := &mock.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() []byte {
			return []byte("rootHash")
		},
	}
	validatorStatisticsProcessor.GetValidatorInfoForRootHashCalled = func(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error) {
		callNumber++
		// first call comes from the constructor
		if callNumber == 1 {
			return nil, nil
		}
		return map[uint32][]*state.ValidatorInfo{
			0: {
				{
					PublicKey: pkEligibleInTrie,
					List:      string(core.EligibleList),
				},
			},
		}, nil
	}
	arg.ValidatorStatistics = validatorStatisticsProcessor

	vsp, _ := NewValidatorsProvider(arg)
	encodedEligible := arg.PubKeyConverter.Encode(pkEligibleInTrie)
	assert.Equal(t, 0, len(vsp.GetCache())) // nothing in cache
	epochStartNotifier.NotifyAll(&block.Header{Nonce: 1, ShardID: 2, Round: 3})
	time.Sleep(arg.CacheRefreshIntervalDurationInSec)
	assert.Equal(t, 1, len(vsp.GetCache()))
	assert.NotNil(t, vsp.GetCache()[encodedEligible])
}

func TestValidatorsProvider_DoesntCallUpdateUpdateCacheWithoutRequests(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	callNumber := 0
	epochStartNotifier := &mock.EpochStartNotifierStub{}
	arg.EpochStartEventNotifier = epochStartNotifier
	arg.CacheRefreshIntervalDurationInSec = 5 * time.Millisecond
	pkEligibleInTrie := []byte("pk1")

	validatorStatisticsProcessor := &mock.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() []byte {
			return []byte("rootHash")
		},
	}
	validatorStatisticsProcessor.GetValidatorInfoForRootHashCalled = func(rootHash []byte) (map[uint32][]*state.ValidatorInfo, error) {
		callNumber++
		// first call comes from the constructor
		if callNumber == 1 {
			return nil, nil
		}
		return map[uint32][]*state.ValidatorInfo{
			0: {
				{
					PublicKey: pkEligibleInTrie,
					List:      string(core.EligibleList),
				},
			},
		}, nil
	}
	arg.ValidatorStatistics = validatorStatisticsProcessor

	vsp, _ := NewValidatorsProvider(arg)
	encodedEligible := arg.PubKeyConverter.Encode(pkEligibleInTrie)
	assert.Equal(t, 0, len(vsp.GetCache())) // nothing in cache
	time.Sleep(arg.CacheRefreshIntervalDurationInSec)
	assert.Equal(t, 0, len(vsp.GetCache())) // nothing in cache
	time.Sleep(arg.CacheRefreshIntervalDurationInSec)
	assert.Equal(t, 0, len(vsp.GetCache())) // nothing in cache

	resp := vsp.GetLatestValidators()
	assert.Equal(t, 1, len(vsp.GetCache()))
	assert.Equal(t, 1, len(resp))
	assert.NotNil(t, vsp.GetCache()[encodedEligible])
}
func createMockValidatorInfo() *state.ValidatorInfo {
	initialInfo := &state.ValidatorInfo{
		PublicKey:                  []byte("a1"),
		ShardId:                    0,
		List:                       "eligible",
		Index:                      1,
		TempRating:                 100,
		Rating:                     1000,
		RewardAddress:              []byte("rewardA1"),
		LeaderSuccess:              1,
		LeaderFailure:              2,
		ValidatorSuccess:           3,
		ValidatorFailure:           4,
		TotalLeaderSuccess:         10,
		TotalLeaderFailure:         20,
		TotalValidatorSuccess:      30,
		TotalValidatorFailure:      40,
		NumSelectedInSuccessBlocks: 5,
		AccumulatedFees:            big.NewInt(100),
	}
	return initialInfo
}

func createDefaultValidatorsProviderArg() ArgValidatorsProvider {
	return ArgValidatorsProvider{
		NodesCoordinator:                  &mock.NodesCoordinatorMock{},
		StartEpoch:                        1,
		EpochStartEventNotifier:           &mock.EpochStartNotifierStub{},
		CacheRefreshIntervalDurationInSec: 1 * time.Millisecond,
		ValidatorStatistics: &mock.ValidatorStatisticsProcessorStub{
			LastFinalizedRootHashCalled: func() []byte {
				return []byte("rootHash")
			},
		},
		MaxRating:       100,
		PubKeyConverter: mock.NewPubkeyConverterMock(32),
	}
}
