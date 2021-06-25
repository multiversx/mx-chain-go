package mock

import (
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

// HeartbeatComponentsStub -
type HeartbeatComponentsStub struct {
	HBMessenger heartbeat.MessageHandler
	HBMonitor   factory.HeartbeatMonitor
	HBSender    factory.HeartbeatSender
	HBStorer    factory.HeartbeatStorer
}

// Create -
func (hbs *HeartbeatComponentsStub) Create() error {
	return nil
}

// Close -
func (hbs *HeartbeatComponentsStub) Close() error {
	return nil
}

// CheckSubcomponents -
func (hbs *HeartbeatComponentsStub) CheckSubcomponents() error {
	return nil
}

// MessageHandler -
func (hbs *HeartbeatComponentsStub) MessageHandler() heartbeat.MessageHandler {
	return hbs.HBMessenger
}

// Monitor -
func (hbs *HeartbeatComponentsStub) Monitor() factory.HeartbeatMonitor {
	return hbs.HBMonitor
}

// Sender -
func (hbs *HeartbeatComponentsStub) Sender() factory.HeartbeatSender {
	return hbs.HBSender
}

// Storer -
func (hbs *HeartbeatComponentsStub) Storer() factory.HeartbeatStorer {
	return hbs.HBStorer
}

// IsInterfaceNil -
func (hbs *HeartbeatComponentsStub) IsInterfaceNil() bool {
	return hbs == nil
}
