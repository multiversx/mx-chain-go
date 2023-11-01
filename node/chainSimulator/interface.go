package chainSimulator

type ChainHandler interface {
	ProcessBlock() error
	IsInterfaceNil() bool
}
