package interceptors

import (
	blockData "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// TxBlockBodyInterceptor represents an interceptor used for transaction block bodies
type TxBlockBodyInterceptor struct {
	*messageChecker
	marshalizer      marshal.Marshalizer
	cache            storage.Cacher
	hasher           hashing.Hasher
	storer           storage.Storer
	shardCoordinator sharding.Coordinator
}

// NewTxBlockBodyInterceptor creates a new instance of a TxBlockBodyInterceptor
func NewTxBlockBodyInterceptor(
	marshalizer marshal.Marshalizer,
	cache storage.Cacher,
	storer storage.Storer,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
) (*TxBlockBodyInterceptor, error) {

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if cache == nil {
		return nil, process.ErrNilCacher
	}
	if storer == nil {
		return nil, process.ErrNilBlockBodyStorage
	}
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}

	return &TxBlockBodyInterceptor{
		messageChecker:   &messageChecker{},
		marshalizer:      marshalizer,
		cache:            cache,
		storer:           storer,
		hasher:           hasher,
		shardCoordinator: shardCoordinator,
	}, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (tbbi *TxBlockBodyInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	err := tbbi.checkMessage(message)
	if err != nil {
		return err
	}

	txBlockBody := block.NewInterceptedTxBlockBody()
	miniBlocks := make([]*blockData.MiniBlock, 0)
	x := message.Data()
	err = tbbi.marshalizer.Unmarshal(&miniBlocks, x)
	if err != nil {
		return err
	}
	txBlockBody.TxBlockBody = miniBlocks

	hash := tbbi.hasher.Compute(string(message.Data()))
	txBlockBody.SetHash(hash)

	err = txBlockBody.IntegrityAndValidity(tbbi.shardCoordinator)
	if err != nil {
		return err
	}

	blockBody, ok := txBlockBody.GetUnderlyingObject().(blockData.Body)
	if !ok {
		return process.ErrCouldNotDecodeUnderlyingBody
	}

	go tbbi.processBlockBody(txBlockBody, blockBody)

	return nil
}

func (tbbi *TxBlockBodyInterceptor) processBlockBody(txBlockBody *block.InterceptedTxBlockBody, blockBody blockData.Body) {
	err := tbbi.storer.Has(txBlockBody.Hash())
	isBlockInStorage := err == nil
	if isBlockInStorage {
		log.Debug("intercepted block body already processed")
		return
	}

	for _, miniblock := range blockBody {
		mbBytes, err := tbbi.marshalizer.Marshal(miniblock)
		if err != nil {
			log.Error(err.Error())
			return
		}

		tbbi.cache.HasOrAdd(tbbi.hasher.Compute(string(mbBytes)), miniblock)
	}
}
