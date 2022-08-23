package pruning

import (
	storageCore "github.com/ElrondNetwork/elrond-go-core/storage"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type fullHistoryTriePruningStorer struct {
	*triePruningStorer
	fhps                           *FullHistoryPruningStorer
	args                           *StorerArgs
	shardId                        string
	oldEpochsActivePersistersCache storage.Cacher
}

// NewFullHistoryTriePruningStorer will return a new instance of PruningStorer without sharded directories' naming scheme
func NewFullHistoryTriePruningStorer(args *FullHistoryStorerArgs) (*fullHistoryTriePruningStorer, error) {
	return initFullHistoryTriePruningStorer(args, "")
}

func initFullHistoryTriePruningStorer(args *FullHistoryStorerArgs, shardId string) (*fullHistoryTriePruningStorer, error) {
	fhps, err := initFullHistoryPruningStorer(args, shardId)
	if err != nil {
		return nil, err
	}

	tps := &triePruningStorer{fhps.PruningStorer}
	fhps.PruningStorer.extendPersisterLifeHandler = tps.extendPersisterLife

	return &fullHistoryTriePruningStorer{
		triePruningStorer: tps,
		fhps:              fhps,
		args:              args.StorerArgs,
		shardId:           shardId,
	}, nil
}

// GetFromEpoch will call GetFromEpoch from the underlying FullHistoryPruningStorer
func (fhtps *fullHistoryTriePruningStorer) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	return fhtps.fhps.GetFromEpoch(key, epoch)
}

// GetBulkFromEpoch will call GetBulkFromEpoch from the underlying FullHistoryPruningStorer
func (fhtps *fullHistoryTriePruningStorer) GetBulkFromEpoch(keys [][]byte, epoch uint32) ([]storageCore.KeyValuePair, error) {
	return fhtps.fhps.GetBulkFromEpoch(keys, epoch)
}

// PutInEpoch will call PutInEpoch from the underlying FullHistoryPruningStorer
func (fhtps *fullHistoryTriePruningStorer) PutInEpoch(key []byte, data []byte, epoch uint32) error {
	return fhtps.fhps.PutInEpoch(key, data, epoch)
}

// Close will call Close from the underlying FullHistoryPruningStorer
func (fhtps *fullHistoryTriePruningStorer) Close() error {
	return fhtps.fhps.Close()
}
