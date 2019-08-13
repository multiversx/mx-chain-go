package sharding_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/assert"
)

func TestValidator_NewValidatorShouldFailOnNilStake(t *testing.T) {
	t.Parallel()

	validator, err := sharding.NewValidator(nil, 0, []byte("pk1"), []byte("addr1"))

	assert.Nil(t, validator)
	assert.Equal(t, sharding.ErrNilStake, err)
}

func TestValidator_NewValidatorShouldFailOnNegativeStake(t *testing.T) {
	t.Parallel()

	validator, err := sharding.NewValidator(big.NewInt(-1), 0, []byte("pk1"), []byte("addr1"))

	assert.Nil(t, validator)
	assert.Equal(t, sharding.ErrNegativeStake, err)
}

func TestValidator_NewValidatorShouldFailOnNilPublickKey(t *testing.T) {
	t.Parallel()

	validator, err := sharding.NewValidator(big.NewInt(0), 0, nil, []byte("addr1"))

	assert.Nil(t, validator)
	assert.Equal(t, sharding.ErrNilPubKey, err)
}

func TestValidator_NewValidatorShouldFailOnNilAddress(t *testing.T) {
	t.Parallel()

	validator, err := sharding.NewValidator(big.NewInt(0), 0, []byte("pk1"), nil)

	assert.Nil(t, validator)
	assert.Equal(t, sharding.ErrNilAddress, err)
}

func TestValidator_NewValidatorShouldWork(t *testing.T) {
	t.Parallel()

	validator, err := sharding.NewValidator(big.NewInt(0), 0, []byte("pk1"), []byte("addr1"))

	assert.NotNil(t, validator)
	assert.Nil(t, err)
}

func TestValidator_StakeShouldWork(t *testing.T) {
	t.Parallel()

	validator, _ := sharding.NewValidator(big.NewInt(1), 0, []byte("pk1"), []byte("addr1"))

	assert.Equal(t, big.NewInt(1), validator.Stake())
}

func TestValidator_PubKeyShouldWork(t *testing.T) {
	t.Parallel()

	validator, _ := sharding.NewValidator(big.NewInt(0), 0, []byte("pk1"), []byte("addr1"))

	assert.Equal(t, []byte("pk1"), validator.PubKey())
}

func TestValidator_AddressShouldWork(t *testing.T) {
	t.Parallel()

	validator, _ := sharding.NewValidator(big.NewInt(0), 0, []byte("pk1"), []byte("addr1"))

	assert.Equal(t, []byte("addr1"), validator.Address())
}
