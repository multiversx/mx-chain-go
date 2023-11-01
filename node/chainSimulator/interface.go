package chainSimulator

// ChainHandler defines what a chain handler should be able to do
type ChainHandler interface {
	CreateNewBlock(nonce uint64, round uint64) error
	IsInterfaceNil() bool
}
