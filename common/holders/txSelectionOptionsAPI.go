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

func (selectionOptionsAPI *txSelectionOptionsAPI) GetWithSender() bool {
	return selectionOptionsAPI.withSender
}

func (selectionOptionsAPI *txSelectionOptionsAPI) GetWithRelayer() bool {
	return selectionOptionsAPI.withRelayer
}

func (selectionOptionsAPI *txSelectionOptionsAPI) GetWithNonce() bool {
	return selectionOptionsAPI.withNonce
}
