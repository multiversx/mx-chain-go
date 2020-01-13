package factory

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data/trie"
	"github.com/ElrondNetwork/elrond-go/update"
)

type ArgsNewTrieSyncersContainerFactory struct {
}

type trieSyncersContainerFactory struct {
}

func NewTrieSyncersContainerFactory() (*trieSyncersContainerFactory, error) {
	return nil, nil
}

func (t *trieSyncersContainerFactory) Create() (update.TrieSyncContainer, error) {
	return nil, nil
}

func (t *trieSyncersContainerFactory) createOneTrieSyncer() (update.TrieSyncer, error) {
	dataTrie, err := trie.NewTrie(t.trieStorage, t.marshalizer, t.hasher)
	if err != nil {
		return nil, err
	}

	trieSyncer, err := trie.NewTrieSyncer(t.resolver, t.nodesCacher, dataTrie, time.Minute)
	if err != nil {
		return nil, err
	}

	return trieSyncer, nil
}

func (t *trieSyncersContainerFactory) IsInterfaceNil() bool {
	return t == nil
}
