package factory

import (
	"github.com/ElrondNetwork/elrond-go/statusHandler/presenter"
	"github.com/ElrondNetwork/elrond-go/statusHandler/view"
)

type presenterFactory struct {
}

// NewPresenterFactory is responsible for creating a new presenter factory object
func NewPresenterFactory() *presenterFactory {
	presenterFactory := presenterFactory{}

	return &presenterFactory
}

// Create returns an presenter object that will hold presenter in the system
func (pf *presenterFactory) Create() view.Presenter {
	presenterStatusHandler := presenter.NewPresenterStatusHandler()

	return presenterStatusHandler
}
