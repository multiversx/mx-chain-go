package state

import (
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

	allValidators := vi.GetAllValidatorsInfo()
	require.Len(t, allValidators, 4)
	require.Contains(t, allValidators, v0)
	require.Contains(t, allValidators, v1)
	require.Contains(t, allValidators, v2)
	require.Contains(t, allValidators, v3)

	validatorsMap := vi.GetShardValidatorsInfoMap()
	require.Len(t, validatorsMap, 3)
	require.Equal(t, []ValidatorInfoHandler{v0, v1}, validatorsMap[0])
	require.Equal(t, []ValidatorInfoHandler{v2}, validatorsMap[1])
	require.Equal(t, []ValidatorInfoHandler{v3}, validatorsMap[core.MetachainShardId])
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
