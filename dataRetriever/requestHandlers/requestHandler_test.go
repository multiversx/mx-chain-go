package requestHandlers

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMetaResolverRequestHandlerNilFinder(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(nil, "topic")

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilResolverFinder, err)
}

func TestNewMetaResolverRequestHandlerEmptyTopic(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(&mock.ResolversFinderStub{}, "")

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrEmptyHeaderRequestTopic, err)
}

func TestNewMetaResolverRequestHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewMetaResolverRequestHandler(&mock.ResolversFinderStub{}, "topic")

	assert.Nil(t, err)
	assert.NotNil(t, rrh)
}

func TestNewShardResolverRequestHandlerNilFinder(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(nil, "topic", "topic", "topic", 1)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrNilResolverFinder, err)
}

func TestNewShardResolverRequestHandlerTxTopicEmpty(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(&mock.ResolversFinderStub{}, "", "topic", "topic", 1)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrEmptyTxRequestTopic, err)
}

func TestNewShardResolverRequestHandlerMBTopicEmpty(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(&mock.ResolversFinderStub{}, "topic", "", "topic", 1)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrEmptyMiniBlockRequestTopic, err)
}

func TestNewShardResolverRequestHandlerHdrTopicEmpty(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(&mock.ResolversFinderStub{}, "topic", "topic", "", 1)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrEmptyHeaderRequestTopic, err)
}

func TestNewShardResolverRequestHandlerMaxTxRequestTooSmall(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(&mock.ResolversFinderStub{}, "topic", "topic", "topic", 0)

	assert.Nil(t, rrh)
	assert.Equal(t, dataRetriever.ErrInvalidMaxTxRequest, err)
}

func TestNewShardResolverRequestHandler(t *testing.T) {
	t.Parallel()

	rrh, err := NewShardResolverRequestHandler(&mock.ResolversFinderStub{}, "topic", "topic", "topic", 1)

	assert.Nil(t, err)
	assert.NotNil(t, rrh)
}
