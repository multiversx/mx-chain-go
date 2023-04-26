package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common/statistics/softwareVersion"
	"github.com/multiversx/mx-chain-go/config"
)

type softwareVersionFactory struct {
	statusHandler core.AppStatusHandler
	config        config.SoftwareVersionConfig
}

// NewSoftwareVersionFactory is responsible for creating a new software version factory object
func NewSoftwareVersionFactory(
	statusHandler core.AppStatusHandler,
	config config.SoftwareVersionConfig,
) (*softwareVersionFactory, error) {
	if check.IfNil(statusHandler) {
		return nil, core.ErrNilAppStatusHandler
	}

	softwareVersionFactoryObject := &softwareVersionFactory{
		statusHandler: statusHandler,
		config:        config,
	}

	return softwareVersionFactoryObject, nil
}

// Create returns a software version checker object
func (svf *softwareVersionFactory) Create() (*softwareVersion.SoftwareVersionChecker, error) {
	stableTagProvider := softwareVersion.NewStableTagProvider(svf.config.StableTagLocation)
	return softwareVersion.NewSoftwareVersionChecker(svf.statusHandler, stableTagProvider, svf.config.PollingIntervalInMinutes)
}
