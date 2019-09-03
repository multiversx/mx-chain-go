package processor_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/stretchr/testify/assert"
)

func createDefaultArgument() *processor.ArgTxInterceptorProcessor {
	return &processor.ArgTxInterceptorProcessor{
		ShardedDataCache: &mock.ShardedDataStub{},
	}
}

func TestNewTxInterceptorProcessor_NilArgumentShouldErr(t *testing.T) {
	t.Parallel()

	txip, err := processor.NewTxInterceptorProcessor(nil)

	assert.Nil(t, txip)
	assert.Equal(t, process.ErrNilArguments, err)
}

func TestNewTxInterceptorProcessor_NilDataPoolShouldErr(t *testing.T) {
	t.Parallel()

	txip, err := processor.NewTxInterceptorProcessor(&processor.ArgTxInterceptorProcessor{})

	assert.Nil(t, txip)
	assert.Equal(t, process.ErrNilDataPoolHolder, err)
}

func TestNewTxInterceptorProcessor_ShouldWork(t *testing.T) {
	t.Parallel()

	txip, err := processor.NewTxInterceptorProcessor(createDefaultArgument())

	assert.False(t, check.ForNil(txip))
	assert.Nil(t, err)
}

//------- Validate

func TestTxInterceptorProcessor_ValidateShouldRetTrue(t *testing.T) {
	t.Parallel()

	txip, _ := processor.NewTxInterceptorProcessor(createDefaultArgument())

	err := txip.Validate(nil)

	assert.Nil(t, err)
}

//------- Save

func TestTxInterceptorProcessor_SaveNilDataShouldErr(t *testing.T) {
	t.Parallel()

	txip, _ := processor.NewTxInterceptorProcessor(createDefaultArgument())

	err := txip.Save(nil)

	assert.Equal(t, process.ErrWrongTypeAssertion, err)
}

func TestTxInterceptorProcessor_SaveShouldWork(t *testing.T) {
	t.Parallel()

	addedWasCalled := false
	txInterceptedData := &struct {
		mock.InterceptedDataStub
		mock.InterceptedTxHandlerStub
	}{
		InterceptedDataStub: mock.InterceptedDataStub{},
		InterceptedTxHandlerStub: mock.InterceptedTxHandlerStub{
			SndShardCalled: func() uint32 {
				return 0
			},
			RcvShardCalled: func() uint32 {
				return 0
			},
			HashCalled: func() []byte {
				return make([]byte, 0)
			},
			TransactionCalled: func() data.TransactionHandler {
				return nil
			},
		},
	}
	arg := createDefaultArgument()
	shardedDataCache := arg.ShardedDataCache.(*mock.ShardedDataStub)
	shardedDataCache.AddDataCalled = func(key []byte, data interface{}, cacheId string) {
		addedWasCalled = true
	}

	txip, _ := processor.NewTxInterceptorProcessor(arg)

	err := txip.Save(txInterceptedData)

	assert.Nil(t, err)
	assert.True(t, addedWasCalled)
}

//------- IsInterfaceNil

func TestIsInterfaceNil_NotInstantiatedShouldRetTrue(t *testing.T) {
	t.Parallel()

	txip, _ := processor.NewTxInterceptorProcessor(createDefaultArgument())
	txip = nil

	assert.True(t, check.ForNil(txip))
}
