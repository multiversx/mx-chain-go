package coordinator

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsPrintDoubleTransactionsDetector() ArgsDoubleTransactionsDetector {
	return ArgsDoubleTransactionsDetector{
		Marshaller:          &marshallerMock.MarshalizerMock{},
		Hasher:              &testscommon.HasherStub{},
		EnableEpochsHandler: enableEpochsHandlerMock.NewEnableEpochsHandlerStub(),
	}
}

func TestNewPrintDoubleTransactionsDetector(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsPrintDoubleTransactionsDetector()
		args.Marshaller = nil

		detector, err := NewDoubleTransactionsDetector(args)
		assert.True(t, check.IfNil(detector))
		assert.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("nil hasher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsPrintDoubleTransactionsDetector()
		args.Hasher = nil

		detector, err := NewDoubleTransactionsDetector(args)
		assert.True(t, check.IfNil(detector))
		assert.Equal(t, process.ErrNilHasher, err)
	})
	t.Run("nil enable epochs handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsPrintDoubleTransactionsDetector()
		args.EnableEpochsHandler = nil

		detector, err := NewDoubleTransactionsDetector(args)
		assert.True(t, check.IfNil(detector))
		assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
	})
	t.Run("invalid enable epochs handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsPrintDoubleTransactionsDetector()
		args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStubWithNoFlagsDefined()

		detector, err := NewDoubleTransactionsDetector(args)
		assert.True(t, check.IfNil(detector))
		assert.True(t, errors.Is(err, core.ErrInvalidEnableEpochsHandler))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsPrintDoubleTransactionsDetector()

		detector, err := NewDoubleTransactionsDetector(args)
		assert.False(t, check.IfNil(detector))
		assert.Nil(t, err)
	})
}

func TestPrintDoubleTransactionsDetector_ProcessBlockBody(t *testing.T) {
	t.Parallel()

	t.Run("nil block body", func(t *testing.T) {
		t.Parallel()

		errorCalled := false
		args := createMockArgsPrintDoubleTransactionsDetector()
		detector, _ := NewDoubleTransactionsDetector(args)
		detector.logger = &testscommon.LoggerStub{
			ErrorCalled: func(message string, args ...interface{}) {
				errorCalled = message == nilBlockBodyMessage
			},
		}

		err := detector.ProcessBlockBody(nil)
		require.Nil(t, err)
		assert.True(t, errorCalled)
	})
	t.Run("empty block body", func(t *testing.T) {
		t.Parallel()

		debugCalled := false
		args := createMockArgsPrintDoubleTransactionsDetector()
		detector, _ := NewDoubleTransactionsDetector(args)
		detector.logger = &testscommon.LoggerStub{
			ErrorCalled: func(message string, args ...interface{}) {
				assert.Fail(t, "should have not called error")
			},
			DebugCalled: func(message string, args ...interface{}) {
				debugCalled = message == noDoubledTransactionsFoundMessage
			},
		}

		err := detector.ProcessBlockBody(&block.Body{})
		require.Nil(t, err)
		assert.True(t, debugCalled)
	})
	t.Run("no doubled transactions", func(t *testing.T) {
		t.Parallel()

		debugCalled := false
		args := createMockArgsPrintDoubleTransactionsDetector()
		detector, _ := NewDoubleTransactionsDetector(args)
		detector.logger = &testscommon.LoggerStub{
			ErrorCalled: func(message string, args ...interface{}) {
				assert.Fail(t, "should have not called error")
			},
			DebugCalled: func(message string, args ...interface{}) {
				debugCalled = message == noDoubledTransactionsFoundMessage
			},
		}

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{[]byte("tx hash1"), []byte("tx hash2")},
				},
				{
					TxHashes: [][]byte{[]byte("tx hash3"), []byte("tx hash4")},
				},
			},
		}
		err := detector.ProcessBlockBody(body)
		require.Nil(t, err)
		assert.True(t, debugCalled)
	})
	t.Run("doubled transactions in different miniblocks but feature not active", func(t *testing.T) {
		t.Parallel()

		debugCalled := false
		args := createMockArgsPrintDoubleTransactionsDetector()
		args.EnableEpochsHandler = enableEpochsHandlerMock.NewEnableEpochsHandlerStub(common.AddFailedRelayedTxToInvalidMBsFlag)
		detector, _ := NewDoubleTransactionsDetector(args)
		detector.logger = &testscommon.LoggerStub{
			ErrorCalled: func(message string, args ...interface{}) {
				assert.Fail(t, "should have not called error")
			},
			DebugCalled: func(message string, args ...interface{}) {
				debugCalled = message == doubledTransactionsFoundButFlagActive
			},
		}

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{[]byte("tx hash1"), []byte("tx hash2")},
				},
				{
					TxHashes: [][]byte{[]byte("tx hash1"), []byte("tx hash4")},
				},
			},
		}
		err := detector.ProcessBlockBody(body)
		require.Nil(t, err)
		assert.True(t, debugCalled)
	})
	t.Run("doubled transactions in different miniblocks", func(t *testing.T) {
		t.Parallel()

		errorCalled := false
		expectedMessage := printReportHeaderNotCritical + ` miniblock hash , type TxBlock, 0 -> 0
  tx hash 7478206861736831
  tx hash 7478206861736832
 miniblock hash , type TxBlock, 0 -> 0
  tx hash 7478206861736831
  tx hash 7478206861736834
`
		args := createMockArgsPrintDoubleTransactionsDetector()
		detector, _ := NewDoubleTransactionsDetector(args)
		detector.logger = &testscommon.LoggerStub{
			ErrorCalled: func(message string, args ...interface{}) {
				assert.Equal(t, expectedMessage, message)
				errorCalled = message == expectedMessage
			},
			DebugCalled: func(message string, args ...interface{}) {
				assert.Fail(t, "should have not called debug")
			},
		}

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{[]byte("tx hash1"), []byte("tx hash2")},
				},
				{
					TxHashes: [][]byte{[]byte("tx hash1"), []byte("tx hash4")},
				},
			},
		}
		err := detector.ProcessBlockBody(body)
		require.Nil(t, err)
		assert.True(t, errorCalled)
	})
	t.Run("doubled transactions in same miniblock", func(t *testing.T) {
		t.Parallel()

		errorCalled := false
		expectedMessage := printReportHeaderNotCritical + ` miniblock hash , type TxBlock, 0 -> 0
  tx hash 7478206861736831
  tx hash 7478206861736831
 miniblock hash , type TxBlock, 0 -> 0
  tx hash 7478206861736832
  tx hash 7478206861736834
`
		args := createMockArgsPrintDoubleTransactionsDetector()
		detector, _ := NewDoubleTransactionsDetector(args)
		detector.logger = &testscommon.LoggerStub{
			ErrorCalled: func(message string, args ...interface{}) {
				assert.Equal(t, expectedMessage, message)
				errorCalled = message == expectedMessage
			},
			DebugCalled: func(message string, args ...interface{}) {
				assert.Fail(t, "should have not called debug")
			},
		}

		body := &block.Body{
			MiniBlocks: []*block.MiniBlock{
				{
					TxHashes: [][]byte{[]byte("tx hash1"), []byte("tx hash1")},
				},
				{
					TxHashes: [][]byte{[]byte("tx hash2"), []byte("tx hash4")},
				},
			},
		}
		err := detector.ProcessBlockBody(body)
		require.Nil(t, err)
		assert.True(t, errorCalled)
	})
}
