package holders

type txSelectionOptionsAPI struct {
	*txSelectionOptions
	withRelayer bool
	withSender  bool
	withNonce   bool
}

// NewTxSelectionOptionsAPI returns a new instance of a selectionOptionsAPI struct
func NewTxSelectionOptionsAPI(options *txSelectionOptions, withSender bool, withRelayer bool, withNonce bool) *txSelectionOptionsAPI {
	return &txSelectionOptionsAPI{
		txSelectionOptions: options,
		withRelayer:        withRelayer,
		withSender:         withSender,
		withNonce:          withNonce,
	}
}

// GetWithSender returns a selection query parameter for the selection simulation endpoint (the sender of the transaction)
func (selectionOptionsAPI *txSelectionOptionsAPI) GetWithSender() bool {
	return selectionOptionsAPI.withSender
}

// GetWithRelayer returns a selection query parameter for the selection simulation endpoint (the relayer of the transaction)
func (selectionOptionsAPI *txSelectionOptionsAPI) GetWithRelayer() bool {
	return selectionOptionsAPI.withRelayer
}

// GetWithNonce returns a selection query parameter for the selection simulation endpoint (the nonce of the transaction)
func (selectionOptionsAPI *txSelectionOptionsAPI) GetWithNonce() bool {
	return selectionOptionsAPI.withNonce
}
