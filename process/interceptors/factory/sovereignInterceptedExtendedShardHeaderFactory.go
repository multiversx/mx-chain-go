package factory

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/interceptedBlocks"
)

// ArgsSovereignInterceptedExtendedHeaderFactory is a struct placeholder for args needed to create an instance of sovereign extended header factory
type ArgsSovereignInterceptedExtendedHeaderFactory struct {
	Marshaller marshal.Marshalizer
	Hasher     hashing.Hasher
}

type sovereignInterceptedShardHeaderDataFactory struct {
	marshaller marshal.Marshalizer
	hasher     hashing.Hasher
}

// NewSovereignInterceptedShardHeaderDataFactory creates an instance of sovereign extended header factory
func NewSovereignInterceptedShardHeaderDataFactory(args ArgsSovereignInterceptedExtendedHeaderFactory) (*sovereignInterceptedShardHeaderDataFactory, error) {
	if check.IfNil(args.Marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}

	return &sovereignInterceptedShardHeaderDataFactory{
		marshaller: args.Marshaller,
		hasher:     args.Hasher,
	}, nil
}

// Create creates instances of sovereign extended header by unmarshalling provided buffer
func (ishdf *sovereignInterceptedShardHeaderDataFactory) Create(buff []byte) (process.InterceptedData, error) {
	arg := interceptedBlocks.ArgsSovereignInterceptedHeader{
		Marshaller:  ishdf.marshaller,
		Hasher:      ishdf.hasher,
		HeaderBytes: buff,
	}

	return interceptedBlocks.NewSovereignInterceptedHeader(arg)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ishdf *sovereignInterceptedShardHeaderDataFactory) IsInterfaceNil() bool {
	return ishdf == nil
}
