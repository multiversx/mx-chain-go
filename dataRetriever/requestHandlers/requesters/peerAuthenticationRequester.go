package requesters

// ArgPeerAuthenticationRequester is the argument structure used to create a new peer authentication requester instance
type ArgPeerAuthenticationRequester struct {
	ArgBaseRequester
}

type peerAuthenticationRequester struct {
	*baseRequester
}

// NewPeerAuthenticationRequester returns a new instance of peer authentication requester
func NewPeerAuthenticationRequester(args ArgPeerAuthenticationRequester) (*peerAuthenticationRequester, error) {
	err := checkArgBase(args.ArgBaseRequester)
	if err != nil {
		return nil, err
	}

	return &peerAuthenticationRequester{
		baseRequester: createBaseRequester(args.ArgBaseRequester),
	}, nil
}

// RequestDataFromHashArray requests peer authentication data from other peers by having multiple public keys' hashes and the epoch as input
func (requester *peerAuthenticationRequester) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return requester.requestDataFromHashArray(hashes, epoch)
}

// IsInterfaceNil returns true if there is no value under the interface
func (requester *peerAuthenticationRequester) IsInterfaceNil() bool {
	return requester == nil
}
