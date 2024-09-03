package requesters

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/stretchr/testify/assert"
)

func createMockArgReceiptRequester(argBase ArgBaseRequester) ArgReceiptRequester {
	return ArgReceiptRequester{
		ArgBaseRequester: argBase,
	}
}

func TestNewReceiptRequester(t *testing.T) {
	t.Parallel()

	t.Run("nil request sender should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockArgBaseRequester()
		argsBase.RequestSender = nil
		requester, err := NewReceiptRequester(createMockArgReceiptRequester(argsBase))
		assert.Equal(t, dataRetriever.ErrNilRequestSender, err)
		assert.True(t, check.IfNil(requester))
	})
	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		argsBase := createMockArgBaseRequester()
		argsBase.Marshaller = nil
		requester, err := NewReceiptRequester(createMockArgReceiptRequester(argsBase))
		assert.Equal(t, dataRetriever.ErrNilMarshalizer, err)
		assert.True(t, check.IfNil(requester))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		requester, err := NewReceiptRequester(createMockArgReceiptRequester(createMockArgBaseRequester()))
		assert.Nil(t, err)
		assert.False(t, check.IfNil(requester))
	})
}
