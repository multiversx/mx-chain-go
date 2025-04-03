package requesters

// ArgEquivalentProofsRequester is the argument structure used to create a new equivalent proofs requester instance
type ArgEquivalentProofsRequester struct {
	ArgBaseRequester
}

type equivalentProofsRequester struct {
	*baseRequester
}

// NewEquivalentProofsRequester returns a new instance of equivalent proofs requester
func NewEquivalentProofsRequester(args ArgEquivalentProofsRequester) (*equivalentProofsRequester, error) {
	err := checkArgBase(args.ArgBaseRequester)
	if err != nil {
		return nil, err
	}

	return &equivalentProofsRequester{
		baseRequester: createBaseRequester(args.ArgBaseRequester),
	}, nil
}

// RequestDataFromHashArray requests equivalent proofs data from other peers by having multiple header hashes and the epoch as input
// all headers must be from the same epoch
func (requester *equivalentProofsRequester) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return requester.requestDataFromHashArray(hashes, epoch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (requester *equivalentProofsRequester) IsInterfaceNil() bool {
	return requester == nil
}
