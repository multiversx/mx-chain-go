package broadcast

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/sharding"
)

// HeaderDataForValidator -
type HeaderDataForValidator struct {
	Round        uint64
	PrevRandSeed []byte
}

// ExtractMetaMiniBlocksAndTransactions -
func (cm *commonMessenger) ExtractMetaMiniBlocksAndTransactions(
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
) (map[uint32][]byte, map[string][][]byte) {
	return cm.extractMetaMiniBlocksAndTransactions(miniBlocks, transactions)
}

// CreateDelayBroadcastDataForValidator creates the delayed broadcast data
func CreateDelayBroadcastDataForValidator(
	headerHash []byte,
	header data.HeaderHandler,
	miniblocksData map[uint32][]byte,
	miniBlockHashes map[string]map[string]struct{},
	transactionsData map[string][][]byte,
	order uint32,
) *delayedBroadcastData {
	return &delayedBroadcastData{
		headerHash:      headerHash,
		header:          header,
		miniBlocksData:  miniblocksData,
		miniBlockHashes: miniBlockHashes,
		transactions:    transactionsData,
		order:           order,
	}
}

// CreateValidatorHeaderBroadcastData creates a validatorHeaderBroadcastData object from the given parameters
func CreateValidatorHeaderBroadcastData(
	headerHash []byte,
	header data.HeaderHandler,
	metaMiniBlocksData map[uint32][]byte,
	metaTransactionsData map[string][][]byte,
	order uint32,
) *validatorHeaderBroadcastData {
	return &validatorHeaderBroadcastData{
		headerHash:           headerHash,
		header:               header,
		metaMiniBlocksData:   metaMiniBlocksData,
		metaTransactionsData: metaTransactionsData,
		order:                order,
	}
}

// CreateDelayBroadcastDataForLeader -
func CreateDelayBroadcastDataForLeader(
	headerHash []byte,
	miniblocks map[uint32][]byte,
	transactions map[string][][]byte,
) *delayedBroadcastData {
	return &delayedBroadcastData{
		headerHash:     headerHash,
		miniBlocksData: miniblocks,
		transactions:   transactions,
	}
}

// HeaderReceived is the callback registered by the shard chain messenger
// to be called when a header is added to the headers pool
func (dbb *delayedBlockBroadcaster) HeaderReceived(headerHandler data.HeaderHandler, hash []byte) {
	dbb.headerReceived(headerHandler, hash)
}

// GetValidatorBroadcastData returns the set validator delayed broadcast data
func (dbb *delayedBlockBroadcaster) GetValidatorBroadcastData() []*delayedBroadcastData {
	dbb.mutDataForBroadcast.RLock()
	copyValBroadcastData := make([]*delayedBroadcastData, len(dbb.valBroadcastData))
	copy(copyValBroadcastData, dbb.valBroadcastData)
	dbb.mutDataForBroadcast.RUnlock()

	return copyValBroadcastData
}

// GetValidatorHeaderBroadcastData -
func (dbb *delayedBlockBroadcaster) GetValidatorHeaderBroadcastData() []*validatorHeaderBroadcastData {
	dbb.mutDataForBroadcast.RLock()
	copyValHeaderBroadcastData := make([]*validatorHeaderBroadcastData, len(dbb.valHeaderBroadcastData))
	copy(copyValHeaderBroadcastData, dbb.valHeaderBroadcastData)
	dbb.mutDataForBroadcast.RUnlock()

	return copyValHeaderBroadcastData
}

// GetLeaderBroadcastData returns the set leader delayed broadcast data
func (dbb *delayedBlockBroadcaster) GetLeaderBroadcastData() []*delayedBroadcastData {
	dbb.mutDataForBroadcast.RLock()
	copyDelayBroadcastData := make([]*delayedBroadcastData, len(dbb.delayedBroadcastData))
	copy(copyDelayBroadcastData, dbb.delayedBroadcastData)
	dbb.mutDataForBroadcast.RUnlock()

	return copyDelayBroadcastData
}

// ValidatorDelayPerOrder -
func ValidatorDelayPerOrder() time.Duration {
	return validatorDelayPerOrder
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

// NewCommonMessenger will return a new instance of a commonMessenger
func NewCommonMessenger(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	shardCoordinator sharding.Coordinator,
	peerSigHandler crypto.PeerSignatureHandler,
	keysHandler consensus.KeysHandler,
) (*commonMessenger, error) {

	return &commonMessenger{
		marshalizer:          marshalizer,
		messenger:            messenger,
		shardCoordinator:     shardCoordinator,
		peerSignatureHandler: peerSigHandler,
		keysHandler:          keysHandler,
	}, nil
}

// Broadcast -
func (cm *commonMessenger) Broadcast(topic string, data []byte, pkBytes []byte) {
	cm.broadcast(topic, data, pkBytes)
}
