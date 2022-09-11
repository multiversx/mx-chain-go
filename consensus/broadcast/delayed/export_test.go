package delayed

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
)

// HeaderDataForValidator -
type HeaderDataForValidator struct {
	Round        uint64
	PrevRandSeed []byte
}

// HeaderReceived is the callback registered by the shard chain messenger
// to be called when a header is added to the headers pool
func (dbb *delayedBlockBroadcaster) HeaderReceived(headerHandler data.HeaderHandler, hash []byte) {
	dbb.headerReceived(headerHandler, hash)
}

// GetValidatorBroadcastData returns the set validator delayed broadcast data
func (dbb *delayedBlockBroadcaster) GetValidatorBroadcastData() []*DelayedBroadcastData {
	dbb.mutDataForBroadcast.RLock()
	copyValBroadcastData := make([]*DelayedBroadcastData, len(dbb.valBroadcastData))
	copy(copyValBroadcastData, dbb.valBroadcastData)
	dbb.mutDataForBroadcast.RUnlock()

	return copyValBroadcastData
}

// GetValidatorHeaderBroadcastData -
func (dbb *delayedBlockBroadcaster) GetValidatorHeaderBroadcastData() []*ValidatorHeaderBroadcastData {
	dbb.mutDataForBroadcast.RLock()
	copyValHeaderBroadcastData := make([]*ValidatorHeaderBroadcastData, len(dbb.valHeaderBroadcastData))
	copy(copyValHeaderBroadcastData, dbb.valHeaderBroadcastData)
	dbb.mutDataForBroadcast.RUnlock()

	return copyValHeaderBroadcastData
}

// GetLeaderBroadcastData returns the set leader delayed broadcast data
func (dbb *delayedBlockBroadcaster) GetLeaderBroadcastData() []*DelayedBroadcastData {
	dbb.mutDataForBroadcast.RLock()
	copyDelayBroadcastData := make([]*DelayedBroadcastData, len(dbb.delayedBroadcastData))
	copy(copyDelayBroadcastData, dbb.delayedBroadcastData)
	dbb.mutDataForBroadcast.RUnlock()

	return copyDelayBroadcastData
}

// ScheduleValidatorBroadcast -
func (dbb *delayedBlockBroadcaster) ScheduleValidatorBroadcast(dataForValidators []*HeaderDataForValidator) {
	dfv := make([]*headerDataForValidator, len(dataForValidators))
	for i, d := range dataForValidators {
		convDfv := &headerDataForValidator{
			round:        d.Round,
			prevRandSeed: d.PrevRandSeed,
		}
		dfv[i] = convDfv
	}
	dbb.scheduleValidatorBroadcast(dfv)
}

// AlarmExpired -
func (dbb *delayedBlockBroadcaster) AlarmExpired(headerHash string) {
	dbb.alarmExpired(headerHash)
}

// RegisterInterceptorCallback -
func (dbb *delayedBlockBroadcaster) RegisterInterceptorCallback(cb func(topic string, hash []byte, data interface{})) error {
	return dbb.registerMiniBlockInterceptorCallback(cb)
}

// InterceptedMiniBlockData -
func (dbb *delayedBlockBroadcaster) InterceptedMiniBlockData(topic string, hash []byte, data interface{}) {
	dbb.interceptedMiniBlockData(topic, hash, data)
}

// InterceptedHeaderData -
func (dbb *delayedBlockBroadcaster) InterceptedHeaderData(topic string, hash []byte, header interface{}) {
	dbb.interceptedHeader(topic, hash, header)
}

// ValidatorDelayPerOrder -
func ValidatorDelayPerOrder() time.Duration {
	return validatorDelayPerOrder
}

// CreateDelayBroadcastDataForValidator creates the delayed broadcast data
func CreateDelayBroadcastDataForValidator(
	headerHash []byte,
	header data.HeaderHandler,
	miniblocksData map[uint32][]byte,
	miniBlockHashes map[string]map[string]struct{},
	transactionsData map[string][][]byte,
	order uint32,
) *DelayedBroadcastData {
	return &DelayedBroadcastData{
		HeaderHash:      headerHash,
		Header:          header,
		MiniBlocksData:  miniblocksData,
		MiniBlockHashes: miniBlockHashes,
		Transactions:    transactionsData,
		Order:           order,
	}
}

// CreateValidatorHeaderBroadcastData creates a validatorHeaderBroadcastData object from the given parameters
func CreateValidatorHeaderBroadcastData(
	headerHash []byte,
	header data.HeaderHandler,
	metaMiniBlocksData map[uint32][]byte,
	metaTransactionsData map[string][][]byte,
	order uint32,
) *ValidatorHeaderBroadcastData {
	return &ValidatorHeaderBroadcastData{
		HeaderHash:           headerHash,
		Header:               header,
		MetaMiniBlocksData:   metaMiniBlocksData,
		MetaTransactionsData: metaTransactionsData,
		Order:                order,
	}
}

// CreateDelayBroadcastDataForLeader -
func CreateDelayBroadcastDataForLeader(
	headerHash []byte,
	miniblocks map[uint32][]byte,
	transactions map[string][][]byte,
) *DelayedBroadcastData {
	return &DelayedBroadcastData{
		HeaderHash:     headerHash,
		MiniBlocksData: miniblocks,
		Transactions:   transactions,
	}
}

// GetShardDataFromMetaChainBlock -
func GetShardDataFromMetaChainBlock(
	headerHandler data.HeaderHandler,
	shardID uint32,
) ([][]byte, []*HeaderDataForValidator, error) {
	headerHashes, dataForValidators, err := getShardDataFromMetaChainBlock(headerHandler, shardID)

	dfv := make([]*HeaderDataForValidator, len(dataForValidators))
	for i, d := range dataForValidators {
		convDfv := &HeaderDataForValidator{
			Round:        d.round,
			PrevRandSeed: d.prevRandSeed,
		}
		dfv[i] = convDfv
	}

	return headerHashes, dfv, err
}
