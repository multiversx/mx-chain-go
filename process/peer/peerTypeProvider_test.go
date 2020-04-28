package peer

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

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
	registerHandlerCalled := false
	populateCacheCalled := false

	arg := createDefaultArg()
	arg.EpochStartEventNotifier = &mock.EpochStartNotifierStub{
		RegisterHandlerCalled: func(handler epochStart.ActionHandler) {
			registerHandlerCalled = true
		},
	}

	arg.ValidatorsProvider = &mock.ValidatorsProviderStub{
		GetLatestValidatorInfosCalled: func() (map[uint32][]*state.ValidatorInfo, error) {
			populateCacheCalled = true
			return nil, nil
		},
	}

	_, _ = NewPeerTypeProvider(arg)
	assert.True(t, populateCacheCalled)
	assert.True(t, registerHandlerCalled)
}

func TestNewPeerTypeProvider_CallsPopulateOnlyAfterTimeout(t *testing.T) {
	populateCacheCalled := 0

	arg := createDefaultArg()

	arg.ValidatorsProvider = &mock.ValidatorsProviderStub{
		GetLatestValidatorInfosCalled: func() (map[uint32][]*state.ValidatorInfo, error) {
			populateCacheCalled++
			return nil, nil
		},
	}
	arg.CacheRefreshIntervalInSec = 1

	ptp, _ := NewPeerTypeProvider(arg)
	populateCacheCalled = 0
	ptp.populateCache(0)
	assert.Equal(t, 0, populateCacheCalled)

	time.Sleep(1 * time.Second)
	ptp.populateCache(0)
	assert.Equal(t, 1, populateCacheCalled)
}

func TestNewPeerTypeProvider_CallsPopulateOnEpochChange(t *testing.T) {
	populateCacheCalled := false
	wg := sync.WaitGroup{}
	wg.Add(2)

	arg := createDefaultArg()
	arg.CacheRefreshIntervalInSec = 0
	epochStartNotifier := &mock.EpochStartNotifierStub{}
	arg.EpochStartEventNotifier = epochStartNotifier
	arg.ValidatorsProvider = &mock.ValidatorsProviderStub{
		GetLatestValidatorInfosCalled: func() (map[uint32][]*state.ValidatorInfo, error) {
			populateCacheCalled = true
			wg.Done()
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
	arg.CacheRefreshIntervalInSec = 1
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
	populateCacheCalled := false

	wg := sync.WaitGroup{}
	wg.Add(2)

	validatorsMap := make(map[uint32][]*state.ValidatorInfo)
	arg.ValidatorsProvider = &mock.ValidatorsProviderStub{
		GetLatestValidatorInfosCalled: func() (map[uint32][]*state.ValidatorInfo, error) {
			populateCacheCalled = true
			wg.Done()
			return validatorsMap, nil
		},
	}
	arg.CacheRefreshIntervalInSec = 0
	ptp, _ := NewPeerTypeProvider(arg)

	populateCacheCalled = false
	validatorsMap[initialShardId] = []*state.ValidatorInfo{
		{
			PublicKey: pk,
			List:      "eligible",
		},
	}

	peerType, shardId, err := ptp.ComputeForPubKey(pk)

	assert.False(t, populateCacheCalled)
	assert.Equal(t, core.ObserverList, peerType)
	assert.Equal(t, uint32(0), shardId)
	assert.Nil(t, err)

	wg.Wait()
	assert.True(t, populateCacheCalled)

	populateCacheCalled = false
	peerType, shardId, _ = ptp.ComputeForPubKey(pk)

	assert.False(t, populateCacheCalled)
	assert.Equal(t, core.EligibleList, peerType)
	assert.Equal(t, initialShardId, shardId)
}

func TestNewPeerTypeProvider_ComputeForKeyEligibleLeaving(t *testing.T) {
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
	arg.CacheRefreshIntervalInSec = 0

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
		NodesCoordinator:          &mock.NodesCoordinatorMock{},
		EpochHandler:              &mock.EpochStartTriggerStub{},
		ValidatorsProvider:        &mock.ValidatorsProviderStub{},
		EpochStartEventNotifier:   &mock.EpochStartNotifierStub{},
		CacheRefreshIntervalInSec: 1,
	}
}
