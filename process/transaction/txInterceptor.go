package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

var log = logger.NewDefaultLogger()

// TxInterceptor is used for intercepting transaction and storing them into a datapool
type TxInterceptor struct {
	process.Interceptor
	txPool        data.ShardedDataCacherNotifier
	addrConverter state.AddressConverter
	hasher        hashing.Hasher
}

// NewTxInterceptor hooks a new interceptor for transactions
func NewTxInterceptor(
	interceptor process.Interceptor,
	txPool data.ShardedDataCacherNotifier,
	addrConverter state.AddressConverter,
	hasher hashing.Hasher,
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

	txIntercept := &TxInterceptor{
		Interceptor:   interceptor,
		txPool:        txPool,
		hasher:        hasher,
		addrConverter: addrConverter,
	}

	interceptor.SetCheckReceivedObjectHandler(txIntercept.processTx)

	return txIntercept, nil
}

func (txi *TxInterceptor) processTx(tx p2p.Newer, rawData []byte) bool {
	if tx == nil {
		log.Debug("nil tx to process")
		return false
	}

	txIntercepted, ok := tx.(process.TransactionInterceptorAdapter)

	if !ok {
		log.Error("bad implementation: transactionInterceptor is not using InterceptedTransaction " +
			"as template object and will always return false")
		return false
	}

	txIntercepted.SetAddressConverter(txi.addrConverter)
	hash := txi.hasher.Compute(string(rawData))
	txIntercepted.SetHash(hash)

	if !txIntercepted.Check() || !txIntercepted.VerifySig() {
		log.Debug("intercepted tx failed the check or verifySig methods")
		return false
	}

	if txIntercepted.IsAddressedToOtherShards() {
		log.Debug("intercepted tx is for other shards")
		return true
	}

	txi.txPool.AddData(hash, txIntercepted.GetTransaction(), txIntercepted.SndShard())
	if txIntercepted.SndShard() != txIntercepted.RcvShard() {
		log.Debug("cross shard tx")
		txi.txPool.AddData(hash, txIntercepted.GetTransaction(), txIntercepted.RcvShard())
	}

	return true
}
