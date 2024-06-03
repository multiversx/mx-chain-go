package process

import (
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignBlockProcessorFactory(t *testing.T) {
	t.Parallel()

	fact := NewSovereignBlockProcessorFactory()

	require.False(t, fact.IsInterfaceNil())
	require.IsType(t, new(sovereignBlockProcessorFactory), fact)
}

func TestNewSovereignBlockProcessorFactory_ProcessBlock(t *testing.T) {
	t.Parallel()

	t.Run("nil block processor should error", func(t *testing.T) {
		fact := NewSovereignBlockProcessorFactory()
		header, block, err := fact.ProcessBlock(nil, &testscommon.HeaderHandlerStub{})
		require.ErrorIs(t, err, process.ErrNilBlockProcessor)
		require.Nil(t, header)
		require.Nil(t, block)
	})
	t.Run("nil header handler should error", func(t *testing.T) {
		fact := NewSovereignBlockProcessorFactory()
		header, block, err := fact.ProcessBlock(&testscommon.BlockProcessorStub{}, nil)
		require.ErrorIs(t, err, process.ErrNilHeaderHandler)
		require.Nil(t, header)
		require.Nil(t, block)
	})
	t.Run("should work", func(t *testing.T) {
		fact := NewSovereignBlockProcessorFactory()

		createWasCalled := false
		processWasCalled := false
		bp := &testscommon.BlockProcessorStub{
			CreateBlockCalled: func(initialHdrData data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
				createWasCalled = true
				return nil, nil, nil
			},
			ProcessBlockCalled: func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
				processWasCalled = true
				return nil, nil, nil
			},
		}

		_, _, err := fact.ProcessBlock(bp, &testscommon.HeaderHandlerStub{})
		require.Nil(t, err)
		require.Equal(t, true, createWasCalled)
		require.Equal(t, true, processWasCalled)
	})
}
