package pruning

import (
	"math"

	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/lrucache"
)

type fullHistoryTriePruningStorer struct {
	*triePruningStorer
	args                           *StorerArgs
	shardId                        string
	oldEpochsActivePersistersCache storage.Cacher
}

// NewFullHistoryTriePruningStorer will return a new instance of PruningStorer without sharded directories' naming scheme
func NewFullHistoryTriePruningStorer(args *FullHistoryStorerArgs) (*fullHistoryTriePruningStorer, error) {
	return initFullHistoryTriePruningStorer(args, "")
}

func initFullHistoryTriePruningStorer(args *FullHistoryStorerArgs, shardId string) (*fullHistoryTriePruningStorer, error) {
	err := checkArgs(args.StorerArgs)
	if err != nil {
		return nil, err
	}

	activePersisters, persistersMapByEpoch, err := initPersistersInEpoch(args.StorerArgs, shardId)
	if err != nil {
		return nil, err
	}

	ps, err := initPruningStorer(args.StorerArgs, shardId, activePersisters, persistersMapByEpoch)
	if err != nil {
		return nil, err
	}

	tps := &triePruningStorer{ps}
	ps.extendPersisterLifeHandler = tps.extendPersisterLife
	tps.registerHandler(args.Notifier)

	if args.NumOfOldActivePersisters < 1 || args.NumOfOldActivePersisters > math.MaxInt32 {
		return nil, storage.ErrInvalidNumberOfOldPersisters
	}

	fhps := &fullHistoryTriePruningStorer{
		triePruningStorer: tps,
		args:              args.StorerArgs,
		shardId:           shardId,
	}
	fhps.oldEpochsActivePersistersCache, err = lrucache.NewCacheWithEviction(int(args.NumOfOldActivePersisters), fhps.onEvicted)
	if err != nil {
		return nil, err
	}

	return fhps, nil
}
