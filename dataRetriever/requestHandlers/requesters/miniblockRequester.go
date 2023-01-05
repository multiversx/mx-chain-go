package requesters

// ArgMiniblockRequester is the argument structure used to create a new miniblock requester instance
type ArgMiniblockRequester struct {
	ArgBaseRequester
}

type miniblockRequester struct {
	*baseRequester
}

// NewMiniblockRequester returns a new instance of miniblock requester
func NewMiniblockRequester(args ArgMiniblockRequester) (*miniblockRequester, error) {
	err := checkArgBase(args.ArgBaseRequester)
	if err != nil {
		return nil, err
	}

	return &miniblockRequester{
		baseRequester: createBaseRequester(args.ArgBaseRequester),
	}, nil
}

// RequestDataFromHashArray requests a block body from other peers by having the miniblocks' hashes and the epoch as input
func (requester *miniblockRequester) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return requester.requestDataFromHashArray(hashes, epoch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (requester *miniblockRequester) IsInterfaceNil() bool {
	return requester == nil
}
