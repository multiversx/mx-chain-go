package gin

import (
	"net/http"
	"testing"

	"github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/stretchr/testify/require"
)

func TestNewHttpServer_NilServerShouldErr(t *testing.T) {
	t.Parallel()

	hs, err := NewHttpServer(nil)
	require.Equal(t, errors.ErrNilHttpServer, err)
	require.True(t, check.IfNil(hs))
}

func TestNewHttpServer_ShouldWork(t *testing.T) {
	t.Parallel()

	server := &http.Server{}
	hs, err := NewHttpServer(server)
	require.NoError(t, err)
	require.False(t, check.IfNil(hs))
}

func TestNewHttpServer_StartAndCloseShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		require.Nil(t, r)
	}()

	server := &http.Server{}
	hs, _ := NewHttpServer(server)

	hs.Start()
	err := hs.Close()
	require.NoError(t, err)
}
