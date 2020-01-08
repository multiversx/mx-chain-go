package genesis

import (
	"github.com/ElrondNetwork/elrond-go/update"
)

type ArgsNewStateExporter struct {
	TrieSyncHandlers update.TrieSyncContainer
}

type stateExport struct {
	accountsContainer update.AccountsHandlerContainer
}

func NewStateExporter() (*stateExport, error) {
	return nil, nil
}
