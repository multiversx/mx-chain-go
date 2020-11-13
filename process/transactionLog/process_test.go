package transactionLog_test

import (
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/process/transactionLog"
	"github.com/stretchr/testify/require"
)

func TestNewTxLogProcessor_NilParameters(t *testing.T) {
	_, nilMarshalizer := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &mock.StorerStub{},
	})

	require.Equal(t, process.ErrNilMarshalizer, nilMarshalizer)

	_, nilStorer := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Marshalizer: &mock.MarshalizerMock{},
	})

	require.Equal(t, process.ErrNilStore, nilStorer)

	_, nilError := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:      &mock.StorerStub{},
		Marshalizer: &mock.MarshalizerMock{},
	})

	require.Nil(t, nilError)
}

func TestTxLogProcessor_SaveLogsNilTxHash(t *testing.T) {
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:      &mock.StorerStub{},
		Marshalizer: &mock.MarshalizerMock{},
	})

	err := txLogProcessor.SaveLog(nil, nil, make([]*vmcommon.LogEntry, 0))
	require.Equal(t, process.ErrNilTxHash, err)
}

func TestTxLogProcessor_SaveLogsNilTx(t *testing.T) {
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:      &mock.StorerStub{},
		Marshalizer: &mock.MarshalizerMock{},
	})

	err := txLogProcessor.SaveLog([]byte("txhash"), nil, make([]*vmcommon.LogEntry, 0))
	require.Equal(t, process.ErrNilTransaction, err)
}

func TestTxLogProcessor_SaveLogsEmptyLogsReturnsNil(t *testing.T) {
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:      &mock.StorerStub{},
		Marshalizer: &mock.MarshalizerMock{},
	})

	err := txLogProcessor.SaveLog([]byte("txhash"), &transaction.Transaction{}, make([]*vmcommon.LogEntry, 0))
	require.Nil(t, err)
}

func TestTxLogProcessor_SaveLogsMarshalErr(t *testing.T) {
	retErr := errors.New("marshal err")
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &mock.StorerStub{},
		Marshalizer: &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, err error) {
				return nil, retErr
			},
		},
	})

	logs := []*vmcommon.LogEntry{
		{Address: []byte("first log")},
	}
	err := txLogProcessor.SaveLog([]byte("txhash"), &transaction.Transaction{}, logs)
	require.Equal(t, retErr, err)
}

func TestTxLogProcessor_SaveLogsStoreErr(t *testing.T) {
	retErr := errors.New("put err")
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &mock.StorerStub{
			PutCalled: func(key, data []byte) error {
				return retErr
			},
		},
		Marshalizer: &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, err error) {
				return nil, nil
			},
		},
	})

	logs := []*vmcommon.LogEntry{
		{Address: []byte("first log")},
	}
	err := txLogProcessor.SaveLog([]byte("txhash"), &transaction.Transaction{}, logs)
	require.Equal(t, retErr, err)
}

func TestTxLogProcessor_SaveLogsCallsPutWithMarshalBuff(t *testing.T) {
	buffExpected := []byte("marshaled log")
	buffActual := []byte("currently wrong value")
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &mock.StorerStub{
			PutCalled: func(key, data []byte) error {
				buffActual = data
				return nil
			},
		},
		Marshalizer: &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, err error) {
				return buffExpected, nil
			},
		},
	})

	logs := []*vmcommon.LogEntry{
		{Address: []byte("first log")},
	}
	_ = txLogProcessor.SaveLog([]byte("txhash"), &transaction.Transaction{}, logs)

	require.Equal(t, buffExpected, buffActual)
}

func TestTxLogProcessor_GetLogErrNotFound(t *testing.T) {
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &mock.StorerStub{
			GetCalled: func(key []byte) (bytes []byte, err error) {
				return nil, errors.New("storer error")
			},
		},
		Marshalizer: &mock.MarshalizerStub{},
	})

	_, err := txLogProcessor.GetLog([]byte("texhash"))

	require.Equal(t, process.ErrLogNotFound, err)
}

func TestTxLogProcessor_GetLogUnmarshalErr(t *testing.T) {
	retErr := errors.New("marshal error")
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &mock.StorerStub{
			GetCalled: func(key []byte) (bytes []byte, err error) {
				return make([]byte, 0), nil
			},
		},
		Marshalizer: &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return retErr
			},
		},
	})

	_, err := txLogProcessor.GetLog([]byte("texhash"))

	require.Equal(t, retErr, err)
}
