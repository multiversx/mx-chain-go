package broadcast

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
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

// HeaderReceived is the callback registered by the shard chain messenger
// to be called when a header is added to the headers pool
func (scm *shardChainMessenger) HeaderReceived(headerHandler data.HeaderHandler, hash []byte) {
	scm.headerReceived(headerHandler, hash)
}

// GetValidatorBroadcastData returns the set validatr delayed broadcast data
func (scm *shardChainMessenger) GetValidatorBroadcastData() []*delayedBroadcastData {
	return scm.valBroadcastData
}

// ValidatorDelayPerOrder -
func ValidatorDelayPerOrder() time.Duration {
	return validatorDelayPerOrder
}

// ScheduleValidatorBroadcast -
func (scm *shardChainMessenger) ScheduleValidatorBroadcast(dataForValidators []*HeaderDataForValidator) {
	dfv := make([]*headerDataForValidator, len(dataForValidators))
	for i, d := range dataForValidators {
		convDfv := &headerDataForValidator{
			round:        d.Round,
			prevRandSeed: d.PrevRandSeed,
		}
		dfv[i] = convDfv
	}
	scm.scheduleValidatorBroadcast(dfv)
}

// AlarmExpired -
func (scm *shardChainMessenger) AlarmExpired(headerHash string) {
	scm.alarmExpired(headerHash)
}

// GetShardDataFromMetaChainBlock -
func (scm *shardChainMessenger) GetShardDataFromMetaChainBlock(
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
func (scm *shardChainMessenger) RegisterInterceptorCallback(cb func(toShard uint32, data []byte)) error {
	return scm.registerInterceptorCallback(cb)
}

// NewCommonMessenger will return a new instance of a commonMessenger
func NewCommonMessenger(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	privateKey crypto.PrivateKey,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
	interceptorsContainer process.InterceptorsContainer,
) (*commonMessenger, error) {

	return &commonMessenger{
		marshalizer:           marshalizer,
		messenger:             messenger,
		privateKey:            privateKey,
		shardCoordinator:      shardCoordinator,
		singleSigner:          singleSigner,
		interceptorsContainer: interceptorsContainer,
	}, nil
}
