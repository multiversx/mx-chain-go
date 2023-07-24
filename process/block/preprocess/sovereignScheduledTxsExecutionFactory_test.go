package preprocess

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSovereignScheduledTxsExecutionFactory(t *testing.T) {
	stef, err := NewSovereignScheduledTxsExecutionFactory()
	require.Nil(t, err)
	require.NotNil(t, stef)
}

func TestSovereignScheduledTxsExecutionFactory_CreateScheduledTxsExecutionHandler(t *testing.T) {
	stef, _ := NewSovereignScheduledTxsExecutionFactory()

	stxeh, err := stef.CreateScheduledTxsExecutionHandler(ScheduledTxsExecutionFactoryArgs{})
	require.Nil(t, err)
	require.NotNil(t, stxeh)
}

func TestSovereignScheduledTxsExecutionFactory_IsInterfaceNil(t *testing.T) {
	stef, _ := NewSovereignScheduledTxsExecutionFactory()
	require.False(t, stef.IsInterfaceNil())
}
