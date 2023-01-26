package processor

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/process"
)

// ArgValidatorInfoInterceptorProcessor is the argument structure used to create a new validator info interceptor processor
type ArgValidatorInfoInterceptorProcessor struct {
	ValidatorInfoPool dataRetriever.ShardedDataCacherNotifier
}

type validatorInfoInterceptorProcessor struct {
	validatorInfoPool dataRetriever.ShardedDataCacherNotifier
}

// NewValidatorInfoInterceptorProcessor creates a new validator info interceptor processor
func NewValidatorInfoInterceptorProcessor(args ArgValidatorInfoInterceptorProcessor) (*validatorInfoInterceptorProcessor, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &validatorInfoInterceptorProcessor{
		validatorInfoPool: args.ValidatorInfoPool,
	}, nil
}

func checkArgs(args ArgValidatorInfoInterceptorProcessor) error {
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

	log.Trace("validatorInfoInterceptorProcessor.Save", "tx hash", hash, "pk", validatorInfo.PublicKey)

	strCache := process.ShardCacherIdentifier(core.MetachainShardId, core.AllShardId)
	viip.validatorInfoPool.AddData(hash, validatorInfo, validatorInfo.Size(), strCache)

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
