package consensus_test

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/stretchr/testify/assert"
)

func TestValidator_NewValidatorShouldFailOnNilStake(t *testing.T) {
	t.Parallel()

	validator, err := consensus.NewValidator(nil, 0, []byte("pk1"))

	assert.Nil(t, validator)
	assert.Equal(t, consensus.ErrNilStake, err)
}

func TestValidator_NewValidatorShouldFailOnNegativeStake(t *testing.T) {
	t.Parallel()

	validator, err := consensus.NewValidator(big.NewInt(-1), 0, []byte("pk1"))

	assert.Nil(t, validator)
	assert.Equal(t, consensus.ErrNegativeStake, err)
}

func TestValidator_NewValidatorShouldFailOnNilPublickKey(t *testing.T) {
	t.Parallel()

	validator, err := consensus.NewValidator(big.NewInt(0), 0, nil)

	assert.Nil(t, validator)
	assert.Equal(t, consensus.ErrNilPubKey, err)
}

func TestValidator_NewValidatorShouldWork(t *testing.T) {
	t.Parallel()

	validator, err := consensus.NewValidator(big.NewInt(0), 0, []byte("pk1"))

	assert.NotNil(t, validator)
	assert.Nil(t, err)
}

func TestValidator_StakeShouldWork(t *testing.T) {
	t.Parallel()

	validator, _ := consensus.NewValidator(big.NewInt(1), 0, []byte("pk1"))

	assert.Equal(t, big.NewInt(1), validator.Stake())
}

func TestValidator_PubKeyShouldWork(t *testing.T) {
	t.Parallel()

	validator, _ := consensus.NewValidator(big.NewInt(0), 0, []byte("pk1"))

	assert.Equal(t, []byte("pk1"), validator.PubKey())
}
