package unsigned

import (
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.DefaultLogger()

// UnsignedTxInterceptor is used for intercepting unsigned transaction and storing them into a datapool
type UnsignedTxInterceptor struct {
	marshalizer      marshal.Marshalizer
	uTxPool          dataRetriever.ShardedDataCacherNotifier
	uTxStorer        storage.Storer
	addrConverter    state.AddressConverter
	hasher           hashing.Hasher
	shardCoordinator sharding.Coordinator
}

// NewUnsignedTxInterceptor hooks a new interceptor for unsigned transactions
func NewUnsignedTxInterceptor(
	marshalizer marshal.Marshalizer,
	uTxPool dataRetriever.ShardedDataCacherNotifier,
	uTxStorer storage.Storer,
	addrConverter state.AddressConverter,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
) (*UnsignedTxInterceptor, error) {

	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, process.ErrNilMarshalizer
	}
	if uTxPool == nil || uTxPool.IsInterfaceNil() {
		return nil, process.ErrNilUTxDataPool
	}
	if uTxStorer == nil || uTxStorer.IsInterfaceNil() {
		return nil, process.ErrNilUTxStorage
	}
	if addrConverter == nil || addrConverter.IsInterfaceNil() {
		return nil, process.ErrNilAddressConverter
	}
	if hasher == nil || hasher.IsInterfaceNil() {
		return nil, process.ErrNilHasher
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return nil, process.ErrNilShardCoordinator
	}

	uTxIntercept := &UnsignedTxInterceptor{
		marshalizer:      marshalizer,
		uTxPool:          uTxPool,
		uTxStorer:        uTxStorer,
		hasher:           hasher,
		addrConverter:    addrConverter,
		shardCoordinator: shardCoordinator,
	}

	return uTxIntercept, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (utxi *UnsignedTxInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error {
	if message == nil || message.IsInterfaceNil() {
		return process.ErrNilMessage
	}

	if message.Data() == nil {
		return process.ErrNilDataToProcess
	}

	uTxsBuff := make([][]byte, 0)
	err := utxi.marshalizer.Unmarshal(&uTxsBuff, message.Data())
	if err != nil {
		return err
	}
	if len(uTxsBuff) == 0 {
		return process.ErrNoUnsignedTransactionInMessage
	}

	filteredUTxsBuffs := make([][]byte, 0)
	lastErrEncountered := error(nil)
	for _, uTxBuff := range uTxsBuff {
		uTxIntercepted, err := NewInterceptedUnsignedTransaction(
			uTxBuff,
			utxi.marshalizer,
			utxi.hasher,
			utxi.addrConverter,
			utxi.shardCoordinator)

		if err != nil {
			lastErrEncountered = err
			continue
		}

		//utx is validated, add it to filtered out utxs
		filteredUTxsBuffs = append(filteredUTxsBuffs, uTxBuff)
		if uTxIntercepted.IsAddressedToOtherShards() {
			log.Debug("intercepted utx is for other shards")

			continue
		}

		go utxi.processUnsignedTransaction(uTxIntercepted)
	}

	var buffToSend []byte
	filteredOutUTxsNeedToBeSend := len(filteredUTxsBuffs) > 0 && lastErrEncountered != nil
	if filteredOutUTxsNeedToBeSend {
		buffToSend, err = utxi.marshalizer.Marshal(filteredUTxsBuffs)
		if err != nil {
			return err
		}
	}

	if broadcastHandler != nil {
		broadcastHandler(buffToSend)
	}

	return lastErrEncountered
}

func (utxi *UnsignedTxInterceptor) processUnsignedTransaction(uTx *InterceptedUnsignedTransaction) {
	//TODO should remove this as it is expensive
	err := utxi.uTxStorer.Has(uTx.Hash())
	isUTxInStorage := err == nil
	if isUTxInStorage {
		log.Debug("intercepted uTx already processed")
		return
	}

	cacherIdentifier := process.ShardCacherIdentifier(uTx.SndShard(), uTx.RcvShard())
	utxi.uTxPool.AddData(
		uTx.Hash(),
		uTx.UnsignedTransaction(),
		cacherIdentifier,
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (utxi *UnsignedTxInterceptor) IsInterfaceNil() bool {
	if utxi == nil {
		return true
	}
	return false
}
