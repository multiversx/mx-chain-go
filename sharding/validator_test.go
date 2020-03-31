package sharding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const defaultSelectionChances = 1

func TestValidator_NewValidatorShouldFailOnNilPublickKey(t *testing.T) {
	t.Parallel()

	v, err := NewValidator(nil, []byte("addr1"), defaultSelectionChances)

	assert.Nil(t, v)
	assert.Equal(t, ErrNilPubKey, err)
}

func TestValidator_NewValidatorShouldFailOnNilAddress(t *testing.T) {
	t.Parallel()

	v, err := NewValidator([]byte("pk1"), nil, defaultSelectionChances)

	assert.Nil(t, v)
	assert.Equal(t, ErrNilAddress, err)
}

func TestValidator_NewValidatorShouldWork(t *testing.T) {
	t.Parallel()

	v, err := NewValidator([]byte("pk1"), []byte("addr1"), defaultSelectionChances)

	assert.NotNil(t, v)
	assert.Nil(t, err)
}

func TestValidator_PubKeyShouldWork(t *testing.T) {
	t.Parallel()

	v, _ := NewValidator([]byte("pk1"), []byte("addr1"), defaultSelectionChances)

	assert.Equal(t, []byte("pk1"), v.PubKey())
}

func TestValidator_AddressShouldWork(t *testing.T) {
	t.Parallel()

	v, _ := NewValidator([]byte("pk1"), []byte("addr1"), defaultSelectionChances)

	assert.Equal(t, []byte("addr1"), v.Address())
}
