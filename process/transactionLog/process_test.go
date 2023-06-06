package transactionLog_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/transactionLog"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewTxLogProcessor_NilParameters(t *testing.T) {
	_, nilMarshalizer := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &storageStubs.StorerStub{},
	})

	require.Equal(t, process.ErrNilMarshalizer, nilMarshalizer)

	_, nilStorer := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Marshalizer:          &mock.MarshalizerMock{},
		SaveInStorageEnabled: true,
	})

	require.Equal(t, process.ErrNilStore, nilStorer)

	_, nilError := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:      &storageStubs.StorerStub{},
		Marshalizer: &mock.MarshalizerMock{},
	})

	require.Nil(t, nilError)
}

func TestTxLogProcessor_SaveLogsNilTxHash(t *testing.T) {
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:      &storageStubs.StorerStub{},
		Marshalizer: &mock.MarshalizerMock{},
	})

	err := txLogProcessor.SaveLog(nil, nil, make([]*vmcommon.LogEntry, 0))
	require.Equal(t, process.ErrNilTxHash, err)
}

func TestTxLogProcessor_SaveLogsNilTx(t *testing.T) {
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:      &storageStubs.StorerStub{},
		Marshalizer: &mock.MarshalizerMock{},
	})

	err := txLogProcessor.SaveLog([]byte("txhash"), nil, make([]*vmcommon.LogEntry, 0))
	require.Equal(t, process.ErrNilTransaction, err)
}

func TestTxLogProcessor_SaveLogsEmptyLogsReturnsNil(t *testing.T) {
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:      &storageStubs.StorerStub{},
		Marshalizer: &mock.MarshalizerMock{},
	})

	err := txLogProcessor.SaveLog([]byte("txhash"), &transaction.Transaction{}, make([]*vmcommon.LogEntry, 0))
	require.Nil(t, err)
}

func TestTxLogProcessor_Clean(t *testing.T) {
	t.Parallel()

	txLogsProc, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer:      &storageStubs.StorerStub{},
		Marshalizer: &mock.MarshalizerMock{},
	})

	logs := []*vmcommon.LogEntry{
		{Address: []byte("first log")},
	}
	err := txLogsProc.SaveLog([]byte("txhash"), &transaction.Transaction{}, logs)
	require.Nil(t, err)
	require.Len(t, txLogsProc.GetAllCurrentLogs(), 1)

	txLogsProc.Clean()
	require.Len(t, txLogsProc.GetAllCurrentLogs(), 0)
}

func TestTxLogProcessor_SaveLogsMarshalErr(t *testing.T) {
	retErr := errors.New("marshal err")
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &storageStubs.StorerStub{},
		Marshalizer: &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, err error) {
				return nil, retErr
			},
		},
		SaveInStorageEnabled: true,
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
		Storer: &storageStubs.StorerStub{
			PutCalled: func(key, data []byte) error {
				return retErr
			},
		},
		Marshalizer: &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, err error) {
				return nil, nil
			},
		},
		SaveInStorageEnabled: true,
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
	expectedLogData := [][]byte{[]byte("data1"), []byte("data2")}

	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &storageStubs.StorerStub{
			PutCalled: func(key, data []byte) error {
				buffActual = data
				return nil
			},
		},
		Marshalizer: &mock.MarshalizerStub{
			MarshalCalled: func(obj interface{}) (bytes []byte, err error) {
				log, _ := obj.(*transaction.Log)
				require.Equal(t, expectedLogData[0], log.Events[0].Data)
				require.Equal(t, expectedLogData, log.Events[0].AdditionalData)
				return buffExpected, nil
			},
		},
		SaveInStorageEnabled: true,
	})

	logs := []*vmcommon.LogEntry{
		{Address: []byte("first log"), Data: expectedLogData},
	}
	_ = txLogProcessor.SaveLog([]byte("txhash"), &transaction.Transaction{}, logs)

	require.Equal(t, buffExpected, buffActual)
}

func TestTxLogProcessor_GetLogErrNotFound(t *testing.T) {
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &storageStubs.StorerStub{
			GetCalled: func(key []byte) (bytes []byte, err error) {
				return nil, errors.New("storer error")
			},
		},
		Marshalizer:          &mock.MarshalizerStub{},
		SaveInStorageEnabled: true,
	})

	_, err := txLogProcessor.GetLog([]byte("texhash"))

	require.Equal(t, process.ErrLogNotFound, err)
}

func TestTxLogProcessor_GetLogUnmarshalErr(t *testing.T) {
	retErr := errors.New("marshal error")
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &storageStubs.StorerStub{
			GetCalled: func(key []byte) (bytes []byte, err error) {
				return make([]byte, 0), nil
			},
		},
		Marshalizer: &mock.MarshalizerStub{
			UnmarshalCalled: func(obj interface{}, buff []byte) error {
				return retErr
			},
		},
		SaveInStorageEnabled: true,
	})

	_, err := txLogProcessor.GetLog([]byte("texhash"))

	require.Equal(t, retErr, err)
}

func TestTxLogProcessor_GetLogFromCache(t *testing.T) {
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &storageStubs.StorerStub{
			PutCalled: func(key, data []byte) error {
				return nil
			},
		},
		Marshalizer: &mock.MarshalizerMock{},
	})
	txLogProcessor.EnableLogToBeSavedInCache()
	_ = txLogProcessor.SaveLog([]byte("txhash"), &transaction.Transaction{}, []*vmcommon.LogEntry{{}})

	logData, found := txLogProcessor.GetLogFromCache([]byte("txhash"))
	require.True(t, found)
	require.Equal(t, "txhash", logData.TxHash)
}

func TestTxLogProcessor_GetLogFromCacheNotInCacheShouldReturnFromStorage(t *testing.T) {
	t.Parallel()

	logs := []*vmcommon.LogEntry{{
		Address: []byte("my-addr"),
	}}

	txLog := &transaction.Log{
		Address: []byte("add"),
	}

	marshalizer := &mock.MarshalizerMock{}
	txLogProcessor, _ := transactionLog.NewTxLogProcessor(transactionLog.ArgTxLogProcessor{
		Storer: &storageStubs.StorerStub{
			PutCalled: func(key, data []byte) error {
				return nil
			},
			GetCalled: func(key []byte) ([]byte, error) {
				logsBytes, _ := marshalizer.Marshal(txLog)
				return logsBytes, nil
			},
		},
		Marshalizer: marshalizer,
	})
	_ = txLogProcessor.SaveLog([]byte("txhash"), &transaction.Transaction{}, logs)

	_, found := txLogProcessor.GetLogFromCache([]byte("txhash"))
	require.True(t, found)
}
