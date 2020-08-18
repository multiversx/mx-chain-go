package outport

import (
	"encoding/binary"
	"os"

	"github.com/ElrondNetwork/arwen-wasm-vm/ipc/marshaling"
	"github.com/ElrondNetwork/elrond-go/outport/messages"
)

// Sender intermediates communication (message sending) via pipes
type Sender struct {
	writer      *os.File
	marshalizer marshaling.Marshalizer
}

// NewSender creates a new sender
func NewSender(writer *os.File, marshalizer marshaling.Marshalizer) *Sender {
	return &Sender{
		writer:      writer,
		marshalizer: marshalizer,
	}
}

// Send sends a message over the pipe
func (sender *Sender) Send(message messages.MessageHandler) (int, error) {
	dataBytes, err := sender.marshalizer.Marshal(message)
	if err != nil {
		return 0, err
	}

	length := len(dataBytes)
	err = sender.sendMessageLengthAndKind(length, message.GetKind())
	if err != nil {
		return 0, err
	}

	_, err = sender.writer.Write(dataBytes)
	if err != nil {
		return 0, err
	}

	return length, err
}

func (sender *Sender) sendMessageLengthAndKind(length int, kind messages.MessageKind) error {
	buffer := make([]byte, 8)
	binary.LittleEndian.PutUint32(buffer[0:4], uint32(length))
	binary.LittleEndian.PutUint32(buffer[4:8], uint32(kind))
	_, err := sender.writer.Write(buffer)
	return err
}

// Shutdown closes the pipe
func (sender *Sender) Shutdown() error {
	err := sender.writer.Close()
	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (sender *Sender) IsInterfaceNil() bool {
	return sender == nil
}
