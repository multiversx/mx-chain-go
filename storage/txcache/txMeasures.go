package txcache

const estimatedSizeOfBoundedTxFields = uint64(128)

// estimateTxSize returns an approximation
func estimateTxSize(tx *WrappedTransaction) uint64 {
	sizeOfData := uint64(len(tx.Tx.GetData()))
	return estimatedSizeOfBoundedTxFields + sizeOfData
}

// estimateTxGas returns an approximation for the necessary computation units (gas units)
func estimateTxGas(tx *WrappedTransaction) uint64 {
	gasLimit := tx.Tx.GetGasLimit()
	return gasLimit
}

// estimateTxFee returns an approximation for the cost of a transaction, in micro ERD (1/1000000 ERD)
// TODO: switch to integer operations (as opposed to float operations).
// TODO: do not assume the order of magnitude of minGasPrice.
func estimateTxFee(tx *WrappedTransaction) uint64 {
	// In order to obtain the result as micro ERD, we have to divide by 10^12 (since 1 ERD = 10^18 gas currency units)
	// In order to have better precision, we divide each of the factors by 10^6
	gasLimit := float32(tx.Tx.GetGasLimit()) / 1000000
	gasPrice := float32(tx.Tx.GetGasPrice()) / 1000000
	feeInMicroERD := gasLimit * gasPrice
	return uint64(feeInMicroERD)
}
