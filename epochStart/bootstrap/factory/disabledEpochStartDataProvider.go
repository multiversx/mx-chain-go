package factory

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/structs"
)

type disabledEpochStartDataProvider struct {
}

// Bootstrap will return an error indicating that the sync is not needed
func (d *disabledEpochStartDataProvider) Bootstrap() (*structs.ComponentsNeededForBootstrap, error) {
	return nil, errors.New("sync not needed")
}
