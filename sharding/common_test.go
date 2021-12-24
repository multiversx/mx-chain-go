package sharding

import (
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerializableValidatorsToValidators_NilPubKey(t *testing.T) {
	t.Parallel()

	serializableValidators := make(map[string][]*SerializableValidator)

	serializableValidator := &SerializableValidator{
		PubKey:  nil,
		Chances: 0,
		Index:   0,
	}

	serializableValidators["0"] = []*SerializableValidator{serializableValidator}

	validators, err := SerializableValidatorsToValidators(serializableValidators)
	require.Nil(t, validators)
	assert.Equal(t, nodesCoordinator.ErrNilPubKey, err)
}

func TestSerializableValidatorsToValidators_InvalidShardIDShouldFail(t *testing.T) {
	t.Parallel()

	serializableValidators := make(map[string][]*SerializableValidator)

	serializableValidator := &SerializableValidator{
		PubKey:  []byte{},
		Chances: 0,
		Index:   0,
	}

	serializableValidators["a"] = []*SerializableValidator{serializableValidator}

	validators, err := SerializableValidatorsToValidators(serializableValidators)
	require.Nil(t, validators)
	assert.Error(t, err)
}

func TestSerializableValidatorsToValidators_ShouldWork(t *testing.T) {
	t.Parallel()

	serializableValidators := make(map[string][]*SerializableValidator)

	pubKey := []byte("dummy pk")
	chances := uint32(1234)
	index := uint32(4321)
	serializableValidator := &SerializableValidator{
		PubKey:  pubKey,
		Chances: chances,
		Index:   index,
	}

	expectedValidator, err := nodesCoordinator.NewValidator(pubKey, chances, index)
	require.Nil(t, err)

	shardId := uint32(0)

	serializableValidators[strconv.Itoa(int(shardId))] = []*SerializableValidator{serializableValidator}

	validatorsPerShard, err := SerializableValidatorsToValidators(serializableValidators)
	require.Nil(t, err)

	assert.Equal(t, 1, len(validatorsPerShard))
	assert.Equal(t, expectedValidator, validatorsPerShard[shardId][0])
}
