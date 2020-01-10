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
	stateSyncer       update.StateSyncer
}

// json marshalizer unmarshal all the trie data and write somewhere the new data

func NewStateExporter() (*stateExport, error) {
	return nil, nil
}

func (se *stateExport) ExportAll() error {
	return nil
}

func (se *stateExport) ExportShard() error {
	return nil
}

func (se *stateExport) ExportMeta() error {
	return nil
}
