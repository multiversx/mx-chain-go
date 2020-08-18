package mock

import (
	"github.com/ElrondNetwork/elrond-go/outport/messages"
)

// SenderMock -
type SenderMock struct {
	SentMessages []messages.MessageHandler
}

// NewSenderMock -
func NewSenderMock() *SenderMock {
	return &SenderMock{
		SentMessages: make([]messages.MessageHandler, 0),
	}
}

// Send -
func (m *SenderMock) Send(message messages.MessageHandler) (int, error) {
	m.SentMessages = append(m.SentMessages, message)
	return 42, nil
}

// GetLatestMessage -
func (m *SenderMock) GetLatestMessage() messages.MessageHandler {
	if len(m.SentMessages) == 0 {
		return nil
	}

	return m.SentMessages[len(m.SentMessages)-1]
}

// Shutdown -
func (m *SenderMock) Shutdown() error {
	return nil
}

// IsInterfaceNil -
func (m *SenderMock) IsInterfaceNil() bool {
	return m == nil
}
