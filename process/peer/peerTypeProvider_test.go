package peer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

const defaultRefreshIntervalDuration = 1 * time.Millisecond

func TestNewPeerTypeProvider_NilNodesCoordinator(t *testing.T) {
	arg := createDefaultArg()
	arg.NodesCoordinator = nil

	ptp, err := NewPeerTypeProvider(arg)
	assert.Nil(t, ptp)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewPeerTypeProvider_NilEpochHandler(t *testing.T) {
	arg := createDefaultArg()
	arg.EpochHandler = nil

	ptp, err := NewPeerTypeProvider(arg)
	assert.Nil(t, ptp)
	assert.Equal(t, process.ErrNilEpochHandler, err)
}

func TestNewPeerTypeProvider_NilEpochStartNotifier(t *testing.T) {
	arg := createDefaultArg()
	arg.EpochStartEventNotifier = nil

	ptp, err := NewPeerTypeProvider(arg)
	assert.Nil(t, ptp)
	assert.Equal(t, process.ErrNilEpochStartNotifier, err)
}

func TestNewPeerTypeProvider_ZeroCacheRefreshInterval(t *testing.T) {
	arg := createDefaultArg()
	arg.PeerTypeRefreshIntervalInSec = 0

	ptp, err := NewPeerTypeProvider(arg)
	assert.Nil(t, ptp)
	assert.Equal(t, process.ErrInvalidPeerTypeRefreshIntervalInSec, err)
}

func TestNewPeerTypeProvider_NilContext(t *testing.T) {
	arg := createDefaultArg()
	arg.Context = nil

	ptp, err := NewPeerTypeProvider(arg)
	assert.Nil(t, ptp)
	assert.Equal(t, process.ErrNilContext, err)
}

func TestNewPeerTypeProvider_NilValidatorsProvider(t *testing.T) {
	arg := createDefaultArg()
	arg.ValidatorsProvider = nil

	ptp, err := NewPeerTypeProvider(arg)
	assert.Nil(t, ptp)
	assert.Equal(t, process.ErrNilValidatorsProvider, err)
}

func TestNewPeerTypeProvider_ShouldWork(t *testing.T) {
	arg := createDefaultArg()

	ptp, err := NewPeerTypeProvider(arg)
	assert.Nil(t, err)
	assert.NotNil(t, ptp)
}

func TestPeerTypeProvider_CallsPopulateAndRegister(t *testing.T) {
	numRegisterHandlerCalled := int32(0)
	numPopulateCacheCalled := int32(0)

	arg := createDefaultArg()
	arg.EpochStartEventNotifier = &mock.EpochStartNotifierStub{
		RegisterHandlerCalled: func(handler epochStart.ActionHandler) {
			atomic.AddInt32(&numRegisterHandlerCalled, 1)
		},
	}

	arg.ValidatorsProvider = &mock.ValidatorsProviderStub{
		GetLatestValidatorInfosCalled: func() (map[uint32][]*state.ValidatorInfo, error) {
			atomic.AddInt32(&numPopulateCacheCalled, 1)
			return nil, nil
		},
	}

	_, _ = NewPeerTypeProvider(arg)

	time.Sleep(arg.PeerTypeRefreshIntervalInSec)

	assert.Equal(t, int32(1), atomic.LoadInt32(&numPopulateCacheCalled))
	assert.Equal(t, int32(1), atomic.LoadInt32(&numRegisterHandlerCalled))
}

func TestPeerTypeProvider_UpdateCache_WithError(t *testing.T) {
	expectedErr := errors.New("expectedError")
	arg := createDefaultArg()

	validatorProviderStub := &mock.ValidatorsProviderStub{}
	arg.ValidatorsProvider = validatorProviderStub
	validatorProviderStub.GetLatestValidatorInfosCalled = func() (map[uint32][]*state.ValidatorInfo, error) {
		return nil, expectedErr
	}

	pk := []byte("pk")
	nodesCoordinator := mock.NewNodesCoordinatorMock()
	nodesCoordinator.GetAllEligibleValidatorsPublicKeysCalled = func() (map[uint32][][]byte, error) {
		return map[uint32][][]byte{
			0: {pk},
		}, nil
	}

	ptp := PeerTypeProvider{
		nodesCoordinator:             nodesCoordinator,
		epochHandler:                 arg.EpochHandler,
		validatorsProvider:           arg.ValidatorsProvider,
		cache:                        nil,
		cacheRefreshIntervalDuration: arg.PeerTypeRefreshIntervalInSec,
		refreshCache:                 nil,
		mutCache:                     sync.RWMutex{},
	}

	ptp.updateCache()

	assert.NotNil(t, ptp.GetCache())
	assert.Equal(t, 1, len(ptp.GetCache()))
}

func TestPeerTypeProvider_UpdateCache(t *testing.T) {
	pk := "pk1"
	initialShardId := uint32(1)
	validatorsMap := make(map[uint32][]*state.ValidatorInfo)
	validatorsMap[initialShardId] = []*state.ValidatorInfo{
		{
			PublicKey: []byte(pk),
			List:      "eligible",
		},
	}
	arg := createDefaultArg()
	arg.ValidatorsProvider = &mock.ValidatorsProviderStub{
		GetLatestValidatorInfosCalled: func() (map[uint32][]*state.ValidatorInfo, error) {
			return validatorsMap, nil
		},
	}

	ptp := PeerTypeProvider{
		nodesCoordinator:             arg.NodesCoordinator,
		epochHandler:                 arg.EpochHandler,
		validatorsProvider:           arg.ValidatorsProvider,
		cache:                        nil,
		cacheRefreshIntervalDuration: arg.PeerTypeRefreshIntervalInSec,
		refreshCache:                 nil,
		mutCache:                     sync.RWMutex{},
	}

	ptp.updateCache()

	assert.NotNil(t, ptp.cache)
	assert.Equal(t, len(validatorsMap[initialShardId]), len(ptp.cache))
	assert.NotNil(t, ptp.cache[pk])
	assert.Equal(t, core.EligibleList, ptp.cache[pk].pType)
	assert.Equal(t, initialShardId, ptp.cache[pk].pShard)
}

func TestPeerTypeProvider_aggregatePType_equal(t *testing.T) {
	pkInactive := "pk1"
	trieInctiveShardId := uint32(0)
	pkEligible := "pk2"
	trieEligibleShardId := uint32(1)
	pkLeaving := "pk3"
	trieLeavingShardId := uint32(2)

	cache := make(map[string]*peerListAndShard)
	cache[pkInactive] = &peerListAndShard{pType: core.InactiveList, pShard: trieInctiveShardId}
	cache[pkEligible] = &peerListAndShard{pType: core.EligibleList, pShard: trieEligibleShardId}
	cache[pkLeaving] = &peerListAndShard{pType: core.LeavingList, pShard: trieLeavingShardId}

	nodesCoordinatorEligibleShardId := uint32(0)
	nodesCoordinatorLeavingShardId := core.MetachainShardId

	validatorsMap := map[uint32][][]byte{
		nodesCoordinatorEligibleShardId: {[]byte(pkEligible)},
		nodesCoordinatorLeavingShardId:  {[]byte(pkLeaving)},
	}
	aggregatePType(cache, validatorsMap, core.EligibleList)

	assert.Equal(t, trieInctiveShardId, cache[pkInactive].pShard)
	assert.Equal(t, core.InactiveList, cache[pkInactive].pType)

	assert.Equal(t, nodesCoordinatorEligibleShardId, cache[pkEligible].pShard)
	assert.Equal(t, core.EligibleList, cache[pkEligible].pType)

	assert.Equal(t, nodesCoordinatorLeavingShardId, cache[pkLeaving].pShard)
	assert.Equal(t, core.PeerType("eligible (leaving)"), cache[pkLeaving].pType)
}

func TestNewPeerTypeProvider_createCache(t *testing.T) {
	pkEligible := "pk1"
	pkWaiting := "pk2"
	pkLeaving := "pk3"
	pkInactive := "pk4"
	pkNew := "pk5"

	validatorsMap := make(map[uint32][]*state.ValidatorInfo)
	eligibleShardId := uint32(0)
	waitingShardId := uint32(1)
	leavingShardId := uint32(2)
	inactiveShardId := uint32(3)
	newShardId := core.MetachainShardId
	validatorsMap[eligibleShardId] = []*state.ValidatorInfo{
		{
			PublicKey: []byte(pkEligible),
			List:      string(core.EligibleList),
		},
	}
	validatorsMap[waitingShardId] = []*state.ValidatorInfo{
		{
			PublicKey: []byte(pkWaiting),
			List:      string(core.WaitingList),
		},
	}
	validatorsMap[leavingShardId] = []*state.ValidatorInfo{
		{
			PublicKey: []byte(pkLeaving),
			List:      string(core.LeavingList),
		},
	}
	validatorsMap[inactiveShardId] = []*state.ValidatorInfo{
		{
			PublicKey: []byte(pkInactive),
			List:      string(core.InactiveList),
		},
	}
	validatorsMap[newShardId] = []*state.ValidatorInfo{
		{
			PublicKey: []byte(pkNew),
			List:      string(core.NewList),
		},
	}
	arg := createDefaultArg()

	ptp := PeerTypeProvider{
		nodesCoordinator:             arg.NodesCoordinator,
		epochHandler:                 arg.EpochHandler,
		validatorsProvider:           arg.ValidatorsProvider,
		cache:                        nil,
		cacheRefreshIntervalDuration: arg.PeerTypeRefreshIntervalInSec,
		mutCache:                     sync.RWMutex{},
	}

	cache := ptp.createNewCache(0, validatorsMap)

	assert.NotNil(t, cache)

	assert.NotNil(t, cache[pkEligible])
	assert.Equal(t, core.EligibleList, cache[pkEligible].pType)
	assert.Equal(t, eligibleShardId, cache[pkEligible].pShard)

	assert.NotNil(t, cache[pkWaiting])
	assert.Equal(t, core.WaitingList, cache[pkWaiting].pType)
	assert.Equal(t, waitingShardId, cache[pkWaiting].pShard)

	assert.NotNil(t, cache[pkLeaving])
	assert.Equal(t, core.LeavingList, cache[pkLeaving].pType)
	assert.Equal(t, leavingShardId, cache[pkLeaving].pShard)

	assert.NotNil(t, cache[pkInactive])
	assert.Equal(t, core.InactiveList, cache[pkInactive].pType)
	assert.Equal(t, inactiveShardId, cache[pkInactive].pShard)

	assert.NotNil(t, cache[pkNew])
	assert.Equal(t, core.NewList, cache[pkNew].pType)
	assert.Equal(t, newShardId, cache[pkNew].pShard)
}

func TestNewPeerTypeProvider_createCache_combined(t *testing.T) {
	pkEligibleInTrie := "pk1"
	pkInactive := "pk2"
	pkLeavingInTrie := "pk3"

	validatorsMap := make(map[uint32][]*state.ValidatorInfo)
	eligibleShardId := uint32(0)
	inactiveShardId := uint32(1)
	leavingShardId := uint32(2)
	validatorsMap[eligibleShardId] = []*state.ValidatorInfo{
		{
			PublicKey: []byte(pkEligibleInTrie),
			List:      string(core.EligibleList),
		},
	}
	validatorsMap[inactiveShardId] = []*state.ValidatorInfo{
		{
			PublicKey: []byte(pkInactive),
			List:      string(core.InactiveList),
		},
	}
	validatorsMap[leavingShardId] = []*state.ValidatorInfo{
		{
			PublicKey: []byte(pkLeavingInTrie),
			List:      string(core.LeavingList),
		},
	}
	arg := createDefaultArg()
	nodesCoordinator := mock.NewNodesCoordinatorMock()
	nodesCoordinatorEligibleShardId := uint32(5)
	nodesCoordinatorLeavingShardId := uint32(6)
	nodesCoordinator.GetAllEligibleValidatorsPublicKeysCalled = func() (map[uint32][][]byte, error) {
		return map[uint32][][]byte{
			nodesCoordinatorEligibleShardId: {[]byte(pkEligibleInTrie)},
			nodesCoordinatorLeavingShardId:  {[]byte(pkLeavingInTrie)},
		}, nil
	}

	ptp := PeerTypeProvider{
		nodesCoordinator:             nodesCoordinator,
		epochHandler:                 arg.EpochHandler,
		validatorsProvider:           arg.ValidatorsProvider,
		cache:                        nil,
		cacheRefreshIntervalDuration: arg.PeerTypeRefreshIntervalInSec,
		mutCache:                     sync.RWMutex{},
	}

	cache := ptp.createNewCache(0, validatorsMap)

	assert.NotNil(t, cache[pkEligibleInTrie])
	assert.Equal(t, core.EligibleList, cache[pkEligibleInTrie].pType)
	assert.Equal(t, nodesCoordinatorEligibleShardId, cache[pkEligibleInTrie].pShard)

	assert.NotNil(t, cache[pkInactive])
	assert.Equal(t, core.InactiveList, cache[pkInactive].pType)
	assert.Equal(t, inactiveShardId, cache[pkInactive].pShard)

	computedPeerType := core.PeerType(fmt.Sprintf(core.CombinedPeerType, core.EligibleList, core.LeavingList))
	assert.NotNil(t, cache[pkLeavingInTrie])
	assert.Equal(t, computedPeerType, cache[pkLeavingInTrie].pType)
	assert.Equal(t, nodesCoordinatorLeavingShardId, cache[pkLeavingInTrie].pShard)
}

func TestNewPeerTypeProvider_CallsPopulateOnlyAfterTimeout(t *testing.T) {
	zeroNumner := int32(0)
	populateCacheCalled := &zeroNumner

	pk := "pk"
	arg := createDefaultArg()
	arg.PeerTypeRefreshIntervalInSec = time.Millisecond * 10
	validatorsProviderStub := &mock.ValidatorsProviderStub{}
	arg.ValidatorsProvider = validatorsProviderStub
	validatorsProviderStub.GetLatestValidatorInfosCalled = func() (map[uint32][]*state.ValidatorInfo, error) {
		atomic.AddInt32(populateCacheCalled, 1)
		return nil, nil
	}

	ptp, _ := NewPeerTypeProvider(arg)

	// allow previous call to through
	time.Sleep(time.Millisecond)

	// inside refreshInterval of 10 milis
	atomic.StoreInt32(populateCacheCalled, 0)
	_, _, _ = ptp.ComputeForPubKey([]byte(pk))
	time.Sleep(time.Millisecond)
	assert.Equal(t, int32(0), atomic.LoadInt32(populateCacheCalled))
	_, _, _ = ptp.ComputeForPubKey([]byte(pk))
	time.Sleep(time.Millisecond)
	assert.Equal(t, int32(0), atomic.LoadInt32(populateCacheCalled))

	// outside of refreshInterval
	time.Sleep(arg.PeerTypeRefreshIntervalInSec)
	_, _, _ = ptp.ComputeForPubKey([]byte(pk))
	//allow call to go through
	time.Sleep(time.Millisecond)
	assert.Equal(t, int32(1), atomic.LoadInt32(populateCacheCalled))

}

func TestNewPeerTypeProvider_CallsUpdateCacheOnEpochChange(t *testing.T) {
	arg := createDefaultArg()
	callNumber := 0
	epochStartNotifier := &mock.EpochStartNotifierStub{}
	arg.EpochStartEventNotifier = epochStartNotifier
	arg.PeerTypeRefreshIntervalInSec = 10 * time.Millisecond
	pkEligibleInTrie := "pk1"
	arg.ValidatorsProvider = &mock.ValidatorsProviderStub{
		GetLatestValidatorInfosCalled: func() (map[uint32][]*state.ValidatorInfo, error) {
			callNumber++
			// first call comes from the constructor
			if callNumber == 1 {
				return nil, nil
			}
			return map[uint32][]*state.ValidatorInfo{
				0: {
					{
						PublicKey: []byte(pkEligibleInTrie),
						List:      string(core.EligibleList),
					},
				},
			}, nil
		},
	}

	ptp, _ := NewPeerTypeProvider(arg)

	assert.Equal(t, 0, len(ptp.GetCache())) // nothing in cache
	epochStartNotifier.NotifyAll(nil)
	time.Sleep(time.Millisecond)
	assert.Equal(t, 1, len(ptp.GetCache()))
	assert.NotNil(t, ptp.GetCache()[pkEligibleInTrie])
}

func TestNewPeerTypeProvider_ComputeForKeyFromCache(t *testing.T) {
	arg := createDefaultArg()
	arg.PeerTypeRefreshIntervalInSec = 10 * time.Millisecond
	pk := []byte("pk1")
	initialShardId := uint32(1)
	popMutex := sync.RWMutex{}
	populateCacheCalled := false
	validatorsMap := make(map[uint32][]*state.ValidatorInfo)
	validatorsMap[initialShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pk,
			List:      "eligible",
		},
	}
	arg.ValidatorsProvider = &mock.ValidatorsProviderStub{
		GetLatestValidatorInfosCalled: func() (map[uint32][]*state.ValidatorInfo, error) {
			popMutex.Lock()
			defer popMutex.Unlock()
			populateCacheCalled = true
			return validatorsMap, nil
		},
	}
	ptp, _ := NewPeerTypeProvider(arg)
	time.Sleep(arg.PeerTypeRefreshIntervalInSec)
	popMutex.Lock()
	populateCacheCalled = false
	popMutex.Unlock()
	peerType, shardId, err := ptp.ComputeForPubKey(pk)

	popMutex.RLock()
	called := populateCacheCalled
	popMutex.RUnlock()
	assert.False(t, called)
	assert.Equal(t, core.EligibleList, peerType)
	assert.Equal(t, initialShardId, shardId)
	assert.Nil(t, err)
}

func TestNewPeerTypeProvider_ComputeForKeyNotFoundInCacheReturnsObserverAndUpdateCache(t *testing.T) {
	arg := createDefaultArg()
	pk := []byte("pk1")
	initialShardId := uint32(1)
	callCount := 0
	validatorsMap := make(map[uint32][]*state.ValidatorInfo)
	validatorsProviderStub := &mock.ValidatorsProviderStub{}
	arg.ValidatorsProvider = validatorsProviderStub
	validatorsMap[initialShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pk,
			List:      "eligible",
		},
	}
	validatorsProviderStub.GetLatestValidatorInfosCalled = func() (map[uint32][]*state.ValidatorInfo, error) {
		callCount++
		if callCount == 1 {
			return nil, nil
		}
		return validatorsMap, nil
	}

	ptp, _ := NewPeerTypeProvider(arg)
	time.Sleep(arg.PeerTypeRefreshIntervalInSec)
	assert.Equal(t, 0, len(ptp.GetCache()))

	peerType, shardId, err := ptp.ComputeForPubKey(pk)

	assert.Equal(t, core.ObserverList, peerType)
	assert.Equal(t, uint32(0), shardId)
	assert.Nil(t, err)

	time.Sleep(arg.PeerTypeRefreshIntervalInSec)

	assert.Equal(t, 1, len(ptp.GetCache()))
}

func TestNewPeerTypeProvider_ComputeForKeyCombinedPeerType(t *testing.T) {
	arg := createDefaultArg()
	pkLeaving := []byte("pkLeaving")
	pkNew := []byte("pkNew")
	pkLeavingList := "leaving"
	pkNewList := "new"

	initialShardId := uint32(1)

	validatorsMap := make(map[uint32][]*state.ValidatorInfo)
	validatorsMap[initialShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pkLeaving,
			List:      pkLeavingList,
		},
		{
			PublicKey: pkNew,
			List:      pkNewList,
		},
	}
	arg.ValidatorsProvider = &mock.ValidatorsProviderStub{
		GetLatestValidatorInfosCalled: func() (map[uint32][]*state.ValidatorInfo, error) {
			return validatorsMap, nil
		},
	}

	eligibleMap := make(map[uint32][][]byte)
	eligibleMap[uint32(1)] = [][]byte{pkLeaving}
	waitingMap := make(map[uint32][][]byte)
	waitingMap[uint32(1)] = [][]byte{pkNew}

	arg.NodesCoordinator = &mock.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func() (map[uint32][][]byte, error) {
			return eligibleMap, nil
		},
		GetAllWaitingValidatorsPublicKeysCalled: func() (map[uint32][][]byte, error) {
			return waitingMap, nil
		},
	}

	ptp, _ := NewPeerTypeProvider(arg)
	time.Sleep(time.Millisecond)
	peerType, shardId, err := ptp.ComputeForPubKey(pkLeaving)
	combinedType := core.PeerType(fmt.Sprintf(core.CombinedPeerType, core.EligibleList, pkLeavingList))
	assert.Equal(t, combinedType, peerType)
	assert.Equal(t, initialShardId, shardId)
	assert.Nil(t, err)

	peerType, shardId, err = ptp.ComputeForPubKey(pkNew)
	combinedType = core.PeerType(fmt.Sprintf(core.CombinedPeerType, core.WaitingList, pkNewList))
	assert.Equal(t, combinedType, peerType)
	assert.Equal(t, initialShardId, shardId)
	assert.Nil(t, err)
}

func TestNewPeerTypeProvider_IsInterfaceNil(t *testing.T) {
	arg := createDefaultArg()

	ptp, _ := NewPeerTypeProvider(arg)
	assert.False(t, ptp.IsInterfaceNil())
}

func createDefaultArg() ArgPeerTypeProvider {
	return ArgPeerTypeProvider{
		NodesCoordinator:             &mock.NodesCoordinatorMock{},
		EpochHandler:                 &mock.EpochStartTriggerStub{},
		ValidatorsProvider:           &mock.ValidatorsProviderStub{},
		EpochStartEventNotifier:      &mock.EpochStartNotifierStub{},
		PeerTypeRefreshIntervalInSec: defaultRefreshIntervalDuration,
		Context:                      context.Background(),
	}
}
