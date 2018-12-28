package mock

import (
	"github.com/pkg/errors"
)

type Facade struct {
	Running          bool
	ShouldErrorStart bool
	ShouldErrorStop  bool
}

func (f *Facade) IsNodeRunning() bool {
	return f.Running
}

func (f *Facade) StartNode() error {
	if f.ShouldErrorStart {
		return errors.New("error")
	}
	return nil
}

func (f *Facade) StopNode() error {
	if f.ShouldErrorStop {
		return errors.New("error")
	}
	f.Running = false
	return nil
}

type WrongFacade struct {
}
