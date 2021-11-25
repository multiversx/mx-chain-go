package sharding

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/sharding/mock"
)

func computeRandomnessAsUint64(randomness []byte, index int) uint64 {
	buffCurrentIndex := make([]byte, 8)
	binary.BigEndian.PutUint64(buffCurrentIndex, uint64(index))

	hasher := &mock.HasherMock{}
	indexHash := hasher.Compute(string(buffCurrentIndex) + string(randomness))

	randomnessAsUint64 := binary.BigEndian.Uint64(indexHash)

	return randomnessAsUint64
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
