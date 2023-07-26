package requestHandlers

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/require"
)

func TestNewResolverRequestHandlerFactory(t *testing.T) {
	t.Parallel()

	rrhf, err := NewResolverRequestHandlerFactory()

	require.Nil(t, err)
	require.NotNil(t, rrhf)
	require.IsType(t, &resolverRequestHandlerFactory{}, rrhf)
}

func TestResolverRequestHandlerFactory_CreateResolverRequestHandler(t *testing.T) {
	t.Parallel()

	rrhf, _ := NewResolverRequestHandlerFactory()

	rrh, err := rrhf.CreateRequestHandler(RequestHandlerArgs{})
	require.NotNil(t, err)
	require.Nil(t, rrh)

	rrh, err = rrhf.CreateRequestHandler(getDefaultArgs())

	require.Nil(t, err)
	require.NotNil(t, rrh)
	require.IsType(t, &resolverRequestHandler{}, rrh)
}

func TestResolverRequestHandlerFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	rrhf, _ := NewResolverRequestHandlerFactory()
	require.False(t, rrhf.IsInterfaceNil())
}

func getDefaultArgs() RequestHandlerArgs {
	return RequestHandlerArgs{
		RequestersFinder:      &dataRetriever.RequestersFinderStub{},
		RequestedItemsHandler: &testscommon.RequestedItemsHandlerStub{},
		WhiteListHandler:      &testscommon.WhiteListHandlerStub{},
		MaxTxsToRequest:       100,
		ShardID:               0,
		RequestInterval:       time.Second,
	}
}
