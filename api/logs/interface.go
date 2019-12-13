package logs

import "io"

type wsConn interface {
	io.Closer
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
}
