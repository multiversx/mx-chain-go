package incomingEventsProc

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEventProcExecutedDepositOperation(t *testing.T) {
	t.Parallel()

	t.Run("nil deposit event proc, should return error", func(t *testing.T) {
		handler, err := NewEventProcExecutedDepositOperation(nil, &eventProcConfirmExecutedOperation{})
		require.Equal(t, errNilEventProcDepositTokens, err)
		require.Nil(t, handler)
	})

	t.Run("nil confirm executed event proc, should return error", func(t *testing.T) {
		handler, err := NewEventProcExecutedDepositOperation(&eventProcDepositTokens{}, nil)
		require.Equal(t, errNilEventProcConfirmExecutedOp, err)
		require.Nil(t, handler)
	})

	t.Run("should work", func(t *testing.T) {
		handler, err := NewEventProcExecutedDepositOperation(&eventProcDepositTokens{}, &eventProcConfirmExecutedOperation{})
		require.NotNil(t, handler)
		require.Nil(t, err)
	})

}
