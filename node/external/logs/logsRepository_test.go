package logs

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	"github.com/stretchr/testify/require"
)

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
