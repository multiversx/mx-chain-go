package smartContract

// SCProcessorCreator defines a scr processor creator
type SCProcessorCreator interface {
	CreateSCProcessor(args ArgsNewSmartContractProcessor) (SCRProcessorHandler, error)
	IsInterfaceNil() bool
}
