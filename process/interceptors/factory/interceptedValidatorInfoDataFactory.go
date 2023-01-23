package factory

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/peer"
)

type interceptedValidatorInfoDataFactory struct {
	marshaller marshal.Marshalizer
	hasher     hashing.Hasher
}

// NewInterceptedValidatorInfoDataFactory creates an instance of interceptedValidatorInfoDataFactory
func NewInterceptedValidatorInfoDataFactory(args ArgInterceptedDataFactory) (*interceptedValidatorInfoDataFactory, error) {
	err := checkInterceptedValidatorInfoDataFactoryArgs(args)
	if err != nil {
		return nil, err
	}

	return &interceptedValidatorInfoDataFactory{
		marshaller: args.CoreComponents.InternalMarshalizer(),
		hasher:     args.CoreComponents.Hasher(),
	}, nil
}

func checkInterceptedValidatorInfoDataFactoryArgs(args ArgInterceptedDataFactory) error {
	if check.IfNil(args.CoreComponents) {
		return process.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.CoreComponents.InternalMarshalizer()) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.CoreComponents.Hasher()) {
		return process.ErrNilHasher
	}

	return nil
}

// Create creates instances of InterceptedData by unmarshalling provided buffer
func (ividf *interceptedValidatorInfoDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	args := peer.ArgInterceptedValidatorInfo{
		DataBuff:    buff,
		Marshalizer: ividf.marshaller,
		Hasher:      ividf.hasher,
	}

	return peer.NewInterceptedValidatorInfo(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ividf *interceptedValidatorInfoDataFactory) IsInterfaceNil() bool {
	return ividf == nil
}
