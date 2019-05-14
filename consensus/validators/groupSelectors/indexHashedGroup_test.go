package groupSelectors_test

import (
	"encoding/binary"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/mock"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/validators/groupSelectors"
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

//------- NewIndexHashedGroupSelector

func TestNewIndexHashedGroupSelector_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	ihgs, err := groupSelectors.NewIndexHashedGroupSelector(1, nil)

	assert.Nil(t, ihgs)
	assert.Equal(t, groupSelectors.ErrNilHasher, err)
}

func TestNewIndexHashedGroupSelector_InvalidConsensusGroupSizeShouldErr(t *testing.T) {
	t.Parallel()

	ihgs, err := groupSelectors.NewIndexHashedGroupSelector(0, mock.HasherMock{})

	assert.Nil(t, ihgs)
	assert.Equal(t, groupSelectors.ErrInvalidConsensusGroupSize, err)
}

func TestNewIndexHashedGroupSelector_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	ihgs, err := groupSelectors.NewIndexHashedGroupSelector(1, mock.HasherMock{})

	assert.NotNil(t, ihgs)
	assert.Nil(t, err)
}

//------- LoadEligibleList

func TestIndexHashedGroupSelector_LoadEligibleListNilListShouldErr(t *testing.T) {
	t.Parallel()

	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(10, mock.HasherMock{})

	assert.Equal(t, groupSelectors.ErrNilInputSlice, ihgs.LoadEligibleList(nil))
}

func TestIndexHashedGroupSelector_OkValShouldWork(t *testing.T) {
	t.Parallel()

	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(10, mock.HasherMock{})

	list := []consensus.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0")),
		mock.NewValidatorMock(big.NewInt(2), 3, []byte("pk1")),
	}

	err := ihgs.LoadEligibleList(list)
	assert.Nil(t, err)
	assert.Equal(t, list, ihgs.EligibleList())
}

//------- ComputeValidatorsGroup

func TestIndexHashedGroupSelector_ComputeValidatorsGroup0SizeShouldErr(t *testing.T) {
	t.Parallel()

	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(1, mock.HasherMock{})

	list := make([]consensus.Validator, 0)

	list, err := ihgs.ComputeValidatorsGroup([]byte("randomness"))

	assert.Nil(t, list)
	assert.Equal(t, groupSelectors.ErrSmallEligibleListSize, err)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupWrongSizeShouldErr(t *testing.T) {
	t.Parallel()

	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(10, mock.HasherMock{})

	list := []consensus.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0")),
		mock.NewValidatorMock(big.NewInt(2), 3, []byte("pk1")),
	}

	_ = ihgs.LoadEligibleList(list)

	list, err := ihgs.ComputeValidatorsGroup([]byte("randomness"))

	assert.Nil(t, list)
	assert.Equal(t, groupSelectors.ErrSmallEligibleListSize, err)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupNilRandomnessShouldErr(t *testing.T) {
	t.Parallel()

	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(2, mock.HasherMock{})

	list := []consensus.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0")),
		mock.NewValidatorMock(big.NewInt(2), 3, []byte("pk1")),
	}

	_ = ihgs.LoadEligibleList(list)

	list2, err := ihgs.ComputeValidatorsGroup(nil)

	assert.Nil(t, list2)
	assert.Equal(t, groupSelectors.ErrNilRandomness, err)
}

//------- functionality tests

func TestIndexHashedGroupSelector_ComputeValidatorsGroup1ValidatorShouldReturnSame(t *testing.T) {
	t.Parallel()

	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(1, mock.HasherMock{})

	list := []consensus.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0")),
	}

	_ = ihgs.LoadEligibleList(list)

	list2, err := ihgs.ComputeValidatorsGroup([]byte("randomness"))

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

	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(2, hasher)

	list := []consensus.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0")),
		mock.NewValidatorMock(big.NewInt(2), 3, []byte("pk1")),
	}

	_ = ihgs.LoadEligibleList(list)

	list2, err := ihgs.ComputeValidatorsGroup([]byte(randomness))

	assert.Nil(t, err)
	assert.Equal(t, list, list2)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupTest2ValidatorsRevertOrder(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherStub{}

	randomness := "randomness"

	//this will return the list in reverse order:
	//element 0 will be the second
	//element 1 will be the first
	hasher.ComputeCalled = func(s string) []byte {
		if string(uint64ToBytes(0))+randomness == s {
			return convertBigIntToBytes(big.NewInt(1))
		}

		if string(uint64ToBytes(1))+randomness == s {
			return convertBigIntToBytes(big.NewInt(0))
		}

		return nil
	}

	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(2, hasher)

	validator0 := mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0"))
	validator1 := mock.NewValidatorMock(big.NewInt(2), 3, []byte("pk1"))

	list := []consensus.Validator{
		validator0,
		validator1,
	}

	_ = ihgs.LoadEligibleList(list)

	list2, err := ihgs.ComputeValidatorsGroup([]byte(randomness))

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

	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(2, hasher)

	list := []consensus.Validator{
		mock.NewValidatorMock(big.NewInt(1), 2, []byte("pk0")),
		mock.NewValidatorMock(big.NewInt(2), 3, []byte("pk1")),
	}

	_ = ihgs.LoadEligibleList(list)

	list2, err := ihgs.ComputeValidatorsGroup([]byte(randomness))

	assert.Nil(t, err)
	assert.Equal(t, list, list2)
}

func TestIndexHashedGroupSelector_ComputeValidatorsGroupTest6From10ValidatorsShouldWork(t *testing.T) {
	t.Parallel()

	hasher := &mock.HasherStub{}

	randomness := "randomness"

	//script:
	// for index 0, hasher will return 11 which will translate to 1, so 1 is the first element
	// for index 1, hasher will return 1 which will translate to 1, 1 is already picked, try the next, 2 is the second element
	// for index 2, hasher will return 9 which will translate to 9, 9 is the 3-rd element
	// for index 3, hasher will return 9 which will translate to 9, 9 is already picked, try the next one, 0 is the 4-th element
	// for index 4, hasher will return 0 which will translate to 0, 0 is already picked, 1 is already picked, 2 is already picked,
	//      3 is the 4-th element
	// for index 5, hasher will return 9 which will translate to 9, so 9, 0, 1, 2, 3 are already picked, 4 is the 5-th element

	script := make(map[string]*big.Int)
	script[string(uint64ToBytes(0))+randomness] = big.NewInt(11) //will translate to 1, add 1
	script[string(uint64ToBytes(1))+randomness] = big.NewInt(1)  //will translate to 1, add 2
	script[string(uint64ToBytes(2))+randomness] = big.NewInt(9)  //will translate to 9, add 9
	script[string(uint64ToBytes(3))+randomness] = big.NewInt(9)  //will translate to 9, add 0
	script[string(uint64ToBytes(4))+randomness] = big.NewInt(0)  //will translate to 0, add 3
	script[string(uint64ToBytes(5))+randomness] = big.NewInt(9)  //will translate to 9, add 4

	hasher.ComputeCalled = func(s string) []byte {
		val, ok := script[s]

		if !ok {
			assert.Fail(t, "should have not got here")
		}

		return convertBigIntToBytes(val)
	}

	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(6, hasher)

	validator0 := mock.NewValidatorMock(big.NewInt(1), 1, []byte("pk0"))
	validator1 := mock.NewValidatorMock(big.NewInt(2), 2, []byte("pk1"))
	validator2 := mock.NewValidatorMock(big.NewInt(3), 3, []byte("pk2"))
	validator3 := mock.NewValidatorMock(big.NewInt(4), 4, []byte("pk3"))
	validator4 := mock.NewValidatorMock(big.NewInt(5), 5, []byte("pk4"))
	validator5 := mock.NewValidatorMock(big.NewInt(6), 6, []byte("pk5"))
	validator6 := mock.NewValidatorMock(big.NewInt(7), 7, []byte("pk6"))
	validator7 := mock.NewValidatorMock(big.NewInt(8), 8, []byte("pk7"))
	validator8 := mock.NewValidatorMock(big.NewInt(9), 9, []byte("pk8"))
	validator9 := mock.NewValidatorMock(big.NewInt(10), 10, []byte("pk9"))

	list := []consensus.Validator{
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

	_ = ihgs.LoadEligibleList(list)

	list2, err := ihgs.ComputeValidatorsGroup([]byte(randomness))

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

	ihgs, _ := groupSelectors.NewIndexHashedGroupSelector(consensusGroupSize, mock.HasherMock{})

	list := make([]consensus.Validator, 0)

	//generate 400 validators
	for i := 0; i < 400; i++ {
		list = append(list, mock.NewValidatorMock(big.NewInt(0), 0, []byte("pk"+strconv.Itoa(i))))
	}
	_ = ihgs.LoadEligibleList(list)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		randomness := strconv.Itoa(i)

		list2, _ := ihgs.ComputeValidatorsGroup([]byte(randomness))

		assert.Equal(b, consensusGroupSize, len(list2))
	}
}
