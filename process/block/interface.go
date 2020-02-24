package block

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type blockProcessor interface {
	CreateNewHeader(round uint64) data.HeaderHandler
}

type stateUpdater interface {
	PruneTrie(rootHash []byte, identifier data.TriePruningIdentifier) error
	CancelPrune(rootHash []byte, identifier data.TriePruningIdentifier)
	SnapshotState(rootHash []byte)
	SetStateCheckpoint(rootHash []byte)
	IsPruningEnabled() bool
}
