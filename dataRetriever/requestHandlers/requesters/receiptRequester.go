package requesters

// ArgReceiptRequester is the argument structure used to create a new receipt requester instance
type ArgReceiptRequester struct {
	ArgBaseRequester
}

type receiptRequester struct {
	*baseRequester
}

// NewReceiptRequester returns a new instance of receipt requester
func NewReceiptRequester(args ArgReceiptRequester) (*receiptRequester, error) {
	err := checkArgBase(args.ArgBaseRequester)
	if err != nil {
		return nil, err
	}

	return &receiptRequester{
		baseRequester: createBaseRequester(args.ArgBaseRequester),
	}, nil
}

// RequestDataFromHashArray requests receipt from other peers by hash array
func (requester *receiptRequester) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return requester.requestDataFromHashArray(hashes, epoch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (requester *receiptRequester) IsInterfaceNil() bool {
	return requester == nil
}
