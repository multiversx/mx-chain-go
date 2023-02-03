package processor_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/interceptors/processor"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func createMockTxArgument() *processor.ArgTxInterceptorProcessor {
	return &processor.ArgTxInterceptorProcessor{
		ShardedDataCache: testscommon.NewShardedDataStub(),
		TxValidator:      &mock.TxValidatorStub{},
	}
}

func TestNewTxInterceptorProcessor_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	txip, err := processor.NewTxInterceptorProcessor(nil)

	assert.Nil(t, txip)
	assert.Equal(t, process.ErrNilArgumentStruct, err)
}

func TestNewTxInterceptorProcessor_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockTxArgument()
	arg.ShardedDataCache = nil
	txip, err := processor.NewTxInterceptorProcessor(arg)

	assert.Nil(t, txip)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewTxInterceptorProcessor_NilTxValidatorShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockTxArgument()
	arg.TxValidator = nil
	txip, err := processor.NewTxInterceptorProcessor(arg)

	assert.Nil(t, txip)
	assert.Equal(t, process.ErrNilTxValidator, err)
}

func TestNewTxInterceptorProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	txip, err := processor.NewTxInterceptorProcessor(createMockTxArgument())

	assert.False(t, check.IfNil(txip))
	assert.Nil(t, err)
}

//------- Validate

func TestTxInterceptorProcessor_ValidateNilTxShouldErr(t *testing.T) {
	t.Parallel()

	txip, _ := processor.NewTxInterceptorProcessor(createMockTxArgument())

	err := txip.Validate(nil, "")

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestTxInterceptorProcessor_ValidateReturnsFalseShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("tx validation error")
	arg := createMockTxArgument()
	arg.TxValidator = &mock.TxValidatorStub{
		CheckTxValidityCalled: func(interceptedTx process.InterceptedTransactionHandler) error {
			return expectedErr
		},
	}
	txip, _ := processor.NewTxInterceptorProcessor(arg)

	txInterceptedData := &struct {
		testscommon.InterceptedDataStub
		mock.InterceptedTxHandlerStub
	}{}
	err := txip.Validate(txInterceptedData, "")

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), expectedErr.Error()))
}

func TestTxInterceptorProcessor_ValidateReturnsTrueShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockTxArgument()
	arg.TxValidator = &mock.TxValidatorStub{
		CheckTxValidityCalled: func(interceptedTx process.InterceptedTransactionHandler) error {
			return nil
		},
	}
	txip, _ := processor.NewTxInterceptorProcessor(arg)

	txInterceptedData := &struct {
		testscommon.InterceptedDataStub
		mock.InterceptedTxHandlerStub
	}{}
	err := txip.Validate(txInterceptedData, "")

	assert.Nil(t, err)
}

//------- Save

func TestTxInterceptorProcessor_SaveNilDataShouldErr(t *testing.T) {
	t.Parallel()

	txip, _ := processor.NewTxInterceptorProcessor(createMockTxArgument())

	err := txip.Save(nil, "", "")

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestTxInterceptorProcessor_SaveShouldWork(t *testing.T) {
	t.Parallel()

	addedWasCalled := false
	txInterceptedData := &struct {
		testscommon.InterceptedDataStub
		mock.InterceptedTxHandlerStub
	}{
		InterceptedDataStub: testscommon.InterceptedDataStub{
			HashCalled: func() []byte {
				return make([]byte, 0)
			},
		},
		InterceptedTxHandlerStub: mock.InterceptedTxHandlerStub{
			SenderShardIdCalled: func() uint32 {
				return 0
			},
			ReceiverShardIdCalled: func() uint32 {
				return 0
			},
			TransactionCalled: func() data.TransactionHandler {
				return &transaction.Transaction{}
			},
		},
	}
	arg := createMockTxArgument()
	shardedDataCache := arg.ShardedDataCache.(*testscommon.ShardedDataStub)
	shardedDataCache.AddDataCalled = func(key []byte, data interface{}, sizeInBytes int, cacheId string) {
		addedWasCalled = true
	}

	txip, _ := processor.NewTxInterceptorProcessor(arg)

	err := txip.Save(txInterceptedData, "", "")

	assert.Nil(t, err)
	assert.True(t, addedWasCalled)
}

//------- IsInterfaceNil

func TestTxInterceptorProcessor_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var txip *processor.TxInterceptorProcessor

	assert.True(t, check.IfNil(txip))
}
