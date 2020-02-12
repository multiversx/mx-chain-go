package sharding

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/stretchr/testify/assert"
)

func TestComputeStartIndexAndNumAppearancesForValidator(t *testing.T) {
	v1 := mock.NewValidatorMock(big.NewInt(10), 5, []byte("pk1"), []byte("addr1"))
	v2 := mock.NewValidatorMock(big.NewInt(10), 5, []byte("pk2"), []byte("addr2"))
	v3 := mock.NewValidatorMock(big.NewInt(10), 5, []byte("pk3"), []byte("addr3"))
	v4 := mock.NewValidatorMock(big.NewInt(10), 5, []byte("pk4"), []byte("addr4"))
	v5 := mock.NewValidatorMock(big.NewInt(10), 5, []byte("pk5"), []byte("addr5"))

	elList := make([]Validator, 0)
	elList = append(elList, v1, v1, v1)     // starts at 0 - count 3
	elList = append(elList, v2, v2, v2, v2) // starts at 3 - count 4
	elList = append(elList, v3, v3, v3)     // starts at 7 - count 3
	elList = append(elList, v4, v4, v4, v4) // starts at 10 - count 4
	elList = append(elList, v5)             // starts at 14 - count 1

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

type reslicingBasedProvider struct {
}

func NewReslicingBasedProvider() *reslicingBasedProvider {
	return &reslicingBasedProvider{}
}

func (rbp *reslicingBasedProvider) Get(randomness []byte, numVal int64, expEligibleList []Validator) ([]Validator, error) {
	expEligibleListClone := make([]Validator, len(expEligibleList))
	copy(expEligibleListClone, expEligibleList)

	valSlice := make([]Validator, 0, numVal)
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

	hasher := &mock.HasherMock{}
	indexHash := hasher.Compute(string(buffCurrentIndex) + string(randomness))

	randomnessAsUint64 := binary.BigEndian.Uint64(indexHash)

	return randomnessAsUint64
}

func reslice(slice []Validator, idx int64) []Validator {
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

func testWithReslicing(rand []byte, numVals int, expElList []Validator) []Validator {
	rbp := NewReslicingBasedProvider()
	res1, _ := rbp.Get(rand, int64(numVals), expElList)
	return res1
}

func testWithSelection(rand []byte, numVals int, expElList []Validator) []Validator {
	sbp := NewSelectionBasedProvider(&mock.HasherMock{}, uint32(numVals))
	res1, _ := sbp.Get(rand, int64(numVals), expElList)
	return res1
}

func testBothAlgorithmsHaveTheSameOutput(t *testing.T, rand []byte, numVals int, expElList []Validator) {
	resReslicing := testWithReslicing(rand, numVals, expElList)
	resSelection := testWithSelection(rand, numVals, expElList)

	assert.Equal(t, resReslicing, resSelection)
}

func displayVals(vals []Validator) {
	for _, val := range vals {
		fmt.Println(hex.EncodeToString(val.PubKey()))
	}
	fmt.Println()
}

func getExpandedEligibleList(num int) []Validator {
	sliceToRet := make([]Validator, 0)

	for i := 1; i <= num; i++ {
		randStake := rand.Intn(1234567)
		randRat := rand.Intn(5) + 8
		pubkey := make([]byte, 32)
		_, _ = rand.Read(pubkey)
		address := make([]byte, 32)
		_, _ = rand.Read(address)
		for j := 0; j < randRat; j++ {
			sliceToRet = append(sliceToRet, mock.NewValidatorMock(big.NewInt(int64(randStake)), int32(randRat), pubkey, address))
		}
	}

	return sliceToRet
}
