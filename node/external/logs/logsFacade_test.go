package logs

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/require"
)

func TestNewLogsFacade(t *testing.T) {
	t.Run("NilStorageService", func(t *testing.T) {
		arguments := ArgsNewLogsFacade{
			StorageService:  nil,
			Marshaller:      marshallerMock.MarshalizerMock{},
			PubKeyConverter: testscommon.NewPubkeyConverterMock(32),
		}

		facade, err := NewLogsFacade(arguments)
		require.ErrorIs(t, err, errCannotCreateLogsFacade)
		require.ErrorContains(t, err, core.ErrNilStore.Error())
		require.Nil(t, facade)
	})

	t.Run("NilMarshaller", func(t *testing.T) {
		arguments := ArgsNewLogsFacade{
			StorageService:  genericMocks.NewChainStorerMock(7),
			Marshaller:      nil,
			PubKeyConverter: testscommon.NewPubkeyConverterMock(32),
		}

		facade, err := NewLogsFacade(arguments)
		require.ErrorIs(t, err, errCannotCreateLogsFacade)
		require.ErrorContains(t, err, core.ErrNilMarshalizer.Error())
		require.Nil(t, facade)
	})

	t.Run("NilPubKeyConverter", func(t *testing.T) {
		arguments := ArgsNewLogsFacade{
			StorageService:  genericMocks.NewChainStorerMock(7),
			Marshaller:      marshallerMock.MarshalizerMock{},
			PubKeyConverter: nil,
		}

		facade, err := NewLogsFacade(arguments)
		require.ErrorIs(t, err, errCannotCreateLogsFacade)
		require.ErrorContains(t, err, core.ErrNilPubkeyConverter.Error())
		require.Nil(t, facade)
	})
}

func TestLogsFacade_GetLogShouldWork(t *testing.T) {
	storageService := genericMocks.NewChainStorerMock(7)
	marshaller := &marshal.GogoProtoMarshalizer{}

	arguments := ArgsNewLogsFacade{
		StorageService:  storageService,
		Marshaller:      marshaller,
		PubKeyConverter: testscommon.NewPubkeyConverterMock(32),
	}

	testLog := &transaction.Log{
		Address: []byte{0xab, 0xba},
		Events: []*transaction.Event{
			{
				Address:    []byte{0xaa, 0xbb},
				Identifier: []byte("helloWorld"),
				Topics:     [][]byte{[]byte("hello"), []byte("world")},
				Data:       []byte("Hello World!"),
			},
		},
	}

	logKey := []byte("hello")
	logBytes, err := marshaller.Marshal(testLog)
	require.Nil(t, err)
	_ = storageService.Logs.Put(logKey, logBytes)

	facade, _ := NewLogsFacade(arguments)
	logOnApi, err := facade.GetLog(logKey, 7)
	require.Nil(t, err)
	require.NotNil(t, logOnApi)
	require.Equal(t, "abba", logOnApi.Address)
	require.Equal(t, "aabb", logOnApi.Events[0].Address)
	require.Equal(t, "helloWorld", logOnApi.Events[0].Identifier)
	require.Equal(t, [][]byte{[]byte("hello"), []byte("world")}, logOnApi.Events[0].Topics)
	require.Equal(t, []byte("Hello World!"), logOnApi.Events[0].Data)
}

func TestLogsFacade_IncludeLogsInTransactionsShouldWork(t *testing.T) {
	storageService := genericMocks.NewChainStorerMock(7)
	marshaller := &marshal.GogoProtoMarshalizer{}

	arguments := ArgsNewLogsFacade{
		StorageService:  storageService,
		Marshaller:      marshaller,
		PubKeyConverter: testscommon.NewPubkeyConverterMock(32),
	}

	facade, _ := NewLogsFacade(arguments)

	transactions := []*transaction.ApiTransactionResult{
		{HashBytes: []byte{0xaa}},
		{HashBytes: []byte{0xbb}},
		{HashBytes: []byte{0xcc}},
		{HashBytes: []byte{0xdd}},
	}

	logOfFirst := &transaction.Log{
		Events: []*transaction.Event{
			{Identifier: []byte("first")},
		},
	}

	logOfSecond := &transaction.Log{
		Events: []*transaction.Event{
			{Identifier: []byte("second")},
		},
	}

	// (no log for the third transaction)

	logOfFourth := &transaction.Log{
		Events: []*transaction.Event{
			{Identifier: []byte("fourth")},
		},
	}

	logOfFirstBytes, _ := marshaller.Marshal(logOfFirst)
	logOfSecondBytes, _ := marshaller.Marshal(logOfSecond)
	logOfFourthBytes, _ := marshaller.Marshal(logOfFourth)
	_ = storageService.Logs.Put([]byte{0xaa}, logOfFirstBytes)
	_ = storageService.Logs.Put([]byte{0xbb}, logOfSecondBytes)
	_ = storageService.Logs.Put([]byte{0xdd}, logOfFourthBytes)

	err := facade.IncludeLogsInTransactions(transactions, [][]byte{{0xaa}, {0xbb}, {0xcc}, {0xdd}}, 7)
	require.Nil(t, err)
	require.Equal(t, "first", transactions[0].Logs.Events[0].Identifier)
	require.Equal(t, "second", transactions[1].Logs.Events[0].Identifier)
	require.Nil(t, transactions[2].Logs)
	require.Equal(t, "fourth", transactions[3].Logs.Events[0].Identifier)
}

func TestLogsFacade_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var lf *logsFacade
	require.True(t, lf.IsInterfaceNil())

	arguments := ArgsNewLogsFacade{
		StorageService:  genericMocks.NewChainStorerMock(7),
		Marshaller:      &marshal.GogoProtoMarshalizer{},
		PubKeyConverter: testscommon.NewPubkeyConverterMock(32),
	}
	lf, _ = NewLogsFacade(arguments)
	require.False(t, lf.IsInterfaceNil())
}
