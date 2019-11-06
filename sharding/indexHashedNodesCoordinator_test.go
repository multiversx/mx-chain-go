package sharding_test

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"math/rand"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go/sharding"
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

func createDummyNodesMap() map[uint32][]sharding.Validator {
	list := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0"), []byte("addr0")),
		mock.NewValidatorMock(big.NewInt(2), 3, []byte("pk1"), []byte("addr1")),
	}

	listMeta := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 1, []byte("pkMeta1"), []byte("addrMeta1")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pkMeta2"), []byte("addrMeta2")),
	}

	nodesMap := make(map[uint32][]sharding.Validator)
	nodesMap[0] = list
	nodesMap[sharding.MetachainShardId] = listMeta

	return nodesMap
}

func genRandSource(round uint64, randomness string) string {
	return fmt.Sprintf("%d-%s", round, core.ToB64([]byte(randomness)))
}

//------- NewIndexHashedNodesCoordinator

func TestNewIndexHashedGroupSelector_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()

	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}

	ihgs, err := sharding.NewIndexHashedNodesCoordinator(arguments)
	assert.Nil(t, ihgs)
	assert.Equal(t, sharding.ErrNilHasher, err)
}

func TestNewIndexHashedGroupSelector_InvalidConsensusGroupSizeShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		MetaConsensusGroupSize: 1,
		Hasher:                 &mock.HasherMock{},
		NbShards:               1,
		Nodes:                  nodesMap,
		SelfPublicKey:          []byte("key"),
	}
	ihgs, err := sharding.NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, sharding.ErrInvalidConsensusGroupSize, err)
}

func TestNewIndexHashedNodesCoordinator_ZeroNbShardsShouldErr(t *testing.T) {
	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                0,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, err := sharding.NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, sharding.ErrInvalidNumberOfShards, err)
}

func TestNewIndexHashedNodesCoordinator_InvalidShardIdShouldErr(t *testing.T) {
	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		ShardId:                 2,
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, err := sharding.NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, sharding.ErrInvalidShardId, err)
}

func TestNewIndexHashedNodesCoordinator_NilSelfPublicKeyShouldErr(t *testing.T) {
	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           nil,
	}
	ihgs, err := sharding.NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, sharding.ErrNilPubKey, err)
}

func TestNewIndexHashedGroupSelector_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}

	ihgs, err := sharding.NewIndexHashedNodesCoordinator(arguments)
	assert.NotNil(t, ihgs)
	assert.Nil(t, err)
}

//------- LoadEligibleList

func TestIndexHashedGroupSelector_SetNilNodesMapShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 2,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}

	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)
	assert.Equal(t, sharding.ErrNilInputNodesMap, ihgs.SetNodesPerShards(nil))
}

func TestIndexHashedGroupSelector_OkValShouldWork(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 2,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}

	ihgs, err := sharding.NewIndexHashedNodesCoordinator(arguments)
	assert.Nil(t, err)
	assert.Equal(t, nodesMap[0], ihgs.EligibleList())
}

//------- ComputeValidatorsGroup

func TestIndexHashedGroupSelector_NewCoordinatorGroup0SizeShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		MetaConsensusGroupSize: 1,
		Hasher:                 &mock.HasherMock{},
		NbShards:               1,
		Nodes:                  nodesMap,
		SelfPublicKey:          []byte("key"),
	}
	ihgs, err := sharding.NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, sharding.ErrInvalidConsensusGroupSize, err)
}

func TestIndexHashedGroupSelector_NewCoordinatorTooFewNodesShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 10,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, err := sharding.NewIndexHashedNodesCoordinator(arguments)

	assert.Nil(t, ihgs)
	assert.Equal(t, sharding.ErrSmallShardEligibleListSize, err)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupNilRandomnessShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 2,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)
	list2, err := ihgs.ComputeValidatorsGroup(nil, 0, 0)

	assert.Nil(t, list2)
	assert.Equal(t, sharding.ErrNilRandomness, err)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupInvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 2,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)
	list2, err := ihgs.ComputeValidatorsGroup([]byte("radomness"), 0, 5)

	assert.Nil(t, list2)
	assert.Equal(t, sharding.ErrInvalidShardId, err)
}

//------- functionality tests

func TestIndexHashedGroupSelector_ComputeValidatorsGroup1ValidatorShouldReturnSame(t *testing.T) {
	t.Parallel()

	list := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0"), []byte("addr0")),
	}

	nodesMap := make(map[uint32][]sharding.Validator)
	nodesMap[0] = list
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)
	list2, err := ihgs.ComputeValidatorsGroup([]byte("randomness"), 0, 0)

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

	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 2,
		MetaConsensusGroupSize:  1,
		Hasher:                  hasher,
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)

	list2, err := ihgs.ComputeValidatorsGroup([]byte(randomness), 0, 0)

	assert.Nil(t, err)
	assert.Equal(t, nodesMap[0], list2)
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

	validator0 := mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0"), []byte("addr0"))
	validator1 := mock.NewValidatorMock(big.NewInt(2), 3, []byte("pk1"), []byte("addr1"))

	list := []sharding.Validator{
		validator0,
		validator1,
	}

	nodesMap := make(map[uint32][]sharding.Validator)
	nodesMap[0] = list
	metaNode, _ := sharding.NewValidator(big.NewInt(1), 1, []byte("pubKeyMeta"), []byte("addressMeta"))
	nodesMap[sharding.MetachainShardId] = []sharding.Validator{metaNode}
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 2,
		MetaConsensusGroupSize:  1,
		Hasher:                  hasher,
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)

	list2, err := ihgs.ComputeValidatorsGroup([]byte(randomness), 0, 0)

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

	nodesMap := createDummyNodesMap()
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 2,
		MetaConsensusGroupSize:  1,
		Hasher:                  hasher,
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)

	list2, err := ihgs.ComputeValidatorsGroup([]byte(randomness), 0, 0)

	assert.Nil(t, err)
	assert.Equal(t, nodesMap[0], list2)
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

	validator0 := mock.NewValidatorMock(big.NewInt(1), 1, []byte("pk0"), []byte("addr0"))
	validator1 := mock.NewValidatorMock(big.NewInt(2), 2, []byte("pk1"), []byte("addr1"))
	validator2 := mock.NewValidatorMock(big.NewInt(3), 3, []byte("pk2"), []byte("addr2"))
	validator3 := mock.NewValidatorMock(big.NewInt(4), 4, []byte("pk3"), []byte("addr3"))
	validator4 := mock.NewValidatorMock(big.NewInt(5), 5, []byte("pk4"), []byte("addr4"))
	validator5 := mock.NewValidatorMock(big.NewInt(6), 6, []byte("pk5"), []byte("addr5"))
	validator6 := mock.NewValidatorMock(big.NewInt(7), 7, []byte("pk6"), []byte("addr6"))
	validator7 := mock.NewValidatorMock(big.NewInt(8), 8, []byte("pk7"), []byte("addr7"))
	validator8 := mock.NewValidatorMock(big.NewInt(9), 9, []byte("pk8"), []byte("addr8"))
	validator9 := mock.NewValidatorMock(big.NewInt(10), 10, []byte("pk9"), []byte("addr9"))

	list := []sharding.Validator{
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

	nodesMap := make(map[uint32][]sharding.Validator)
	nodesMap[0] = list
	validatorMeta, _ := sharding.NewValidator(big.NewInt(1), 1, []byte("pubKeyMeta"), []byte("addressMeta"))
	nodesMap[sharding.MetachainShardId] = []sharding.Validator{validatorMeta}
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 6,
		MetaConsensusGroupSize:  1,
		Hasher:                  hasher,
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)

	list2, err := ihgs.ComputeValidatorsGroup([]byte(randomness), 0, 0)

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
	list := make([]sharding.Validator, 0)

	//generate 400 validators
	for i := 0; i < 400; i++ {
		list = append(list, mock.NewValidatorMock(big.NewInt(0), 0, []byte("pk"+strconv.Itoa(i)), []byte("addr"+strconv.Itoa(i))))
	}

	nodesMap := make(map[uint32][]sharding.Validator)
	nodesMap[0] = list

	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: consensusGroupSize,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomness := strconv.Itoa(i)
		list2, _ := ihgs.ComputeValidatorsGroup([]byte(randomness), 0, 0)

		assert.Equal(b, consensusGroupSize, len(list2))
	}
}

func TestIndexHashedGroupSelector_GetValidatorWithPublicKeyShouldReturnErrNilPubKey(t *testing.T) {
	t.Parallel()

	list := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0"), []byte("addr0")),
	}
	nodesMap := make(map[uint32][]sharding.Validator)
	nodesMap[0] = list
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)

	_, _, err := ihgs.GetValidatorWithPublicKey(nil)
	assert.Equal(t, sharding.ErrNilPubKey, err)
}

func TestIndexHashedGroupSelector_GetValidatorWithPublicKeyShouldReturnErrValidatorNotFound(t *testing.T) {
	t.Parallel()

	list := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0"), []byte("addr0")),
	}

	nodesMap := make(map[uint32][]sharding.Validator)
	nodesMap[0] = list
	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                1,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)

	_, _, err := ihgs.GetValidatorWithPublicKey([]byte("pk1"))
	assert.Equal(t, sharding.ErrValidatorNotFound, err)
}

func TestIndexHashedGroupSelector_GetValidatorWithPublicKeyShouldWork(t *testing.T) {
	t.Parallel()

	listMeta := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0_meta"), []byte("addr0_meta")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk1_meta"), []byte("addr1_meta")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk2_meta"), []byte("addr2_meta")),
	}
	listShard0 := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0_shard0"), []byte("addr0_shard0")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk1_shard0"), []byte("addr1_shard0")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk2_shard0"), []byte("addr2_shard0")),
	}
	listShard1 := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0_shard1"), []byte("addr0_shard1")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk1_shard1"), []byte("addr1_shard1")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk2_shard1"), []byte("addr2_shard1")),
	}

	nodesMap := make(map[uint32][]sharding.Validator)
	nodesMap[sharding.MetachainShardId] = listMeta
	nodesMap[0] = listShard0
	nodesMap[1] = listShard1

	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		NbShards:                2,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)

	validator, shardId, err := ihgs.GetValidatorWithPublicKey([]byte("pk0_meta"))
	assert.Nil(t, err)
	assert.Equal(t, sharding.MetachainShardId, shardId)
	assert.Equal(t, []byte("addr0_meta"), validator.Address())

	validator, shardId, err = ihgs.GetValidatorWithPublicKey([]byte("pk1_shard0"))
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), shardId)
	assert.Equal(t, []byte("addr1_shard0"), validator.Address())

	validator, shardId, err = ihgs.GetValidatorWithPublicKey([]byte("pk2_shard1"))
	assert.Nil(t, err)
	assert.Equal(t, uint32(1), shardId)
	assert.Equal(t, []byte("addr2_shard1"), validator.Address())
}

func TestIndexHashedGroupSelector_GetAllValidatorsPublicKeys(t *testing.T) {
	t.Parallel()

	shardZeroId := uint32(0)
	shardOneId := uint32(1)
	expectedValidatorsPubKeys := map[uint32][][]byte{
		shardZeroId:               {[]byte("pk0_shard0"), []byte("pk1_shard0"), []byte("pk2_shard0")},
		shardOneId:                {[]byte("pk0_shard1"), []byte("pk1_shard1"), []byte("pk2_shard1")},
		sharding.MetachainShardId: {[]byte("pk0_meta"), []byte("pk1_meta"), []byte("pk2_meta")},
	}

	listMeta := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, expectedValidatorsPubKeys[sharding.MetachainShardId][0], []byte("addr0_meta")),
		mock.NewValidatorMock(big.NewInt(1), 2, expectedValidatorsPubKeys[sharding.MetachainShardId][1], []byte("addr1_meta")),
		mock.NewValidatorMock(big.NewInt(1), 2, expectedValidatorsPubKeys[sharding.MetachainShardId][2], []byte("addr2_meta")),
	}
	listShard0 := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, expectedValidatorsPubKeys[shardZeroId][0], []byte("addr0_shard0")),
		mock.NewValidatorMock(big.NewInt(1), 2, expectedValidatorsPubKeys[shardZeroId][1], []byte("addr1_shard0")),
		mock.NewValidatorMock(big.NewInt(1), 2, expectedValidatorsPubKeys[shardZeroId][2], []byte("addr2_shard0")),
	}
	listShard1 := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, expectedValidatorsPubKeys[shardOneId][0], []byte("addr0_shard1")),
		mock.NewValidatorMock(big.NewInt(1), 2, expectedValidatorsPubKeys[shardOneId][1], []byte("addr1_shard1")),
		mock.NewValidatorMock(big.NewInt(1), 2, expectedValidatorsPubKeys[shardOneId][2], []byte("addr2_shard1")),
	}

	nodesMap := make(map[uint32][]sharding.Validator)
	nodesMap[sharding.MetachainShardId] = listMeta
	nodesMap[shardZeroId] = listShard0
	nodesMap[shardOneId] = listShard1

	arguments := sharding.ArgNodesCoordinator{
		ShardConsensusGroupSize: 1,
		MetaConsensusGroupSize:  1,
		Hasher:                  &mock.HasherMock{},
		ShardId:                 shardZeroId,
		NbShards:                2,
		Nodes:                   nodesMap,
		SelfPublicKey:           []byte("key"),
	}

	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(arguments)

	allValidatorsPublicKeys := ihgs.GetAllValidatorsPublicKeys()
	assert.Equal(t, expectedValidatorsPubKeys, allValidatorsPublicKeys)
}

func BenchmarkIndexHashedGroupSelector_TestExpandList(b *testing.B) {
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)

	fmt.Println(m.TotalAlloc)

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

	fmt.Println(m2.TotalAlloc)
	fmt.Println(fmt.Sprintf("Used %d MB", (m2.TotalAlloc-m.TotalAlloc)/1024/1024))
	//fmt.Print(array[0:100])
}

func BenchmarkIndexHashedGroupSelector_TestHashes(b *testing.B) {
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
