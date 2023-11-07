package chainSimulator

// ChainHandler defines what a chain handler should be able to do
type ChainHandler interface {
	CreateNewBlock() error
	IsInterfaceNil() bool
}
