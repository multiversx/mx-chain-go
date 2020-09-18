package indexer

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewDataDispatcher_InvalidCacheSize(t *testing.T) {
	t.Parallel()

	dataDist, err := NewDataDispatcher(-1)

	require.Nil(t, dataDist)
	require.Equal(t, ErrNegativeCacheSize, err)
}
