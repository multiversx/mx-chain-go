package requestHandlers

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/require"
)

func TestNewResolverRequestHandlerFactory(t *testing.T) {
	rrhf, err := NewResolverRequestHandlerFactory()

	require.Nil(t, err)
	require.NotNil(t, rrhf)
}

func TestResolverRequestHandlerFactory_CreateResolverRequestHandler(t *testing.T) {
	rrhf, _ := NewResolverRequestHandlerFactory()

	rrh, err := rrhf.CreateResolverRequestHandler(GetDefaultArgs())

	require.Nil(t, err)
	require.NotNil(t, rrh)
}

func TestResolverRequestHandlerFactory_IsInterfaceNil(t *testing.T) {
	rrhf, _ := NewResolverRequestHandlerFactory()

	require.False(t, rrhf.IsInterfaceNil())
}

func GetDefaultArgs() ResolverRequestArgs {
	return ResolverRequestArgs{
		RequestersFinder:      &dataRetriever.RequestersFinderStub{},
		RequestedItemsHandler: &testscommon.RequestedItemsHandlerStub{},
		WhiteListHandler:      &testscommon.WhiteListHandlerStub{},
		MaxTxsToRequest:       100,
		ShardID:               0,
		RequestInterval:       time.Second,
	}
}
