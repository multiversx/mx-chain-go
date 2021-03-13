package interceptedBlocks

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ process.InterceptedData = (*InterceptedMiniblock)(nil)

// InterceptedMiniblock is a wrapper over a miniblock
type InterceptedMiniblock struct {
	miniblock         *block.MiniBlock
	marshalizer       marshal.Marshalizer
	hasher            hashing.Hasher
	shardCoordinator  sharding.Coordinator
	hash              []byte
	isForCurrentShard bool
}

// NewInterceptedMiniblock creates a new instance of InterceptedMiniblock struct
func NewInterceptedMiniblock(arg *ArgInterceptedMiniblock) (*InterceptedMiniblock, error) {
	err := checkMiniblockArgument(arg)
	if err != nil {
		return nil, err
	}

	miniblock, err := createMiniblock(arg.Marshalizer, arg.MiniblockBuff)
	if err != nil {
		return nil, err
	}

	inMiniblock := &InterceptedMiniblock{
		miniblock:        miniblock,
		marshalizer:      arg.Marshalizer,
		hasher:           arg.Hasher,
		shardCoordinator: arg.ShardCoordinator,
	}
	inMiniblock.processFields(arg.MiniblockBuff)

	return inMiniblock, nil
}

func createMiniblock(marshalizer marshal.Marshalizer, miniblockBuff []byte) (*block.MiniBlock, error) {
	miniblock := &block.MiniBlock{}
	err := marshalizer.Unmarshal(miniblock, miniblockBuff)
	if err != nil {
		return nil, err
	}

	return miniblock, nil
}

func (inMb *InterceptedMiniblock) processFields(mbBuff []byte) {
	inMb.hash = inMb.hasher.Compute(string(mbBuff))

	inMb.processIsForCurrentShard()
}

func (inMb *InterceptedMiniblock) processIsForCurrentShard() {
	isForCurrentShardRecv := inMb.miniblock.ReceiverShardID == inMb.shardCoordinator.SelfId()
	isForCurrentShardSender := inMb.miniblock.SenderShardID == inMb.shardCoordinator.SelfId()
	isForAllShards := inMb.miniblock.ReceiverShardID == core.AllShardId

	inMb.isForCurrentShard = isForCurrentShardRecv || isForCurrentShardSender || isForAllShards
}

// Hash gets the hash of this transaction block body
func (inMb *InterceptedMiniblock) Hash() []byte {
	return inMb.hash
}

// Miniblock returns the miniblock held by this wrapper
func (inMb *InterceptedMiniblock) Miniblock() *block.MiniBlock {
	return inMb.miniblock
}

// CheckValidity checks if the received tx block body is valid (not nil fields)
func (inMb *InterceptedMiniblock) CheckValidity() error {
	return inMb.integrity()
}

// IsForCurrentShard returns true if at least one contained miniblock is for current shard
func (inMb *InterceptedMiniblock) IsForCurrentShard() bool {
	return inMb.isForCurrentShard
}

// integrity checks the integrity of the tx block body
func (inMb *InterceptedMiniblock) integrity() error {
	miniblock := inMb.miniblock

	receiverNotCurrentShard := miniblock.ReceiverShardID >= inMb.shardCoordinator.NumberOfShards() &&
		(miniblock.ReceiverShardID != core.MetachainShardId && miniblock.ReceiverShardID != core.AllShardId)
	if receiverNotCurrentShard {
		return process.ErrInvalidShardId
	}

	senderNotCurrentShard := miniblock.SenderShardID >= inMb.shardCoordinator.NumberOfShards() &&
		miniblock.SenderShardID != core.MetachainShardId
	if senderNotCurrentShard {
		return process.ErrInvalidShardId
	}

	for _, txHash := range miniblock.TxHashes {
		if txHash == nil {
			return process.ErrNilTxHash
		}
	}

	//TODO: Remove this condition after the final decision about reserved field from mini block
	//if len(miniblock.Reserved) > 0 {
	//	return process.ErrReservedFieldNotSupportedYet
	//}

	return nil
}

// Type returns the type of this intercepted data
func (inMb *InterceptedMiniblock) Type() string {
	return "intercepted miniblock"
}

// String returns the transactions body's most important fields as string
func (inMb *InterceptedMiniblock) String() string {
	return fmt.Sprintf("miniblock type=%s, numTxs=%d, sender shardid=%d, recv shardid=%d",
		inMb.miniblock.Type.String(),
		len(inMb.miniblock.TxHashes),
		inMb.miniblock.SenderShardID,
		inMb.miniblock.ReceiverShardID,
	)
}

// Identifiers returns the identifiers used in requests
func (inMb *InterceptedMiniblock) Identifiers() [][]byte {
	return [][]byte{inMb.hash}
}

// IsInterfaceNil returns true if there is no value under the interface
func (inMb *InterceptedMiniblock) IsInterfaceNil() bool {
	return inMb == nil
}
