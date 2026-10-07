package dblookupext

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestSaveExecutionResultsFromHeader(t *testing.T) {
	t.Parallel()

	t.Run("header v1 should do nothing", func(t *testing.T) {
		t.Parallel()

		called := false
		storer := &storage.StorerStub{
			PutInEpochCalled: func(key, data []byte, epoch uint32) error {
				called = true
				return nil
			},
		}
		marshaller := marshallerMock.MarshalizerMock{}
		proc := newExecutionResultsProcessor(storer, marshaller)

		header := &block.Header{}
		err := proc.saveExecutionResultsFromHeader(header)
		require.NoError(t, err)
		require.False(t, called)
	})

	t.Run("header v2 should do nothing", func(t *testing.T) {
		t.Parallel()

		called := false
		storer := &storage.StorerStub{
			PutInEpochCalled: func(key, data []byte, epoch uint32) error {
				called = true
				return nil
			},
		}
		marshaller := marshallerMock.MarshalizerMock{}
		proc := newExecutionResultsProcessor(storer, marshaller)

		header := &block.HeaderV2{}
		err := proc.saveExecutionResultsFromHeader(header)
		require.NoError(t, err)
		require.False(t, called)
	})

	t.Run("header v3 no execution results", func(t *testing.T) {
		t.Parallel()

		called := false
		storer := &storage.StorerStub{
			PutInEpochCalled: func(key, data []byte, epoch uint32) error {
				called = true
				return nil
			},
		}
		marshaller := marshallerMock.MarshalizerMock{}
		proc := newExecutionResultsProcessor(storer, marshaller)

		header := &block.HeaderV3{}
		err := proc.saveExecutionResultsFromHeader(header)
		require.NoError(t, err)
		require.False(t, called)
	})

	t.Run("header v3 with execution result should error", func(t *testing.T) {
		t.Parallel()

		localError := errors.New("local error")
		storer := &storage.StorerStub{
			PutInEpochCalled: func(key, data []byte, epoch uint32) error {
				return localError
			},
		}
		marshaller := marshallerMock.MarshalizerMock{}
		proc := newExecutionResultsProcessor(storer, marshaller)

		header := &block.HeaderV3{
			ExecutionResults: []*block.ExecutionResult{
				{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderNonce: 1,
					},
				},
			},
		}
		err := proc.saveExecutionResultsFromHeader(header)
		require.Equal(t, localError, err)
	})

	t.Run("header v3 with execution result should call put in epoch", func(t *testing.T) {
		t.Parallel()

		called := false
		storer := &storage.StorerStub{
			PutInEpochCalled: func(key, data []byte, epoch uint32) error {
				called = true
				return nil
			},
		}
		marshaller := marshallerMock.MarshalizerMock{}
		proc := newExecutionResultsProcessor(storer, marshaller)

		header := &block.HeaderV3{
			ExecutionResults: []*block.ExecutionResult{
				nil,
				{
					BaseExecutionResult: &block.BaseExecutionResult{
						HeaderNonce: 1,
					},
				},
			},
		}
		err := proc.saveExecutionResultsFromHeader(header)
		require.NoError(t, err)
		require.True(t, called)
	})

}
