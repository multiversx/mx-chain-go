package pruning_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/storage/database"
	"github.com/multiversx/mx-chain-go/storage/mock"
	"github.com/multiversx/mx-chain-go/storage/pruning"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/require"
)

type countedEpochPersister struct {
	storage.Persister
	reads *int
}

func (p *countedEpochPersister) Get(key []byte) ([]byte, error) {
	*p.reads++
	return p.Persister.Get(key)
}

func TestEpochBoundaryRead(t *testing.T) {
	for _, fullHistory := range []bool{false, true} {
		for _, scenario := range []string{"cache", "same epoch", "next epoch", "missing", "outside range"} {
			t.Run(fmt.Sprintf("fullHistory=%t/%s", fullHistory, scenario), func(t *testing.T) {
				args := getDefaultArgs()
				args.EpochsData.StartingEpoch = 12
				args.EpochsData.NumOfActivePersisters = 3
				args.EpochsData.NumOfEpochsToKeep = 3
				if fullHistory {
					args.EpochsData.StartingEpoch = 20
				}
				args.PersistersTracker = pruning.NewPersistersTracker(args.EpochsData)
				args.PathManager = &testscommon.PathManagerStub{
					PathForEpochCalled: func(_ string, epoch uint32, _ string) string {
						return fmt.Sprint(epoch)
					},
				}
				reads := 0
				persisters := make(map[string]storage.Persister)
				args.PersisterFactory = &mock.PersisterFactoryStub{
					CreateCalled: func(path string) (storage.Persister, error) {
						if persisters[path] == nil {
							persisters[path] = &countedEpochPersister{Persister: database.NewMemDB(), reads: &reads}
						}
						return persisters[path], nil
					},
				}
				var store interface {
					storage.Storer
					GetFromEpochOrNext([]byte, uint32) ([]byte, error)
				}
				var err error
				if fullHistory {
					store, err = pruning.NewFullHistoryPruningStorer(pruning.FullHistoryStorerArgs{
						StorerArgs: args, NumOfOldActivePersisters: 4,
					})
				} else {
					store, err = pruning.NewPruningStorer(args)
				}
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, store.Close()) })
				key, value := []byte("hash"), []byte("result")
				writeEpoch := uint32(10)
				switch scenario {
				case "next epoch", "cache":
					writeEpoch = 11
				case "outside range":
					writeEpoch = 12
				}
				if scenario != "missing" {
					require.NoError(t, store.PutInEpoch(key, value, writeEpoch))
				}
				if scenario != "cache" {
					store.ClearCache()
				}
				reads = 0
				actual, err := store.GetFromEpochOrNext(key, 10)
				if scenario == "missing" || scenario == "outside range" {
					require.Error(t, err)
					require.Nil(t, actual)
					require.Equal(t, 2, reads)
				} else {
					require.NoError(t, err)
					require.Equal(t, value, actual)
					expectedReads := 1
					if scenario == "next epoch" || scenario == "cache" {
						expectedReads = 2
					}
					if scenario == "cache" {
						expectedReads = 0
					}
					require.Equal(t, expectedReads, reads)

					reads = 0
					actual, err = store.GetFromEpochOrNext(key, 10)
					require.NoError(t, err)
					require.Equal(t, value, actual)
					require.Zero(t, reads, "successful reads must warm the shared cache")
				}
			})
		}
	}
}
