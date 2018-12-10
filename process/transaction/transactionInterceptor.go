package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

var log = logger.NewDefaultLogger()

type transactionInterceptor struct {
	intercept     *process.Interceptor
	txPool        *dataPool.DataPool
	addrConverter state.AddressConverter
}

// NewTransactionInterceptor hooks a new interceptor for transactions
func NewTransactionInterceptor(
	messenger p2p.Messenger,
	txPool *dataPool.DataPool,
	addrConverter state.AddressConverter) (*transactionInterceptor, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if txPool == nil {
		return nil, process.ErrNilTxDataPool
	}

	if addrConverter == nil {
		return nil, process.ErrNilAddressConverter
	}

	intercept, err := process.NewInterceptor("tx", messenger, NewInterceptedTransaction())
	if err != nil {
		return nil, err
	}

	txIntercept := &transactionInterceptor{
		intercept:     intercept,
		txPool:        txPool,
		addrConverter: addrConverter,
	}

	intercept.CheckReceivedObject = txIntercept.processTx

	return txIntercept, nil
}

func (txi *transactionInterceptor) processTx(tx p2p.Newer, rawData []byte, hasher hashing.Hasher) bool {
	if tx == nil {
		log.Debug("nil tx to process")
		return false
	}

	if hasher == nil {
		return false
	}

	txIntercepted, ok := tx.(process.TransactionInterceptorAdapter)

	if !ok {
		log.Error("bad implementation: transactionInterceptor is not using InterceptedTransaction " +
			"as template object and will always return false")
		return false
	}

	txIntercepted.SetAddressConverter(txi.addrConverter)
	hash := hasher.Compute(string(rawData))
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
