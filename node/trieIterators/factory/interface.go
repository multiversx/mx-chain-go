package factory

import (
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/trieIterators"
)

type DelegatedListHandler interface {
	CreateDelegatedListHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DelegatedListHandler, error)
	IsInterfaceNil() bool
}

type DirectStakedListHandler interface {
	CreateDirectStakedListHandler(args trieIterators.ArgTrieIteratorProcessor) (external.DirectStakedListHandler, error)
	IsInterfaceNil() bool
}
