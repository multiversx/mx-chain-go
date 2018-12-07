package transaction

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/interceptors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transactionPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type transactionInterceptor struct {
	intercept *interceptors.Interceptor
	txPool    *transactionPool.TransactionPool
	messenger p2p.Messenger
}

func NewTransactionInterceptor(messenger p2p.Messenger, txPool *transactionPool.TransactionPool) (*transactionInterceptor, error) {
	if messenger == nil {
		return nil, interceptors.ErrNilMessenger
	}

	if txPool == nil {
		return nil, interceptors.ErrNilInterceptor
	}

	intercept, err := interceptors.NewInterceptor("tx", messenger, NewTransactionNewer())
	if err != nil {
		return nil, err
	}

	txIntercept := &transactionInterceptor{
		intercept: intercept,
		txPool:    txPool,
		messenger: messenger,
	}

	intercept.CheckReceivedObject = txIntercept.processTx

	return txIntercept, nil
}

func (txi *transactionInterceptor) processTx(tx p2p.Newer, rawData []byte) bool {
	txNewer, ok := tx.(*TransactionNewer)

	if !ok {
		//it's ok to panic as this might happen only when implementing new interceptors
		panic("transactionInterceptor is not using TransactionNewer as template object")
	}

	txOk := txi.checkSanityTx(txNewer)
	if !txOk {
		return false
	}

	hash := txi.messenger.Hasher().Compute(string(rawData))
	srcShard := txi.computeSourceShard(txNewer)
	destShard := txi.computeDestinationShard(txNewer)

	txi.txPool.AddTransaction(hash, txNewer.Transaction, srcShard)
	if srcShard != destShard {
		txi.txPool.AddTransaction(hash, txNewer.Transaction, destShard)
	}

	return true
}

func (txi *transactionInterceptor) checkSanityTx(txNewer *TransactionNewer) bool {
	return txi.checkNilFields(txNewer) &&
		txi.checkSignature(txNewer)
}

func (txi *transactionInterceptor) checkNilFields(txNewer *TransactionNewer) bool {
	return txNewer.Transaction != nil &&
		txNewer.Signature != nil &&
		txNewer.Challenge != nil &&
		txNewer.RcvAddr != nil &&
		txNewer.SndAddr != nil
}

func (txi *transactionInterceptor) checkSignature(txNewer *TransactionNewer) bool {
	//TODO add real checking here
	return true
}

func (txi *transactionInterceptor) computeSourceShard(txNewer *TransactionNewer) uint32 {
	//TODO add real implementation here
	return 0
}

func (txi *transactionInterceptor) computeDestinationShard(txNewer *TransactionNewer) uint32 {
	//TODO add real implementation here
	return 0
}
