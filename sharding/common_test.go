package sharding

import (
	"errors"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/stretchr/testify/assert"
)

func TestSerializableValidatorsToValidators_NilPubKey(t *testing.T) {
	serializableValidators := make(map[string][]*SerializableValidator)

	serializableValidator := &SerializableValidator{
		PubKey:  nil,
		Chances: 0,
		Index:   0,
	}

	serializableValidators[strconv.Itoa(0)] = []*SerializableValidator{serializableValidator}

	_, err := SerializableValidatorsToValidators(serializableValidators)
	assert.True(t, errors.Is(err, nodesCoordinator.ErrNilPubKey))
}

func TestSerializableValidatorsToValidators_ShouldFail(t *testing.T) {
	serializableValidators := make(map[string][]*SerializableValidator)

	serializableValidator := &SerializableValidator{
		PubKey:  []byte{},
		Chances: 0,
		Index:   0,
	}

	serializableValidators["a"] = []*SerializableValidator{serializableValidator}

	_, err := SerializableValidatorsToValidators(serializableValidators)
	assert.Error(t, err)
}

func TestSerializableValidatorsToValidators_ShouldWork(t *testing.T) {
	serializableValidators := make(map[string][]*SerializableValidator)

	serializableValidator := &SerializableValidator{
		PubKey:  []byte("dummy pk"),
		Chances: 0,
		Index:   0,
	}

	shardId := uint32(0)

	serializableValidators[strconv.Itoa(0)] = []*SerializableValidator{serializableValidator}

	validatorsPerShard, err := SerializableValidatorsToValidators(serializableValidators)
	assert.Nil(t, err)

	assert.Equal(t, 1, len(validatorsPerShard))
	assert.Equal(t, []byte("dummy pk"), validatorsPerShard[shardId][0].PubKey())
}
