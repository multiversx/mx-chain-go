package metachain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSovereignValidatorInfoCreatorFactory_CreateValidatorInfoCreator(t *testing.T) {
	t.Parallel()

	factory := NewSovereignValidatorInfoCreatorFactory()
	require.False(t, factory.IsInterfaceNil())

	args := createMockEpochValidatorInfoCreatorsArguments()
	vic, err := factory.CreateValidatorInfoCreator(args)
	require.Nil(t, err)
	require.Equal(t, "*metachain.sovereignValidatorInfoCreator", fmt.Sprintf("%T", vic))
}
