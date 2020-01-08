package factory

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/update"
)

type ArgsNewSyncHandlerFactory struct {
	ShardCoordinator sharding.Coordinator
	Hasher           hashing.Hasher
	Marshalizer      marshal.Marshalizer
}

type stateSyncHandlerFactory struct {
}

func NewSyncHandlerFactory() (*stateSyncHandlerFactory, error) {
	return nil, nil
}

func (shf *stateSyncHandlerFactory) SyncResolversContainer() dataRetriever.ResolversContainer {
	return nil
}

func (shf *stateSyncHandlerFactory) AccountsDBContainer() update.AccountsHandlerContainer {
	return nil
}
