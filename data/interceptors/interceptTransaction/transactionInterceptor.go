package interceptTransaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

var log = logger.NewDefaultLogger()

type transactionInterceptor struct {
	intercept     *interceptors.Interceptor
	txPool        *transactionPool.TransactionPool
	addrConverter state.AddressConverter
}

func NewTransactionInterceptor(
	messenger p2p.Messenger,
	txPool *transactionPool.TransactionPool,
	addrConverter state.AddressConverter) (*transactionInterceptor, error) {

	if messenger == nil {
		return nil, interceptors.ErrNilMessenger
	}

	if txPool == nil {
		return nil, interceptors.ErrNilTxPool
	}

	if addrConverter == nil {
		return nil, interceptors.ErrNilAddressConverter
	}

	intercept, err := interceptors.NewInterceptor("tx", messenger, NewInterceptedTransaction())
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

func (txi *transactionInterceptor) processTx(tx p2p.Newer, rawData []byte) bool {
	if tx == nil {
		log.Debug("nil tx to process")
		return false
	}

	txIntercepted, ok := tx.(interceptors.TransactionInterceptorAdapter)

	if !ok {
		log.Error("bad implementation: transactionInterceptor is not using InterceptedTransaction " +
			"as template object and will always return false")
		return false
	}

	txIntercepted.SetAddressConverter(txi.addrConverter)

	if !txIntercepted.Check() || !txIntercepted.VerifySig() {
		return false
	}

	mes := txi.intercept.Messenger()
	if mes == nil {
		return false
	}
	hasher := mes.Hasher()
	if hasher == nil {
		return false
	}

	if txIntercepted.IsAddressedToOtherShards() {
		return true
	}

	hash := hasher.Compute(string(rawData))

	txi.txPool.AddTransaction(hash, txIntercepted.GetTransaction(), txIntercepted.SndShard())
	if txIntercepted.SndShard() != txIntercepted.RcvShard() {
		log.Debug("cross shard tx")
		txi.txPool.AddTransaction(hash, txIntercepted.GetTransaction(), txIntercepted.RcvShard())
	}

	return true
}
