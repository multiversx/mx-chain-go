package state

import (
	"strconv"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/stretchr/testify/require"
)

func TestShardValidatorsInfoMap_Add_GetShardValidatorsInfoMap_GetAllValidatorsInfo(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap(3)

	v0 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk0")}
	v1 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk1")}
	v2 := &ValidatorInfo{ShardId: 1, PublicKey: []byte("pk2")}
	v3 := &ValidatorInfo{ShardId: core.MetachainShardId, PublicKey: []byte("pk3")}

	vi.Add(v0)
	vi.Add(v1)
	vi.Add(v2)
	vi.Add(v3)
	vi.Add(v3)

	allValidators := vi.GetAllValidatorsInfo()
	require.Len(t, allValidators, 4)
	require.Contains(t, allValidators, v0)
	require.Contains(t, allValidators, v1)
	require.Contains(t, allValidators, v2)
	require.Contains(t, allValidators, v3)

	validatorsMap := vi.GetShardValidatorsInfoMap()
	expectedValidatorsMap := map[uint32][]ValidatorInfoHandler{
		0:                     {v0, v1},
		1:                     {v2},
		core.MetachainShardId: {v3},
	}
	require.Equal(t, validatorsMap, expectedValidatorsMap)

	validatorPointersMap := vi.GetValInfoPointerMap()
	expectedValidatorPointersMap := map[uint32][]*ValidatorInfo{
		0:                     {v0, v1},
		1:                     {v2},
		core.MetachainShardId: {v3},
	}
	require.Equal(t, expectedValidatorPointersMap, validatorPointersMap)
}

func TestShardValidatorsInfoMap_GetValidatorWithBLSKey(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap(1)

	pubKey0 := []byte("pk0")
	pubKey1 := []byte("pk1")
	v0 := &ValidatorInfo{ShardId: 0, PublicKey: pubKey0}
	v1 := &ValidatorInfo{ShardId: 1, PublicKey: pubKey1}

	vi.Add(v0)
	vi.Add(v1)

	require.Equal(t, v0, vi.GetValidator(pubKey0))
	require.Equal(t, v1, vi.GetValidator(pubKey1))
	require.Nil(t, vi.GetValidator([]byte("pk2")))
}

func TestShardValidatorsInfoMap_Delete(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap(2)

	v0 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk0")}
	v1 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk1")}
	v2 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk2")}
	v3 := &ValidatorInfo{ShardId: 1, PublicKey: []byte("pk3")}

	vi.Add(v0)
	vi.Add(v1)
	vi.Add(v2)
	vi.Add(v3)

	vi.Delete(&ValidatorInfo{ShardId: 0, PublicKey: []byte("pk3")})
	vi.Delete(&ValidatorInfo{ShardId: 1, PublicKey: []byte("pk0")})
	require.Len(t, vi.GetAllValidatorsInfo(), 4)

	vi.Delete(&ValidatorInfo{ShardId: 0, PublicKey: []byte("pk1")})
	require.Len(t, vi.GetAllValidatorsInfo(), 3)
	require.Equal(t, []ValidatorInfoHandler{v0, v2}, vi.GetShardValidatorsInfoMap()[0])
}

func TestShardValidatorsInfoMap_Replace(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap(2)

	v0 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk0")}
	v1 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk1")}

	vi.Add(v0)
	vi.Add(v1)

	vi.Replace(v0, &ValidatorInfo{ShardId: 1, PublicKey: []byte("pk2")})
	require.Equal(t, []ValidatorInfoHandler{v0, v1}, vi.GetShardValidatorsInfoMap()[0])

	v2 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk2")}
	vi.Replace(v0, v2)
	require.Equal(t, []ValidatorInfoHandler{v2, v1}, vi.GetShardValidatorsInfoMap()[0])
}

func TestShardValidatorsInfoMap_SetValidatorsInShard(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap(2)

	v0 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk0")}
	vi.Add(v0)

	v1 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk1")}
	v2 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk2")}
	v3 := &ValidatorInfo{ShardId: 1, PublicKey: []byte("pk3")}
	shard0Validators := []ValidatorInfoHandler{v1, v2}
	shard1Validators := []ValidatorInfoHandler{v3}

	vi.SetValidatorsInShard(1, shard0Validators)
	require.Equal(t, []ValidatorInfoHandler{v0}, vi.GetShardValidatorsInfoMap()[0])

	vi.SetValidatorsInShard(0, []ValidatorInfoHandler{v1, v2, v3})
	require.Equal(t, shard0Validators, vi.GetShardValidatorsInfoMap()[0])

	vi.SetValidatorsInShard(1, shard1Validators)
	require.Equal(t, shard0Validators, vi.GetShardValidatorsInfoMap()[0])
	require.Equal(t, shard1Validators, vi.GetShardValidatorsInfoMap()[1])
}

func TestShardValidatorsInfoMap_GettersShouldReturnCopiesOfInternalData(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap(2)

	v0 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk0")}
	v1 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk1")}
	v2 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk2")}

	vi.Add(v0)
	vi.Add(v1)
	vi.Add(v2)

	validatorsMap := vi.GetShardValidatorsInfoMap()
	delete(validatorsMap, 0)

	validatorPointersMap := vi.GetValInfoPointerMap()
	delete(validatorPointersMap, 0)

	validators := vi.GetAllValidatorsInfo()
	validators = append(validators, &ValidatorInfo{ShardId: 1, PublicKey: []byte("pk3")})

	validator := vi.GetValidator([]byte("pk0"))
	validator.SetShardId(1)

	require.Equal(t, []ValidatorInfoHandler{v0, v1, v2}, vi.GetAllValidatorsInfo())
}

func TestShardValidatorsInfoMap_Concurrency(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap(2)

	numValidatorsShard0 := 100
	numValidatorsShard1 := 50
	numValidators := numValidatorsShard0 + numValidatorsShard1

	shard0Validators := createValidatorsInfo(0, numValidatorsShard0)
	shard1Validators := createValidatorsInfo(1, numValidatorsShard1)

	firstHalfShard0 := shard0Validators[:numValidatorsShard0/2]
	secondHalfShard0 := shard0Validators[numValidatorsShard0/2:]

	firstHalfShard1 := shard1Validators[:numValidatorsShard1/2]
	secondHalfShard1 := shard1Validators[numValidatorsShard1/2:]

	wg := &sync.WaitGroup{}

	wg.Add(numValidators)
	go addValidatorsInShardConcurrently(vi, shard0Validators, wg)
	go addValidatorsInShardConcurrently(vi, shard1Validators, wg)
	wg.Wait()
	requireSameValidatorsDifferentOrder(t, vi.GetShardValidatorsInfoMap()[0], shard0Validators)
	requireSameValidatorsDifferentOrder(t, vi.GetShardValidatorsInfoMap()[1], shard1Validators)

	wg.Add(numValidators / 2)
	go deleteValidatorsConcurrently(vi, firstHalfShard0, wg)
	go deleteValidatorsConcurrently(vi, firstHalfShard1, wg)
	wg.Wait()
	requireSameValidatorsDifferentOrder(t, vi.GetShardValidatorsInfoMap()[0], secondHalfShard0)
	requireSameValidatorsDifferentOrder(t, vi.GetShardValidatorsInfoMap()[1], secondHalfShard1)

	wg.Add(numValidators / 2)
	go replaceValidatorsConcurrently(vi, vi.GetShardValidatorsInfoMap()[0], firstHalfShard0, wg)
	go replaceValidatorsConcurrently(vi, vi.GetShardValidatorsInfoMap()[1], firstHalfShard1, wg)
	wg.Wait()
	requireSameValidatorsDifferentOrder(t, vi.GetShardValidatorsInfoMap()[0], firstHalfShard0)
	requireSameValidatorsDifferentOrder(t, vi.GetShardValidatorsInfoMap()[1], firstHalfShard1)

	wg.Add(2)
	go func() {
		vi.SetValidatorsInShard(0, shard0Validators)
		wg.Done()
	}()
	go func() {
		vi.SetValidatorsInShard(1, shard1Validators)
		wg.Done()
	}()
	wg.Wait()
	requireSameValidatorsDifferentOrder(t, vi.GetShardValidatorsInfoMap()[0], shard0Validators)
	requireSameValidatorsDifferentOrder(t, vi.GetShardValidatorsInfoMap()[1], shard1Validators)
}

func requireSameValidatorsDifferentOrder(t *testing.T, dest []ValidatorInfoHandler, src []ValidatorInfoHandler) {
	require.Equal(t, len(dest), len(src))

	for _, v := range src {
		require.Contains(t, dest, v)
	}
}

func createValidatorsInfo(shardID uint32, numOfValidators int) []ValidatorInfoHandler {
	ret := make([]ValidatorInfoHandler, 0, numOfValidators)

	for i := 0; i < numOfValidators; i++ {
		ret = append(ret, &ValidatorInfo{
			ShardId:   shardID,
			PublicKey: []byte(strconv.Itoa(int(shardID)) + "pubKey" + strconv.Itoa(i)),
		})
	}

	return ret
}

func addValidatorsInShardConcurrently(
	vi ShardValidatorsInfoMapHandler,
	validators []ValidatorInfoHandler,
	wg *sync.WaitGroup,
) {
	for _, validator := range validators {
		go func(val ValidatorInfoHandler) {
			vi.Add(val)
			wg.Done()
		}(validator)
	}
}

func deleteValidatorsConcurrently(
	vi ShardValidatorsInfoMapHandler,
	validators []ValidatorInfoHandler,
	wg *sync.WaitGroup,
) {
	for _, validator := range validators {
		go func(val ValidatorInfoHandler) {
			vi.Delete(val)
			wg.Done()
		}(validator)
	}
}

func replaceValidatorsConcurrently(
	vi ShardValidatorsInfoMapHandler,
	oldValidators []ValidatorInfoHandler,
	newValidators []ValidatorInfoHandler,
	wg *sync.WaitGroup,
) {
	for idx := range oldValidators {
		go func(old ValidatorInfoHandler, new ValidatorInfoHandler) {
			vi.Replace(old, new)
			wg.Done()
		}(oldValidators[idx], newValidators[idx])
	}
}
