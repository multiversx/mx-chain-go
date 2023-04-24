package requesters

// ArgValidatorInfoRequester is the argument structure used to create a new validator info requester instance
type ArgValidatorInfoRequester struct {
	ArgBaseRequester
}

type validatorInfoRequester struct {
	*baseRequester
}

// NewValidatorInfoRequester returns a new instance of validator info requester
func NewValidatorInfoRequester(args ArgValidatorInfoRequester) (*validatorInfoRequester, error) {
	err := checkArgBase(args.ArgBaseRequester)
	if err != nil {
		return nil, err
	}

	return &validatorInfoRequester{
		baseRequester: createBaseRequester(args.ArgBaseRequester),
	}, nil
}

// RequestDataFromHashArray requests validator info from other peers by hash array
func (requester *validatorInfoRequester) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return requester.requestDataFromHashArray(hashes, epoch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (requester *validatorInfoRequester) IsInterfaceNil() bool {
	return requester == nil
}
