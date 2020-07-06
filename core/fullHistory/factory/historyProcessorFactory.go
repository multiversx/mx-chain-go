package factory

import (
	"github.com/ElrondNetwork/elrond-go/core/fullHistory"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgsHistoryProcessorFactory holds all dependencies required by the history processor factory in order to create
// new instances
type ArgsHistoryProcessorFactory struct {
	IsEnabled   bool
	SelfShardID uint32
	Store       storage.Storer
	Marshalizer marshal.Marshalizer
	Hasher      hashing.Hasher
}

type historyProcessorFactory struct {
	IsEnabled   bool
	SelfShardID uint32
	Store       storage.Storer
	Marshalizer marshal.Marshalizer
	Hasher      hashing.Hasher
}

// NewHistoryProcessorFactory creates an instance of historyProcessorFactory
func NewHistoryProcessorFactory(args *ArgsHistoryProcessorFactory) (fullHistory.HistoryProcessorFactory, error) {
	return &historyProcessorFactory{
		IsEnabled:   args.IsEnabled,
		SelfShardID: args.SelfShardID,
		Store:       args.Store,
		Marshalizer: args.Marshalizer,
		Hasher:      args.Hasher,
	}, nil
}

// Create creates instances of HistoryHandler
func (hpf *historyProcessorFactory) Create() (fullHistory.HistoryHandler, error) {
	if !hpf.IsEnabled {
		return fullHistory.NewNilHistoryProcessor()
	}

	historyProcArgs := fullHistory.HistoryProcessorArguments{
		Hasher:      hpf.Hasher,
		Marshalizer: hpf.Marshalizer,
		Store:       hpf.Store,
		SelfShardID: hpf.SelfShardID,
	}
	return fullHistory.NewHistoryProcessor(historyProcArgs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hpf *historyProcessorFactory) IsInterfaceNil() bool {
	return hpf == nil
}
