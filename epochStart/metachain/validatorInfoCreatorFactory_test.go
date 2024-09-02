package metachain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatorInfoCreatorFactory_CreateValidatorInfoCreator(t *testing.T) {
	t.Parallel()

	factory := NewValidatorInfoCreatorFactory()
	require.False(t, factory.IsInterfaceNil())

	args := createMockEpochValidatorInfoCreatorsArguments()
	vic, err := factory.CreateValidatorInfoCreator(args)
	require.Nil(t, err)
	require.Equal(t, "*metachain.validatorInfoCreator", fmt.Sprintf("%T", vic))
}
