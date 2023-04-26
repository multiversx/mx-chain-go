package pruning

import (
	storageCore "github.com/multiversx/mx-chain-core-go/storage"
)

type fullHistoryTriePruningStorer struct {
	*triePruningStorer
	storerWithEpochOperations storerWithEpochOperations
	args                      StorerArgs
	shardId                   string
}

// NewFullHistoryTriePruningStorer will return a new instance of PruningStorer without sharded directories' naming scheme
func NewFullHistoryTriePruningStorer(args FullHistoryStorerArgs) (*fullHistoryTriePruningStorer, error) {
	return initFullHistoryTriePruningStorer(args, "")
}

func initFullHistoryTriePruningStorer(args FullHistoryStorerArgs, shardId string) (*fullHistoryTriePruningStorer, error) {
	fhps, err := initFullHistoryPruningStorer(args, shardId)
	if err != nil {
		return nil, err
	}

	tps := &triePruningStorer{
		PruningStorer: fhps.PruningStorer,
	}
	fhps.PruningStorer.lastEpochNeededHandler = tps.lastEpochNeeded
	tps.registerHandler(args.Notifier)

	return &fullHistoryTriePruningStorer{
		triePruningStorer:         tps,
		storerWithEpochOperations: fhps,
		args:                      args.StorerArgs,
		shardId:                   shardId,
	}, nil
}

// GetFromEpoch will call the same function from the underlying FullHistoryPruningStorer
func (fhtps *fullHistoryTriePruningStorer) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	return fhtps.storerWithEpochOperations.GetFromEpoch(key, epoch)
}

// GetBulkFromEpoch will call the same function from the underlying FullHistoryPruningStorer
func (fhtps *fullHistoryTriePruningStorer) GetBulkFromEpoch(keys [][]byte, epoch uint32) ([]storageCore.KeyValuePair, error) {
	return fhtps.storerWithEpochOperations.GetBulkFromEpoch(keys, epoch)
}

// PutInEpoch will call the same function from the underlying FullHistoryPruningStorer
func (fhtps *fullHistoryTriePruningStorer) PutInEpoch(key []byte, data []byte, epoch uint32) error {
	return fhtps.storerWithEpochOperations.PutInEpoch(key, data, epoch)
}

// Close will call the same function from the underlying FullHistoryPruningStorer
func (fhtps *fullHistoryTriePruningStorer) Close() error {
	return fhtps.storerWithEpochOperations.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (fhtps *fullHistoryTriePruningStorer) IsInterfaceNil() bool {
	return fhtps == nil
}
