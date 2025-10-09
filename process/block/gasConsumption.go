package block

import (
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
)

// gasType defines the type of gas consumption
type gasType uint8

const (
	incoming gasType = iota
	outgoingIntra
	outgoingCross
)

const (
	zeroLimitsFactor       = uint64(0)
	minPercentLimitsFactor = uint64(10)  // 10%
	maxPercentLimitsFactor = uint64(500) // 500%
	percentSplitBlock      = uint64(50)  // 50%
	initialLastIndex       = -1
)

// ArgsGasConsumption holds the arguments needed to create a gasConsumption instance
type ArgsGasConsumption struct {
	EconomicsFee                      process.FeeHandler
	ShardCoordinator                  process.ShardCoordinator
	GasHandler                        process.GasHandler
	BlockCapacityOverestimationFactor uint64
	PercentDecreaseLimitsStep         uint64
}

// gasConsumption implements the GasComputation interface for managing gas limits during block creation
type gasConsumption struct {
	mut                              sync.RWMutex
	economicsFee                     process.FeeHandler
	shardCoordinator                 process.ShardCoordinator
	gasHandler                       process.GasHandler
	totalGasConsumed                 map[gasType]uint64
	gasConsumedByMiniBlock           map[string]uint64
	pendingMiniBlocks                []data.MiniBlockHeaderHandler
	transactionsForPendingMiniBlocks map[string][]data.TransactionHandler
	incomingLimitFactor              uint64
	outgoingLimitFactor              uint64
	decreaseStep                     uint64
	initialOverestimationFactor      uint64
	percentDecreaseLimitsStep        uint64
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
	if args.BlockCapacityOverestimationFactor <= minPercentLimitsFactor || args.BlockCapacityOverestimationFactor > maxPercentLimitsFactor {
		return nil, fmt.Errorf("%w for BlockCapacityOverestimationFactor, received %d, min allowed %d, max allowed %d",
			process.ErrInvalidValue,
			args.BlockCapacityOverestimationFactor,
			minPercentLimitsFactor,
			maxPercentLimitsFactor)
	}

	return &gasConsumption{
		economicsFee:                     args.EconomicsFee,
		shardCoordinator:                 args.ShardCoordinator,
		gasHandler:                       args.GasHandler,
		totalGasConsumed:                 make(map[gasType]uint64),
		gasConsumedByMiniBlock:           make(map[string]uint64),
		transactionsForPendingMiniBlocks: make(map[string][]data.TransactionHandler),
		incomingLimitFactor:              args.BlockCapacityOverestimationFactor,
		outgoingLimitFactor:              args.BlockCapacityOverestimationFactor,
		decreaseStep:                     args.BlockCapacityOverestimationFactor * args.PercentDecreaseLimitsStep / 100,
		percentDecreaseLimitsStep:        args.PercentDecreaseLimitsStep,
		initialOverestimationFactor:      args.BlockCapacityOverestimationFactor,
	}, nil
}

// CheckIncomingMiniBlocks verifies if an incoming mini block and its transactions can be included within gas limits
// returns the last mini block index included, the number of pending mini blocks left and error if needed
// This must be called first, before CheckOutgoingTransactions!
func (gc *gasConsumption) CheckIncomingMiniBlocks(
	miniBlocks []data.MiniBlockHeaderHandler,
	transactions map[string][]data.TransactionHandler,
) (lastMiniBlockIndex int, pendingMiniBlocks int, err error) {
	if len(miniBlocks) == 0 || len(transactions) == 0 {
		return initialLastIndex, 0, nil
	}

	gc.mut.Lock()
	defer gc.mut.Unlock()

	if gc.incomingLimitFactor == 0 {
		return initialLastIndex, 0, fmt.Errorf("%w for incoming mini blocks", process.ErrZeroLimit)
	}

	bandwidthForIncomingMiniBlocks := gc.getGasLimitForOneDirection(incoming, gc.shardCoordinator.SelfId())

	lastMiniBlockIndex = initialLastIndex
	shouldSavePending := false
	for i := 0; i < len(miniBlocks); i++ {
		shouldSavePending, err = gc.checkIncomingMiniBlock(miniBlocks[i], transactions, bandwidthForIncomingMiniBlocks)
		if shouldSavePending {
			gc.pendingMiniBlocks = append(gc.pendingMiniBlocks, miniBlocks[i:]...)
			gc.transactionsForPendingMiniBlocks = transactions

			return lastMiniBlockIndex, len(gc.pendingMiniBlocks), err
		}
		if err != nil {
			return lastMiniBlockIndex, 0, err
		}
		lastMiniBlockIndex = i
	}

	return lastMiniBlockIndex, 0, nil
}

func (gc *gasConsumption) checkIncomingMiniBlock(
	mb data.MiniBlockHeaderHandler,
	transactions map[string][]data.TransactionHandler,
	bandwidthForIncomingMiniBlocks uint64,
) (bool, error) {
	if mb == nil {
		return false, nil
	}

	mbHash := mb.GetHash()
	numTxs := mb.GetTxCount()

	transactionsForMB, found := transactions[string(mbHash)]
	if !found {
		// do not save any pending mini blocks, as this one is invalid
		return false, fmt.Errorf("%w, could not find mini block hash in transactions map", process.ErrInvalidValue)
	}

	if int(numTxs) != len(transactionsForMB) {
		// do not save any pending mini blocks, as this one is invalid
		return false, fmt.Errorf("%w, the provided mini block does not match the number of transactions provided", process.ErrInvalidValue)
	}

	gasConsumedByMB, err := gc.checkGasConsumedByMiniBlock(mb, transactionsForMB)
	if err != nil {
		return false, err
	}

	mbsLimitReached := gc.totalGasConsumed[incoming]+gasConsumedByMB > bandwidthForIncomingMiniBlocks
	if !mbsLimitReached {
		// limit not reached, continue
		// this method might be called either from handling all mini blocks,
		// either from handling pending, where the pending ones
		// should have continuous indexes after the ones already included
		gc.gasConsumedByMiniBlock[string(mbHash)] = gasConsumedByMB
		gc.totalGasConsumed[incoming] += gasConsumedByMB

		return false, nil
	}

	return true, nil
}

func (gc *gasConsumption) checkGasConsumedByMiniBlock(mb data.MiniBlockHeaderHandler, transactionsForMB []data.TransactionHandler) (uint64, error) {
	gasConsumedByMB := uint64(0)
	for i := 0; i < len(transactionsForMB); i++ {
		tx := transactionsForMB[i]
		if check.IfNil(tx) {
			continue
		}

		// we only care about the gas consumed in receiver shard as all mini blocks are coming to current shard
		_, gasConsumedInReceiverShard, err := gc.gasHandler.ComputeGasProvidedByTx(mb.GetSenderShardID(), mb.GetReceiverShardID(), tx)
		if err != nil {
			// do not save any pending mini blocks, as this one is invalid
			return 0, err
		}

		maxGasLimitPerTx := gc.economicsFee.MaxGasLimitPerTx()
		if gasConsumedInReceiverShard > maxGasLimitPerTx {
			// this should not happen, transactions with higher gas should have been already rejected
			// return the last saved mini block, and the proper error, without saving the rest of mini blocks
			// do not save any pending mini blocks, as this one is invalid
			return 0, process.ErrMaxGasLimitPerTransactionIsReached
		}

		gasConsumedByMB += gasConsumedInReceiverShard
	}

	maxGasLimitPerMB := gc.maxGasLimitPerMiniBlock(mb.GetReceiverShardID())
	if gasConsumedByMB > maxGasLimitPerMB {
		// return the last saved mini block, and the proper error, without saving the rest of mini blocks
		// do not save any pending mini blocks, as this one is invalid
		return 0, process.ErrMaxGasLimitPerMiniBlockIsReached
	}

	return gasConsumedByMB, nil
}

func (gc *gasConsumption) checkPendingIncomingMiniBlocks() ([]data.MiniBlockHeaderHandler, error) {
	addedMiniBlocks := make([]data.MiniBlockHeaderHandler, 0)
	// checking if any pending mini blocks are left to fill the block
	hasPendingMiniBlocks := len(gc.pendingMiniBlocks) > 0
	if !hasPendingMiniBlocks {
		return addedMiniBlocks, nil
	}

	// won't return error, but don't add them further
	// most probably will never happen
	if gc.incomingLimitFactor == 0 {
		return addedMiniBlocks, nil
	}

	bandwidthForIncomingMiniBlocks := gc.getGasLimitForOneDirection(incoming, gc.shardCoordinator.SelfId())
	bandwidthForIncomingMiniBlocks += gc.getGasLeftFromTransactions()
	lastIndexAdded := 0
	for i := 0; i < len(gc.pendingMiniBlocks); i++ {
		mb := gc.pendingMiniBlocks[i]
		_, err := gc.checkIncomingMiniBlock(mb, gc.transactionsForPendingMiniBlocks, bandwidthForIncomingMiniBlocks)
		if err != nil {
			return nil, err
		}

		addedMiniBlocks = append(addedMiniBlocks, mb)
		lastIndexAdded = i
	}

	gc.pendingMiniBlocks = gc.pendingMiniBlocks[lastIndexAdded+1:]

	return addedMiniBlocks, nil
}

// CheckOutgoingTransactions verifies the outgoing transactions and returns:
//   - the index of the last valid transaction
//   - the pending mini blocks added if any
//   - error if so
//
// only returns error if a transaction is invalid, with too much gas
// This method assumes that incoming mini blocks were already handled, trying to add any remaining pending ones at the end
func (gc *gasConsumption) CheckOutgoingTransactions(
	txHashes [][]byte,
	transactions []data.TransactionHandler,
) (addedTxHashes [][]byte, pendingMiniBlocksAdded []data.MiniBlockHeaderHandler, err error) {
	if len(transactions) == 0 || len(txHashes) == 0 {
		return nil, nil, nil
	}

	if len(transactions) != len(txHashes) {
		return nil, nil, process.ErrInvalidValue
	}

	gc.mut.Lock()
	defer gc.mut.Unlock()

	if gc.outgoingLimitFactor == 0 {
		return nil, nil, fmt.Errorf("%w for outgoing transactions", process.ErrZeroLimit)
	}

	skippedSenders := make(map[string]struct{})
	addedHashes := make([][]byte, 0)
	for i := 0; i < len(transactions); i++ {
		_, shouldSkipSender := skippedSenders[string(transactions[i].GetSndAddr())]
		if shouldSkipSender {
			continue
		}

		shouldSkipSender = gc.checkOutgoingTransaction(transactions[i])
		if shouldSkipSender {
			skippedSenders[string(transactions[i].GetSndAddr())] = struct{}{}
			continue
		}

		addedHashes = append(addedHashes, txHashes[i])
	}

	// reaching this point means that transactions were added and the limit for outgoing was not reached
	pendingMiniBlocksAdded, err = gc.checkPendingIncomingMiniBlocks()
	return addedHashes, pendingMiniBlocksAdded, err
}

// must be called under mutex protection
func (gc *gasConsumption) checkOutgoingTransaction(
	tx data.TransactionHandler,
) bool {
	if check.IfNil(tx) {
		return false
	}

	senderShard := gc.shardCoordinator.SelfId()
	receiverShard := gc.shardCoordinator.ComputeId(tx.GetRcvAddr())
	gasConsumedInSenderShard, gasConsumedInReceiverShard, err := gc.checkGasConsumedByTx(senderShard, receiverShard, tx)
	if err != nil {
		log.Warn("checkOutgoingTransaction.checkGasConsumedByTx failed", "error", err)
		return true
	}

	return gc.checkShardsLimits(senderShard, receiverShard, gasConsumedInSenderShard, gasConsumedInReceiverShard)
}

func (gc *gasConsumption) checkGasConsumedByTx(
	senderShard uint32,
	receiverShard uint32,
	tx data.TransactionHandler,
) (uint64, uint64, error) {
	gasConsumedInSenderShard, gasConsumedInReceiverShard, err := gc.gasHandler.ComputeGasProvidedByTx(senderShard, receiverShard, tx)
	if err != nil {
		return 0, 0, err
	}
	maxGasLimitPerTx := gc.economicsFee.MaxGasLimitPerTx()
	if gasConsumedInSenderShard > maxGasLimitPerTx || gasConsumedInReceiverShard > maxGasLimitPerTx {
		// this should not happen, transactions with higher gas should have been already rejected
		// return the last saved transaction, and the proper error, without saving the rest of transactions
		return 0, 0, process.ErrMaxGasLimitPerTransactionIsReached
	}

	return gasConsumedInSenderShard, gasConsumedInReceiverShard, nil
}

func (gc *gasConsumption) checkShardsLimits(
	senderShard uint32,
	receiverShard uint32,
	gasConsumedInSenderShard uint64,
	gasConsumedInReceiverShard uint64,
) bool {
	isCrossShard := senderShard != receiverShard

	bandwidthForOutgoingTransactionsIntra := gc.getGasLimitForOneDirection(outgoingIntra, senderShard)
	// if mini blocks are already handled, use the space left only for intra shard limit
	bandwidthForOutgoingTransactionsIntra += gc.getGasLeftFromMiniBlocks(senderShard)

	txsLimitReachedIntra := gc.totalGasConsumed[outgoingIntra]+gasConsumedInSenderShard > bandwidthForOutgoingTransactionsIntra
	txsLimitReachedCross := false
	if isCrossShard {
		bandwidthForOutgoingCrossTransactions := gc.getGasLimitForOneDirection(outgoingCross, receiverShard)
		txsLimitReachedCross = gc.totalGasConsumed[outgoingCross]+gasConsumedInReceiverShard > bandwidthForOutgoingCrossTransactions
	}

	// if one of the limits is reached, do not count the remaining gas
	// also return true so sender will be skipped
	if txsLimitReachedCross || txsLimitReachedIntra {
		return true
	}

	// add the consumed gas, for receiver shard too if needed
	gc.totalGasConsumed[outgoingIntra] += gasConsumedInSenderShard
	if isCrossShard {
		gc.totalGasConsumed[outgoingCross] += gasConsumedInReceiverShard
	}

	return false
}

func (gc *gasConsumption) getGasLeftFromMiniBlocks(senderShard uint32) uint64 {
	bandwidthForIncomingMiniBlocks := gc.getGasLimitForOneDirection(incoming, senderShard)
	gasConsumedByIncomingMiniBlocks := gc.totalGasConsumed[incoming]
	if gasConsumedByIncomingMiniBlocks >= bandwidthForIncomingMiniBlocks {
		return 0
	}

	return bandwidthForIncomingMiniBlocks - gasConsumedByIncomingMiniBlocks
}

func (gc *gasConsumption) getGasLeftFromTransactions() uint64 {
	bandwidthForOutgoingIntra := gc.getGasLimitForOneDirection(outgoingIntra, gc.shardCoordinator.SelfId())
	gasConsumedByOutgoingIntra := gc.totalGasConsumed[outgoingIntra]
	if gasConsumedByOutgoingIntra >= bandwidthForOutgoingIntra {
		return 0
	}

	return bandwidthForOutgoingIntra - gasConsumedByOutgoingIntra
}

// GetBandwidthForTransactions returns the total bandwidth left for transactions after mini blocks selection
func (gc *gasConsumption) GetBandwidthForTransactions() uint64 {
	gc.mut.RLock()
	defer gc.mut.RUnlock()

	gasLeftFromMiniBlocks := gc.getGasLeftFromMiniBlocks(gc.shardCoordinator.SelfId())
	gasLeftFromTransactions := gc.getGasLeftFromTransactions()
	return gasLeftFromMiniBlocks + gasLeftFromTransactions
}

// TotalGasConsumed returns the total gas consumed for both incoming and outgoing transactions
func (gc *gasConsumption) TotalGasConsumed() uint64 {
	gc.mut.RLock()
	defer gc.mut.RUnlock()

	totalGasConsumed := uint64(0)
	for _, gasConsumed := range gc.totalGasConsumed {
		totalGasConsumed += gasConsumed
	}

	return totalGasConsumed
}

// DecreaseIncomingLimit reduces the gas limit for incoming mini blocks by a configured percentage
func (gc *gasConsumption) DecreaseIncomingLimit() {
	gc.mut.Lock()
	defer gc.mut.Unlock()

	if gc.incomingLimitFactor == zeroLimitsFactor {
		return
	}

	if gc.incomingLimitFactor <= zeroLimitsFactor ||
		gc.incomingLimitFactor <= gc.decreaseStep {
		gc.incomingLimitFactor = zeroLimitsFactor
		return
	}

	gc.incomingLimitFactor = gc.incomingLimitFactor - gc.decreaseStep
}

// DecreaseOutgoingLimit reduces the gas limit for outgoing transactions by a configured percentage
func (gc *gasConsumption) DecreaseOutgoingLimit() {
	gc.mut.Lock()
	defer gc.mut.Unlock()

	if gc.outgoingLimitFactor == zeroLimitsFactor {
		return
	}

	if gc.outgoingLimitFactor <= zeroLimitsFactor ||
		gc.outgoingLimitFactor <= gc.decreaseStep {
		gc.outgoingLimitFactor = zeroLimitsFactor
		return
	}

	gc.outgoingLimitFactor = gc.outgoingLimitFactor - gc.decreaseStep
}

// ZeroIncomingLimit sets the incoming limit factor to 0, effectively disabling incoming mini blocks
func (gc *gasConsumption) ZeroIncomingLimit() {
	gc.mut.Lock()
	defer gc.mut.Unlock()

	log.Debug("setting incoming limit to zero...")

	gc.incomingLimitFactor = 0
}

// ZeroOutgoingLimit sets the outgoing limit factor to 0, effectively disabling outgoing transactions
func (gc *gasConsumption) ZeroOutgoingLimit() {
	gc.mut.Lock()
	defer gc.mut.Unlock()

	log.Debug("setting outgoing limit to zero...")

	gc.outgoingLimitFactor = 0
}

// ResetIncomingLimit resets the gas limit for incoming mini blocks to its initial value
func (gc *gasConsumption) ResetIncomingLimit() {
	gc.mut.Lock()
	defer gc.mut.Unlock()

	gc.incomingLimitFactor = gc.initialOverestimationFactor
}

// ResetOutgoingLimit resets the gas limit for outgoing transactions to its initial value
func (gc *gasConsumption) ResetOutgoingLimit() {
	gc.mut.Lock()
	defer gc.mut.Unlock()

	gc.outgoingLimitFactor = gc.initialOverestimationFactor
}

// Reset clears all gas consumption data except for the limits factors
func (gc *gasConsumption) Reset() {
	gc.mut.Lock()
	defer gc.mut.Unlock()

	gc.totalGasConsumed = make(map[gasType]uint64)
	gc.gasConsumedByMiniBlock = make(map[string]uint64)
	gc.pendingMiniBlocks = make([]data.MiniBlockHeaderHandler, 0)
	gc.transactionsForPendingMiniBlocks = make(map[string][]data.TransactionHandler, 0)
}

// GetPendingMiniBlocks returns the pending mini blocks
func (gc *gasConsumption) GetPendingMiniBlocks() []data.MiniBlockHeaderHandler {
	gc.mut.RLock()
	defer gc.mut.RUnlock()

	pendingMbs := make([]data.MiniBlockHeaderHandler, len(gc.pendingMiniBlocks))
	copy(pendingMbs, gc.pendingMiniBlocks)

	return pendingMbs
}

func (gc *gasConsumption) getGasLimitForOneDirection(gasType gasType, shardID uint32) uint64 {
	totalBlockLimit := gc.maxGasLimitPerBlock(gasType, shardID)
	return totalBlockLimit * percentSplitBlock / 100
}

// must be called under mutex protection as it access blockLimitFactor
func (gc *gasConsumption) maxGasLimitPerBlock(gasType gasType, shardID uint32) uint64 {
	limitFactor := gc.incomingLimitFactor
	if gasType == outgoingIntra || gasType == outgoingCross {
		limitFactor = gc.outgoingLimitFactor
	}

	isCrossShard := shardID != gc.shardCoordinator.SelfId()
	if isCrossShard {
		return gc.economicsFee.MaxGasLimitPerBlockForSafeCrossShard() * limitFactor / 100
	}

	return gc.economicsFee.MaxGasLimitPerBlock(shardID) * limitFactor / 100
}

// must be called under mutex protection as it access miniBlockLimitFactor
func (gc *gasConsumption) maxGasLimitPerMiniBlock(shardID uint32) uint64 {
	isCrossShard := shardID != gc.shardCoordinator.SelfId()
	if isCrossShard {
		return gc.economicsFee.MaxGasLimitPerMiniBlockForSafeCrossShard()
	}

	return gc.economicsFee.MaxGasLimitPerMiniBlock(shardID)
}

// IsInterfaceNil checks if the interface is nil
func (gc *gasConsumption) IsInterfaceNil() bool {
	return gc == nil
}
