package logs_test

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/multiversx/mx-chain-go/api/logs"
	"github.com/multiversx/mx-chain-go/api/mock"
	"github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
)

func removeWriterFromLogSubsystem(w io.Writer) {
	_ = logger.RemoveLogObserver(w)
}

func createMockLogSender() (*logs.LogSender, *mock.WsConnStub, io.Writer) {
	conn := &mock.WsConnStub{}
	conn.SetCloseHandler(func() error {
		return nil
	})
	conn.SetReadMessageHandler(func() (messageType int, p []byte, err error) {
		profile := logger.Profile{LogLevelPatterns: "*:INFO"}
		profileJson, _ := profile.Marshal()
		return websocket.TextMessage, profileJson, nil
	})

	ls, _ := logs.NewLogSender(
		&mock.MarshalizerStub{},
		conn,
		&mock.LoggerStub{},
	)
	removeWriterFromLogSubsystem(ls.Writer())
	ls.SetWriter(logs.NewLogWriter())

	lsender := &logs.LogSender{}
	lsender.Set(ls)
	return lsender, conn, ls.Writer()
}

// ------- NewLogSender

func TestNewLogSender_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	ls, err := logs.NewLogSender(nil, &mock.WsConnStub{}, &mock.LoggerStub{})

	assert.Nil(t, ls)
	assert.Equal(t, logs.ErrNilMarshalizer, err)
}

func TestNewLogSender_NilConnectionShouldErr(t *testing.T) {
	t.Parallel()

	ls, err := logs.NewLogSender(&mock.MarshalizerStub{}, nil, &mock.LoggerStub{})

	assert.Nil(t, ls)
	assert.Equal(t, logs.ErrNilWsConn, err)
}

func TestNewLogSender_NilLoggerShouldErr(t *testing.T) {
	t.Parallel()

	ls, err := logs.NewLogSender(&mock.MarshalizerStub{}, &mock.WsConnStub{}, nil)

	assert.Nil(t, ls)
	assert.Equal(t, logs.ErrNilLogger, err)
}

func TestNewLogSender_ShouldWork(t *testing.T) {
	t.Parallel()

	ls, err := logs.NewLogSender(&mock.MarshalizerStub{}, &mock.WsConnStub{}, &mock.LoggerStub{})

	assert.NotNil(t, ls)
	assert.Nil(t, err)
	assert.NotNil(t, ls.Writer())

	removeWriterFromLogSubsystem(ls.Writer())
}

// ------- StartSendingBlocking

func TestLogSender_StartSendingBlockingConnReadMessageErrShouldCloseConn(t *testing.T) {
	t.Parallel()

	closeCalled := false
	conn := &mock.WsConnStub{}
	conn.SetCloseHandler(func() error {
		closeCalled = true
		return nil
	})
	conn.SetReadMessageHandler(func() (messageType int, p []byte, err error) {
		return websocket.TextMessage, nil, errors.New("")
	})
	ls, _ := logs.NewLogSender(
		&mock.MarshalizerStub{},
		conn,
		&mock.LoggerStub{},
	)
	removeWriterFromLogSubsystem(ls.Writer())

	ls.StartSendingBlocking()

	assert.True(t, closeCalled)
}

func TestLogSender_StartSendingBlockingWrongPatternShouldCloseConn(t *testing.T) {
	t.Parallel()

	closeCalled := false
	conn := &mock.WsConnStub{}
	conn.SetCloseHandler(func() error {
		closeCalled = true
		return nil
	})
	conn.SetReadMessageHandler(func() (messageType int, p []byte, err error) {
		return websocket.TextMessage, []byte("wrong log pattern"), nil
	})
	ls, _ := logs.NewLogSender(
		&mock.MarshalizerStub{},
		conn,
		&mock.LoggerStub{},
	)
	removeWriterFromLogSubsystem(ls.Writer())

	ls.StartSendingBlocking()

	assert.True(t, closeCalled)
}

func TestLogSender_StartSendingBlockingSendsMessage(t *testing.T) {
	t.Parallel()

	ls, conn, writer := createMockLogSender()
	data := []byte("random data")
	var retrievedData []byte
	conn.SetWriteMessageHandler(func(messageType int, data []byte) error {
		retrievedData = data
		return nil
	})

	go func() {
		// watchdog function
		time.Sleep(time.Millisecond * 10)

		_ = ls.Writer().Close()
	}()

	_, err := writer.Write(data)
	ls.StartSendingBlocking()

	assert.Nil(t, err)
	assert.Equal(t, data, retrievedData)
}

func TestLogSender_StartSendingBlockingSendsMessageAndStopsWhenReadClose(t *testing.T) {
	t.Parallel()

	ls, conn, writer := createMockLogSender()
	data := []byte("random data")
	var retrievedData []byte
	conn.SetWriteMessageHandler(func(messageType int, data []byte) error {
		retrievedData = data
		return nil
	})

	go func() {
		// watchdog function
		time.Sleep(time.Millisecond * 10)

		conn.SetReadMessageHandler(func() (messageType int, p []byte, err error) {
			return websocket.CloseMessage, []byte(""), nil
		})
	}()

	_, err := writer.Write(data)
	ls.StartSendingBlocking()

	assert.Nil(t, err)
	assert.Equal(t, data, retrievedData)
}

func TestLogSender_StartSendingBlockingConnWriteFailsShouldStop(t *testing.T) {
	t.Parallel()

	ls, conn, writer := createMockLogSender()
	data := []byte("random data")
	closeCalled := false
	conn.SetWriteMessageHandler(func(messageType int, data []byte) error {
		return errors.New("")
	})
	conn.SetCloseHandler(func() error {
		closeCalled = true
		return nil
	})

	_, _ = writer.Write(data)
	ls.StartSendingBlocking()

	assert.True(t, closeCalled)
}
