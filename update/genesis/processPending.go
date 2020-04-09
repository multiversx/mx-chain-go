package genesis

// ArgsPendingTransactionProcessor defines the arguments structure
type ArgsPendingTransactionProcessor struct {
}

type pendingProcessor struct {
}

// NewPendingTransactionProcessor creates a pending transaction processor to be used after hardfork import
func NewPendingTransactionProcessor(args ArgsPendingTransactionProcessor) (*pendingProcessor, error) {
	return &pendingProcessor{}, nil
}

// IsInterfaceNil returns true if underlying object is nil
func (p *pendingProcessor) IsInterfaceNil() bool {
	return p == nil
}
