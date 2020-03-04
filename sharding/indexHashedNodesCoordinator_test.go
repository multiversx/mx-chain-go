package sharding

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
	"github.com/stretchr/testify/require"
)

func uint64ToBytes(value uint64) []byte {
	buff := make([]byte, 8)
	binary.BigEndian.PutUint64(buff, value)

	return buff
}

func createDummyNodesMap(nodesPerShard uint32, nbShards uint32, suffix string) map[uint32][]Validator {
	nodesMap := make(map[uint32][]Validator)

	for i := uint32(0); i <= nbShards; i++ {
		shard := i
		list := make([]Validator, 0)
		if i == nbShards {
			shard = core.MetachainShardId
		}

		for j := uint32(0); j < nodesPerShard; j++ {
			pk := []byte(fmt.Sprintf("pk%s_%d_%d", suffix, i, j))
			addr := []byte(fmt.Sprintf("addr%s_%d_%d", suffix, i, j))
			list = append(list, mock.NewValidatorMock(pk, addr))
		}

		nodesMap[shard] = list
	}

	return nodesMap
}

func containStrings(a []string, b []string) bool {
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
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                nbShards,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("test"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}
	return arguments
}

func genRandSource(round uint64, randomness string) string {
	return fmt.Sprintf("%d-%s", round, []byte(randomness))
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
	arguments.ShardId = 10
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
	require.Equal(t, ErrNilInputNodesMap, ihgs.SetNodesPerShards(nil, waitingMap, 0))
}

func TestIndexHashedNodesCoordinator_SetNilWaitingMapShouldErr(t *testing.T) {
	t.Parallel()

	eligibleMap := createDummyNodesMap(10, 3, "eligible")
	arguments := createArguments()

	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
	require.Equal(t, ErrNilInputNodesMap, ihgs.SetNodesPerShards(eligibleMap, nil, 0))
}

func TestIndexHashedNodesCoordinator_OkValShouldWork(t *testing.T) {
	t.Parallel()

	eligibleMap := createDummyNodesMap(10, 3, "eligible")
	waitingMap := createDummyNodesMap(3, 3, "waiting")
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 2,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	readEligible := ihgs.nodesConfig[arguments.Epoch].eligibleMap[0]
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
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 10,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
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

	list := []Validator{
		mock.NewValidatorMock([]byte("pk0"), []byte("addr0")),
	}
	tmp := createDummyNodesMap(2, 1, "meta")
	nodesMap := make(map[uint32][]Validator)
	nodesMap[0] = list
	nodesMap[core.MetachainShardId] = tmp[core.MetachainShardId]
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           nodesMap,
		WaitingNodes:            make(map[uint32][]Validator),
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
	list2, err := ihgs.ComputeConsensusGroup([]byte("randomness"), 0, 0, 0)

	require.Equal(t, list, list2)
	require.Nil(t, err)
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroupTest2Validators(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherStub{}

	randomness := "randomness"

	//this will return the list in order:
	//element 0 will be first element
	//element 1 will be the second
	hasher.ComputeCalled = func(s string) []byte {
		if strings.Contains(s, "0-") {
			return uint64ToBytes(0)
		}

		if strings.Contains(s, "1-") {
			return uint64ToBytes(1)
		}

		return nil
	}

	eligibleMap := createDummyNodesMap(10, 3, "eligible")
	waitingMap := createDummyNodesMap(3, 3, "waiting")
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 2,
		MetaConsensusGroupSize:  1,
		Hasher:                  hasher,
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	list2, err := ihgs.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

	require.Equal(t, eligibleMap[0][:2], list2)
	require.Nil(t, err)
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroupTest2ValidatorsRevertOrder(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherStub{}

	randomness := "randomness"
	randSource := genRandSource(0, randomness)

	//this will return the list in reverse order:
	//element 0 will be the second
	//element 1 will be the first
	hasher.ComputeCalled = func(s string) []byte {
		if string(uint64ToBytes(0))+randSource == s {
			return uint64ToBytes(1)
		}

		if string(uint64ToBytes(1))+randSource == s {
			return uint64ToBytes(0)
		}

		return nil
	}

	validator0 := mock.NewValidatorMock([]byte("pk0"), []byte("addr0"))
	validator1 := mock.NewValidatorMock([]byte("pk1"), []byte("addr1"))

	list := []Validator{
		validator0,
		validator1,
	}

	eligibleMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	metaNode, _ := NewValidator([]byte("pubKeyMeta"), []byte("addressMeta"))
	eligibleMap[core.MetachainShardId] = []Validator{metaNode}
	waitingMap := make(map[uint32][]Validator)
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 2,
		MetaConsensusGroupSize:  1,
		Hasher:                  hasher,
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	list2, err := ihgs.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

	require.Nil(t, err)
	require.Equal(t, validator0, list2[1])
	require.Equal(t, validator1, list2[0])
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroupTest2ValidatorsSameIndex(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherStub{}

	randomness := "randomness"

	//this will return the list in order:
	//element 0 will be the first
	//element 1 will be the second as the same index is being returned and 0 is already in list
	hasher.ComputeCalled = func(s string) []byte {
		if strings.Contains(s, "0-") {
			return uint64ToBytes(0)
		}

		if strings.Contains(s, "1-") {
			return uint64ToBytes(1)
		}

		return nil
	}

	eligibleMap := createDummyNodesMap(10, 3, "eligible")
	waitingMap := createDummyNodesMap(3, 3, "waiting")
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 2,
		MetaConsensusGroupSize:  1,
		Hasher:                  hasher,
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	list2, err := ihgs.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

	require.Nil(t, err)
	require.Equal(t, eligibleMap[0][:2], list2)
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroupTest6From10ValidatorsShouldWork(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherStub{}
	selfPubKey := []byte("key")
	randomness := "randomness"
	randomnessWithRound := genRandSource(0, randomness)

	//script:
	// for index 0, hasher will return 11 which will translate to 1, so index 1 will be used ; num appearances = 1 => size = 1

	// for index 1, hasher will return 1 which will translate to 1, 1 is already picked, size will be added so the
	// new calculated index will 2 ; appearances = 1 => size = 2

	// for index 2, hasher will return 9 , 9 % (10 - 2) = 1 ; 1 is already picked so add the size (2) and the new
	// validator will be from index 3 ; appearances = 1 => size = 3

	// for index 3, hasher will return 9 ; 9 % (10 - 3) = 2 ; 2 > 1 (first element in slice) so add the size (3) and the new
	// validator will be from index 5 ; appearances = 1 => size = 4

	// for index 4, hasher will return 0 ; 0 % (10 - 4) = 0 so the new validator will be from index 0 ;
	// num appearances = 1 => size = 5

	// for index 5, hasher will return 9 ; 9 % (10 - 5) = 4 ; 4 > 0 (first element in sorted slice) so size will be added
	// and will return the index 9 for the validator
	script := make(map[string]uint64)

	script[string(uint64ToBytes(0))+randomnessWithRound] = 11 //will translate to 1, add 1
	script[string(uint64ToBytes(1))+randomnessWithRound] = 1  //will translate to 1, add 2
	script[string(uint64ToBytes(2))+randomnessWithRound] = 9  //will translate to 9, add 9
	script[string(uint64ToBytes(3))+randomnessWithRound] = 9  //will translate to 9, add 0
	script[string(uint64ToBytes(4))+randomnessWithRound] = 0  //will translate to 0, add 3
	script[string(uint64ToBytes(5))+randomnessWithRound] = 9  //will translate to 9, add 4

	hasher.ComputeCalled = func(s string) []byte {
		if s == string(selfPubKey) {
			return []byte(s)
		}

		val, ok := script[s]
		if !ok {
			require.Fail(t, "should have not got here")
		}

		return uint64ToBytes(val)
	}

	validator0 := mock.NewValidatorMock([]byte("pk0"), []byte("addr0"))
	validator1 := mock.NewValidatorMock([]byte("pk1"), []byte("addr1"))
	validator2 := mock.NewValidatorMock([]byte("pk2"), []byte("addr2"))
	validator3 := mock.NewValidatorMock([]byte("pk3"), []byte("addr3"))
	validator4 := mock.NewValidatorMock([]byte("pk4"), []byte("addr4"))
	validator5 := mock.NewValidatorMock([]byte("pk5"), []byte("addr5"))
	validator6 := mock.NewValidatorMock([]byte("pk6"), []byte("addr6"))
	validator7 := mock.NewValidatorMock([]byte("pk7"), []byte("addr7"))
	validator8 := mock.NewValidatorMock([]byte("pk8"), []byte("addr8"))
	validator9 := mock.NewValidatorMock([]byte("pk9"), []byte("addr9"))

	list := []Validator{
		validator0,
		validator1,
		validator2,
		validator3,
		validator4,
		validator5,
		validator6,
		validator7,
		validator8,
		validator9,
	}

	eligibleMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	validatorMeta, _ := NewValidator([]byte("pubKeyMeta"), []byte("addressMeta"))
	eligibleMap[core.MetachainShardId] = []Validator{validatorMeta}
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 6,
		MetaConsensusGroupSize:  1,
		Hasher:                  hasher,
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            make(map[uint32][]Validator),
		SelfPublicKey:           selfPubKey,
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	list2, err := ihgs.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

	require.Nil(t, err)
	require.Equal(t, 6, len(list2))
	//check order as described in script
	require.Equal(t, validator1, list2[0])
	require.Equal(t, validator2, list2[1])
	require.Equal(t, validator3, list2[2])
	require.Equal(t, validator5, list2[3])
	require.Equal(t, validator0, list2[4])
	require.Equal(t, validator9, list2[5])
}

func TestIndexHashedNodesCoordinator_ComputeValidatorsGroup400of400For10locksNoMemoization(t *testing.T) {
	consensusGroupSize := 400
	nodesPerShard := uint32(400)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")
	nodeShuffler := NewXorValidatorsShuffler(nodesPerShard, nodesPerShard, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	getCounter := int32(0)
	putCounter := int32(0)

	cache := &mock.NodesCoordinatorCacheMock{
		PutCalled: func(key []byte, value interface{}) (evicted bool) {
			atomic.AddInt32(&putCounter, 1)
			return false
		},
		GetCalled: func(key []byte) (value interface{}, ok bool) {
			atomic.AddInt32(&getCounter, 1)
			return nil, false
		},
	}

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: consensusGroupSize,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     cache,
	}

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
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
	nodeShuffler := NewXorValidatorsShuffler(nodesPerShard, nodesPerShard, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	getCounter := 0
	putCounter := 0

	mut := sync.Mutex{}

	//consensusGroup := list[0:21]
	cacheMap := make(map[string]interface{})
	cache := &mock.NodesCoordinatorCacheMock{
		PutCalled: func(key []byte, value interface{}) (evicted bool) {
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
		ShardConsensusGroupSize: consensusGroupSize,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     cache,
	}

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
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

func BenchmarkIndexHashedGroupSelector_ComputeValidatorsGroup21of400(b *testing.B) {
	consensusGroupSize := 21
	nodesPerShard := uint32(400)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap := createDummyNodesMap(nodesPerShard, 1, "eligible")
	nodeShuffler := NewXorValidatorsShuffler(nodesPerShard, nodesPerShard, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: consensusGroupSize,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomness := strconv.Itoa(i)
		list2, _ := ihgs.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

		require.Equal(b, consensusGroupSize, len(list2))
	}
}

func runBenchmark(consensusGroupCache Cacher, consensusGroupSize int, nodesMap map[uint32][]Validator, b *testing.B) {
	waitingMap := make(map[uint32][]Validator)
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: consensusGroupSize,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		EpochStartSubscriber:    epochStartSubscriber,
		Shuffler:                nodeShuffler,
		NbShards:                1,
		EligibleNodes:           nodesMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     consensusGroupCache,
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

func computeMemoryRequirements(consensusGroupCache Cacher, consensusGroupSize int, nodesMap map[uint32][]Validator, b *testing.B) {
	waitingMap := make(map[uint32][]Validator)
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: consensusGroupSize,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		EpochStartSubscriber:    epochStartSubscriber,
		Shuffler:                nodeShuffler,
		NbShards:                1,
		EligibleNodes:           nodesMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     consensusGroupCache,
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

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

	fmt.Println(fmt.Sprintf("Used %d MB", (m2.HeapAlloc-m.HeapAlloc)/1024/1024))
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

	_, _, err := ihgs.GetValidatorWithPublicKey(nil, 0)
	require.Equal(t, ErrNilPubKey, err)
}

func TestIndexHashedNodesCoordinator_GetValidatorWithPublicKeyShouldReturnErrValidatorNotFound(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	_, _, err := ihgs.GetValidatorWithPublicKey([]byte("pk1"), 0)
	require.Equal(t, ErrValidatorNotFound, err)
}

func TestIndexHashedNodesCoordinator_GetValidatorWithPublicKeyShouldWork(t *testing.T) {
	t.Parallel()

	listMeta := []Validator{
		mock.NewValidatorMock([]byte("pk0_meta"), []byte("addr0_meta")),
		mock.NewValidatorMock([]byte("pk1_meta"), []byte("addr1_meta")),
		mock.NewValidatorMock([]byte("pk2_meta"), []byte("addr2_meta")),
	}
	listShard0 := []Validator{
		mock.NewValidatorMock([]byte("pk0_shard0"), []byte("addr0_shard0")),
		mock.NewValidatorMock([]byte("pk1_shard0"), []byte("addr1_shard0")),
		mock.NewValidatorMock([]byte("pk2_shard0"), []byte("addr2_shard0")),
	}
	listShard1 := []Validator{
		mock.NewValidatorMock([]byte("pk0_shard1"), []byte("addr0_shard1")),
		mock.NewValidatorMock([]byte("pk1_shard1"), []byte("addr1_shard1")),
		mock.NewValidatorMock([]byte("pk2_shard1"), []byte("addr2_shard1")),
	}

	eligibleMap := make(map[uint32][]Validator)
	eligibleMap[core.MetachainShardId] = listMeta
	eligibleMap[0] = listShard0
	eligibleMap[1] = listShard1
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                2,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            make(map[uint32][]Validator),
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	v, shardId, err := ihgs.GetValidatorWithPublicKey([]byte("pk0_meta"), 0)
	require.Nil(t, err)
	require.Equal(t, core.MetachainShardId, shardId)
	require.Equal(t, []byte("addr0_meta"), v.Address())

	v, shardId, err = ihgs.GetValidatorWithPublicKey([]byte("pk1_shard0"), 0)
	require.Nil(t, err)
	require.Equal(t, uint32(0), shardId)
	require.Equal(t, []byte("addr1_shard0"), v.Address())

	v, shardId, err = ihgs.GetValidatorWithPublicKey([]byte("pk2_shard1"), 0)
	require.Nil(t, err)
	require.Equal(t, uint32(1), shardId)
	require.Equal(t, []byte("addr2_shard1"), v.Address())
}

func TestNewIndexHashedNodesCoordinator_GetValidatorWithPublicKeyNotExistingEpoch(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	_, _, err := ihgs.GetValidatorWithPublicKey(arguments.EligibleNodes[0][0].PubKey(), 1)
	require.Equal(t, ErrEpochNodesConfigDesNotExist, err)
}

func TestIndexHashedNodesCoordinator_GetAllValidatorsPublicKeys(t *testing.T) {
	t.Parallel()

	shardZeroId := uint32(0)
	shardOneId := uint32(1)
	expectedValidatorsPubKeys := map[uint32][][]byte{
		shardZeroId:           {[]byte("pk0_shard0"), []byte("pk1_shard0"), []byte("pk2_shard0")},
		shardOneId:            {[]byte("pk0_shard1"), []byte("pk1_shard1"), []byte("pk2_shard1")},
		core.MetachainShardId: {[]byte("pk0_meta"), []byte("pk1_meta"), []byte("pk2_meta")},
	}

	listMeta := []Validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][0], []byte("addr0_meta")),
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][1], []byte("addr1_meta")),
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][2], []byte("addr2_meta")),
	}
	listShard0 := []Validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][0], []byte("addr0_shard0")),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][1], []byte("addr1_shard0")),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][2], []byte("addr2_shard0")),
	}
	listShard1 := []Validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][0], []byte("addr0_shard1")),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][1], []byte("addr1_shard1")),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][2], []byte("addr2_shard1")),
	}

	eligibleMap := make(map[uint32][]Validator)
	eligibleMap[core.MetachainShardId] = listMeta
	eligibleMap[shardZeroId] = listShard0
	eligibleMap[shardOneId] = listShard1
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		ShardId:                 shardZeroId,
		NbShards:                2,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            make(map[uint32][]Validator),
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}

	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	allValidatorsPublicKeys, err := ihgs.GetAllValidatorsPublicKeys(0)
	require.Equal(t, expectedValidatorsPubKeys, allValidatorsPublicKeys)
	require.Nil(t, err)
}

func TestIndexHashedNodesCoordinator_EpochStart(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	header := &mock.HeaderHandlerStub{
		GetPrevRandSeedCalled: func() []byte {
			return []byte("rand seed")
		},
		IsStartOfEpochBlockCalled: func() bool {
			return true
		},
		GetEpochCaled: func() uint32 {
			return 1
		},
	}

	ihgs.EpochStartPrepare(header)
	ihgs.EpochStartAction(header)

	validators, err := ihgs.GetAllValidatorsPublicKeys(1)
	require.Nil(t, err)
	require.NotNil(t, validators)

	computedShardId := ihgs.computeShardForPublicKey(ihgs.nodesConfig[0])
	// should remain in same shard with intra shard shuffling
	require.Equal(t, arguments.ShardId, computedShardId)
}

func TestIndexHashedNodesCoordinator_GetConsensusValidatorsPublicKeysNotExistingEpoch(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	var pKeys []string
	randomness := []byte("randomness")
	pKeys, err = ihgs.GetConsensusValidatorsPublicKeys(randomness, 0, 0, 1)
	require.Equal(t, ErrEpochNodesConfigDesNotExist, err)
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
	require.True(t, containStrings(pKeys, shard0PubKeys))
}

func TestIndexHashedNodesCoordinator_GetConsensusValidatorsRewardsAddressesInvalidRandomness(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	var addresses []string
	addresses, err = ihgs.GetConsensusValidatorsRewardsAddresses(nil, 0, 0, 0)
	require.Equal(t, ErrNilRandomness, err)
	require.Nil(t, addresses)
}

func TestIndexHashedNodesCoordinator_GetConsensusValidatorsRewardsAddressesOK(t *testing.T) {
	t.Parallel()

	args := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(args)
	require.Nil(t, err)

	var addresses []string
	randomness := []byte("randomness")
	addresses, err = ihgs.GetConsensusValidatorsRewardsAddresses(randomness, 0, 0, 0)
	require.Nil(t, err)
	require.True(t, len(addresses) > 0)
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

	epoch := uint32(1)
	header := &mock.HeaderHandlerStub{
		GetPrevRandSeedCalled: func() []byte {
			return []byte("rand seed")
		},
		IsStartOfEpochBlockCalled: func() bool {
			return true
		},
		GetEpochCaled: func() uint32 {
			return atomic.LoadUint32(&epoch)
		},
	}

	ihgs.EpochStartPrepare(header)
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
	require.Equal(t, ErrEpochNodesConfigDesNotExist, err)
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

	nodesCurrentEpoch, err := ihgs.GetAllValidatorsPublicKeys(0)
	require.Nil(t, err)

	allNodesList := make([]string, 0)
	for _, nodesList := range nodesCurrentEpoch {
		for _, nodeKey := range nodesList {
			allNodesList = append(allNodesList, string(nodeKey))
		}
	}

	whitelistedNodes, err := ihgs.GetConsensusWhitelistedNodes(0)
	require.Nil(t, err)
	require.True(t, len(whitelistedNodes) > 0)

	for key := range whitelistedNodes {
		require.True(t, containStrings([]string{key}, allNodesList))
	}
}

func TestIndexHashedNodesCoordinator_GetConsensusWhitelistedNodesEpoch1(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	header := &mock.HeaderHandlerStub{
		GetPrevRandSeedCalled: func() []byte {
			return []byte("rand seed")
		},
		IsStartOfEpochBlockCalled: func() bool {
			return true
		},
		GetEpochCaled: func() uint32 {
			return 1
		},
	}

	ihgs.EpochStartPrepare(header)
	ihgs.EpochStartAction(header)

	nodesPrevEpoch, err := ihgs.GetAllValidatorsPublicKeys(0)
	require.Nil(t, err)
	nodesCurrentEpoch, err := ihgs.GetAllValidatorsPublicKeys(1)
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
	require.True(t, len(whitelistedNodes) > 0)

	for key := range whitelistedNodes {
		require.True(t, containStrings([]string{key}, allNodesList))
	}
}

func TestIndexHashedNodesCoordinator_GetConsensusWhitelistedNodesAfterRevertToEpoch(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)

	epoch := uint32(1)
	header := &mock.HeaderHandlerStub{
		GetPrevRandSeedCalled: func() []byte {
			return []byte("rand seed")
		},
		IsStartOfEpochBlockCalled: func() bool {
			return true
		},
		GetEpochCaled: func() uint32 {
			return atomic.LoadUint32(&epoch)
		},
	}

	ihgs.EpochStartPrepare(header)
	ihgs.EpochStartAction(header)

	atomic.StoreUint32(&epoch, 2)
	ihgs.EpochStartPrepare(header)
	ihgs.EpochStartAction(header)

	nodesEpoch1, err := ihgs.GetAllValidatorsPublicKeys(1)
	require.Nil(t, err)

	allNodesList := make([]string, 0)
	for _, nodesList := range nodesEpoch1 {
		for _, nodeKey := range nodesList {
			allNodesList = append(allNodesList, string(nodeKey))
		}
	}

	whitelistedNodes, err := ihgs.GetConsensusWhitelistedNodes(1)
	require.Nil(t, err)
	require.True(t, len(whitelistedNodes) > 0)

	for key := range whitelistedNodes {
		require.True(t, containStrings([]string{key}, allNodesList))
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
