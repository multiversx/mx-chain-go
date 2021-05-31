package sharding

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	mathRand "math/rand"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	shuffleBetweenShards = false
	adaptivity           = false
	hysteresis           = float32(0.2)
	eligiblePerShard     = 100
	waitingPerShard      = 30
)

func generateRandomByteArray(size int) []byte {
	r := make([]byte, size)
	_, _ = rand.Read(r)

	return r
}

func generateValidatorList(number int) []Validator {
	v := make([]Validator, number)

	for i := 0; i < number; i++ {
		v[i] = &validator{
			pubKey: generateRandomByteArray(32),
		}
	}

	return v
}

func generateValidatorMap(
	nodesPerShard int,
	nbShards uint32,
) map[uint32][]Validator {
	validatorsMap := make(map[uint32][]Validator)

	for i := uint32(0); i < nbShards; i++ {
		validatorsMap[i] = generateValidatorList(nodesPerShard)
	}

	validatorsMap[core.MetachainShardId] = generateValidatorList(nodesPerShard)

	return validatorsMap
}

func contains(a []Validator, b []Validator) bool {
	var found bool
	for _, va := range a {
		found = false
		for _, vb := range b {
			if reflect.DeepEqual(va, vb) {
				found = true
				break
			}
		}
		if !found {
			return found
		}
	}

	return found
}

func testRemoveValidators(
	t *testing.T,
	initialValidators []Validator,
	validatorsToRemove []Validator,
	remaining []Validator,
	removed []Validator,
	maxToRemove int,
) {
	nbRemoved := maxToRemove
	if nbRemoved > len(validatorsToRemove) {
		nbRemoved = len(validatorsToRemove)
	}

	assert.Equal(t, nbRemoved, len(removed))
	assert.Equal(t, len(initialValidators)-len(remaining), nbRemoved)

	all := append(remaining, removed...)
	assert.True(t, contains(all, initialValidators))
	assert.Equal(t, len(initialValidators), len(all))
}

func testDistributeValidators(
	t *testing.T,
	initialMap map[uint32][]Validator,
	resultedMap map[uint32][]Validator,
	distributedNodes []Validator,
) {
	totalResultingValidators := make([]Validator, 0)
	totalLen := 0
	for _, valList := range resultedMap {
		totalResultingValidators = append(totalResultingValidators, valList...)
		totalLen += len(valList)
	}

	totalValidators := make([]Validator, 0)
	for _, valList := range initialMap {
		totalValidators = append(totalValidators, valList...)
	}
	assert.Equal(t, len(totalValidators)+len(distributedNodes), totalLen)

	totalValidators = append(totalValidators, distributedNodes...)
	assert.True(t, contains(totalResultingValidators, totalValidators))
}

func numberMatchingNodes(searchList []Validator, toFind []Validator) int {
	nbFound := 0
	for _, v1 := range toFind {
		for _, v2 := range searchList {
			if bytes.Equal(v1.PubKey(), v2.PubKey()) {
				nbFound++
				break
			}
		}
	}

	return nbFound
}

func testLeaving(
	t *testing.T,
	eligible map[uint32][]Validator,
	waiting map[uint32][]Validator,
	prevLeaving []Validator,
	newLeaving []Validator,
) (int, map[uint32]int) {
	nbLeavingPerShard := make(map[uint32]int)

	nbLeavingFromEligible := 0
	for i, eligibleList := range eligible {
		nbWantingToLeaveFromList := numberMatchingNodes(eligibleList, prevLeaving)
		maxAllowedToLeaveFromList := len(waiting[i])
		nbLeaving := nbWantingToLeaveFromList
		if nbLeaving > maxAllowedToLeaveFromList {
			nbLeaving = maxAllowedToLeaveFromList
		}

		nbLeavingPerShard[i] += nbLeaving
		nbLeavingFromEligible += nbLeaving
	}
	assert.Equal(t, nbLeavingFromEligible, len(prevLeaving)-len(newLeaving))

	return nbLeavingFromEligible, nbLeavingPerShard
}

func testShuffledOut(
	t *testing.T,
	eligibleMap map[uint32][]Validator,
	waitingMap map[uint32][]Validator,
	newEligible map[uint32][]Validator,
	shuffledOut []Validator,
	prevleaving []Validator,
	newleaving []Validator,
) {
	nbAllLeaving, _ := testLeaving(t, eligibleMap, waitingMap, prevleaving, newleaving)
	allWaiting := getValidatorsInMap(waitingMap)
	allEligible := getValidatorsInMap(eligibleMap)
	assert.Equal(t, len(shuffledOut)+nbAllLeaving, len(allWaiting))

	allNewEligible := getValidatorsInMap(newEligible)
	assert.Equal(t, len(allEligible)-len(shuffledOut)-nbAllLeaving, len(allNewEligible))

	newNodes := append(allNewEligible, shuffledOut...)
	assert.NotEqual(t, allEligible, newNodes)
	assert.True(t, contains(newNodes, allEligible))
}

func createHashShufflerInter() (*randHashShuffler, error) {
	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           eligiblePerShard,
		NodesMeta:            eligiblePerShard,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: true,
		MaxNodesEnableConfig: nil,
	}

	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)

	return shuffler, err
}

func createHashShufflerIntraShards() (*randHashShuffler, error) {
	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           eligiblePerShard,
		NodesMeta:            eligiblePerShard,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)

	return shuffler, err
}

func getValidatorsInMap(valMap map[uint32][]Validator) []Validator {
	result := make([]Validator, 0)

	for _, valList := range valMap {
		result = append(result, valList...)
	}

	return result
}

func Test_copyValidatorMap(t *testing.T) {
	t.Parallel()

	valMap := generateValidatorMap(30, 2)
	v2 := copyValidatorMap(valMap)
	assert.Equal(t, valMap, v2)

	valMap[0] = valMap[0][1:]
	assert.NotEqual(t, valMap, v2)
}

func Test_promoteWaitingToEligibleEmptyList(t *testing.T) {
	t.Parallel()

	eligibleMap := generateValidatorMap(30, 2)
	waitingMap := generateValidatorMap(0, 2)
	eligibleMapCopy := copyValidatorMap(eligibleMap)

	for k := range eligibleMap {
		assert.Equal(t, eligibleMap[k], eligibleMapCopy[k])
		assert.Empty(t, waitingMap[k])
	}
}

// eligible 40: 0 in eligible, 10 waiting
func Test_promoteWaitingToEligible_ZeroEligible(t *testing.T) {
	t.Parallel()

	numMeta := uint32(40)
	numShard := uint32(40)
	numWaiting := 10

	eligibleMap := generateValidatorMap(0, 2)
	waitingMap := generateValidatorMap(numWaiting, 2)

	eligibleMapCopy := copyValidatorMap(eligibleMap)
	waitingMapCopy := copyValidatorMap(waitingMap)

	err := moveMaxNumNodesToMap(eligibleMap, waitingMap, numMeta, numShard)
	assert.Nil(t, err)

	for k := range eligibleMap {
		assert.Equal(t, append(eligibleMapCopy[k], waitingMapCopy[k][0:numWaiting]...), eligibleMap[k])
		assert.Empty(t, waitingMap[k])
	}
}

// eligible 40: 30 in eligible, 6 waiting
func Test_promoteWaitingToEligible_LessWaitingThanRemainingSize(t *testing.T) {
	t.Parallel()

	numMeta := uint32(40)
	numShard := uint32(40)
	numWaiting := 6

	eligibleMap := generateValidatorMap(30, 2)
	waitingMap := generateValidatorMap(numWaiting, 2)

	eligibleMapCopy := copyValidatorMap(eligibleMap)
	waitingMapCopy := copyValidatorMap(waitingMap)

	err := moveMaxNumNodesToMap(eligibleMap, waitingMap, numMeta, numShard)
	assert.Nil(t, err)

	for k := range eligibleMap {
		assert.Equal(t, append(eligibleMapCopy[k], waitingMapCopy[k][0:numWaiting]...), eligibleMap[k])
		assert.Empty(t, waitingMap[k])
	}
}

// eligible 40: 30 in eligible, 10 waiting
func Test_promoteWaitingToEligible_ExactlyWaitingToRemainingSize(t *testing.T) {
	t.Parallel()

	numMeta := uint32(40)
	numShard := uint32(40)
	numWaiting := 10

	eligibleMap := generateValidatorMap(30, 2)
	waitingMap := generateValidatorMap(numWaiting, 2)

	eligibleMapCopy := copyValidatorMap(eligibleMap)
	waitingMapCopy := copyValidatorMap(waitingMap)

	err := moveMaxNumNodesToMap(eligibleMap, waitingMap, numMeta, numShard)
	assert.Nil(t, err)

	for k := range eligibleMap {
		assert.Equal(t, eligibleMap[k], append(eligibleMapCopy[k], waitingMapCopy[k]...))
		assert.Empty(t, waitingMap[k])
	}
}

// eligible 40: 30 in eligible, 15 waiting
func Test_promoteWaitingToEligible_MoreWaitingThanRemainingSize(t *testing.T) {
	t.Parallel()

	numMeta := uint32(40)
	numShard := uint32(40)
	numWaiting := 15
	numRemaining := 10
	eligibleMap := generateValidatorMap(30, 2)
	waitingMap := generateValidatorMap(numWaiting, 2)

	eligibleMapCopy := copyValidatorMap(eligibleMap)
	waitingMapCopy := copyValidatorMap(waitingMap)

	err := moveMaxNumNodesToMap(eligibleMap, waitingMap, numMeta, numShard)
	assert.Nil(t, err)

	for k := range eligibleMap {
		assert.Equal(t, append(eligibleMapCopy[k], waitingMapCopy[k][0:numRemaining]...), eligibleMap[k])
		assert.Equal(t, waitingMapCopy[k][numRemaining:], waitingMap[k])
	}
}

// eligible 40: 20 in eligible, 80 waiting
func Test_promoteWaitingToEligible_2xMoreWaitingThanRemainingSize(t *testing.T) {
	t.Parallel()

	numMeta := uint32(40)
	numShard := uint32(40)
	numWaiting := 80
	numRemaining := 20
	eligibleMap := generateValidatorMap(20, 2)
	waitingMap := generateValidatorMap(numWaiting, 2)

	eligibleMapCopy := copyValidatorMap(eligibleMap)
	waitingMapCopy := copyValidatorMap(waitingMap)

	err := moveMaxNumNodesToMap(eligibleMap, waitingMap, numMeta, numShard)
	assert.Nil(t, err)

	for k := range eligibleMap {
		assert.Equal(t, append(eligibleMapCopy[k], waitingMapCopy[k][0:numRemaining]...), eligibleMap[k])
		assert.Equal(t, waitingMapCopy[k][numRemaining:], waitingMap[k])
	}

	// cleanup eligible to have space for another 40
	for k := range eligibleMap {
		eligibleMap[k] = make([]Validator, 0)
	}
	err = moveMaxNumNodesToMap(eligibleMap, waitingMap, numMeta, numShard)
	assert.Nil(t, err)

	remainingIndex := 60
	for k := range eligibleMap {
		assert.Equal(t, waitingMapCopy[k][numRemaining:remainingIndex], eligibleMap[k])
		assert.Equal(t, waitingMapCopy[k][remainingIndex:], waitingMap[k])
	}

	// cleanup eligible to have space for another 40
	for k := range eligibleMap {
		eligibleMap[k] = make([]Validator, 0)
	}
	err = moveMaxNumNodesToMap(eligibleMap, waitingMap, numMeta, numShard)
	assert.Nil(t, err)

	for k := range eligibleMap {
		assert.Equal(t, waitingMapCopy[k][remainingIndex:80], eligibleMap[k])
		assert.Empty(t, waitingMap[k])
	}
}

func Test_promoteWaitingToEligible_(t *testing.T) {
	t.Parallel()

	eligibleMap := generateValidatorMap(30, 2)
	waitingMap := generateValidatorMap(22, 2)

	eligibleMapCopy := copyValidatorMap(eligibleMap)
	waitingMapCopy := copyValidatorMap(waitingMap)

	err := moveNodesToMap(eligibleMap, waitingMap)
	assert.Nil(t, err)

	for k := range eligibleMap {
		assert.Equal(t, eligibleMap[k], append(eligibleMapCopy[k], waitingMapCopy[k]...))
		assert.Empty(t, waitingMap[k])
	}
}

func Test_promoteWaitingToNilEligible(t *testing.T) {
	t.Parallel()

	waitingMap := generateValidatorMap(22, 2)

	err := moveNodesToMap(nil, waitingMap)

	assert.Equal(t, ErrNilOrEmptyDestinationForDistribute, err)
}

func Test_removeValidatorFromListFirst(t *testing.T) {
	t.Parallel()

	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	_ = copy(validatorsCopy, validators)

	v := removeValidatorFromList(validators, 0)
	assert.Equal(t, validatorsCopy[len(validatorsCopy)-1], v[0])
	assert.NotEqual(t, validatorsCopy[0], v[0])
	assert.Equal(t, len(validatorsCopy)-1, len(v))

	for i := 1; i < len(v); i++ {
		assert.Equal(t, validatorsCopy[i], v[i])
	}
}

func Test_removeValidatorFromListLast(t *testing.T) {
	t.Parallel()

	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	_ = copy(validatorsCopy, validators)

	v := removeValidatorFromList(validators, len(validators)-1)
	assert.Equal(t, len(validatorsCopy)-1, len(v))
	assert.Equal(t, validatorsCopy[:len(validatorsCopy)-1], v)
}

func Test_removeValidatorFromListMiddle(t *testing.T) {
	t.Parallel()

	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	_ = copy(validatorsCopy, validators)

	v := removeValidatorFromList(validators, len(validators)/2)
	assert.Equal(t, len(validatorsCopy)-1, len(v))
	assert.Equal(t, validatorsCopy[len(validatorsCopy)-1], v[len(validatorsCopy)/2])
}

func Test_removeValidatorFromListIndexNegativeNoAction(t *testing.T) {
	t.Parallel()

	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	_ = copy(validatorsCopy, validators)

	v := removeValidatorFromList(validators, -1)
	assert.Equal(t, len(validatorsCopy), len(v))
	assert.Equal(t, validatorsCopy, v)
}

func Test_removeValidatorFromListIndexTooBigNoAction(t *testing.T) {
	t.Parallel()

	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	_ = copy(validatorsCopy, validators)

	v := removeValidatorFromList(validators, len(validators))
	assert.Equal(t, len(validatorsCopy), len(v))
	assert.Equal(t, validatorsCopy, v)
}

func Test_removeValidatorsFromListRemoveFromStart(t *testing.T) {
	t.Parallel()

	validatorsToRemoveFromStart := 3
	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	validatorsToRemove := make([]Validator, 0)

	_ = copy(validatorsCopy, validators)
	validatorsToRemove = append(validatorsToRemove, validators[:validatorsToRemoveFromStart]...)

	v, removed := removeValidatorsFromList(validators, validatorsToRemove, len(validatorsToRemove))
	testRemoveValidators(t, validatorsCopy, validatorsToRemove, v, removed, len(validatorsToRemove))
}

func Test_removeValidatorsFromListRemoveFromLast(t *testing.T) {
	t.Parallel()

	validatorsToRemoveFromEnd := 3
	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	validatorsToRemove := make([]Validator, 0)

	_ = copy(validatorsCopy, validators)
	validatorsToRemove = append(validatorsToRemove, validators[len(validators)-validatorsToRemoveFromEnd:]...)

	v, removed := removeValidatorsFromList(validators, validatorsToRemove, len(validatorsToRemove))
	testRemoveValidators(t, validatorsCopy, validatorsToRemove, v, removed, len(validatorsToRemove))
}

func Test_removeValidatorsFromListRemoveFromFirstMaxSmaller(t *testing.T) {
	t.Parallel()

	validatorsToRemoveFromStart := 3
	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	validatorsToRemove := make([]Validator, 0)
	maxToRemove := validatorsToRemoveFromStart - 1

	_ = copy(validatorsCopy, validators)
	validatorsToRemove = append(validatorsToRemove, validators[:validatorsToRemoveFromStart]...)

	v, removed := removeValidatorsFromList(validators, validatorsToRemove, maxToRemove)
	testRemoveValidators(t, validatorsCopy, validatorsToRemove, v, removed, maxToRemove)
}

func Test_removeValidatorsFromListRemoveFromFirstMaxGreater(t *testing.T) {
	t.Parallel()

	validatorsToRemoveFromStart := 3
	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	validatorsToRemove := make([]Validator, 0)
	maxToRemove := validatorsToRemoveFromStart + 1

	_ = copy(validatorsCopy, validators)
	validatorsToRemove = append(validatorsToRemove, validators[:validatorsToRemoveFromStart]...)

	v, removed := removeValidatorsFromList(validators, validatorsToRemove, maxToRemove)
	testRemoveValidators(t, validatorsCopy, validatorsToRemove, v, removed, maxToRemove)
}

func Test_removeValidatorsFromListRemoveFromLastMaxSmaller(t *testing.T) {
	t.Parallel()

	validatorsToRemoveFromEnd := 3
	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	validatorsToRemove := make([]Validator, 0)
	maxToRemove := validatorsToRemoveFromEnd - 1

	_ = copy(validatorsCopy, validators)
	validatorsToRemove = append(validatorsToRemove, validators[len(validators)-validatorsToRemoveFromEnd:]...)
	assert.Equal(t, validatorsToRemoveFromEnd, len(validatorsToRemove))

	v, removed := removeValidatorsFromList(validators, validatorsToRemove, maxToRemove)
	testRemoveValidators(t, validatorsCopy, validatorsToRemove, v, removed, maxToRemove)
}

func Test_removeValidatorsFromListRemoveFromLastMaxGreater(t *testing.T) {
	t.Parallel()

	validatorsToRemoveFromEnd := 3
	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	validatorsToRemove := make([]Validator, 0)
	maxToRemove := validatorsToRemoveFromEnd + 1

	_ = copy(validatorsCopy, validators)
	validatorsToRemove = append(validatorsToRemove, validators[len(validators)-validatorsToRemoveFromEnd:]...)
	assert.Equal(t, validatorsToRemoveFromEnd, len(validatorsToRemove))

	v, removed := removeValidatorsFromList(validators, validatorsToRemove, maxToRemove)
	testRemoveValidators(t, validatorsCopy, validatorsToRemove, v, removed, maxToRemove)
}

func Test_removeValidatorsFromListRandomValidatorsMaxSmaller(t *testing.T) {
	t.Parallel()

	nbValidatotrsToRemove := 10
	maxToRemove := nbValidatotrsToRemove - 3
	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	validatorsToRemove := make([]Validator, 0)

	_ = copy(validatorsCopy, validators)

	sort.Sort(validatorList(validators))

	validatorsToRemove = append(validatorsToRemove, validators[:nbValidatotrsToRemove]...)

	v, removed := removeValidatorsFromList(validators, validatorsToRemove, maxToRemove)
	testRemoveValidators(t, validatorsCopy, validatorsToRemove, v, removed, maxToRemove)
}

func Test_removeValidatorsFromListRandomValidatorsMaxGreater(t *testing.T) {
	t.Parallel()

	nbValidatotrsToRemove := 10
	maxToRemove := nbValidatotrsToRemove + 3
	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	validatorsToRemove := make([]Validator, 0)

	_ = copy(validatorsCopy, validators)

	sort.Sort(validatorList(validators))

	validatorsToRemove = append(validatorsToRemove, validators[:nbValidatotrsToRemove]...)

	v, removed := removeValidatorsFromList(validators, validatorsToRemove, maxToRemove)
	testRemoveValidators(t, validatorsCopy, validatorsToRemove, v, removed, maxToRemove)
}

func Test_removeDupplicates_NoDupplicates(t *testing.T) {
	t.Parallel()

	firstList := generateValidatorList(30)
	secondList := generateValidatorList(30)

	firstListCopy := make([]Validator, len(firstList))
	copy(firstListCopy, firstList)

	secondListCopy := make([]Validator, len(secondList))
	copy(secondListCopy, secondList)

	secondListAfterRemove := removeDupplicates(firstList, secondList)

	assert.Equal(t, firstListCopy, firstList)
	assert.Equal(t, secondListCopy, secondListAfterRemove)
}

func Test_removeDupplicates_SomeDupplicates(t *testing.T) {
	t.Parallel()

	firstList := generateValidatorList(30)
	secondList := generateValidatorList(20)
	validatorsFromFirstList := firstList[0:10]
	secondList = append(secondList, validatorsFromFirstList...)

	firstListCopy := make([]Validator, len(firstList))
	copy(firstListCopy, firstList)

	secondListCopy := make([]Validator, len(secondList))
	copy(secondListCopy, secondList)

	secondListAfterRemove := removeDupplicates(firstList, secondList)

	assert.Equal(t, firstListCopy, firstList)
	assert.Equal(t, len(secondListCopy)-len(validatorsFromFirstList), len(secondListAfterRemove))
	assert.Equal(t, secondListCopy[:20], secondListAfterRemove)
}

func Test_removeDupplicates_AllDupplicates(t *testing.T) {
	t.Parallel()

	firstList := generateValidatorList(30)
	secondList := make([]Validator, len(firstList))
	copy(secondList, firstList)

	firstListCopy := make([]Validator, len(firstList))
	copy(firstListCopy, firstList)
	secondListCopy := make([]Validator, len(secondList))
	copy(secondListCopy, secondList)

	secondListAfterRemove := removeDupplicates(firstList, secondList)

	assert.Equal(t, firstListCopy, firstList)
	assert.Equal(t, len(secondListCopy)-len(firstListCopy), len(secondListAfterRemove))
	assert.Equal(t, 0, len(secondListAfterRemove))
}

func Test_shuffleList(t *testing.T) {
	t.Parallel()

	randomness := generateRandomByteArray(32)
	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, 0)
	validatorsCopy = append(validatorsCopy, validators...)

	shuffled := shuffleList(validators, randomness)
	assert.Equal(t, len(validatorsCopy), len(shuffled))
	assert.NotEqual(t, validatorsCopy, shuffled)
	assert.True(t, contains(shuffled, validatorsCopy))
}

func Test_shuffleListParameterNotChanged(t *testing.T) {
	t.Parallel()

	randomness := generateRandomByteArray(32)
	validators := generateValidatorList(30)
	validatorsCopy := make([]Validator, len(validators))
	_ = copy(validatorsCopy, validators)

	_ = shuffleList(validators, randomness)
	assert.Equal(t, validatorsCopy, validators)
}

func Test_shuffleListConsistentShuffling(t *testing.T) {
	t.Parallel()

	randomness := generateRandomByteArray(32)
	validators := generateValidatorList(30)

	nbTrials := 10
	shuffled := shuffleList(validators, randomness)
	for i := 0; i < nbTrials; i++ {
		shuffled2 := shuffleList(validators, randomness)
		assert.Equal(t, shuffled, shuffled2)
	}
}

func Test_equalizeValidatorsListsZeroToDistributeNoError(t *testing.T) {
	t.Parallel()
	valToGenerate := map[uint32]int{0: 20, 1: 0, core.MetachainShardId: 10}
	validatorsMap := make(map[uint32][]Validator)
	for i, nbValidators := range valToGenerate {
		validatorsMap[i] = generateValidatorList(nbValidators)
	}

	nbToDistribute := 0
	validatorsToDistribute := generateValidatorList(nbToDistribute)
	validatorsCopy := copyValidatorMap(validatorsMap)
	remainingToDistribute := equalizeValidatorsLists(validatorsMap, validatorsToDistribute)
	v, _ := removeValidatorsFromList(validatorsToDistribute, remainingToDistribute, len(remainingToDistribute))
	require.Equal(t, nbToDistribute, len(v))

	testDistributeValidators(t, validatorsCopy, validatorsMap, v)
}

func Test_equalizeValidatorsListsOneEmptyWaitingListNotEnoughToEqualize(t *testing.T) {
	t.Parallel()
	valToGenerate := map[uint32]int{0: 20, 1: 0, core.MetachainShardId: 10}
	validatorsMap := make(map[uint32][]Validator)
	for i, nbValidators := range valToGenerate {
		validatorsMap[i] = generateValidatorList(nbValidators)
	}

	nbToDistribute := 10
	validatorsToDistribute := generateValidatorList(nbToDistribute)
	validatorsCopy := copyValidatorMap(validatorsMap)
	remainingToDistribute := equalizeValidatorsLists(validatorsMap, validatorsToDistribute)
	v, _ := removeValidatorsFromList(validatorsToDistribute, remainingToDistribute, len(remainingToDistribute))
	require.Equal(t, nbToDistribute, len(v))

	testDistributeValidators(t, validatorsCopy, validatorsMap, v)
}

func Test_equalizeValidatorsListsOneEmptyWaitingListExactNumberToEqualize(t *testing.T) {
	t.Parallel()
	valToGenerate := map[uint32]int{0: 20, 1: 0, core.MetachainShardId: 10}
	validatorsMap := make(map[uint32][]Validator)
	for i, nbValidators := range valToGenerate {
		validatorsMap[i] = generateValidatorList(nbValidators)
	}

	nbToDistribute := 30
	validatorsToDistribute := generateValidatorList(nbToDistribute)
	validatorsCopy := copyValidatorMap(validatorsMap)
	remainingToDistribute := equalizeValidatorsLists(validatorsMap, validatorsToDistribute)
	v, _ := removeValidatorsFromList(validatorsToDistribute, remainingToDistribute, len(remainingToDistribute))
	require.Equal(t, nbToDistribute, len(v))
	for _, nbValidators := range validatorsMap {
		require.Equal(t, 20, len(nbValidators))
	}

	testDistributeValidators(t, validatorsCopy, validatorsMap, v)
}

func Test_equalizeValidatorsListsUnbalancedListsEnoughToDistribute(t *testing.T) {
	t.Parallel()
	valToGenerate := map[uint32]int{0: 30, 1: 10, core.MetachainShardId: 20}
	validatorsMap := make(map[uint32][]Validator)
	for i, nbValidators := range valToGenerate {
		validatorsMap[i] = generateValidatorList(nbValidators)
	}

	nbLists := len(validatorsMap)
	newNodesPerShard := 20
	validatorsToDistribute := generateValidatorList(nbLists * newNodesPerShard)
	validatorsCopy := copyValidatorMap(validatorsMap)

	remainingToDistribute := equalizeValidatorsLists(validatorsMap, validatorsToDistribute)
	for _, nbValidators := range validatorsMap {
		require.Equal(t, 30, len(nbValidators))
	}

	v, _ := removeValidatorsFromList(validatorsToDistribute, remainingToDistribute, len(remainingToDistribute))

	testDistributeValidators(t, validatorsCopy, validatorsMap, v)
}

func Test_equalizeValidatorsListsUnbalancedListsNotEnoughToDistribute(t *testing.T) {
	t.Parallel()
	valToGenerate := map[uint32]int{0: 30, 1: 10, core.MetachainShardId: 20}
	validatorsMap := make(map[uint32][]Validator)
	for i, nbValidators := range valToGenerate {
		validatorsMap[i] = generateValidatorList(nbValidators)
	}

	validatorsToDistribute := generateValidatorList(25)
	validatorsCopy := copyValidatorMap(validatorsMap)

	remainingToDistribute := equalizeValidatorsLists(validatorsMap, validatorsToDistribute)
	v, _ := removeValidatorsFromList(validatorsToDistribute, remainingToDistribute, len(remainingToDistribute))

	testDistributeValidators(t, validatorsCopy, validatorsMap, v)
}

func Test_distributeValidatorsEqualNumber(t *testing.T) {
	t.Parallel()

	nbShards := uint32(2)
	randomness := generateRandomByteArray(32)
	nodesPerShard := 30
	newNodesPerShard := 10
	validatorsMap := generateValidatorMap(nodesPerShard, nbShards)
	validatorsCopy := copyValidatorMap(validatorsMap)

	nbLists := len(validatorsMap)
	validatorsToDistribute := generateValidatorList(nbLists * newNodesPerShard)
	err := distributeValidators(validatorsMap, validatorsToDistribute, randomness, false)
	testDistributeValidators(t, validatorsCopy, validatorsMap, validatorsToDistribute)
	assert.Nil(t, err)
}

func Test_distributeValidatorsEqualNumberNoMeta(t *testing.T) {
	t.Parallel()

	nbShards := uint32(2)
	randomness := generateRandomByteArray(32)
	nodesPerShard := 30
	newNodesPerShard := 10
	validatorsMap := make(map[uint32][]Validator)

	for i := uint32(0); i < nbShards; i++ {
		validatorsMap[i] = generateValidatorList(nodesPerShard)
	}

	validatorsCopy := copyValidatorMap(validatorsMap)

	nbLists := len(validatorsMap)
	validatorsToDistribute := generateValidatorList(nbLists * newNodesPerShard)
	err := distributeValidators(validatorsMap, validatorsToDistribute, randomness, false)
	testDistributeValidators(t, validatorsCopy, validatorsMap, validatorsToDistribute)
	assert.Nil(t, err)
	assert.Nil(t, validatorsMap[core.MetachainShardId])
}

func Test_distributeValidatorsEqualNumberConsistent(t *testing.T) {
	t.Parallel()

	nbShards := uint32(2)
	randomness := generateRandomByteArray(32)
	nodesPerShard := 30
	newNodesPerShard := 10
	validatorsMap := generateValidatorMap(nodesPerShard, nbShards)
	validatorsCopy := copyValidatorMap(validatorsMap)

	nbLists := len(validatorsMap)
	validatorsToDistribute := generateValidatorList(nbLists * newNodesPerShard)
	err := distributeValidators(validatorsMap, validatorsToDistribute, randomness, false)
	testDistributeValidators(t, validatorsCopy, validatorsMap, validatorsToDistribute)
	assert.Nil(t, err)

	err = distributeValidators(validatorsCopy, validatorsToDistribute, randomness, false)
	assert.Nil(t, err)
	for i := range validatorsCopy {
		assert.Equal(t, validatorsMap[i], validatorsCopy[i])
	}
}

func Test_distributeValidatorsUnequalNumber(t *testing.T) {
	t.Parallel()

	randomness := generateRandomByteArray(32)
	nodesPerShard := 30
	nbShards := uint32(2)
	validatorsMap := generateValidatorMap(nodesPerShard, nbShards)
	validatorsCopy := copyValidatorMap(validatorsMap)

	nbLists := len(validatorsMap)
	maxNewNodesPerShard := 10
	newNodes := nbLists*maxNewNodesPerShard - 1
	validatorsToDistribute := generateValidatorList(nbLists*newNodes - 1)
	err := distributeValidators(validatorsMap, validatorsToDistribute, randomness, false)
	assert.Nil(t, err)
	testDistributeValidators(t, validatorsCopy, validatorsMap, validatorsToDistribute)
}

func Test_distributeValidatorsUnequalNumberConsistent(t *testing.T) {
	t.Parallel()

	randomness := generateRandomByteArray(32)
	nodesPerShard := 30
	nbShards := uint32(2)
	validatorsMap := generateValidatorMap(nodesPerShard, nbShards)
	validatorsCopy := copyValidatorMap(validatorsMap)

	nbLists := len(validatorsMap)
	maxNewNodesPerShard := 10
	newNodes := nbLists*maxNewNodesPerShard - 1
	validatorsToDistribute := generateValidatorList(nbLists*newNodes - 1)
	err := distributeValidators(validatorsMap, validatorsToDistribute, randomness, false)
	assert.Nil(t, err)
	testDistributeValidators(t, validatorsCopy, validatorsMap, validatorsToDistribute)

	err = distributeValidators(validatorsCopy, validatorsToDistribute, randomness, false)
	assert.Nil(t, err)
	for i := range validatorsCopy {
		assert.Equal(t, validatorsMap[i], validatorsCopy[i])
	}
}

func Test_distributeValidatorsNilOrEmptyDestination(t *testing.T) {
	t.Parallel()

	randomness := generateRandomByteArray(32)
	nodesPerShard := 30
	nbShards := uint32(2)
	validatorsMap := generateValidatorMap(nodesPerShard, nbShards)

	nbLists := len(validatorsMap)
	maxNewNodesPerShard := 10
	newNodes := nbLists*maxNewNodesPerShard - 1
	validatorsToDistribute := generateValidatorList(nbLists*newNodes - 1)
	err := distributeValidators(nil, validatorsToDistribute, randomness, false)
	assert.Equal(t, ErrNilOrEmptyDestinationForDistribute, err)

	err = distributeValidators(make(map[uint32][]Validator), validatorsToDistribute, randomness, false)
	assert.Equal(t, ErrNilOrEmptyDestinationForDistribute, err)
}

func Test_shuffleOutNodesNoLeaving(t *testing.T) {
	t.Parallel()

	randomness := generateRandomByteArray(32)
	eligibleNodesPerShard := eligiblePerShard
	waitingNodesPerShard := 40
	nbShards := uint32(2)
	var leaving []Validator

	eligibleMap := generateValidatorMap(eligibleNodesPerShard, nbShards)
	waitingMap := generateValidatorMap(waitingNodesPerShard, nbShards)
	stillRemainingInLeaving := make([]Validator, 0)
	numToRemove := make(map[uint32]int)
	for shardId := range waitingMap {
		numToRemove[shardId] = len(waitingMap[shardId])
	}

	shuffledOut, newEligible := shuffleOutNodes(eligibleMap, numToRemove, randomness)
	shuffleOutList := make([]Validator, 0)
	for _, shuffledOutPerShard := range shuffledOut {
		shuffleOutList = append(shuffleOutList, shuffledOutPerShard...)
	}
	testShuffledOut(t, eligibleMap, waitingMap, newEligible, shuffleOutList, leaving, stillRemainingInLeaving)
}

func Test_shuffleOutNodesWithLeaving(t *testing.T) {
	t.Parallel()

	randomness := generateRandomByteArray(32)
	eligibleNodesPerShard := eligiblePerShard
	waitingNodesPerShard := 40
	nbShards := uint32(2)
	leaving := make([]Validator, 0)

	eligibleMap := generateValidatorMap(eligibleNodesPerShard, nbShards)
	waitingMap := generateValidatorMap(waitingNodesPerShard, nbShards)
	for _, valList := range eligibleMap {
		leaving = append(leaving, valList[:len(valList)/5]...)
	}

	numToRemove := make(map[uint32]int)
	for shardId := range waitingMap {
		numToRemove[shardId] = len(waitingMap[shardId])
	}

	copyEligibleMap := copyValidatorMap(eligibleMap)
	copyWaitingMap := copyValidatorMap(waitingMap)
	newEligible, _, stillRemainingInLeaving := removeLeavingNodesFromValidatorMaps(
		copyEligibleMap,
		copyWaitingMap,
		numToRemove,
		leaving,
		eligibleNodesPerShard,
		eligibleNodesPerShard,
		true)
	shuffledOut, newEligible := shuffleOutNodes(newEligible, numToRemove, randomness)
	shuffleOutList := make([]Validator, 0)
	for _, shuffledOutPerShard := range shuffledOut {
		shuffleOutList = append(shuffleOutList, shuffledOutPerShard...)
	}
	testShuffledOut(t, eligibleMap, waitingMap, newEligible, shuffleOutList, leaving, stillRemainingInLeaving)
}

func Test_shuffleOutNodesWithLeavingMoreThanWaiting(t *testing.T) {
	t.Parallel()

	randomness := generateRandomByteArray(32)
	eligibleNodesPerShard := eligiblePerShard
	waitingNodesPerShard := 40
	nbShards := uint32(2)
	leaving := make([]Validator, 0)

	eligibleMap := generateValidatorMap(eligibleNodesPerShard, nbShards)
	waitingMap := generateValidatorMap(waitingNodesPerShard, nbShards)
	for _, valList := range eligibleMap {
		leaving = append(leaving, valList[:len(valList)/2+1]...)
	}

	numToRemove := make(map[uint32]int)
	for shardId := range waitingMap {
		numToRemove[shardId] = len(waitingMap[shardId])
	}
	copyEligibleMap := copyValidatorMap(eligibleMap)
	copyWaitingMap := copyValidatorMap(waitingMap)

	newEligible, _, stillRemainingInLeaving := removeLeavingNodesFromValidatorMaps(
		copyEligibleMap,
		copyWaitingMap,
		numToRemove,
		leaving,
		eligibleNodesPerShard,
		eligibleNodesPerShard,
		true)

	shuffledOut, newEligible := shuffleOutNodes(newEligible, numToRemove, randomness)
	shuffleOutList := make([]Validator, 0)
	for _, shuffledOutPerShard := range shuffledOut {
		shuffleOutList = append(shuffleOutList, shuffledOutPerShard...)
	}
	testShuffledOut(t, eligibleMap, waitingMap, newEligible, shuffleOutList, leaving, stillRemainingInLeaving)
}

func Test_removeLeavingNodesFromValidatorMaps(t *testing.T) {
	t.Parallel()

	maxShuffleOutNumber := 20
	eligibleNodesPerShard := eligiblePerShard
	waitingNodesPerShard := 40
	nbShards := uint32(2)

	tests := []struct {
		waitingFixEnabled bool
		remainingToRemove int
	}{
		{
			waitingFixEnabled: false,
			remainingToRemove: 18,
		},
		{
			waitingFixEnabled: true,
			remainingToRemove: 20,
		},
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {

			leaving := make([]Validator, 0)

			eligibleMap := generateValidatorMap(eligibleNodesPerShard, nbShards)
			waitingMap := generateValidatorMap(waitingNodesPerShard, nbShards)
			for _, waitingValidators := range waitingMap {
				leaving = append(leaving, waitingValidators[:2]...)
			}

			numToRemove := make(map[uint32]int)

			for shardId := range waitingMap {
				numToRemove[shardId] = maxShuffleOutNumber
			}
			copyEligibleMap := copyValidatorMap(eligibleMap)
			copyWaitingMap := copyValidatorMap(waitingMap)

			_, _, _ = removeLeavingNodesFromValidatorMaps(
				copyEligibleMap,
				copyWaitingMap,
				numToRemove,
				leaving,
				eligibleNodesPerShard,
				eligibleNodesPerShard,
				tt.waitingFixEnabled,
			)

			for _, remainingToRemove := range numToRemove {
				require.Equal(t, tt.remainingToRemove, remainingToRemove)
			}
		})
	}
}

func TestNewHashValidatorsShuffler(t *testing.T) {
	t.Parallel()

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           eligiblePerShard,
		NodesMeta:            eligiblePerShard,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	assert.Nil(t, err)
	assert.NotNil(t, shuffler)
}

func TestRandHashShuffler_computeNewShardsNotChanging(t *testing.T) {
	t.Parallel()

	currentNbShards := uint32(3)
	shuffler, err := createHashShufflerInter()
	require.Nil(t, err)

	eligible := generateValidatorMap(int(shuffler.nodesShard), currentNbShards)
	nbShards := currentNbShards + 1 // account for meta
	maxNodesNoSplit := (nbShards + 1) * (shuffler.nodesShard + shuffler.shardHysteresis)
	nbWaitingPerShard := int(maxNodesNoSplit/nbShards - shuffler.nodesShard)
	waiting := generateValidatorMap(nbWaitingPerShard, currentNbShards)
	newNodes := generateValidatorList(0)
	leavingUnstake := generateValidatorList(0)
	leavingRating := generateValidatorList(0)

	numNewNodes := len(newNodes)
	numLeaving := len(leavingUnstake) + len(leavingRating)

	newNbShards := shuffler.computeNewShards(eligible, waiting, numNewNodes, numLeaving, currentNbShards)
	assert.Equal(t, currentNbShards, newNbShards)
}

func TestRandHashShuffler_computeNewShardsWithSplit(t *testing.T) {
	t.Parallel()

	currentNbShards := uint32(3)
	shuffler, err := createHashShufflerInter()
	require.Nil(t, err)

	eligible := generateValidatorMap(int(shuffler.nodesShard), currentNbShards)
	nbShards := currentNbShards + 1 // account for meta
	maxNodesNoSplit := (nbShards + 1) * (shuffler.nodesShard + shuffler.shardHysteresis)
	nbWaitingPerShard := int(maxNodesNoSplit/nbShards-shuffler.nodesShard) + 1
	waiting := generateValidatorMap(nbWaitingPerShard, currentNbShards)
	newNodes := generateValidatorList(0)
	leavingUnstake := generateValidatorList(0)
	leavingRating := generateValidatorList(0)

	numNewNodes := len(newNodes)
	numLeaving := len(leavingUnstake) + len(leavingRating)

	newNbShards := shuffler.computeNewShards(eligible, waiting, numNewNodes, numLeaving, currentNbShards)
	assert.Equal(t, currentNbShards+1, newNbShards)
}

func TestRandHashShuffler_computeNewShardsWithMerge(t *testing.T) {
	t.Parallel()

	currentNbShards := uint32(3)
	shuffler, err := createHashShufflerInter()
	require.Nil(t, err)

	eligible := generateValidatorMap(int(shuffler.nodesShard), currentNbShards)
	nbWaitingPerShard := 0
	waiting := generateValidatorMap(nbWaitingPerShard, currentNbShards)
	newNodes := generateValidatorList(0)
	leavingUnstake := generateValidatorList(1)
	leavingRating := generateValidatorList(0)

	numNewNodes := len(newNodes)
	numLeaving := len(leavingUnstake) + len(leavingRating)

	newNbShards := shuffler.computeNewShards(eligible, waiting, numNewNodes, numLeaving, currentNbShards)
	assert.Equal(t, currentNbShards-1, newNbShards)
}

func TestRandHashShuffler_UpdateParams(t *testing.T) {
	t.Parallel()

	shuffler, err := createHashShufflerInter()
	require.Nil(t, err)

	shuffler2 := &randHashShuffler{
		nodesShard:            200,
		nodesMeta:             200,
		shardHysteresis:       0,
		metaHysteresis:        0,
		adaptivity:            true,
		shuffleBetweenShards:  true,
		validatorDistributor:  &CrossShardValidatorDistributor{},
		availableNodesConfigs: nil,
	}

	shuffler.UpdateParams(
		shuffler2.nodesShard,
		shuffler2.nodesMeta,
		0,
		shuffler2.adaptivity,
	)

	assert.Equal(t, shuffler2, shuffler)
}

func TestRandHashShuffler_UpdateNodeListsNoReSharding(t *testing.T) {
	t.Parallel()

	shuffler, err := createHashShufflerInter()
	require.Nil(t, err)

	eligiblePerShard := int(shuffler.nodesShard)
	waitingPerShard := 30
	nbShards := uint32(3)
	randomness := generateRandomByteArray(32)

	leavingNodes := make([]Validator, 0)
	extraLeavingNodes := make([]Validator, 0)
	newNodes := make([]Validator, 0)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	args := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          newNodes,
		UnStakeLeaving:    leavingNodes,
		AdditionalLeaving: extraLeavingNodes,
		Rand:              randomness,
		NbShards:          nbShards,
	}

	resUpdateNodeList, err := shuffler.UpdateNodeLists(args)
	require.Nil(t, err)

	allPrevEligible := getValidatorsInMap(eligibleMap)
	allNewEligible := getValidatorsInMap(resUpdateNodeList.Eligible)
	allPrevWaiting := getValidatorsInMap(waitingMap)
	allNewWaiting := getValidatorsInMap(resUpdateNodeList.Waiting)

	assert.Equal(t, len(allPrevEligible)+len(allPrevWaiting), len(allNewEligible)+len(allNewWaiting))
}

func TestRandHashShuffler_UpdateNodeListsWithUnstakeLeavingRemovesFromEligible(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 10
	eligibleMeta := 10

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           uint32(eligiblePerShard),
		NodesMeta:            uint32(eligibleMeta),
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}

	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	waitingPerShard := 2
	nbShards := 0

	args := createShufflerArgs(eligiblePerShard, waitingPerShard, uint32(nbShards))

	args.UnStakeLeaving = []Validator{
		args.Eligible[core.MetachainShardId][0],
		args.Eligible[core.MetachainShardId][1],
	}

	resUpdateNodeList, err := shuffler.UpdateNodeLists(args)
	require.Nil(t, err)

	allNewEligible := getValidatorsInMap(resUpdateNodeList.Eligible)
	allNewWaiting := getValidatorsInMap(resUpdateNodeList.Waiting)

	allNewEligibleMap := listToValidatorMap(allNewEligible)

	for _, leavingValidator := range resUpdateNodeList.Leaving {
		assert.Nil(t, allNewEligibleMap[string(leavingValidator.PubKey())])
	}

	previousNumberOfNodes := (eligiblePerShard + waitingPerShard) * (nbShards + 1)
	currentNumberOfNodes := len(allNewEligible) + len(allNewWaiting) + len(resUpdateNodeList.Leaving)
	assert.Equal(t, previousNumberOfNodes, currentNumberOfNodes)
}

func TestRandHashShuffler_UpdateNodeListsWaitingListFixDisabled(t *testing.T) {
	t.Parallel()

	testUpdateNodesAndCheckNumLeaving(t, true)
}

func TestRandHashShuffler_UpdateNodeListsWithWaitingListFixEnabled(t *testing.T) {
	t.Parallel()

	testUpdateNodesAndCheckNumLeaving(t, false)
}

func testUpdateNodesAndCheckNumLeaving(t *testing.T, beforeFix bool) {
	eligiblePerShard := 400
	eligibleMeta := 10

	waitingPerShard := 400
	nbShards := 1

	numLeaving := 200

	numNodesToShuffle := 80

	waitingListFixEnableEpoch := 0
	if beforeFix {
		waitingListFixEnableEpoch = 9999
	}

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           uint32(eligiblePerShard),
		NodesMeta:            uint32(eligibleMeta),
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: []config.MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            100,
				NodesToShufflePerShard: uint32(numNodesToShuffle),
			},
		},
		WaitingListFixEnableEpoch: uint32(waitingListFixEnableEpoch),
	}

	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	args := createShufflerArgs(eligiblePerShard, waitingPerShard, uint32(nbShards))

	for i := 0; i < numLeaving; i++ {
		args.UnStakeLeaving = append(args.UnStakeLeaving, args.Waiting[0][i])
	}

	resUpdateNodeList, err := shuffler.UpdateNodeLists(args)
	require.Nil(t, err)

	eligibleInShard0 := resUpdateNodeList.Eligible[0]
	assert.Equal(t, numValidatorsInEligibleList, len(eligibleInShard0))

	expectedNumLeaving := numLeaving
	if beforeFix {
		expectedNumLeaving = numNodesToShuffle
	}
	assert.Equal(t, expectedNumLeaving, len(resUpdateNodeList.Leaving))

	for i := 0; i < numNodesToShuffle; i++ {
		assert.Equal(t, resUpdateNodeList.Leaving[i], args.UnStakeLeaving[i])
	}
}

func TestRandHashShuffler_UpdateNodeListsWaitingListWithFixCheckWaitingDisabled(t *testing.T) {
	t.Parallel()

	testUpdateNodeListsAndCheckWaitingList(t, true)
}

func TestRandHashShuffler_UpdateNodeListsWaitingListWithFixCheckWaitingEnabled(t *testing.T) {
	t.Parallel()

	testUpdateNodeListsAndCheckWaitingList(t, false)
}

func testUpdateNodeListsAndCheckWaitingList(t *testing.T, beforeFix bool) {
	eligiblePerShard := 400
	eligibleMeta := 10

	waitingPerShard := 400
	nbShards := 1

	numLeaving := 2

	numNodesToShuffle := 80

	waitingListFixEnableEpoch := 0
	if beforeFix {
		waitingListFixEnableEpoch = 9999
	}

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           uint32(eligiblePerShard),
		NodesMeta:            uint32(eligibleMeta),
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: []config.MaxNodesChangeConfig{
			{
				EpochEnable:            0,
				MaxNumNodes:            100,
				NodesToShufflePerShard: uint32(numNodesToShuffle),
			},
		},
		WaitingListFixEnableEpoch: uint32(waitingListFixEnableEpoch),
	}

	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	args := createShufflerArgs(eligiblePerShard, waitingPerShard, uint32(nbShards))

	initialWaitingListCopy := make(map[uint32][]Validator, len(args.Waiting))
	for shID, val := range args.Waiting {
		initialWaitingListCopy[shID] = make([]Validator, len(val))
		copy(initialWaitingListCopy[shID], args.Waiting[shID])
	}

	for i := 0; i < numLeaving; i++ {
		args.UnStakeLeaving = append(args.UnStakeLeaving, args.Waiting[0][i])
	}

	resUpdateNodeList, err := shuffler.UpdateNodeLists(args)
	require.Nil(t, err)

	eligibleInShard0 := resUpdateNodeList.Eligible[0]
	assert.Equal(t, numValidatorsInEligibleList, len(eligibleInShard0))

	numWaitingListToEligible := 0
	for _, waiting := range initialWaitingListCopy[0] {
		for _, eligible := range resUpdateNodeList.Eligible[0] {
			if bytes.Equal(waiting.PubKey(), eligible.PubKey()) {
				numWaitingListToEligible++
			}
		}
	}

	expectedNumWaitingMovedToEligible := numNodesToShuffle
	if beforeFix {
		expectedNumWaitingMovedToEligible -= numLeaving
	}

	assert.Equal(t, expectedNumWaitingMovedToEligible, numWaitingListToEligible)
}

func TestRandHashShuffler_UpdateNodeListsWithUnstakeLeavingRemovesFromWaiting(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 10
	eligibleMeta := 10

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           uint32(eligiblePerShard),
		NodesMeta:            uint32(eligibleMeta),
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}

	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	waitingPerShard := 2
	nbShards := 0

	args := createShufflerArgs(eligiblePerShard, waitingPerShard, uint32(nbShards))

	args.UnStakeLeaving = []Validator{
		args.Waiting[core.MetachainShardId][0],
		args.Waiting[core.MetachainShardId][1],
	}

	resUpdateNodeList, err := shuffler.UpdateNodeLists(args)
	require.Nil(t, err)

	allNewEligible := getValidatorsInMap(resUpdateNodeList.Eligible)
	allNewWaiting := getValidatorsInMap(resUpdateNodeList.Waiting)

	allNewWaitingMap := listToValidatorMap(allNewWaiting)

	for _, leavingValidator := range resUpdateNodeList.Leaving {
		assert.Nil(t, allNewWaitingMap[string(leavingValidator.PubKey())])
	}

	previousNumberOfNodes := (eligiblePerShard + waitingPerShard) * (nbShards + 1)
	currentNumberOfNodes := len(allNewEligible) + len(allNewWaiting) + len(resUpdateNodeList.Leaving)
	assert.Equal(t, previousNumberOfNodes, currentNumberOfNodes)
}

func TestRandHashShuffler_UpdateNodeListsWithNonExistentUnstakeLeavingDoesNotRemove(t *testing.T) {
	t.Parallel()

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           10,
		NodesMeta:            10,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	eligiblePerShard := int(shuffler.nodesShard)
	waitingPerShard := 2
	nbShards := uint32(0)

	args := createShufflerArgs(eligiblePerShard, waitingPerShard, nbShards)

	args.UnStakeLeaving = []Validator{
		&validator{
			pubKey: generateRandomByteArray(32),
		},
		&validator{
			pubKey: generateRandomByteArray(32),
		},
	}

	resUpdateNodeList, err := shuffler.UpdateNodeLists(args)
	require.Nil(t, err)

	allNewEligible := getValidatorsInMap(resUpdateNodeList.Eligible)
	allNewEligibleMap := listToValidatorMap(allNewEligible)
	allNewWaiting := getValidatorsInMap(resUpdateNodeList.Waiting)
	allNewWaitingMap := listToValidatorMap(allNewWaiting)

	for _, leavingValidator := range resUpdateNodeList.Leaving {
		assert.Nil(t, allNewEligibleMap[string(leavingValidator.PubKey())])
	}
	for _, leavingValidator := range resUpdateNodeList.Leaving {
		assert.Nil(t, allNewWaitingMap[string(leavingValidator.PubKey())])
	}

	for i := range resUpdateNodeList.Leaving {
		assert.Equal(t, args.UnStakeLeaving[i], resUpdateNodeList.Leaving[i])
	}
}

func TestRandHashShuffler_UpdateNodeListsWithRangeOnMaps(t *testing.T) {
	t.Parallel()
	shuffleBetweenShards := []bool{true, false}

	for _, shuffle := range shuffleBetweenShards {
		shufflerArgs := &NodesShufflerArgs{
			NodesShard:           uint32(eligiblePerShard),
			NodesMeta:            uint32(eligiblePerShard),
			Hysteresis:           hysteresis,
			Adaptivity:           adaptivity,
			ShuffleBetweenShards: shuffle,
			MaxNodesEnableConfig: nil,
		}

		shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
		require.Nil(t, err)

		waitingPerShard := 20
		numShards := 2

		numLeaving := (waitingPerShard + waitingPerShard/2) * (numShards + 1)

		args := createShufflerArgs(eligiblePerShard, waitingPerShard, uint32(numShards))

		allValidators := make([]Validator, 0)

		for _, shardValidators := range args.Eligible {
			allValidators = append(allValidators, shardValidators...)
		}

		for _, shardValidators := range args.Waiting {
			allValidators = append(allValidators, shardValidators...)
		}

		leavingValidators := make([]Validator, numLeaving)

		for i := 0; i < numLeaving; i++ {
			randIndex := mathRand.Intn(len(allValidators))
			leavingValidators[i] = allValidators[randIndex]
			allValidators = removeValidatorFromList(allValidators, randIndex)
		}

		args.UnStakeLeaving = leavingValidators

		resUpdateNodeListInitial, err := shuffler.UpdateNodeLists(args)
		require.Nil(t, err)

		for i := 0; i < 100; i++ {
			resUpdateNodeList, err := shuffler.UpdateNodeLists(args)
			require.Nil(t, err)

			assert.Equal(t, resUpdateNodeListInitial.Eligible, resUpdateNodeList.Eligible)
			assert.Equal(t, resUpdateNodeListInitial.Waiting, resUpdateNodeList.Waiting)
			assert.Equal(t, resUpdateNodeList.Leaving, resUpdateNodeList.Leaving)
		}
	}
}

func TestRandHashShuffler_UpdateNodeListsNoReShardingIntraShardShuffling(t *testing.T) {
	t.Parallel()

	nbShards := uint32(3)
	randomness := generateRandomByteArray(32)

	shuffler, err := createHashShufflerIntraShards()
	require.Nil(t, err)

	leavingNodes := make([]Validator, 0)
	additionalLeavingNodes := make([]Validator, 0)
	newNodes := make([]Validator, 0)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	args := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          newNodes,
		UnStakeLeaving:    leavingNodes,
		AdditionalLeaving: additionalLeavingNodes,
		Rand:              randomness,
		NbShards:          nbShards,
	}

	resUpdateNodeList, err := shuffler.UpdateNodeLists(args)
	require.Nil(t, err)

	allPrevEligible := getValidatorsInMap(eligibleMap)
	allNewEligible := getValidatorsInMap(resUpdateNodeList.Eligible)
	allPrevWaiting := getValidatorsInMap(waitingMap)
	allNewWaiting := getValidatorsInMap(resUpdateNodeList.Waiting)

	assert.Equal(t, len(allPrevEligible)+len(allPrevWaiting), len(allNewEligible)+len(allNewWaiting))
}

func TestRandHashShuffler_RemoveLeavingNodesNotExistingInEligibleOrWaiting_WithAllKnown(t *testing.T) {
	t.Parallel()

	nbShards := uint32(3)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	leavingValidators := []Validator{
		eligibleMap[core.MetachainShardId][0],
		eligibleMap[core.MetachainShardId][1],
		waitingMap[0][1],
		waitingMap[0][2],
	}

	stillRemaining, removed := removeLeavingNodesNotExistingInEligibleOrWaiting(leavingValidators, eligibleMap, waitingMap)

	assert.Equal(t, len(leavingValidators), len(stillRemaining))
	assert.Equal(t, 0, len(removed))

	for _, leavingValidator := range leavingValidators {
		foundInEligible, _ := searchInMap(eligibleMap, leavingValidator.PubKey())
		foundInWaiting, _ := searchInMap(waitingMap, leavingValidator.PubKey())
		assert.True(t, foundInEligible || foundInWaiting)
	}

}

func TestRandHashShuffler_RemoveLeavingNodesNotExistingInEligibleOrWaiting_WithAllNotKnown(t *testing.T) {
	t.Parallel()

	leavingNumber := 4
	nbShards := uint32(3)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	leavingValidators := generateValidatorList(leavingNumber)

	stillRemaining, removed := removeLeavingNodesNotExistingInEligibleOrWaiting(leavingValidators, eligibleMap, waitingMap)

	assert.Equal(t, 0, len(stillRemaining))
	assert.Equal(t, len(leavingValidators), len(removed))

	for _, leavingValidator := range leavingValidators {
		foundInEligible, _ := searchInMap(eligibleMap, leavingValidator.PubKey())
		foundInWaiting, _ := searchInMap(waitingMap, leavingValidator.PubKey())
		assert.False(t, foundInEligible || foundInWaiting)
	}

	for i := range leavingValidators {
		assert.Equal(t, leavingValidators[i], removed[i])
	}
}

func TestRandHashShuffler_RemoveLeavingNodesNotExistingInEligibleOrWaiting_3Found3NotFound(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 10
	waitingPerShard := 3
	leavingNumber := 3
	nbShards := uint32(3)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	leavingValidators := generateValidatorList(leavingNumber)
	leavingValidators = append(leavingValidators, eligibleMap[core.MetachainShardId][0])
	leavingValidators = append(leavingValidators, eligibleMap[0][eligiblePerShard/2])
	leavingValidators = append(leavingValidators, waitingMap[1][waitingPerShard-1])

	stillRemaining, removed := removeLeavingNodesNotExistingInEligibleOrWaiting(leavingValidators, eligibleMap, waitingMap)

	assert.Equal(t, 3, len(stillRemaining))
	assert.Equal(t, 3, len(removed))

	for i := range leavingValidators[:3] {
		assert.Equal(t, leavingValidators[i], removed[i])
	}

	for _, leavingValidator := range leavingValidators[3:] {
		foundInEligible, _ := searchInMap(eligibleMap, leavingValidator.PubKey())
		foundInWaiting, _ := searchInMap(waitingMap, leavingValidator.PubKey())
		assert.True(t, foundInEligible || foundInWaiting)
	}
}

func TestRandHashShuffler_RemoveLeavingNodesFromValidatorMaps_FromEligible(t *testing.T) {
	t.Parallel()

	nbShards := uint32(0)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	leavingValidators := []Validator{
		eligibleMap[core.MetachainShardId][0],
	}

	numToRemove := make(map[uint32]int)
	for shardId := range waitingMap {
		numToRemove[shardId] = waitingPerShard
	}

	eligibleCopy := copyValidatorMap(eligibleMap)
	waitingCopy := copyValidatorMap(waitingMap)

	newEligible, newWaiting, stillRemaining := removeLeavingNodesFromValidatorMaps(
		eligibleCopy,
		waitingCopy,
		numToRemove,
		leavingValidators,
		eligiblePerShard,
		eligiblePerShard,
		true)

	assert.Equal(t, eligiblePerShard-1, len(newEligible[core.MetachainShardId]))
	assert.Equal(t, waitingPerShard, len(newWaiting[core.MetachainShardId]))
	assert.Equal(t, 0, len(stillRemaining))

	for _, leavingValidator := range leavingValidators {
		foundInEligible, _ := searchInMap(eligibleMap, leavingValidator.PubKey())
		assert.True(t, foundInEligible)
		foundInEligible, _ = searchInMap(newEligible, leavingValidator.PubKey())
		assert.False(t, foundInEligible)
	}
}

func TestRandHashShuffler_RemoveLeavingNodesFromValidatorMaps_FromWaiting(t *testing.T) {
	t.Parallel()

	nbShards := uint32(0)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	leavingValidators := []Validator{
		waitingMap[core.MetachainShardId][1],
	}

	numToRemove := make(map[uint32]int)
	for shardId := range waitingMap {
		numToRemove[shardId] = waitingPerShard
	}

	eligibleCopy := copyValidatorMap(eligibleMap)
	waitingCopy := copyValidatorMap(waitingMap)

	newEligible, newWaiting, stillRemaining := removeLeavingNodesFromValidatorMaps(
		eligibleCopy,
		waitingCopy,
		numToRemove,
		leavingValidators,
		eligiblePerShard,
		eligiblePerShard,
		true)

	assert.Equal(t, eligiblePerShard, len(newEligible[core.MetachainShardId]))
	assert.Equal(t, waitingPerShard-1, len(newWaiting[core.MetachainShardId]))
	assert.Equal(t, 0, len(stillRemaining))

	for _, leavingValidator := range leavingValidators {
		foundInEligible, _ := searchInMap(waitingMap, leavingValidator.PubKey())
		assert.True(t, foundInEligible)
		foundInEligible, _ = searchInMap(newWaiting, leavingValidator.PubKey())
		assert.False(t, foundInEligible)
	}
}

func TestRandHashShuffler_RemoveLeavingNodesFromValidatorMaps_NonExisting(t *testing.T) {
	t.Parallel()

	nbShards := uint32(0)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	leavingValidators := generateValidatorList(1)

	numToRemove := make(map[uint32]int)
	for shardId := range waitingMap {
		numToRemove[shardId] = waitingPerShard
	}

	eligibleCopy := copyValidatorMap(eligibleMap)
	waitingCopy := copyValidatorMap(waitingMap)

	newEligible, newWaiting, stillRemaining := removeLeavingNodesFromValidatorMaps(
		eligibleCopy,
		waitingCopy,
		numToRemove,
		leavingValidators,
		eligiblePerShard,
		eligiblePerShard,
		true)

	assert.Equal(t, eligiblePerShard, len(newEligible[core.MetachainShardId]))
	assert.Equal(t, waitingPerShard, len(newWaiting[core.MetachainShardId]))
	assert.Equal(t, len(leavingValidators), len(stillRemaining))

	for _, leavingValidator := range leavingValidators {
		foundInEligible, _ := searchInMap(newEligible, leavingValidator.PubKey())
		foundInWaiting, _ := searchInMap(newWaiting, leavingValidator.PubKey())
		assert.False(t, foundInEligible || foundInWaiting)
	}
}

func TestRandHashShuffler_RemoveLeavingNodesFromValidatorMaps_2Eligible2Waiting2NonExisting(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 10
	waitingPerShard := 6
	nbShards := uint32(0)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	leavingValidators := generateValidatorList(2)
	leavingValidators = append(leavingValidators, []Validator{
		eligibleMap[core.MetachainShardId][0],
		eligibleMap[core.MetachainShardId][5]}...)
	leavingValidators = append(leavingValidators, []Validator{
		waitingMap[core.MetachainShardId][1],
		waitingMap[core.MetachainShardId][2]}...)

	numToRemove := make(map[uint32]int)
	for shardId := range waitingMap {
		numToRemove[shardId] = waitingPerShard
	}

	eligibleCopy := copyValidatorMap(eligibleMap)
	waitingCopy := copyValidatorMap(waitingMap)

	newEligible, newWaiting, stillRemaining := removeLeavingNodesFromValidatorMaps(
		eligibleCopy,
		waitingCopy,
		numToRemove,
		leavingValidators,
		eligiblePerShard,
		eligiblePerShard,
		true)

	remainingInEligible := eligiblePerShard - 2
	remainingInWaiting := waitingPerShard - 2
	stillLeaving := 2

	assert.Equal(t, remainingInEligible, len(newEligible[core.MetachainShardId]))
	assert.Equal(t, remainingInWaiting, len(newWaiting[core.MetachainShardId]))
	assert.Equal(t, stillLeaving, len(stillRemaining))

	for _, leavingValidator := range leavingValidators[:2] {
		foundInEligible, _ := searchInMap(newEligible, leavingValidator.PubKey())
		foundInWaiting, _ := searchInMap(newWaiting, leavingValidator.PubKey())
		assert.False(t, foundInEligible || foundInWaiting)
	}

	for _, leavingValidator := range leavingValidators[2:4] {
		foundInWaiting, _ := searchInMap(newWaiting, leavingValidator.PubKey())
		assert.False(t, foundInWaiting)
	}

	for _, leavingValidator := range leavingValidators[4:6] {
		foundInEligible, _ := searchInMap(newEligible, leavingValidator.PubKey())
		assert.False(t, foundInEligible)
	}
}

func TestRandHashShuffler_RemoveLeavingNodesFromValidatorMaps_2FromEligible2FromWaitingFrom3Waiting(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 10
	waitingPerShard := 3
	nbShards := uint32(0)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	leavingValidators := make([]Validator, 0)
	leavingValidators = append(leavingValidators, []Validator{
		eligibleMap[core.MetachainShardId][0],
		eligibleMap[core.MetachainShardId][5]}...)
	leavingValidators = append(leavingValidators, []Validator{
		waitingMap[core.MetachainShardId][1],
		waitingMap[core.MetachainShardId][2]}...)

	numToRemove := make(map[uint32]int)
	for shardId := range waitingMap {
		numToRemove[shardId] = waitingPerShard
	}

	eligibleCopy := copyValidatorMap(eligibleMap)
	waitingCopy := copyValidatorMap(waitingMap)

	newEligible, newWaiting, stillRemaining := removeLeavingNodesFromValidatorMaps(
		eligibleCopy,
		waitingCopy,
		numToRemove,
		leavingValidators,
		eligiblePerShard,
		eligiblePerShard,
		true)

	// removed first 2 from waiting and just one from eligible
	remainingInEligible := eligiblePerShard - 1
	remainingInWaiting := waitingPerShard - 2
	stillLeaving := 1

	assert.Equal(t, remainingInEligible, len(newEligible[core.MetachainShardId]))
	assert.Equal(t, remainingInWaiting, len(newWaiting[core.MetachainShardId]))
	assert.Equal(t, stillLeaving, len(stillRemaining))

	for _, leavingValidator := range leavingValidators[2:4] {
		foundInWaiting, _ := searchInMap(newWaiting, leavingValidator.PubKey())
		assert.False(t, foundInWaiting)
	}

	foundInEligible, _ := searchInMap(newEligible, leavingValidators[0].PubKey())
	assert.False(t, foundInEligible)

	foundInEligible, _ = searchInMap(newEligible, leavingValidators[1].PubKey())
	assert.True(t, foundInEligible)
}

func TestRandHashShuffler_UpdateNodeLists_WithUnstakeLeaving(t *testing.T) {
	t.Parallel()

	nbShards := uint32(2)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	unStakeLeaving := make(map[uint32][]Validator)
	additionalLeaving := make(map[uint32][]Validator)
	unStakeLeaving[core.MetachainShardId] = []Validator{
		eligibleMap[core.MetachainShardId][0],
		eligibleMap[core.MetachainShardId][eligiblePerShard-1],
		waitingMap[core.MetachainShardId][0],
		waitingMap[core.MetachainShardId][1],
	}

	unStakeLeaving[0] = []Validator{
		waitingMap[0][0],
		waitingMap[0][waitingPerShard/2]}

	unStakeLeaving[1] = []Validator{
		eligibleMap[1][0],
		eligibleMap[1][1],
		eligibleMap[1][2],
		eligibleMap[1][eligiblePerShard-1]}

	additionaLeavingList := make([]Validator, 0)
	for _, shardLeaving := range unStakeLeaving {
		additionaLeavingList = append(additionaLeavingList, shardLeaving...)
	}

	shuffler, err := createHashShufflerIntraShards()
	require.Nil(t, err)

	arg := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          make([]Validator, 0),
		UnStakeLeaving:    additionaLeavingList,
		AdditionalLeaving: make([]Validator, 0),
		Rand:              generateRandomByteArray(32),
		NbShards:          nbShards,
	}

	result, err := shuffler.UpdateNodeLists(arg)
	require.Nil(t, err)

	leavingPerShardMap, stillRemainingPerShardMap := createActuallyLeavingPerShards(unStakeLeaving, additionalLeaving, result.Leaving)

	for i := uint32(0); i < nbShards+1; i++ {
		shardId := i
		if i == nbShards {
			shardId = core.MetachainShardId
		}
		verifyResultsIntraShardShuffling(t,
			eligibleMap[shardId],
			waitingMap[shardId],
			unStakeLeaving[shardId],
			additionalLeaving[shardId],
			result.Eligible[shardId],
			result.Waiting[shardId],
			leavingPerShardMap[shardId],
			stillRemainingPerShardMap[shardId],
		)
	}
}

func TestRandHashShuffler_UpdateNodeLists_WithUnstakeLeaving_EnoughRemaining(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 100
	waitingPerShard := 30
	nbShards := uint32(2)

	shuffler, err := createHashShufflerIntraShards()
	require.Nil(t, err)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)
	currentShardId := uint32(1)
	unstakeLeaving := eligibleMap[currentShardId]                  // unstake 100
	eligibleMap[currentShardId] = eligibleMap[currentShardId][:70] // eligible - 70

	arg := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          make([]Validator, 0),
		UnStakeLeaving:    unstakeLeaving,
		AdditionalLeaving: make([]Validator, 0),
		Rand:              generateRandomByteArray(32),
		NbShards:          nbShards,
	}

	result, err := shuffler.UpdateNodeLists(arg)
	assert.NotNil(t, result)
	assert.Nil(t, err)
}

func TestRandHashShuffler_UpdateNodeLists_WithUnstakeLeaving_NotEnoughRemaining(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 100
	waitingPerShard := 30
	nbShards := uint32(2)

	shuffler, err := createHashShufflerIntraShards()
	require.Nil(t, err)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	unstakeLeaving := eligibleMap[core.MetachainShardId] // unstake 100

	eligibleMap[core.MetachainShardId] = make([]Validator, 0) // eligible - no validators

	arg := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          make([]Validator, 0),
		UnStakeLeaving:    unstakeLeaving,
		AdditionalLeaving: make([]Validator, 0),
		Rand:              generateRandomByteArray(32),
		NbShards:          nbShards,
	}

	_, err = shuffler.UpdateNodeLists(arg)
	assert.True(t, errors.Is(err, ErrSmallShardEligibleListSize))
	assert.True(t, strings.Contains(err.Error(), fmt.Sprintf("%v", core.MetachainShardId)))

	eligibleMap = generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap = generateValidatorMap(waitingPerShard, nbShards)

	currentShardId := uint32(0)
	unstakeLeaving = eligibleMap[currentShardId]                   // unstake 100
	eligibleMap[currentShardId] = eligibleMap[currentShardId][:69] // eligible - 69

	arg = ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          make([]Validator, 0),
		UnStakeLeaving:    unstakeLeaving,
		AdditionalLeaving: make([]Validator, 0),
		Rand:              generateRandomByteArray(32),
		NbShards:          uint32(len(eligibleMap)),
	}

	_, err = shuffler.UpdateNodeLists(arg)
	assert.True(t, errors.Is(err, ErrSmallShardEligibleListSize))
	assert.True(t, strings.Contains(err.Error(), fmt.Sprintf("%v", currentShardId)))
}

func TestRandHashShuffler_UpdateNodeLists_WithAdditionalLeaving(t *testing.T) {
	t.Parallel()

	nbShards := uint32(2)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	additionalLeaving := make(map[uint32][]Validator)
	unstakeLeaving := make(map[uint32][]Validator)
	additionalLeaving[core.MetachainShardId] = []Validator{
		eligibleMap[core.MetachainShardId][0],
		eligibleMap[core.MetachainShardId][eligiblePerShard-1],
		waitingMap[core.MetachainShardId][0],
		waitingMap[core.MetachainShardId][1],
	}

	additionalLeaving[0] = []Validator{
		waitingMap[0][0],
		waitingMap[0][waitingPerShard/2]}

	additionalLeaving[1] = []Validator{
		eligibleMap[1][0],
		eligibleMap[1][1],
		eligibleMap[1][2],
		eligibleMap[1][eligiblePerShard-1]}

	unstakeLeavingList, additionalLeavingList := prepareListsFromMaps(unstakeLeaving, additionalLeaving)

	shuffler, err := createHashShufflerIntraShards()
	require.Nil(t, err)

	arg := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          make([]Validator, 0),
		UnStakeLeaving:    unstakeLeavingList,
		AdditionalLeaving: additionalLeavingList,
		Rand:              generateRandomByteArray(32),
		NbShards:          nbShards,
	}

	result, err := shuffler.UpdateNodeLists(arg)
	require.Nil(t, err)

	leavingPerShardMap, stillRemainingPerShardMap := createActuallyLeavingPerShards(unstakeLeaving, additionalLeaving, result.Leaving)

	for i := uint32(0); i < nbShards+1; i++ {
		shardId := i
		if i == nbShards {
			shardId = core.MetachainShardId
		}
		verifyResultsIntraShardShuffling(t,
			eligibleMap[shardId],
			waitingMap[shardId],
			additionalLeaving[shardId],
			unstakeLeaving[shardId],
			result.Eligible[shardId],
			result.Waiting[shardId],
			leavingPerShardMap[shardId],
			stillRemainingPerShardMap[shardId],
		)
	}
}

func TestRandHashShuffler_UpdateNodeLists_WithUnstakeAndAdditionalLeaving_NoDupplicates(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 100
	waitingPerShard := 30
	nbShards := uint32(2)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	unstakeLeaving := make(map[uint32][]Validator)
	unstakeLeaving[core.MetachainShardId] = []Validator{
		eligibleMap[core.MetachainShardId][0],
		waitingMap[core.MetachainShardId][0],
	}

	unstakeLeaving[0] = []Validator{
		waitingMap[0][0],
	}

	unstakeLeaving[1] = []Validator{
		eligibleMap[1][2],
		eligibleMap[1][eligiblePerShard-1]}

	additionalLeaving := make(map[uint32][]Validator)
	additionalLeaving[core.MetachainShardId] = []Validator{
		eligibleMap[core.MetachainShardId][eligiblePerShard-1],
		waitingMap[core.MetachainShardId][1],
	}

	additionalLeaving[0] = []Validator{
		waitingMap[0][waitingPerShard/2]}

	additionalLeaving[1] = []Validator{
		eligibleMap[1][0],
		eligibleMap[1][1],
	}

	unstakeLeavingList, additionalLeavingList := prepareListsFromMaps(unstakeLeaving, additionalLeaving)

	shuffler, err := createHashShufflerIntraShards()
	require.Nil(t, err)

	arg := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          make([]Validator, 0),
		UnStakeLeaving:    unstakeLeavingList,
		AdditionalLeaving: additionalLeavingList,
		Rand:              generateRandomByteArray(32),
		NbShards:          nbShards,
	}

	result, err := shuffler.UpdateNodeLists(arg)
	require.Nil(t, err)

	leavingPerShardMap, stillRemainingPerShardMap := createActuallyLeavingPerShards(unstakeLeaving, additionalLeaving, result.Leaving)

	for i := uint32(0); i < nbShards+1; i++ {
		shardId := i
		if i == nbShards {
			shardId = core.MetachainShardId
		}
		verifyResultsIntraShardShuffling(t,
			eligibleMap[shardId],
			waitingMap[shardId],
			additionalLeaving[shardId],
			unstakeLeaving[shardId],
			result.Eligible[shardId],
			result.Waiting[shardId],
			leavingPerShardMap[shardId],
			stillRemainingPerShardMap[shardId],
		)
	}
}
func TestRandHashShuffler_UpdateNodeLists_WithAdditionalLeaving_WithDupplicates(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 100
	waitingPerShard := 30
	nbShards := uint32(2)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	unstakeLeaving := make(map[uint32][]Validator)
	unstakeLeaving[core.MetachainShardId] = []Validator{
		eligibleMap[core.MetachainShardId][0],
		eligibleMap[core.MetachainShardId][eligiblePerShard-1],
		waitingMap[core.MetachainShardId][0],
		waitingMap[core.MetachainShardId][1],
	}

	unstakeLeaving[0] = []Validator{
		waitingMap[0][0],
	}

	unstakeLeaving[1] = []Validator{
		eligibleMap[1][1],
		eligibleMap[1][2],
		eligibleMap[1][eligiblePerShard-1],
	}

	additionalLeaving := make(map[uint32][]Validator)
	additionalLeaving[core.MetachainShardId] = []Validator{
		eligibleMap[core.MetachainShardId][0],
		waitingMap[core.MetachainShardId][0],
		waitingMap[core.MetachainShardId][1],
	}

	additionalLeaving[0] = []Validator{
		waitingMap[0][0],
		waitingMap[0][waitingPerShard/2]}

	additionalLeaving[1] = []Validator{
		eligibleMap[1][0],
		eligibleMap[1][1],
		eligibleMap[1][2],
	}

	unstakeLeavingList, additionalLeavingList := prepareListsFromMaps(unstakeLeaving, additionalLeaving)

	shuffler, err := createHashShufflerIntraShards()
	require.Nil(t, err)

	arg := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          make([]Validator, 0),
		UnStakeLeaving:    unstakeLeavingList,
		AdditionalLeaving: additionalLeavingList,
		Rand:              generateRandomByteArray(32),
		NbShards:          nbShards,
	}

	result, err := shuffler.UpdateNodeLists(arg)
	require.Nil(t, err)

	leavingPerShardMap, stillRemainingPerShardMap := createActuallyLeavingPerShards(unstakeLeaving, additionalLeaving, result.Leaving)

	for i := uint32(0); i < nbShards+1; i++ {
		shardId := i
		if i == nbShards {
			shardId = core.MetachainShardId
		}
		verifyResultsIntraShardShuffling(t,
			eligibleMap[shardId],
			waitingMap[shardId],
			additionalLeaving[shardId],
			unstakeLeaving[shardId],
			result.Eligible[shardId],
			result.Waiting[shardId],
			leavingPerShardMap[shardId],
			stillRemainingPerShardMap[shardId],
		)
	}
}

func TestRandHashShuffler_UpdateNodeLists_All(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 4
	waitingPerShard := 2
	nbShards := uint32(2)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	unstakeLeaving := make(map[uint32][]Validator)
	additionalLeaving := make(map[uint32][]Validator)

	// duplicates on metachain both eligible and waiting
	firstRemovedMeta := waitingMap[core.MetachainShardId][1]
	secondRemovedMeta := eligibleMap[core.MetachainShardId][0]
	notRemovedMeta := eligibleMap[core.MetachainShardId][eligiblePerShard-1]

	unstakeLeaving[core.MetachainShardId] = []Validator{
		secondRemovedMeta,
		firstRemovedMeta,
	}
	additionalLeaving[core.MetachainShardId] = []Validator{
		secondRemovedMeta,
		firstRemovedMeta,
		notRemovedMeta,
	}

	// no duplicates on shard 0
	firstRemovedShard0 := eligibleMap[0][3]
	secondRemovedShard0 := waitingMap[0][0]
	unstakeLeaving[0] = []Validator{
		firstRemovedShard0,
	}
	additionalLeaving[0] = []Validator{
		secondRemovedShard0,
	}

	// just 1 from waiting to be removed from shard 1
	firstRemovedShard1 := waitingMap[1][1]
	unstakeLeaving[1] = []Validator{
		firstRemovedShard1,
	}
	additionalLeaving[1] = nil

	unstakeLeavingList, additionalLeavingList := prepareListsFromMaps(unstakeLeaving, additionalLeaving)

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           uint32(eligiblePerShard),
		NodesMeta:            uint32(eligiblePerShard),
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	arg := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          make([]Validator, 0),
		UnStakeLeaving:    unstakeLeavingList,
		AdditionalLeaving: additionalLeavingList,
		Rand:              generateRandomByteArray(32),
		NbShards:          nbShards,
	}

	result, err := shuffler.UpdateNodeLists(arg)
	require.Nil(t, err)

	leavingPerShardMap, stillRemainingPerShardMap := createActuallyLeavingPerShards(unstakeLeaving, additionalLeaving, result.Leaving)

	for i := uint32(0); i < nbShards+1; i++ {
		shardId := i
		if i == nbShards {
			shardId = core.MetachainShardId
		}
		verifyResultsIntraShardShuffling(t,
			eligibleMap[shardId],
			waitingMap[shardId],
			additionalLeaving[shardId],
			unstakeLeaving[shardId],
			result.Eligible[shardId],
			result.Waiting[shardId],
			leavingPerShardMap[shardId],
			stillRemainingPerShardMap[shardId],
		)
	}

	removedFromMeta := []Validator{firstRemovedMeta, secondRemovedMeta}
	sort.Sort(validatorList(removedFromMeta))
	sort.Sort(validatorList(leavingPerShardMap[core.MetachainShardId]))
	assert.Equal(t, removedFromMeta, leavingPerShardMap[core.MetachainShardId])
	found, _ := searchInMap(result.Eligible, firstRemovedMeta.PubKey())
	assert.False(t, found)
	found, _ = searchInMap(result.Waiting, secondRemovedMeta.PubKey())
	assert.False(t, found)

	remainingInMeta := []Validator{notRemovedMeta}
	sort.Sort(validatorList(remainingInMeta))
	sort.Sort(validatorList(stillRemainingPerShardMap[core.MetachainShardId]))
	assert.Equal(t, remainingInMeta, stillRemainingPerShardMap[core.MetachainShardId])
	found, shardId := searchInMap(result.Eligible, notRemovedMeta.PubKey())
	assert.True(t, found)
	assert.Equal(t, core.MetachainShardId, shardId)

	removedFromShard0 := []Validator{firstRemovedShard0, secondRemovedShard0}
	sort.Sort(validatorList(removedFromShard0))
	sort.Sort(validatorList(leavingPerShardMap[0]))
	assert.Equal(t, removedFromShard0, leavingPerShardMap[0])
	found, _ = searchInMap(result.Eligible, firstRemovedShard0.PubKey())
	assert.False(t, found)
	found, _ = searchInMap(result.Waiting, secondRemovedShard0.PubKey())
	assert.False(t, found)

	removedFromShard1 := []Validator{firstRemovedShard1}
	sort.Sort(validatorList(removedFromShard1))
	sort.Sort(validatorList(leavingPerShardMap[1]))
	assert.Equal(t, removedFromShard1, leavingPerShardMap[1])
	found, _ = searchInMap(result.Waiting, secondRemovedShard0.PubKey())
	assert.False(t, found)
}

func TestRandHashShuffler_UpdateNodeLists_WithNewNodes_NoWaiting(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 10
	newNodesPerShard := 10
	numWaitingPerShard := 0
	nbShards := uint32(2)
	randomness := generateRandomByteArray(32)

	leavingNodes := make([]Validator, 0)
	additionalLeaving := make([]Validator, 0)
	newNodes := generateValidatorList(newNodesPerShard * (int(nbShards) + 1))

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(numWaitingPerShard, nbShards)

	args := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          newNodes,
		UnStakeLeaving:    leavingNodes,
		AdditionalLeaving: additionalLeaving,
		Rand:              randomness,
		NbShards:          nbShards,
	}

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           uint32(eligiblePerShard),
		NodesMeta:            uint32(eligiblePerShard),
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}

	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	resUpdateNodeList, err := shuffler.UpdateNodeLists(args)
	require.Nil(t, err)

	allNewEligible := getValidatorsInMap(resUpdateNodeList.Eligible)
	allNewWaiting := getValidatorsInMap(resUpdateNodeList.Waiting)

	oldEligible := getValidatorsInMap(args.Eligible)
	sort.Sort(validatorList(allNewEligible))
	sort.Sort(validatorList(oldEligible))
	assert.Equal(t, oldEligible, allNewEligible)

	sort.Sort(validatorList(allNewWaiting))
	sort.Sort(validatorList(args.NewNodes))
	assert.Equal(t, args.NewNodes, allNewWaiting)

	for _, waitingInShard := range resUpdateNodeList.Waiting {
		assert.Equal(t, newNodesPerShard, len(waitingInShard))
	}

	previousNumberOfNodes := (eligiblePerShard+numWaitingPerShard)*(int(nbShards)+1) + len(args.NewNodes)
	currentNumberOfNodes := len(allNewEligible) + len(allNewWaiting) + len(resUpdateNodeList.Leaving)
	assert.Equal(t, previousNumberOfNodes, currentNumberOfNodes)
}

func TestRandHashShuffler_UpdateNodeLists_WithNewNodes_NilOrEmptyWaiting(t *testing.T) {
	t.Parallel()

	eligiblePerShard := 10
	newNodesPerShard := 10
	nbShards := uint32(2)
	randomness := generateRandomByteArray(32)

	leavingNodes := make([]Validator, 0)
	additionalLeaving := make([]Validator, 0)
	newNodes := generateValidatorList(newNodesPerShard * (int(nbShards) + 1))

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)

	args := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           nil,
		NewNodes:          newNodes,
		UnStakeLeaving:    leavingNodes,
		AdditionalLeaving: additionalLeaving,
		Rand:              randomness,
		NbShards:          nbShards,
	}

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           uint32(eligiblePerShard),
		NodesMeta:            uint32(eligiblePerShard),
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	resUpdateNodeList, err := shuffler.UpdateNodeLists(args)
	require.Nil(t, err)
	require.Equal(t, int(nbShards+1), len(resUpdateNodeList.Waiting))

	args = ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           make(map[uint32][]Validator),
		NewNodes:          newNodes,
		UnStakeLeaving:    leavingNodes,
		AdditionalLeaving: additionalLeaving,
		Rand:              randomness,
		NbShards:          nbShards,
	}

	resUpdateNodeList, err = shuffler.UpdateNodeLists(args)
	require.Nil(t, err)
	require.Equal(t, int(nbShards+1), len(resUpdateNodeList.Waiting))
}

func TestRandHashShuffler_UpdateNodeLists_WithNewNodes_WithWaiting(t *testing.T) {
	t.Parallel()

	numEligiblePerShard := 100
	newNodesPerShard := 100
	numWaitingPerShard := 30
	nbShards := 2
	randomness := generateRandomByteArray(32)

	leavingNodes := make([]Validator, 0)
	additionalLeaving := make([]Validator, 0)
	newNodes := generateValidatorList(newNodesPerShard * (nbShards + 1))

	eligibleMap := generateValidatorMap(numEligiblePerShard, uint32(nbShards))
	waitingMap := generateValidatorMap(numWaitingPerShard, uint32(nbShards))

	args := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          newNodes,
		UnStakeLeaving:    leavingNodes,
		AdditionalLeaving: additionalLeaving,
		Rand:              randomness,
		NbShards:          uint32(nbShards),
	}

	shuffler, err := createHashShufflerIntraShards()
	require.Nil(t, err)

	resUpdateNodeList, err := shuffler.UpdateNodeLists(args)
	require.Nil(t, err)

	allNewEligible := getValidatorsInMap(resUpdateNodeList.Eligible)
	allNewWaiting := getValidatorsInMap(resUpdateNodeList.Waiting)

	for _, newNode := range args.NewNodes {
		found, _ := searchInMap(resUpdateNodeList.Waiting, newNode.PubKey())
		assert.True(t, found)
	}

	for shardId, waitingInShard := range resUpdateNodeList.Waiting {
		assert.Equal(t, newNodesPerShard+len(waitingMap[shardId]), len(waitingInShard))
	}

	previousNumberOfNodes := (numEligiblePerShard+numWaitingPerShard)*(nbShards+1) + len(args.NewNodes)
	currentNumberOfNodes := len(allNewEligible) + len(allNewWaiting) + len(resUpdateNodeList.Leaving)
	assert.Equal(t, previousNumberOfNodes, currentNumberOfNodes)
}

func TestRandHashShuffler_UpdateNodeLists_WithNewNodes_WithWaiting_WithLeaving(t *testing.T) {
	t.Parallel()

	numEligiblePerShard := 10
	newNodesPerShard := 10
	numWaitingPerShard := 3
	nbShards := uint32(2)
	randomness := generateRandomByteArray(32)

	leavingNodes := make([]Validator, 0)
	additionalLeaving := make([]Validator, 0)
	newNodes := generateValidatorList(newNodesPerShard * (int(nbShards) + 1))

	eligibleMap := generateValidatorMap(numEligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(numWaitingPerShard, nbShards)

	// meta leaving 4 nodes, 1 dupplicate
	leavingNodes = append(leavingNodes, []Validator{
		eligibleMap[core.MetachainShardId][0],
		eligibleMap[core.MetachainShardId][1],
		waitingMap[core.MetachainShardId][2],
	}...)

	additionalLeaving = append(additionalLeaving, []Validator{
		eligibleMap[core.MetachainShardId][0],
		waitingMap[core.MetachainShardId][1],
	}...)

	// shard 0 leaving 2 nodes
	leavingNodes = append(leavingNodes, []Validator{
		waitingMap[0][2],
	}...)
	additionalLeaving = append(additionalLeaving, []Validator{
		eligibleMap[0][5],
	}...)

	// shard 1 leaving
	additionalLeaving = append(additionalLeaving, []Validator{
		eligibleMap[1][2],
	}...)

	args := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          newNodes,
		UnStakeLeaving:    leavingNodes,
		AdditionalLeaving: additionalLeaving,
		Rand:              randomness,
		NbShards:          nbShards,
	}

	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           uint32(numEligiblePerShard),
		NodesMeta:            uint32(numEligiblePerShard),
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: nil,
	}
	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	resUpdateNodeList, err := shuffler.UpdateNodeLists(args)
	require.Nil(t, err)

	for _, newNode := range args.NewNodes {
		found, _ := searchInMap(resUpdateNodeList.Waiting, newNode.PubKey())
		assert.True(t, found)
	}

	assert.Equal(t, numEligiblePerShard, len(args.Eligible[core.MetachainShardId]))
	assert.Equal(t, newNodesPerShard, len(resUpdateNodeList.Waiting[core.MetachainShardId])) // 3 left, just new nodes in waiting
	assert.Equal(t, 1, len(resUpdateNodeList.StillRemaining))

	numWaitingShard0 := newNodesPerShard + len(waitingMap[0]) - 2 //2 left => 1 in waiting
	assert.Equal(t, numEligiblePerShard, len(resUpdateNodeList.Eligible[0]))
	assert.Equal(t, numWaitingShard0, len(resUpdateNodeList.Waiting[0]))

	numWaitingShard1 := newNodesPerShard + len(waitingMap[1]) - 1 //1 left => 2 in waiting
	assert.Equal(t, numEligiblePerShard, len(resUpdateNodeList.Eligible[1]))
	assert.Equal(t, numWaitingShard1, len(resUpdateNodeList.Waiting[1]))
}

func prepareListsFromMaps(unstakeLeaving map[uint32][]Validator, additionalLeaving map[uint32][]Validator) ([]Validator, []Validator) {
	return getValidatorsInMap(unstakeLeaving), getValidatorsInMap(additionalLeaving)
}

func verifyResultsIntraShardShuffling(
	t *testing.T,
	eligible []Validator,
	waiting []Validator,
	unstakeLeaving []Validator,
	additionalLeaving []Validator,
	newEligible []Validator,
	newWaiting []Validator,
	leaving []Validator,
	stillRemaining []Validator) {

	removedNodes := make([]Validator, 0)
	assert.Equal(t, len(eligible), len(newEligible))
	initialNumWaiting := len(waiting)
	numToRemove := initialNumWaiting

	additionalLeaving = removeDupplicates(unstakeLeaving, additionalLeaving)

	computedNewWaiting, removedFromWaiting := removeValidatorsFromList(waiting, unstakeLeaving, numToRemove)
	removedNodes = append(removedNodes, removedFromWaiting...)
	numToRemove = numToRemove - len(removedFromWaiting)

	computedNewEligible, removedFromEligible := removeValidatorsFromList(eligible, unstakeLeaving, numToRemove)
	numToRemove = numToRemove - len(removedFromEligible)
	removedNodes = append(removedNodes, removedFromEligible...)

	_, removedFromWaiting = removeValidatorsFromList(computedNewWaiting, additionalLeaving, numToRemove)
	numToRemove = numToRemove - len(removedFromWaiting)
	removedNodes = append(removedNodes, removedFromWaiting...)

	_, removedFromEligible = removeValidatorsFromList(computedNewEligible, additionalLeaving, numToRemove)
	numToRemove = numToRemove - len(removedFromEligible)
	removedNodes = append(removedNodes, removedFromEligible...)

	actuallyRemoved := initialNumWaiting - numToRemove
	previousNrValidatorsInShard := len(eligible) + len(waiting)
	currentNrValidatorsInShard := len(newEligible) + len(newWaiting)

	assert.Equal(t, previousNrValidatorsInShard, currentNrValidatorsInShard+actuallyRemoved)
	assert.True(t, initialNumWaiting >= len(removedNodes))

	allLeavingValidators := append(unstakeLeaving, additionalLeaving...)
	computedStillRemaining, _ := removeValidatorsFromList(allLeavingValidators, removedNodes, len(removedNodes))

	assert.Equal(t, len(removedNodes), len(leaving))
	assert.Equal(t, len(allLeavingValidators), len(leaving)+len(stillRemaining))

	assert.Equal(t, len(computedStillRemaining), len(stillRemaining))

	// TODO: see how to correctly compute the removed validators
	//assert.Equal(t, removedNodes, leaving)
	//if len(computedStillRemaining) > 0 {
	//	assert.Equal(t, computedStillRemaining, stillRemaining)
	//}
}

func TestRandHashShuffler_ShuffleOutShard_WithMoreThanValidatorsShuffledOut(t *testing.T) {
	t.Parallel()

	validatorsNumber := 10
	numToShuffle := 20
	randomness := generateRandomByteArray(32)

	validators := generateValidatorList(validatorsNumber)

	shuffledOut, remaining := shuffleOutShard(validators, numToShuffle, randomness)

	assert.Equal(t, validatorsNumber, len(shuffledOut))
	assert.Equal(t, 0, len(remaining))
}

func TestRandHashShuffler_ShuffleOutShard_WithLessThanValidatorsShuffledOut(t *testing.T) {
	t.Parallel()

	validatorsNumber := 10
	numToShuffle := 6
	randomness := generateRandomByteArray(32)

	validators := generateValidatorList(validatorsNumber)

	shuffledOut, remaining := shuffleOutShard(validators, numToShuffle, randomness)

	assert.Equal(t, numToShuffle, len(shuffledOut))
	assert.Equal(t, validatorsNumber-numToShuffle, len(remaining))
}

func TestRandHashShuffler_ShuffleOutShard_WithZeroValidatorsShuffledOut(t *testing.T) {
	t.Parallel()

	validatorsNumber := 10
	numToShuffle := 0
	randomness := generateRandomByteArray(32)

	validators := generateValidatorList(validatorsNumber)

	shuffledOut, remaining := shuffleOutShard(validators, numToShuffle, randomness)

	assert.Equal(t, 0, len(shuffledOut))
	assert.Equal(t, validatorsNumber, len(remaining))
}

func listToValidatorMap(validators []Validator) map[string]Validator {
	validatorMap := make(map[string]Validator)

	for _, v := range validators {
		validatorMap[string(v.PubKey())] = v
	}

	return validatorMap
}

func createShufflerArgs(eligiblePerShard int, waitingPerShard int, nbShards uint32) ArgsUpdateNodes {
	randomness := generateRandomByteArray(32)

	leavingNodes := make([]Validator, 0)
	additionalLeaving := make([]Validator, 0)
	newNodes := make([]Validator, 0)

	eligibleMap := generateValidatorMap(eligiblePerShard, nbShards)
	waitingMap := generateValidatorMap(waitingPerShard, nbShards)

	args := ArgsUpdateNodes{
		Eligible:          eligibleMap,
		Waiting:           waitingMap,
		NewNodes:          newNodes,
		UnStakeLeaving:    leavingNodes,
		AdditionalLeaving: additionalLeaving,
		Rand:              randomness,
		NbShards:          nbShards,
	}
	return args
}

func BenchmarkRandHashShuffler_RemoveWithReslice(b *testing.B) {
	nrValidators := 50000
	validators := generateValidatorList(nrValidators)
	validatorsCopy := make([]Validator, len(validators))
	_ = copy(validatorsCopy, validators)

	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(validators)-2; j++ {
			validators = removeValidatorFromListKeepOrder(validators, 2)
		}
	}

	m2 := runtime.MemStats{}
	runtime.ReadMemStats(&m2)
	fmt.Printf("Used %d MB\n", (m2.HeapAlloc-m.HeapAlloc)/1024/1024)
}

func BenchmarkRandHashShuffler_RemoveWithoutReslice(b *testing.B) {
	nrValidators := 50000
	validators := generateValidatorList(nrValidators)
	validatorsCopy := make([]Validator, len(validators))
	_ = copy(validatorsCopy, validators)

	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(validators)-2; j++ {
			validators = removeValidatorFromList(validators, 2)
		}
	}

	m2 := runtime.MemStats{}
	runtime.ReadMemStats(&m2)
	fmt.Printf("Used %d MB\n", (m2.HeapAlloc-m.HeapAlloc)/1024/1024)
}

func TestRandHashShuffler_sortConfigs(t *testing.T) {
	t.Parallel()

	orderedConfigs := getDummyShufflerConfigs()

	shuffledConfigs := make([]config.MaxNodesChangeConfig, len(orderedConfigs))
	copy(shuffledConfigs, orderedConfigs)

	n := 100

	for i := 0; i < n; i++ {
		mathRand.Shuffle(len(shuffledConfigs), func(i, j int) {
			shuffledConfigs[i], shuffledConfigs[j] = shuffledConfigs[j], shuffledConfigs[i]
		})
		require.NotEqual(t, orderedConfigs, shuffledConfigs)

		shufflerArgs := &NodesShufflerArgs{
			NodesShard:           eligiblePerShard,
			NodesMeta:            eligiblePerShard,
			Hysteresis:           hysteresis,
			Adaptivity:           adaptivity,
			ShuffleBetweenShards: shuffleBetweenShards,
			MaxNodesEnableConfig: shuffledConfigs,
		}
		shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
		require.Nil(t, err)
		require.Equal(t, orderedConfigs, shuffler.availableNodesConfigs)
	}
}

func TestRandHashShuffler_UpdateShufflerConfig(t *testing.T) {
	t.Parallel()

	orderedConfigs := getDummyShufflerConfigs()
	shufflerArgs := &NodesShufflerArgs{
		NodesShard:           eligiblePerShard,
		NodesMeta:            eligiblePerShard,
		Hysteresis:           hysteresis,
		Adaptivity:           adaptivity,
		ShuffleBetweenShards: shuffleBetweenShards,
		MaxNodesEnableConfig: orderedConfigs,
	}
	shuffler, err := NewHashValidatorsShuffler(shufflerArgs)
	require.Nil(t, err)

	i := 0
	for epoch := orderedConfigs[0].EpochEnable; epoch < orderedConfigs[len(orderedConfigs)-1].EpochEnable; epoch++ {
		if epoch == orderedConfigs[(i+1)%len(orderedConfigs)].EpochEnable {
			i++
		}
		shuffler.UpdateShufflerConfig(epoch)
		require.Equal(t, orderedConfigs[i], shuffler.activeNodesConfig)
	}
}

func getDummyShufflerConfigs() []config.MaxNodesChangeConfig {
	return []config.MaxNodesChangeConfig{
		{EpochEnable: 0, MaxNumNodes: 2500, NodesToShufflePerShard: 400},
		{EpochEnable: 3, MaxNumNodes: 3500, NodesToShufflePerShard: 143},
		{EpochEnable: 4, MaxNumNodes: 2000, NodesToShufflePerShard: 80},
		{EpochEnable: 5, MaxNumNodes: 1000, NodesToShufflePerShard: 120},
		{EpochEnable: 20, MaxNumNodes: 10000, NodesToShufflePerShard: 400},
		{EpochEnable: 101, MaxNumNodes: 10000, NodesToShufflePerShard: 80},
		{EpochEnable: 102, MaxNumNodes: 10000, NodesToShufflePerShard: 400},
		{EpochEnable: 221, MaxNumNodes: 3200, NodesToShufflePerShard: 400},
		{EpochEnable: 1000, MaxNumNodes: 2900, NodesToShufflePerShard: 400},
		{EpochEnable: 2300, MaxNumNodes: 5400, NodesToShufflePerShard: 400},
	}
}
