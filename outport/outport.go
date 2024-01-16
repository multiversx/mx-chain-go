package outport

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/storage/txcache"
	logger "github.com/multiversx/mx-chain-logger-go"
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
	messageCounter    uint64
	config            outportcore.OutportConfig
}

// NewOutport will create a new instance of proxy
func NewOutport(retrialInterval time.Duration, cfg outportcore.OutportConfig) (*outport, error) {
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
		config:            cfg,
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
		blockData, err := prepareBlockData(args.HeaderDataWithBody, driver)
		if err != nil {
			return err
		}

		args.OutportBlock.BlockData = blockData
		o.saveBlockBlocking(args.OutportBlock, driver)
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

func (o *outport) monitorCompletionOnDriver(function string, driver Driver) chan struct{} {
	counter := atomic.AddUint64(&o.messageCounter, 1)

	o.logHandler(logger.LogDebug, "outport.monitorCompletionOnDriver starting",
		"function", function, "driver", driverString(driver), "message counter", counter)
	ch := make(chan struct{})
	go func(startTime time.Time) {
		timer := time.NewTimer(o.timeForDriverCall)

		select {
		case <-ch:
			o.logHandler(logger.LogDebug, "outport.monitorCompletionOnDriver ended",
				"function", function, "driver", driverString(driver), "message counter", counter, "time", time.Since(startTime))
		case <-timer.C:
			o.logHandler(logger.LogWarning, "outport.monitorCompletionOnDriver took too long",
				"function", function, "driver", driverString(driver), "message counter", counter, "time", o.timeForDriverCall)
		}

		timer.Stop()
	}(time.Now())

	return ch
}

func (o *outport) saveBlockBlocking(args *outportcore.OutportBlock, driver Driver) {
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
func (o *outport) RevertIndexedBlock(headerDataWithBody *outportcore.HeaderDataWithBody) error {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		blockData, err := prepareBlockData(headerDataWithBody, driver)
		if err != nil {
			return err
		}

		o.revertIndexedBlockBlocking(blockData, driver)
	}

	return nil
}

func (o *outport) revertIndexedBlockBlocking(blockData *outportcore.BlockData, driver Driver) {
	ch := o.monitorCompletionOnDriver("revertIndexedBlockBlocking", driver)
	defer close(ch)

	for {
		err := driver.RevertIndexedBlock(blockData)
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
func (o *outport) SaveRoundsInfo(roundsInfo *outportcore.RoundsInfo) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		o.saveRoundsInfoBlocking(roundsInfo, driver)
	}
}

func (o *outport) saveRoundsInfoBlocking(roundsInfo *outportcore.RoundsInfo, driver Driver) {
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
func (o *outport) SaveValidatorsPubKeys(validatorsPubKeys *outportcore.ValidatorsPubKeys) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		o.saveValidatorsPubKeysBlocking(validatorsPubKeys, driver)
	}
}

func (o *outport) saveValidatorsPubKeysBlocking(validatorsPubKeys *outportcore.ValidatorsPubKeys, driver Driver) {
	ch := o.monitorCompletionOnDriver("saveValidatorsPubKeysBlocking", driver)
	defer close(ch)

	for {
		err := driver.SaveValidatorsPubKeys(validatorsPubKeys)
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
func (o *outport) SaveValidatorsRating(validatorsRating *outportcore.ValidatorsRating) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		o.saveValidatorsRatingBlocking(validatorsRating, driver)
	}
}

func (o *outport) saveValidatorsRatingBlocking(validatorsRating *outportcore.ValidatorsRating, driver Driver) {
	ch := o.monitorCompletionOnDriver("saveValidatorsRatingBlocking", driver)
	defer close(ch)

	for {
		err := driver.SaveValidatorsRating(validatorsRating)
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
func (o *outport) SaveAccounts(accounts *outportcore.Accounts) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		o.saveAccountsBlocking(accounts, driver)
	}
}

// NewTransactionHandlerInPool gets called whenever a new transaction (transaction, unsigned transaction or reward transaction)
// was added to the pool
func (o *outport) NewTransactionHandlerInPool(key []byte, value interface{}) {
	if check.IfNilReflect(value) {
		return
	}

	switch t := value.(type) {
	case *txcache.WrappedTransaction:
		tx, isTransaction := t.Tx.(*transaction.Transaction)
		if !isTransaction {
			log.Warn("programming error in NewTransactionHandlerInPool, improper value",
				"value type", fmt.Sprintf("%T", value),
				"value.Tx type", fmt.Sprintf("%T", t.Tx))
			return
		}

		// TODO something with the transaction and remove the following line
		_ = tx
	case *rewardTx.RewardTx:
		// TODO something with the reward transaction
		_ = t
	case *smartContractResult.SmartContractResult:
		// TODO something with the smart contract result transaction
		_ = t
	default:
		log.Warn("programming error in NewTransactionHandlerInPool, improper value",
			"value type", fmt.Sprintf("%T", value))
	}
}

func (o *outport) saveAccountsBlocking(accounts *outportcore.Accounts, driver Driver) {
	ch := o.monitorCompletionOnDriver("saveAccountsBlocking", driver)
	defer close(ch)

	for {
		err := driver.SaveAccounts(accounts)
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
func (o *outport) FinalizedBlock(finalizedBlock *outportcore.FinalizedBlock) {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	for _, driver := range o.drivers {
		o.finalizedBlockBlocking(finalizedBlock, driver)
	}
}

func (o *outport) finalizedBlockBlocking(finalizedBlock *outportcore.FinalizedBlock, driver Driver) {
	ch := o.monitorCompletionOnDriver("finalizedBlockBlocking", driver)
	defer close(ch)

	for {
		err := driver.FinalizedBlock(finalizedBlock)
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
