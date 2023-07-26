package preprocess

import (
	"testing"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignScheduledTxsExecutionFactory(t *testing.T) {
	t.Parallel()

	stef, err := NewSovereignScheduledTxsExecutionFactory()
	require.Nil(t, err)
	require.NotNil(t, stef)
	require.IsType(t, &sovereignScheduledTxsExecutionFactory{}, stef)
}

func TestSovereignScheduledTxsExecutionFactory_CreateScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	stef, _ := NewSovereignScheduledTxsExecutionFactory()

	stxeh, err := stef.CreateScheduledTxsExecutionHandler(ScheduledTxsExecutionFactoryArgs{})
	require.Nil(t, err)
	require.NotNil(t, stxeh)
	require.Implements(t, new(process.ScheduledTxsExecutionHandler), stxeh)
}

func TestSovereignScheduledTxsExecutionFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	stef, _ := NewSovereignScheduledTxsExecutionFactory()
	require.False(t, stef.IsInterfaceNil())
}
