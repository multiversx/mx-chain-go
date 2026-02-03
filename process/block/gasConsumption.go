package block

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"golang.org/x/exp/slices"

	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/preprocess"
)

const (
	incoming = "incoming"
	outgoing = "outgoing"
	pending  = "pending"
)

const (
	zeroLimitsFactor       = uint64(0)
	minPercentLimitsFactor = uint64(10)  // 10%
	maxPercentLimitsFactor = uint64(500) // 500%
	percentSplitBlock      = uint64(50)  // 50%
	initialLastIndex       = -1
)

// TODO: rename to include size check

// ArgsGasConsumption holds the arguments needed to create a gasConsumption instance
type ArgsGasConsumption struct {
	EconomicsFee                      process.FeeHandler
	ShardCoordinator                  process.ShardCoordinator
	GasHandler                        process.GasHandler
	BlockCapacityOverestimationFactor uint64
	PercentDecreaseLimitsStep         uint64
	BlockSizeComputation              preprocess.BlockSizeComputationHandler
}

// gasConsumption implements the GasComputation interface for managing gas limits during block creation
type gasConsumption struct {
	mut                              sync.RWMutex
	economicsFee                     process.FeeHandler
	shardCoordinator                 process.ShardCoordinator
	gasHandler                       process.GasHandler
	blockSizeComputation             preprocess.BlockSizeComputationHandler
	totalGasConsumed                 map[string]uint64
	gasConsumedByMiniBlock           map[string]uint64
	numTxsPerMiniBlock               map[string]uint32
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
	if check.IfNil(args.BlockSizeComputation) {
		return nil, process.ErrNilBlockSizeComputationHandler
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
		blockSizeComputation:             args.BlockSizeComputation,
		totalGasConsumed:                 make(map[string]uint64),
		gasConsumedByMiniBlock:           make(map[string]uint64),
		numTxsPerMiniBlock:               make(map[string]uint32),
		transactionsForPendingMiniBlocks: make(map[string][]data.TransactionHandler),
		incomingLimitFactor:              args.BlockCapacityOverestimationFactor,
		outgoingLimitFactor:              args.BlockCapacityOverestimationFactor,
		decreaseStep:                     args.BlockCapacityOverestimationFactor * args.PercentDecreaseLimitsStep / 100,
		percentDecreaseLimitsStep:        args.PercentDecreaseLimitsStep,
		initialOverestimationFactor:      args.BlockCapacityOverestimationFactor,
	}, nil
}

// AddIncomingMiniBlocks verifies if an incoming mini block and its transactions can be included within gas limits.
// returns the last mini block index included, the number of pending mini blocks left and error if needed
// This must be called first, before AddOutgoingTransactions!
func (gc *gasConsumption) AddIncomingMiniBlocks(
	miniBlocks []data.MiniBlockHeaderHandler,
	transactions map[string][]data.TransactionHandler,
) (lastMiniBlockIndex int, pendingMiniBlocks int, err error) {
	if len(miniBlocks) == 0 || len(transactions) == 0 {
		return initialLastIndex, 0, nil
	}

	gc.mut.Lock()
	defer gc.mut.Unlock()

	if gc.incomingLimitFactor == zeroLimitsFactor {
		return initialLastIndex, 0, nil
	}

	// if we already have some pending mini blocks, the new ones should only be appended as pending
	if len(gc.pendingMiniBlocks) > 0 {
		errSavePending := gc.savePendingMiniBlocksNoLock(miniBlocks, transactions)
		if errSavePending != nil {
			return initialLastIndex, 0, errSavePending
		}

		return initialLastIndex, len(miniBlocks), nil
	}

	bandwidthForIncomingMiniBlocks := gc.getGasLimitForOneDirection(incoming, gc.shardCoordinator.SelfId())

	lastMiniBlockIndex = initialLastIndex
	shouldSavePending := false
	for i := 0; i < len(miniBlocks); i++ {
		mbType := miniBlocks[i].GetTypeInt32()
		if mbType == int32(block.RewardsBlock) || mbType == int32(block.PeerBlock) {
			// rewards and validator info have 0 gas limit, thus they should be included anyway without checking their transactions
			gc.addMiniBlockToBlockSizeComputation(miniBlocks[i])
			lastMiniBlockIndex = i
			continue
		}

		shouldSavePending, err = gc.addIncomingMiniBlock(miniBlocks[i], transactions, bandwidthForIncomingMiniBlocks)
		if shouldSavePending {
			// saving pending starting with idx i, as it was not included either
			errSavePending := gc.savePendingMiniBlocksNoLock(miniBlocks[i:], transactions)
			if errSavePending != nil {
				return lastMiniBlockIndex, 0, errSavePending
			}

			return lastMiniBlockIndex, len(gc.pendingMiniBlocks), err
		}
		if err != nil {
			return lastMiniBlockIndex, 0, err
		}

		lastMiniBlockIndex = i
	}

	return lastMiniBlockIndex, 0, nil
}

func (gc *gasConsumption) addMiniBlockToBlockSizeComputation(
	mb data.MiniBlockHeaderHandler,
) {
	gc.blockSizeComputation.AddNumMiniBlocks(1)
	gc.blockSizeComputation.AddNumTxs(int(mb.GetTxCount()))

	gc.numTxsPerMiniBlock[string(mb.GetHash())] = mb.GetTxCount()
}

func (gc *gasConsumption) savePendingMiniBlocksNoLock(
	miniBlocks []data.MiniBlockHeaderHandler,
	transactions map[string][]data.TransactionHandler,
) error {
	for _, mb := range miniBlocks {
		hashStr := string(mb.GetHash())
		transactionsForMB, found := transactions[hashStr]
		if !found {
			log.Warn("could not find transaction for pending mini block", "hash", mb.GetHash())
			return fmt.Errorf("%w, could not find mini block hash in transactions map", process.ErrInvalidValue)
		}
		gasConsumedByPendingMb, errCheckPending := gc.checkGasConsumedByMiniBlock(mb, transactionsForMB)
		if errCheckPending != nil {
			log.Warn("failed to check gas consumed by pending mini block", "hash", mb.GetHash(), "error", errCheckPending)
			return errCheckPending
		}

		gc.transactionsForPendingMiniBlocks[hashStr] = transactions[hashStr]
		gc.totalGasConsumed[pending] += gasConsumedByPendingMb
	}

	gc.pendingMiniBlocks = append(gc.pendingMiniBlocks, miniBlocks...)

	return nil
}

// RevertIncomingMiniBlocks gets a list of mini block hashes and removes them from the local state
func (gc *gasConsumption) RevertIncomingMiniBlocks(miniBlockHashes [][]byte) {
	if len(miniBlockHashes) == 0 {
		return
	}

	gc.mut.Lock()
	defer gc.mut.Unlock()

	for _, miniBlockHash := range miniBlockHashes {
		// do not check here if it was found or not, as some pending mini blocks may be missing from this map
		gasConsumedByMb := gc.gasConsumedByMiniBlock[string(miniBlockHash)]
		delete(gc.gasConsumedByMiniBlock, string(miniBlockHash))

		isPending, idxInPendingSlice := gc.isPendingMiniBlock(miniBlockHash)
		if isPending {
			gc.revertPendingMiniBlock(miniBlockHash, idxInPendingSlice)
			gc.totalGasConsumed[pending] -= gasConsumedByMb
			continue
		}

		gc.revertBlockSizeLimits(miniBlockHash)

		// if the mini block is not pending, remove it from the total gas consumed
		gc.totalGasConsumed[incoming] -= gasConsumedByMb
	}
}

func (gc *gasConsumption) revertBlockSizeLimits(
	miniBlockHash []byte,
) {
	gc.blockSizeComputation.DecNumMiniBlocks(1)

	numTxsPerMiniBlock := gc.numTxsPerMiniBlock[string(miniBlockHash)]
	delete(gc.numTxsPerMiniBlock, string(miniBlockHash))
	gc.blockSizeComputation.DecNumTxs(int(numTxsPerMiniBlock))
}

func (gc *gasConsumption) isPendingMiniBlock(blockHash []byte) (bool, int) {
	for idx, miniBlock := range gc.pendingMiniBlocks {
		if bytes.Equal(miniBlock.GetHash(), blockHash) {
			return true, idx
		}
	}

	return false, initialLastIndex
}

func (gc *gasConsumption) revertPendingMiniBlock(miniBlockHash []byte, idxInPendingSlice int) {
	// if the mini block was saved as pending, remove its transactions
	// and remove it from pending slice
	delete(gc.transactionsForPendingMiniBlocks, string(miniBlockHash))
	gc.pendingMiniBlocks = slices.Delete(gc.pendingMiniBlocks, idxInPendingSlice, idxInPendingSlice+1)
}

func (gc *gasConsumption) addIncomingMiniBlock(
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

	gc.gasConsumedByMiniBlock[string(mbHash)] = gasConsumedByMB

	mbsGasLimitReached := gc.totalGasConsumed[incoming]+gasConsumedByMB > bandwidthForIncomingMiniBlocks
	mbsSizeLimitReached := gc.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(1, int(numTxs))
	mbsLimitReached := mbsGasLimitReached || mbsSizeLimitReached

	if !mbsLimitReached {
		// limit not reached, continue
		// this method might be called either from handling all mini blocks,
		// either from handling pending, where the pending ones
		// should have continuous indexes after the ones already included
		gc.totalGasConsumed[incoming] += gasConsumedByMB

		gc.addMiniBlockToBlockSizeComputation(mb)

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
		_, gasConsumedInReceiverShard, err := gc.gasHandler.ComputeGasProvidedByTx(mb.GetSenderShardID(), mb.GetReceiverShardID(), tx)
		if err != nil {
			// do not save any pending mini blocks, as this one is invalid
			return 0, err
		}

		maxGasLimitPerTx := gc.economicsFee.MaxGasLimitPerTx()
		if gasConsumedInReceiverShard > maxGasLimitPerTx || tx.GetGasLimit() > maxGasLimitPerTx {
			// this should not happen, transactions with higher gas should have been already rejected
			// return the last saved mini block, and the proper error, without saving the rest of mini blocks
			// do not save any pending mini blocks, as this one is invalid
			return 0, process.ErrMaxGasLimitPerTransactionIsReached
		}

		gasConsumedByMB += gasConsumedInReceiverShard
	}

	maxGasLimitPerMB := gc.maxGasLimitPerMiniBlock(mb.GetReceiverShardID())
	if gasConsumedByMB <= maxGasLimitPerMB {
		return gasConsumedByMB, nil
	}

	// if the limit for mini block is reached and:
	//	- there is only one tx that satisfied the MaxGasLimitPerTx, allow it into the block
	//	- there are more than one tx, return error
	if len(transactionsForMB) == 1 {
		return gasConsumedByMB, nil
	}

	return 0, process.ErrMaxGasLimitPerMiniBlockIsReached
}

func (gc *gasConsumption) addPendingIncomingMiniBlocks() ([]data.MiniBlockHeaderHandler, error) {
	addedMiniBlocks := make([]data.MiniBlockHeaderHandler, 0)
	// checking if any pending mini blocks are left to fill the block
	hasPendingMiniBlocks := len(gc.pendingMiniBlocks) > 0
	if !hasPendingMiniBlocks {
		return addedMiniBlocks, nil
	}

	// won't return error, but don't add them further
	// most probably will never happen
	if gc.incomingLimitFactor == zeroLimitsFactor {
		return addedMiniBlocks, nil
	}

	bandwidthForIncomingMiniBlocks := gc.getGasLimitForOneDirection(incoming, gc.shardCoordinator.SelfId())
	bandwidthForIncomingMiniBlocks += gc.getGasLeftFromTransactions()
	lastIndexAdded := 0
	for i := 0; i < len(gc.pendingMiniBlocks); i++ {
		mb := gc.pendingMiniBlocks[i]
		gasConsumedByIncomingBefore := gc.totalGasConsumed[incoming]
		shouldSavePending, err := gc.addIncomingMiniBlock(mb, gc.transactionsForPendingMiniBlocks, bandwidthForIncomingMiniBlocks)
		if err != nil {
			return nil, err
		}
		// if should save pending, it means the limit was reached, thus break the loop
		if shouldSavePending {
			break
		}

		addedMiniBlocks = append(addedMiniBlocks, mb)
		lastIndexAdded = i

		gasConsumedByIncomingAfter := gc.totalGasConsumed[incoming]
		gasConsumedByMiniBlock := gasConsumedByIncomingAfter - gasConsumedByIncomingBefore
		if gasConsumedByMiniBlock <= gc.totalGasConsumed[pending] {
			gc.totalGasConsumed[pending] -= gasConsumedByMiniBlock
		}
	}

	gc.pendingMiniBlocks = gc.pendingMiniBlocks[lastIndexAdded+1:]

	return addedMiniBlocks, nil
}

// AddOutgoingTransactions verifies the outgoing transactions and returns:
//   - the index of the last valid transaction
//   - the pending mini blocks added if any
//   - error if so
//
// only returns error if a transaction is invalid, with too much gas
// This method assumes that incoming mini blocks were already handled, trying to add any remaining pending ones at the end
func (gc *gasConsumption) AddOutgoingTransactions(
	txHashes [][]byte,
	transactions []data.TransactionHandler,
) (addedTxHashes [][]byte, pendingMiniBlocksAdded []data.MiniBlockHeaderHandler, err error) {
	if len(transactions) != len(txHashes) {
		return nil, nil, process.ErrInvalidValue
	}

	gc.mut.Lock()
	defer gc.mut.Unlock()

	if gc.outgoingLimitFactor == 0 {
		return make([][]byte, 0), make([]data.MiniBlockHeaderHandler, 0), nil
	}

	skippedSenders := make(map[string]struct{})
	addedHashes := make([][]byte, 0)
	for i := 0; i < len(transactions); i++ {
		_, shouldSkipSender := skippedSenders[string(transactions[i].GetSndAddr())]
		if shouldSkipSender {
			continue
		}

		shouldSkipSender = gc.addOutgoingTransaction(transactions[i])
		if shouldSkipSender {
			skippedSenders[string(transactions[i].GetSndAddr())] = struct{}{}
			continue
		}

		gc.blockSizeComputation.AddNumTxs(1)

		addedHashes = append(addedHashes, txHashes[i])
	}

	if len(addedHashes) > 0 {
		gc.blockSizeComputation.AddNumMiniBlocks(1)
	}

	// reaching this point means that transactions were added and the limit for outgoing was not reached
	pendingMiniBlocksAdded, err = gc.addPendingIncomingMiniBlocks()
	return addedHashes, pendingMiniBlocksAdded, err
}

// must be called under mutex protection
func (gc *gasConsumption) addOutgoingTransaction(
	tx data.TransactionHandler,
) bool {
	if check.IfNil(tx) {
		return false
	}

	senderShard := gc.shardCoordinator.SelfId()
	receiverShard := gc.shardCoordinator.ComputeId(tx.GetRcvAddr())
	gasConsumedInSenderShard, gasConsumedInReceiverShard, err := gc.checkGasConsumedByTx(senderShard, receiverShard, tx)
	if err != nil {
		log.Warn("addOutgoingTransaction.checkGasConsumedByTx failed", "error", err)
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
	if gasConsumedInSenderShard > maxGasLimitPerTx ||
		gasConsumedInReceiverShard > maxGasLimitPerTx ||
		tx.GetGasLimit() > maxGasLimitPerTx {
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

	gasKeyOutgoingIntra := gc.getGasKeyForOutgoingShard(senderShard)
	gasKeyOutgoingCross := gc.getGasKeyForOutgoingShard(receiverShard)
	bandwidthForOutgoingTransactionsIntra := gc.getGasLimitForOneDirection(gasKeyOutgoingIntra, senderShard)
	// if mini blocks are already handled, use the space left only for intra shard limit
	bandwidthForOutgoingTransactionsIntra += gc.getGasLeftFromMiniBlocks(senderShard)

	txsLimitReachedIntra := gc.totalGasConsumed[gasKeyOutgoingIntra]+gasConsumedInSenderShard > bandwidthForOutgoingTransactionsIntra
	txsLimitReachedCross := false
	if isCrossShard {
		bandwidthForOutgoingCrossTransactions := gc.getGasLimitForOneDirection(gasKeyOutgoingCross, receiverShard)
		txsLimitReachedCross = gc.totalGasConsumed[gasKeyOutgoingCross]+gasConsumedInReceiverShard > bandwidthForOutgoingCrossTransactions
	}

	txsSizeLimit := gc.blockSizeComputation.IsMaxBlockSizeWithoutThrottleReached(1, 1)

	// if one of the limits is reached, do not count the remaining gas
	// also return true so sender will be skipped
	if txsLimitReachedCross || txsLimitReachedIntra || txsSizeLimit {
		return true
	}

	// add the consumed gas, for receiver shard too if needed
	gc.totalGasConsumed[gasKeyOutgoingIntra] += gasConsumedInSenderShard
	if isCrossShard {
		gc.totalGasConsumed[gasKeyOutgoingCross] += gasConsumedInReceiverShard
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
	gasKeyOutgoingIntra := gc.getGasKeyForOutgoingShard(gc.shardCoordinator.SelfId())
	bandwidthForOutgoingIntra := gc.getGasLimitForOneDirection(gasKeyOutgoingIntra, gc.shardCoordinator.SelfId())
	gasConsumedByOutgoingIntra := gc.totalGasConsumed[gasKeyOutgoingIntra]
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

// TotalGasConsumedInSelfShard returns the total gas consumed for both incoming and outgoing transactions in self shard
func (gc *gasConsumption) TotalGasConsumedInSelfShard() uint64 {
	gc.mut.RLock()
	defer gc.mut.RUnlock()

	totalGasConsumedInSelfShard := gc.totalGasConsumed[incoming]
	gasKeyOutgoingIntra := gc.getGasKeyForOutgoingShard(gc.shardCoordinator.SelfId())
	totalGasConsumedInSelfShard += gc.totalGasConsumed[gasKeyOutgoingIntra]

	return totalGasConsumedInSelfShard
}

// TotalGasConsumedInShard returns the total gas consumed for outgoing transactions in the provided shard
func (gc *gasConsumption) TotalGasConsumedInShard(shard uint32) uint64 {
	gc.mut.RLock()
	defer gc.mut.RUnlock()

	gasKeyOutgoingCross := gc.getGasKeyForOutgoingShard(shard)
	return gc.totalGasConsumed[gasKeyOutgoingCross]
}

// CanAddPendingIncomingMiniBlocks returns true if more pending incoming mini blocks can be added without reaching the block limits
func (gc *gasConsumption) CanAddPendingIncomingMiniBlocks() bool {
	gc.mut.RLock()
	defer gc.mut.RUnlock()

	totalGasToBeConsumedByPending := uint64(0)
	totalGasConsumed := uint64(0)
	for typeOfGas, gasConsumed := range gc.totalGasConsumed {
		if typeOfGas == pending {
			totalGasToBeConsumedByPending += gasConsumed
			continue
		}

		totalGasConsumed += gasConsumed
	}

	maxGasLimitPerBlock := gc.maxGasLimitPerBlock(incoming, gc.shardCoordinator.SelfId())
	if maxGasLimitPerBlock <= totalGasConsumed {
		return false
	}

	gasLeft := maxGasLimitPerBlock - totalGasConsumed
	return totalGasToBeConsumedByPending < gasLeft
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

	gc.incomingLimitFactor = zeroLimitsFactor
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

	gc.totalGasConsumed = make(map[string]uint64)
	gc.gasConsumedByMiniBlock = make(map[string]uint64)
	gc.numTxsPerMiniBlock = make(map[string]uint32)
	gc.pendingMiniBlocks = make([]data.MiniBlockHeaderHandler, 0)
	gc.transactionsForPendingMiniBlocks = make(map[string][]data.TransactionHandler, 0)

	gc.blockSizeComputation.Init()
}

// GetPendingMiniBlocks returns the pending mini blocks
func (gc *gasConsumption) GetPendingMiniBlocks() []data.MiniBlockHeaderHandler {
	gc.mut.RLock()
	defer gc.mut.RUnlock()

	pendingMbs := make([]data.MiniBlockHeaderHandler, len(gc.pendingMiniBlocks))
	copy(pendingMbs, gc.pendingMiniBlocks)

	return pendingMbs
}

func (gc *gasConsumption) getGasLimitForOneDirection(gasType string, shardID uint32) uint64 {
	totalBlockLimit := gc.maxGasLimitPerBlock(gasType, shardID)
	if shardID != gc.shardCoordinator.SelfId() {
		// allow the full block for receiver shard
		return totalBlockLimit
	}

	return totalBlockLimit * percentSplitBlock / 100
}

// must be called under mutex protection as it access blockLimitFactor
func (gc *gasConsumption) maxGasLimitPerBlock(gasType string, shardID uint32) uint64 {
	limitFactor := gc.incomingLimitFactor
	if strings.Contains(gasType, outgoing) {
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

func (gc *gasConsumption) getGasKeyForOutgoingShard(shard uint32) string {
	return fmt.Sprintf("%s_%d", outgoing, shard)
}

// IsInterfaceNil checks if the interface is nil
func (gc *gasConsumption) IsInterfaceNil() bool {
	return gc == nil
}
