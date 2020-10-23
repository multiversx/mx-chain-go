package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/statusHandler/view"
	"github.com/ElrondNetwork/elrond-go/statusHandler/view/termuic"
)

type viewsFactory struct {
	presenter                 view.Presenter
	refreshTimeInMilliseconds int
}

// NewViewsFactory is responsible for creating a new viewers factory object
func NewViewsFactory(presenter view.Presenter, refreshTimeInMilliseconds int) (*viewsFactory, error) {
	if check.IfNil(presenter) {
		return nil, statusHandler.ErrNilPresenterInterface
	}

	return &viewsFactory{
		presenter:                 presenter,
		refreshTimeInMilliseconds: refreshTimeInMilliseconds,
	}, nil
}

// Create returns an view slice that will hold all views in the system
func (wf *viewsFactory) Create() ([]Viewer, error) {
	views := make([]Viewer, 0)

	termuiConsole, err := wf.createTermuiConsole()
	if err != nil {
		return nil, err
	}
	views = append(views, termuiConsole)

	return views, nil
}

func (wf *viewsFactory) createTermuiConsole() (*termuic.TermuiConsole, error) {
	termuiConsole, err := termuic.NewTermuiConsole(wf.presenter, wf.refreshTimeInMilliseconds)
	if err != nil {
		return nil, err
	}

	return termuiConsole, nil
}
