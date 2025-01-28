package nodesCoordinator

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
)

func TestComputeStartIndexAndNumAppearancesForValidator(t *testing.T) {
	elList := make([]uint32, 0)
	elList = append(elList, 0, 0, 0)    // starts at 0 - count 3
	elList = append(elList, 1, 1, 1, 1) // starts at 3 - count 4
	elList = append(elList, 2, 2, 2)    // starts at 7 - count 3
	elList = append(elList, 3, 3, 3, 3) // starts at 10 - count 4
	elList = append(elList, 4)          // starts at 14 - count 1

	type result struct {
		start int64
		num   int64
	}
	type fieldsStruct struct {
		indexes []int64
		res     result
	}

	// test all cases for the above list
	fields := []fieldsStruct{
		{
			indexes: []int64{0, 1, 2},
			res:     result{0, 3},
		},
		{
			indexes: []int64{3, 4, 5, 6},
			res:     result{3, 4},
		},
		{
			indexes: []int64{7, 8, 9},
			res:     result{7, 3},
		},
		{
			indexes: []int64{10, 11, 12, 13},
			res:     result{10, 4},
		},
		{
			indexes: []int64{14},
			res:     result{14, 1},
		},
	}

	for _, field := range fields {
		for _, idx := range field.indexes {
			resStart, resNum := computeStartIndexAndNumAppearancesForValidator(elList, idx)
			assert.Equal(t, field.res.start, resStart)
			assert.Equal(t, field.res.num, resNum)
		}
	}
}

// ------------- comparison between the selection algorithm and an algorithm which actually does reslicing

func GetValidatorsByReslicing(randomness []byte, numVal int64, expEligibleList []uint32) ([]uint32, error) {
	expEligibleListClone := make([]uint32, len(expEligibleList))
	copy(expEligibleListClone, expEligibleList)

	valSlice := make([]uint32, 0, numVal)
	for i := int64(0); i < numVal; i++ {
		rndmnss := computeRandomnessAsUint64(randomness, int(i))
		randomIdx := rndmnss % uint64(len(expEligibleListClone))
		valSlice = append(valSlice, expEligibleListClone[randomIdx])
		expEligibleListClone = reslice(expEligibleListClone, int64(randomIdx))
	}

	return valSlice, nil
}

func computeRandomnessAsUint64(randomness []byte, index int) uint64 {
	buffCurrentIndex := make([]byte, 8)
	binary.BigEndian.PutUint64(buffCurrentIndex, uint64(index))

	hasher := &hashingMocks.HasherMock{}
	indexHash := hasher.Compute(string(buffCurrentIndex) + string(randomness))

	randomnessAsUint64 := binary.BigEndian.Uint64(indexHash)

	return randomnessAsUint64
}

func reslice(slice []uint32, idx int64) []uint32 {
	startIdx, nbEntries := computeStartIndexAndNumAppearancesForValidator(slice, idx)
	endIdx := startIdx + nbEntries - 1

	retSl := append(slice[:startIdx], slice[endIdx+1:]...)

	return retSl
}

func TestBoth(t *testing.T) {
	numVals := 400
	expElList := getExpandedEligibleList(400)
	randomness := []byte("randomness")

	for i := 0; i < 100; i++ {
		testBothAlgorithmsHaveTheSameOutput(t, randomness, numVals, expElList)
	}
}

func testWithReslicing(rand []byte, numVals int, expElList []uint32) []uint32 {
	res1, _ := GetValidatorsByReslicing(rand, int64(numVals), expElList)
	return res1
}

func testWithSelection(rand []byte, numVals int, expElList []uint32) []uint32 {
	sbp := NewSelectionBasedProvider(&hashingMocks.HasherMock{}, uint32(numVals))
	res1, _ := sbp.Get(rand, int64(numVals), expElList)
	return res1
}

func testBothAlgorithmsHaveTheSameOutput(t *testing.T, rand []byte, numVals int, expElList []uint32) {
	resReslicing := testWithReslicing(rand, numVals, expElList)
	resSelection := testWithSelection(rand, numVals, expElList)

	assert.Equal(t, resReslicing, resSelection)
}

func displayVals(vals []uint32) {
	for _, v := range vals {
		fmt.Println(v)
	}
	fmt.Println()
}

func getExpandedEligibleList(num int) []uint32 {
	sliceToRet := make([]uint32, 0)

	for i := 1; i <= num; i++ {
		randBigInt, _ := rand.Int(rand.Reader, big.NewInt(5))
		randRat := int(randBigInt.Uint64()) + 8
		for j := 0; j < randRat; j++ {
			sliceToRet = append(sliceToRet, uint32(i))
		}
	}

	return sliceToRet
}

func newValidatorMock(pubKey []byte, chances uint32, index uint32) *validator {
	return &validator{pubKey: pubKey, index: index, chances: chances}
}

func TestSerializableShardValidatorListToValidatorListShouldErrNilPubKey(t *testing.T) {
	t.Parallel()

	listOfSerializableValidators := []*SerializableValidator{
		{
			PubKey:  nil,
			Chances: 1,
			Index:   1,
		},
	}

	_, err := SerializableShardValidatorListToValidatorList(listOfSerializableValidators)
	require.Equal(t, ErrNilPubKey, err)
}

func TestSerializableShardValidatorListToValidatorListShouldWork(t *testing.T) {
	t.Parallel()

	listOfSerializableValidators := []*SerializableValidator{
		{
			PubKey:  []byte("pubkey"),
			Chances: 1,
			Index:   1,
		},
	}

	expectedListOfValidators := make([]Validator, 1)
	v, _ := NewValidator(listOfSerializableValidators[0].PubKey, listOfSerializableValidators[0].Chances, listOfSerializableValidators[0].Index)
	require.NotNil(t, v)
	expectedListOfValidators[0] = v

	valReturned, err := SerializableShardValidatorListToValidatorList(listOfSerializableValidators)

	require.Nil(t, err)
	require.Equal(t, expectedListOfValidators, valReturned)
}

func TestSerializableValidatorsToValidatorsShouldWork(t *testing.T) {
	t.Parallel()

	mapOfSerializableValidators := make(map[string][]*SerializableValidator, 1)
	mapOfSerializableValidators["1"] = []*SerializableValidator{
		{
			PubKey:  []byte("pubkey"),
			Chances: 1,
			Index:   1,
		},
	}

	expectedMapOfValidators := make(map[uint32][]Validator, 1)

	v, _ := NewValidator(mapOfSerializableValidators["1"][0].PubKey, mapOfSerializableValidators["1"][0].Chances, mapOfSerializableValidators["1"][0].Index)
	expectedMapOfValidators[uint32(1)] = []Validator{v}

	require.NotNil(t, v)

	valReturned, err := SerializableValidatorsToValidators(mapOfSerializableValidators)

	require.Nil(t, err)
	require.Equal(t, expectedMapOfValidators, valReturned)
}

func TestSerializableValidatorsToValidatorsShouldErrNilPubKey(t *testing.T) {
	t.Parallel()

	mapOfSerializableValidators := make(map[string][]*SerializableValidator, 1)
	mapOfSerializableValidators["1"] = []*SerializableValidator{
		{
			PubKey:  nil,
			Chances: 1,
			Index:   1,
		},
	}

	_, err := SerializableValidatorsToValidators(mapOfSerializableValidators)

	require.Equal(t, ErrNilPubKey, err)
}

func TestSerializableValidatorsToValidatorsShouldErrEmptyString(t *testing.T) {
	t.Parallel()

	mapOfSerializableValidators := make(map[string][]*SerializableValidator, 1)
	mapOfSerializableValidators[""] = []*SerializableValidator{
		{
			PubKey:  []byte("pubkey"),
			Chances: 1,
			Index:   1,
		},
	}

	expectedMapOfValidators := make(map[uint32][]Validator, 1)

	v, _ := NewValidator(mapOfSerializableValidators[""][0].PubKey, mapOfSerializableValidators[""][0].Chances, mapOfSerializableValidators[""][0].Index)
	require.NotNil(t, v)
	expectedMapOfValidators[uint32(1)] = []Validator{v}

	_, err := SerializableValidatorsToValidators(mapOfSerializableValidators)

	require.Equal(t, &strconv.NumError{Func: "ParseUint", Num: "", Err: strconv.ErrSyntax}, err)
}
