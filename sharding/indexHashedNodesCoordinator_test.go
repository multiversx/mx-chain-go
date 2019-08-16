package sharding_test

import (
	"encoding/binary"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/core"
	"math/big"
	"strconv"
	"testing"

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

	ihgs, err := sharding.NewIndexHashedNodesCoordinator(
		1,
		1,
		nil,
		0,
		1,
		nodesMap,
	)

	assert.Nil(t, ihgs)
	assert.Equal(t, sharding.ErrNilHasher, err)
}

func TestNewIndexHashedGroupSelector_InvalidConsensusGroupSizeShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	ihgs, err := sharding.NewIndexHashedNodesCoordinator(
		0,
		1,
		mock.HasherMock{},
		0,
		1,
		nodesMap,
	)

	assert.Nil(t, ihgs)
	assert.Equal(t, sharding.ErrInvalidConsensusGroupSize, err)
}

func TestNewIndexHashedGroupSelector_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	ihgs, err := sharding.NewIndexHashedNodesCoordinator(
		1,
		1,
		mock.HasherMock{},
		0,
		1,
		nodesMap,
	)

	assert.NotNil(t, ihgs)
	assert.Nil(t, err)
}

//------- LoadEligibleList

func TestIndexHashedGroupSelector_SetNilNodesMapShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(
		2,
		1,
		mock.HasherMock{},
		0,
		1,
		nodesMap,
	)

	assert.Equal(t, sharding.ErrNilInputNodesMap, ihgs.SetNodesPerShards(nil))
}

func TestIndexHashedGroupSelector_OkValShouldWork(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	ihgs, err := sharding.NewIndexHashedNodesCoordinator(
		2,
		1,
		mock.HasherMock{},
		0,
		1,
		nodesMap,
	)

	assert.Nil(t, err)
	assert.Equal(t, nodesMap[0], ihgs.EligibleList())
}

//------- ComputeValidatorsGroup

func TestIndexHashedGroupSelector_NewCoordinatorGroup0SizeShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	ihgs, err := sharding.NewIndexHashedNodesCoordinator(
		0,
		1,
		mock.HasherMock{},
		0,
		1,
		nodesMap,
	)

	assert.Nil(t, ihgs)
	assert.Equal(t, sharding.ErrInvalidConsensusGroupSize, err)
}

func TestIndexHashedGroupSelector_NewCoordinatorTooFewNodesShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	ihgs, err := sharding.NewIndexHashedNodesCoordinator(
		10,
		1,
		mock.HasherMock{},
		0,
		1,
		nodesMap,
	)

	assert.Nil(t, ihgs)
	assert.Equal(t, sharding.ErrSmallShardEligibleListSize, err)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupNilRandomnessShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(
		2,
		1,
		mock.HasherMock{},
		0,
		1,
		nodesMap,
	)

	list2, err := ihgs.ComputeValidatorsGroup(nil, 0, 0)

	assert.Nil(t, list2)
	assert.Equal(t, sharding.ErrNilRandomness, err)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupInvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	nodesMap := createDummyNodesMap()
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(
		2,
		1,
		mock.HasherMock{},
		0,
		1,
		nodesMap,
	)

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
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(
		1,
		1,
		mock.HasherMock{},
		0,
		1,
		nodesMap,
	)

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
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(
		2,
		1,
		hasher,
		0,
		1,
		nodesMap)

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
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(
		2,
		1,
		hasher,
		0,
		1,
		nodesMap,
	)

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
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(
		2,
		1,
		hasher,
		0,
		1,
		nodesMap,
	)

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
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(
		6,
		1,
		hasher,
		0,
		1,
		nodesMap,
	)

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

	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(
		consensusGroupSize,
		1,
		mock.HasherMock{},
		0,
		1,
		nodesMap,
	)

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
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(
		1,
		1,
		mock.HasherMock{},
		0,
		1,
		nodesMap,
	)

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
	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(
		1,
		1,
		mock.HasherMock{},
		0,
		1,
		nodesMap,
	)

	_, _, err := ihgs.GetValidatorWithPublicKey([]byte("pk1"))

	assert.Equal(t, sharding.ErrValidatorNotFound, err)
}

func TestIndexHashedGroupSelector_GetValidatorWithPublicKeyShouldWork(t *testing.T) {
	t.Parallel()

	list_meta := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0_meta"), []byte("addr0_meta")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk1_meta"), []byte("addr1_meta")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk2_meta"), []byte("addr2_meta")),
	}
	list_shard0 := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0_shard0"), []byte("addr0_shard0")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk1_shard0"), []byte("addr1_shard0")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk2_shard0"), []byte("addr2_shard0")),
	}
	list_shard1 := []sharding.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0_shard1"), []byte("addr0_shard1")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk1_shard1"), []byte("addr1_shard1")),
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk2_shard1"), []byte("addr2_shard1")),
	}

	nodesMap := make(map[uint32][]sharding.Validator)
	nodesMap[sharding.MetachainShardId] = list_meta
	nodesMap[0] = list_shard0
	nodesMap[1] = list_shard1

	ihgs, _ := sharding.NewIndexHashedNodesCoordinator(
		1,
		1,
		mock.HasherMock{},
		0,
		2,
		nodesMap,
	)

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
