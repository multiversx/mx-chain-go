package txpool

// Economics interface contains economics-related functions required by the txpool
type Economics interface {
	MinGasPrice() uint64
}
