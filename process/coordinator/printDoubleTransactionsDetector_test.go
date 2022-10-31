package coordinator

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockArgsPrintDoubleTransactionsDetector() ArgsPrintDoubleTransactionsDetector {
	return ArgsPrintDoubleTransactionsDetector{
		Marshaller:          &testscommon.MarshalizerMock{},
		Hasher:              &testscommon.HasherStub{},
		EnableEpochsHandler: &testscommon.EnableEpochsHandlerStub{},
	}
}

func TestNewPrintDoubleTransactionsDetector(t *testing.T) {
	t.Parallel()

	t.Run("nil marshaller should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsPrintDoubleTransactionsDetector()
		args.Marshaller = nil

		detector, err := NewPrintDoubleTransactionsDetector(args)
		assert.True(t, check.IfNil(detector))
		assert.Equal(t, process.ErrNilMarshalizer, err)
	})
	t.Run("nil hasher should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsPrintDoubleTransactionsDetector()
		args.Hasher = nil

		detector, err := NewPrintDoubleTransactionsDetector(args)
		assert.True(t, check.IfNil(detector))
		assert.Equal(t, process.ErrNilHasher, err)
	})
	t.Run("nil enable epochs handler should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsPrintDoubleTransactionsDetector()
		args.EnableEpochsHandler = nil

		detector, err := NewPrintDoubleTransactionsDetector(args)
		assert.True(t, check.IfNil(detector))
		assert.Equal(t, process.ErrNilEnableEpochsHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := createMockArgsPrintDoubleTransactionsDetector()

		detector, err := NewPrintDoubleTransactionsDetector(args)
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
		detector, _ := NewPrintDoubleTransactionsDetector(args)
		detector.logger = &testscommon.LoggerStub{
			ErrorCalled: func(message string, args ...interface{}) {
				errorCalled = message == nilBlockBodyMessage
			},
		}

		detector.ProcessBlockBody(nil)
		assert.True(t, errorCalled)
	})
	t.Run("empty block body", func(t *testing.T) {
		t.Parallel()

		debugCalled := false
		args := createMockArgsPrintDoubleTransactionsDetector()
		detector, _ := NewPrintDoubleTransactionsDetector(args)
		detector.logger = &testscommon.LoggerStub{
			ErrorCalled: func(message string, args ...interface{}) {
				assert.Fail(t, "should have not called error")
			},
			DebugCalled: func(message string, args ...interface{}) {
				debugCalled = message == noDoubledTransactionsFoundMessage
			},
		}

		detector.ProcessBlockBody(&block.Body{})
		assert.True(t, debugCalled)
	})
	t.Run("no doubled transactions", func(t *testing.T) {
		t.Parallel()

		debugCalled := false
		args := createMockArgsPrintDoubleTransactionsDetector()
		detector, _ := NewPrintDoubleTransactionsDetector(args)
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
		detector.ProcessBlockBody(body)
		assert.True(t, debugCalled)
	})
	t.Run("doubled transactions in different miniblocks but feature not active", func(t *testing.T) {
		t.Parallel()

		debugCalled := false
		args := createMockArgsPrintDoubleTransactionsDetector()
		args.EnableEpochsHandler = &testscommon.EnableEpochsHandlerStub{
			IsAddFailedRelayedTxToInvalidMBsFlagField: true,
		}
		detector, _ := NewPrintDoubleTransactionsDetector(args)
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
		detector.ProcessBlockBody(body)
		assert.True(t, debugCalled)
	})
	t.Run("doubled transactions in different miniblocks", func(t *testing.T) {
		t.Parallel()

		errorCalled := false
		expectedMessage := printReportHeader + ` miniblock hash , type TxBlock, 0 -> 0
  tx hash 7478206861736831
  tx hash 7478206861736832
 miniblock hash , type TxBlock, 0 -> 0
  tx hash 7478206861736831
  tx hash 7478206861736834
`
		args := createMockArgsPrintDoubleTransactionsDetector()
		detector, _ := NewPrintDoubleTransactionsDetector(args)
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
		detector.ProcessBlockBody(body)
		assert.True(t, errorCalled)
	})
	t.Run("doubled transactions in same miniblock", func(t *testing.T) {
		t.Parallel()

		errorCalled := false
		expectedMessage := printReportHeader + ` miniblock hash , type TxBlock, 0 -> 0
  tx hash 7478206861736831
  tx hash 7478206861736831
 miniblock hash , type TxBlock, 0 -> 0
  tx hash 7478206861736832
  tx hash 7478206861736834
`
		args := createMockArgsPrintDoubleTransactionsDetector()
		detector, _ := NewPrintDoubleTransactionsDetector(args)
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
		detector.ProcessBlockBody(body)
		assert.True(t, errorCalled)
	})
}
