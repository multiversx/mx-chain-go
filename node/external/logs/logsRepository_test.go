package logs

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	storageStubs "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestNewLogsRepository(t *testing.T) {
	t.Parallel()

	t.Run("storer not found", func(t *testing.T) {
		t.Parallel()

		repository := newLogsRepository(&storageStubs.ChainStorerStub{
			GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
				return nil, errors.New("new error")
			},
		}, marshallerMock.MarshalizerMock{})
		require.Nil(t, repository)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		repository := newLogsRepository(&genericMocks.ChainStorerMock{}, marshallerMock.MarshalizerMock{})
		require.NotNil(t, repository)
	})
}

func TestLogsRepository_GetLogsShouldWork(t *testing.T) {
	epoch := uint32(7)

	storageService := genericMocks.NewChainStorerMock(epoch)
	marshaller := &marshal.GogoProtoMarshalizer{}

	firstLog := &transaction.Log{Events: []*transaction.Event{{Identifier: []byte("first")}}}
	secondLog := &transaction.Log{Events: []*transaction.Event{{Identifier: []byte("second")}}}

	firstLogBytes, _ := marshaller.Marshal(firstLog)
	secondLogBytes, _ := marshaller.Marshal(secondLog)
	_ = storageService.Logs.Put([]byte{0xaa}, firstLogBytes)
	_ = storageService.Logs.Put([]byte{0xbb}, secondLogBytes)

	repository := newLogsRepository(storageService, marshaller)

	firstLogFetched, err := repository.getLog([]byte{0xaa}, epoch)
	require.Nil(t, err)
	require.Equal(t, []byte("first"), firstLogFetched.Events[0].Identifier)

	secondLogFetched, err := repository.getLog([]byte{0xbb}, epoch)
	require.Nil(t, err)
	require.Equal(t, []byte("second"), secondLogFetched.Events[0].Identifier)

	bothLogsFetched, err := repository.getLogs([][]byte{{0xaa}, {0xbb}}, epoch)
	require.Nil(t, err)
	require.Len(t, bothLogsFetched, 2)

	noLogsFetched, err := repository.getLogs([][]byte{{0xcc}, {0xdd}}, epoch)
	require.Nil(t, err)
	require.Len(t, noLogsFetched, 0)
}

func TestLogsRepository_GetLogShouldFallbackToPreviousEpoch(t *testing.T) {
	storageService := genericMocks.NewChainStorerMock(uint32(0))
	marshaller := &marshal.GogoProtoMarshalizer{}
	repository := newLogsRepository(storageService, marshaller)

	logEntry := &transaction.Log{Events: []*transaction.Event{{Identifier: []byte("foo")}}}
	logEntryBytes, _ := marshaller.Marshal(logEntry)
	_ = storageService.Logs.PutInEpoch([]byte{0xaa}, logEntryBytes, 41)

	// logEntry is missing in epoch 42 (edge-case), but is present in epoch 41
	logEntryFetched, err := repository.getLog([]byte{0xaa}, 42)
	require.Nil(t, err)
	require.Equal(t, []byte("foo"), logEntryFetched.Events[0].Identifier)
}

func TestLogsRepository_GetLogShouldNotFallbackToPreviousEpochIfZero(t *testing.T) {
	marshaller := &marshal.GogoProtoMarshalizer{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
					if epoch != 0 {
						require.Fail(t, "unexpected")
					}

					return nil, errors.New("expected")
				},
			}, nil
		},
	}

	repository := newLogsRepository(storageService, marshaller)
	_, err := repository.getLog([]byte{0xaa}, 0)
	require.Error(t, err, "expected")
}

func TestLogsRepository_GetLogsShouldFallbackToPreviousEpoch(t *testing.T) {
	storageService := genericMocks.NewChainStorerMock(uint32(0))
	marshaller := &marshal.GogoProtoMarshalizer{}
	repository := newLogsRepository(storageService, marshaller)

	fooBytes, _ := marshaller.Marshal(&transaction.Log{Events: []*transaction.Event{{Identifier: []byte("foo")}}})
	barBytes, _ := marshaller.Marshal(&transaction.Log{Events: []*transaction.Event{{Identifier: []byte("bar")}}})
	_ = storageService.Logs.PutInEpoch([]byte{0xaa}, fooBytes, 41)
	_ = storageService.Logs.PutInEpoch([]byte{0xbb}, barBytes, 41)

	// entries are missing in epoch 42 (edge-case), but are present in epoch 41
	logEntriesFetched, err := repository.getLogs([][]byte{{0xaa}, {0xbb}}, 42)
	require.Nil(t, err)
	require.Equal(t, []byte("foo"), logEntriesFetched[string([]byte{0xaa})].Events[0].Identifier)
	require.Equal(t, []byte("bar"), logEntriesFetched[string([]byte{0xbb})].Events[0].Identifier)
}

func TestLogsRepository_GetLogsShouldNotFallbackToPreviousEpochIfZero(t *testing.T) {
	marshaller := &marshal.GogoProtoMarshalizer{}
	storageService := &storageStubs.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &storageStubs.StorerStub{
				GetBulkFromEpochCalled: func(keys [][]byte, epoch uint32) ([]data.KeyValuePair, error) {
					if epoch != 0 {
						require.Fail(t, "unexpected")
					}

					return nil, errors.New("expected")
				},
			}, nil
		},
	}

	repository := newLogsRepository(storageService, marshaller)
	_, err := repository.getLogs([][]byte{{0xaa}, {0xbb}}, 0)
	require.Error(t, err, "expected")
}

func TestLogsRepository_GetLogsShouldErr(t *testing.T) {
	epoch := uint32(7)

	storageService := genericMocks.NewChainStorerMock(epoch)
	marshaller := &marshal.GogoProtoMarshalizer{}

	repository := newLogsRepository(storageService, marshaller)

	// Missing log
	missingLog, err := repository.getLog([]byte{0xcc}, epoch)
	require.ErrorIs(t, err, errCannotLoadLogs)
	require.Nil(t, missingLog)

	// Badly serialized log
	_ = storageService.Logs.Put([]byte{0xaa}, []byte{0xa, 0xb, 0xc, 0xd})
	badLog, err := repository.getLog([]byte{0xaa}, epoch)
	require.ErrorIs(t, err, errCannotUnmarshalLog)
	require.Nil(t, badLog)
}
