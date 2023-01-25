package nodesCoordinator

import (
	"fmt"
	"math/big"
	"math/rand"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/endProcess"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/hashing/sha256"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/sharding/mock"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	vic "github.com/multiversx/mx-chain-go/testscommon/validatorInfoCacher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIndexHashedNodesCoordinatorWithRater_NilRaterShouldErr(t *testing.T) {
	nc, _ := NewIndexHashedNodesCoordinator(createArguments())
	ihnc, err := NewIndexHashedNodesCoordinatorWithRater(nc, nil)

	assert.Nil(t, ihnc)
	assert.Equal(t, ErrNilChanceComputer, err)
}

func TestNewIndexHashedNodesCoordinatorWithRater_NilNodesCoordinatorShouldErr(t *testing.T) {
	ihnc, err := NewIndexHashedNodesCoordinatorWithRater(nil, &mock.RaterMock{})

	assert.Nil(t, ihnc)
	assert.Equal(t, ErrNilNodesCoordinator, err)
}

func TestNewIndexHashedGroupSelectorWithRater_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	nc, _ := NewIndexHashedNodesCoordinator(createArguments())
	ihnc, err := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{})
	assert.NotNil(t, ihnc)
	assert.Nil(t, err)
}

//------- LoadEligibleList

func TestIndexHashedGroupSelectorWithRater_SetNilEligibleMapShouldErr(t *testing.T) {
	t.Parallel()
	waiting := createDummyNodesMap(2, 1, "waiting")
	nc, _ := NewIndexHashedNodesCoordinator(createArguments())
	ihnc, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{})
	assert.Equal(t, ErrNilInputNodesMap, ihnc.setNodesPerShards(nil, waiting, nil, 0))
}

func TestIndexHashedGroupSelectorWithRater_OkValShouldWork(t *testing.T) {
	t.Parallel()

	eligibleMap := createDummyNodesMap(3, 1, "waiting")
	waitingMap := make(map[uint32][]Validator)
	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           3,
		NodesMeta:            3,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler:  &mock.EnableEpochsHandlerMock{},
	}
	nodeShuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := genericMocks.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 2,
		MetaConsensusGroupSize:  1,
		Marshalizer:             &mock.MarshalizerMock{},
		Hasher:                  &hashingMocks.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("test"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		ChanStopNode:            make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler:     &mock.EnableEpochsHandlerMock{},
		ValidatorInfoCacher:     &vic.ValidatorInfoCacherStub{},
	}
	nc, err := NewIndexHashedNodesCoordinator(arguments)
	assert.Nil(t, err)
	readEligible := nc.nodesConfig[0].eligibleMap[0]
	assert.Equal(t, eligibleMap[0], readEligible)

	rater := &mock.RaterMock{}
	ihnc, err := NewIndexHashedNodesCoordinatorWithRater(nc, rater)
	assert.Nil(t, err)

	readEligible = ihnc.nodesConfig[0].eligibleMap[0]
	assert.Equal(t, eligibleMap[0], readEligible)
}

//------- functionality tests

func TestIndexHashedGroupSelectorWithRater_ComputeValidatorsGroup1ValidatorShouldNotCallGetRating(t *testing.T) {
	t.Parallel()

	list := []Validator{
		newValidatorMock([]byte("pk0"), 1, defaultSelectionChances),
	}

	arguments := createArguments()
	arguments.EligibleNodes[0] = list

	chancesCalled := false
	rater := &mock.RaterMock{
		GetChancesCalled: func(u uint32) uint32 {
			chancesCalled = true
			return 1
		}}

	nc, err := NewIndexHashedNodesCoordinator(arguments)
	assert.Nil(t, err)
	assert.Equal(t, false, chancesCalled)
	ihnc, _ := NewIndexHashedNodesCoordinatorWithRater(nc, rater)
	assert.Equal(t, true, chancesCalled)
	list2, err := ihnc.ComputeConsensusGroup([]byte("randomness"), 0, 0, 0)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(list2))
}

func BenchmarkIndexHashedGroupSelectorWithRater_ComputeValidatorsGroup63of400(b *testing.B) {
	b.ReportAllocs()

	consensusGroupSize := 63
	list := make([]Validator, 0)

	//generate 400 validators
	for i := 0; i < 400; i++ {
		list = append(list, newValidatorMock([]byte("pk"+strconv.Itoa(i)), 1, defaultSelectionChances))
	}
	listMeta := []Validator{
		newValidatorMock([]byte("pkMeta1"), 1, defaultSelectionChances),
		newValidatorMock([]byte("pkMeta2"), 1, defaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	eligibleMap[core.MetachainShardId] = listMeta

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           400,
		NodesMeta:            1,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler:  &mock.EnableEpochsHandlerMock{},
	}
	nodeShuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(b, err)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := genericMocks.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: consensusGroupSize,
		MetaConsensusGroupSize:  1,
		Marshalizer:             &mock.MarshalizerMock{},
		Hasher:                  &hashingMocks.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
		ChanStopNode:            make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler:     &mock.EnableEpochsHandlerMock{},
		ValidatorInfoCacher:     &vic.ValidatorInfoCacherStub{},
	}
	ihnc, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(b, err)
	ihncRater, err := NewIndexHashedNodesCoordinatorWithRater(ihnc, &mock.RaterMock{})
	require.Nil(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomness := strconv.Itoa(0)
		list2, _ := ihncRater.ComputeConsensusGroup([]byte(randomness), uint64(0), 0, 0)

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
		list = append(list, newValidatorMock([]byte(fmt.Sprintf("pk%v", i)), 1, defaultSelectionChances))
	}
	listMeta := []Validator{
		newValidatorMock([]byte("pkMeta1"), 1, defaultSelectionChances),
		newValidatorMock([]byte("pkMeta2"), 1, defaultSelectionChances),
	}

	consensusAppearances := make(map[string]uint64)
	leaderAppearances := make(map[string]uint64)
	for _, v := range list {
		consensusAppearances[string(v.PubKey())] = 0
		leaderAppearances[string(v.PubKey())] = 0
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	eligibleMap[core.MetachainShardId] = listMeta
	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           shardSize,
		NodesMeta:            1,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler:  &mock.EnableEpochsHandlerMock{},
	}
	nodeShuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := genericMocks.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: consensusGroupSize,
		MetaConsensusGroupSize:  1,
		Hasher:                  &hashingMocks.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
		ChanStopNode:            make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler:     &mock.EnableEpochsHandlerMock{},
		ValidatorInfoCacher:     &vic.ValidatorInfoCacherStub{},
	}
	ihnc, _ := NewIndexHashedNodesCoordinator(arguments)
	numRounds := uint64(1000000)
	hasher := sha256.NewSha256()
	for i := uint64(0); i < numRounds; i++ {
		randomness := hasher.Compute(fmt.Sprintf("%v%v", i, time.Millisecond))
		consensusGroup, _ := ihnc.ComputeConsensusGroup(randomness, uint64(0), 0, 0)
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
		newValidatorMock([]byte("pk0"), 1, defaultSelectionChances),
	}
	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	eligibleMap[core.MetachainShardId] = list
	sufflerArgs := &NodesShufflerArgs{
		NodesShard:           1,
		NodesMeta:            1,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler:  &mock.EnableEpochsHandlerMock{},
	}
	nodeShuffler, err := NewHashValidatorsShuffler(sufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := genericMocks.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Marshalizer:             &mock.MarshalizerMock{},
		Hasher:                  &hashingMocks.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		ChanStopNode:            make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler:     &mock.EnableEpochsHandlerMock{},
		ValidatorInfoCacher:     &vic.ValidatorInfoCacherStub{},
	}
	nc, _ := NewIndexHashedNodesCoordinator(arguments)
	ihnc, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{})

	_, _, err = ihnc.GetValidatorWithPublicKey(nil)
	assert.Equal(t, ErrNilPubKey, err)
}

func TestIndexHashedGroupSelectorWithRater_GetValidatorWithPublicKeyShouldReturnErrValidatorNotFound(t *testing.T) {
	t.Parallel()

	list := []Validator{
		newValidatorMock([]byte("pk0"), 1, defaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	eligibleMap[core.MetachainShardId] = list

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           1,
		NodesMeta:            1,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler:  &mock.EnableEpochsHandlerMock{},
	}
	nodeShuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := genericMocks.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Marshalizer:             &mock.MarshalizerMock{},
		Hasher:                  &hashingMocks.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		ChanStopNode:            make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler:     &mock.EnableEpochsHandlerMock{},
		ValidatorInfoCacher:     &vic.ValidatorInfoCacherStub{},
	}
	nc, _ := NewIndexHashedNodesCoordinator(arguments)
	ihnc, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{})

	_, _, err = ihnc.GetValidatorWithPublicKey([]byte("pk1"))
	assert.Equal(t, ErrValidatorNotFound, err)
}

func TestIndexHashedGroupSelectorWithRater_GetValidatorWithPublicKeyShouldWork(t *testing.T) {
	t.Parallel()

	listMeta := []Validator{
		newValidatorMock([]byte("pk0_meta"), 1, defaultSelectionChances),
		newValidatorMock([]byte("pk1_meta"), 1, defaultSelectionChances),
		newValidatorMock([]byte("pk2_meta"), 1, defaultSelectionChances),
	}
	listShard0 := []Validator{
		newValidatorMock([]byte("pk0_shard0"), 1, defaultSelectionChances),
		newValidatorMock([]byte("pk1_shard0"), 1, defaultSelectionChances),
		newValidatorMock([]byte("pk2_shard0"), 1, defaultSelectionChances),
	}
	listShard1 := []Validator{
		newValidatorMock([]byte("pk0_shard1"), 1, defaultSelectionChances),
		newValidatorMock([]byte("pk1_shard1"), 1, defaultSelectionChances),
		newValidatorMock([]byte("pk2_shard1"), 1, defaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           3,
		NodesMeta:            3,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler:  &mock.EnableEpochsHandlerMock{},
	}
	nodeShuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := genericMocks.NewStorerMock()

	eligibleMap[core.MetachainShardId] = listMeta
	eligibleMap[0] = listShard0
	eligibleMap[1] = listShard1

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Marshalizer:             &mock.MarshalizerMock{},
		Hasher:                  &hashingMocks.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                2,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		ChanStopNode:            make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler:     &mock.EnableEpochsHandlerMock{},
		ValidatorInfoCacher:     &vic.ValidatorInfoCacherStub{},
	}
	nc, _ := NewIndexHashedNodesCoordinator(arguments)
	ihnc, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{})

	_, shardId, err := ihnc.GetValidatorWithPublicKey([]byte("pk0_meta"))
	assert.Nil(t, err)
	assert.Equal(t, core.MetachainShardId, shardId)

	_, shardId, err = ihnc.GetValidatorWithPublicKey([]byte("pk1_shard0"))
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), shardId)

	_, shardId, err = ihnc.GetValidatorWithPublicKey([]byte("pk2_shard1"))
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), shardId)
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
		newValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][0], 1, defaultSelectionChances),
		newValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][1], 1, defaultSelectionChances),
		newValidatorMock(expectedValidatorsPubKeys[core.MetachainShardId][2], 1, defaultSelectionChances),
	}
	listShard0 := []Validator{
		newValidatorMock(expectedValidatorsPubKeys[shardZeroId][0], 1, defaultSelectionChances),
		newValidatorMock(expectedValidatorsPubKeys[shardZeroId][1], 1, defaultSelectionChances),
		newValidatorMock(expectedValidatorsPubKeys[shardZeroId][2], 1, defaultSelectionChances),
	}
	listShard1 := []Validator{
		newValidatorMock(expectedValidatorsPubKeys[shardOneId][0], 1, defaultSelectionChances),
		newValidatorMock(expectedValidatorsPubKeys[shardOneId][1], 1, defaultSelectionChances),
		newValidatorMock(expectedValidatorsPubKeys[shardOneId][2], 1, defaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           3,
		NodesMeta:            3,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler:  &mock.EnableEpochsHandlerMock{},
	}
	nodeShuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)
	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := genericMocks.NewStorerMock()

	eligibleMap[core.MetachainShardId] = listMeta
	eligibleMap[shardZeroId] = listShard0
	eligibleMap[shardOneId] = listShard1

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Marshalizer:             &mock.MarshalizerMock{},
		Hasher:                  &hashingMocks.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      epochStartSubscriber,
		BootStorer:              bootStorer,
		ShardIDAsObserver:       shardZeroId,
		NbShards:                2,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		ChanStopNode:            make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler:     &mock.EnableEpochsHandlerMock{},
		ValidatorInfoCacher:     &vic.ValidatorInfoCacherStub{},
	}

	nc, _ := NewIndexHashedNodesCoordinator(arguments)
	ihnc, err := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{})
	assert.Nil(t, err)

	allValidatorsPublicKeys, err := ihnc.GetAllEligibleValidatorsPublicKeys(0)
	assert.Nil(t, err)
	assert.Equal(t, expectedValidatorsPubKeys, allValidatorsPublicKeys)
}

func TestIndexHashedGroupSelectorWithRater_ComputeAdditionalLeaving(t *testing.T) {
	t.Parallel()

	minChances := uint32(5)
	belowRatingThresholdChances := uint32(1)
	nc, _ := NewIndexHashedNodesCoordinator(createArguments())
	ihnc, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{
		GetChancesCalled: func(rating uint32) uint32 {
			if rating == 0 {
				return minChances
			}
			if rating < 10 {
				return belowRatingThresholdChances
			}
			return 10
		},
	})

	leavingValidator := &state.ShardValidatorInfo{
		PublicKey:  []byte("eligible"),
		ShardId:    core.MetachainShardId,
		List:       string(common.EligibleList),
		Index:      7,
		TempRating: 5,
	}

	shardValidatorInfo := []*state.ShardValidatorInfo{
		leavingValidator,
	}

	additionalLeaving, err := ihnc.ComputeAdditionalLeaving(shardValidatorInfo)
	assert.NotNil(t, additionalLeaving)
	assert.Nil(t, err)

	found, shardId := searchInMap(additionalLeaving, leavingValidator.PublicKey)
	assert.True(t, found)
	assert.Equal(t, leavingValidator.ShardId, shardId)

	val := additionalLeaving[shardId][0]
	assert.Equal(t, leavingValidator.PublicKey, val.PubKey())
	assert.Equal(t, belowRatingThresholdChances, val.Chances())
	assert.Equal(t, leavingValidator.Index, val.Index())
}

func TestIndexHashedGroupSelectorWithRater_ComputeAdditionalLeaving_ShouldAddNewEligibleWaiting(t *testing.T) {
	t.Parallel()

	minChances := uint32(5)

	nc, _ := NewIndexHashedNodesCoordinator(createArguments())
	ihnc, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{
		GetChancesCalled: func(rating uint32) uint32 {
			if rating == 0 {
				return minChances
			}
			return 0
		},
	})

	newValidator := &state.ShardValidatorInfo{
		PublicKey:  []byte("new"),
		ShardId:    0,
		List:       string(common.NewList),
		Index:      1,
		TempRating: 5,
	}
	eligibleValidator := &state.ShardValidatorInfo{
		PublicKey:  []byte("eligible"),
		ShardId:    core.MetachainShardId,
		List:       string(common.EligibleList),
		Index:      1,
		TempRating: 5,
	}
	waitingValidator := &state.ShardValidatorInfo{
		PublicKey:  []byte("waiting"),
		ShardId:    1,
		List:       string(common.WaitingList),
		Index:      1,
		TempRating: 5,
	}

	shardValidatorInfo := []*state.ShardValidatorInfo{
		newValidator,
		eligibleValidator,
		waitingValidator,
	}

	additionalLeaving, err := ihnc.ComputeAdditionalLeaving(shardValidatorInfo)
	assert.NotNil(t, additionalLeaving)
	assert.Nil(t, err)

	foundNew, _ := searchInMap(additionalLeaving, newValidator.PublicKey)
	assert.True(t, foundNew)

	foundEligible, _ := searchInMap(additionalLeaving, eligibleValidator.PublicKey)
	assert.True(t, foundEligible)

	foundWaiting, _ := searchInMap(additionalLeaving, waitingValidator.PublicKey)
	assert.True(t, foundWaiting)
}

func TestIndexHashedGroupSelectorWithRater_ComputeAdditionalLeaving_ShouldNotAddInactiveAndJailed(t *testing.T) {
	t.Parallel()

	minChances := uint32(5)

	nc, _ := NewIndexHashedNodesCoordinator(createArguments())
	ihnc, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{
		GetChancesCalled: func(rating uint32) uint32 {
			if rating == 0 {
				return minChances
			}
			return 0
		},
	})

	inactiveValidator := &state.ShardValidatorInfo{
		PublicKey:  []byte("inactive"),
		ShardId:    0,
		List:       string(common.InactiveList),
		Index:      1,
		TempRating: 5,
	}
	jailedValidator := &state.ShardValidatorInfo{
		PublicKey:  []byte("jailed"),
		ShardId:    core.MetachainShardId,
		List:       string(common.JailedList),
		Index:      1,
		TempRating: 5,
	}

	shardValidatorInfo := []*state.ShardValidatorInfo{
		inactiveValidator,
		jailedValidator,
	}

	additionalLeaving, err := ihnc.ComputeAdditionalLeaving(shardValidatorInfo)
	assert.NotNil(t, additionalLeaving)
	assert.Nil(t, err)

	foundInactive, _ := searchInMap(additionalLeaving, inactiveValidator.PublicKey)
	assert.False(t, foundInactive)

	foundJailed, _ := searchInMap(additionalLeaving, jailedValidator.PublicKey)
	assert.False(t, foundJailed)
}

func TestIndexHashedGroupSelectorWithRater_ComputeAdditionalLeaving_ShouldAddBelowMinRating(t *testing.T) {
	t.Parallel()

	minRating := uint32(10)
	minChances := uint32(5)

	nc, _ := NewIndexHashedNodesCoordinator(createArguments())
	ihnc, _ := NewIndexHashedNodesCoordinatorWithRater(nc, &mock.RaterMock{
		GetChancesCalled: func(rating uint32) uint32 {
			if rating == 0 {
				return minChances
			}
			if rating < minRating {
				return 0
			}
			return 10
		},
	})

	eligibleValidator := &state.ShardValidatorInfo{
		PublicKey:  []byte("eligible"),
		ShardId:    0,
		List:       string(common.EligibleList),
		Index:      1,
		TempRating: 50,
	}
	belowRatingValidator := &state.ShardValidatorInfo{
		PublicKey:  []byte("eligibleBelow"),
		ShardId:    core.MetachainShardId,
		List:       string(common.EligibleList),
		Index:      1,
		TempRating: 5,
	}

	shardValidatorInfo := []*state.ShardValidatorInfo{
		eligibleValidator,
		belowRatingValidator,
	}

	additionalLeaving, err := ihnc.ComputeAdditionalLeaving(shardValidatorInfo)
	assert.NotNil(t, additionalLeaving)
	assert.Nil(t, err)

	foundEligible, _ := searchInMap(additionalLeaving, eligibleValidator.PublicKey)
	assert.False(t, foundEligible)

	foundBelowRatingValidator, _ := searchInMap(additionalLeaving, belowRatingValidator.PublicKey)
	assert.True(t, foundBelowRatingValidator)
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
	fmt.Printf("Used %d MB\n", (m2.HeapAlloc-m.HeapAlloc)/1024/1024)
}

func BenchmarkIndexHashedGroupSelectorWithRater_TestHashes(b *testing.B) {
	nrElementsInList := int64(4000000)
	nrHashes := 100

	hasher := blake2b.NewBlake2b()

	randomBits := ""

	for i := 0; i < nrHashes; i++ {
		randomBits = fmt.Sprintf("%s%d", randomBits, rand.Intn(2))
	}

	for i := 0; i < nrHashes; i++ {
		computedHash := hasher.Compute(randomBits + fmt.Sprintf("%d", i))
		computedLargeIndex := big.NewInt(0)
		computedLargeIndex.SetBytes(computedHash)
		fmt.Println(big.NewInt(0).Mod(computedLargeIndex, big.NewInt(nrElementsInList)).Int64())
	}
}

func BenchmarkIndexHashedWithRaterGroupSelector_ComputeValidatorsGroup21of400(b *testing.B) {
	consensusGroupSize := 21
	list := make([]Validator, 0)

	//generate 400 validators
	for i := 0; i < 400; i++ {
		list = append(list, newValidatorMock([]byte("pk"+strconv.Itoa(i)), 1, defaultSelectionChances))
	}

	listMeta := []Validator{
		newValidatorMock([]byte("pkMeta1"), 1, defaultSelectionChances),
		newValidatorMock([]byte("pkMeta2"), 1, defaultSelectionChances),
	}

	eligibleMap := make(map[uint32][]Validator)
	waitingMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	eligibleMap[core.MetachainShardId] = listMeta
	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           400,
		NodesMeta:            1,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
		EnableEpochsHandler:  &mock.EnableEpochsHandlerMock{},
	}
	nodeShuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(b, err)

	epochStartSubscriber := &mock.EpochStartNotifierStub{}
	bootStorer := genericMocks.NewStorerMock()

	arguments := ArgNodesCoordinator{
		ShardConsensusGroupSize: consensusGroupSize,
		MetaConsensusGroupSize:  1,
		Marshalizer:             &mock.MarshalizerMock{},
		Hasher:                  &hashingMocks.HasherMock{},
		Shuffler:                nodeShuffler,
		EpochStartNotifier:      epochStartSubscriber,
		BootStorer:              bootStorer,
		NbShards:                1,
		EligibleNodes:           eligibleMap,
		WaitingNodes:            waitingMap,
		SelfPublicKey:           []byte("key"),
		ConsensusGroupCache:     &mock.NodesCoordinatorCacheMock{},
		ShuffledOutHandler:      &mock.ShuffledOutHandlerStub{},
		ChanStopNode:            make(chan endProcess.ArgEndProcess),
		NodeTypeProvider:        &nodeTypeProviderMock.NodeTypeProviderStub{},
		IsFullArchive:           false,
		EnableEpochsHandler:     &mock.EnableEpochsHandlerMock{},
		ValidatorInfoCacher:     &vic.ValidatorInfoCacherStub{},
	}
	ihnc, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(b, err)
	ihncRater, err := NewIndexHashedNodesCoordinatorWithRater(ihnc, &mock.RaterMock{})
	require.Nil(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomness := strconv.Itoa(i)
		list2, _ := ihncRater.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

		assert.Equal(b, consensusGroupSize, len(list2))
	}
}
