package transaction

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// TxInterceptor is used for intercepting transaction and storing them into a datapool
type TxInterceptor struct {
	marshalizer      marshal.Marshalizer
	txPool           dataRetriever.ShardedDataCacherNotifier
	txValidator      process.TxValidator
	addrConverter    state.AddressConverter
	hasher           hashing.Hasher
	singleSigner     crypto.SingleSigner
	keyGen           crypto.KeyGenerator
	shardCoordinator sharding.Coordinator
}

// NewTxInterceptor hooks a new interceptor for transactions
func NewTxInterceptor(
	marshalizer marshal.Marshalizer,
	txPool dataRetriever.ShardedDataCacherNotifier,
	txValidator process.TxValidator,
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
	if txValidator == nil {
		return nil, process.ErrNilTxHandlerValidator
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
		txValidator:      txValidator,
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
func (txi *TxInterceptor) ProcessReceivedMessage(message p2p.MessageP2P, broadcastHandler func(buffToSend []byte)) error {
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

		go txi.processTransaction(txIntercepted)
	}

	var buffToSend []byte
	filteredOutTxsNeedToBeSend := len(filteredTxsBuffs) > 0 && lastErrEncountered != nil
	if filteredOutTxsNeedToBeSend {
		buffToSend, err = txi.marshalizer.Marshal(filteredTxsBuffs)
		if err != nil {
			return err
		}
	}

	if broadcastHandler != nil {
		broadcastHandler(buffToSend)
	}

	return lastErrEncountered
}

func (txi *TxInterceptor) processTransaction(tx *InterceptedTransaction) {
	isTxValid := txi.txValidator.IsTxValidForProcessing(tx.Transaction())
	if !isTxValid {
		log.Debug(fmt.Sprintf("intercepted tx with hash %s is not valid", hex.EncodeToString(tx.hash)))
		return
	}

	cacherIdentifier := process.ShardCacherIdentifier(tx.SndShard(), tx.RcvShard())
	txi.txPool.AddData(
		tx.Hash(),
		tx.Transaction(),
		cacherIdentifier,
	)
}
