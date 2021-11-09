package outport

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

type outport struct {
	mutex   sync.RWMutex
	drivers []Driver
}

var log = logger.GetOrCreate("outport")

// NewOutport will create a new instance of proxy
func NewOutport() *outport {
	return &outport{
		drivers: make([]Driver, 0),
		mutex:   sync.RWMutex{},
	}
}

// SaveBlock will save block for every driver
func (o *outport) SaveBlock(args *indexer.ArgsSaveBlockData) error {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		err := driver.SaveBlock(args)
		if err != nil {
			return err
		}
	}

	return nil
}

// RevertIndexedBlock will revert block for every driver
func (o *outport) RevertIndexedBlock(header data.HeaderHandler, body data.BodyHandler) error {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		err := driver.RevertIndexedBlock(header, body)
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveRoundsInfo will save rounds information for every driver
func (o *outport) SaveRoundsInfo(roundsInfos []*indexer.RoundInfo) error {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		err := driver.SaveRoundsInfo(roundsInfos)
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveValidatorsPubKeys will save validators public keys for every driver
func (o *outport) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) error {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		err := driver.SaveValidatorsPubKeys(validatorsPubKeys, epoch)
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveValidatorsRating will save validators rating for every driver
func (o *outport) SaveValidatorsRating(indexID string, infoRating []*indexer.ValidatorRatingInfo) error {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		err := driver.SaveValidatorsRating(indexID, infoRating)
		if err != nil {
			return err
		}
	}

	return nil
}

// SaveAccounts will save accounts  for every driver
func (o *outport) SaveAccounts(blockTimestamp uint64, acc []data.UserAccountHandler) error {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		err := driver.SaveAccounts(blockTimestamp, acc)
		if err != nil {
			return err
		}
	}

	return nil
}

// FinalizedBlock will call all the drivers that a block is finalized
func (o *outport) FinalizedBlock(headerHash []byte) error {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		err := driver.FinalizedBlock(headerHash)
		if err != nil {
			return err
		}
	}

	return nil
}

// Close will close all the drivers that are in outport
func (o *outport) Close() error {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	var err error
	for _, driver := range o.drivers {
		errClose := driver.Close()
		if errClose != nil {
			log.Error("cannot close driver", "error", errClose.Error())
			err = errClose
		}

	}

	return err
}

// HasDrivers returns true if there is at least one driver in the outport
func (o *outport) HasDrivers() bool {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	return len(o.drivers) != 0
}

// SubscribeDriver can subscribe a driver to the outport
func (o *outport) SubscribeDriver(driver Driver) error {
	if check.IfNil(driver) {
		return ErrNilDriver
	}

	o.mutex.Lock()
	o.drivers = append(o.drivers, driver)
	o.mutex.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (o *outport) IsInterfaceNil() bool {
	return o == nil
}
