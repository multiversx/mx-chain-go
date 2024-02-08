package components

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	nodeFactory "github.com/multiversx/mx-chain-go/cmd/node/factory"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/factory"
	bootstrapComp "github.com/multiversx/mx-chain-go/factory/bootstrap"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

// ArgsBootstrapComponentsHolder will hold the components needed for the bootstrap components holders
type ArgsBootstrapComponentsHolder struct {
	CoreComponents       factory.CoreComponentsHolder
	CryptoComponents     factory.CryptoComponentsHolder
	NetworkComponents    factory.NetworkComponentsHolder
	StatusCoreComponents factory.StatusCoreComponentsHolder
	WorkingDir           string
	FlagsConfig          config.ContextFlagsConfig
	ImportDBConfig       config.ImportDbConfig
	PrefsConfig          config.Preferences
	Config               config.Config
	ShardIDStr           string
}

type bootstrapComponentsHolder struct {
	closeHandler            *closeHandler
	epochStartBootstrapper  factory.EpochStartBootstrapper
	epochBootstrapParams    factory.BootstrapParamsHolder
	nodeType                core.NodeType
	shardCoordinator        sharding.Coordinator
	versionedHeaderFactory  nodeFactory.VersionedHeaderFactory
	headerVersionHandler    nodeFactory.HeaderVersionHandler
	headerIntegrityVerifier nodeFactory.HeaderIntegrityVerifierHandler
	guardedAccountHandler   process.GuardedAccountHandler
}

// CreateBootstrapComponents will create a new instance of bootstrap components holder
func CreateBootstrapComponents(args ArgsBootstrapComponentsHolder) (factory.BootstrapComponentsHandler, error) {
	instance := &bootstrapComponentsHolder{
		closeHandler: NewCloseHandler(),
	}

	args.PrefsConfig.Preferences.DestinationShardAsObserver = args.ShardIDStr

	bootstrapComponentsFactoryArgs := bootstrapComp.BootstrapComponentsFactoryArgs{
		Config:               args.Config,
		PrefConfig:           args.PrefsConfig,
		ImportDbConfig:       args.ImportDBConfig,
		FlagsConfig:          args.FlagsConfig,
		WorkingDir:           args.WorkingDir,
		CoreComponents:       args.CoreComponents,
		CryptoComponents:     args.CryptoComponents,
		NetworkComponents:    args.NetworkComponents,
		StatusCoreComponents: args.StatusCoreComponents,
	}

	bootstrapComponentsFactory, err := bootstrapComp.NewBootstrapComponentsFactory(bootstrapComponentsFactoryArgs)
	if err != nil {
		return nil, fmt.Errorf("NewBootstrapComponentsFactory failed: %w", err)
	}

	managedBootstrapComponents, err := bootstrapComp.NewManagedBootstrapComponents(bootstrapComponentsFactory)
	if err != nil {
		return nil, err
	}

	err = managedBootstrapComponents.Create()
	if err != nil {
		return nil, err
	}

	instance.epochStartBootstrapper = managedBootstrapComponents.EpochStartBootstrapper()
	instance.epochBootstrapParams = managedBootstrapComponents.EpochBootstrapParams()
	instance.nodeType = managedBootstrapComponents.NodeType()
	instance.shardCoordinator = managedBootstrapComponents.ShardCoordinator()
	instance.versionedHeaderFactory = managedBootstrapComponents.VersionedHeaderFactory()
	instance.headerVersionHandler = managedBootstrapComponents.HeaderVersionHandler()
	instance.headerIntegrityVerifier = managedBootstrapComponents.HeaderIntegrityVerifier()
	instance.guardedAccountHandler = managedBootstrapComponents.GuardedAccountHandler()

	instance.collectClosableComponents()

	return instance, nil
}

// EpochStartBootstrapper will return the epoch start bootstrapper
func (b *bootstrapComponentsHolder) EpochStartBootstrapper() factory.EpochStartBootstrapper {
	return b.epochStartBootstrapper
}

// EpochBootstrapParams will return the epoch bootstrap params
func (b *bootstrapComponentsHolder) EpochBootstrapParams() factory.BootstrapParamsHolder {
	return b.epochBootstrapParams
}

// NodeType will return the node type
func (b *bootstrapComponentsHolder) NodeType() core.NodeType {
	return b.nodeType
}

// ShardCoordinator will return the shardCoordinator
func (b *bootstrapComponentsHolder) ShardCoordinator() sharding.Coordinator {
	return b.shardCoordinator
}

// VersionedHeaderFactory will return the versioned header factory
func (b *bootstrapComponentsHolder) VersionedHeaderFactory() nodeFactory.VersionedHeaderFactory {
	return b.versionedHeaderFactory
}

// HeaderVersionHandler will return header version handler
func (b *bootstrapComponentsHolder) HeaderVersionHandler() nodeFactory.HeaderVersionHandler {
	return b.headerVersionHandler
}

// HeaderIntegrityVerifier will return header integrity verifier
func (b *bootstrapComponentsHolder) HeaderIntegrityVerifier() nodeFactory.HeaderIntegrityVerifierHandler {
	return b.headerIntegrityVerifier
}

// GuardedAccountHandler will return guarded account handler
func (b *bootstrapComponentsHolder) GuardedAccountHandler() process.GuardedAccountHandler {
	return b.guardedAccountHandler
}

func (b *bootstrapComponentsHolder) collectClosableComponents() {
	b.closeHandler.AddComponent(b.epochStartBootstrapper)
}

// Close will call the Close methods on all inner components
func (b *bootstrapComponentsHolder) Close() error {
	return b.closeHandler.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (b *bootstrapComponentsHolder) IsInterfaceNil() bool {
	return b == nil
}

// Create will do nothing
func (b *bootstrapComponentsHolder) Create() error {
	return nil
}

// CheckSubcomponents will do nothing
func (b *bootstrapComponentsHolder) CheckSubcomponents() error {
	return nil
}

// String will do nothing
func (b *bootstrapComponentsHolder) String() string {
	return ""
}
