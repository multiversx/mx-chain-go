package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// TxInterceptor is used for intercepting transaction and storing them into a datapool
type TxInterceptor struct {
	process.Interceptor
	txPool           data.ShardedDataCacherNotifier
	txStorer         storage.Storer
	addrConverter    state.AddressConverter
	hasher           hashing.Hasher
	singleSigner     crypto.SingleSigner
	keyGen           crypto.KeyGenerator
	shardCoordinator sharding.ShardCoordinator
}

// NewTxInterceptor hooks a new interceptor for transactions
func NewTxInterceptor(
	interceptor process.Interceptor,
	txPool data.ShardedDataCacherNotifier,
	txStorer storage.Storer,
	addrConverter state.AddressConverter,
	hasher hashing.Hasher,
	singleSigner crypto.SingleSigner,
	keyGen crypto.KeyGenerator,
	shardCoordinator sharding.ShardCoordinator,
) (*TxInterceptor, error) {

	if interceptor == nil {
		return nil, process.ErrNilInterceptor
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
		Interceptor:      interceptor,
		txPool:           txPool,
		txStorer:         txStorer,
		hasher:           hasher,
		addrConverter:    addrConverter,
		singleSigner:     singleSigner,
		keyGen:           keyGen,
		shardCoordinator: shardCoordinator,
	}

	interceptor.SetCheckReceivedObjectHandler(txIntercept.processTx)

	return txIntercept, nil
}

func (txi *TxInterceptor) processTx(tx p2p.Creator, rawData []byte) error {
	if tx == nil {
		return process.ErrNilTransaction
	}

	if rawData == nil {
		return process.ErrNilDataToProcess
	}

	txIntercepted, ok := tx.(process.TransactionInterceptorAdapter)

	if !ok {
		return process.ErrBadInterceptorTopicImplementation
	}

	txIntercepted.SetAddressConverter(txi.addrConverter)
	txIntercepted.SetSingleSignKeyGen(txi.keyGen)
	hashWithSig := txi.hasher.Compute(string(rawData))
	txIntercepted.SetHash(hashWithSig)

	copiedTx := *txIntercepted.GetTransaction()
	copiedTx.Signature = nil

	marshalizer := txi.Marshalizer()
	if marshalizer == nil {
		return process.ErrNilMarshalizer
	}

	buffCopiedTx, err := marshalizer.Marshal(&copiedTx)
	if err != nil {
		return err
	}
	txIntercepted.SetTxBuffWithoutSig(buffCopiedTx)

	err = txIntercepted.IntegrityAndValidity(txi.shardCoordinator)
	if err != nil {
		return err
	}

	err = txIntercepted.VerifySig()
	if err != nil {
		return err
	}

	if txIntercepted.IsAddressedToOtherShards() {
		log.Debug("intercepted tx is for other shards")
		return nil
	}

	isTxInStorage, _ := txi.txStorer.Has(hashWithSig)

	if isTxInStorage {
		log.Debug("intercepted tx already processed")
		return nil
	}

	txi.txPool.AddData(hashWithSig, txIntercepted.GetTransaction(), txIntercepted.SndShard())
	return nil
}
