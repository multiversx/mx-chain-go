package factory

import (
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/statusHandler/view"
	"github.com/ElrondNetwork/elrond-go/statusHandler/view/termuic"
)

type viewsFactory struct {
	presenter view.Presenter
}

// NewViewsFactory is responsible for creating a new viewers factory object
func NewViewsFactory(presenter view.Presenter) (*viewsFactory, error) {
	if presenter == nil || presenter.IsInterfaceNil() {
		return nil, statusHandler.ErrorNilPresenterInterface
	}

	return &viewsFactory{
		presenter,
	}, nil
}

// Create returns an view slice that will hold all views in the system
func (wf *viewsFactory) Create() ([]statusHandler.Viewer, error) {
	viewers := make([]statusHandler.Viewer, 0)

	termuiConsole, err := wf.createTermuiConsole()
	if err != nil {
		return nil, err
	}
	viewers = append(viewers, termuiConsole)
	return viewers, nil

}

func (wf *viewsFactory) createTermuiConsole() (*termuic.TermuiConsole, error) {
	termuiConsole, err := termuic.NewTermuiConsole(wf.presenter)
	if err != nil {
		return nil, err
	}

	return termuiConsole, nil
}
