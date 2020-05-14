package factory

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	stateFactory "github.com/ElrondNetwork/elrond-go/data/state/factory"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters/uint64ByteSlice"
	"github.com/ElrondNetwork/elrond-go/hashing"
	factoryHasher "github.com/ElrondNetwork/elrond-go/hashing/factory"
	"github.com/ElrondNetwork/elrond-go/marshal"
	factoryMarshalizer "github.com/ElrondNetwork/elrond-go/marshal/factory"
	"github.com/ElrondNetwork/elrond-go/statusHandler"
)

// CoreComponentsFactoryArgs holds the arguments needed for creating a core components factory
type CoreComponentsFactoryArgs struct {
	Config  config.Config
	ChainID []byte
}

// CoreComponentsFactory is responsible for creating the core components
type CoreComponentsFactory struct {
	config      config.Config
	chainID     []byte
	marshalizer marshal.Marshalizer
	hasher      hashing.Hasher
}

// coreComponents is the DTO used for core components
type coreComponents struct {
	Hasher                   hashing.Hasher
	InternalMarshalizer      marshal.Marshalizer
	VmMarshalizer            marshal.Marshalizer
	TxSignMarshalizer        marshal.Marshalizer
	Uint64ByteSliceConverter typeConverters.Uint64ByteSliceConverter
	addressPubKeyConverter   state.PubkeyConverter
	validatorPubKeyConverter state.PubkeyConverter
	StatusHandler            core.AppStatusHandler
	ChainID                  []byte
}

// NewCoreComponentsFactory initializes the factory which is responsible to creating core components
func NewCoreComponentsFactory(args CoreComponentsFactoryArgs) *CoreComponentsFactory {
	return &CoreComponentsFactory{
		config:  args.Config,
		chainID: args.ChainID,
	}
}

// Create creates the core components
func (ccf *CoreComponentsFactory) Create() (*coreComponents, error) {
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

	return &coreComponents{
		Hasher:                   hasher,
		InternalMarshalizer:      internalMarshalizer,
		VmMarshalizer:            vmMarshalizer,
		TxSignMarshalizer:        txSignMarshalizer,
		Uint64ByteSliceConverter: uint64ByteSliceConverter,
		addressPubKeyConverter:   addressPubkeyConverter,
		validatorPubKeyConverter: validatorPubkeyConverter,
		StatusHandler:            statusHandler.NewNilStatusHandler(),
		ChainID:                  ccf.chainID,
	}, nil
}
