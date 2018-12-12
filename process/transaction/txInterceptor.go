package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/interceptor"
)

var log = logger.NewDefaultLogger()

// TxInterceptor is used for intercepting transaction and storing them into a datapool
type TxInterceptor struct {
	*interceptor.Interceptor
	txPool        *shardedData.ShardedData
	addrConverter state.AddressConverter
	hasher        hashing.Hasher
}

// NewTxInterceptor hooks a new interceptor for transactions
func NewTxInterceptor(
	messenger p2p.Messenger,
	txPool *shardedData.ShardedData,
	addrConverter state.AddressConverter,
	hasher hashing.Hasher,
) (*TxInterceptor, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
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

	intercept, err := interceptor.NewInterceptor(
		process.TxInterceptor,
		messenger,
		NewInterceptedTransaction())
	if err != nil {
		return nil, err
	}

	txIntercept := &TxInterceptor{
		Interceptor:   intercept,
		txPool:        txPool,
		addrConverter: addrConverter,
		hasher:        hasher,
	}

	intercept.CheckReceivedObject = txIntercept.processTx

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
		return false
	}

	if txIntercepted.IsAddressedToOtherShards() {
		return true
	}

	txi.txPool.AddData(hash, txIntercepted.GetTransaction(), txIntercepted.SndShard())
	if txIntercepted.SndShard() != txIntercepted.RcvShard() {
		log.Debug("cross shard tx")
		txi.txPool.AddData(hash, txIntercepted.GetTransaction(), txIntercepted.RcvShard())
	}

	return true
}
