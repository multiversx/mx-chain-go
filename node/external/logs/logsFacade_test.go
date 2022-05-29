package logs

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	marshalizerFactory "github.com/ElrondNetwork/elrond-go-core/marshal/factory"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/stretchr/testify/require"
)

func TestNewLogsFacade(t *testing.T) {
	t.Run("NilStorageService", func(t *testing.T) {
		arguments := ArgsNewLogsFacade{
			StorageService:  nil,
			Marshalizer:     testscommon.MarshalizerMock{},
			PubKeyConverter: testscommon.NewPubkeyConverterMock(32),
		}

		facade, err := NewLogsFacade(arguments)
		require.Equal(t, core.ErrNilStore, err)
		require.True(t, check.IfNil(facade))
	})

	t.Run("NilMarshalizer", func(t *testing.T) {
		arguments := ArgsNewLogsFacade{
			StorageService:  genericMocks.NewChainStorerMock(7),
			Marshalizer:     nil,
			PubKeyConverter: testscommon.NewPubkeyConverterMock(32),
		}

		facade, err := NewLogsFacade(arguments)
		require.Equal(t, core.ErrNilMarshalizer, err)
		require.True(t, check.IfNil(facade))
	})

	t.Run("NilPubKeyConverter", func(t *testing.T) {
		arguments := ArgsNewLogsFacade{
			StorageService:  genericMocks.NewChainStorerMock(7),
			Marshalizer:     testscommon.MarshalizerMock{},
			PubKeyConverter: nil,
		}

		facade, err := NewLogsFacade(arguments)
		require.Equal(t, core.ErrNilPubkeyConverter, err)
		require.True(t, check.IfNil(facade))
	})
}

func TestLogsFacade_GetLogShouldWork(t *testing.T) {
	storageService := genericMocks.NewChainStorerMock(7)
	marshalizer, _ := marshalizerFactory.NewMarshalizer("gogo protobuf")

	arguments := ArgsNewLogsFacade{
		StorageService:  storageService,
		Marshalizer:     marshalizer,
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
	logBytes, err := marshalizer.Marshal(testLog)
	require.Nil(t, err)
	storageService.Logs.Put(logKey, logBytes)

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
	marshalizer, _ := marshalizerFactory.NewMarshalizer("gogo protobuf")

	arguments := ArgsNewLogsFacade{
		StorageService:  storageService,
		Marshalizer:     marshalizer,
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

	logOfFirstBytes, _ := marshalizer.Marshal(logOfFirst)
	logOfSecondBytes, _ := marshalizer.Marshal(logOfSecond)
	logOfFourthBytes, _ := marshalizer.Marshal(logOfFourth)
	storageService.Logs.Put([]byte{0xaa}, logOfFirstBytes)
	storageService.Logs.Put([]byte{0xbb}, logOfSecondBytes)
	storageService.Logs.Put([]byte{0xdd}, logOfFourthBytes)

	err := facade.IncludeLogsInTransactions(transactions, [][]byte{{0xaa}, {0xbb}, {0xcc}, {0xdd}}, 7)
	require.Nil(t, err)
	require.Equal(t, "first", transactions[0].Logs.Events[0].Identifier)
	require.Equal(t, "second", transactions[1].Logs.Events[0].Identifier)
	require.Nil(t, transactions[2].Logs)
	require.Equal(t, "fourth", transactions[3].Logs.Events[0].Identifier)
}
