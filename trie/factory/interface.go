package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type coreComponentsHandler interface {
	InternalMarshalizer() marshal.Marshalizer
	Hasher() hashing.Hasher
	PathHandler() storage.PathManagerHandler
}
