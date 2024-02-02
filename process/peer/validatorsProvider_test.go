package peer

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAtomic "github.com/multiversx/mx-chain-core-go/core/atomic"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/validator"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/stakingcommon"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestNewValidatorsProvider_WithNilValidatorPubKeyConverterShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.ValidatorPubKeyConverter = nil
	vp, err := NewValidatorsProvider(arg)

	assert.True(t, errors.Is(err, process.ErrNilPubkeyConverter))
	assert.True(t, strings.Contains(err.Error(), "validator"))
	assert.True(t, check.IfNil(vp))
}

func TestNewValidatorsProvider_WithNilAddressPubkeyConverterShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.AddressPubKeyConverter = nil
	vp, err := NewValidatorsProvider(arg)

	assert.True(t, errors.Is(err, process.ErrNilPubkeyConverter))
	assert.True(t, strings.Contains(err.Error(), "address"))
	assert.True(t, check.IfNil(vp))
}

func TestNewValidatorsProvider_WithNilStakingDataProviderShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.StakingDataProvider = nil
	vp, err := NewValidatorsProvider(arg)

	assert.Equal(t, process.ErrNilStakingDataProvider, err)
	assert.True(t, check.IfNil(vp))
}

func TestNewValidatorsProvider_WithNilNodesCoordinatorShouldErr(t *testing.T) {
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

func TestNewValidatorsProvider_WithZeroRefreshCacheIntervalInSecShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.CacheRefreshIntervalDurationInSec = 0
	vp, err := NewValidatorsProvider(arg)

	assert.Equal(t, process.ErrInvalidCacheRefreshIntervalInSec, err)
	assert.True(t, check.IfNil(vp))
}

func TestNewValidatorsProvider_WithNilAuctionListSelectorShouldErr(t *testing.T) {
	arg := createDefaultValidatorsProviderArg()
	arg.AuctionListSelector = nil
	vp, err := NewValidatorsProvider(arg)

	require.Nil(t, vp)
	require.Equal(t, epochStart.ErrNilAuctionListSelector, err)
}

func TestValidatorsProvider_GetLatestValidatorsSecondHashDoesNotExist(t *testing.T) {
	mut := sync.Mutex{}
	root := []byte("rootHash")
	e := errors.Errorf("not ok")
	initialInfo := createMockValidatorInfo()

	validatorInfos := state.NewShardValidatorsInfoMap()
	_ = validatorInfos.Add(initialInfo)

	gotOk := false
	gotNil := false
	vs := &testscommon.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() (bytes []byte) {
			mut.Lock()
			defer mut.Unlock()
			return root
		},
		GetValidatorInfoForRootHashCalled: func(rootHash []byte) (m state.ShardValidatorsInfoMapHandler, err error) {
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

	nc := &shardingMocks.NodesCoordinatorMock{GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (map[uint32][][]byte, error) {
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

	arg.ValidatorStatistics = &testscommon.ValidatorStatisticsProcessorStub{
		GetValidatorInfoForRootHashCalled: func(rootHash []byte) (state.ShardValidatorsInfoMapHandler, error) {
			atomic.AddInt32(&numPopulateCacheCalled, 1)
			return state.NewShardValidatorsInfoMap(), nil
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

	validatorProc := &testscommon.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() []byte {
			return []byte("rootHash")
		},
	}
	validatorProc.GetValidatorInfoForRootHashCalled = func(rootHash []byte) (state.ShardValidatorsInfoMapHandler, error) {
		return nil, expectedErr
	}

	pk := []byte("pk")
	nodesCoordinator := shardingMocks.NewNodesCoordinatorMock()
	nodesCoordinator.GetAllEligibleValidatorsPublicKeysCalled = func(epoch uint32) (map[uint32][][]byte, error) {
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
		validatorPubKeyConverter:     testscommon.NewPubkeyConverterMock(32),
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
		cache:                        make(map[string]*validator.ValidatorStatistics),
		cacheRefreshIntervalDuration: arg.CacheRefreshIntervalDurationInSec,
		refreshCache:                 make(chan uint32),
		lock:                         sync.RWMutex{},
		stakingDataProvider:          &stakingcommon.StakingDataProviderStub{},
		auctionListSelector:          &stakingcommon.AuctionListSelectorStub{},
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
	initialList := string(common.EligibleList)
	validatorsMap := state.NewShardValidatorsInfoMap()
	_ = validatorsMap.Add(&state.ValidatorInfo{
		PublicKey: pk,
		List:      initialList,
		ShardId:   initialShardId,
	})

	arg := createDefaultValidatorsProviderArg()
	validatorProc := &testscommon.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() []byte {
			return []byte("rootHash")
		},
	}
	validatorProc.GetValidatorInfoForRootHashCalled = func(rootHash []byte) (state.ShardValidatorsInfoMapHandler, error) {
		return validatorsMap, nil
	}

	vsp := validatorsProvider{
		nodesCoordinator:             arg.NodesCoordinator,
		validatorStatistics:          validatorProc,
		cache:                        nil,
		cacheRefreshIntervalDuration: arg.CacheRefreshIntervalDurationInSec,
		refreshCache:                 nil,
		validatorPubKeyConverter:     testscommon.NewPubkeyConverterMock(32),
		lock:                         sync.RWMutex{},
	}

	vsp.updateCache()

	assert.NotNil(t, vsp.cache)
	assert.Equal(t, len(validatorsMap.GetShardValidatorsInfoMap()[initialShardId]), len(vsp.cache))
	encodedKey, _ := arg.ValidatorPubKeyConverter.Encode(pk)
	assert.NotNil(t, vsp.cache[encodedKey])
	assert.Equal(t, initialList, vsp.cache[encodedKey].ValidatorStatus)
	assert.Equal(t, initialShardId, vsp.cache[encodedKey].ShardId)
}

func TestValidatorsProvider_aggregatePType_equal(t *testing.T) {
	pubKeyConverter := testscommon.NewPubkeyConverterMock(32)
	pkInactive := []byte("pk1")
	trieInctiveShardId := uint32(0)
	inactiveList := string(common.InactiveList)
	pkEligible := []byte("pk2")
	trieEligibleShardId := uint32(1)
	eligibleList := string(common.EligibleList)
	pkLeaving := []byte("pk3")
	trieLeavingShardId := uint32(2)
	leavingList := string(common.LeavingList)

	encodedEligible, _ := pubKeyConverter.Encode(pkEligible)
	encondedInactive, _ := pubKeyConverter.Encode(pkInactive)
	encodedLeaving, _ := pubKeyConverter.Encode(pkLeaving)
	cache := make(map[string]*validator.ValidatorStatistics)
	cache[encondedInactive] = &validator.ValidatorStatistics{ValidatorStatus: inactiveList, ShardId: trieInctiveShardId}
	cache[encodedEligible] = &validator.ValidatorStatistics{ValidatorStatus: eligibleList, ShardId: trieEligibleShardId}
	cache[encodedLeaving] = &validator.ValidatorStatistics{ValidatorStatus: leavingList, ShardId: trieLeavingShardId}

	nodesCoordinatorEligibleShardId := uint32(0)
	nodesCoordinatorLeavingShardId := core.MetachainShardId

	validatorsMap := map[uint32][][]byte{
		nodesCoordinatorEligibleShardId: {pkEligible},
		nodesCoordinatorLeavingShardId:  {pkLeaving},
	}

	vp := validatorsProvider{
		validatorPubKeyConverter: pubKeyConverter,
	}

	vp.aggregateLists(cache, validatorsMap, common.EligibleList)

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
	eligibleList := string(common.EligibleList)
	pkWaiting := []byte("pk2")
	waitingList := string(common.WaitingList)
	pkLeaving := []byte("pk3")
	leavingList := string(common.LeavingList)
	pkInactive := []byte("pk4")
	inactiveList := string(common.InactiveList)
	pkNew := []byte("pk5")
	newList := string(common.NewList)

	validatorsMap := state.NewShardValidatorsInfoMap()
	eligibleShardId := uint32(0)
	waitingShardId := uint32(1)
	leavingShardId := uint32(2)
	inactiveShardId := uint32(3)
	newShardId := core.MetachainShardId
	_ = validatorsMap.Add(&state.ValidatorInfo{
		PublicKey: pkEligible,
		ShardId:   eligibleShardId,
		List:      eligibleList,
	})
	_ = validatorsMap.Add(&state.ValidatorInfo{

		PublicKey: pkWaiting,
		ShardId:   waitingShardId,
		List:      waitingList,
	})
	_ = validatorsMap.Add(&state.ValidatorInfo{

		PublicKey: pkLeaving,
		ShardId:   leavingShardId,
		List:      leavingList,
	})
	_ = validatorsMap.Add(&state.ValidatorInfo{

		PublicKey: pkInactive,
		ShardId:   inactiveShardId,
		List:      inactiveList,
	})
	_ = validatorsMap.Add(&state.ValidatorInfo{

		PublicKey: pkNew,
		ShardId:   newShardId,
		List:      newList,
	})
	arg := createDefaultValidatorsProviderArg()
	pubKeyConverter := testscommon.NewPubkeyConverterMock(32)
	vsp := validatorsProvider{
		nodesCoordinator:             arg.NodesCoordinator,
		validatorStatistics:          arg.ValidatorStatistics,
		cache:                        nil,
		cacheRefreshIntervalDuration: arg.CacheRefreshIntervalDurationInSec,
		validatorPubKeyConverter:     pubKeyConverter,
		lock:                         sync.RWMutex{},
	}

	cache := vsp.createNewCache(0, validatorsMap)

	assert.NotNil(t, cache)

	encodedPkEligible, _ := pubKeyConverter.Encode(pkEligible)
	assert.NotNil(t, cache[encodedPkEligible])
	assert.Equal(t, eligibleList, cache[encodedPkEligible].ValidatorStatus)
	assert.Equal(t, eligibleShardId, cache[encodedPkEligible].ShardId)

	encodedPkWaiting, _ := pubKeyConverter.Encode(pkWaiting)
	assert.NotNil(t, cache[encodedPkWaiting])
	assert.Equal(t, waitingList, cache[encodedPkWaiting].ValidatorStatus)
	assert.Equal(t, waitingShardId, cache[encodedPkWaiting].ShardId)

	encodedPkLeaving, _ := pubKeyConverter.Encode(pkLeaving)
	assert.NotNil(t, cache[encodedPkLeaving])
	assert.Equal(t, leavingList, cache[encodedPkLeaving].ValidatorStatus)
	assert.Equal(t, leavingShardId, cache[encodedPkLeaving].ShardId)

	encodedPkNew, _ := pubKeyConverter.Encode(pkNew)
	assert.NotNil(t, cache[encodedPkNew])
	assert.Equal(t, newList, cache[encodedPkNew].ValidatorStatus)
	assert.Equal(t, newShardId, cache[encodedPkNew].ShardId)
}

func TestValidatorsProvider_createCache_combined(t *testing.T) {
	pkEligibleInTrie := []byte("pk1")
	eligibleList := string(common.EligibleList)
	pkInactive := []byte("pk2")
	inactiveList := string(common.InactiveList)
	pkLeavingInTrie := []byte("pk3")
	leavingList := string(common.LeavingList)

	validatorsMap := state.NewShardValidatorsInfoMap()
	eligibleShardId := uint32(0)
	inactiveShardId := uint32(1)
	leavingShardId := uint32(2)
	_ = validatorsMap.Add(&state.ValidatorInfo{
		PublicKey: pkEligibleInTrie,
		ShardId:   eligibleShardId,
		List:      eligibleList,
	})
	_ = validatorsMap.Add(&state.ValidatorInfo{
		PublicKey: pkInactive,
		ShardId:   inactiveShardId,
		List:      inactiveList,
	})
	_ = validatorsMap.Add(&state.ValidatorInfo{
		PublicKey: pkLeavingInTrie,
		ShardId:   leavingShardId,
		List:      leavingList,
	})
	arg := createDefaultValidatorsProviderArg()
	nodesCoordinator := shardingMocks.NewNodesCoordinatorMock()
	nodesCoordinatorEligibleShardId := uint32(5)
	nodesCoordinatorLeavingShardId := uint32(6)
	nodesCoordinator.GetAllEligibleValidatorsPublicKeysCalled = func(epoch uint32) (map[uint32][][]byte, error) {
		return map[uint32][][]byte{
			nodesCoordinatorEligibleShardId: {pkEligibleInTrie},
			nodesCoordinatorLeavingShardId:  {pkLeavingInTrie},
		}, nil
	}

	vsp := validatorsProvider{
		nodesCoordinator:             nodesCoordinator,
		validatorStatistics:          arg.ValidatorStatistics,
		validatorPubKeyConverter:     arg.ValidatorPubKeyConverter,
		cache:                        nil,
		cacheRefreshIntervalDuration: arg.CacheRefreshIntervalDurationInSec,
		lock:                         sync.RWMutex{},
	}

	cache := vsp.createNewCache(0, validatorsMap)

	encodedPkEligible, _ := arg.ValidatorPubKeyConverter.Encode(pkEligibleInTrie)
	assert.NotNil(t, cache[encodedPkEligible])
	assert.Equal(t, eligibleList, cache[encodedPkEligible].ValidatorStatus)
	assert.Equal(t, nodesCoordinatorEligibleShardId, cache[encodedPkEligible].ShardId)

	encodedPkLeavingInTrie, _ := arg.ValidatorPubKeyConverter.Encode(pkLeavingInTrie)
	computedPeerType := fmt.Sprintf(common.CombinedPeerType, common.EligibleList, common.LeavingList)
	assert.NotNil(t, cache[encodedPkLeavingInTrie])
	assert.Equal(t, computedPeerType, cache[encodedPkLeavingInTrie].ValidatorStatus)
	assert.Equal(t, nodesCoordinatorLeavingShardId, cache[encodedPkLeavingInTrie].ShardId)
}

func TestValidatorsProvider_CallsPopulateOnlyAfterTimeout(t *testing.T) {
	zeroNumner := int32(0)
	populateCacheCalled := &zeroNumner

	arg := createDefaultValidatorsProviderArg()
	arg.CacheRefreshIntervalDurationInSec = time.Millisecond * 10
	validatorStatisticsProcessor := &testscommon.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() []byte {
			return []byte("rootHash")
		},
	}
	validatorStatisticsProcessor.GetValidatorInfoForRootHashCalled = func(rootHash []byte) (state.ShardValidatorsInfoMapHandler, error) {
		atomic.AddInt32(populateCacheCalled, 1)
		return state.NewShardValidatorsInfoMap(), nil
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

	// outside refreshInterval
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

	validatorStatisticsProcessor := &testscommon.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() []byte {
			return []byte("rootHash")
		},
	}
	validatorStatisticsProcessor.GetValidatorInfoForRootHashCalled = func(rootHash []byte) (state.ShardValidatorsInfoMapHandler, error) {
		callNumber++
		// first call comes from the constructor
		if callNumber == 1 {
			return state.NewShardValidatorsInfoMap(), nil
		}
		validatorsMap := state.NewShardValidatorsInfoMap()
		_ = validatorsMap.Add(&state.ValidatorInfo{
			ShardId:   0,
			PublicKey: pkEligibleInTrie,
			List:      string(common.EligibleList),
		})
		return validatorsMap, nil
	}
	arg.ValidatorStatistics = validatorStatisticsProcessor

	vsp, _ := NewValidatorsProvider(arg)
	encodedEligible, _ := arg.ValidatorPubKeyConverter.Encode(pkEligibleInTrie)
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

	validatorStatisticsProcessor := &testscommon.ValidatorStatisticsProcessorStub{
		LastFinalizedRootHashCalled: func() []byte {
			return []byte("rootHash")
		},
	}
	validatorStatisticsProcessor.GetValidatorInfoForRootHashCalled = func(rootHash []byte) (state.ShardValidatorsInfoMapHandler, error) {
		callNumber++
		// first call comes from the constructor
		if callNumber == 1 {
			return state.NewShardValidatorsInfoMap(), nil
		}
		validatorsMap := state.NewShardValidatorsInfoMap()
		_ = validatorsMap.Add(&state.ValidatorInfo{
			ShardId:   0,
			PublicKey: pkEligibleInTrie,
			List:      string(common.EligibleList),
		})
		return validatorsMap, nil
	}
	arg.ValidatorStatistics = validatorStatisticsProcessor

	vsp, _ := NewValidatorsProvider(arg)
	encodedEligible, _ := arg.ValidatorPubKeyConverter.Encode(pkEligibleInTrie)
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

func TestValidatorsProvider_GetAuctionList(t *testing.T) {
	t.Parallel()

	t.Run("error getting root hash", func(t *testing.T) {
		t.Parallel()
		args := createDefaultValidatorsProviderArg()
		args.ValidatorStatistics = &testscommon.ValidatorStatisticsProcessorStub{
			LastFinalizedRootHashCalled: func() []byte {
				return nil
			},
		}
		vp, _ := NewValidatorsProvider(args)
		time.Sleep(args.CacheRefreshIntervalDurationInSec)

		list, err := vp.GetAuctionList()
		require.Nil(t, list)
		require.Equal(t, state.ErrNilRootHash, err)
	})

	t.Run("error getting validators info for root hash", func(t *testing.T) {
		t.Parallel()
		args := createDefaultValidatorsProviderArg()
		expectedErr := errors.New("local error")
		expectedRootHash := []byte("root hash")
		args.ValidatorStatistics = &testscommon.ValidatorStatisticsProcessorStub{
			LastFinalizedRootHashCalled: func() []byte {
				return expectedRootHash
			},
			GetValidatorInfoForRootHashCalled: func(rootHash []byte) (state.ShardValidatorsInfoMapHandler, error) {
				require.Equal(t, expectedRootHash, rootHash)
				return nil, expectedErr
			},
		}
		vp, _ := NewValidatorsProvider(args)
		time.Sleep(args.CacheRefreshIntervalDurationInSec)

		list, err := vp.GetAuctionList()
		require.Nil(t, list)
		require.Equal(t, expectedErr, err)
	})

	t.Run("error filling validator info, staking data provider cache should be cleaned", func(t *testing.T) {
		t.Parallel()
		args := createDefaultValidatorsProviderArg()

		cleanCalled := &coreAtomic.Flag{}
		expectedValidator := &state.ValidatorInfo{PublicKey: []byte("pubKey"), List: string(common.AuctionList)}
		expectedErr := errors.New("local error")
		expectedRootHash := []byte("root hash")
		args.ValidatorStatistics = &testscommon.ValidatorStatisticsProcessorStub{
			LastFinalizedRootHashCalled: func() []byte {
				return expectedRootHash
			},
			GetValidatorInfoForRootHashCalled: func(rootHash []byte) (state.ShardValidatorsInfoMapHandler, error) {
				require.Equal(t, expectedRootHash, rootHash)
				validatorsMap := state.NewShardValidatorsInfoMap()
				_ = validatorsMap.Add(expectedValidator)
				return validatorsMap, nil
			},
		}
		args.StakingDataProvider = &stakingcommon.StakingDataProviderStub{
			FillValidatorInfoCalled: func(validator state.ValidatorInfoHandler) error {
				require.Equal(t, expectedValidator, validator)
				return expectedErr
			},
			CleanCalled: func() {
				cleanCalled.SetValue(true)
			},
		}
		vp, _ := NewValidatorsProvider(args)
		time.Sleep(args.CacheRefreshIntervalDurationInSec)

		list, err := vp.GetAuctionList()
		require.Nil(t, list)
		require.Equal(t, expectedErr, err)
		require.True(t, cleanCalled.IsSet())
	})

	t.Run("error selecting nodes from auction, staking data provider cache should be cleaned", func(t *testing.T) {
		t.Parallel()
		args := createDefaultValidatorsProviderArg()

		cleanCalled := &coreAtomic.Flag{}
		expectedErr := errors.New("local error")
		args.AuctionListSelector = &stakingcommon.AuctionListSelectorStub{
			SelectNodesFromAuctionListCalled: func(validatorsInfoMap state.ShardValidatorsInfoMapHandler, randomness []byte) error {
				return expectedErr
			},
		}
		args.StakingDataProvider = &stakingcommon.StakingDataProviderStub{
			CleanCalled: func() {
				cleanCalled.SetValue(true)
			},
		}
		vp, _ := NewValidatorsProvider(args)
		time.Sleep(args.CacheRefreshIntervalDurationInSec)

		list, err := vp.GetAuctionList()
		require.Nil(t, list)
		require.Equal(t, expectedErr, err)
		require.True(t, cleanCalled.IsSet())
	})

	t.Run("empty list, check normal flow is executed", func(t *testing.T) {
		t.Parallel()
		args := createDefaultValidatorsProviderArg()

		expectedRootHash := []byte("root hash")
		ctRootHashCalled := uint32(0)
		ctGetValidatorsInfoForRootHash := uint32(0)
		ctSelectNodesFromAuctionList := uint32(0)
		ctFillValidatorInfoCalled := uint32(0)
		ctGetOwnersDataCalled := uint32(0)
		ctComputeUnqualifiedNodes := uint32(0)

		args.ValidatorStatistics = &testscommon.ValidatorStatisticsProcessorStub{
			LastFinalizedRootHashCalled: func() []byte {
				atomic.AddUint32(&ctRootHashCalled, 1)
				return expectedRootHash
			},
			GetValidatorInfoForRootHashCalled: func(rootHash []byte) (state.ShardValidatorsInfoMapHandler, error) {
				atomic.AddUint32(&ctGetValidatorsInfoForRootHash, 1)
				require.Equal(t, expectedRootHash, rootHash)
				return state.NewShardValidatorsInfoMap(), nil
			},
		}
		args.AuctionListSelector = &stakingcommon.AuctionListSelectorStub{
			SelectNodesFromAuctionListCalled: func(validatorsInfoMap state.ShardValidatorsInfoMapHandler, randomness []byte) error {
				atomic.AddUint32(&ctSelectNodesFromAuctionList, 1)
				require.Equal(t, expectedRootHash, randomness)
				return nil
			},
		}
		args.StakingDataProvider = &stakingcommon.StakingDataProviderStub{
			FillValidatorInfoCalled: func(validator state.ValidatorInfoHandler) error {
				atomic.AddUint32(&ctFillValidatorInfoCalled, 1)
				return nil
			},
			GetOwnersDataCalled: func() map[string]*epochStart.OwnerData {
				atomic.AddUint32(&ctGetOwnersDataCalled, 1)
				return nil
			},
			ComputeUnQualifiedNodesCalled: func(validatorInfos state.ShardValidatorsInfoMapHandler) ([][]byte, map[string][][]byte, error) {
				atomic.AddUint32(&ctComputeUnqualifiedNodes, 1)
				return nil, nil, nil
			},
		}
		vp, _ := NewValidatorsProvider(args)
		time.Sleep(args.CacheRefreshIntervalDurationInSec)

		list, err := vp.GetAuctionList()
		require.Nil(t, err)
		require.Empty(t, list)
		require.Equal(t, ctRootHashCalled, uint32(2))               // another call is from constructor in startRefreshProcess.updateCache
		require.Equal(t, ctGetValidatorsInfoForRootHash, uint32(2)) // another call is from constructor in startRefreshProcess.updateCache
		require.Equal(t, ctFillValidatorInfoCalled, uint32(0))
		require.Equal(t, ctGetOwnersDataCalled, uint32(1))
		require.Equal(t, ctComputeUnqualifiedNodes, uint32(1))
		require.Equal(t, expectedRootHash, vp.cachedRandomness)
	})

	t.Run("normal flow, check data is correctly computed", func(t *testing.T) {
		t.Parallel()
		args := createDefaultValidatorsProviderArg()

		v1 := &state.ValidatorInfo{PublicKey: []byte("pk1"), List: string(common.AuctionList)}
		v2 := &state.ValidatorInfo{PublicKey: []byte("pk2"), List: string(common.AuctionList)}
		v3 := &state.ValidatorInfo{PublicKey: []byte("pk3"), List: string(common.AuctionList)}
		v4 := &state.ValidatorInfo{PublicKey: []byte("pk4"), List: string(common.AuctionList)}
		v5 := &state.ValidatorInfo{PublicKey: []byte("pk5"), List: string(common.AuctionList)}
		v6 := &state.ValidatorInfo{PublicKey: []byte("pk6"), List: string(common.AuctionList)}
		v7 := &state.ValidatorInfo{PublicKey: []byte("pk7"), List: string(common.EligibleList)}
		v8 := &state.ValidatorInfo{PublicKey: []byte("pk8"), List: string(common.WaitingList)}
		v9 := &state.ValidatorInfo{PublicKey: []byte("pk9"), List: string(common.LeavingList)}
		v10 := &state.ValidatorInfo{PublicKey: []byte("pk10"), List: string(common.JailedList)}
		v11 := &state.ValidatorInfo{PublicKey: []byte("pk11"), List: string(common.AuctionList)}
		v12 := &state.ValidatorInfo{PublicKey: []byte("pk12"), List: string(common.AuctionList)}

		owner1 := "owner1"
		owner2 := "owner2"
		owner3 := "owner3"
		owner4 := "owner4"
		owner5 := "owner5"
		owner6 := "owner6"
		owner7 := "owner7"
		ownersData := map[string]*epochStart.OwnerData{
			owner1: {
				NumStakedNodes: 3,
				NumActiveNodes: 1,
				TotalTopUp:     big.NewInt(7500),
				TopUpPerNode:   big.NewInt(2500),
				AuctionList:    []state.ValidatorInfoHandler{v1, v2}, // owner1 will have v1 & v2 selected
				Qualified:      true,                                 // with qualifiedTopUp = 2500
			},
			owner2: {
				NumStakedNodes: 3,
				NumActiveNodes: 1,
				TotalTopUp:     big.NewInt(3000),
				TopUpPerNode:   big.NewInt(1000),
				AuctionList:    []state.ValidatorInfoHandler{v3, v4}, // owner2 will have v3 selected
				Qualified:      true,                                 // with qualifiedTopUp = 1500
			},
			owner3: {
				NumStakedNodes: 2,
				NumActiveNodes: 0,
				TotalTopUp:     big.NewInt(4000),
				TopUpPerNode:   big.NewInt(2000),
				AuctionList:    []state.ValidatorInfoHandler{v5, v6}, // owner3 will have v5 selected
				Qualified:      true,                                 // with qualifiedTopUp = 4000
			},
			owner4: {
				NumStakedNodes: 3,
				NumActiveNodes: 2,
				TotalTopUp:     big.NewInt(0),
				TopUpPerNode:   big.NewInt(0),
				AuctionList:    []state.ValidatorInfoHandler{v7}, // owner4 has one node in auction, but is not qualified
				Qualified:      false,                            // should be sent at the bottom of the list
			},
			owner5: {
				NumStakedNodes: 5,
				NumActiveNodes: 5,
				TotalTopUp:     big.NewInt(5000),
				TopUpPerNode:   big.NewInt(1000),
				AuctionList:    []state.ValidatorInfoHandler{}, // owner5 has no nodes in auction, will not appear in API list
				Qualified:      true,
			},
			// owner6 has same stats as owner7. After selection, owner7 will have its node selected => should be listed above owner 6
			owner6: {
				NumStakedNodes: 1,
				NumActiveNodes: 0,
				TotalTopUp:     big.NewInt(0),
				TopUpPerNode:   big.NewInt(0),
				AuctionList:    []state.ValidatorInfoHandler{v11},
				Qualified:      true, // should be added
			},
			owner7: {
				NumStakedNodes: 1,
				NumActiveNodes: 0,
				TotalTopUp:     big.NewInt(0),
				TopUpPerNode:   big.NewInt(0),
				AuctionList:    []state.ValidatorInfoHandler{v12},
				Qualified:      true,
			},
		}

		validatorsMap := state.NewShardValidatorsInfoMap()
		_ = validatorsMap.Add(v1)
		_ = validatorsMap.Add(v2)
		_ = validatorsMap.Add(v3)
		_ = validatorsMap.Add(v4)
		_ = validatorsMap.Add(v5)
		_ = validatorsMap.Add(v6)
		_ = validatorsMap.Add(v7)
		_ = validatorsMap.Add(v8)
		_ = validatorsMap.Add(v9)
		_ = validatorsMap.Add(v10)
		_ = validatorsMap.Add(v11)
		_ = validatorsMap.Add(v12)

		rootHash := []byte("root hash")
		args.ValidatorStatistics = &testscommon.ValidatorStatisticsProcessorStub{
			LastFinalizedRootHashCalled: func() []byte {
				return rootHash
			},
			GetValidatorInfoForRootHashCalled: func(rootHash []byte) (state.ShardValidatorsInfoMapHandler, error) {
				return validatorsMap, nil
			},
		}
		args.AuctionListSelector = &stakingcommon.AuctionListSelectorStub{
			SelectNodesFromAuctionListCalled: func(validatorsInfoMap state.ShardValidatorsInfoMapHandler, randomness []byte) error {
				selectedV1 := v1.ShallowClone()
				selectedV1.SetList(string(common.SelectedFromAuctionList))
				_ = validatorsInfoMap.Replace(v1, selectedV1)

				selectedV2 := v2.ShallowClone()
				selectedV2.SetList(string(common.SelectedFromAuctionList))
				_ = validatorsInfoMap.Replace(v2, selectedV2)

				selectedV3 := v3.ShallowClone()
				selectedV3.SetList(string(common.SelectedFromAuctionList))
				_ = validatorsInfoMap.Replace(v3, selectedV3)

				selectedV5 := v5.ShallowClone()
				selectedV5.SetList(string(common.SelectedFromAuctionList))
				_ = validatorsInfoMap.Replace(v5, selectedV5)

				selectedV12 := v12.ShallowClone()
				selectedV12.SetList(string(common.SelectedFromAuctionList))
				_ = validatorsInfoMap.Replace(v12, selectedV12)

				return nil
			},
		}
		args.StakingDataProvider = &stakingcommon.StakingDataProviderStub{
			GetOwnersDataCalled: func() map[string]*epochStart.OwnerData {
				return ownersData
			},
		}

		vp, _ := NewValidatorsProvider(args)
		time.Sleep(args.CacheRefreshIntervalDurationInSec)

		expectedList := []*common.AuctionListValidatorAPIResponse{
			{
				Owner:          args.AddressPubKeyConverter.SilentEncode([]byte(owner3), log),
				NumStakedNodes: 2,
				TotalTopUp:     "4000",
				TopUpPerNode:   "2000",
				QualifiedTopUp: "4000",
				AuctionList: []*common.AuctionNode{
					{
						BlsKey:    args.ValidatorPubKeyConverter.SilentEncode(v5.PublicKey, log),
						Qualified: true,
					},
					{
						BlsKey:    args.ValidatorPubKeyConverter.SilentEncode(v6.PublicKey, log),
						Qualified: false,
					},
				},
			},
			{
				Owner:          args.AddressPubKeyConverter.SilentEncode([]byte(owner1), log),
				NumStakedNodes: 3,
				TotalTopUp:     "7500",
				TopUpPerNode:   "2500",
				QualifiedTopUp: "2500",
				AuctionList: []*common.AuctionNode{
					{
						BlsKey:    args.ValidatorPubKeyConverter.SilentEncode(v1.PublicKey, log),
						Qualified: true,
					},
					{
						BlsKey:    args.ValidatorPubKeyConverter.SilentEncode(v2.PublicKey, log),
						Qualified: true,
					},
				},
			},
			{
				Owner:          args.AddressPubKeyConverter.SilentEncode([]byte(owner2), log),
				NumStakedNodes: 3,
				TotalTopUp:     "3000",
				TopUpPerNode:   "1000",
				QualifiedTopUp: "1500",
				AuctionList: []*common.AuctionNode{
					{
						BlsKey:    args.ValidatorPubKeyConverter.SilentEncode(v3.PublicKey, log),
						Qualified: true,
					},
					{
						BlsKey:    args.ValidatorPubKeyConverter.SilentEncode(v4.PublicKey, log),
						Qualified: false,
					},
				},
			},
			{
				Owner:          args.AddressPubKeyConverter.SilentEncode([]byte(owner7), log),
				NumStakedNodes: 1,
				TotalTopUp:     "0",
				TopUpPerNode:   "0",
				QualifiedTopUp: "0",
				AuctionList: []*common.AuctionNode{
					{
						BlsKey:    args.ValidatorPubKeyConverter.SilentEncode(v12.PublicKey, log),
						Qualified: true,
					},
				},
			},
			{
				Owner:          args.AddressPubKeyConverter.SilentEncode([]byte(owner6), log),
				NumStakedNodes: 1,
				TotalTopUp:     "0",
				TopUpPerNode:   "0",
				QualifiedTopUp: "0",
				AuctionList: []*common.AuctionNode{
					{
						BlsKey:    args.ValidatorPubKeyConverter.SilentEncode(v11.PublicKey, log),
						Qualified: false,
					},
				},
			},
			{
				Owner:          args.AddressPubKeyConverter.SilentEncode([]byte(owner4), log),
				NumStakedNodes: 3,
				TotalTopUp:     "0",
				TopUpPerNode:   "0",
				QualifiedTopUp: "0",
				AuctionList: []*common.AuctionNode{
					{
						BlsKey:    args.ValidatorPubKeyConverter.SilentEncode(v7.PublicKey, log),
						Qualified: false,
					},
				},
			},
		}

		list, err := vp.GetAuctionList()
		require.Nil(t, err)
		require.Equal(t, expectedList, list)
	})

}

func createMockValidatorInfo() *state.ValidatorInfo {
	initialInfo := &state.ValidatorInfo{
		PublicKey:                  []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
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

func createMockShardValidatorInfo() *state.ShardValidatorInfo {
	initialInfo := &state.ShardValidatorInfo{
		PublicKey:  bytes.Repeat([]byte("a"), 96),
		ShardId:    0,
		List:       "eligible",
		Index:      1,
		TempRating: 100,
	}
	return initialInfo
}

func createDefaultValidatorsProviderArg() ArgValidatorsProvider {
	return ArgValidatorsProvider{
		NodesCoordinator:                  &shardingMocks.NodesCoordinatorMock{},
		StartEpoch:                        1,
		EpochStartEventNotifier:           &mock.EpochStartNotifierStub{},
		StakingDataProvider:               &stakingcommon.StakingDataProviderStub{},
		CacheRefreshIntervalDurationInSec: 1 * time.Millisecond,
		ValidatorStatistics: &testscommon.ValidatorStatisticsProcessorStub{
			LastFinalizedRootHashCalled: func() []byte {
				return []byte("rootHash")
			},
		},
		MaxRating:                100,
		ValidatorPubKeyConverter: testscommon.NewPubkeyConverterMock(32),
		AddressPubKeyConverter:   testscommon.NewPubkeyConverterMock(32),
		AuctionListSelector:      &stakingcommon.AuctionListSelectorStub{},
	}
}
