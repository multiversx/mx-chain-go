package clean

import (
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type oldDataCleanerProvider struct {
	nodeTypeProvider            sharding.NodeTypeProviderHandler
	validatorCleanOldEpochsData bool
	observerCleanOldEpochsData  bool
}

// NewOldDataCleanerProvider returns a new instance of oldDataCleanerProvider
func NewOldDataCleanerProvider(
	nodeTypeProvider sharding.NodeTypeProviderHandler,
	pruningStorerConfig config.StoragePruningConfig,
) (*oldDataCleanerProvider, error) {
	if check.IfNil(nodeTypeProvider) {
		return nil, storage.ErrNilNodeTypeProvider
	}
	return &oldDataCleanerProvider{
		nodeTypeProvider:            nodeTypeProvider,
		validatorCleanOldEpochsData: pruningStorerConfig.ValidatorCleanOldEpochsData,
		observerCleanOldEpochsData:  pruningStorerConfig.ObserverCleanOldEpochsData,
	}, nil
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

	// TODO: remove this log before merging
	log.Debug("oldDataCleanerProvider.ShouldClean", "node type", nodeType, "value", shouldClean)

	return shouldClean
}

// IsInterfaceNil returns true if there is no value under the interface
func (odcp *oldDataCleanerProvider) IsInterfaceNil() bool {
	return odcp == nil
}
