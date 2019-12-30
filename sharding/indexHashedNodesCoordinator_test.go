package sharding

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/stretchr/testify/assert"
)

func convertBigIntToBytes(value *big.Int) []byte {
	return value.Bytes()
}

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
	}
	return arguments
}

func genRandSource(round uint64, randomness string) string {
	return fmt.Sprintf("%d-%s", round, core.ToB64([]byte(randomness)))
}

//------- NewIndexHashedNodesCoordinator

func TestNewIndexHashedGroupSelector_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.Hasher = nil
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, ErrNilHasher, err)
}

func TestNewIndexHashedGroupSelector_InvalidConsensusGroupSizeShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.ShardConsensusGroupSize = 0
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, ErrInvalidConsensusGroupSize, err)
}

func TestNewIndexHashedNodesCoordinator_ZeroNbShardsShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.NbShards = 0
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, ErrInvalidNumberOfShards, err)
}

func TestNewIndexHashedNodesCoordinator_InvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.ShardId = 10
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, ErrInvalidShardId, err)
}

func TestNewIndexHashedNodesCoordinator_NilSelfPublicKeyShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.SelfPublicKey = nil
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, ErrNilPubKey, err)
}

func TestNewIndexHashedGroupSelector_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	assert.NotNil(t, ihgs)
	assert.Nil(t, err)
}

//------- LoadEligibleList

func TestIndexHashedGroupSelector_SetNilEligibleMapShouldErr(t *testing.T) {
	t.Parallel()

	waitingMap := createDummyNodesMap(3, 3, "waiting")
	arguments := createArguments()

	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
	assert.Equal(t, ErrNilInputNodesMap, ihgs.SetNodesPerShards(nil, waitingMap, 0))
}

func TestIndexHashedGroupSelector_SetNilWaitingMapShouldErr(t *testing.T) {
	t.Parallel()

	eligibleMap := createDummyNodesMap(10, 3, "eligible")
	arguments := createArguments()

	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
	assert.Equal(t, ErrNilInputNodesMap, ihgs.SetNodesPerShards(eligibleMap, nil, 0))
}

func TestIndexHashedGroupSelector_OkValShouldWork(t *testing.T) {
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
	}

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	assert.Nil(t, err)

	readEligible := ihgs.nodesConfig[arguments.Epoch].eligibleMap[0]
	assert.Equal(t, eligibleMap[0], readEligible)
}

//------- ComputeValidatorsGroup

func TestIndexHashedGroupSelector_NewCoordinatorGroup0SizeShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	arguments.MetaConsensusGroupSize = 0
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, ErrInvalidConsensusGroupSize, err)
}

func TestIndexHashedGroupSelector_NewCoordinatorTooFewNodesShouldErr(t *testing.T) {
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
	}
	ihgs, err := NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, ErrSmallShardEligibleListSize, err)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupNilRandomnessShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
	list2, err := ihgs.ComputeConsensusGroup(nil, 0, 0, 0)

	assert.Nil(t, list2)
	assert.Equal(t, ErrNilRandomness, err)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupInvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
	list2, err := ihgs.ComputeConsensusGroup([]byte("radomness"), 0, 5, 0)

	assert.Nil(t, list2)
	assert.Equal(t, ErrInvalidShardId, err)
}

//------- functionality tests

func TestIndexHashedGroupSelector_ComputeValidatorsGroup1ValidatorShouldReturnSame(t *testing.T) {
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
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
	list2, err := ihgs.ComputeConsensusGroup([]byte("randomness"), 0, 0, 0)

	assert.Nil(t, err)
	assert.Equal(t, list, list2)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupTest2Validators(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherStub{}

	randomness := "randomness"

	//this will return the list in order:
	//element 0 will be first element
	//element 1 will be the second
	hasher.ComputeCalled = func(s string) []byte {
		if string(uint64ToBytes(0))+randomness == s {
			return convertBigIntToBytes(big.NewInt(0))
		}

		if string(uint64ToBytes(1))+randomness == s {
			return convertBigIntToBytes(big.NewInt(1))
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
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	list2, err := ihgs.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

	assert.Nil(t, err)
	assert.Equal(t, eligibleMap[0][:2], list2)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupTest2ValidatorsRevertOrder(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherStub{}

	randomness := "randomness"
	randSource := genRandSource(0, randomness)

	//this will return the list in reverse order:
	//element 0 will be the second
	//element 1 will be the first
	hasher.ComputeCalled = func(s string) []byte {
		if string(uint64ToBytes(0))+randSource == s {
			return convertBigIntToBytes(big.NewInt(1))
		}

		if string(uint64ToBytes(1))+randSource == s {
			return convertBigIntToBytes(big.NewInt(0))
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
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	list2, err := ihgs.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

	assert.Nil(t, err)
	assert.Equal(t, validator0, list2[1])
	assert.Equal(t, validator1, list2[0])
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupTest2ValidatorsSameIndex(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherStub{}

	randomness := "randomness"

	//this will return the list in order:
	//element 0 will be the first
	//element 1 will be the second as the same index is being returned and 0 is already in list
	hasher.ComputeCalled = func(s string) []byte {
		if string(uint64ToBytes(0))+randomness == s {
			return convertBigIntToBytes(big.NewInt(0))
		}

		if string(uint64ToBytes(1))+randomness == s {
			return convertBigIntToBytes(big.NewInt(0))
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
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	list2, err := ihgs.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

	assert.Nil(t, err)
	assert.Equal(t, eligibleMap[0][:2], list2)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupTest6From10ValidatorsShouldWork(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherStub{}

	randomness := "randomness"
	randomnessWithRound := genRandSource(0, randomness)

	//script:
	// for index 0, hasher will return 11 which will translate to 1, so 1 is the first element
	// for index 1, hasher will return 1 which will translate to 1, 1 is already picked, try the next, 2 is the second element
	// for index 2, hasher will return 9 which will translate to 9, 9 is the 3-rd element
	// for index 3, hasher will return 9 which will translate to 9, 9 is already picked, try the next one, 0 is the 4-th element
	// for index 4, hasher will return 0 which will translate to 0, 0 is already picked, 1 is already picked, 2 is already picked,
	//      3 is the 4-th element
	// for index 5, hasher will return 9 which will translate to 9, so 9, 0, 1, 2, 3 are already picked, 4 is the 5-th element
	script := make(map[string]*big.Int)

	script[string(uint64ToBytes(0))+randomnessWithRound] = big.NewInt(11) //will translate to 1, add 1
	script[string(uint64ToBytes(1))+randomnessWithRound] = big.NewInt(1)  //will translate to 1, add 2
	script[string(uint64ToBytes(2))+randomnessWithRound] = big.NewInt(9)  //will translate to 9, add 9
	script[string(uint64ToBytes(3))+randomnessWithRound] = big.NewInt(9)  //will translate to 9, add 0
	script[string(uint64ToBytes(4))+randomnessWithRound] = big.NewInt(0)  //will translate to 0, add 3
	script[string(uint64ToBytes(5))+randomnessWithRound] = big.NewInt(9)  //will translate to 9, add 4

	hasher.ComputeCalled = func(s string) []byte {
		val, ok := script[s]

		if !ok {
			assert.Fail(t, "should have not got here")
		}

		return convertBigIntToBytes(val)
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
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	list2, err := ihgs.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

	assert.Nil(t, err)
	assert.Equal(t, 6, len(list2))
	//check order as described in script
	assert.Equal(t, validator1, list2[0])
	assert.Equal(t, validator2, list2[1])
	assert.Equal(t, validator9, list2[2])
	assert.Equal(t, validator0, list2[3])
	assert.Equal(t, validator3, list2[4])
	assert.Equal(t, validator4, list2[5])
}

func BenchmarkIndexHashedGroupSelector_ComputeValidatorsGroup21of400(b *testing.B) {
	consensusGroupSize := 21
	list := make([]Validator, 0)

	//generate 400 validators
	for i := 0; i < 400; i++ {
		list = append(list, mock.NewValidatorMock([]byte("pk"+strconv.Itoa(i)), []byte("addr"+strconv.Itoa(i))))
	}

	eligibleMap := make(map[uint32][]Validator)
	eligibleMap[0] = list
	nodeShuffler := NewXorValidatorsShuffler(10, 10, 0, false)
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
		WaitingNodes:            make(map[uint32][]Validator),
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomness := strconv.Itoa(i)
		list2, _ := ihgs.ComputeConsensusGroup([]byte(randomness), 0, 0, 0)

		assert.Equal(b, consensusGroupSize, len(list2))
	}
}

func TestIndexHashedGroupSelector_GetValidatorWithPublicKeyShouldReturnErrNilPubKey(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	_, _, err := ihgs.GetValidatorWithPublicKey(nil, 0)
	assert.Equal(t, ErrNilPubKey, err)
}

func TestIndexHashedGroupSelector_GetValidatorWithPublicKeyShouldReturnErrValidatorNotFound(t *testing.T) {
	t.Parallel()

	arguments := createArguments()
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	_, _, err := ihgs.GetValidatorWithPublicKey([]byte("pk1"), 0)
	assert.Equal(t, ErrValidatorNotFound, err)
}

func TestIndexHashedGroupSelector_GetValidatorWithPublicKeyShouldWork(t *testing.T) {
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
	}
	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	validator, shardId, err := ihgs.GetValidatorWithPublicKey([]byte("pk0_meta"), 0)
	assert.Nil(t, err)
	assert.Equal(t, core.MetachainShardId, shardId)
	assert.Equal(t, []byte("addr0_meta"), validator.Address())

	validator, shardId, err = ihgs.GetValidatorWithPublicKey([]byte("pk1_shard0"), 0)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), shardId)
	assert.Equal(t, []byte("addr1_shard0"), validator.Address())

	validator, shardId, err = ihgs.GetValidatorWithPublicKey([]byte("pk2_shard1"), 0)
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), shardId)
	assert.Equal(t, []byte("addr2_shard1"), validator.Address())
}

func TestIndexHashedGroupSelector_GetAllValidatorsPublicKeys(t *testing.T) {
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
	}

	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)

	allValidatorsPublicKeys, err := ihgs.GetAllValidatorsPublicKeys(0)
	assert.Equal(t, expectedValidatorsPubKeys, allValidatorsPublicKeys)
	assert.Nil(t, err)
}

func TestIndexHashedGroupSelector_EpochStart(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	ihgs, _ := NewIndexHashedNodesCoordinator(arguments)
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
	assert.Nil(t, err)
	assert.NotNil(t, validators)

	computedShardId := ihgs.computeShardForPublicKey(ihgs.nodesConfig[0])
	// should remain in same shard with intra shard shuffling
	assert.Equal(t, arguments.ShardId, computedShardId)
}
