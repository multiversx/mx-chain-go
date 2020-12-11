package websockets

import (
	outportTypes "github.com/ElrondNetwork/elrond-go/outport/types"
	"github.com/gorilla/websocket"
)

// ClientHandler -
type ClientHandler interface {
	StartSendingBlocking(conn *websocket.Conn)
}

// DataPreparer  -
type DataPreparer interface {
	SaveBlock(args outportTypes.ArgsSaveBlocks) error
	SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) error
}
