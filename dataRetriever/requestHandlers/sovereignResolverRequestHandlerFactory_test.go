package requestHandlers

import (
	"testing"

	"github.com/multiversx/mx-chain-go/errors"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignResolverRequestHandlerFactory(t *testing.T) {
	t.Parallel()

	rrhf, err := NewSovereignResolverRequestHandlerFactory(nil)

	require.Nil(t, rrhf)
	require.Equal(t, err, errors.ErrNilResolverRequestFactoryHandler)

	rf, _ := NewResolverRequestHandlerFactory()
	rrhf, err = NewSovereignResolverRequestHandlerFactory(rf)

	require.Nil(t, err)
	require.NotNil(t, rrhf)
	require.IsType(t, &sovereignResolverRequestHandlerFactory{}, rrhf)
}

func TestSovereignResolverRequestHandlerFactory_CreateResolverRequestHandler(t *testing.T) {
	t.Parallel()

	rf, _ := NewResolverRequestHandlerFactory()
	rrhf, _ := NewSovereignResolverRequestHandlerFactory(rf)

	rrh, err := rrhf.CreateRequestHandler(RequestHandlerArgs{})
	require.NotNil(t, err)
	require.Nil(t, rrh)

	rrh, err = rrhf.CreateRequestHandler(getDefaultArgs())

	require.Nil(t, err)
	require.NotNil(t, rrh)
	require.IsType(t, &sovereignResolverRequestHandler{}, rrh)
}

func TestSovereignResolverRequestHandlerFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	rf, _ := NewResolverRequestHandlerFactory()
	rrhf, _ := NewSovereignResolverRequestHandlerFactory(rf)

	require.False(t, rrhf.IsInterfaceNil())

	rrhf = (*sovereignResolverRequestHandlerFactory)(nil)
	require.True(t, rrhf.IsInterfaceNil())
}
