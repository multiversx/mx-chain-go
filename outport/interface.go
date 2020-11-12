package outport

import (
	"github.com/ElrondNetwork/elrond-go/outport/drivers"
)

// OutportHandler is interface that defines what a proxy implementation should be able to do
type OutportHandler interface {
	drivers.Driver
	SubscribeDriver(driver drivers.Driver) error
	HasDrivers() bool
}
