package host

import "github.com/multiversx/mx-chain-communication-go/websocket"

// SenderHost defines the actions that a host sender should do
type SenderHost interface {
	Send(payload []byte, topic string) error
	SetPayloadHandler(handler websocket.PayloadHandler) error
	Close() error
	IsInterfaceNil() bool
}

type payloadProcessorHandler interface {
	websocket.PayloadHandler
	SetHandlerFunc(handler func()) error
}
