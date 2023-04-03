package requestHandlers_test

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/dataRetriever/requestHandlers"
	"github.com/multiversx/mx-chain-go/process"
	dataRetrieverMocks "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/stretchr/testify/assert"
)

func TestNewSovereignResolverRequestHandler_ShouldErrNilRequestHandler(t *testing.T) {
	t.Parallel()

	srrh, err := requestHandlers.NewSovereignResolverRequestHandler(nil)
	assert.Nil(t, srrh)
	assert.Equal(t, process.ErrNilRequestHandler, err)
}

func TestNewSovereignResolverRequestHandler_ShouldWork(t *testing.T) {
	t.Parallel()

	rrh, _ := requestHandlers.NewResolverRequestHandler(
		&dataRetrieverMocks.RequestersFinderStub{},
		&mock.RequestedItemsHandlerStub{},
		&mock.WhiteListHandlerStub{},
		1,
		0,
		time.Second,
	)

	srrh, err := requestHandlers.NewSovereignResolverRequestHandler(rrh)
	assert.NotNil(t, srrh)
	assert.Nil(t, err)
}
