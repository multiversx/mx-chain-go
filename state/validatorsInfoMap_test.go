package state

import (
	"encoding/hex"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/require"
)

func TestShardValidatorsInfoMap_OperationsWithNilValidators(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap()

	t.Run("add nil validator", func(t *testing.T) {
		t.Parallel()

		err := vi.Add(nil)
		require.Equal(t, ErrNilValidatorInfo, err)
	})

	t.Run("delete nil validator", func(t *testing.T) {
		t.Parallel()

		err := vi.Delete(nil)
		require.Equal(t, ErrNilValidatorInfo, err)
	})

	t.Run("replace nil validator", func(t *testing.T) {
		t.Parallel()

		err := vi.Replace(nil, &ValidatorInfo{})
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), ErrNilValidatorInfo.Error()))
		require.True(t, strings.Contains(err.Error(), "old"))

		err = vi.Replace(&ValidatorInfo{}, nil)
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), ErrNilValidatorInfo.Error()))
		require.True(t, strings.Contains(err.Error(), "new"))
	})

	t.Run("set nil validators in shard", func(t *testing.T) {
		t.Parallel()

		v := &ValidatorInfo{ShardId: 3, PublicKey: []byte("pk")}
		err := vi.SetValidatorsInShard(3, []ValidatorInfoHandler{v, nil})
		require.Error(t, err)
		require.True(t, strings.Contains(err.Error(), ErrNilValidatorInfo.Error()))
		require.True(t, strings.Contains(err.Error(), "index 1"))
	})
}

func TestShardValidatorsInfoMap_Add_GetShardValidatorsInfoMap_GetAllValidatorsInfo_GetValInfoPointerMap(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap()

	v0 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk0")}
	v1 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk1")}
	v2 := &ValidatorInfo{ShardId: 1, PublicKey: []byte("pk2")}
	v3 := &ValidatorInfo{ShardId: core.MetachainShardId, PublicKey: []byte("pk3")}

	_ = vi.Add(v0)
	_ = vi.Add(v1)
	_ = vi.Add(v2)
	_ = vi.Add(v3)

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
}

func TestShardValidatorsInfoMap_GetValidator(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap()

	pubKey0 := []byte("pk0")
	pubKey1 := []byte("pk1")
	v0 := &ValidatorInfo{ShardId: 0, PublicKey: pubKey0}
	v1 := &ValidatorInfo{ShardId: 1, PublicKey: pubKey1}

	_ = vi.Add(v0)
	_ = vi.Add(v1)

	require.Equal(t, v0, vi.GetValidator(pubKey0))
	require.Equal(t, v1, vi.GetValidator(pubKey1))
	require.Nil(t, vi.GetValidator([]byte("pk2")))
}

func TestShardValidatorsInfoMap_Delete(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap()

	v0 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk0")}
	v1 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk1")}
	v2 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk2")}
	v3 := &ValidatorInfo{ShardId: 1, PublicKey: []byte("pk3")}

	_ = vi.Add(v0)
	_ = vi.Add(v1)
	_ = vi.Add(v2)
	_ = vi.Add(v3)

	_ = vi.Delete(&ValidatorInfo{ShardId: 0, PublicKey: []byte("pk3")})
	_ = vi.Delete(&ValidatorInfo{ShardId: 1, PublicKey: []byte("pk0")})
	require.Len(t, vi.GetAllValidatorsInfo(), 4)

	_ = vi.Delete(v1)
	require.Len(t, vi.GetAllValidatorsInfo(), 3)
	require.Equal(t, []ValidatorInfoHandler{v0, v2}, vi.GetShardValidatorsInfoMap()[0])
	require.Equal(t, []ValidatorInfoHandler{v3}, vi.GetShardValidatorsInfoMap()[1])

	_ = vi.Delete(v3)
	require.Len(t, vi.GetAllValidatorsInfo(), 2)
	require.Equal(t, []ValidatorInfoHandler{v0, v2}, vi.GetShardValidatorsInfoMap()[0])
}

func TestShardValidatorsInfoMap_Replace(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap()

	v0 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk0")}
	v1 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk1")}

	_ = vi.Add(v0)
	_ = vi.Add(v1)

	err := vi.Replace(v0, &ValidatorInfo{ShardId: 1, PublicKey: []byte("pk2")})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), ErrValidatorsDifferentShards.Error()))
	require.Equal(t, []ValidatorInfoHandler{v0, v1}, vi.GetShardValidatorsInfoMap()[0])

	v2 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk2")}
	err = vi.Replace(v0, v2)
	require.Nil(t, err)
	require.Equal(t, []ValidatorInfoHandler{v2, v1}, vi.GetShardValidatorsInfoMap()[0])

	v3 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk3")}
	v4 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk4")}
	err = vi.Replace(v3, v4)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), ErrValidatorNotFound.Error()))
	require.True(t, strings.Contains(err.Error(), hex.EncodeToString(v3.PublicKey)))
	require.Equal(t, []ValidatorInfoHandler{v2, v1}, vi.GetShardValidatorsInfoMap()[0])
}

func TestShardValidatorsInfoMap_SetValidatorsInShard(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap()

	v0 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk0")}
	_ = vi.Add(v0)

	v1 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk1")}
	v2 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk2")}
	v3 := &ValidatorInfo{ShardId: 1, PublicKey: []byte("pk3")}
	shard0Validators := []ValidatorInfoHandler{v1, v2}
	shard1Validators := []ValidatorInfoHandler{v3}

	err := vi.SetValidatorsInShard(1, shard0Validators)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), ErrValidatorsDifferentShards.Error()))
	require.True(t, strings.Contains(err.Error(), hex.EncodeToString(v1.PublicKey)))
	require.Equal(t, []ValidatorInfoHandler{v0}, vi.GetShardValidatorsInfoMap()[0])
	require.Empty(t, vi.GetShardValidatorsInfoMap()[1])

	err = vi.SetValidatorsInShard(0, []ValidatorInfoHandler{v1, v2, v3})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), ErrValidatorsDifferentShards.Error()))
	require.True(t, strings.Contains(err.Error(), hex.EncodeToString(v3.PublicKey)))
	require.Equal(t, []ValidatorInfoHandler{v0}, vi.GetShardValidatorsInfoMap()[0])
	require.Empty(t, vi.GetShardValidatorsInfoMap()[1])

	err = vi.SetValidatorsInShard(0, shard0Validators)
	require.Nil(t, err)
	require.Equal(t, shard0Validators, vi.GetShardValidatorsInfoMap()[0])

	err = vi.SetValidatorsInShard(1, shard1Validators)
	require.Nil(t, err)
	require.Equal(t, shard1Validators, vi.GetShardValidatorsInfoMap()[1])
}

func TestShardValidatorsInfoMap_GettersShouldReturnCopiesOfInternalData(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap()

	v0 := &ValidatorInfo{ShardId: 0, PublicKey: []byte("pk0")}
	v1 := &ValidatorInfo{ShardId: 1, PublicKey: []byte("pk1")}

	_ = vi.Add(v0)
	_ = vi.Add(v1)

	validatorsMap := vi.GetShardValidatorsInfoMap()
	delete(validatorsMap, 0)
	validatorsMap[1][0].SetPublicKey([]byte("rnd"))

	validators := vi.GetAllValidatorsInfo()
	validators = append(validators, &ValidatorInfo{ShardId: 1, PublicKey: []byte("pk3")})

	validator := vi.GetValidator([]byte("pk0"))
	require.False(t, validator == v0) // require not same pointer
	validator.SetShardId(2)

	require.Len(t, vi.GetAllValidatorsInfo(), 2)
	require.True(t, vi.GetShardValidatorsInfoMap()[0][0] == v0) // check by pointer
	require.True(t, vi.GetShardValidatorsInfoMap()[1][0] == v1) // check by pointer
	require.NotEqual(t, vi.GetAllValidatorsInfo(), validators)
}

func TestShardValidatorsInfoMap_Concurrency(t *testing.T) {
	t.Parallel()

	vi := NewShardValidatorsInfoMap()

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
		_ = vi.SetValidatorsInShard(0, shard0Validators)
		wg.Done()
	}()
	go func() {
		_ = vi.SetValidatorsInShard(1, shard1Validators)
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
			_ = vi.Add(val)
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
			_ = vi.Delete(val)
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
			_ = vi.Replace(old, new)
			wg.Done()
		}(oldValidators[idx], newValidators[idx])
	}
}
