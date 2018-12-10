package mock

import (
	"github.com/pkg/errors"
)

type Facade struct {
	Running bool
	ShouldErrorStart bool
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

type WrongFacade struct {

}