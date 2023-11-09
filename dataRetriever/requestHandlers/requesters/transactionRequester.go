package requesters

// ArgTransactionRequester is the argument structure used to create a new transaction requester instance
type ArgTransactionRequester struct {
	ArgBaseRequester
}

type transactionRequester struct {
	*baseRequester
}

// NewTransactionRequester returns a new instance of transaction requester
func NewTransactionRequester(args ArgTransactionRequester) (*transactionRequester, error) {
	err := checkArgBase(args.ArgBaseRequester)
	if err != nil {
		return nil, err
	}

	return &transactionRequester{
		baseRequester: createBaseRequester(args.ArgBaseRequester),
	}, nil
}

// RequestDataFromHashArray requests a list of tx hashes from other peers
func (requester *transactionRequester) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return requester.requestDataFromHashArray(hashes, epoch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (requester *transactionRequester) IsInterfaceNil() bool {
	return requester == nil
}
