package factory

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSovereignValidatorAccountsSyncerFactory_CreateValidatorAccountsSyncer(t *testing.T) {
	t.Parallel()

	args := getArgs()
	factory := NewSovereignValidatorAccountsSyncerFactory()
	require.False(t, factory.IsInterfaceNil())

	valSyncer, err := factory.CreateValidatorAccountsSyncer(args)
	require.Nil(t, err)
	require.Equal(t, "*syncer.validatorAccountsSyncer", fmt.Sprintf("%T", valSyncer))
}
