package latestData

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLatestDataProviderFactory_CreateLatestDataProvider(t *testing.T) {
	t.Parallel()

	factory := NewLatestDataProviderFactory()
	require.False(t, factory.IsInterfaceNil())

	ldp, err := factory.CreateLatestDataProvider(getLatestDataProviderArgs())
	require.NotNil(t, ldp)
	require.Nil(t, err)
	require.Equal(t, "*latestData.latestDataProvider", fmt.Sprintf("%T", ldp))
}
