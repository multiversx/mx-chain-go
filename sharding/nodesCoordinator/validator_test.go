package nodesCoordinator

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestValidator_NewValidatorShouldFailOnNilPublickKey(t *testing.T) {
	t.Parallel()

	v, err := NewValidator(nil, defaultSelectionChances, 1)

	assert.Nil(t, v)
	assert.Equal(t, ErrNilPubKey, err)
}

func TestValidator_NewValidatorShouldWork(t *testing.T) {
	t.Parallel()

	v, err := NewValidator([]byte("pk1"), defaultSelectionChances, 1)

	assert.NotNil(t, v)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(v))
}

func TestValidator_PubKeyShouldWork(t *testing.T) {
	t.Parallel()

	v, _ := NewValidator([]byte("pk1"), defaultSelectionChances, 1)

	assert.Equal(t, []byte("pk1"), v.PubKey())
}
