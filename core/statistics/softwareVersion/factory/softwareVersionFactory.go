package factory

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics/softwareVersion"
)

const stableTagLocation = "https://api.github.com/repos/ElrondNetwork/elrond-go/releases/latest"

type softwareVersionFactory struct {
	statusHandler core.AppStatusHandler
}

// NewSoftwareVersionFactory is responsible for creating a new software version factory object
func NewSoftwareVersionFactory(statusHandler core.AppStatusHandler) (*softwareVersionFactory, error) {
	if check.IfNil(statusHandler) {
		return nil, core.ErrNilAppStatusHandler
	}

	softwareVersionFactoryObject := &softwareVersionFactory{
		statusHandler: statusHandler,
	}

	return softwareVersionFactoryObject, nil
}

// Create returns an software version checker object
func (svf *softwareVersionFactory) Create() (*softwareVersion.SoftwareVersionChecker, error) {
	stableTagProvider := softwareVersion.NewStableTagProvider(stableTagLocation)
	softwareVersionChecker, err := softwareVersion.NewSoftwareVersionChecker(svf.statusHandler, stableTagProvider)

	return softwareVersionChecker, err
}
