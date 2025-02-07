package preprocess

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSovereignScheduledTxsExecutionFactory(t *testing.T) {
	t.Parallel()

	stef := NewSovereignScheduledTxsExecutionFactory()
	require.IsType(t, &sovereignScheduledTxsExecutionFactory{}, stef)
	require.False(t, stef.IsInterfaceNil())
}

func TestSovereignScheduledTxsExecutionFactory_CreateScheduledTxsExecutionHandler(t *testing.T) {
	t.Parallel()

	stef := NewSovereignScheduledTxsExecutionFactory()

	stxeh, err := stef.CreateScheduledTxsExecutionHandler(ScheduledTxsExecutionFactoryArgs{})
	require.Nil(t, err)
	require.Equal(t, "*disabled.ScheduledTxsExecutionHandler", fmt.Sprintf("%T", stxeh))
}
