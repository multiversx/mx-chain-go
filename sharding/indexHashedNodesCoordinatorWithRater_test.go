package sharding

import (
	"fmt"
	"math/big"
	"math/rand"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIndexHashedNodesCoordinatorWithRater_NilRaterShouldErr(t *testing.T) {
	nc, _ := NewIndexHashedNodesCoordinator(createArguments())
	ihgs, err := NewIndexHashedNodesCoordinatorWithRater(nc, nil)

	assert.Nil(t, ihgs)
	assert.Equal(t, ErrNilRater, err)
}

func TestNewIndexHashedNodesCoordinatorWithRater_NilNodesCoordinatorShouldErr(t *testing.T) {
	ihgs, err := NewIndexHashedNodesCoordinatorWithRater(nil, &mock.RaterMock{})

	assert.Nil(t, ihgs)
	assert.Equal(t, ErrNilNodesCoordinator, err)
}

func TestNewIndexHashedGroupSelectorWithRater_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	nc, _ := NewIndexHashedNodesCoordinator(createArguments())
	ihgs, err := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{})
	assert.NotNil(t, ihgs)
	assert.Nil(t, err)
}

//------- LoadEligibleList

func TestIndexHashedGroupSelectorWithRater_SetNilEligibleMapShouldErr(t *testing.T) {
	t.Parallel()
	waiting := createDummyNodesMap(2, 1, "waiting")
	nc, _ := NewIndexHashedNodesCoordinator(createArguments())
	ihgs, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{})
	assert.Equal(t, ErrNilInputNodesMap, ihgs.setNodesPerShards(nil, waiting, 0))
}

func TestIndexHashedGroupSelectorWithRater_OkValShouldWork(t *testing.T) {
	t.Parallel()

	eligibleMap := createDummyNodesMap(3, 1, "waiting")
	waitingMap := make(map[uint32][]Validator)
	nodeShuffler := NewXorValidatorsShuffler(3, 3, 0, false)
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
		SelfPublicKey:           []byte("test"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}
	nc, err := NewIndexHashedNodesCoordinator(arguments)
	assert.Nil(t, err)
	readEligible := nc.nodesConfig[0].eligibleMap[0]
	assert.Equal(t, eligibleMap[0], readEligible)

	rater := &mock.RaterMock{}
	ihgs, err := NewIndexHashedNodesCoordinatorWithRater(nc, rater)
	assert.Nil(t, err)

	for i := 0; i < len(eligibleMap[0]); i++ {
		eligibleMap[0][i].SetChances(rater.GetChance(0))
	}
	readEligible = ihgs.nodesConfig[0].eligibleMap[0]
	assert.Equal(t, eligibleMap[0], readEligible)
}

//------- functionality tests

func TestIndexHashedGroupSelectorWithRater_ComputeValidatorsGroup1ValidatorShouldNotCallGetRating(t *testing.T) {
	t.Parallel()

	list := []Validator{
		mock.NewValidatorMock([]byte("pk0"), []byte("addr0"), defaultSelectionChances),
	}

	arguments := createArguments()
	arguments.EligibleNodes[0] = list

	raterCalled := false
	chancesCalled := false
	rater := &mock.RaterMock{GetRatingCalled: func(string) uint32 {
		raterCalled = true
		return 1
	},
		GetChancesCalled: func(u uint32) uint32 {
			chancesCalled = true
			return 1
		}}

	nc, err := NewIndexHashedNodesCoordinator(arguments)
	assert.Nil(t, err)
	assert.Equal(t, false, raterCalled)
	assert.Equal(t, false, chancesCalled)
	ihgs, _ := NewIndexHashedNodesCoordinatorWithRater(nc, rater)
	assert.Equal(t, true, raterCalled)
	assert.Equal(t, true, chancesCalled)
	list2, err := ihgs.ComputeConsensusGroup([]byte("randomness"), 0, 0, 0)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(list2))
}

func TestNewIndexHashedGroupSelectorWithRaterSetsChances(t *testing.T) {
	t.Parallel()

	list := []Validator{
		mock.NewValidatorMock([]byte("pk0"), []byte("addr0"), defaultSelectionChances),
		mock.NewValidatorMock([]byte("pk1"), []byte("addr1"), defaultSelectionChances),
	}

	listMeta := []Validator{
		mock.NewValidatorMock([]byte("pkMeta1"), []byte("addrMeta1"), defaultSelectionChances),
		mock.NewValidatorMock([]byte("pkMeta2"), []byte("addrMeta2"), defaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	eligibleMap[core.MetachainShardId] = listMeta
	waitingMap := make(map[uint32][]Validator)
	nodeShuffler := NewXorValidatorsShuffler(2, 2, 0, false)
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

	ratingPk0 := uint32(5)
	ratingPk1 := uint32(1)
	rater := &mock.RaterMock{
		GetRatingCalled: func(pk string) uint32 {
			if pk == "pk0" {
				return ratingPk0
			}
			if pk == "pk1" {
				return ratingPk1
			}
			return 1
		},
		GetChancesCalled: func(rating uint32) uint32 {
			return rating
		},
	}

	nc, _ := NewIndexHashedNodesCoordinator(arguments)
	ihgs, _ := NewIndexHashedNodesCoordinatorWithRater(nc, rater)

	eligibleNodes := ihgs.nodesConfig[0].eligibleMap[0]
	for _, v := range eligibleNodes {
		chances := v.Chances()
		require.Equal(t, rater.GetRatingCalled(string(v.PubKey())), chances)
	}
}

func BenchmarkIndexHashedGroupSelectorWithRater_ComputeValidatorsGroup63of400(b *testing.B) {
	b.ReportAllocs()

	consensusGroupSize := 63
	list := make([]Validator, 0)

	//generate 400 validators
	for i := 0; i < 400; i++ {
		list = append(list, mock.NewValidatorMock([]byte("pk"+strconv.Itoa(i)), []byte("addr"+strconv.Itoa(i)), defaultSelectionChances))
	}
	listMeta := []Validator{
		mock.NewValidatorMock([]byte("pkMeta1"), []byte("addrMeta1"), defaultSelectionChances),
		mock.NewValidatorMock([]byte("pkMeta2"), []byte("addrMeta2"), defaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	eligibleMap[core.MetachainShardId] = listMeta
	nodeShuffler := NewXorValidatorsShuffler(400, 1, 0, false)
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
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(b, err)
	ihgsRater, err := NewIndexHashedNodesCoordinatorWithRater(ihgs, &mock.RaterMock{})
	require.Nil(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomness := strconv.Itoa(0)
		list2, _ := ihgsRater.ComputeConsensusGroup([]byte(randomness), uint64(0), 0, 0)

		assert.Equal(b, consensusGroupSize, len(list2))
	}
}

func Test_ComputeValidatorsGroup63of400(t *testing.T) {
	t.Skip("Long test")

	consensusGroupSize := 63
	shardSize := uint32(400)
	list := make([]Validator, 0)

	//generate 400 validators
	for i := uint32(0); i < shardSize; i++ {
		list = append(list, mock.NewValidatorMock([]byte(fmt.Sprintf("pk%v", i)), []byte(fmt.Sprintf("addr%v", i)), defaultSelectionChances))
	}
	listMeta := []Validator{
		mock.NewValidatorMock([]byte("pkMeta1"), []byte("addrMeta1"), defaultSelectionChances),
		mock.NewValidatorMock([]byte("pkMeta2"), []byte("addrMeta2"), defaultSelectionChances),
	}

	consensusAppearances := make(map[string]uint64)
	leaderAppearances := make(map[string]uint64)
	for _, validator := range list {
		consensusAppearances[string(validator.PubKey())] = 0
		leaderAppearances[string(validator.PubKey())] = 0
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	eligibleMap[core.MetachainShardId] = listMeta
	nodeShuffler := NewXorValidatorsShuffler(shardSize, 1, 0, false)
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
	numRounds := uint64(1000000)
	hasher := sha256.Sha256{}
	for i := uint64(0); i < numRounds; i++ {
		randomness := hasher.Compute(fmt.Sprintf("%v%v", i, time.Millisecond))
		consensusGroup, _ := ihgs.ComputeConsensusGroup((randomness), uint64(0), 0, 0)
		leaderAppearances[string(consensusGroup[0].PubKey())]++
		for _, v := range consensusGroup {
			consensusAppearances[string(v.PubKey())]++
		}
	}

	leaderAverage := numRounds / uint64(shardSize)
	percentDifference := leaderAverage * 5 / 100
	for pk, v := range leaderAppearances {
		if v < leaderAverage-percentDifference || v > leaderAverage+percentDifference {
			log.Warn("leader outside of 5%", "pk", pk, "leaderAverage", leaderAverage, "actual", v)
		}
	}

	validatorAverage := numRounds * uint64(consensusGroupSize) / uint64(shardSize)
	percentDifference = validatorAverage * 5 / 100
	for pk, v := range consensusAppearances {
		if v < validatorAverage-percentDifference || v > validatorAverage+percentDifference {
			log.Warn("validator outside of 5%", "pk", pk, "validatorAverage", validatorAverage, "actual", v)
		}
	}

}

func TestIndexHashedGroupSelectorWithRater_GetValidatorWithPublicKeyShouldReturnErrNilPubKey(t *testing.T) {
	t.Parallel()

	list := []Validator{
		mock.NewValidatorMock([]byte("pk0"), []byte("addr0"), defaultSelectionChances),
	}
	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	eligibleMap[core.MetachainShardId] = list
	nodeShuffler := NewXorValidatorsShuffler(1, 1, 0, false)
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
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}
	nc, _ := NewIndexHashedNodesCoordinator(arguments)
	ihgs, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{})

	_, _, err := ihgs.GetValidatorWithPublicKey(nil, 0)
	assert.Equal(t, ErrNilPubKey, err)
}

func TestIndexHashedGroupSelectorWithRater_GetValidatorWithPublicKeyShouldReturnErrValidatorNotFound(t *testing.T) {
	t.Parallel()

	list := []Validator{
		mock.NewValidatorMock([]byte("pk0"), []byte("addr0"), defaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	eligibleMap[core.MetachainShardId] = list
	nodeShuffler := NewXorValidatorsShuffler(1, 1, 0, false)
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
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}
	nc, _ := NewIndexHashedNodesCoordinator(arguments)
	ihgs, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{})

	_, _, err := ihgs.GetValidatorWithPublicKey([]byte("pk1"), 0)
	assert.Equal(t, ErrValidatorNotFound, err)
}

func TestIndexHashedGroupSelectorWithRater_GetValidatorWithPublicKeyShouldWork(t *testing.T) {
	t.Parallel()

	listMeta := []Validator{
		mock.NewValidatorMock([]byte("pk0_meta"), []byte("addr0_meta"), defaultSelectionChances),
		mock.NewValidatorMock([]byte("pk1_meta"), []byte("addr1_meta"), defaultSelectionChances),
		mock.NewValidatorMock([]byte("pk2_meta"), []byte("addr2_meta"), defaultSelectionChances),
	}
	listShard0 := []Validator{
		mock.NewValidatorMock([]byte("pk0_shard0"), []byte("addr0_shard0"), defaultSelectionChances),
		mock.NewValidatorMock([]byte("pk1_shard0"), []byte("addr1_shard0"), defaultSelectionChances),
		mock.NewValidatorMock([]byte("pk2_shard0"), []byte("addr2_shard0"), defaultSelectionChances),
	}
	listShard1 := []Validator{
		mock.NewValidatorMock([]byte("pk0_shard1"), []byte("addr0_shard1"), defaultSelectionChances),
		mock.NewValidatorMock([]byte("pk1_shard1"), []byte("addr1_shard1"), defaultSelectionChances),
		mock.NewValidatorMock([]byte("pk2_shard1"), []byte("addr2_shard1"), defaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	nodeShuffler := NewXorValidatorsShuffler(3, 3, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	eligibleMap[core.MetachainShardId] = listMeta
	eligibleMap[0] = listShard0
	eligibleMap[1] = listShard1

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                2,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}
	nc, _ := NewIndexHashedNodesCoordinator(arguments)
	ihgs, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{})

	v, shardId, err := ihgs.GetValidatorWithPublicKey([]byte("pk0_meta"), 0)
	assert.Nil(t, err)
	assert.Equal(t, core.MetachainShardId, shardId)
	assert.Equal(t, []byte("addr0_meta"), v.Address())

	v, shardId, err = ihgs.GetValidatorWithPublicKey([]byte("pk1_shard0"), 0)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), shardId)
	assert.Equal(t, []byte("addr1_shard0"), v.Address())

	v, shardId, err = ihgs.GetValidatorWithPublicKey([]byte("pk2_shard1"), 0)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), shardId)
	assert.Equal(t, []byte("addr2_shard1"), v.Address())
}

func TestIndexHashedGroupSelectorWithRater_GetAllEligibleValidatorsPublicKeys(t *testing.T) {
	t.Parallel()

	shardZeroId := uint32(0)
	shardOneId := uint32(1)
	expectedValidatorsPubKeys := map[uint32][][]byte{
		shardZeroId:           {[]byte("pk0_shard0"), []byte("pk1_shard0"), []byte("pk2_shard0")},
		shardOneId:            {[]byte("pk0_shard1"), []byte("pk1_shard1"), []byte("pk2_shard1")},
		core.MetachainShardId: {[]byte("pk0_meta"), []byte("pk1_meta"), []byte("pk2_meta")},
	}

	listMeta := []Validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][0], []byte("addr0_meta"), defaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][1], []byte("addr1_meta"), defaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][2], []byte("addr2_meta"), defaultSelectionChances),
	}
	listShard0 := []Validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][0], []byte("addr0_shard0"), defaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][1], []byte("addr1_shard0"), defaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardZeroId][2], []byte("addr2_shard0"), defaultSelectionChances),
	}
	listShard1 := []Validator{
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][0], []byte("addr0_shard1"), defaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][1], []byte("addr1_shard1"), defaultSelectionChances),
		mock.NewValidatorMock(expectedValidatorsPubKeys[shardOneId][2], []byte("addr2_shard1"), defaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	nodeShuffler := NewXorValidatorsShuffler(3, 3, 0, false)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := mock.NewStorerMock()

	eligibleMap[core.MetachainShardId] = listMeta
	eligibleMap[shardZeroId] = listShard0
	eligibleMap[shardOneId] = listShard1

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartSubscriber:    epochStartSubscriber,
		BootStorer:              bootStorer,
		ShardIDAsObserver:       shardZeroId,
		NbShards:                2,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
	}

	nc, _ := NewIndexHashedNodesCoordinator(arguments)
	ihgs, err := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{})
	assert.Nil(t, err)

	allValidatorsPublicKeys, err := ihgs.GetAllEligibleValidatorsPublicKeys(0)
	assert.Nil(t, err)
	assert.Equal(t, expectedValidatorsPubKeys, allValidatorsPublicKeys)
}

func BenchmarkIndexHashedGroupSelectorWithRater_TestExpandList(b *testing.B) {
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)

	fmt.Println(m.HeapAlloc)

	nrNodes := 40000
	ratingSteps := 100
	array := make([]int, nrNodes*ratingSteps)
	for i := 0; i < nrNodes; i++ {
		for j := 0; j < ratingSteps; j++ {
			array[i*ratingSteps+j] = i
		}
	}

	//a := []int{1, 2, 3, 4, 5, 6, 7, 8}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(array), func(i, j int) { array[i], array[j] = array[j], array[i] })
	m2 := runtime.MemStats{}

	runtime.ReadMemStats(&m2)

	fmt.Println(m2.HeapAlloc)
	fmt.Println(fmt.Sprintf("Used %d MB", (m2.HeapAlloc-m.HeapAlloc)/1024/1024))
	//fmt.Print(array[0:100])
}

func BenchmarkIndexHashedGroupSelectorWithRater_TestHashes(b *testing.B) {
	nrElementsInList := int64(4000000)
	nrHashes := 100

	hasher := blake2b.Blake2b{}

	randomBits := ""

	for i := 0; i < nrHashes; i++ {
		randomBits = fmt.Sprintf("%s%d", randomBits, rand.Intn(2))
	}
	//computedListIndex := int64(0)
	for i := 0; i < nrHashes; i++ {
		computedHash := hasher.Compute(randomBits + fmt.Sprintf("%d", i))
		computedLargeIndex := big.NewInt(0)
		computedLargeIndex.SetBytes(computedHash)
		fmt.Println(big.NewInt(0).Mod(computedLargeIndex, big.NewInt(nrElementsInList)).Int64())
	}

	//fmt.Print(array[0:100])
}

func BenchmarkIndexHashedWithRaterGroupSelector_ComputeValidatorsGroup21of400(b *testing.B) {
	consensusGroupSize := 21
	list := make([]Validator, 0)

	//generate 400 validators
	for i := 0; i < 400; i++ {
		list = append(list, mock.NewValidatorMock([]byte("pk"+strconv.Itoa(i)), []byte("addr"+strconv.Itoa(i)), defaultSelectionChances))
	}

	listMeta := []Validator{
		mock.NewValidatorMock([]byte("pkMeta1"), []byte("addrMeta1"), defaultSelectionChances),
		mock.NewValidatorMock([]byte("pkMeta2"), []byte("addrMeta2"), defaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	eligibleMap[core.MetachainShardId] = listMeta
	nodeShuffler := NewXorValidatorsShuffler(400, 1, 0, false)
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
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(b, err)
	ihgsRater, err := NewIndexHashedNodesCoordinatorWithRater(ihgs, &mock.RaterMock{})
	require.Nil(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomness := strconv.Itoa(i)
		list2, _ := ihgsRater.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

		assert.Equal(b, consensusGroupSize, len(list2))
	}
}

func TestIndexHashedGroupSelectorWithRater_ComputeLeavingShouldWork(t *testing.T) {
	t.Parallel()
	pkLeaving := "leavingPK"
	pkNotLeaving := "notLeavingPK"
	nc, _ := NewIndexHashedNodesCoordinator(createArguments())
	ihgs, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{
		GetRatingCalled: func(s string) uint32 {
			if s == pkLeaving {
				return 1
			}
			return 100
		},
		GetChancesCalled: func(u uint32) uint32 {
			if u == 1 {
				return 0
			}
			return 10
		},
	})

	leavingValidator := mock.NewValidatorMock([]byte(pkLeaving), []byte(pkLeaving), defaultSelectionChances)
	notLeavingValidator := mock.NewValidatorMock([]byte(pkNotLeaving), []byte(pkNotLeaving), defaultSelectionChances)
	validators := []Validator{leavingValidator, notLeavingValidator}
	leavingValidators := ihgs.ComputeLeaving(validators)

	require.Equal(t, 1, len(leavingValidators))
	require.Equal(t, leavingValidator, leavingValidators[0])
}

func TestIndexHashedGroupSelectorWithRater_ComputeLeavingWithAllChancesBigShouldNotRemove(t *testing.T) {
	t.Parallel()
	pkLeaving := "leavingPK"
	pkNotLeaving := "notLeavingPK"
	nc, _ := NewIndexHashedNodesCoordinator(createArguments())
	ihgs, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{
		GetRatingCalled: func(s string) uint32 {
			return 100
		},
		GetChancesCalled: func(u uint32) uint32 {
			return 10
		},
	})

	leavingValidator := mock.NewValidatorMock([]byte(pkLeaving), []byte(pkLeaving), defaultSelectionChances)
	notLeavingValidator := mock.NewValidatorMock([]byte(pkNotLeaving), []byte(pkNotLeaving), defaultSelectionChances)
	validators := []Validator{leavingValidator, notLeavingValidator}
	leavingValidators := ihgs.ComputeLeaving(validators)

	require.Equal(t, 0, len(leavingValidators))
}
