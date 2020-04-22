package process

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

// TODO: do factory which creates the needed components
// TODO: consensus on the first created block -> integrate the hardfork after processor into the normal block processor ?
// TODO: save blocks to storage, marshalize, broadcast.
// TODO: set previous roothash to genesis block
// TODO: use new genesis block process

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
	if args.ImportHandler != nil {
		return nil, update.ErrNilImportHandler
	}
	if args.Hasher != nil {
		return nil, update.ErrNilHasher
	}
	if args.Marshalizer != nil {
		return nil, update.ErrNilMarshalizer
	}
	if args.ShardCoordinator != nil {
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

// CreateAllBlockAfterHardfork creates all the blocks after hardfork
func (a *afterHardFork) CreateAllBlocksAfterHardfork(
	chainID string,
	round uint64,
	nonce uint64,
	epoch uint32,
) (map[uint32]data.HeaderHandler, map[uint32]data.BodyHandler, error) {
	mapHeaders := make(map[uint32]data.HeaderHandler)
	mapBodies := make(map[uint32]data.BodyHandler)

	shardIDs := make([]uint32, 0, a.shardCoordinator.NumberOfShards()+1)
	for i := uint32(0); i < a.shardCoordinator.NumberOfShards(); i++ {
		shardIDs[i] = i
	}
	shardIDs[a.shardCoordinator.NumberOfShards()] = core.MetachainShardId

	for _, shardId := range shardIDs {
		blockProcessor, ok := a.mapBlockProcessors[shardId]
		if !ok {
			return nil, nil, update.ErrNilHardForkBlockProcessor
		}

		hdr, body, err := blockProcessor.CreateNewBlock(chainID, round, nonce, epoch)
		if err != nil {
			return nil, nil, err
		}

		mapHeaders[shardId] = hdr
		mapBodies[shardId] = body
	}

	return mapHeaders, mapBodies, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (a *afterHardFork) IsInterfaceNil() bool {
	return a == nil
}
