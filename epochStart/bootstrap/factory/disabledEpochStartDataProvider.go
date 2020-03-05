package factory

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
)

type disabledEpochStartDataProvider struct {
}

// Bootstrap will return an error indicating that the sync is not needed
func (d *disabledEpochStartDataProvider) Bootstrap() (*bootstrap.ComponentsNeededForBootstrap, error) {
	return nil, errors.New("sync not needed")
}
