package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// InterceptedTxBlockBody represents the wrapper over TxBlockBodyWrapper struct.
type InterceptedTxBlockBody struct {
	txBlockBody       block.Body
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	shardCoordinator  sharding.Coordinator
	hash              []byte
	isForCurrentShard bool
}

// NewInterceptedTxBlockBody creates a new instance of InterceptedTxBlockBody struct
func NewInterceptedTxBlockBody(arg *ArgInterceptedTxBlockBody) (*InterceptedTxBlockBody, error) {
	err := checkTxBlockBodyArgument(arg)
	if err != nil {
		return nil, err
	}

	txBlockBody, err := createTxBlockBody(arg.Marshalizer, arg.TxBlockBodyBuff)
	if err != nil {
		return nil, err
	}

	inTxBody := &InterceptedTxBlockBody{
		txBlockBody:      txBlockBody,
		marshalizer:      arg.Marshalizer,
		hasher:           arg.Hasher,
		shardCoordinator: arg.ShardCoordinator,
	}
	inTxBody.processFields(arg.TxBlockBodyBuff)

	return inTxBody, nil
}

func createTxBlockBody(marshalizer marshal.Marshalizer, txBlockBodyBuff []byte) (block.Body, error) {
	txBlockBody := make(block.Body, 0)
	err := marshalizer.Unmarshal(&txBlockBody, txBlockBodyBuff)
	if err != nil {
		return nil, err
	}

	return txBlockBody, nil
}

func (inTxBlkBdy *InterceptedTxBlockBody) processFields(txBuff []byte) {
	inTxBlkBdy.hash = inTxBlkBdy.hasher.Compute(string(txBuff))

	inTxBlkBdy.processIsForCurrentShard(inTxBlkBdy.txBlockBody)
}

func (inTxBlkBdy *InterceptedTxBlockBody) processIsForCurrentShard(txBlockBody block.Body) {
	inTxBlkBdy.isForCurrentShard = false
	for _, miniblock := range inTxBlkBdy.txBlockBody {
		inTxBlkBdy.isForCurrentShard = inTxBlkBdy.miniblockForCurrentShard(miniblock)
		if inTxBlkBdy.isForCurrentShard {
			return
		}
	}
}

func (inTxBlkBdy *InterceptedTxBlockBody) miniblockForCurrentShard(miniblock *block.MiniBlock) bool {
	isForCurrentShardRecv := miniblock.ReceiverShardID == inTxBlkBdy.shardCoordinator.SelfId()
	isForCurrentShardSender := miniblock.SenderShardID == inTxBlkBdy.shardCoordinator.SelfId()

	return isForCurrentShardRecv || isForCurrentShardSender
}

// Hash gets the hash of this transaction block body
func (inTxBlkBdy *InterceptedTxBlockBody) Hash() []byte {
	return inTxBlkBdy.hash
}

// TxBlockBody returns the block body held by this wrapper
func (inTxBlkBdy *InterceptedTxBlockBody) TxBlockBody() block.Body {
	return inTxBlkBdy.txBlockBody
}

// CheckValidity checks if the received tx block body is valid (not nil fields)
func (inTxBlkBdy *InterceptedTxBlockBody) CheckValidity() error {
	return inTxBlkBdy.integrity()
}

// IsForCurrentShard returns true if at least one contained miniblock is for current shard
func (inTxBlkBdy *InterceptedTxBlockBody) IsForCurrentShard() bool {
	return inTxBlkBdy.isForCurrentShard
}

// integrity checks the integrity of the tx block body
func (inTxBlkBdy *InterceptedTxBlockBody) integrity() error {
	for _, miniBlock := range inTxBlkBdy.txBlockBody {
		if miniBlock.TxHashes == nil {
			return process.ErrNilTxHashes
		}

		if miniBlock.ReceiverShardID >= inTxBlkBdy.shardCoordinator.NumberOfShards() {
			return process.ErrInvalidShardId
		}

		if miniBlock.SenderShardID >= inTxBlkBdy.shardCoordinator.NumberOfShards() {
			return process.ErrInvalidShardId
		}

		for _, txHash := range miniBlock.TxHashes {
			if txHash == nil {
				return process.ErrNilTxHash
			}
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (inTxBlkBdy *InterceptedTxBlockBody) IsInterfaceNil() bool {
	if inTxBlkBdy == nil {
		return true
	}
	return false
}
