package mock

import (
	"github.com/pkg/errors"
)

// Facade is the mock implementation of a node router handler
type Facade struct {
	Running                    bool
	ShouldErrorStart           bool
	ShouldErrorStop            bool
	GetCurrentPublicKeyHandler func() string
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

// WrongFacade is a struct that can be used as a wrong implementation of the node router handler
type WrongFacade struct {
}
