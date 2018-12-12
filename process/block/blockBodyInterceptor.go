package block

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/process/interceptor"
)

// GenericBlockBodyInterceptor represents an interceptor used for all types of block bodies
type GenericBlockBodyInterceptor struct {
	intercept     *interceptor.Interceptor
	bodyBlockPool *shardedData.ShardedData
	hasher        hashing.Hasher
}

// NewGenericBlockBodyInterceptor hooks a new interceptor for block bodies
// Fetched data blocks will be placed inside data pools
func NewGenericBlockBodyInterceptor(
	name string,
	messenger p2p.Messenger,
	bodyBlockPool *shardedData.ShardedData,
	hasher hashing.Hasher,
	templateObj process.BlockBodyInterceptorAdapter,
) (*GenericBlockBodyInterceptor, error) {

	if messenger == nil {
		return nil, process.ErrNilMessenger
	}

	if bodyBlockPool == nil {
		return nil, process.ErrNilBodyBlockDataPool
	}

	if hasher == nil {
		return nil, process.ErrNilHasher
	}

	if templateObj == nil {
		return nil, process.ErrNilTemplateObj
	}

	intercept, err := interceptor.NewInterceptor(name, messenger, templateObj)
	if err != nil {
		return nil, err
	}

	bbIntercept := &GenericBlockBodyInterceptor{
		intercept:     intercept,
		bodyBlockPool: bodyBlockPool,
		hasher:        hasher,
	}

	intercept.CheckReceivedObject = bbIntercept.processBodyBlock

	return bbIntercept, nil
}

func (gbbi *GenericBlockBodyInterceptor) processBodyBlock(bodyBlock p2p.Newer, rawData []byte) bool {
	if bodyBlock == nil {
		log.Debug("nil body block to process")
		return false
	}

	txBlockBodyIntercepted, ok := bodyBlock.(process.BlockBodyInterceptorAdapter)

	if !ok {
		log.Error("bad implementation: BlockBodyInterceptor is not using BlockBodyInterceptorAdapter " +
			"as template object and will always return false")
		return false
	}

	hash := gbbi.hasher.Compute(string(rawData))
	txBlockBodyIntercepted.SetHash(hash)

	if !txBlockBodyIntercepted.Check() {
		return false
	}

	gbbi.bodyBlockPool.AddData(hash, txBlockBodyIntercepted, txBlockBodyIntercepted.Shard())
	return true
}
