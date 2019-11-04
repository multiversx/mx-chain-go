package factory

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/statistics/softwareVersion"
)

type softwareVersionFactory struct {
	statusHandler core.AppStatusHandler
}

// NewSoftwareVersionFactory is responsible for creating a new software version factory object
func NewSoftwareVersionFactory(statusHandler core.AppStatusHandler) (*softwareVersionFactory, error) {
	if statusHandler == nil || statusHandler.IsInterfaceNil() {
		return nil, core.ErrNilAppStatusHandler
	}

	softwareVersionFactory := &softwareVersionFactory{
		statusHandler: statusHandler,
	}

	return softwareVersionFactory, nil
}

// Create returns an software version checker object
func (svf *softwareVersionFactory) Create() (*softwareVersion.SoftwareVersionChecker, error) {
	softwareVersionChecker, err := softwareVersion.NewSoftwareVersionChecker(svf.statusHandler)

	return softwareVersionChecker, err
}
