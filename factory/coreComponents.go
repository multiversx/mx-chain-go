package factory

import (
	"fmt"
	"path/filepath"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	stateFactory "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/hashing"
	factoryHasher "github.com/ElrondNetwork/elrond-go/hashing/factory"
	"github.com/ElrondNetwork/elrond-go/marshal"
	factoryMarshalizer "github.com/ElrondNetwork/elrond-go/marshal/factory"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/ElrondNetwork/elrond-go/storage/pathmanager"
)

// CoreComponentsFactoryArgs holds the arguments needed for creating a core components factory
type CoreComponentsFactoryArgs struct {
	Config           config.Config
	WorkingDirectory string
}

// coreComponentsFactory is responsible for creating the core components
type coreComponentsFactory struct {
	config     config.Config
	workingDir string
}

// coreComponents is the DTO used for core components
type coreComponents struct {
	hasher                   hashing.Hasher
	internalMarshalizer      marshal.Marshalizer
	vmMarshalizer            marshal.Marshalizer
	txSignMarshalizer        marshal.Marshalizer
	uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	addressPubKeyConverter   core.PubkeyConverter
	validatorPubKeyConverter core.PubkeyConverter
	statusHandler            core.AppStatusHandler
	pathHandler              storage.PathManagerHandler
	chainID                  string
	minTransactionVersion    uint32
}

// NewCoreComponentsFactory initializes the factory which is responsible to creating core components
func NewCoreComponentsFactory(args CoreComponentsFactoryArgs) *coreComponentsFactory {
	return &coreComponentsFactory{
		config:     args.Config,
		workingDir: args.WorkingDirectory,
	}
}

// Create creates the core components
func (ccf *coreComponentsFactory) Create() (*coreComponents, error) {
	hasher, err := factoryHasher.NewHasher(ccf.config.Hasher.Type)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrHasherCreation, err.Error())
	}

	internalMarshalizer, err := factoryMarshalizer.NewMarshalizer(ccf.config.Marshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (internal): %s", ErrMarshalizerCreation, err.Error())
	}

	vmMarshalizer, err := factoryMarshalizer.NewMarshalizer(ccf.config.VmMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (vm): %s", ErrMarshalizerCreation, err.Error())
	}

	txSignMarshalizer, err := factoryMarshalizer.NewMarshalizer(ccf.config.TxSignMarshalizer.Type)
	if err != nil {
		return nil, fmt.Errorf("%w (tx sign): %s", ErrMarshalizerCreation, err.Error())
	}

	uint64ByteSliceConverter := uint64ByteSlice.NewBigEndianConverter()

	addressPubkeyConverter, err := stateFactory.NewPubkeyConverter(ccf.config.AddressPubkeyConverter)
	if err != nil {
		return nil, fmt.Errorf("%w for AddressPubkeyConverter", err)
	}
	validatorPubkeyConverter, err := stateFactory.NewPubkeyConverter(ccf.config.ValidatorPubkeyConverter)
	if err != nil {
		return nil, fmt.Errorf("%w for AddressPubkeyConverter", err)
	}

	pruningStorerPathTemplate, staticStorerPathTemplate := ccf.createStorerTemplatePaths()
	pathHandler, err := pathmanager.NewPathManager(pruningStorerPathTemplate, staticStorerPathTemplate)
	if err != nil {
		return nil, err
	}

	return &coreComponents{
		hasher:                   hasher,
		internalMarshalizer:      internalMarshalizer,
		vmMarshalizer:            vmMarshalizer,
		txSignMarshalizer:        txSignMarshalizer,
		uint64ByteSliceConverter: uint64ByteSliceConverter,
		addressPubKeyConverter:   addressPubkeyConverter,
		validatorPubKeyConverter: validatorPubkeyConverter,
		statusHandler:            statusHandler.NewNilStatusHandler(),
		pathHandler:              pathHandler,
		chainID:                  ccf.config.GeneralSettings.ChainID,
		minTransactionVersion:    ccf.config.GeneralSettings.MinTransactionVersion,
	}, nil
}

func (ccf *coreComponentsFactory) createStorerTemplatePaths() (string, string) {
	pathTemplateForPruningStorer := filepath.Join(
		ccf.workingDir,
		core.DefaultDBPath,
		ccf.config.GeneralSettings.ChainID,
		fmt.Sprintf("%s_%s", core.DefaultEpochString, core.PathEpochPlaceholder),
		fmt.Sprintf("%s_%s", core.DefaultShardString, core.PathShardPlaceholder),
		core.PathIdentifierPlaceholder)

	pathTemplateForStaticStorer := filepath.Join(
		ccf.workingDir,
		core.DefaultDBPath,
		ccf.config.GeneralSettings.ChainID,
		core.DefaultStaticDbString,
		fmt.Sprintf("%s_%s", core.DefaultShardString, core.PathShardPlaceholder),
		core.PathIdentifierPlaceholder)

	return pathTemplateForPruningStorer, pathTemplateForStaticStorer
}
