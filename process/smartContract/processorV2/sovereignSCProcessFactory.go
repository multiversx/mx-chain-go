package processorV2

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/smartContract/scrCommon"
)

type sovereignSCProcessFactory struct {
	scrProcessorCreator scrCommon.SCProcessorCreator
}

// NewSovereignSCProcessFactory creates a new smart contract process factory
func NewSovereignSCProcessFactory(creator scrCommon.SCProcessorCreator) (*sovereignSCProcessFactory, error) {
	if check.IfNil(creator) {
		return nil, process.ErrNilSCProcessorCreator
	}
	return &sovereignSCProcessFactory{
		scrProcessorCreator: creator,
	}, nil
}

// CreateSCProcessor creates a new smart contract processor
func (scpf *sovereignSCProcessFactory) CreateSCProcessor(args scrCommon.ArgsNewSmartContractProcessor) (scrCommon.SCRProcessorHandler, error) {
	sp, err := scpf.scrProcessorCreator.CreateSCProcessor(args)
	if err != nil {
		return nil, err
	}

	scProc, ok := sp.(process.SmartContractProcessorFacade)
	if !ok {
		return nil, process.ErrWrongTypeAssertion
	}

	scpHelper, err := scrCommon.NewSCProcessHelper(scrCommon.SCProcessHelperArgs{
		Accounts:         args.AccountsDB,
		ShardCoordinator: args.ShardCoordinator,
		Marshalizer:      args.Marshalizer,
		Hasher:           args.Hasher,
		PubkeyConverter:  args.PubkeyConv,
	})

	if err != nil {
		return nil, err
	}

	return NewSovereignSCRProcessor(SovereignSCProcessArgs{
		ArgsParser:             args.ArgsParser,
		TxTypeHandler:          args.TxTypeHandler,
		SmartContractProcessor: scProc,
		SCProcessHelperHandler: scpHelper,
	})
}

// IsInterfaceNil returns true if there is no value under the interface
func (scpf *sovereignSCProcessFactory) IsInterfaceNil() bool {
	return scpf == nil
}
