package gin

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/errors"
	"github.com/stretchr/testify/require"
)

func TestNewHttpServer_NilServerShouldErr(t *testing.T) {
	t.Parallel()

	hs, err := NewHttpServer(nil)
	require.Equal(t, errors.ErrNilHttpServer, err)
	require.True(t, check.IfNil(hs))
}
