package sharding

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidator_NewValidatorShouldFailOnNilPublickKey(t *testing.T) {
	t.Parallel()

	validator, err := NewValidator(nil, []byte("addr1"))

	assert.Nil(t, validator)
	assert.Equal(t, ErrNilPubKey, err)
}

func TestValidator_NewValidatorShouldFailOnNilAddress(t *testing.T) {
	t.Parallel()

	validator, err := NewValidator([]byte("pk1"), nil)

	assert.Nil(t, validator)
	assert.Equal(t, ErrNilAddress, err)
}

func TestValidator_NewValidatorShouldWork(t *testing.T) {
	t.Parallel()

	validator, err := NewValidator([]byte("pk1"), []byte("addr1"))

	assert.NotNil(t, validator)
	assert.Nil(t, err)
}

func TestValidator_PubKeyShouldWork(t *testing.T) {
	t.Parallel()

	validator, _ := NewValidator([]byte("pk1"), []byte("addr1"))

	assert.Equal(t, []byte("pk1"), validator.PubKey())
}

func TestValidator_AddressShouldWork(t *testing.T) {
	t.Parallel()

	validator, _ := NewValidator([]byte("pk1"), []byte("addr1"))

	assert.Equal(t, []byte("addr1"), validator.Address())
}
