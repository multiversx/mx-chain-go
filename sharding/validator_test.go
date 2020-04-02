package sharding

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestValidator_NewValidatorShouldFailOnNilPublickKey(t *testing.T) {
	t.Parallel()

	v, err := NewValidator(nil, 1, defaultSelectionChances)

	assert.Nil(t, v)
	assert.Equal(t, ErrNilPubKey, err)
}

func TestValidator_NewValidatorShouldWork(t *testing.T) {
	t.Parallel()

	v, err := NewValidator([]byte("pk1"), 1, defaultSelectionChances)

	assert.NotNil(t, v)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(v))
}

func TestValidator_PubKeyShouldWork(t *testing.T) {
	t.Parallel()

	v, _ := NewValidator([]byte("pk1"), 1, defaultSelectionChances)

	assert.Equal(t, []byte("pk1"), v.PubKey())
}
