package mock

import (
	"bytes"
	"io"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-protocol"
)

type streamMock struct {
	mutData      sync.Mutex
	buffStream   *bytes.Buffer
	pid          protocol.ID
	streamClosed bool
	canRead      bool
}

func NewStreamMock() *streamMock {
	return &streamMock{
		mutData:      sync.Mutex{},
		buffStream:   new(bytes.Buffer),
		streamClosed: false,
		canRead:      false,
	}
}

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

func (sm *streamMock) Write(p []byte) (int, error) {
	sm.mutData.Lock()
	n, err := sm.buffStream.Write(p)
	if err == nil {
		sm.canRead = true
	}
	sm.mutData.Unlock()

	return n, err
}

func (sm *streamMock) Close() error {
	sm.mutData.Lock()
	defer sm.mutData.Unlock()

	sm.streamClosed = true
	return nil
}

func (sm *streamMock) Reset() error {
	sm.mutData.Lock()
	defer sm.mutData.Unlock()

	sm.buffStream.Reset()
	sm.canRead = false
	return nil
}

func (sm *streamMock) SetDeadline(time.Time) error {
	panic("implement me")
}

func (sm *streamMock) SetReadDeadline(time.Time) error {
	panic("implement me")
}

func (sm *streamMock) SetWriteDeadline(time.Time) error {
	panic("implement me")
}

func (sm *streamMock) Protocol() protocol.ID {
	return sm.pid
}

func (sm *streamMock) SetProtocol(pid protocol.ID) {
	sm.pid = pid
}

func (sm *streamMock) Stat() net.Stat {
	return net.Stat{
		Direction: net.DirOutbound,
	}
}

func (sm *streamMock) Conn() net.Conn {
	panic("implement me")
}
