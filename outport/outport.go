package outport

import (
	"fmt"
	"sync"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("outport")

type outport struct {
	mutex          sync.RWMutex
	drivers        []Driver
	chanClose      chan struct{}
	messageCounter uint64
	config         outportcore.OutportConfig
}

// NewOutport will create a new instance of proxy
func NewOutport(retrialInterval time.Duration, cfg outportcore.OutportConfig) (*outport, error) {
	return &outport{
		drivers:   make([]Driver, 0),
		mutex:     sync.RWMutex{},
		chanClose: make(chan struct{}),
		config:    cfg,
	}, nil
}

// SaveBlock will save block for every driver
func (o *outport) SaveBlock(args *outportcore.OutportBlockWithHeaderAndBody) error {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	if args == nil {
		return fmt.Errorf("outport.SaveBlock error: %w", errNilSaveBlockArgs)
	}

	for _, driver := range o.drivers {
		go func(driver Driver) {
			blockData, err := prepareBlockData(args.HeaderDataWithBody, driver)
			if err != nil {
				log.Error("error preparing block data for SaveBlock", "error", err)
				return
			}

			args.OutportBlock.BlockData = blockData
			o.doSaveBlock(args.OutportBlock, driver)
		}(driver)
	}

	return nil
}

func prepareBlockData(
	headerBodyData *outportcore.HeaderDataWithBody,
	driver Driver,
) (*outportcore.BlockData, error) {
	if headerBodyData == nil {
		return nil, fmt.Errorf("outport.prepareBlockData error: %w", errNilHeaderAndBodyArgs)
	}

	marshaller := driver.GetMarshaller()
	headerBytes, headerType, err := outportcore.GetHeaderBytesAndType(marshaller, headerBodyData.Header)
	if err != nil {
		return nil, err
	}
	body, err := outportcore.GetBody(headerBodyData.Body)
	if err != nil {
		return nil, err
	}

	return &outportcore.BlockData{
		ShardID:              headerBodyData.Header.GetShardID(),
		HeaderBytes:          headerBytes,
		HeaderType:           string(headerType),
		HeaderHash:           headerBodyData.HeaderHash,
		Body:                 body,
		IntraShardMiniBlocks: headerBodyData.IntraShardMiniBlocks,
	}, nil
}

func (o *outport) doSaveBlock(args *outportcore.OutportBlock, driver Driver) {
	err := driver.SaveBlock(args)
	if err != nil {
		log.Error("error calling SaveBlock, will retry",
			"driver", driverString(driver),
			"error", err)
	}
}

// RevertIndexedBlock will revert block for every driver
func (o *outport) RevertIndexedBlock(headerDataWithBody *outportcore.HeaderDataWithBody) error {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		go func(driver Driver) {
			blockData, err := prepareBlockData(headerDataWithBody, driver)
			if err != nil {
				log.Error("error preparing block data for RevertIndexedBlock", "error", err)
				return
			}

			o.doRevertIndexedBlock(blockData, driver)
		}(driver)

	}

	return nil
}

func (o *outport) doRevertIndexedBlock(blockData *outportcore.BlockData, driver Driver) {
	err := driver.RevertIndexedBlock(blockData)
	if err != nil {
		log.Error("error calling RevertIndexedBlock, will retry",
			"driver", driverString(driver),
			"error", err)
	}
}

// SaveRoundsInfo will save rounds information for every driver
func (o *outport) SaveRoundsInfo(roundsInfo *outportcore.RoundsInfo) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		go o.doSaveRoundsInfo(roundsInfo, driver)
	}
}

func (o *outport) doSaveRoundsInfo(roundsInfo *outportcore.RoundsInfo, driver Driver) {
	err := driver.SaveRoundsInfo(roundsInfo)
	if err != nil {
		log.Error("error calling SaveRoundsInfo, will retry",
			"driver", driverString(driver),
			"error", err)
	}
}

// SaveValidatorsPubKeys will save validators public keys for every driver
func (o *outport) SaveValidatorsPubKeys(validatorsPubKeys *outportcore.ValidatorsPubKeys) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		go o.doSaveValidatorsPubKeys(validatorsPubKeys, driver)
	}
}

func (o *outport) doSaveValidatorsPubKeys(validatorsPubKeys *outportcore.ValidatorsPubKeys, driver Driver) {
	err := driver.SaveValidatorsPubKeys(validatorsPubKeys)
	if err != nil {
		log.Error("error calling SaveValidatorsPubKeys, will retry",
			"driver", driverString(driver),
			"error", err)
	}
}

// SaveValidatorsRating will save validators rating for every driver
func (o *outport) SaveValidatorsRating(validatorsRating *outportcore.ValidatorsRating) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		go o.doSaveValidatorsRating(validatorsRating, driver)
	}
}

func (o *outport) doSaveValidatorsRating(validatorsRating *outportcore.ValidatorsRating, driver Driver) {
	err := driver.SaveValidatorsRating(validatorsRating)
	if err != nil {
		log.Error("error calling SaveValidatorsRating, will retry",
			"driver", driverString(driver),
			"error", err)
	}
}

// SaveAccounts will save accounts  for every driver
func (o *outport) SaveAccounts(accounts *outportcore.Accounts) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		go o.saveAccounts(accounts, driver)
	}
}

func (o *outport) saveAccounts(accounts *outportcore.Accounts, driver Driver) {
	err := driver.SaveAccounts(accounts)
	if err != nil {
		log.Error("error calling SaveAccounts, will retry",
			"driver", driverString(driver),
			"error", err)
	}
}

// FinalizedBlock will call all the drivers that a block is finalized
func (o *outport) FinalizedBlock(finalizedBlock *outportcore.FinalizedBlock) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		go o.doFinalizedBlock(finalizedBlock, driver)
	}
}

func (o *outport) doFinalizedBlock(finalizedBlock *outportcore.FinalizedBlock, driver Driver) {
	err := driver.FinalizedBlock(finalizedBlock)
	if err != nil {
		log.Error("error calling FinalizedBlock, will retry",
			"driver", driverString(driver),
			"error", err)
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

	callback := func() error {
		return driver.SetCurrentSettings(o.config)
	}

	err := driver.RegisterHandler(callback, outportcore.TopicSettings)
	if err != nil {
		return err
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
