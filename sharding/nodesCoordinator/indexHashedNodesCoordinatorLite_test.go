package nodesCoordinator

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexHashedNodesCoordinator_SetNodesConfigFromValidatorsInfo(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)
	epoch := uint32(1)

	randomness := []byte("rand seed")

	ihgs.nodesConfig[epoch] = ihgs.nodesConfig[0]

	body := createBlockBodyFromNodesCoordinator(ihgs, epoch)
	validatorsInfo, err := createValidatorInfoFromBody(body, arguments.Marshalizer, 10)
	require.Nil(t, err)

	err = ihgs.SetNodesConfigFromValidatorsInfo(epoch, randomness, validatorsInfo)

	validators, err := ihgs.GetAllEligibleValidatorsPublicKeys(epoch)
	require.Nil(t, err)
	require.NotNil(t, validators)
}

func TestIndexHashedNodesCoordinator_SetNodesConfigFromValidatorsInfoMultipleEpochs(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)
	epoch := uint32(1)

	randomness := []byte("rand seed")

	ihgs.nodesConfig[epoch] = ihgs.nodesConfig[0]

	body := createBlockBodyFromNodesCoordinator(ihgs, epoch)
	validatorsInfo, err := createValidatorInfoFromBody(body, arguments.Marshalizer, 10)
	require.Nil(t, err)

	ihgs.SetNodesConfigFromValidatorsInfo(epoch, randomness, validatorsInfo)
	ihgs.SetNodesConfigFromValidatorsInfo(epoch+1, randomness, validatorsInfo)

	_, ok := ihgs.nodesConfig[epoch+1]
	assert.True(t, ok)

	ihgs.SetNodesConfigFromValidatorsInfo(epoch+4, randomness, validatorsInfo)

	_, ok = ihgs.nodesConfig[epoch]
	assert.False(t, ok)

	validators, err := ihgs.GetAllEligibleValidatorsPublicKeys(epoch)
	require.Nil(t, validators)
	require.True(t, errors.Is(err, ErrEpochNodesConfigDoesNotExist))
}

func TestIndexHashedNodesCoordinator_IsEpochInConfig(t *testing.T) {
	t.Parallel()

	arguments := createArguments()

	ihgs, err := NewIndexHashedNodesCoordinator(arguments)
	require.Nil(t, err)
	epoch := uint32(1)

	ihgs.nodesConfig[epoch] = ihgs.nodesConfig[0]

	body := createBlockBodyFromNodesCoordinator(ihgs, epoch)
	validatorsInfo, _ := createValidatorInfoFromBody(body, arguments.Marshalizer, 10)

	err = ihgs.SetNodesConfigFromValidatorsInfo(epoch, []byte{}, validatorsInfo)

	status := ihgs.IsEpochInConfig(epoch)
	assert.True(t, status)

	status = ihgs.IsEpochInConfig(epoch + 1)
	assert.False(t, status)
}
