package genesis

import (
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/update"
)

type ArgsNewStateExporter struct {
	TrieSyncHandlers update.TrieSyncContainer
}

type stateExport struct {
	accountsContainer update.AccountsHandlerContainer
	stateStore        storage.Storer
}

// json marshalizer unmarshal all the trie data and write somewhere the new data

func NewStateExporter() (*stateExport, error) {
	return nil, nil
}
