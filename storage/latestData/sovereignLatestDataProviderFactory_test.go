package latestData

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSovereignLatestDataProviderFactory_CreateLatestDataProvider(t *testing.T) {
	t.Parallel()

	factory := NewSovereignLatestDataProviderFactory()
	require.False(t, factory.IsInterfaceNil())

	ldp, err := factory.CreateLatestDataProvider(getLatestDataProviderArgs())
	require.NotNil(t, ldp)
	require.Nil(t, err)
	require.Equal(t, "*latestData.sovereignLatestDataProvider", fmt.Sprintf("%T", ldp))
}
