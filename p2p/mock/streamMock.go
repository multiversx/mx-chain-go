package mock

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/protocol"
)

type streamMock struct {
	mutData      sync.Mutex
	buffStream   *bytes.Buffer
	pid          protocol.ID
	streamClosed bool
	canRead      bool
	conn         network.Conn
}

// NewStreamMock -
func NewStreamMock() *streamMock {
	return &streamMock{
		mutData:      sync.Mutex{},
		buffStream:   new(bytes.Buffer),
		streamClosed: false,
		canRead:      false,
	}
}

// Read -
func (sm *streamMock) Read(p []byte) (n int, err error) {
	//just a mock implementation of blocking read
	for {
		time.Sleep(time.Millisecond * 10)

		sm.mutData.Lock()
		if sm.streamClosed {
			sm.mutData.Unlock()
			return 0, io.EOF
		}

		if sm.canRead {
			n, err := sm.buffStream.Read(p)
			sm.canRead = false
			sm.mutData.Unlock()

			return n, err
		}
		sm.mutData.Unlock()
	}
}

// Write -
func (sm *streamMock) Write(p []byte) (int, error) {
	sm.mutData.Lock()
	n, err := sm.buffStream.Write(p)
	if err == nil {
		sm.canRead = true
	}
	sm.mutData.Unlock()

	return n, err
}

// Close -
func (sm *streamMock) Close() error {
	sm.mutData.Lock()
	defer sm.mutData.Unlock()

	sm.streamClosed = true
	return nil
}

// Reset -
func (sm *streamMock) Reset() error {
	sm.mutData.Lock()
	defer sm.mutData.Unlock()

	sm.buffStream.Reset()
	sm.canRead = false
	return nil
}

// SetDeadline -
func (sm *streamMock) SetDeadline(time.Time) error {
	panic("implement me")
}

// SetReadDeadline -
func (sm *streamMock) SetReadDeadline(time.Time) error {
	panic("implement me")
}

// SetWriteDeadline -
func (sm *streamMock) SetWriteDeadline(time.Time) error {
	panic("implement me")
}

// Protocol -
func (sm *streamMock) Protocol() protocol.ID {
	return sm.pid
}

// SetProtocol -
func (sm *streamMock) SetProtocol(pid protocol.ID) {
	sm.pid = pid
}

// Stat -
func (sm *streamMock) Stat() network.Stat {
	return network.Stat{
		Direction: network.DirOutbound,
	}
}

// Conn -
func (sm *streamMock) Conn() network.Conn {
	return sm.conn
}

func (sm *streamMock) SetConn(conn network.Conn) {
	sm.conn = conn
}
