package preprocess

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/scheduled"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type scrInfo struct {
	txHash    []byte
	txHandler data.TransactionHandler
}

type scheduledTxsExecution struct {
	txProcessor       process.TransactionProcessor
	txCoordinator     process.TransactionCoordinator
	mapScheduledTxs   map[string]data.TransactionHandler
	mapScheduledSCRs  map[block.Type][]data.TransactionHandler
	scheduledTxs      []data.TransactionHandler
	scheduledRootHash []byte
	gasAndFees        scheduled.GasAndFees
	storer            storage.Storer
	marshaller        marshal.Marshalizer
	mutScheduledTxs   sync.RWMutex
	shardCoordinator  sharding.Coordinator
}

// NewScheduledTxsExecution creates a new object which handles the execution of scheduled transactions
func NewScheduledTxsExecution(
	txProcessor process.TransactionProcessor,
	txCoordinator process.TransactionCoordinator,
	storer storage.Storer,
	marshaller marshal.Marshalizer,
	shardCoordinator sharding.Coordinator,
) (*scheduledTxsExecution, error) {

	if check.IfNil(txProcessor) {
		return nil, process.ErrNilTxProcessor
	}
	if check.IfNil(txCoordinator) {
		return nil, process.ErrNilTransactionCoordinator
	}
	if check.IfNil(storer) {
		return nil, process.ErrNilStorage
	}
	if check.IfNil(marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	ste := &scheduledTxsExecution{
		txProcessor:       txProcessor,
		txCoordinator:     txCoordinator,
		mapScheduledTxs:   make(map[string]data.TransactionHandler),
		mapScheduledSCRs:  make(map[block.Type][]data.TransactionHandler),
		scheduledTxs:      make([]data.TransactionHandler, 0),
		gasAndFees:        process.GetZeroGasAndFees(),
		storer:            storer,
		marshaller:        marshaller,
		scheduledRootHash: nil,
		shardCoordinator:  shardCoordinator,
	}

	return ste, nil
}

// Init method removes all the scheduled transactions
func (ste *scheduledTxsExecution) Init() {
	ste.mutScheduledTxs.Lock()
	log.Debug("scheduledTxsExecution.Init", "num of last scheduled txs", len(ste.scheduledTxs))
	ste.mapScheduledTxs = make(map[string]data.TransactionHandler)
	ste.scheduledTxs = make([]data.TransactionHandler, 0)
	ste.mutScheduledTxs.Unlock()
}

// Add method adds a scheduled transaction to be executed
func (ste *scheduledTxsExecution) Add(txHash []byte, tx data.TransactionHandler) bool {
	ste.mutScheduledTxs.Lock()
	defer ste.mutScheduledTxs.Unlock()

	_, exist := ste.mapScheduledTxs[string(txHash)]
	if exist {
		return false
	}

	ste.mapScheduledTxs[string(txHash)] = tx
	ste.scheduledTxs = append(ste.scheduledTxs, tx)

	return true
}

// Execute method executes the given scheduled transaction
func (ste *scheduledTxsExecution) Execute(txHash []byte) error {
	ste.mutScheduledTxs.RLock()
	defer ste.mutScheduledTxs.RUnlock()

	txHandler, exist := ste.mapScheduledTxs[string(txHash)]
	if !exist {
		return fmt.Errorf("%w: in scheduledTxsExecution.Execute", process.ErrMissingTransaction)
	}

	err := ste.execute(txHandler)
	if err != nil && !errors.Is(err, process.ErrFailedTransaction) {
		return err
	}

	return nil
}

// ExecuteAll method executes all the scheduled transactions
func (ste *scheduledTxsExecution) ExecuteAll(haveTime func() time.Duration) error {
	ste.mutScheduledTxs.Lock()
	defer ste.mutScheduledTxs.Unlock()

	if haveTime == nil {
		return process.ErrNilHaveTimeHandler
	}

	log.Debug("scheduledTxsExecution.ExecuteAll", "num of scheduled txs to be executed", len(ste.scheduledTxs))

	mapAllIntermediateTxsBeforeScheduledExecution := ste.txCoordinator.GetAllIntermediateTxs()

	for _, txHandler := range ste.scheduledTxs {
		if haveTime() <= 0 {
			return process.ErrTimeIsOut
		}

		err := ste.execute(txHandler)
		if err != nil {
			log.Debug("scheduledTxsExecution.ExecuteAll: execute(txHandler)",
				"nonce", txHandler.GetNonce(),
				"value", txHandler.GetValue(),
				"gas limit", txHandler.GetGasLimit(),
				"gas price", txHandler.GetGasPrice(),
				"sender address", string(txHandler.GetSndAddr()),
				"receiver address", string(txHandler.GetRcvAddr()),
				"data", string(txHandler.GetData()),
				"error", err.Error())
			if !errors.Is(err, process.ErrFailedTransaction) {
				return err
			}
		}
	}

	mapAllIntermediateTxsAfterScheduledExecution := ste.txCoordinator.GetAllIntermediateTxs()
	ste.computeScheduledSCRs(mapAllIntermediateTxsBeforeScheduledExecution, mapAllIntermediateTxsAfterScheduledExecution)

	return nil
}

func (ste *scheduledTxsExecution) execute(txHandler data.TransactionHandler) error {
	tx, ok := txHandler.(*transaction.Transaction)
	if !ok {
		return fmt.Errorf("%w: in scheduledTxsExecution.execute", process.ErrWrongTypeAssertion)
	}

	_, err := ste.txProcessor.ProcessTransaction(tx)
	return err
}

func (ste *scheduledTxsExecution) computeScheduledSCRs(
	mapAllIntermediateTxsBeforeScheduledExecution map[block.Type]map[string]data.TransactionHandler,
	mapAllIntermediateTxsAfterScheduledExecution map[block.Type]map[string]data.TransactionHandler,
) {
	numScheduledSCRs := 0
	ste.mapScheduledSCRs = make(map[block.Type][]data.TransactionHandler)
	for blockType, allIntermediateTxsAfterScheduledExecution := range mapAllIntermediateTxsAfterScheduledExecution {
		scrsInfo := ste.getAllIntermediateTxsAfterScheduledExecution(
			mapAllIntermediateTxsBeforeScheduledExecution,
			allIntermediateTxsAfterScheduledExecution,
			blockType,
		)
		if len(scrsInfo) == 0 {
			continue
		}

		sort.Slice(scrsInfo, func(a, b int) bool {
			return bytes.Compare(scrsInfo[a].txHash, scrsInfo[b].txHash) < 0
		})

		ste.mapScheduledSCRs[blockType] = make([]data.TransactionHandler, len(scrsInfo))
		for scrIndex, scrInfo := range scrsInfo {
			ste.mapScheduledSCRs[blockType][scrIndex] = scrInfo.txHandler
			log.Trace("scheduledTxsExecution.computeScheduledSCRs", "blockType", blockType, "sender", ste.mapScheduledSCRs[blockType][scrIndex].GetSndAddr(), "receiver", ste.mapScheduledSCRs[blockType][scrIndex].GetRcvAddr())
		}

		numScheduledSCRs += len(scrsInfo)
	}

	log.Debug("scheduledTxsExecution.computeScheduledSCRs", "num of scheduled scrs created", numScheduledSCRs)
}

func (ste *scheduledTxsExecution) getAllIntermediateTxsAfterScheduledExecution(
	mapAllIntermediateTxsBeforeScheduledExecution map[block.Type]map[string]data.TransactionHandler,
	allIntermediateTxsAfterScheduledExecution map[string]data.TransactionHandler,
	blockType block.Type,
) []*scrInfo {
	scrsInfo := make([]*scrInfo, 0)
	for txHash, txHandler := range allIntermediateTxsAfterScheduledExecution {
		scrs, blockTypeExists := mapAllIntermediateTxsBeforeScheduledExecution[blockType]
		if blockTypeExists {
			_, txExists := scrs[txHash]
			if txExists {
				continue
			}
		}

		if ste.shardCoordinator.SameShard(txHandler.GetSndAddr(), txHandler.GetRcvAddr()) {
			log.Trace("scheduledTxsExecution.getAllIntermediateTxsAfterScheduledExecution: intra shard scr skipped", "hash", []byte(txHash))
			continue
		}

		scrsInfo = append(scrsInfo, &scrInfo{
			txHash:    []byte(txHash),
			txHandler: txHandler,
		})
	}

	return scrsInfo
}

// GetScheduledSCRs gets the resulted SCRs after the execution of scheduled transactions
func (ste *scheduledTxsExecution) GetScheduledSCRs() map[block.Type][]data.TransactionHandler {
	ste.mutScheduledTxs.RLock()
	defer ste.mutScheduledTxs.RUnlock()

	numScheduledSCRs := 0
	mapScheduledSCRs := make(map[block.Type][]data.TransactionHandler)
	for blockType, scheduledSCRs := range ste.mapScheduledSCRs {
		if len(scheduledSCRs) == 0 {
			continue
		}

		mapScheduledSCRs[blockType] = make([]data.TransactionHandler, len(scheduledSCRs))
		for scrIndex, txHandler := range scheduledSCRs {
			mapScheduledSCRs[blockType][scrIndex] = txHandler
			log.Trace("scheduledTxsExecution.GetScheduledSCRs", "blockType", blockType, "sender", mapScheduledSCRs[blockType][scrIndex].GetSndAddr(), "receiver", mapScheduledSCRs[blockType][scrIndex].GetRcvAddr())
		}
		numScheduledSCRs += len(scheduledSCRs)
	}

	log.Debug("scheduledTxsExecution.GetScheduledSCRs", "num of scheduled scrs", numScheduledSCRs)

	return mapScheduledSCRs
}

// SetScheduledRootHashSCRsGasAndFees sets the resulted scheduled root hash, SCRs, gas and fees after the execution of scheduled transactions
func (ste *scheduledTxsExecution) SetScheduledRootHashSCRsGasAndFees(rootHash []byte, mapSCRs map[block.Type][]data.TransactionHandler, gasAndFees scheduled.GasAndFees) {
	ste.mutScheduledTxs.Lock()
	defer ste.mutScheduledTxs.Unlock()

	ste.scheduledRootHash = rootHash

	numScheduledSCRs := 0
	ste.mapScheduledSCRs = make(map[block.Type][]data.TransactionHandler)
	for blockType, scrs := range mapSCRs {
		if len(scrs) == 0 {
			continue
		}

		ste.mapScheduledSCRs[blockType] = make([]data.TransactionHandler, len(scrs))
		for scrIndex, txHandler := range scrs {
			ste.mapScheduledSCRs[blockType][scrIndex] = txHandler
			log.Trace("scheduledTxsExecution.SetScheduledRootHashSCRsGasAndFees", "blockType", blockType, "sender", ste.mapScheduledSCRs[blockType][scrIndex].GetSndAddr(), "receiver", ste.mapScheduledSCRs[blockType][scrIndex].GetRcvAddr())
		}

		numScheduledSCRs += len(scrs)
	}

	ste.gasAndFees = gasAndFees

	log.Debug("scheduledTxsExecution.SetScheduledRootHashSCRsGasAndFees",
		"scheduled root hash", rootHash,
		"num of scheduled scrs", numScheduledSCRs,
		"accumulatedFees", gasAndFees.AccumulatedFees.String(),
		"developerFees", gasAndFees.DeveloperFees.String(),
		"gasProvided", gasAndFees.GasProvided,
		"gasPenalized", gasAndFees.GasPenalized,
		"gasRefunded", gasAndFees.GasRefunded)
}

// GetScheduledRootHash gets the resulted root hash after the execution of scheduled transactions
func (ste *scheduledTxsExecution) GetScheduledRootHash() []byte {
	ste.mutScheduledTxs.RLock()
	rootHash := ste.scheduledRootHash
	ste.mutScheduledTxs.RUnlock()

	log.Debug("scheduledTxsExecution.GetScheduledRootHash", "scheduled root hash", rootHash)

	return rootHash
}

// GetScheduledGasAndFees returns the gas and fees for the scheduled transactions in last processed block
// if there are no scheduled transactions in the last processed block, the returned struct has zero values
func (ste *scheduledTxsExecution) GetScheduledGasAndFees() scheduled.GasAndFees {
	ste.mutScheduledTxs.RLock()
	gasAndFees := ste.gasAndFees
	ste.mutScheduledTxs.RUnlock()

	log.Debug("scheduledTxsExecution.GetScheduledGasAndFees",
		"accumulatedFees", gasAndFees.AccumulatedFees.String(),
		"developerFees", gasAndFees.DeveloperFees.String(),
		"gasProvided", gasAndFees.GasProvided,
		"gasPenalized", gasAndFees.GasPenalized,
		"gasRefunded", gasAndFees.GasRefunded)

	return gasAndFees
}

// SetScheduledRootHash sets the resulted root hash after the execution of scheduled transactions
func (ste *scheduledTxsExecution) SetScheduledRootHash(rootHash []byte) {
	ste.mutScheduledTxs.Lock()
	ste.scheduledRootHash = rootHash
	ste.mutScheduledTxs.Unlock()

	log.Debug("scheduledTxsExecution.SetScheduledRootHash", "scheduled root hash", rootHash)
}

// SetScheduledGasAndFees sets the gas and fees for the scheduled transactions
func (ste *scheduledTxsExecution) SetScheduledGasAndFees(gasAndFees scheduled.GasAndFees) {
	ste.mutScheduledTxs.Lock()
	ste.gasAndFees = gasAndFees
	ste.mutScheduledTxs.Unlock()

	log.Debug("scheduledTxsExecution.SetScheduledGasAndFees",
		"accumulatedFees", gasAndFees.AccumulatedFees.String(),
		"developerFees", gasAndFees.DeveloperFees.String(),
		"gasProvided", gasAndFees.GasProvided,
		"gasPenalized", gasAndFees.GasPenalized,
		"gasRefunded", gasAndFees.GasRefunded)
}

// SetTransactionProcessor sets the transaction processor needed by scheduled txs execution component
func (ste *scheduledTxsExecution) SetTransactionProcessor(txProcessor process.TransactionProcessor) {
	ste.txProcessor = txProcessor
}

// SetTransactionCoordinator sets the transaction coordinator needed by scheduled txs execution component
func (ste *scheduledTxsExecution) SetTransactionCoordinator(txCoordinator process.TransactionCoordinator) {
	ste.txCoordinator = txCoordinator
}

// GetScheduledRootHashForHeader gets scheduled root hash of the given header from storage
func (ste *scheduledTxsExecution) GetScheduledRootHashForHeader(
	headerHash []byte,
) ([]byte, error) {
	rootHash, _, _, err := ste.getScheduledRootHashSCRsGasAndFeesForHeader(headerHash)

	log.Debug("scheduledTxsExecution.GetScheduledRootHashForHeader", "header hash", headerHash, "scheduled root hash", rootHash)

	return rootHash, err
}

// RollBackToBlock rolls back the scheduled txs execution handler to the given header
func (ste *scheduledTxsExecution) RollBackToBlock(headerHash []byte) error {
	scheduledRootHash, mapScheduledSCRs, gasAndFees, err := ste.getScheduledRootHashSCRsGasAndFeesForHeader(headerHash)
	if err != nil {
		return err
	}

	log.Debug("scheduledTxsExecution.RollBackToBlock",
		"header hash", headerHash,
		"scheduled root hash", scheduledRootHash,
		"num of scheduled scrs", getNumScheduledSCRs(mapScheduledSCRs),
		"accumulatedFees", gasAndFees.AccumulatedFees.String(),
		"developerFees", gasAndFees.DeveloperFees.String(),
		"gasProvided", gasAndFees.GasProvided,
		"gasPenalized", gasAndFees.GasPenalized,
		"gasRefunded", gasAndFees.GasRefunded)

	ste.SetScheduledRootHashSCRsGasAndFees(scheduledRootHash, mapScheduledSCRs, *gasAndFees)

	return nil
}

// SaveStateIfNeeded saves the scheduled Txs Execution state for the given header hash, if there are scheduled TXs
func (ste *scheduledTxsExecution) SaveStateIfNeeded(headerHash []byte) {
	scheduledRootHash := ste.GetScheduledRootHash()
	mapScheduledSCRs := ste.GetScheduledSCRs()
	ste.mutScheduledTxs.RLock()
	gasAndFees := ste.gasAndFees
	numScheduledTxs := len(ste.scheduledTxs)
	ste.mutScheduledTxs.RUnlock()
	log.Debug("scheduledTxsExecution.SaveStateIfNeeded",
		"header hash", headerHash,
		"scheduled root hash", scheduledRootHash,
		"num of scheduled txs", numScheduledTxs,
		"num of scheduled scrs", getNumScheduledSCRs(mapScheduledSCRs),
		"accumulatedFees", gasAndFees.AccumulatedFees.String(),
		"developerFees", gasAndFees.DeveloperFees.String(),
		"gasProvided", gasAndFees.GasProvided,
		"gasPenalized", gasAndFees.GasPenalized,
		"gasRefunded", gasAndFees.GasRefunded)

	if numScheduledTxs > 0 {
		ste.SaveState(headerHash, scheduledRootHash, mapScheduledSCRs, gasAndFees)
	}
}

// SaveState saves the scheduled SC execution state
func (ste *scheduledTxsExecution) SaveState(
	headerHash []byte,
	scheduledRootHash []byte,
	mapScheduledSCRs map[block.Type][]data.TransactionHandler,
	gasAndFees scheduled.GasAndFees,
) {
	marshalledScheduledData, err := ste.getMarshalledScheduledRootHashSCRsGasAndFees(scheduledRootHash, mapScheduledSCRs, gasAndFees)
	if err != nil {
		log.Warn("scheduledTxsExecution.SaveState getMarshalledScheduledRootHashSCRsGasAndFees", "error", err.Error())
		return
	}

	log.Debug("scheduledTxsExecution.SaveState Put",
		"header hash", headerHash,
		"scheduled root hash", scheduledRootHash,
		"num of scheduled scrs", getNumScheduledSCRs(mapScheduledSCRs),
		"gasAndFees.AccumulatedFees", gasAndFees.AccumulatedFees.String(),
		"gasAndFees.DeveloperFees", gasAndFees.DeveloperFees.String(),
		"gasAndFees.GasProvided", gasAndFees.GasProvided,
		"gasAndFees.GasPenalized", gasAndFees.GasPenalized,
		"gasAndFees.GasRefunded", gasAndFees.GasRefunded,
		"length of marshalized scheduled SCRs", len(marshalledScheduledData))
	err = ste.storer.Put(headerHash, marshalledScheduledData)
	if err != nil {
		log.Warn("scheduledTxsExecution.SaveState Put -> ScheduledSCRsUnit", "error", err.Error())
	}
}

// getScheduledRootHashSCRsGasAndFeesForHeader gets scheduled root hash, the SCRs, gas and fees of the given header from storage
func (ste *scheduledTxsExecution) getScheduledRootHashSCRsGasAndFeesForHeader(
	headerHash []byte,
) ([]byte, map[block.Type][]data.TransactionHandler, *scheduled.GasAndFees, error) {
	var err error
	defer func() {
		if err != nil {
			log.Trace("getScheduledRootHashSCRsGasAndFeesForHeader: given header does not have scheduled txs",
				"header hash", headerHash,
			)
		}
	}()

	marshalledSCRsSavedData, err := ste.storer.Get(headerHash)
	if err != nil {
		return nil, nil, nil, err
	}

	scheduledSCRs := &scheduled.ScheduledSCRs{}
	err = ste.marshaller.Unmarshal(scheduledSCRs, marshalledSCRsSavedData)
	if err != nil {
		return nil, nil, nil, err
	}

	scheduledRootHash := scheduledSCRs.RootHash
	txHandlersMap := scheduledSCRs.GetTransactionHandlersMap()
	gasAndFees := *scheduledSCRs.GasAndFees

	return scheduledRootHash, txHandlersMap, &gasAndFees, nil
}

func (ste *scheduledTxsExecution) getMarshalledScheduledRootHashSCRsGasAndFees(
	scheduledRootHash []byte,
	mapScheduledSCRs map[block.Type][]data.TransactionHandler,
	gasAndFees scheduled.GasAndFees,
) ([]byte, error) {
	scheduledSCRs := &scheduled.ScheduledSCRs{
		RootHash:   scheduledRootHash,
		GasAndFees: &gasAndFees,
	}

	err := scheduledSCRs.SetTransactionHandlersMap(mapScheduledSCRs)
	if err != nil {
		return nil, err
	}

	return ste.marshaller.Marshal(scheduledSCRs)
}

// IsScheduledTx returns true if the given txHash was scheduled for execution for the current block
func (ste *scheduledTxsExecution) IsScheduledTx(txHash []byte) bool {
	ste.mutScheduledTxs.RLock()
	_, ok := ste.mapScheduledTxs[string(txHash)]
	ste.mutScheduledTxs.RUnlock()

	return ok
}

func getNumScheduledSCRs(mapScheduledSCRs map[block.Type][]data.TransactionHandler) int {
	numScheduledSCRs := 0
	for _, scheduledSCRs := range mapScheduledSCRs {
		numScheduledSCRs += len(scheduledSCRs)
	}

	return numScheduledSCRs
}

// IsInterfaceNil returns true if there is no value under the interface
func (ste *scheduledTxsExecution) IsInterfaceNil() bool {
	return ste == nil
}
