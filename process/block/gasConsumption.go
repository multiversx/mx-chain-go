package block

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

const (
	initialLimitsFactor = uint64(100)
	// TODO: move these to config
	percentSplitBlock         = uint64(50) // 50%
	percentDecreaseLimitsStep = uint64(10) // 10%
)

type gasConsumption struct {
	sync.RWMutex
	economicsFee            process.FeeHandler
	shardCoordinator        sharding.Coordinator
	totalGasConsumedInShard map[uint32]uint64
	gasConsumedByMiniBlock  map[string]uint64
	limitsFactor            uint64
}

// NewGasConsumption returns a new instance of gasConsumption
func NewGasConsumption(
	economicsFee process.FeeHandler,
	shardCoordinator sharding.Coordinator,
) (*gasConsumption, error) {
	if check.IfNil(economicsFee) {
		return nil, process.ErrNilEconomicsData
	}
	if check.IfNil(shardCoordinator) {
		return nil, process.ErrNilShardCoordinator
	}

	return &gasConsumption{
		economicsFee:            economicsFee,
		shardCoordinator:        shardCoordinator,
		totalGasConsumedInShard: make(map[uint32]uint64),
		gasConsumedByMiniBlock:  make(map[string]uint64),
		limitsFactor:            initialLimitsFactor,
	}, nil
}

// CheckTransactionsForMiniBlock checks if the provided transactions can be added in a mini block
// for the current selection session
func (gc *gasConsumption) CheckTransactionsForMiniBlock(
	miniBlockHash []byte,
	txs []data.TransactionHandler,
) (uint32, error) {
	if len(txs) == 0 {
		return 0, nil
	}

	gc.Lock()
	defer gc.Unlock()

	lastIndexAdded := uint32(0)

	senderShard := gc.shardCoordinator.ComputeId(txs[0].GetSndAddr())

	maxGasLimitPerTx := gc.economicsFee.MaxGasLimitPerTx()
	maxGasLimitPerMB := gc.maxGasLimitPerMiniBlock(senderShard)
	gasConsumedByMB := uint64(0)
	for i := uint32(0); i < uint32(len(txs)); i++ {
		tx := txs[i]
		if tx.GetGasLimit() > maxGasLimitPerTx {
			return lastIndexAdded, process.ErrInvalidMaxGasLimitPerTx
		}

		gasConsumedByMB += tx.GetGasLimit()
		if gasConsumedByMB > maxGasLimitPerMB {
			return lastIndexAdded, process.ErrMaxGasLimitPerMiniBlockIsReached
		}

		gasLeftForShard := gc.getMaxGasLeftForShard(senderShard)
		if gasLeftForShard <= 0 {
			return lastIndexAdded, process.ErrMaxGasLimitPerBlockIsReached
		}

		gc.totalGasConsumedInShard[senderShard] += tx.GetGasLimit()
		lastIndexAdded = i
	}

	gc.gasConsumedByMiniBlock[string(miniBlockHash)] = gasConsumedByMB

	return lastIndexAdded, nil
}

// GetGasConsumedByMiniBlock returns the gas consumed by the provided mini block hash
func (gc *gasConsumption) GetGasConsumedByMiniBlock(hash []byte) uint64 {
	gc.RLock()
	defer gc.RUnlock()

	return gc.gasConsumedByMiniBlock[string(hash)]
}

// DecreaseLimits lowers the block limits
func (gc *gasConsumption) DecreaseLimits() {
	gc.Lock()
	defer gc.Unlock()

	decreaseStep := initialLimitsFactor * percentDecreaseLimitsStep / 100
	gc.limitsFactor = gc.limitsFactor - decreaseStep
}

// Reset resets the accumulated values to the initial state
func (gc *gasConsumption) Reset() {
	gc.Lock()
	defer gc.Unlock()

	gc.limitsFactor = initialLimitsFactor
	gc.totalGasConsumedInShard = make(map[uint32]uint64)
	gc.gasConsumedByMiniBlock = make(map[string]uint64)
}

func (gc *gasConsumption) getMaxGasLeftForShard(senderShard uint32) int64 {
	maxGasLimitPerBlock := gc.maxGasLimitPerBlock(senderShard)
	limit := maxGasLimitPerBlock * percentSplitBlock / 100

	if senderShard == gc.shardCoordinator.SelfId() {
		gasConsumed := gc.totalGasConsumedInShard[senderShard]
		return int64(limit) - int64(gasConsumed)
	}

	consumedInOtherShards := int64(0)
	consumedIntraShard := uint64(0)
	for shard, consumed := range gc.totalGasConsumedInShard {
		if shard == gc.shardCoordinator.SelfId() {
			consumedIntraShard = consumed
			continue
		}

		consumedInOtherShards += int64(consumed)
	}

	// if intra shard was already processed,
	// use the remaining unused gas from the intra shard allocation for cross shard
	if consumedIntraShard != 0 && consumedIntraShard < limit {
		gasLeftFromIntra := limit - consumedIntraShard
		limit += gasLeftFromIntra
	}

	return int64(limit) - consumedInOtherShards
}

func (gc *gasConsumption) maxGasLimitPerBlock(shardID uint32) uint64 {
	return gc.economicsFee.MaxGasLimitPerBlock(shardID) * gc.limitsFactor / 100
}

func (gc *gasConsumption) maxGasLimitPerMiniBlock(shardID uint32) uint64 {
	return gc.economicsFee.MaxGasLimitPerMiniBlock(shardID) * gc.limitsFactor / 100
}

// IsInterfaceNil returns true if there is no value under the interface
func (gc *gasConsumption) IsInterfaceNil() bool {
	return gc == nil
}
