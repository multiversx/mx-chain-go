package peer

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

func TestNewPeerTypeProvider_NilNodesCoordinator(t *testing.T) {
	arg := createDefaultArgPeerTypeProvider()
	arg.NodesCoordinator = nil

	ptp, err := NewPeerTypeProvider(arg)
	assert.Nil(t, ptp)
	assert.Equal(t, process.ErrNilNodesCoordinator, err)
}

func TestNewPeerTypeProvider_NilEpochStartNotifier(t *testing.T) {
	arg := createDefaultArgPeerTypeProvider()
	arg.EpochStartEventNotifier = nil

	ptp, err := NewPeerTypeProvider(arg)
	assert.Nil(t, ptp)
	assert.Equal(t, process.ErrNilEpochStartNotifier, err)
}

func TestNewPeerTypeProvider_ShouldWork(t *testing.T) {
	arg := createDefaultArgPeerTypeProvider()

	ptp, err := NewPeerTypeProvider(arg)
	assert.Nil(t, err)
	assert.NotNil(t, ptp)
}

func TestPeerTypeProvider_CallsPopulateAndRegister(t *testing.T) {
	numRegisterHandlerCalled := int32(0)
	numPopulateCacheCalled := int32(0)

	arg := createDefaultArgPeerTypeProvider()
	arg.EpochStartEventNotifier = &mock.EpochStartNotifierStub{
		RegisterHandlerCalled: func(handler epochStart.ActionHandler) {
			atomic.AddInt32(&numRegisterHandlerCalled, 1)
		},
	}

	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(_ uint32) (map[uint32][][]byte, error) {
			atomic.AddInt32(&numPopulateCacheCalled, 1)
			return nil, nil
		},
	}

	_, _ = NewPeerTypeProvider(arg)

	assert.Equal(t, int32(1), atomic.LoadInt32(&numPopulateCacheCalled))
	assert.Equal(t, int32(1), atomic.LoadInt32(&numRegisterHandlerCalled))
}

func TestPeerTypeProvider_UpdateCache(t *testing.T) {
	pk := "pk1"
	initialShardId := uint32(1)
	eligibleMap := make(map[uint32][][]byte)
	eligibleMap[initialShardId] = [][]byte{
		[]byte(pk),
	}
	arg := createDefaultArgPeerTypeProvider()
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(_ uint32) (map[uint32][][]byte, error) {
			return eligibleMap, nil
		},
	}

	ptp := PeerTypeProvider{
		nodesCoordinator: arg.NodesCoordinator,
		cache:            nil,
		mutCache:         sync.RWMutex{},
	}

	ptp.updateCache(0)

	assert.NotNil(t, ptp.cache)
	assert.Equal(t, len(eligibleMap[initialShardId]), len(ptp.cache))
	assert.NotNil(t, ptp.cache[pk])
	assert.Equal(t, common.EligibleList, ptp.cache[pk].pType)
	assert.Equal(t, initialShardId, ptp.cache[pk].pShard)
}

func TestNewPeerTypeProvider_createCache(t *testing.T) {
	pkEligible := "pk1"
	pkWaiting := "pk2"

	eligibleMap := make(map[uint32][][]byte)
	waitingMap := make(map[uint32][][]byte)
	eligibleShardId := uint32(0)
	waitingShardId := uint32(1)
	eligibleMap[eligibleShardId] = [][]byte{
		[]byte(pkEligible),
	}
	waitingMap[waitingShardId] = [][]byte{
		[]byte(pkWaiting),
	}

	arg := createDefaultArgPeerTypeProvider()
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(_ uint32) (map[uint32][][]byte, error) {
			return eligibleMap, nil
		},
		GetAllWaitingValidatorsPublicKeysCalled: func() (map[uint32][][]byte, error) {
			return waitingMap, nil
		},
	}

	ptp := PeerTypeProvider{
		nodesCoordinator: arg.NodesCoordinator,
		cache:            nil,
		mutCache:         sync.RWMutex{},
	}

	cache := ptp.createNewCache(0)

	assert.NotNil(t, cache)

	assert.NotNil(t, cache[pkEligible])
	assert.Equal(t, common.EligibleList, cache[pkEligible].pType)
	assert.Equal(t, eligibleShardId, cache[pkEligible].pShard)

	assert.NotNil(t, cache[pkWaiting])
	assert.Equal(t, common.WaitingList, cache[pkWaiting].pType)
	assert.Equal(t, waitingShardId, cache[pkWaiting].pShard)
}

func TestNewPeerTypeProvider_CallsUpdateCacheOnEpochChange(t *testing.T) {
	arg := createDefaultArgPeerTypeProvider()
	callNumber := 0
	epochStartNotifier := &mock.EpochStartNotifierStub{}
	arg.EpochStartEventNotifier = epochStartNotifier
	pkEligibleInTrie := "pk1"
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(_ uint32) (map[uint32][][]byte, error) {
			callNumber++
			// first call comes from the constructor
			if callNumber == 1 {
				return nil, nil
			}
			return map[uint32][][]byte{
				0: {
					[]byte(pkEligibleInTrie),
				},
			}, nil
		},
	}

	ptp, _ := NewPeerTypeProvider(arg)

	assert.Equal(t, 0, len(ptp.GetCache())) // nothing in cache
	epochStartNotifier.NotifyAll(&block.Header{Nonce: 1, ShardID: 2, Round: 3})
	assert.Equal(t, 1, len(ptp.GetCache()))
	assert.NotNil(t, ptp.GetCache()[pkEligibleInTrie])
}

func TestNewPeerTypeProvider_ComputeForKeyFromCache(t *testing.T) {
	arg := createDefaultArgPeerTypeProvider()
	pk := []byte("pk1")
	initialShardId := uint32(1)
	popMutex := sync.RWMutex{}
	populateCacheCalled := false
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(_ uint32) (map[uint32][][]byte, error) {
			populateCacheCalled = true
			return map[uint32][][]byte{
				initialShardId: {pk},
			}, nil
		},
	}

	ptp, _ := NewPeerTypeProvider(arg)
	popMutex.Lock()
	populateCacheCalled = false
	popMutex.Unlock()
	peerType, shardId, err := ptp.ComputeForPubKey(pk)

	popMutex.RLock()
	called := populateCacheCalled
	popMutex.RUnlock()
	assert.False(t, called)
	assert.Equal(t, common.EligibleList, peerType)
	assert.Equal(t, initialShardId, shardId)
	assert.Nil(t, err)
}

func TestNewPeerTypeProvider_ComputeForKeyNotFoundInCacheReturnsObserver(t *testing.T) {
	arg := createDefaultArgPeerTypeProvider()
	pk := []byte("pk1")
	arg.NodesCoordinator = &shardingMocks.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(_ uint32) (map[uint32][][]byte, error) {
			return map[uint32][][]byte{}, nil
		},
	}

	ptp, _ := NewPeerTypeProvider(arg)

	peerType, shardId, err := ptp.ComputeForPubKey(pk)

	assert.Equal(t, common.ObserverList, peerType)
	assert.Equal(t, uint32(0), shardId)
	assert.Nil(t, err)
}

func TestNewPeerTypeProvider_IsInterfaceNil(t *testing.T) {
	arg := createDefaultArgPeerTypeProvider()

	ptp, _ := NewPeerTypeProvider(arg)
	assert.False(t, ptp.IsInterfaceNil())
}

func createDefaultArgPeerTypeProvider() ArgPeerTypeProvider {
	return ArgPeerTypeProvider{
		NodesCoordinator:        &shardingMocks.NodesCoordinatorMock{},
		StartEpoch:              0,
		EpochStartEventNotifier: &mock.EpochStartNotifierStub{},
	}
}
