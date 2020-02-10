package consensusGroupProviders

import (
	"fmt"
	"math/big"
	"math/rand"
	"runtime"
	"testing"

	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/mock"
	"github.com/stretchr/testify/assert"
)

func BenchmarkReslicingBasedProvider_Get(b *testing.B) {
	numVals := 400
	expElList := getExpandedEligibleList(400)
	randomness := uint64(12345)

	for i := 0; i < b.N; i++ {
		testWithReslicing(randomness, numVals, expElList)
	}
}

func BenchmarkSelectionBasedProvider_Get(b *testing.B) {
	numVals := 400
	expElList := getExpandedEligibleList(400)
	randomness := uint64(12345)

	for i := 0; i < b.N; i++ {
		testWithSelection(randomness, numVals, expElList)
	}
}

func TestMemUsageWithReslicing(t *testing.T) {
	numVals := 400
	expElList := getExpandedEligibleList(400)
	randomness := uint64(12345)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Println("\nAlloc", memStats.Alloc)
	fmt.Println("HeapAlloc", memStats.HeapAlloc)
	fmt.Println("HeapSys", memStats.HeapSys)
	fmt.Println("HeapReleased", memStats.HeapReleased)
	fmt.Println("TotalAlloc", memStats.TotalAlloc)
	fmt.Println("----------------")

	for i := 0; i < 1; i++ {
		testWithReslicing(randomness, numVals, expElList)
	}

	runtime.ReadMemStats(&memStats)
	fmt.Println("\nAlloc", memStats.Alloc)
	fmt.Println("HeapAlloc", memStats.HeapAlloc)
	fmt.Println("HeapSys", memStats.HeapSys)
	fmt.Println("HeapReleased", memStats.HeapReleased)
	fmt.Println("TotalAlloc", memStats.TotalAlloc)
}

func TestMemUsageWithSelection(t *testing.T) {
	numVals := 400
	expElList := getExpandedEligibleList(400)
	randomness := uint64(12345)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Println("\nAlloc", memStats.Alloc)
	fmt.Println("HeapAlloc", memStats.HeapAlloc)
	fmt.Println("HeapSys", memStats.HeapSys)
	fmt.Println("HeapReleased", memStats.HeapReleased)
	fmt.Println("TotalAlloc", memStats.TotalAlloc)
	fmt.Println("----------------")

	for i := 0; i < 1; i++ {
		testWithSelection(randomness, numVals, expElList)
	}

	runtime.ReadMemStats(&memStats)
	fmt.Println("\nAlloc", memStats.Alloc)
	fmt.Println("HeapAlloc", memStats.HeapAlloc)
	fmt.Println("HeapSys", memStats.HeapSys)
	fmt.Println("HeapReleased", memStats.HeapReleased)
	fmt.Println("TotalAlloc", memStats.TotalAlloc)
}

func TestBoth(t *testing.T) {
	numVals := 400
	expElList := getExpandedEligibleList(400)
	randomness := uint64(12345)
	for i := 0; i < 100; i++ {
		testBothAlgorithmsHaveTheSameOutput(t, randomness, numVals, expElList)
	}
}

func testWithReslicing(rand uint64, numVals int, expElList []sharding.Validator) []sharding.Validator {
	rbp := NewReslicingBasedProvider()
	res1, _ := rbp.Get(rand, int64(numVals), expElList)
	return res1
}

func testWithSelection(rand uint64, numVals int, expElList []sharding.Validator) []sharding.Validator {
	sbp := NewSelectionBasedProvider()
	res1, _ := sbp.Get(rand, int64(numVals), expElList)
	return res1
}

func testBothAlgorithmsHaveTheSameOutput(t *testing.T, rand uint64, numVals int, expElList []sharding.Validator) {
	resReslicing := testWithReslicing(rand, numVals, expElList)
	resSelection := testWithSelection(rand, numVals, expElList)

	assert.Equal(t, resReslicing, resSelection)
}

func displayVals(vals []sharding.Validator) {
	for _, val := range vals {
		fmt.Println(string(val.Address()))
	}
	fmt.Println()
}

func getExpandedEligibleList(num int) []sharding.Validator {
	sliceToRet := make([]sharding.Validator, 0)

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
