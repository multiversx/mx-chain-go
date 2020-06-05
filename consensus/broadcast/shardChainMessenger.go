package broadcast

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

const validatorDelayPerOrder = time.Second

var _ consensus.BroadcastMessenger = (*shardChainMessenger)(nil)

type shardChainMessenger struct {
	*commonMessenger
	delayedBlockBroadcaster DelayedBroadcaster
}

// ShardChainMessengerArgs holds the arguments for creating a shardChainMessenger instance
type ShardChainMessengerArgs struct {
	CommonMessengerArgs
	HeadersSubscriber          consensus.HeadersPoolSubscriber
	InterceptorsContainer      process.InterceptorsContainer
	MaxDelayCacheSize          uint32
	MaxValidatorDelayCacheSize uint32
}

// NewShardChainMessenger creates a new shardChainMessenger object
func NewShardChainMessenger(
	args ShardChainMessengerArgs,
) (*shardChainMessenger, error) {

	err := checkShardChainNilParameters(args)
	if err != nil {
		return nil, err
	}

	cm := &commonMessenger{
		marshalizer:      args.Marshalizer,
		messenger:        args.Messenger,
		privateKey:       args.PrivateKey,
		shardCoordinator: args.ShardCoordinator,
		singleSigner:     args.SingleSigner,
	}

	dbbArgs := &DelayedBlockBroadcasterArgs{
		InterceptorsContainer: args.InterceptorsContainer,
		HeadersSubscriber:     args.HeadersSubscriber,
		LeaderCacheSize:       args.MaxDelayCacheSize,
		ValidatorCacheSize:    args.MaxValidatorDelayCacheSize,
		ShardCoordinator:      args.ShardCoordinator,
	}

	dbb, err := NewDelayedBlockBroadcaster(dbbArgs)
	if err != nil {
		return nil, err
	}

	scm := &shardChainMessenger{
		commonMessenger:         cm,
		delayedBlockBroadcaster: dbb,
	}

	err = dbb.SetBroadcastHandlers(scm.BroadcastMiniBlocks, scm.BroadcastTransactions)
	if err != nil {
		return nil, err
	}

	return scm, nil
}

func checkShardChainNilParameters(
	args ShardChainMessengerArgs,
) error {
	err := checkCommonMessengerNilParameters(args.CommonMessengerArgs)
	if err != nil {
		return err
	}
	if check.IfNil(args.InterceptorsContainer) {
		return spos.ErrNilInterceptorsContainer
	}
	if check.IfNil(args.HeadersSubscriber) {
		return spos.ErrNilHeadersSubscriber
	}
	if args.MaxDelayCacheSize == 0 {
		return spos.ErrInvalidCacheSize
	}

	return nil
}

// BroadcastBlock will send on in-shard headers topic and on in-shard miniblocks topic the header and block body
func (scm *shardChainMessenger) BroadcastBlock(blockBody data.BodyHandler, header data.HeaderHandler) error {
	if check.IfNil(blockBody) {
		return spos.ErrNilBody
	}

	err := blockBody.IntegrityAndValidity()
	if err != nil {
		return err
	}

	if check.IfNil(header) {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	b := blockBody.(*block.Body)
	msgBlockBody, err := scm.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	headerIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	selfIdentifier := scm.shardCoordinator.CommunicationIdentifier(scm.shardCoordinator.SelfId())

	go scm.messenger.Broadcast(factory.ShardBlocksTopic+headerIdentifier, msgHeader)
	go scm.messenger.Broadcast(factory.MiniBlocksTopic+selfIdentifier, msgBlockBody)

	return nil
}

// BroadcastHeader will send on in-shard headers topic the header
func (scm *shardChainMessenger) BroadcastHeader(header data.HeaderHandler) error {
	if check.IfNil(header) {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	shardIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	go scm.messenger.Broadcast(factory.ShardBlocksTopic+shardIdentifier, msgHeader)

	return nil
}

// SetLeaderDelayBroadcast sets the miniBlocks and transactions to be broadcast with delay
func (scm *shardChainMessenger) SetLeaderDelayBroadcast(
	headerHash []byte,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
) error {
	if len(headerHash) == 0 {
		return spos.ErrNilHeaderHash
	}
	if len(miniBlocks) == 0 {
		return nil
	}

	broadcastData := &delayedBroadcastData{
		headerHash:     headerHash,
		miniBlocksData: miniBlocks,
		transactions:   transactions,
	}

	return scm.delayedBlockBroadcaster.SetLeaderData(broadcastData)
}

// SetValidatorDelayBroadcast sets the miniBlocks and transactions to be broadcast with delay
// The broadcast will only be done in case the consensus leader fails its broadcast
// and all other validators with a lower order fail their broadcast as well.
func (scm *shardChainMessenger) SetValidatorDelayBroadcast(headerHash []byte, prevRandSeed []byte, round uint64, miniBlocks map[uint32][]byte, miniBlockHashes map[uint32]map[string]struct{}, transactions map[string][][]byte, order uint32) error {
	if len(headerHash) == 0 {
		return spos.ErrNilHeaderHash
	}
	if len(prevRandSeed) == 0 {
		return spos.ErrNilPrevRandSeed
	}
	if len(miniBlocks) == 0 && len(miniBlockHashes) == 0 {
		return nil
	}
	if len(miniBlocks) == 0 || len(miniBlockHashes) == 0 {
		return spos.ErrInvalidDataToBroadcast
	}

	broadcastData := &delayedBroadcastData{
		headerHash:      headerHash,
		prevRandSeed:    prevRandSeed,
		round:           round,
		miniBlocksData:  miniBlocks,
		miniBlockHashes: miniBlockHashes,
		transactions:    transactions,
		order:           order,
	}

	return scm.delayedBlockBroadcaster.SetValidatorData(broadcastData)
}

// Close closes all the started infinite looping goroutines and subcomponents
func (scm *shardChainMessenger) Close() {
	scm.delayedBlockBroadcaster.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *shardChainMessenger) IsInterfaceNil() bool {
	return scm == nil
}
