package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// TxInterceptor is used for intercepting transaction and storing them into a datapool
type TxInterceptor struct {
	marshalizer      marshal.Marshalizer
	txPool           data.ShardedDataCacherNotifier
	txStorer         storage.Storer
	addrConverter    state.AddressConverter
	hasher           hashing.Hasher
	singleSigner     crypto.SingleSigner
	keyGen           crypto.KeyGenerator
	shardCoordinator sharding.Coordinator
}

// NewTxInterceptor hooks a new interceptor for transactions
func NewTxInterceptor(
	marshalizer marshal.Marshalizer,
	txPool data.ShardedDataCacherNotifier,
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
func (txi *TxInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) ([]byte, error) {
	if message == nil {
		return nil, process.ErrNilMessage
	}

	if message.Data() == nil {
		return nil, process.ErrNilDataToProcess
	}

	txsBuff := make([][]byte, 0)
	err := txi.marshalizer.Unmarshal(&txsBuff, message.Data())
	if err != nil {
		return nil, err
	}
	if len(txsBuff) == 0 {
		return nil, process.ErrNoTransactionInMessage
	}

	filteredTxsBuffs := make([][]byte, 0)
	lastErrEncountered := error(nil)
	for _, txBuff := range txsBuff {
		tx := &transaction.Transaction{}
		err := txi.marshalizer.Unmarshal(tx, txBuff)
		if err != nil {
			lastErrEncountered = err
			continue
		}

		txIntercepted := NewInterceptedTransaction(txi.singleSigner)
		txIntercepted.Transaction = tx
		txIntercepted.SetAddressConverter(txi.addrConverter)
		txIntercepted.SetSingleSignKeyGen(txi.keyGen)

		copiedTx := *txIntercepted.GetTransaction()
		copiedTx.Signature = nil
		buffCopiedTx, err := txi.marshalizer.Marshal(&copiedTx)
		if err != nil {
			lastErrEncountered = err
			continue
		}

		txIntercepted.SetTxBuffWithoutSig(buffCopiedTx)
		hashWithSig := txi.hasher.Compute(string(txBuff))
		txIntercepted.SetHash(hashWithSig)
		err = txIntercepted.IntegrityAndValidity(txi.shardCoordinator)
		if err != nil {
			lastErrEncountered = err
			continue
		}

		err = txIntercepted.VerifySig()
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

		isTxInStorage, _ := txi.txStorer.Has(hashWithSig)
		if isTxInStorage {
			log.Debug("intercepted tx already processed")
			continue
		}

		cacherIdentifier := process.ShardCacherIdentifier(txIntercepted.SndShard(), txIntercepted.RcvShard())
		txi.txPool.AddData(
			hashWithSig,
			txIntercepted.GetTransaction(),
			cacherIdentifier,
		)
	}

	var buffToSend []byte
	filteredOutTxsNeedToBeSend := len(filteredTxsBuffs) > 0 && lastErrEncountered != nil
	if filteredOutTxsNeedToBeSend {
		buffToSend, err = txi.marshalizer.Marshal(filteredTxsBuffs)
		if err != nil {
			return nil, err
		}
	}

	return buffToSend, lastErrEncountered
}
