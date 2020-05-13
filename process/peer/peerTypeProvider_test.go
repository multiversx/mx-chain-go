package peer

import (
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
	arg.CacheRefreshIntervalDuration = 0

	ptp, err := NewPeerTypeProvider(arg)
	assert.Nil(t, ptp)
	assert.Equal(t, process.ErrInvalidCacheRefreshIntervalDuration, err)
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

func TestNewPeerTypeProvider_CallsPopulateAndRegister(t *testing.T) {
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
	assert.Equal(t, int32(1), numPopulateCacheCalled)
	assert.Equal(t, int32(1), numRegisterHandlerCalled)
}

func TestNewPeerTypeProvider_CallsPopulateWithError(t *testing.T) {
	expectedErr := errors.New("expectedError")
	arg := createDefaultArg()

	validatorProviderStub := &mock.ValidatorsProviderStub{}
	arg.ValidatorsProvider = validatorProviderStub
	ptp, _ := NewPeerTypeProvider(arg)
	time.Sleep(1 * time.Millisecond)
	wg := sync.WaitGroup{}
	wg.Add(1)
	validatorProviderStub.GetLatestValidatorInfosCalled = func() (map[uint32][]*state.ValidatorInfo, error) {
		defer wg.Done()
		assert.True(t, ptp.isUpdating)
		return nil, expectedErr
	}

	ptp.populateCache(0)
	wg.Wait()
	time.Sleep(1 * time.Millisecond)
	assert.False(t, ptp.isUpdating)
}

func TestNewPeerTypeProvider_CallsPopulateOnlyAfterTimeout(t *testing.T) {
	populateCacheCalled := int32(0)
	refreshIntervalDuration := 10 * time.Millisecond
	arg := createDefaultArg()
	arg.CacheRefreshIntervalDuration = refreshIntervalDuration
	validatorsProviderStub := &mock.ValidatorsProviderStub{}
	arg.ValidatorsProvider = validatorsProviderStub

	ptp, _ := NewPeerTypeProvider(arg)

	wg := sync.WaitGroup{}
	wg.Add(1)
	validatorsProviderStub.GetLatestValidatorInfosCalled = func() (map[uint32][]*state.ValidatorInfo, error) {
		defer wg.Done()
		atomic.AddInt32(&populateCacheCalled, 1)
		return nil, nil
	}

	// inside refreshInterval because newPeerType populated
	populateCacheCalled = 0
	ptp.populateCache(0)
	assert.Equal(t, int32(0), populateCacheCalled)

	// outside refreshInterval because newPeerType
	time.Sleep(refreshIntervalDuration)
	ptp.populateCache(0)
	wg.Wait()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, int32(1), populateCacheCalled)
}

func TestNewPeerTypeProvider_CallsPopulateOnlyOnceIfIsUpdating(t *testing.T) {
	numPopulateCacheCalled := 0

	arg := createDefaultArg()

	validatorsProviderStub := &mock.ValidatorsProviderStub{}
	arg.ValidatorsProvider = validatorsProviderStub

	ptp, _ := NewPeerTypeProvider(arg)
	time.Sleep(defaultRefreshIntervalDuration)
	wg := sync.WaitGroup{}
	wg.Add(1)
	validatorsProviderStub.GetLatestValidatorInfosCalled = func() (map[uint32][]*state.ValidatorInfo, error) {
		defer wg.Done()
		numPopulateCacheCalled++
		time.Sleep(20 * time.Millisecond)
		return nil, nil
	}

	numPopulateCacheCalled = 0
	ptp.populateCache(0)
	ptp.populateCache(0)
	ptp.populateCache(0)
	wg.Wait()
	assert.Equal(t, 1, numPopulateCacheCalled)
}

func TestNewPeerTypeProvider_CallsPopulateOnEpochChange(t *testing.T) {
	populateCacheCalled := false
	wg := sync.WaitGroup{}
	wg.Add(2)

	arg := createDefaultArg()
	arg.CacheRefreshIntervalDuration = 0
	epochStartNotifier := &mock.EpochStartNotifierStub{}
	arg.EpochStartEventNotifier = epochStartNotifier
	arg.ValidatorsProvider = &mock.ValidatorsProviderStub{
		GetLatestValidatorInfosCalled: func() (map[uint32][]*state.ValidatorInfo, error) {
			defer wg.Done()
			populateCacheCalled = true
			return nil, nil
		},
	}

	_, _ = NewPeerTypeProvider(arg)
	populateCacheCalled = false
	epochStartNotifier.NotifyAll(nil)
	wg.Wait()
	assert.True(t, populateCacheCalled)
}

func TestNewPeerTypeProvider_ComputeForKeyFromCache(t *testing.T) {
	arg := createDefaultArg()
	pk := []byte("pk1")
	initialShardId := uint32(1)
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
			populateCacheCalled = true
			return validatorsMap, nil
		},
	}
	arg.CacheRefreshIntervalDuration = 1
	ptp, _ := NewPeerTypeProvider(arg)
	populateCacheCalled = false
	peerType, shardId, err := ptp.ComputeForPubKey(pk)

	assert.False(t, populateCacheCalled)
	assert.Equal(t, core.EligibleList, peerType)
	assert.Equal(t, initialShardId, shardId)
	assert.Nil(t, err)
}

func TestNewPeerTypeProvider_ComputeForKeyNotFoundInCacheReturnsObserverAndPopulateCache(t *testing.T) {
	arg := createDefaultArg()
	pk := []byte("pk1")
	initialShardId := uint32(1)
	numPopulateCacheCalled := 1

	validatorsMap := make(map[uint32][]*state.ValidatorInfo)
	validatorsProviderStub := &mock.ValidatorsProviderStub{}
	arg.ValidatorsProvider = validatorsProviderStub
	arg.CacheRefreshIntervalDuration = 0
	ptp, _ := NewPeerTypeProvider(arg)

	wg := sync.WaitGroup{}
	wg.Add(1)
	validatorsProviderStub.GetLatestValidatorInfosCalled = func() (map[uint32][]*state.ValidatorInfo, error) {
		defer wg.Done()
		numPopulateCacheCalled++
		return validatorsMap, nil
	}

	numPopulateCacheCalled = 0
	validatorsMap[initialShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pk,
			List:      "eligible",
		},
	}

	peerType, shardId, err := ptp.ComputeForPubKey(pk)

	assert.Equal(t, core.ObserverList, peerType)
	assert.Equal(t, uint32(0), shardId)
	assert.Nil(t, err)

	wg.Wait()
	assert.Equal(t, 1, numPopulateCacheCalled)
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
	arg.CacheRefreshIntervalDuration = 0

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
		CacheRefreshIntervalDuration: defaultRefreshIntervalDuration,
	}
}
