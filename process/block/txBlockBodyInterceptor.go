package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/dataPool"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type txBlockBodyInterceptor struct {
	intercept  *process.Interceptor
	txBodyPool *dataPool.DataPool
}

func NewTxBlockBodyInterceptor(
	messenger p2p.Messenger,
	txBodyPool *dataPool.DataPool,
) (*txBlockBodyInterceptor, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if txBodyPool == nil {
		return nil, process.ErrNilTxBodyBlockDataPool
	}

	intercept, err := process.NewInterceptor("hdr", messenger, NewInterceptedTxBlockBody())
	if err != nil {
		return nil, err
	}

	txBodyBlockIntercept := &txBlockBodyInterceptor{
		intercept:  intercept,
		txBodyPool: txBodyPool,
	}

	intercept.CheckReceivedObject = txBodyBlockIntercept.processTxBodyBlock

	return txBodyBlockIntercept, nil
}

func (txbbi *txBlockBodyInterceptor) processTxBodyBlock(txBodyBlock p2p.Newer, rawData []byte, hasher hashing.Hasher) bool {
	if txBodyBlock == nil {
		log.Debug("nil tx body block to process")
		return false
	}

	if hasher == nil {
		return false
	}

	txBlockBodyIntercepted, ok := txBodyBlock.(process.TxBlockBodyInterceptorAdapter)

	if !ok {
		log.Error("bad implementation: txBlockBodyInterceptor is not using InterceptedTxBlockBody " +
			"as template object and will always return false")
		return false
	}

	hash := hasher.Compute(string(rawData))
	txBlockBodyIntercepted.SetHash(hash)

	if !txBlockBodyIntercepted.Check() {
		return false
	}

	txbbi.txBodyPool.AddData(hash, txBlockBodyIntercepted, txBlockBodyIntercepted.Shard())
	return true
}
