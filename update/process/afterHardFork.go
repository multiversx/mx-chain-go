package process

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

// TODO: save blocks and transactions to storage
// TODO: marshalize if there are cross shard results (smart contract results in special)
// TODO: use new genesis block process, integrate with it

// ArgsAfterHardFork defines the arguments for the new after hard fork process handler
type ArgsAfterHardFork struct {
	MapBlockProcessors map[uint32]update.HardForkBlockProcessor
	ImportHandler      update.ImportHandler
	ShardCoordinator   sharding.Coordinator
	Hasher             hashing.Hasher
	Marshalizer        marshal.Marshalizer
}

type afterHardFork struct {
	mapBlockProcessors map[uint32]update.HardForkBlockProcessor
	importHandler      update.ImportHandler
	shardCoordinator   sharding.Coordinator
	hasher             hashing.Hasher
	marshalizer        marshal.Marshalizer
}

// NewAfterHardForkBlockCreation creates the after hard fork block creator process handler
func NewAfterHardForkBlockCreation(args ArgsAfterHardFork) (*afterHardFork, error) {
	if args.MapBlockProcessors == nil {
		return nil, update.ErrNilHardForkBlockProcessor
	}
	if check.IfNil(args.ImportHandler) {
		return nil, update.ErrNilImportHandler
	}
	if check.IfNil(args.Hasher) {
		return nil, update.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, update.ErrNilMarshalizer
	}
	if check.IfNil(args.ShardCoordinator) {
		return nil, update.ErrNilShardCoordinator
	}

	return &afterHardFork{
		mapBlockProcessors: args.MapBlockProcessors,
		importHandler:      args.ImportHandler,
		shardCoordinator:   args.ShardCoordinator,
		hasher:             args.Hasher,
		marshalizer:        args.Marshalizer,
	}, nil
}

// CreateAllBlocksAfterHardfork creates all the blocks after hardfork
func (a *afterHardFork) CreateAllBlocksAfterHardfork(
	chainID string,
	round uint64,
	nonce uint64,
	epoch uint32,
) (map[uint32]data.HeaderHandler, map[uint32]*block.Body, error) {
	mapHeaders := make(map[uint32]data.HeaderHandler)
	mapBodies := make(map[uint32]*block.Body)

	shardIDs := make([]uint32, a.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < a.shardCoordinator.NumberOfShards(); i++ {
		shardIDs[i] = i
	}
	shardIDs[a.shardCoordinator.NumberOfShards()] = core.MetachainShardId

	lastPostMbs, err := update.CreateBody(a.hasher, a.marshalizer, shardIDs, mapBodies, a.mapBlockProcessors)
	if err != nil {
		return nil, nil, err
	}

	err = update.CreatePostMiniBlocks(a.hasher, a.marshalizer, shardIDs, lastPostMbs, mapBodies, a.mapBlockProcessors)
	if err != nil {
		return nil, nil, err
	}

	err = a.createHeaders(shardIDs, mapBodies, mapHeaders, chainID, round, nonce, epoch)
	if err != nil {
		return nil, nil, err
	}

	return mapHeaders, mapBodies, nil
}

func (a *afterHardFork) createHeaders(
	shardIDs []uint32,
	mapBodies map[uint32]*block.Body,
	mapHeaders map[uint32]data.HeaderHandler,
	chainID string,
	round uint64,
	nonce uint64,
	epoch uint32,
) error {
	for _, shardID := range shardIDs {
		log.Debug("afterHardFork.createHeader", "shard", shardID)

		blockProcessor, ok := a.mapBlockProcessors[shardID]
		if !ok {
			return update.ErrNilHardForkBlockProcessor
		}

		hdr, err := blockProcessor.CreateBlock(mapBodies[shardID], chainID, round, nonce, epoch)
		if err != nil {
			return err
		}

		mapHeaders[shardID] = hdr
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (a *afterHardFork) IsInterfaceNil() bool {
	return a == nil
}
