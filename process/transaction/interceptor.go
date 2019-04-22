package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// TxInterceptor is used for intercepting transaction and storing them into a datapool
type TxInterceptor struct {
	marshalizer              marshal.Marshalizer
	txPool                   dataRetriever.ShardedDataCacherNotifier
	txStorer                 storage.Storer
	addrConverter            state.AddressConverter
	hasher                   hashing.Hasher
	singleSigner             crypto.SingleSigner
	keyGen                   crypto.KeyGenerator
	shardCoordinator         sharding.Coordinator
	broadcastCallbackHandler func(buffToSend []byte)
}

// NewTxInterceptor hooks a new interceptor for transactions
func NewTxInterceptor(
	marshalizer marshal.Marshalizer,
	txPool dataRetriever.ShardedDataCacherNotifier,
	txStorer storage.Storer,
	addrConverter state.AddressConverter,
	hasher hashing.Hasher,
	singleSigner crypto.SingleSigner,
	keyGen crypto.KeyGenerator,
	shardCoordinator sharding.Coordinator,
) (*TxInterceptor, error) {

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if txPool == nil {
		return nil, process.ErrNilTxDataPool
	}
	if txStorer == nil {
		return nil, process.ErrNilTxStorage
	}
	if addrConverter == nil {
		return nil, process.ErrNilAddressConverter
	}
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if singleSigner == nil {
		return nil, process.ErrNilSingleSigner
	}
	if keyGen == nil {
		return nil, process.ErrNilKeyGen
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}

	txIntercept := &TxInterceptor{
		marshalizer:      marshalizer,
		txPool:           txPool,
		txStorer:         txStorer,
		hasher:           hasher,
		addrConverter:    addrConverter,
		singleSigner:     singleSigner,
		keyGen:           keyGen,
		shardCoordinator: shardCoordinator,
	}

	return txIntercept, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (txi *TxInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	if message == nil {
		return process.ErrNilMessage
	}

	if message.Data() == nil {
		return process.ErrNilDataToProcess
	}

	txsBuff := make([][]byte, 0)
	err := txi.marshalizer.Unmarshal(&txsBuff, message.Data())
	if err != nil {
		return err
	}
	if len(txsBuff) == 0 {
		return process.ErrNoTransactionInMessage
	}

	filteredTxsBuffs := make([][]byte, 0)
	lastErrEncountered := error(nil)
	for _, txBuff := range txsBuff {
		txIntercepted, err := NewInterceptedTransaction(
			txBuff,
			txi.marshalizer,
			txi.hasher,
			txi.keyGen,
			txi.singleSigner,
			txi.addrConverter,
			txi.shardCoordinator)

		if err != nil {
			lastErrEncountered = err
			continue
		}

		//tx is validated, add it to filtered out txs
		filteredTxsBuffs = append(filteredTxsBuffs, txBuff)
		if txIntercepted.IsAddressedToOtherShards() {
			log.Debug("intercepted tx is for other shards")

			continue
		}

		isTxInStorage, _ := txi.txStorer.Has(txIntercepted.Hash())
		if isTxInStorage {
			log.Debug("intercepted tx already processed")
			continue
		}

		cacherIdentifier := process.ShardCacherIdentifier(txIntercepted.SndShard(), txIntercepted.RcvShard())
		txi.txPool.AddData(
			txIntercepted.Hash(),
			txIntercepted.Transaction(),
			cacherIdentifier,
		)
	}

	var buffToSend []byte
	filteredOutTxsNeedToBeSend := len(filteredTxsBuffs) > 0 && lastErrEncountered != nil
	if filteredOutTxsNeedToBeSend {
		buffToSend, err = txi.marshalizer.Marshal(filteredTxsBuffs)
		if err != nil {
			return err
		}
	}

	if txi.broadcastCallbackHandler != nil {
		txi.broadcastCallbackHandler(buffToSend)
	}

	return lastErrEncountered
}

// SetBroadcastCallback sets the callback method to send filtered out message
func (txi *TxInterceptor) SetBroadcastCallback(callback func(buffToSend []byte)) {
	txi.broadcastCallbackHandler = callback
}
