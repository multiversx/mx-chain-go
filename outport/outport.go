package outport

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/outport/drivers"
	"github.com/ElrondNetwork/elrond-go/outport/types"
)

type proxy struct {
	mutex   sync.RWMutex
	drivers []drivers.Driver
}

// NewOutport will create a new instance of proxy
func NewOutport() *proxy {
	return &proxy{
		drivers: make([]drivers.Driver, 0),
		mutex:   sync.RWMutex{},
	}
}

// SaveBlock --
func (o *proxy) SaveBlock(args types.ArgsSaveBlocks) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, indexer := range o.drivers {
		indexer.SaveBlock(args)
	}
}

// RevertBlock -
func (o *proxy) RevertBlock(header data.HeaderHandler, body data.BodyHandler) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, indexer := range o.drivers {
		indexer.RevertBlock(header, body)
	}
}

// SaveRoundsInfo -
func (o *proxy) SaveRoundsInfo(roundsInfos []types.RoundInfo) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, indexer := range o.drivers {
		indexer.SaveRoundsInfo(roundsInfos)
	}
}

// UpdateTPS -
func (o *proxy) UpdateTPS(tpsBenchmark statistics.TPSBenchmark) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	for _, indexer := range o.drivers {
		indexer.UpdateTPS(tpsBenchmark)
	}
}

// SaveValidatorsPubKeys -
func (o *proxy) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, indexer := range o.drivers {
		indexer.SaveValidatorsPubKeys(validatorsPubKeys, epoch)
	}
}

// SaveValidatorsRating -
func (o *proxy) SaveValidatorsRating(indexID string, infoRating []types.ValidatorRatingInfo) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, indexer := range o.drivers {
		indexer.SaveValidatorsRating(indexID, infoRating)
	}
}

// SaveAccounts -
func (o *proxy) SaveAccounts(acc []state.UserAccountHandler) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, indexer := range o.drivers {
		indexer.SaveAccounts(acc)
	}
}

// Close -
func (o *proxy) Close() error {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, indexer := range o.drivers {
		_ = indexer.Close()
	}

	return nil
}

// HasDrivers -
func (o *proxy) HasDrivers() bool {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return len(o.drivers) != 0
}

// SubscribeDriver -
func (o *proxy) SubscribeDriver(driver drivers.Driver) error {
	if check.IfNil(driver) {
		return ErrNilDriver
	}

	o.mutex.Lock()
	o.drivers = append(o.drivers, driver)
	o.mutex.Unlock()

	return nil
}

// IsInterfaceNil -
func (o *proxy) IsInterfaceNil() bool {
	return o == nil
}
