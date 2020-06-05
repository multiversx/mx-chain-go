package broadcast

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// HeaderDataForValidator -
type HeaderDataForValidator struct {
	Round        uint64
	PrevRandSeed []byte
}

// SignMessage will sign and return the given message
func (cm *commonMessenger) SignMessage(message *consensus.Message) ([]byte, error) {
	return cm.signMessage(message)
}

// CreateDelayBroadcastDataForValidator creates the delayed broadcast data
func CreateDelayBroadcastDataForValidator(
	headerHash []byte,
	prevRandSeed []byte,
	round uint64,
	miniblocksData map[uint32][]byte,
	miniBlockHashes map[uint32]map[string]struct{},
	transactionsData map[string][][]byte,
	order uint32,
) *delayedBroadcastData {
	return &delayedBroadcastData{
		headerHash:      headerHash,
		prevRandSeed:    prevRandSeed,
		round:           round,
		miniBlocksData:  miniblocksData,
		miniBlockHashes: miniBlockHashes,
		transactions:    transactionsData,
		order:           order,
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
func (dbb *delayedBlockBroadcaster) RegisterInterceptorCallback(cb func(toShard uint32, data []byte)) error {
	return dbb.registerInterceptorCallback(cb)
}

// InterceptedMiniBlockData -
func (dbb *delayedBlockBroadcaster) InterceptedMiniBlockData(toShardTopic uint32, data []byte) {
	dbb.interceptedMiniBlockData(toShardTopic, data)
}

// NewCommonMessenger will return a new instance of a commonMessenger
func NewCommonMessenger(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	privateKey crypto.PrivateKey,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
) (*commonMessenger, error) {

	return &commonMessenger{
		marshalizer:      marshalizer,
		messenger:        messenger,
		privateKey:       privateKey,
		shardCoordinator: shardCoordinator,
		singleSigner:     singleSigner,
	}, nil
}
