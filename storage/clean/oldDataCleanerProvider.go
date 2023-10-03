package clean

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/storage"
)

// ArgOldDataCleanerProvider is the argument used to create a new oldDataCleanerProvider instance
type ArgOldDataCleanerProvider struct {
	NodeTypeProvider    NodeTypeProviderHandler
	PruningStorerConfig config.StoragePruningConfig
	ManagedPeersHolder  storage.ManagedPeersHolder
}

type oldDataCleanerProvider struct {
	nodeTypeProvider            NodeTypeProviderHandler
	managedPeersHolder          storage.ManagedPeersHolder
	validatorCleanOldEpochsData bool
	observerCleanOldEpochsData  bool
}

// NewOldDataCleanerProvider returns a new instance of oldDataCleanerProvider
func NewOldDataCleanerProvider(args ArgOldDataCleanerProvider) (*oldDataCleanerProvider, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &oldDataCleanerProvider{
		nodeTypeProvider:            args.NodeTypeProvider,
		validatorCleanOldEpochsData: args.PruningStorerConfig.ValidatorCleanOldEpochsData,
		observerCleanOldEpochsData:  args.PruningStorerConfig.ObserverCleanOldEpochsData,
		managedPeersHolder:          args.ManagedPeersHolder,
	}, nil
}

func checkArgs(args ArgOldDataCleanerProvider) error {
	if check.IfNil(args.NodeTypeProvider) {
		return storage.ErrNilNodeTypeProvider
	}
	if check.IfNil(args.ManagedPeersHolder) {
		return storage.ErrNilManagedPeersHolder
	}

	return nil
}

// ShouldClean returns true if old data can be cleaned, based on current configuration,
func (odcp *oldDataCleanerProvider) ShouldClean() bool {
	nodeType := odcp.nodeTypeProvider.GetType()
	shouldClean := false

	if nodeType == core.NodeTypeValidator {
		shouldClean = odcp.validatorCleanOldEpochsData
	}

	if nodeType == core.NodeTypeObserver {
		shouldClean = odcp.observerCleanOldEpochsData
	}

	isMultiKey := odcp.managedPeersHolder.IsMultiKeyMode()
	if isMultiKey {
		shouldClean = odcp.validatorCleanOldEpochsData
	}

	log.Debug("oldDataCleanerProvider.ShouldClean", "node type", nodeType, "is multi key", isMultiKey, "value", shouldClean)

	return shouldClean
}

// IsInterfaceNil returns true if there is no value under the interface
func (odcp *oldDataCleanerProvider) IsInterfaceNil() bool {
	return odcp == nil
}
