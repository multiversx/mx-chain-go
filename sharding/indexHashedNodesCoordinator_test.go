package sharding

import (
	"encoding/hex"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	shuffleBetweenShards = false
	adaptivity           = false
	hysteresis           = float32(0.2)
	eligiblePerShard     = 100
	waitingPerShard      = 30
)

func createDummyNodesList(nbNodes uint32, suffix string) []validator {
	list := make([]validator, 0)
	hasher := sha256.NewSha256()

	for j := uint32(0); j < nbNodes; j++ {
		pk := hasher.Compute(fmt.Sprintf("pk%s_%d", suffix, j))
		list = append(list, mock.NewValidatorMock(pk, 1, nodesCoordinator.DefaultSelectionChances))
	}

	return list
}

func createDummyNodesMap(nodesPerShard uint32, nbShards uint32, suffix string) map[uint32][]validator {
	nodesMap := make(map[uint32][]validator)

	var shard uint32

	for i := uint32(0); i <= nbShards; i++ {
		shard = i
		if i == nbShards {
			shard = core.MetachainShardId
		}
		list := createDummyNodesList(nodesPerShard, suffix+"_i")
		nodesMap[shard] = list
	}

	return nodesMap
}

func isStringSubgroup(a []string, b []string) bool {
	var found bool
	for _, va := range a {
		found = false
		for _, vb := range b {
			if va == vb {
				found = true
				break
			}
		}
		if !found {
			return found
		}
	}

	return found
}

func createArguments() ArgNodesCoordinator {
	nbShards := uint32(1)
	eligibleMap := createDummyNodesMap(10, nbShards, "eligible")
	waitingMap := createDummyNodesMap(3, nbShards, "waiting")
	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           10,
		NodesMeta:            10,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, _ := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    1,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		NbShards:                   nbShards,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("test"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		IsFullArchive:              false,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	return arguments
}

func validatorsPubKeys(validators []validator) []string {
	pKeys := make([]string, len(validators))
	for _, v := range validators {
		pKeys = append(pKeys, string(v.PubKey()))
	}

	return pKeys
}

//------- NewIndexHashedNodesCoordinator

func TestNewIndexHashedNodesCoordinator_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.Hasher = nil
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	require.Equal(t, ErrNilHasher, err)
	require.Nil(t, ihgs)
}

func TestNewIndexHashedNodesCoordinator_InvalidConsensusGroupSizeShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.ShardConsensusGroupSize = 0
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	require.Equal(t, ErrInvalidConsensusGroupSize, err)
	require.Nil(t, ihgs)
}

func TestNewIndexHashedNodesCoordinator_ZeroNbShardsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.NbShards = 0
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	require.Equal(t, ErrInvalidNumberOfShards, err)
	require.Nil(t, ihgs)
}

func TestNewIndexHashedNodesCoordinator_InvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.ShardIDAsObserver = 10
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	require.Equal(t, ErrInvalidShardId, err)
	require.Nil(t, ihgs)
}

func TestNewIndexHashedNodesCoordinator_NilSelfPublicKeyShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.SelfPublicKey = nil
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	require.Equal(t, ErrNilPubKey, err)
	require.Nil(t, ihgs)
}

func TestNewIndexHashedNodesCoordinator_NilCacherShouldErr(t *testing.T) {
	arguments := createArguments()
	arguments.ConsensusGroupCache = nil
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	require.Equal(t, ErrNilCacher, err)
	require.Nil(t, ihgs)
}

func TestNewIndexHashedGroupSelector_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	require.NotNil(t, ihgs)
	require.Nil(t, err)
}

//------- LoadEligibleList

func TestIndexHashedNodesCoordinator_SetNilEligibleMapShouldErr(t *testing.T) {
	t.Parallel()

	waitingMap := createDummyNodesMap(3, 3, "waiting")
	arguments := createArguments()

	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
	require.Equal(t, ErrNilInputNodesMap, ihgs.SetNodesPerShards(nil, waitingMap, nil, 0))
}

func TestIndexHashedNodesCoordinator_SetNilWaitingMapShouldErr(t *testing.T) {
	t.Parallel()

	eligibleMap := createDummyNodesMap(10, 3, "eligible")
	arguments := createArguments()

	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
	require.Equal(t, ErrNilInputNodesMap, ihgs.SetNodesPerShards(eligibleMap, nil, nil, 0))
}

func TestIndexHashedNodesCoordinator_OkValShouldWork(t *testing.T) {
	t.Parallel()

	eligibleMap := createDummyNodesMap(10, 3, "eligible")
	waitingMap := createDummyNodesMap(3, 3, "waiting")

	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           10,
		NodesMeta:            10,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    2,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		NbShards:                   1,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	nodeConfig, _ := ihgs.GetNodesConfigPerEpoch(arguments.Epoch)
	readEligible := nodeConfig.EligibleMap[0]
	require.Equal(t, eligibleMap[0], readEligible)
}

//------- ComputeValidatorsGroup

func TestIndexHashedNodesCoordinator_NewCoordinatorGroup0SizeShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.MetaConsensusGroupSize = 0
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	require.Equal(t, ErrInvalidConsensusGroupSize, err)
	require.Nil(t, ihgs)
}

func TestIndexHashedNodesCoordinator_NewCoordinatorTooFewNodesShouldErr(t *testing.T) {
	t.Parallel()

	eligibleMap := createDummyNodesMap(5, 3, "eligible")
	waitingMap := createDummyNodesMap(3, 3, "waiting")
	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           10,
		NodesMeta:            10,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    10,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		NbShards:                   1,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	require.Equal(t, ErrSmallShardEligibleListSize, err)
	require.Nil(t, ihgs)
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroupNilRandomnessShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
	list2, err := ihgs.ComputeConsensusGroup(nil, 0, 0, 0)

	require.Equal(t, ErrNilRandomness, err)
	require.Nil(t, list2)
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroupInvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
	list2, err := ihgs.ComputeConsensusGroup([]byte("radomness"), 0, 5, 0)

	require.Equal(t, ErrInvalidShardId, err)
	require.Nil(t, list2)
}

//------- functionality tests

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroup1ValidatorShouldReturnSame(t *testing.T) {
	t.Parallel()

	list := []validator{
		mock.NewValidatorMock([]byte("pk0"), 1, nodesCoordinator.DefaultSelectionChances),
	}
	tmp := createDummyNodesMap(2, 1, "meta")
	nodesMap := make(map[uint32][]validator)
	nodesMap[0] = list
	nodesMap[core.MetachainShardId] = tmp[core.MetachainShardId]
	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           10,
		NodesMeta:            10,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    1,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		NbShards:                   1,
		EligibleNodes:              nodesMap,
		WaitingNodes:               make(map[uint32][]validator),
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
	list2, err := ihgs.ComputeConsensusGroup([]byte("randomness"), 0, 0, 0)

	require.Equal(t, list, list2)
	require.Nil(t, err)
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroup400of400For10locksNoMemoization(t *testing.T) {
	consensusGroupSize := 400
	nodesPerShard := uint32(400)
	waitingMap := make(map[uint32][]validator)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")
	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           nodesPerShard,
		NodesMeta:            nodesPerShard,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	getCounter := int32(0)
	putCounter := int32(0)

	cache := &mock.NodesCoordinatorCacheMock{
		PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
			atomic.AddInt32(&putCounter, 1)
			return false
		},
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			atomic.AddInt32(&getCounter, 1)
			return nil, false
		},
	}

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    consensusGroupSize,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		NbShards:                   1,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        cache,
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	miniBlocks := 10

	var list2 []validator
	for i := 0; i < miniBlocks; i++ {
		for j := 0; j <= i; j++ {
			randomness := strconv.Itoa(j)
			list2, err = ihgs.ComputeConsensusGroup([]byte(randomness), uint64(j), 0, 0)
			require.Nil(t, err)
			require.Equal(t, consensusGroupSize, len(list2))
		}
	}

	computationNr := miniBlocks * (miniBlocks + 1) / 2

	require.Equal(t, int32(computationNr), getCounter)
	require.Equal(t, int32(computationNr), putCounter)
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroup400of400For10BlocksMemoization(t *testing.T) {
	consensusGroupSize := 400
	nodesPerShard := uint32(400)
	waitingMap := make(map[uint32][]validator)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")
	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           nodesPerShard,
		NodesMeta:            nodesPerShard,
		Hysteresis:           0,
		Adaptivity:           false,
		ShuffleBetweenShards: false,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	getCounter := 0
	putCounter := 0

	mut := sync.Mutex{}

	//consensusGroup := list[0:21]
	cacheMap := make(map[string]interface{})
	cache := &mock.NodesCoordinatorCacheMock{
		PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
			mut.Lock()
			defer mut.Unlock()
			putCounter++
			cacheMap[string(key)] = value
			return false
		},
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			mut.Lock()
			defer mut.Unlock()
			getCounter++
			val, ok := cacheMap[string(key)]
			if ok {
				return val, true
			}
			return nil, false
		},
	}

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    consensusGroupSize,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		NbShards:                   1,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        cache,
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	miniBlocks := 10

	var list2 []validator
	for i := 0; i < miniBlocks; i++ {
		for j := 0; j <= i; j++ {
			randomness := strconv.Itoa(j)
			list2, err = ihgs.ComputeConsensusGroup([]byte(randomness), uint64(j), 0, 0)
			require.Nil(t, err)
			require.Equal(t, consensusGroupSize, len(list2))
		}
	}

	computationNr := miniBlocks * (miniBlocks + 1) / 2

	require.Equal(t, computationNr, getCounter)
	require.Equal(t, miniBlocks, putCounter)
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroup63of400TestEqualSameParams(t *testing.T) {
	t.Skip("testing consistency - to be run manually")
	cache := &mock.NodesCoordinatorCacheMock{
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			return nil, false
		},
		PutCalled: func(key []byte, value interface{}, sizeInBytes int) (evicted bool) {
			return false
		},
	}

	consensusGroupSize := 63
	nodesPerShard := uint32(400)
	waitingMap := make(map[uint32][]validator)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")

	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           nodesPerShard,
		NodesMeta:            nodesPerShard,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    consensusGroupSize,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		NbShards:                   1,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        cache,
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	nbDifferentSamplings := 1000
	repeatPerSampling := 100

	list := make([][]validator, repeatPerSampling)
	for i := 0; i < nbDifferentSamplings; i++ {
		randomness := arguments.Hasher.Compute(strconv.Itoa(i))
		fmt.Printf("starting selection with randomness: %s\n", hex.EncodeToString(randomness))
		for j := 0; j < repeatPerSampling; j++ {
			list[j], err = ihgs.ComputeConsensusGroup(randomness, 0, 0, 0)
			require.Nil(t, err)
			require.Equal(t, consensusGroupSize, len(list[j]))
		}

		for j := 1; j < repeatPerSampling; j++ {
			require.Equal(t, list[0], list[j])
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func BenchmarkIndexHashedGroupSelector_ComputeValidatorsGroup21of400(b *testing.B) {
	consensusGroupSize := 21
	nodesPerShard := uint32(400)
	waitingMap := make(map[uint32][]validator)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")
	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           nodesPerShard,
		NodesMeta:            nodesPerShard,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(b, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    consensusGroupSize,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		NbShards:                   1,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomness := strconv.Itoa(i)
		list2, _ := ihgs.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

		require.Equal(b, consensusGroupSize, len(list2))
	}
}

func runBenchmark(consensusGroupCache nodesCoordinator.Cacher, consensusGroupSize int, nodesMap map[uint32][]validator, b *testing.B) {
	waitingMap := make(map[uint32][]validator)
	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           10,
		NodesMeta:            10,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(b, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    consensusGroupSize,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		EpochStartNotifier:         epochStartSubscriber,
		Shuffler:                   nodeShuffler,
		BootStorer:                 bootStorer,
		NbShards:                   1,
		EligibleNodes:              nodesMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        consensusGroupCache,
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		missedBlocks := 1000
		for j := 0; j < missedBlocks; j++ {
			randomness := strconv.Itoa(j)
			list2, _ := ihgs.ComputeConsensusGroup([]byte(randomness), uint64(j), 0, 0)
			require.Equal(b, consensusGroupSize, len(list2))
		}
	}
}

func computeMemoryRequirements(consensusGroupCache nodesCoordinator.Cacher, consensusGroupSize int, nodesMap map[uint32][]validator, b *testing.B) {
	waitingMap := make(map[uint32][]validator)
	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           10,
		NodesMeta:            10,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(b, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    consensusGroupSize,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		EpochStartNotifier:         epochStartSubscriber,
		Shuffler:                   nodeShuffler,
		BootStorer:                 bootStorer,
		NbShards:                   1,
		EligibleNodes:              nodesMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        consensusGroupCache,
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(b, err)

	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)

	missedBlocks := 1000
	for i := 0; i < missedBlocks; i++ {
		randomness := strconv.Itoa(i)
		list2, _ := ihgs.ComputeConsensusGroup([]byte(randomness), uint64(i), 0, 0)
		require.Equal(b, consensusGroupSize, len(list2))
	}

	m2 := runtime.MemStats{}
	runtime.ReadMemStats(&m2)

	fmt.Printf("Used %d MB\n", (m2.HeapAlloc-m.HeapAlloc)/1024/1024)
}

func BenchmarkIndexHashedNodesCoordinator_ComputeValidatorsGroup63of400RecomputeEveryGroup(b *testing.B) {
	consensusGroupSize := 63
	nodesPerShard := uint32(400)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")

	consensusGroupCache, _ := lrucache.NewCache(1)
	computeMemoryRequirements(consensusGroupCache, consensusGroupSize, eligibleMap, b)
	consensusGroupCache, _ = lrucache.NewCache(1)
	runBenchmark(consensusGroupCache, consensusGroupSize, eligibleMap, b)
}

func BenchmarkIndexHashedNodesCoordinator_ComputeValidatorsGroup400of400RecomputeEveryGroup(b *testing.B) {
	consensusGroupSize := 400
	nodesPerShard := uint32(400)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")

	consensusGroupCache, _ := lrucache.NewCache(1)
	computeMemoryRequirements(consensusGroupCache, consensusGroupSize, eligibleMap, b)
	consensusGroupCache, _ = lrucache.NewCache(1)
	runBenchmark(consensusGroupCache, consensusGroupSize, eligibleMap, b)
}

func BenchmarkIndexHashedNodesCoordinator_ComputeValidatorsGroup63of400Memoization(b *testing.B) {
	consensusGroupSize := 63
	nodesPerShard := uint32(400)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")

	consensusGroupCache, _ := lrucache.NewCache(10000)
	computeMemoryRequirements(consensusGroupCache, consensusGroupSize, eligibleMap, b)
	consensusGroupCache, _ = lrucache.NewCache(10000)
	runBenchmark(consensusGroupCache, consensusGroupSize, eligibleMap, b)
}

func BenchmarkIndexHashedNodesCoordinator_ComputeValidatorsGroup400of400Memoization(b *testing.B) {
	consensusGroupSize := 400
	nodesPerShard := uint32(400)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")

	consensusGroupCache, _ := lrucache.NewCache(1000)
	computeMemoryRequirements(consensusGroupCache, consensusGroupSize, eligibleMap, b)
	consensusGroupCache, _ = lrucache.NewCache(1000)
	runBenchmark(consensusGroupCache, consensusGroupSize, eligibleMap, b)
}

func TestIndexHashedNodesCoordinator_GetValidatorWithPublicKeyShouldReturnErrNilPubKey(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	_, _, err := ihgs.GetValidatorWithPublicKey(nil)
	require.Equal(t, ErrNilPubKey, err)
}

func TestIndexHashedNodesCoordinator_GetValidatorWithPublicKeyShouldReturnErrValidatorNotFound(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	_, _, err := ihgs.GetValidatorWithPublicKey([]byte("pk1"))
	require.Equal(t, ErrValidatorNotFound, err)
}

func TestIndexHashedNodesCoordinator_GetValidatorWithPublicKeyShouldWork(t *testing.T) {
	t.Parallel()

	listMeta := []validator{
		mock.NewValidatorMock([]byte("pk0_meta"), 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock([]byte("pk1_meta"), 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock([]byte("pk2_meta"), 1, nodesCoordinator.DefaultSelectionChances),
	}
	listShard0 := []validator{
		mock.NewValidatorMock([]byte("pk0_shard0"), 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock([]byte("pk1_shard0"), 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock([]byte("pk2_shard0"), 1, nodesCoordinator.DefaultSelectionChances),
	}
	listShard1 := []validator{
		mock.NewValidatorMock([]byte("pk0_shard1"), 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock([]byte("pk1_shard1"), 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock([]byte("pk2_shard1"), 1, nodesCoordinator.DefaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]validator)
	eligibleMap[core.MetachainShardId] = listMeta
	eligibleMap[0] = listShard0
	eligibleMap[1] = listShard1
	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           10,
		NodesMeta:            10,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    1,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		NbShards:                   2,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               make(map[uint32][]validator),
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	v, shardId, err := ihgs.GetValidatorWithPublicKey([]byte("pk0_meta"))
	require.Nil(t, err)
	require.Equal(t, core.MetachainShardId, shardId)
	require.Equal(t, []byte("pk0_meta"), v.PubKey())

	v, shardId, err = ihgs.GetValidatorWithPublicKey([]byte("pk1_shard0"))
	require.Nil(t, err)
	require.Equal(t, uint32(0), shardId)
	require.Equal(t, []byte("pk1_shard0"), v.PubKey())

	v, shardId, err = ihgs.GetValidatorWithPublicKey([]byte("pk2_shard1"))
	require.Nil(t, err)
	require.Equal(t, uint32(1), shardId)
	require.Equal(t, []byte("pk2_shard1"), v.PubKey())
}

func TestIndexHashedGroupSelector_GetAllEligibleValidatorsPublicKeys(t *testing.T) {
	t.Parallel()

	shardZeroId := uint32(0)
	shardOneId := uint32(1)
	expectedValidatorsPubKeys := map[uint32][][]byte{
		shardZeroId:           {[]byte("pk0_shard0"), []byte("pk1_shard0"), []byte("pk2_shard0")},
		shardOneId:            {[]byte("pk0_shard1"), []byte("pk1_shard1"), []byte("pk2_shard1")},
		core.MetachainShardId: {[]byte("pk0_meta"), []byte("pk1_meta"), []byte("pk2_meta")},
	}

	listMeta := []validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][0], 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][1], 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][2], 1, nodesCoordinator.DefaultSelectionChances),
	}
	listShard0 := []validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][0], 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][1], 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][2], 1, nodesCoordinator.DefaultSelectionChances),
	}
	listShard1 := []validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][0], 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][1], 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][2], 1, nodesCoordinator.DefaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]validator)
	eligibleMap[core.MetachainShardId] = listMeta
	eligibleMap[shardZeroId] = listShard0
	eligibleMap[shardOneId] = listShard1
	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           10,
		NodesMeta:            10,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    1,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		ShardIDAsObserver:          shardZeroId,
		NbShards:                   2,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               make(map[uint32][]validator),
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	allValidatorsPublicKeys, err := ihgs.GetAllEligibleValidatorsPublicKeys(0)
	require.Equal(t, expectedValidatorsPubKeys, allValidatorsPublicKeys)
	require.Nil(t, err)
}

func TestIndexHashedGroupSelector_GetAllWaitingValidatorsPublicKeys(t *testing.T) {
	t.Parallel()

	shardZeroId := uint32(0)
	shardOneId := uint32(1)
	expectedValidatorsPubKeys := map[uint32][][]byte{
		shardZeroId:           {[]byte("pk0_shard0"), []byte("pk1_shard0"), []byte("pk2_shard0")},
		shardOneId:            {[]byte("pk0_shard1"), []byte("pk1_shard1"), []byte("pk2_shard1")},
		core.MetachainShardId: {[]byte("pk0_meta"), []byte("pk1_meta"), []byte("pk2_meta")},
	}

	listMeta := []validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][0], 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][1], 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][2], 1, nodesCoordinator.DefaultSelectionChances),
	}
	listShard0 := []validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][0], 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][1], 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][2], 1, nodesCoordinator.DefaultSelectionChances),
	}
	listShard1 := []validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][0], 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][1], 1, nodesCoordinator.DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][2], 1, nodesCoordinator.DefaultSelectionChances),
	}

	waitingMap := make(map[uint32][]validator)
	waitingMap[core.MetachainShardId] = listMeta
	waitingMap[shardZeroId] = listShard0
	waitingMap[shardOneId] = listShard1

	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           10,
		NodesMeta:            10,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	eligibleMap := make(map[uint32][]validator)
	eligibleMap[core.MetachainShardId] = []validator{&mock.ValidatorMock{}}
	eligibleMap[shardZeroId] = []validator{&mock.ValidatorMock{}}

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    1,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		ShardIDAsObserver:          shardZeroId,
		NbShards:                   2,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	allValidatorsPublicKeys, err := ihgs.GetAllWaitingValidatorsPublicKeys(0)
	require.Equal(t, expectedValidatorsPubKeys, allValidatorsPublicKeys)
	require.Nil(t, err)
}

func createBlockBodyFromNodesCoordinator(ihgs *indexHashedNodesCoordinator, epoch uint32) *block.Body {
	body := &block.Body{MiniBlocks: make([]*block.MiniBlock, 0)}

	nodeConfig, _ := ihgs.GetNodesConfigPerEpoch(epoch)

	mbs := createMiniBlocksForNodesMap(nodeConfig.EligibleMap, string(common.EligibleList), ihgs.marshalizer)
	body.MiniBlocks = append(body.MiniBlocks, mbs...)

	mbs = createMiniBlocksForNodesMap(nodeConfig.WaitingMap, string(common.WaitingList), ihgs.marshalizer)
	body.MiniBlocks = append(body.MiniBlocks, mbs...)

	mbs = createMiniBlocksForNodesMap(nodeConfig.LeavingMap, string(common.LeavingList), ihgs.marshalizer)
	body.MiniBlocks = append(body.MiniBlocks, mbs...)

	return body
}

func createMiniBlocksForNodesMap(nodesMap map[uint32][]validator, list string, marshalizer marshal.Marshalizer) []*block.MiniBlock {
	miniBlocks := make([]*block.MiniBlock, 0)
	for shId, eligibleList := range nodesMap {
		miniBlock := &block.MiniBlock{Type: block.PeerBlock}
		for index, eligible := range eligibleList {
			shardVInfo := &state.ShardValidatorInfo{
				PublicKey:  eligible.PubKey(),
				ShardId:    shId,
				List:       list,
				Index:      uint32(index),
				TempRating: 10,
			}

			marshaledData, _ := marshalizer.Marshal(shardVInfo)
			miniBlock.TxHashes = append(miniBlock.TxHashes, marshaledData)
		}
		miniBlocks = append(miniBlocks, miniBlock)
	}
	return miniBlocks
}

func TestIndexHashedNodesCoordinator_EpochStart(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)
	epoch := uint32(1)

	header := &block.MetaBlock{
		PrevRandSeed: []byte("rand seed"),
		EpochStart:   block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		Epoch:        epoch,
	}

	nodeConfig, _ := ihgs.GetNodesConfigPerEpoch(0)

	ihgs.SetNodesConfigPerEpoch(epoch, nodeConfig)

	body := createBlockBodyFromNodesCoordinator(ihgs, epoch)
	ihgs.EpochStartPrepare(header, body)
	ihgs.EpochStartAction(header)

	validators, err := ihgs.GetAllEligibleValidatorsPublicKeys(epoch)
	require.Nil(t, err)
	require.NotNil(t, validators)

	computedShardId, isValidator := ihgs.computeShardForSelfPublicKey(nodeConfig)
	// should remain in same shard with intra shard shuffling
	require.Equal(t, arguments.ShardIDAsObserver, computedShardId)
	require.False(t, isValidator)
}

func TestIndexHashedNodesCoordinator_setNodesPerShardsShouldTriggerWrongConfiguration(t *testing.T) {
	t.Parallel()

	chanStopNode := make(chan endProcess.ArgEndProcess, 1)
	arguments := createArguments()
	arguments.ChanStopNode = chanStopNode
	arguments.IsFullArchive = true

	pk := []byte("pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	eligibleMap := map[uint32][]validator{
		core.MetachainShardId: {
			mock.NewValidatorMock(pk, 1, 1),
		},
	}

	err = ihgs.SetNodesPerShards(eligibleMap, map[uint32][]validator{}, map[uint32][]validator{}, 2)
	require.NoError(t, err)

	value := <-chanStopNode
	require.Equal(t, common.WrongConfiguration, value.Reason)
}

func TestIndexHashedNodesCoordinator_setNodesPerShardsShouldNotTriggerWrongConfiguration(t *testing.T) {
	t.Parallel()

	chanStopNode := make(chan endProcess.ArgEndProcess, 1)
	arguments := createArguments()
	arguments.ChanStopNode = chanStopNode
	arguments.IsFullArchive = false

	pk := []byte("pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	eligibleMap := map[uint32][]validator{
		core.MetachainShardId: {
			mock.NewValidatorMock(pk, 1, 1),
		},
	}

	err = ihgs.SetNodesPerShards(eligibleMap, map[uint32][]validator{}, map[uint32][]validator{}, 2)
	require.NoError(t, err)

	require.Empty(t, chanStopNode)
}

func TestIndexHashedNodesCoordinator_setNodesPerShardsShouldSetNodeTypeValidator(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.IsFullArchive = false

	var nodeTypeResult core.NodeType
	var setTypeWasCalled bool
	arguments.NodeTypeProvider = &nodeTypeProviderMock.NodeTypeProviderStub{
		SetTypeCalled: func(nodeType core.NodeType) {
			nodeTypeResult = nodeType
			setTypeWasCalled = true
		},
	}

	pk := []byte("pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	eligibleMap := map[uint32][]validator{
		core.MetachainShardId: {
			mock.NewValidatorMock(pk, 1, 1),
		},
	}

	err = ihgs.SetNodesPerShards(eligibleMap, map[uint32][]validator{}, map[uint32][]validator{}, 2)
	require.NoError(t, err)
	require.True(t, setTypeWasCalled)
	require.Equal(t, core.NodeTypeValidator, nodeTypeResult)
}

func TestIndexHashedNodesCoordinator_setNodesPerShardsShouldSetNodeTypeObserver(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.IsFullArchive = false

	var nodeTypeResult core.NodeType
	var setTypeWasCalled bool
	arguments.NodeTypeProvider = &nodeTypeProviderMock.NodeTypeProviderStub{
		SetTypeCalled: func(nodeType core.NodeType) {
			nodeTypeResult = nodeType
			setTypeWasCalled = true
		},
	}

	pk := []byte("observer pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	eligibleMap := map[uint32][]validator{
		core.MetachainShardId: {
			mock.NewValidatorMock([]byte("validator pk"), 1, 1),
		},
	}

	err = ihgs.SetNodesPerShards(eligibleMap, map[uint32][]validator{}, map[uint32][]validator{}, 2)
	require.NoError(t, err)
	require.True(t, setTypeWasCalled)
	require.Equal(t, core.NodeTypeObserver, nodeTypeResult)
}

func TestIndexHashedNodesCoordinator_EpochStartInEligible(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	pk := []byte("pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)
	epoch := uint32(2)

	header := &block.MetaBlock{
		PrevRandSeed: []byte("rand seed"),
		EpochStart:   block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		Epoch:        epoch,
	}

	validatorShard := core.MetachainShardId
	ihgs.SetNodesConfig(map[uint32]*nodesCoordinator.EpochNodesConfig{
		epoch: {
			ShardID: validatorShard,
			EligibleMap: map[uint32][]validator{
				validatorShard: {mock.NewValidatorMock(pk, 1, 1)},
			},
		},
	})
	body := createBlockBodyFromNodesCoordinator(ihgs, epoch)
	ihgs.EpochStartPrepare(header, body)
	ihgs.EpochStartAction(header)

	nodeConfig, _ := ihgs.GetNodesConfigPerEpoch(epoch)
	computedShardId, isValidator := ihgs.computeShardForSelfPublicKey(nodeConfig)

	require.Equal(t, validatorShard, computedShardId)
	require.True(t, isValidator)
}

func TestIndexHashedNodesCoordinator_EpochStartInWaiting(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	pk := []byte("pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	epoch := uint32(2)
	header := &block.MetaBlock{
		PrevRandSeed: []byte("rand seed"),
		EpochStart:   block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		Epoch:        epoch,
	}

	validatorShard := core.MetachainShardId
	ihgs.SetNodesConfig(map[uint32]*nodesCoordinator.EpochNodesConfig{
		epoch: {
			ShardID: validatorShard,
			WaitingMap: map[uint32][]validator{
				validatorShard: {mock.NewValidatorMock(pk, 1, 1)},
			},
		},
	})
	body := createBlockBodyFromNodesCoordinator(ihgs, epoch)
	ihgs.EpochStartPrepare(header, body)
	ihgs.EpochStartAction(header)

	nodeConfig, _ := ihgs.GetNodesConfigPerEpoch(epoch)
	computedShardId, isValidator := ihgs.computeShardForSelfPublicKey(nodeConfig)
	require.Equal(t, validatorShard, computedShardId)
	require.True(t, isValidator)
}

func TestIndexHashedNodesCoordinator_EpochStartInLeaving(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	pk := []byte("pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	epoch := uint32(2)
	header := &block.MetaBlock{
		PrevRandSeed: []byte("rand seed"),
		EpochStart:   block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		Epoch:        epoch,
	}

	validatorShard := core.MetachainShardId
	ihgs.SetNodesConfig(map[uint32]*nodesCoordinator.EpochNodesConfig{
		epoch: {
			ShardID: validatorShard,
			EligibleMap: map[uint32][]validator{
				validatorShard: {
					mock.NewValidatorMock([]byte("eligiblePk"), 1, 1),
				},
			},
			LeavingMap: map[uint32][]validator{
				validatorShard: {mock.NewValidatorMock(pk, 1, 1)},
			},
		},
	})
	body := createBlockBodyFromNodesCoordinator(ihgs, epoch)
	ihgs.EpochStartPrepare(header, body)
	ihgs.EpochStartAction(header)

	nodeConfig, _ := ihgs.GetNodesConfigPerEpoch(epoch)
	computedShardId, isValidator := ihgs.computeShardForSelfPublicKey(nodeConfig)
	require.Equal(t, validatorShard, computedShardId)
	require.True(t, isValidator)
}

func TestIndexHashedNodesCoordinator_EpochStart_EligibleSortedAscendingByIndex(t *testing.T) {
	t.Parallel()

	nbShards := uint32(1)
	eligibleMap := make(map[uint32][]validator)

	pk1 := []byte{2}
	pk2 := []byte{1}

	list := []validator{
		mock.NewValidatorMock(pk1, 1, 1),
		mock.NewValidatorMock(pk2, 1, 1),
	}
	eligibleMap[core.MetachainShardId] = list

	shufflerArgs := &nodesCoordinator.NodesShufflerArgs{
		NodesShard:           2,
		NodesMeta:            2,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	nodeShuffler, err := nodesCoordinator.NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize:    1,
		MetaConsensusGroupSize:     1,
		Marshalizer:                &mock.MarshalizerMock{},
		Hasher:                     &mock.HasherMock{},
		Shuffler:                   nodeShuffler,
		EpochStartNotifier:         epochStartSubscriber,
		BootStorer:                 bootStorer,
		NbShards:                   nbShards,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               map[uint32][]validator{},
		SelfPublicKey:              []byte("test"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:         &mock.ShuffledOutHandlerStub{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)
	epoch := uint32(1)

	header := &block.MetaBlock{
		PrevRandSeed: []byte("rand seed"),
		EpochStart:   block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		Epoch:        epoch,
	}

	nodeConfig, _ := ihgs.GetNodesConfigPerEpoch(0)

	ihgs.SetNodesConfigPerEpoch(epoch, nodeConfig)

	body := createBlockBodyFromNodesCoordinator(ihgs, epoch)
	ihgs.EpochStartPrepare(header, body)

	nodeConfig, _ = ihgs.GetNodesConfigPerEpoch(1)
	newNodesConfig := nodeConfig

	firstEligible := newNodesConfig.EligibleMap[core.MetachainShardId][0]
	secondEligible := newNodesConfig.EligibleMap[core.MetachainShardId][1]
	assert.True(t, firstEligible.Index() < secondEligible.Index())
}

func TestIndexHashedNodesCoordinator_GetConsensusValidatorsPublicKeysNotExistingEpoch(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	var pKeys []string
	randomness := []byte("randomness")
	pKeys, err = ihgs.GetConsensusValidatorsPublicKeys(randomness, 0, 0, 1)
	require.True(t, errors.Is(err, nodesCoordinator.ErrEpochNodesConfigDoesNotExist))
	require.Nil(t, pKeys)
}

func TestIndexHashedNodesCoordinator_GetConsensusValidatorsPublicKeysExistingEpoch(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	shard0PubKeys := validatorsPubKeys(args.EligibleNodes[0])

	var pKeys []string
	randomness := []byte("randomness")
	pKeys, err = ihgs.GetConsensusValidatorsPublicKeys(randomness, 0, 0, 0)
	require.Nil(t, err)
	require.True(t, len(pKeys) > 0)
	require.True(t, isStringSubgroup(pKeys, shard0PubKeys))
}

func TestIndexHashedNodesCoordinator_GetValidatorsIndexes(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)
	randomness := []byte("randomness")

	var pKeys []string
	pKeys, err = ihgs.GetConsensusValidatorsPublicKeys(randomness, 0, 0, 0)
	require.Nil(t, err)

	var indexes []uint64
	indexes, err = ihgs.GetValidatorsIndexes(pKeys, 0)
	require.Nil(t, err)
	require.Equal(t, len(pKeys), len(indexes))
}

func TestIndexHashedNodesCoordinator_GetValidatorsIndexesInvalidPubKey(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)
	randomness := []byte("randomness")

	var pKeys []string
	pKeys, err = ihgs.GetConsensusValidatorsPublicKeys(randomness, 0, 0, 0)
	require.Nil(t, err)

	var indexes []uint64
	pKeys[0] = "dummy"
	indexes, err = ihgs.GetValidatorsIndexes(pKeys, 0)
	require.Equal(t, ErrInvalidNumberPubKeys, err)
	require.Nil(t, indexes)
}

func TestIndexHashedNodesCoordinator_GetSavedStateKey(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	header := &block.MetaBlock{
		PrevRandSeed: []byte("rand seed"),
		EpochStart:   block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		Epoch:        1,
	}

	body := createBlockBodyFromNodesCoordinator(ihgs, 0)
	ihgs.EpochStartPrepare(header, body)
	ihgs.EpochStartAction(header)

	key := ihgs.GetSavedStateKey()
	require.Equal(t, []byte("rand seed"), key)
}

func TestIndexHashedNodesCoordinator_GetSavedStateKeyEpoch0(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	expectedKey := args.Hasher.Compute(string(args.SelfPublicKey))
	key := ihgs.GetSavedStateKey()
	require.Equal(t, expectedKey, key)
}

func TestIndexHashedNodesCoordinator_ShardIdForEpochInvalidEpoch(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	shardId, err := ihgs.ShardIdForEpoch(1)
	require.True(t, errors.Is(err, nodesCoordinator.ErrEpochNodesConfigDoesNotExist))
	require.Equal(t, uint32(0), shardId)
}

func TestIndexHashedNodesCoordinator_ShardIdForEpochValidEpoch(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	shardId, err := ihgs.ShardIdForEpoch(0)
	require.Nil(t, err)
	require.Equal(t, uint32(0), shardId)
}

func TestIndexHashedNodesCoordinator_GetConsensusWhitelistedNodesEpoch0(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	nodesCurrentEpoch, err := ihgs.GetAllEligibleValidatorsPublicKeys(0)
	require.Nil(t, err)

	allNodesList := make([]string, 0)
	for _, nodesList := range nodesCurrentEpoch {
		for _, nodeKey := range nodesList {
			allNodesList = append(allNodesList, string(nodeKey))
		}
	}

	whitelistedNodes, err := ihgs.GetConsensusWhitelistedNodes(0)
	require.Nil(t, err)
	require.Greater(t, len(whitelistedNodes), 0)

	for key := range whitelistedNodes {
		require.True(t, isStringSubgroup([]string{key}, allNodesList))
	}
}

func TestIndexHashedNodesCoordinator_GetConsensusWhitelistedNodesEpoch1(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	header := &block.MetaBlock{
		PrevRandSeed: []byte("rand seed"),
		EpochStart:   block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		Epoch:        1,
	}

	body := createBlockBodyFromNodesCoordinator(ihgs, 0)
	ihgs.EpochStartPrepare(header, body)
	ihgs.EpochStartAction(header)

	nodesPrevEpoch, err := ihgs.GetAllEligibleValidatorsPublicKeys(0)
	require.Nil(t, err)
	nodesCurrentEpoch, err := ihgs.GetAllEligibleValidatorsPublicKeys(1)
	require.Nil(t, err)

	allNodesList := make([]string, 0)
	for shardId := range nodesPrevEpoch {
		for _, nodeKey := range nodesPrevEpoch[shardId] {
			allNodesList = append(allNodesList, string(nodeKey))
		}
		for _, nodeKey := range nodesCurrentEpoch[shardId] {
			allNodesList = append(allNodesList, string(nodeKey))
		}
	}

	whitelistedNodes, err := ihgs.GetConsensusWhitelistedNodes(1)
	require.Nil(t, err)
	require.Greater(t, len(whitelistedNodes), 0)

	for key := range whitelistedNodes {
		require.True(t, isStringSubgroup([]string{key}, allNodesList))
	}
}

func TestIndexHashedNodesCoordinator_GetConsensusWhitelistedNodesAfterRevertToEpoch(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	header := &block.MetaBlock{
		PrevRandSeed: []byte("rand seed"),
		EpochStart:   block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		Epoch:        1,
	}

	body := createBlockBodyFromNodesCoordinator(ihgs, 0)
	ihgs.EpochStartPrepare(header, body)
	ihgs.EpochStartAction(header)

	body = createBlockBodyFromNodesCoordinator(ihgs, 1)
	header = &block.MetaBlock{
		PrevRandSeed: []byte("rand seed"),
		EpochStart:   block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		Epoch:        2,
	}
	ihgs.EpochStartPrepare(header, body)
	ihgs.EpochStartAction(header)

	body = createBlockBodyFromNodesCoordinator(ihgs, 2)
	header = &block.MetaBlock{
		PrevRandSeed: []byte("rand seed"),
		EpochStart:   block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		Epoch:        3,
	}
	ihgs.EpochStartPrepare(header, body)
	ihgs.EpochStartAction(header)

	body = createBlockBodyFromNodesCoordinator(ihgs, 3)
	header = &block.MetaBlock{
		PrevRandSeed: []byte("rand seed"),
		EpochStart:   block.EpochStart{LastFinalizedHeaders: []block.EpochStartShardData{{}}},
		Epoch:        4,
	}
	ihgs.EpochStartPrepare(header, body)
	ihgs.EpochStartAction(header)

	nodesEpoch1, err := ihgs.GetAllEligibleValidatorsPublicKeys(1)
	require.Nil(t, err)

	allNodesList := make([]string, 0)
	for _, nodesList := range nodesEpoch1 {
		for _, nodeKey := range nodesList {
			allNodesList = append(allNodesList, string(nodeKey))
		}
	}

	whitelistedNodes, err := ihgs.GetConsensusWhitelistedNodes(1)
	require.Nil(t, err)
	require.Greater(t, len(whitelistedNodes), 0)

	for key := range whitelistedNodes {
		require.True(t, isStringSubgroup([]string{key}, allNodesList))
	}
}

func TestIndexHashedNodesCoordinator_ConsensusGroupSize(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	consensusSizeShard := ihgs.ConsensusGroupSize(0)
	consensusSizeMeta := ihgs.ConsensusGroupSize(core.MetachainShardId)

	require.Equal(t, arguments.ShardConsensusGroupSize, consensusSizeShard)
	require.Equal(t, arguments.MetaConsensusGroupSize, consensusSizeMeta)
}

func TestIndexHashedNodesCoordinator_GetNumTotalEligible(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	expectedNbNodes := uint64(0)
	for _, nodesList := range arguments.EligibleNodes {
		expectedNbNodes += uint64(len(nodesList))
	}

	nbNodes := ihgs.GetNumTotalEligible()
	require.Equal(t, expectedNbNodes, nbNodes)
}

func TestIndexHashedNodesCoordinator_GetOwnPublicKey(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	ownPubKey := ihgs.GetOwnPublicKey()
	require.Equal(t, arguments.SelfPublicKey, ownPubKey)
}

func TestIndexHashedNodesCoordinator_ShuffleOutWithEligible(t *testing.T) {
	t.Parallel()

	processCalled := false
	newShard := uint32(0)

	arguments := createArguments()
	arguments.ShuffledOutHandler = &mock.ShuffledOutHandlerStub{
		ProcessCalled: func(newShardID uint32) error {
			processCalled = true
			newShard = newShardID
			return nil
		},
	}
	pk := []byte("pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	epoch := uint32(2)
	validatorShard := uint32(7)
	ihgs.SetNodesConfig(map[uint32]*nodesCoordinator.EpochNodesConfig{
		epoch: {
			ShardID: validatorShard,
			EligibleMap: map[uint32][]validator{
				validatorShard: {mock.NewValidatorMock(pk, 1, 1)},
			},
		},
	})

	ihgs.ShuffleOutForEpoch(epoch)
	require.True(t, processCalled)
	require.Equal(t, validatorShard, newShard)
}

func TestIndexHashedNodesCoordinator_ShuffleOutWithWaiting(t *testing.T) {
	t.Parallel()

	processCalled := false
	newShard := uint32(0)

	arguments := createArguments()
	arguments.ShuffledOutHandler = &mock.ShuffledOutHandlerStub{
		ProcessCalled: func(newShardID uint32) error {
			processCalled = true
			newShard = newShardID
			return nil
		},
	}
	pk := []byte("pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	epoch := uint32(2)
	validatorShard := uint32(7)
	ihgs.SetNodesConfig(map[uint32]*nodesCoordinator.EpochNodesConfig{
		epoch: {
			ShardID: validatorShard,
			WaitingMap: map[uint32][]validator{
				validatorShard: {mock.NewValidatorMock(pk, 1, 1)},
			},
		},
	})

	ihgs.ShuffleOutForEpoch(epoch)
	require.True(t, processCalled)
	require.Equal(t, validatorShard, newShard)
}

func TestIndexHashedNodesCoordinator_ShuffleOutWithObserver(t *testing.T) {
	t.Parallel()

	processCalled := false
	newShard := uint32(0)

	arguments := createArguments()
	arguments.ShuffledOutHandler = &mock.ShuffledOutHandlerStub{
		ProcessCalled: func(newShardID uint32) error {
			processCalled = true
			newShard = newShardID
			return nil
		},
	}
	pk := []byte("pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	epoch := uint32(2)
	validatorShard := uint32(7)
	ihgs.SetNodesConfig(map[uint32]*nodesCoordinator.EpochNodesConfig{
		epoch: {
			ShardID: validatorShard,
			EligibleMap: map[uint32][]validator{
				validatorShard: {mock.NewValidatorMock([]byte("eligibleKey"), 1, 1)},
			},
			WaitingMap: map[uint32][]validator{
				validatorShard: {mock.NewValidatorMock([]byte("waitingKey"), 1, 1)},
			},
			LeavingMap: map[uint32][]validator{
				validatorShard: {mock.NewValidatorMock(pk, 1, 1)}},
		},
	})

	ihgs.ShuffleOutForEpoch(epoch)
	require.False(t, processCalled)
	expectedShardForLeaving := uint32(0)
	require.Equal(t, expectedShardForLeaving, newShard)
}

func TestIndexHashedNodesCoordinator_ShuffleOutNotFound(t *testing.T) {
	t.Parallel()

	processCalled := false
	newShard := uint32(0)

	arguments := createArguments()
	arguments.ShuffledOutHandler = &mock.ShuffledOutHandlerStub{
		ProcessCalled: func(newShardID uint32) error {
			processCalled = true
			newShard = newShardID
			return nil
		},
	}
	pk := []byte("pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	epoch := uint32(2)
	validatorShard := uint32(7)
	ihgs.SetNodesConfig(map[uint32]*nodesCoordinator.EpochNodesConfig{
		epoch: {
			ShardID: validatorShard,
			EligibleMap: map[uint32][]validator{
				validatorShard: {mock.NewValidatorMock([]byte("eligibleKey"), 1, 1)},
			},
			WaitingMap: map[uint32][]validator{
				validatorShard: {mock.NewValidatorMock([]byte("waitingKey"), 1, 1)},
			},
			LeavingMap: map[uint32][]validator{
				validatorShard: {mock.NewValidatorMock([]byte("observerKey"), 1, 1)},
			},
		},
	})

	ihgs.ShuffleOutForEpoch(epoch)
	require.False(t, processCalled)
	expectedShardForNotFound := uint32(0)
	require.Equal(t, expectedShardForNotFound, newShard)
}

func TestIndexHashedNodesCoordinator_ShuffleOutNilConfig(t *testing.T) {
	t.Parallel()

	processCalled := false
	newShard := uint32(0)

	arguments := createArguments()
	arguments.ShuffledOutHandler = &mock.ShuffledOutHandlerStub{
		ProcessCalled: func(newShardID uint32) error {
			processCalled = true
			newShard = newShardID
			return nil
		},
	}
	pk := []byte("pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	epoch := uint32(2)
	ihgs.SetNodesConfig(map[uint32]*nodesCoordinator.EpochNodesConfig{
		epoch: nil,
	})

	ihgs.ShuffleOutForEpoch(epoch)
	require.False(t, processCalled)
	expectedShardForNotFound := uint32(0)
	require.Equal(t, expectedShardForNotFound, newShard)
}

func TestIndexHashedNodesCoordinator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var ihgs NodesCoordinator
	require.True(t, check.IfNil(ihgs))

	var ihgs2 *indexHashedNodesCoordinator
	require.True(t, check.IfNil(ihgs2))

	arguments := createArguments()
	ihgs3, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)
	require.False(t, check.IfNil(ihgs3))
}
