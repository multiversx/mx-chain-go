package api

// ApiProcessorCompsCreatorHandler should be able to create api process components
type ApiProcessorCompsCreatorHandler interface {
	CreateAPIComps(args ArgsCreateAPIProcessComps) (*APIProcessComps, error)
	IsInterfaceNil() bool
}
