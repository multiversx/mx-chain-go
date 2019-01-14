package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// TxInterceptor is used for intercepting transaction and storing them into a datapool
type TxInterceptor struct {
	process.Interceptor
	txPool           data.ShardedDataCacherNotifier
	addrConverter    state.AddressConverter
	hasher           hashing.Hasher
	singleSignKeyGen crypto.KeyGenerator
	shardCoordinator sharding.ShardCoordinator
}

// NewTxInterceptor hooks a new interceptor for transactions
func NewTxInterceptor(
	interceptor process.Interceptor,
	txPool data.ShardedDataCacherNotifier,
	addrConverter state.AddressConverter,
	hasher hashing.Hasher,
	singleSignKeyGen crypto.KeyGenerator,
	shardCoordinator sharding.ShardCoordinator,

) (*TxInterceptor, error) {

	if interceptor == nil {
		return nil, process.ErrNilInterceptor
	}

	if txPool == nil {
		return nil, process.ErrNilTxDataPool
	}

	if addrConverter == nil {
		return nil, process.ErrNilAddressConverter
	}

	if hasher == nil {
		return nil, process.ErrNilHasher
	}

	if singleSignKeyGen == nil {
		return nil, process.ErrNilSingleSignKeyGen
	}

	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}

	txIntercept := &TxInterceptor{
		Interceptor:      interceptor,
		txPool:           txPool,
		hasher:           hasher,
		addrConverter:    addrConverter,
		singleSignKeyGen: singleSignKeyGen,
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
	txIntercepted.SetSingleSignKeyGen(txi.singleSignKeyGen)
	hash := txi.hasher.Compute(string(rawData))
	txIntercepted.SetHash(hash)

	err := txIntercepted.IntegrityAndValidity(txi.shardCoordinator)
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

	txi.txPool.AddData(hash, txIntercepted.GetTransaction(), txIntercepted.SndShard())

	return nil
}
