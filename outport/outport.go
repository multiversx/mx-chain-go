package outport

import (
	"fmt"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("outport")

const maxTimeForDriverCall = time.Second * 30
const minimumRetrialInterval = time.Millisecond * 10

type outport struct {
	mutex             sync.RWMutex
	drivers           []Driver
	retrialInterval   time.Duration
	chanClose         chan struct{}
	logHandler        func(logLevel logger.LogLevel, message string, args ...interface{})
	timeForDriverCall time.Duration
}

// NewOutport will create a new instance of proxy
func NewOutport(retrialInterval time.Duration) (*outport, error) {
	if retrialInterval < minimumRetrialInterval {
		return nil, fmt.Errorf("%w, provided: %d, minimum: %d", ErrInvalidRetrialInterval, retrialInterval, minimumRetrialInterval)
	}

	return &outport{
		drivers:           make([]Driver, 0),
		mutex:             sync.RWMutex{},
		retrialInterval:   retrialInterval,
		chanClose:         make(chan struct{}),
		logHandler:        log.Log,
		timeForDriverCall: maxTimeForDriverCall,
	}, nil
}

// SaveBlock will save block for every driver
func (o *outport) SaveBlock(args *indexer.ArgsSaveBlockData) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		o.saveBlockBlocking(args, driver)
	}
}

func (o *outport) monitorCompletionOnDriver(function string, driver Driver) chan struct{} {
	o.logHandler(logger.LogDebug, "outport.monitorCompletionOnDriver starting",
		"function", function, "driver", driverString(driver))
	ch := make(chan struct{})
	go func() {
		timer := time.NewTimer(o.timeForDriverCall)

		select {
		case <-ch:
			o.logHandler(logger.LogDebug, "outport.monitorCompletionOnDriver ended",
				"function", function, "driver", driverString(driver))
		case <-timer.C:
			o.logHandler(logger.LogError, "outport.monitorCompletionOnDriver took too long",
				"function", function, "driver", driverString(driver), "time", o.timeForDriverCall)
		}

		timer.Stop()
	}()

	return ch
}

func (o *outport) saveBlockBlocking(args *indexer.ArgsSaveBlockData, driver Driver) {
	ch := o.monitorCompletionOnDriver("saveBlockBlocking", driver)
	defer close(ch)

	for {
		err := driver.SaveBlock(args)
		if err == nil {
			return
		}

		log.Error("error calling SaveBlock, will retry",
			"driver", driverString(driver),
			"retrial in", o.retrialInterval,
			"error", err)

		if o.shouldTerminate() {
			return
		}
	}
}

func (o *outport) shouldTerminate() bool {
	select {
	case <-o.chanClose:
		return true
	case <-time.After(o.retrialInterval):
		return false
	}
}

// RevertIndexedBlock will revert block for every driver
func (o *outport) RevertIndexedBlock(header data.HeaderHandler, body data.BodyHandler) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		o.revertIndexedBlockBlocking(header, body, driver)
	}
}

func (o *outport) revertIndexedBlockBlocking(header data.HeaderHandler, body data.BodyHandler, driver Driver) {
	ch := o.monitorCompletionOnDriver("revertIndexedBlockBlocking", driver)
	defer close(ch)

	for {
		err := driver.RevertIndexedBlock(header, body)
		if err == nil {
			return
		}

		log.Error("error calling RevertIndexedBlock, will retry",
			"driver", driverString(driver),
			"retrial in", o.retrialInterval,
			"error", err)

		if o.shouldTerminate() {
			return
		}
	}
}

// SaveRoundsInfo will save rounds information for every driver
func (o *outport) SaveRoundsInfo(roundsInfo []*indexer.RoundInfo) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		o.saveRoundsInfoBlocking(roundsInfo, driver)
	}
}

func (o *outport) saveRoundsInfoBlocking(roundsInfo []*indexer.RoundInfo, driver Driver) {
	ch := o.monitorCompletionOnDriver("saveRoundsInfoBlocking", driver)
	defer close(ch)

	for {
		err := driver.SaveRoundsInfo(roundsInfo)
		if err == nil {
			return
		}

		log.Error("error calling SaveRoundsInfo, will retry",
			"driver", driverString(driver),
			"retrial in", o.retrialInterval,
			"error", err)

		if o.shouldTerminate() {
			return
		}
	}
}

// SaveValidatorsPubKeys will save validators public keys for every driver
func (o *outport) SaveValidatorsPubKeys(validatorsPubKeys map[uint32][][]byte, epoch uint32) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		o.saveValidatorsPubKeysBlocking(validatorsPubKeys, epoch, driver)
	}
}

func (o *outport) saveValidatorsPubKeysBlocking(validatorsPubKeys map[uint32][][]byte, epoch uint32, driver Driver) {
	ch := o.monitorCompletionOnDriver("saveValidatorsPubKeysBlocking", driver)
	defer close(ch)

	for {
		err := driver.SaveValidatorsPubKeys(validatorsPubKeys, epoch)
		if err == nil {
			return
		}

		log.Error("error calling SaveValidatorsPubKeys, will retry",
			"driver", driverString(driver),
			"retrial in", o.retrialInterval,
			"error", err)

		if o.shouldTerminate() {
			return
		}
	}
}

// SaveValidatorsRating will save validators rating for every driver
func (o *outport) SaveValidatorsRating(indexID string, infoRating []*indexer.ValidatorRatingInfo) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		o.saveValidatorsRatingBlocking(indexID, infoRating, driver)
	}
}

func (o *outport) saveValidatorsRatingBlocking(indexID string, infoRating []*indexer.ValidatorRatingInfo, driver Driver) {
	ch := o.monitorCompletionOnDriver("saveValidatorsRatingBlocking", driver)
	defer close(ch)

	for {
		err := driver.SaveValidatorsRating(indexID, infoRating)
		if err == nil {
			return
		}

		log.Error("error calling SaveValidatorsRating, will retry",
			"driver", driverString(driver),
			"retrial in", o.retrialInterval,
			"error", err)

		if o.shouldTerminate() {
			return
		}
	}
}

// SaveAccounts will save accounts  for every driver
func (o *outport) SaveAccounts(blockTimestamp uint64, acc []data.UserAccountHandler) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		o.saveAccountsBlocking(blockTimestamp, acc, driver)
	}
}

func (o *outport) saveAccountsBlocking(blockTimestamp uint64, acc []data.UserAccountHandler, driver Driver) {
	ch := o.monitorCompletionOnDriver("saveAccountsBlocking", driver)
	defer close(ch)

	for {
		err := driver.SaveAccounts(blockTimestamp, acc)
		if err == nil {
			return
		}

		log.Error("error calling SaveAccounts, will retry",
			"driver", driverString(driver),
			"retrial in", o.retrialInterval,
			"error", err)

		if o.shouldTerminate() {
			return
		}
	}
}

// FinalizedBlock will call all the drivers that a block is finalized
func (o *outport) FinalizedBlock(headerHash []byte) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		o.finalizedBlockBlocking(headerHash, driver)
	}
}

func (o *outport) finalizedBlockBlocking(headerHash []byte, driver Driver) {
	ch := o.monitorCompletionOnDriver("finalizedBlockBlocking", driver)
	defer close(ch)

	for {
		err := driver.FinalizedBlock(headerHash)
		if err == nil {
			return
		}

		log.Error("error calling FinalizedBlock, will retry",
			"driver", driverString(driver),
			"retrial in", o.retrialInterval,
			"error", err)

		if o.shouldTerminate() {
			return
		}
	}
}

// Close will close all the drivers that are in outport
func (o *outport) Close() error {
	close(o.chanClose)

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

	log.Debug("outport.SubscribeDriver new driver added", "driver", driverString(driver))

	return nil
}

func driverString(driver Driver) string {
	return fmt.Sprintf("%T", driver)
}

// IsInterfaceNil returns true if there is no value under the interface
func (o *outport) IsInterfaceNil() bool {
	return o == nil
}
