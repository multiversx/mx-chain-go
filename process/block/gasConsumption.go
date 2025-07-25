package block

import (
	"fmt"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
)

// gasType defines the type of gas consumption
type gasType string

const (
	incoming gasType = "incoming"
	outgoing gasType = "outgoing"

	// TODO: move these to config
	initialLimitsFactor       = uint64(200) // 200%
	percentSplitBlock         = uint64(50)  // 50%
	percentDecreaseLimitsStep = uint64(10)  // 10%
	minPercentLimitsFactor    = uint64(10)  // 10%
	initialLastIndex          = -1
)

// ArgsGasConsumption holds the arguments needed to create a gasConsumption instance
type ArgsGasConsumption struct {
	EconomicsFee     process.FeeHandler
	ShardCoordinator process.ShardCoordinator
	GasHandler       process.GasHandler
}

// gasConsumption implements the GasComputation interface for managing gas limits during block creation
type gasConsumption struct {
	sync.RWMutex
	economicsFee                     process.FeeHandler
	shardCoordinator                 process.ShardCoordinator
	gasHandler                       process.GasHandler
	totalGasConsumed                 map[string]uint64
	gasConsumedByMiniBlock           map[string]uint64
	pendingMiniBlocks                []data.MiniBlockHeaderHandler
	pendingTransactions              []data.TransactionHandler
	transactionsForPendingMiniBlocks map[string][]data.TransactionHandler
	lastMiniBlockIndex               int
	lastTransactionIndex             int
	isMiniBlockSelectionDone         bool
	isTransactionSelectionDone       bool
	miniBlockLimitFactor             uint64
	blockLimitFactor                 uint64
}

// NewGasConsumption creates a new instance of gasConsumption
func NewGasConsumption(args ArgsGasConsumption) (*gasConsumption, error) {
	if check.IfNil(args.EconomicsFee) {
		return nil, process.ErrNilEconomicsFeeHandler
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}
	if check.IfNil(args.GasHandler) {
		return nil, process.ErrNilGasHandler
	}

	return &gasConsumption{
		economicsFee:                     args.EconomicsFee,
		shardCoordinator:                 args.ShardCoordinator,
		gasHandler:                       args.GasHandler,
		totalGasConsumed:                 make(map[string]uint64),
		gasConsumedByMiniBlock:           make(map[string]uint64),
		transactionsForPendingMiniBlocks: make(map[string][]data.TransactionHandler),
		miniBlockLimitFactor:             initialLimitsFactor,
		blockLimitFactor:                 initialLimitsFactor,
		lastMiniBlockIndex:               initialLastIndex,
		lastTransactionIndex:             initialLastIndex,
	}, nil
}

// CheckIncomingMiniBlocks verifies if an incoming mini block and its transactions can be included within gas limits
// returns the last mini block index included, the number of pending mini blocks left and error if needed
func (gc *gasConsumption) CheckIncomingMiniBlocks(
	miniBlocks []data.MiniBlockHeaderHandler,
	transactions map[string][]data.TransactionHandler,
) (int, int, error) {
	if len(miniBlocks) == 0 || len(transactions) == 0 {
		return gc.lastMiniBlockIndex, 0, nil
	}

	gc.Lock()
	defer gc.Unlock()

	if gc.isMiniBlockSelectionDone {
		return initialLastIndex, 0, process.ErrMiniBlocksAlreadyProcessed
	}

	defer func() {
		// no matter the result, mark the mini blocks selection done
		gc.isMiniBlockSelectionDone = true
	}()

	blockGasLimitForOneDirection := gc.getGasLimitForOneDirection(gc.shardCoordinator.SelfId())

	for i := 0; i < len(miniBlocks); i++ {
		shouldSavePending, shouldStop, err := gc.checkIncomingMiniBlock(miniBlocks[i], transactions, blockGasLimitForOneDirection)
		if shouldSavePending {
			gc.pendingMiniBlocks = append(gc.pendingMiniBlocks, miniBlocks[i:]...)
			gc.transactionsForPendingMiniBlocks = transactions

			return gc.lastMiniBlockIndex, len(gc.pendingMiniBlocks), err
		}
		if err != nil || shouldStop {
			return gc.lastMiniBlockIndex, 0, err
		}
	}

	// reaching this point means that mini blocks were added and the limit for incoming was not reached
	err := gc.checkPendingOutgoingTransactions()
	return gc.lastMiniBlockIndex, 0, err
}

func (gc *gasConsumption) checkIncomingMiniBlock(
	mb data.MiniBlockHeaderHandler,
	transactions map[string][]data.TransactionHandler,
	blockGasLimitForOneDirection uint64,
) (bool, bool, error) {
	if mb == nil {
		return false, true, nil
	}

	mbHash := mb.GetHash()
	numTxs := mb.GetTxCount()

	transactionsForMB, found := transactions[string(mbHash)]
	if !found {
		// do not save any pending mini blocks
		return false, true, fmt.Errorf("%w, could not find mini block hash in transactions map", process.ErrInvalidValue)
	}

	if int(numTxs) != len(transactionsForMB) {
		// do not save any pending mini blocks
		return false, true, fmt.Errorf("%w, the provided mini block does not match the number of transactions provided", process.ErrInvalidValue)
	}

	gasConsumedByMB := uint64(0)
	for j := 0; j < len(transactionsForMB); j++ {
		tx := transactionsForMB[j]
		if check.IfNil(tx) {
			continue
		}

		// we only care about the gas consumed in receiver shard as all mini blocks are coming to current shard
		_, gasConsumedInReceiverShard, err := gc.gasHandler.ComputeGasProvidedByTx(mb.GetSenderShardID(), mb.GetReceiverShardID(), tx)
		if err != nil {
			return false, true, err
		}

		maxGasLimitPerTx := gc.economicsFee.MaxGasLimitPerTx()
		if gasConsumedInReceiverShard > maxGasLimitPerTx {
			// this should not happen, transactions with higher gas should have been already rejected
			// return the last saved mini block, and the proper error, without saving the rest of mini blocks
			return false, true, process.ErrMaxGasLimitPerTransactionIsReached
		}

		gasConsumedByMB += gasConsumedInReceiverShard
	}

	maxGasLimitPerMB := gc.maxGasLimitPerMiniBlock(mb.GetReceiverShardID())
	if gasConsumedByMB > maxGasLimitPerMB {
		// return the last saved mini block, and the proper error, without saving the rest of mini blocks
		return false, true, process.ErrMaxGasLimitPerMiniBlockIsReached
	}

	mbsLimitReached := gc.totalGasConsumed[string(incoming)]+gasConsumedByMB > blockGasLimitForOneDirection
	if !mbsLimitReached {
		// limit not reached, continue
		// simply increment the lastMiniBlockIndex as this method might be called either from
		// handling all mini blocks, either from handling pending, where the pending ones
		// should have continuos indexes after the ones already included
		gc.lastMiniBlockIndex++
		gc.gasConsumedByMiniBlock[string(mbHash)] = gasConsumedByMB
		gc.totalGasConsumed[string(incoming)] += gasConsumedByMB

		return false, false, nil
	}

	// if limit is reached, check if outgoing transactions were already handled
	// if transactions not handled yet:
	//	- save the rest of mini blocks as pending and stop
	// if transactions handled and:
	//	- there is some space left, add more mini blocks
	//	- there is no space left, return the last index added, do not save the pending mini blocks
	if !gc.isTransactionSelectionDone {
		// transactions not handled yet, save the rest of mini blocks, return latest index and no error
		return true, true, nil
	}

	gasConsumedByOutgoingTransactions := gc.getTotalOutgoingGas()
	gasLeftFromTransactions := blockGasLimitForOneDirection - gasConsumedByOutgoingTransactions
	if gasLeftFromTransactions < maxGasLimitPerMB {
		// transactions handled, but the remaining space is not enough
		// returning no error, not saving pending mini blocks, block is full
		return false, true, nil
	}

	return false, false, nil
}

func (gc *gasConsumption) checkPendingIncomingMiniBlocks() error {
	// checking if any pending mini blocks are left in order to fill the block
	hasPendingMiniBlocks := gc.isMiniBlockSelectionDone && len(gc.pendingMiniBlocks) > 0
	if !hasPendingMiniBlocks {
		return nil
	}

	blockGasLimitForOneDirection := gc.getGasLimitForOneDirection(gc.shardCoordinator.SelfId())
	for i := 0; i < len(gc.pendingMiniBlocks); i++ {
		mb := gc.pendingMiniBlocks[i]
		_, shouldStop, err := gc.checkIncomingMiniBlock(mb, gc.transactionsForPendingMiniBlocks, blockGasLimitForOneDirection)
		if err != nil || shouldStop {
			return err
		}
	}

	return nil
}

// CheckOutgoingTransactions verifies the outgoing transactions and returns the index of the last valid transaction
// only returns error if a transaction is invalid, with too much gas
func (gc *gasConsumption) CheckOutgoingTransactions(transactions []data.TransactionHandler) (int, error) {
	if len(transactions) == 0 {
		return initialLastIndex, nil
	}

	gc.Lock()
	defer gc.Unlock()

	if gc.isTransactionSelectionDone {
		return initialLastIndex, process.ErrTransactionsAlreadyProcessed
	}

	defer func() {
		// no matter the result, mark the transactions selection done
		gc.isTransactionSelectionDone = true
	}()

	for i := 0; i < len(transactions); i++ {
		shouldSavePending, shouldStop, err := gc.checkOutgoingTransaction(transactions[i])
		if shouldSavePending {
			gc.pendingTransactions = append(gc.pendingTransactions, transactions[i:]...)

			return gc.lastTransactionIndex, err
		}
		if err != nil || shouldStop {
			return gc.lastTransactionIndex, err
		}
	}

	// reaching this point means that transactions were added and the limit for outgoing was not reached
	err := gc.checkPendingIncomingMiniBlocks()
	return gc.lastTransactionIndex, err
}

// must be called under mutex protection
func (gc *gasConsumption) checkOutgoingTransaction(
	tx data.TransactionHandler,
) (bool, bool, error) {
	if check.IfNil(tx) {
		return false, true, nil
	}

	senderShard := gc.shardCoordinator.SelfId()
	receiverShard := gc.shardCoordinator.ComputeId(tx.GetRcvAddr())
	isCrossShard := senderShard != receiverShard
	gasConsumedInSenderShard, gasConsumedInReceiverShard, err := gc.gasHandler.ComputeGasProvidedByTx(senderShard, receiverShard, tx)
	if err != nil {
		return false, true, err
	}
	maxGasLimitPerTx := gc.economicsFee.MaxGasLimitPerTx()
	if gasConsumedInSenderShard > maxGasLimitPerTx || gasConsumedInReceiverShard > maxGasLimitPerTx {
		// this should not happen, transactions with higher gas should have been already rejected
		// return the last saved transaction, and the proper error, without saving the rest of transactions
		return false, true, process.ErrMaxGasLimitPerTransactionIsReached
	}

	outgoingSelfKey := getOutgoingKey(senderShard)
	outgoingDestKey := getOutgoingKey(receiverShard)
	bandwidthForOutgoingTransactions := gc.getGasLimitForOneDirection(senderShard)
	txsLimitReachedForSelfShard := gc.totalGasConsumed[outgoingSelfKey]+gasConsumedInSenderShard > bandwidthForOutgoingTransactions
	txsLimitReachedForDestShard := gc.totalGasConsumed[outgoingDestKey]+gasConsumedInReceiverShard > bandwidthForOutgoingTransactions
	gasConsumedByOutgoingTransactions := gc.getTotalOutgoingGas()
	txsLimitReachedForTotalOutgoing := gasConsumedByOutgoingTransactions+gasConsumedInSenderShard > bandwidthForOutgoingTransactions
	txsLimitReachedInAnyShard := txsLimitReachedForSelfShard || txsLimitReachedForDestShard || txsLimitReachedForTotalOutgoing
	if !txsLimitReachedInAnyShard {
		// limit not reached, continue
		// simply increment the lastTransactionIndex as this method might be called either from
		// handling all transactions, either from handling pending, where the pending ones
		// should have continuos indexes after the ones already included
		// also add the consumed gas, for receiver shard too if needed
		gc.lastTransactionIndex++
		gc.totalGasConsumed[outgoingSelfKey] += gasConsumedInSenderShard
		if isCrossShard {
			gc.totalGasConsumed[outgoingDestKey] += gasConsumedInReceiverShard
		}

		return false, false, nil
	}

	// if limit is reached for any shard, transaction won't be added
	// if mini blocks not handled yet:
	//	- save the rest of transactions as pending
	// if mini blocks handled and:
	//	- there is some space left, add more transactions
	//	- there is no space left, return the last index added, do not save the pending transactions
	if !gc.isMiniBlockSelectionDone {
		// mini blocks not handled yet, save the rest of transactions, return no error
		return true, true, nil
	}

	gasConsumedByIncomingMiniBlocks := gc.totalGasConsumed[string(incoming)]
	totalBandwidthForIncoming := gc.getGasLimitForOneDirection(senderShard)
	bandwidthLeftFromIncomingMiniBlocks := totalBandwidthForIncoming - gasConsumedByIncomingMiniBlocks
	// check if at least one tx can be added
	if bandwidthLeftFromIncomingMiniBlocks < maxGasLimitPerTx {
		// mini blocks handled, but the remaining space is not enough
		// returning no error, not saving pending transactions, block is full
		return false, true, nil
	}

	return false, false, nil
}

func (gc *gasConsumption) checkPendingOutgoingTransactions() error {
	// checking if any pending transactions are left in order to fill the block
	hasPendingTransactions := gc.isTransactionSelectionDone && len(gc.pendingTransactions) > 0
	if !hasPendingTransactions {
		return nil
	}

	for i := 0; i < len(gc.pendingTransactions); i++ {
		tx := gc.pendingTransactions[i]
		_, shouldStop, err := gc.checkOutgoingTransaction(tx)
		if err != nil || shouldStop {
			return err
		}
	}

	return nil
}

// TotalGasConsumed returns the total gas consumed for both incoming and outgoing transactions
func (gc *gasConsumption) TotalGasConsumed() uint64 {
	gc.RLock()
	defer gc.RUnlock()

	totalGasConsumed := uint64(0)
	for _, gasConsumed := range gc.totalGasConsumed {
		totalGasConsumed += gasConsumed
	}

	return totalGasConsumed
}

// GetLastMiniBlockIndexIncluded returns the last mini block index added
func (gc *gasConsumption) GetLastMiniBlockIndexIncluded() int {
	gc.RLock()
	defer gc.RUnlock()

	return gc.lastMiniBlockIndex
}

// GetLasTransactionIndexIncluded returns the last transactions index added
func (gc *gasConsumption) GetLasTransactionIndexIncluded() int {
	gc.RLock()
	defer gc.RUnlock()

	return gc.lastTransactionIndex
}

// DecreaseMiniBlockLimit reduces the mini block gas limit by a configured percentage
func (gc *gasConsumption) DecreaseMiniBlockLimit() {
	gc.Lock()
	defer gc.Unlock()

	if gc.miniBlockLimitFactor == minPercentLimitsFactor {
		return
	}

	decreaseStep := initialLimitsFactor * percentDecreaseLimitsStep / 100
	gc.miniBlockLimitFactor = gc.miniBlockLimitFactor - decreaseStep

	if gc.miniBlockLimitFactor <= minPercentLimitsFactor {
		gc.miniBlockLimitFactor = minPercentLimitsFactor
	}
}

// ResetMiniBlockLimit resets the mini block gas limit to its initial value
func (gc *gasConsumption) ResetMiniBlockLimit() {
	gc.Lock()
	defer gc.Unlock()

	gc.miniBlockLimitFactor = initialLimitsFactor
}

// DecreaseBlockLimit reduces the block gas limit by a configured percentage
func (gc *gasConsumption) DecreaseBlockLimit() {
	gc.Lock()
	defer gc.Unlock()

	if gc.blockLimitFactor == minPercentLimitsFactor {
		return
	}

	decreaseStep := initialLimitsFactor * percentDecreaseLimitsStep / 100
	gc.blockLimitFactor = gc.blockLimitFactor - decreaseStep

	if gc.blockLimitFactor <= minPercentLimitsFactor {
		gc.blockLimitFactor = minPercentLimitsFactor
	}
}

// ResetBlockLimit resets the block gas limit to its initial value
func (gc *gasConsumption) ResetBlockLimit() {
	gc.Lock()
	defer gc.Unlock()

	gc.blockLimitFactor = initialLimitsFactor
}

// Reset clears all gas consumption data except for the limits factors
func (gc *gasConsumption) Reset() {
	gc.Lock()
	defer gc.Unlock()

	gc.totalGasConsumed = make(map[string]uint64)
	gc.gasConsumedByMiniBlock = make(map[string]uint64)
	gc.isTransactionSelectionDone = false
	gc.isMiniBlockSelectionDone = false
	gc.pendingMiniBlocks = make([]data.MiniBlockHeaderHandler, 0)
	gc.pendingTransactions = make([]data.TransactionHandler, 0)
	gc.transactionsForPendingMiniBlocks = make(map[string][]data.TransactionHandler, 0)
	gc.lastMiniBlockIndex = initialLastIndex
	gc.lastTransactionIndex = initialLastIndex
}

func (gc *gasConsumption) getGasLimitForOneDirection(shardID uint32) uint64 {
	totalBlockLimit := gc.maxGasLimitPerBlock(shardID)
	return totalBlockLimit * percentSplitBlock / 100
}

// must be called under mutex protection as it access blockLimitFactor
func (gc *gasConsumption) maxGasLimitPerBlock(shardID uint32) uint64 {
	isCrossShard := shardID != gc.shardCoordinator.SelfId()
	if isCrossShard {
		return gc.economicsFee.MaxGasLimitPerBlockForSafeCrossShard() * gc.blockLimitFactor / 100
	}

	return gc.economicsFee.MaxGasLimitPerBlock(shardID) * gc.blockLimitFactor / 100
}

// must be called under mutex protection as it access miniBlockLimitFactor
func (gc *gasConsumption) maxGasLimitPerMiniBlock(shardID uint32) uint64 {
	isCrossShard := shardID != gc.shardCoordinator.SelfId()
	if isCrossShard {
		return gc.economicsFee.MaxGasLimitPerMiniBlockForSafeCrossShard() * gc.miniBlockLimitFactor / 100
	}

	return gc.economicsFee.MaxGasLimitPerMiniBlock(shardID) * gc.miniBlockLimitFactor / 100
}

// must be called under mutex protection as it access totalGasConsumed
func (gc *gasConsumption) getTotalOutgoingGas() uint64 {
	totalGasConsumed := uint64(0)
	for key, gasConsumed := range gc.totalGasConsumed {
		if !strings.Contains(key, string(outgoing)) {
			continue
		}

		totalGasConsumed += gasConsumed
	}

	return totalGasConsumed
}

func getOutgoingKey(shardID uint32) string {
	return fmt.Sprintf("%s_%d", outgoing, shardID)
}

// IsInterfaceNil checks if the interface is nil
func (gc *gasConsumption) IsInterfaceNil() bool {
	return gc == nil
}
