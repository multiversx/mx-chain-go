package requesters

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

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

// RequestDataFromNonce requests equivalent proofs data from other peers for the specified nonce-shard key
func (requester *equivalentProofsRequester) RequestDataFromNonce(nonceShardKey []byte, epoch uint32) error {
	log.Trace("equivalentProofsRequester.RequestDataFromNonce",
		"nonce-shard", string(nonceShardKey),
		"epoch", epoch,
		"topic", requester.RequestTopic())

	return requester.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.NonceType,
			Value: nonceShardKey,
			Epoch: epoch,
		},
		[][]byte{nonceShardKey},
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (requester *equivalentProofsRequester) IsInterfaceNil() bool {
	return requester == nil
}
