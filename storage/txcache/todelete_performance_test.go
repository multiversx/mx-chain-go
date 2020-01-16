package txcache

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/require"
)

func Test_SortManySenders(t *testing.T) {
	mapOfSenders := newTxListBySenderMap(16)
	nSenders := 2500000

	stopWatch := core.NewStopWatch()

	stopWatch.Start("add")

	for senderTag := 0; senderTag < nSenders; senderTag++ {
		sender := createFakeSenderAddress(senderTag)
		mapOfSenders.addSender(string(sender))
	}

	stopWatch.Stop("add")

	stopWatch.Start("sortByOrderNumber")
	lists := mapOfSenders.GetListsSortedByOrderNumber()
	stopWatch.Stop("sortByOrderNumber")

	stopWatch.Start("sortByScore")
	lists = mapOfSenders.GetListsSortedBySmartScore()
	stopWatch.Stop("sortByScore")
	fmt.Println(len(lists))

	fmt.Println(stopWatch.GetMeasurements())

	require.True(t, false)
}

func Test_GoSearch(t *testing.T) {
	stopWatch := core.NewStopWatch()

	rand.Seed(time.Now().UnixNano())
	mySlice := make([]int, 1000000)
	for i := 0; i < len(mySlice); i++ {
		mySlice[i] = rand.Intn(int(100))
	}

	myCopy := make([]int, len(mySlice))
	copy(myCopy, mySlice)
	stopWatch.Start("golangsort")
	sort.Slice(myCopy, func(i, j int) bool {
		return myCopy[i] < myCopy[j]
	})
	stopWatch.Stop("golangsort")

	myCopy = make([]int, len(mySlice))
	copy(myCopy, mySlice)

	stopWatch.Start("countsort")
	CountingSort(myCopy)
	stopWatch.Stop("countsort")

	fmt.Println(stopWatch.GetMeasurements())
}

// Counting sort assumes that each of the n input elements is an integer
// in the range 0 to k, for some integer k.
// 1. Create an array(slice) of the size of the maximum value + 1.
// 2. Count each element.
// 3. Add up the elements.
// 4. Put them back to result.
func CountingSort(arr []int) []int {

	// 1. Create an array(slice) of the size of the maximum value + 1
	k := GetMaxIntArray(arr)
	count := make([]int, k+1)

	// 2. Count each element
	for i := 0; i < len(arr); i++ {
		count[arr[i]] = count[arr[i]] + 1
	}
	// fmt.Println(count)

	// 3. Add up the elements
	for i := 1; i < k+1; i++ {
		count[i] = count[i] + count[i-1]
	}
	// fmt.Println(count)

	// 4. Put them back to result
	result := make([]int, len(arr)+1)
	for j := 0; j < len(arr); j++ {
		result[count[arr[j]]] = arr[j]
		count[arr[j]] = count[arr[j]] - 1
	}
	// fmt.Println(count)

	return result
}

// Return the maximum value in an integer array.
func GetMaxIntArray(arr []int) int {
	max_num := arr[0]
	for _, elem := range arr {
		if max_num < elem {
			max_num = elem
		}
	}
	return max_num
}

// CountIntArray counts the element frequencies.
func CountIntArray(arr []int) map[int]int {
	model := make(map[int]int)
	for _, elem := range arr {
		// first element is already initialized with 0
		model[elem] += 1
	}
	return model
}
