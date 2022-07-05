package mock

import (
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

// HeartbeatComponentsStub -
type HeartbeatComponentsStub struct {
	MessageHandlerField heartbeat.MessageHandler
	MonitorField        factory.HeartbeatMonitor
	SenderField         factory.HeartbeatSender
	StorerField         factory.HeartbeatStorer
}

// Create -
func (hbc *HeartbeatComponentsStub) Create() error {
	return nil
}

// Close -
func (hbc *HeartbeatComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (hbc *HeartbeatComponentsStub) CheckSubcomponents() error {
	return nil
}

// String -
func (hbc *HeartbeatComponentsStub) String() string {
	return ""
}

// MessageHandler -
func (hbc *HeartbeatComponentsStub) MessageHandler() heartbeat.MessageHandler {
	return hbc.MessageHandlerField
}

// Monitor -
func (hbc *HeartbeatComponentsStub) Monitor() factory.HeartbeatMonitor {
	return hbc.MonitorField
}

// Sender -
func (hbc *HeartbeatComponentsStub) Sender() factory.HeartbeatSender {
	return hbc.SenderField
}

// Storer -
func (hbc *HeartbeatComponentsStub) Storer() factory.HeartbeatStorer {
	return hbc.StorerField
}

// IsInterfaceNil -
func (hbc *HeartbeatComponentsStub) IsInterfaceNil() bool {
	return hbc == nil
}
