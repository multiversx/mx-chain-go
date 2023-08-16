package factory

import (
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/storage"
)

type coreComponentsHandler interface {
	InternalMarshalizer() marshal.Marshalizer
	Hasher() hashing.Hasher
	PathHandler() storage.PathManagerHandler
	ProcessStatusHandler() common.ProcessStatusHandler
	EnableEpochsHandler() common.EnableEpochsHandler
}
