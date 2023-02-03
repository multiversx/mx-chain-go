package clean

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	"github.com/multiversx/mx-chain-go/storage"
)

// StorageListProviderHandler defines the actions needed for returning all storers
type StorageListProviderHandler interface {
	GetAllStorers() map[dataRetriever.UnitType]storage.Storer
	IsInterfaceNil() bool
}

// EpochStartNotifier defines what a component which will handle registration to epoch start event should do
type EpochStartNotifier interface {
	RegisterHandler(handler epochStart.ActionHandler)
	IsInterfaceNil() bool
}

// OldDataCleanerProvider defines what a component that handles the deletion or keeping of old data should do
type OldDataCleanerProvider interface {
	ShouldClean() bool
	IsInterfaceNil() bool
}

// NodeTypeProviderHandler defines the actions needed for a component that can handle the node type
type NodeTypeProviderHandler interface {
	SetType(nodeType core.NodeType)
	GetType() core.NodeType
	IsInterfaceNil() bool
}
