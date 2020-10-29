package indexer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadResponseBody_NilBodyNilDest(t *testing.T) {
	t.Parallel()

	err := loadResponseBody(nil, nil)
	require.NoError(t, err)
}

func TestLoadResponseBody_NilBodyNotNilDest(t *testing.T) {
	t.Parallel()

	err := loadResponseBody(nil, struct{}{})
	require.NoError(t, err)
}


