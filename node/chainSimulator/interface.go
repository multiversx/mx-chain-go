package chainSimulator

type ChainHandler interface {
	ProcessBlock(nonce uint64, round uint64) error
	IsInterfaceNil() bool
}
