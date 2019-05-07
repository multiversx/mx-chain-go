package mock

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go-sandbox/node/heartbeat"

	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
)

// Facade is the mock implementation of a node router handler
type Facade struct {
	Running                    bool
	ShouldErrorStart           bool
	ShouldErrorStop            bool
	GetCurrentPublicKeyHandler func() string
	TpsBenchmarkHandler        func() *statistics.TpsBenchmark
	GetHeartbeatsHandler       func() ([]heartbeat.PubKeyHeartbeat, error)
}

// IsNodeRunning is the mock implementation of a handler's IsNodeRunning method
func (f *Facade) IsNodeRunning() bool {
	return f.Running
}

// StartNode is the mock implementation of a handler's StartNode method
func (f *Facade) StartNode() error {
	if f.ShouldErrorStart {
		return errors.New("error")
	}
	return nil
}

// TpsBenchmark is the mock implementation for retreiving the TpsBenchmark
func (f *Facade) TpsBenchmark() *statistics.TpsBenchmark {
	if f.TpsBenchmarkHandler != nil {
		return f.TpsBenchmarkHandler()
	}
	return nil
}

// StopNode is the mock implementation of a handler's StopNode method
func (f *Facade) StopNode() error {
	if f.ShouldErrorStop {
		return errors.New("error")
	}
	f.Running = false
	return nil
}

// GetCurrentPublicKey is the mock implementation of a handler's StopNode method
func (f *Facade) GetCurrentPublicKey() string {
	return f.GetCurrentPublicKeyHandler()
}

func (f *Facade) GetHeartbeats() ([]heartbeat.PubKeyHeartbeat, error) {
	return f.GetHeartbeatsHandler()
}

// WrongFacade is a struct that can be used as a wrong implementation of the node router handler
type WrongFacade struct {
}
