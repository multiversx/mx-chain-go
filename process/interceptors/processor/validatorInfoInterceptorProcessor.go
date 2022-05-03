package processor

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// ArgValidatorInfoInterceptorProcessor is the argument structure used to create a new validator info interceptor processor
type ArgValidatorInfoInterceptorProcessor struct {
	Marshaller        marshal.Marshalizer
	ValidatorInfoPool storage.Cacher
}

type validatorInfoInterceptorProcessor struct {
	marshaller        marshal.Marshalizer
	validatorInfoPool storage.Cacher
}

// NewValidatorInfoInterceptorProcessor creates a new validator info interceptor processor
func NewValidatorInfoInterceptorProcessor(args ArgValidatorInfoInterceptorProcessor) (*validatorInfoInterceptorProcessor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &validatorInfoInterceptorProcessor{
		marshaller:        args.Marshaller,
		validatorInfoPool: args.ValidatorInfoPool,
	}, nil
}

func checkArgs(args ArgValidatorInfoInterceptorProcessor) error {
	if check.IfNil(args.Marshaller) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.ValidatorInfoPool) {
		return process.ErrNilValidatorInfoPool
	}

	return nil
}

// Validate returns nil as validation is done on Save
func (viip *validatorInfoInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted validator info into the cache
func (viip *validatorInfoInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, _ string) error {
	ivi, ok := data.(interceptedValidatorInfo)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	validatorInfo := ivi.ValidatorInfo()
	hash := ivi.Hash()

	viip.validatorInfoPool.HasOrAdd(hash, validatorInfo, validatorInfo.Size())

	return nil
}

// RegisterHandler registers a callback function to be notified of incoming validator info
func (viip *validatorInfoInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("validatorInfoInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (viip *validatorInfoInterceptorProcessor) IsInterfaceNil() bool {
	return viip == nil
}
