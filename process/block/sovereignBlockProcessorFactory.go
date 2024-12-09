package block

import (
	"errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing/factory"
	mxErrors "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
	processFactory "github.com/multiversx/mx-chain-go/process/factory"
)

type sovereignBlockProcessorFactory struct {
	shardBlockProcessorFactory BlockProcessorCreator
}

// NewSovereignBlockProcessorFactory creates a new sovereign block processor factory
func NewSovereignBlockProcessorFactory(sbpf BlockProcessorCreator) (*sovereignBlockProcessorFactory, error) {
	if check.IfNil(sbpf) {
		return nil, mxErrors.ErrNilBlockProcessorFactory
	}

	return &sovereignBlockProcessorFactory{
		shardBlockProcessorFactory: sbpf,
	}, nil
}

// CreateBlockProcessor creates a new sovereign block processor for the chain run type sovereign
func (s *sovereignBlockProcessorFactory) CreateBlockProcessor(argumentsBaseProcessor ArgBaseProcessor, argsMetaProcessorCreateFunc ExtraMetaBlockProcessorCreateFunc) (process.DebuggerBlockProcessor, error) {
	sp, err := s.shardBlockProcessorFactory.CreateBlockProcessor(argumentsBaseProcessor, argsMetaProcessorCreateFunc)
	if err != nil {
		return nil, errors.New("could not create shard block processor: " + err.Error())
	}

	shardProc, ok := sp.(*shardProcessor)
	if !ok {
		return nil, mxErrors.ErrWrongTypeAssertion
	}

	outgoingOpFormatter, err := sovereign.CreateOutgoingOperationsFormatter(
		argumentsBaseProcessor.Config.SovereignConfig.OutgoingSubscribedEvents.SubscribedEvents,
		argumentsBaseProcessor.CoreComponents.AddressPubKeyConverter(),
		argumentsBaseProcessor.RunTypeComponents.DataCodecHandler(),
		argumentsBaseProcessor.RunTypeComponents.TopicsCheckerHandler())
	if err != nil {
		return nil, err
	}

	operationsHasher, err := factory.NewHasher(argumentsBaseProcessor.Config.SovereignConfig.OutGoingBridge.Hasher)
	if err != nil {
		return nil, err
	}

	systemVM, err := argumentsBaseProcessor.VmContainer.Get(processFactory.SystemVirtualMachine)
	if err != nil {
		return nil, err
	}

	argsMetaProcessor, err := argsMetaProcessorCreateFunc(systemVM)
	if err != nil {
		return nil, err
	}

	args := ArgsSovereignChainBlockProcessor{
		ShardProcessor:                  shardProc,
		ValidatorStatisticsProcessor:    argumentsBaseProcessor.ValidatorStatisticsProcessor,
		OutgoingOperationsFormatter:     outgoingOpFormatter,
		OutGoingOperationsPool:          argumentsBaseProcessor.RunTypeComponents.OutGoingOperationsPoolHandler(),
		OperationsHasher:                operationsHasher,
		ValidatorInfoCreator:            argsMetaProcessor.EpochValidatorInfoCreator,
		EpochRewardsCreator:             argsMetaProcessor.EpochRewardsCreator,
		EpochStartDataCreator:           argsMetaProcessor.EpochStartDataCreator,
		EpochSystemSCProcessor:          argsMetaProcessor.EpochSystemSCProcessor,
		SCToProtocol:                    argsMetaProcessor.SCToProtocol,
		EpochEconomics:                  argsMetaProcessor.EpochEconomics,
		MainChainNotarizationStartRound: argumentsBaseProcessor.Config.SovereignConfig.MainChainNotarization.MainChainNotarizationStartRound,
	}

	return NewSovereignChainBlockProcessor(args)
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *sovereignBlockProcessorFactory) IsInterfaceNil() bool {
	return s == nil
}
