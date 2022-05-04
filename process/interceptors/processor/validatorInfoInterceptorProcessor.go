package processor

import (
	"strconv"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/storage"
)

const (
	epochBase = 10
	epochSize = 32
)

// ArgValidatorInfoInterceptorProcessor is the argument structure used to create a new validator info interceptor processor
type ArgValidatorInfoInterceptorProcessor struct {
	Marshaller           marshal.Marshalizer
	ValidatorInfoPool    storage.Cacher
	ValidatorInfoStorage storage.Storer
}

type validatorInfoInterceptorProcessor struct {
	marshaller           marshal.Marshalizer
	validatorInfoPool    storage.Cacher
	validatorInfoStorage storage.Storer
}

// NewValidatorInfoInterceptorProcessor creates a new validator info interceptor processor
func NewValidatorInfoInterceptorProcessor(args ArgValidatorInfoInterceptorProcessor) (*validatorInfoInterceptorProcessor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &validatorInfoInterceptorProcessor{
		marshaller:           args.Marshaller,
		validatorInfoPool:    args.ValidatorInfoPool,
		validatorInfoStorage: args.ValidatorInfoStorage,
	}, nil
}

func checkArgs(args ArgValidatorInfoInterceptorProcessor) error {
	if check.IfNil(args.Marshaller) {
		return process.ErrNilMarshalizer
	}
	if check.IfNil(args.ValidatorInfoPool) {
		return process.ErrNilValidatorInfoPool
	}
	if check.IfNil(args.ValidatorInfoStorage) {
		return process.ErrNilValidatorInfoStorage
	}

	return nil
}

// Validate returns nil as validation is done on Save
func (viip *validatorInfoInterceptorProcessor) Validate(_ process.InterceptedData, _ core.PeerID) error {
	return nil
}

// Save will save the intercepted validator info into the cache
func (viip *validatorInfoInterceptorProcessor) Save(data process.InterceptedData, _ core.PeerID, epoch string) error {
	ivi, ok := data.(interceptedValidatorInfo)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	validatorInfo := ivi.ValidatorInfo()
	hash := ivi.Hash()

	viip.validatorInfoPool.HasOrAdd(hash, validatorInfo, validatorInfo.Size())

	return viip.updateStorage(hash, validatorInfo, epoch)
}

func (viip *validatorInfoInterceptorProcessor) updateStorage(hash []byte, validatorInfo state.ValidatorInfo, epoch string) error {
	buff, err := viip.marshaller.Marshal(&validatorInfo)
	if err != nil {
		return err
	}

	epochUint, err := strconv.ParseUint(epoch, epochBase, epochSize)
	if err != nil {
		return err
	}

	return viip.validatorInfoStorage.PutInEpoch(hash, buff, uint32(epochUint))
}

// RegisterHandler registers a callback function to be notified of incoming validator info
func (viip *validatorInfoInterceptorProcessor) RegisterHandler(_ func(topic string, hash []byte, data interface{})) {
	log.Error("validatorInfoInterceptorProcessor.RegisterHandler", "error", "not implemented")
}

// IsInterfaceNil returns true if there is no value under the interface
func (viip *validatorInfoInterceptorProcessor) IsInterfaceNil() bool {
	return viip == nil
}
