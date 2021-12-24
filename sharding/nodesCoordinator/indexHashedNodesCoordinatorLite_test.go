package nodesCoordinator

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
	"github.com/ElrondNetwork/elrond-go-core/data/endProcess"
	"github.com/ElrondNetwork/elrond-go-core/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/stretchr/testify/require"
)

func createDummyNodesList(nbNodes uint32, suffix string) []Validator {
	list := make([]Validator, 0)
	hasher := sha256.NewSha256()

	for j := uint32(0); j < nbNodes; j++ {
		pk := hasher.Compute(fmt.Sprintf("pk%s_%d", suffix, j))
		list = append(list, mock.NewValidatorMock(pk, 1, DefaultSelectionChances))
	}

	return list
}

func createDummyNodesMap(nodesPerShard uint32, nbShards uint32, suffix string) map[uint32][]Validator {
	nodesMap := make(map[uint32][]Validator)

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

func createArguments() ArgNodesCoordinatorLite {
	nbShards := uint32(1)
	eligibleMap := createDummyNodesMap(10, nbShards, "eligible")
	waitingMap := createDummyNodesMap(3, nbShards, "waiting")

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    1,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		NbShards:                   nbShards,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("test"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		WaitingListFixEnabledEpoch: 0,
		IsFullArchive:              false,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	return arguments
}

func validatorsPubKeys(validators []Validator) []string {
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
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)

	require.Equal(t, ErrNilHasher, err)
	require.Nil(t, ihgs)
}

func TestNewIndexHashedNodesCoordinator_InvalidConsensusGroupSizeShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.ShardConsensusGroupSize = 0
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)

	require.Equal(t, ErrInvalidConsensusGroupSize, err)
	require.Nil(t, ihgs)
}

func TestNewIndexHashedNodesCoordinator_ZeroNbShardsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.NbShards = 0
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)

	require.Equal(t, ErrInvalidNumberOfShards, err)
	require.Nil(t, ihgs)
}

func TestNewIndexHashedNodesCoordinator_InvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.ShardIDAsObserver = 10
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)

	require.Equal(t, ErrInvalidShardId, err)
	require.Nil(t, ihgs)
}

func TestNewIndexHashedNodesCoordinator_NilSelfPublicKeyShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.SelfPublicKey = nil
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)

	require.Equal(t, ErrNilPubKey, err)
	require.Nil(t, ihgs)
}

func TestNewIndexHashedNodesCoordinator_NilCacherShouldErr(t *testing.T) {
	arguments := createArguments()
	arguments.ConsensusGroupCache = nil
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)

	require.Equal(t, ErrNilCacher, err)
	require.Nil(t, ihgs)
}

func TestNewIndexHashedGroupSelector_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)

	require.NotNil(t, ihgs)
	require.Nil(t, err)
}

//------- LoadEligibleList

func TestIndexHashedNodesCoordinator_SetNilEligibleMapShouldErr(t *testing.T) {
	t.Parallel()

	waitingMap := createDummyNodesMap(3, 3, "waiting")
	arguments := createArguments()

	ihgs, _ := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Equal(t, ErrNilInputNodesMap, ihgs.SetNodesPerShards(nil, waitingMap, nil, 0))
}

func TestIndexHashedNodesCoordinator_SetNilWaitingMapShouldErr(t *testing.T) {
	t.Parallel()

	eligibleMap := createDummyNodesMap(10, 3, "eligible")
	arguments := createArguments()

	ihgs, _ := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Equal(t, ErrNilInputNodesMap, ihgs.SetNodesPerShards(eligibleMap, nil, nil, 0))
}

func TestIndexHashedNodesCoordinator_OkValShouldWork(t *testing.T) {
	t.Parallel()

	eligibleMap := createDummyNodesMap(10, 3, "eligible")
	waitingMap := createDummyNodesMap(3, 3, "waiting")

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    2,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		NbShards:                   1,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Nil(t, err)

	readEligible := ihgs.nodesConfig[arguments.Epoch].EligibleMap[0]
	require.Equal(t, eligibleMap[0], readEligible)
}

//------- ComputeValidatorsGroup

func TestIndexHashedNodesCoordinator_NewCoordinatorGroup0SizeShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.MetaConsensusGroupSize = 0
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)

	require.Equal(t, ErrInvalidConsensusGroupSize, err)
	require.Nil(t, ihgs)
}

func TestIndexHashedNodesCoordinator_NewCoordinatorTooFewNodesShouldErr(t *testing.T) {
	t.Parallel()

	eligibleMap := createDummyNodesMap(5, 3, "eligible")
	waitingMap := createDummyNodesMap(3, 3, "waiting")

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    10,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		NbShards:                   1,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)

	require.Equal(t, ErrSmallShardEligibleListSize, err)
	require.Nil(t, ihgs)
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroupNilRandomnessShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinatorLite(arguments)
	list2, err := ihgs.ComputeConsensusGroup(nil, 0, 0, 0)

	require.Equal(t, ErrNilRandomness, err)
	require.Nil(t, list2)
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroupInvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinatorLite(arguments)
	list2, err := ihgs.ComputeConsensusGroup([]byte("radomness"), 0, 5, 0)

	require.Equal(t, ErrInvalidShardId, err)
	require.Nil(t, list2)
}

//------- functionality tests

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroup1ValidatorShouldReturnSame(t *testing.T) {
	t.Parallel()

	list := []Validator{
		mock.NewValidatorMock([]byte("pk0"), 1, DefaultSelectionChances),
	}
	tmp := createDummyNodesMap(2, 1, "meta")
	nodesMap := make(map[uint32][]Validator)
	nodesMap[0] = list
	nodesMap[core.MetachainShardId] = tmp[core.MetachainShardId]

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    1,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		NbShards:                   1,
		EligibleNodes:              nodesMap,
		WaitingNodes:               make(map[uint32][]Validator),
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinatorLite(arguments)
	list2, err := ihgs.ComputeConsensusGroup([]byte("randomness"), 0, 0, 0)

	require.Equal(t, list, list2)
	require.Nil(t, err)
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroup400of400For10locksNoMemoization(t *testing.T) {
	consensusGroupSize := 400
	nodesPerShard := uint32(400)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")

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

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    consensusGroupSize,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		NbShards:                   1,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        cache,
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Nil(t, err)

	miniBlocks := 10

	var list2 []Validator
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
	waitingMap := make(map[uint32][]Validator)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")

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

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    consensusGroupSize,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		NbShards:                   1,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        cache,
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Nil(t, err)

	miniBlocks := 10

	var list2 []Validator
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
	waitingMap := make(map[uint32][]Validator)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    consensusGroupSize,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		NbShards:                   1,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        cache,
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Nil(t, err)

	nbDifferentSamplings := 1000
	repeatPerSampling := 100

	list := make([][]Validator, repeatPerSampling)
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
	waitingMap := make(map[uint32][]Validator)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    consensusGroupSize,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		NbShards:                   1,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinatorLite(arguments)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomness := strconv.Itoa(i)
		list2, _ := ihgs.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

		require.Equal(b, consensusGroupSize, len(list2))
	}
}

func runBenchmark(consensusGroupCache Cacher, consensusGroupSize int, nodesMap map[uint32][]Validator, b *testing.B) {
	waitingMap := make(map[uint32][]Validator)

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    consensusGroupSize,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		NbShards:                   1,
		EligibleNodes:              nodesMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        consensusGroupCache,
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinatorLite(arguments)

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

func computeMemoryRequirements(consensusGroupCache Cacher, consensusGroupSize int, nodesMap map[uint32][]Validator, b *testing.B) {
	waitingMap := make(map[uint32][]Validator)

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    consensusGroupSize,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		NbShards:                   1,
		EligibleNodes:              nodesMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        consensusGroupCache,
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)
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
	ihgs, _ := NewIndexHashedNodesCoordinatorLite(arguments)

	_, _, err := ihgs.GetValidatorWithPublicKey(nil)
	require.Equal(t, ErrNilPubKey, err)
}

func TestIndexHashedNodesCoordinator_GetValidatorWithPublicKeyShouldReturnErrValidatorNotFound(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinatorLite(arguments)

	_, _, err := ihgs.GetValidatorWithPublicKey([]byte("pk1"))
	require.Equal(t, ErrValidatorNotFound, err)
}

func TestIndexHashedNodesCoordinator_GetValidatorWithPublicKeyShouldWork(t *testing.T) {
	t.Parallel()

	listMeta := []Validator{
		mock.NewValidatorMock([]byte("pk0_meta"), 1, DefaultSelectionChances),
		mock.NewValidatorMock([]byte("pk1_meta"), 1, DefaultSelectionChances),
		mock.NewValidatorMock([]byte("pk2_meta"), 1, DefaultSelectionChances),
	}
	listShard0 := []Validator{
		mock.NewValidatorMock([]byte("pk0_shard0"), 1, DefaultSelectionChances),
		mock.NewValidatorMock([]byte("pk1_shard0"), 1, DefaultSelectionChances),
		mock.NewValidatorMock([]byte("pk2_shard0"), 1, DefaultSelectionChances),
	}
	listShard1 := []Validator{
		mock.NewValidatorMock([]byte("pk0_shard1"), 1, DefaultSelectionChances),
		mock.NewValidatorMock([]byte("pk1_shard1"), 1, DefaultSelectionChances),
		mock.NewValidatorMock([]byte("pk2_shard1"), 1, DefaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	eligibleMap[core.MetachainShardId] = listMeta
	eligibleMap[0] = listShard0
	eligibleMap[1] = listShard1

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    1,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		NbShards:                   2,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               make(map[uint32][]Validator),
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinatorLite(arguments)

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

	listMeta := []Validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][0], 1, DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][1], 1, DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][2], 1, DefaultSelectionChances),
	}
	listShard0 := []Validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][0], 1, DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][1], 1, DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][2], 1, DefaultSelectionChances),
	}
	listShard1 := []Validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][0], 1, DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][1], 1, DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][2], 1, DefaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	eligibleMap[core.MetachainShardId] = listMeta
	eligibleMap[shardZeroId] = listShard0
	eligibleMap[shardOneId] = listShard1

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    1,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		ShardIDAsObserver:          shardZeroId,
		NbShards:                   2,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               make(map[uint32][]Validator),
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, _ := NewIndexHashedNodesCoordinatorLite(arguments)

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

	listMeta := []Validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][0], 1, DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][1], 1, DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][2], 1, DefaultSelectionChances),
	}
	listShard0 := []Validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][0], 1, DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][1], 1, DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][2], 1, DefaultSelectionChances),
	}
	listShard1 := []Validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][0], 1, DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][1], 1, DefaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][2], 1, DefaultSelectionChances),
	}

	waitingMap := make(map[uint32][]Validator)
	waitingMap[core.MetachainShardId] = listMeta
	waitingMap[shardZeroId] = listShard0
	waitingMap[shardOneId] = listShard1

	eligibleMap := make(map[uint32][]Validator)
	eligibleMap[core.MetachainShardId] = []Validator{&mock.ValidatorMock{}}
	eligibleMap[shardZeroId] = []Validator{&mock.ValidatorMock{}}

	arguments := ArgNodesCoordinatorLite{
		ShardConsensusGroupSize:    1,
		MetaConsensusGroupSize:     1,
		Hasher:                     &mock.HasherMock{},
		ShardIDAsObserver:          shardZeroId,
		NbShards:                   2,
		EligibleNodes:              eligibleMap,
		WaitingNodes:               waitingMap,
		SelfPublicKey:              []byte("key"),
		ConsensusGroupCache:        &mock.NodesCoordinatorCacheMock{},
		WaitingListFixEnabledEpoch: 0,
		ChanStopNode:               make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:           &nodeTypeProviderMock.NodeTypeProviderStub{},
	}

	ihgs, _ := NewIndexHashedNodesCoordinatorLite(arguments)

	allValidatorsPublicKeys, err := ihgs.GetAllWaitingValidatorsPublicKeys(0)
	require.Equal(t, expectedValidatorsPubKeys, allValidatorsPublicKeys)
	require.Nil(t, err)
}

func TestIndexHashedNodesCoordinator_setNodesPerShardsShouldTriggerWrongConfiguration(t *testing.T) {
	t.Parallel()

	chanStopNode := make(chan endProcess.ArgEndProcess, 1)
	arguments := createArguments()
	arguments.ChanStopNode = chanStopNode
	arguments.IsFullArchive = true

	pk := []byte("pk")
	arguments.SelfPublicKey = pk
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Nil(t, err)

	eligibleMap := map[uint32][]Validator{
		core.MetachainShardId: {
			mock.NewValidatorMock(pk, 1, 1),
		},
	}

	err = ihgs.SetNodesPerShards(eligibleMap, map[uint32][]Validator{}, map[uint32][]Validator{}, 2)
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
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Nil(t, err)

	eligibleMap := map[uint32][]Validator{
		core.MetachainShardId: {
			mock.NewValidatorMock(pk, 1, 1),
		},
	}

	err = ihgs.SetNodesPerShards(eligibleMap, map[uint32][]Validator{}, map[uint32][]Validator{}, 2)
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
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Nil(t, err)

	eligibleMap := map[uint32][]Validator{
		core.MetachainShardId: {
			mock.NewValidatorMock(pk, 1, 1),
		},
	}

	err = ihgs.SetNodesPerShards(eligibleMap, map[uint32][]Validator{}, map[uint32][]Validator{}, 2)
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
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Nil(t, err)

	eligibleMap := map[uint32][]Validator{
		core.MetachainShardId: {
			mock.NewValidatorMock([]byte("validator pk"), 1, 1),
		},
	}

	err = ihgs.SetNodesPerShards(eligibleMap, map[uint32][]Validator{}, map[uint32][]Validator{}, 2)
	require.NoError(t, err)
	require.True(t, setTypeWasCalled)
	require.Equal(t, core.NodeTypeObserver, nodeTypeResult)
}

func TestIndexHashedNodesCoordinator_GetConsensusValidatorsPublicKeysNotExistingEpoch(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinatorLite(args)
	require.Nil(t, err)

	var pKeys []string
	randomness := []byte("randomness")
	pKeys, err = ihgs.GetConsensusValidatorsPublicKeys(randomness, 0, 0, 1)
	require.True(t, errors.Is(err, ErrEpochNodesConfigDoesNotExist))
	require.Nil(t, pKeys)
}

func TestIndexHashedNodesCoordinator_GetConsensusValidatorsPublicKeysExistingEpoch(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinatorLite(args)
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
	ihgs, err := NewIndexHashedNodesCoordinatorLite(args)
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
	ihgs, err := NewIndexHashedNodesCoordinatorLite(args)
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

func TestIndexHashedNodesCoordinator_ShardIdForEpochInvalidEpoch(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinatorLite(args)
	require.Nil(t, err)

	shardId, err := ihgs.ShardIdForEpoch(1)
	require.True(t, errors.Is(err, ErrEpochNodesConfigDoesNotExist))
	require.Equal(t, uint32(0), shardId)
}

func TestIndexHashedNodesCoordinator_ShardIdForEpochValidEpoch(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinatorLite(args)
	require.Nil(t, err)

	shardId, err := ihgs.ShardIdForEpoch(0)
	require.Nil(t, err)
	require.Equal(t, uint32(0), shardId)
}

func TestIndexHashedNodesCoordinator_ConsensusGroupSize(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Nil(t, err)

	consensusSizeShard := ihgs.ConsensusGroupSize(0)
	consensusSizeMeta := ihgs.ConsensusGroupSize(core.MetachainShardId)

	require.Equal(t, arguments.ShardConsensusGroupSize, consensusSizeShard)
	require.Equal(t, arguments.MetaConsensusGroupSize, consensusSizeMeta)
}

func TestIndexHashedNodesCoordinator_GetNumTotalEligible(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)
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
	ihgs, err := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Nil(t, err)

	ownPubKey := ihgs.GetOwnPublicKey()
	require.Equal(t, arguments.SelfPublicKey, ownPubKey)
}

func TestIndexHashedNodesCoordinator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var ihgs NodesCoordinatorLite
	require.True(t, check.IfNil(ihgs))

	var ihgs2 *IndexHashedNodesCoordinatorLite
	require.True(t, check.IfNil(ihgs2))

	arguments := createArguments()
	ihgs3, err := NewIndexHashedNodesCoordinatorLite(arguments)
	require.Nil(t, err)
	require.False(t, check.IfNil(ihgs3))
}
