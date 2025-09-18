package holders

type txSelectionOptionsAPI struct {
	*txSelectionOptions
	requestedFields string
}

// NewTxSelectionOptionsAPI returns a new instance of a selectionOptionsAPI struct
func NewTxSelectionOptionsAPI(options *txSelectionOptions, requestedFields string) *txSelectionOptionsAPI {
	return &txSelectionOptionsAPI{
		txSelectionOptions: options,
		requestedFields:    requestedFields,
	}
}

// GetRequestedFields returns a selection query parameter for the selection simulation endpoint (the requested fields of the transaction)
func (selectionOptionsAPI *txSelectionOptionsAPI) GetRequestedFields() string {
	return selectionOptionsAPI.requestedFields
}
