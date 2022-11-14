package requesters

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

// ArgHeaderRequester is the argument structure used to create a new header requester instance
type ArgHeaderRequester struct {
	ArgBaseRequester
	NonceConverter typeConverters.Uint64ByteSliceConverter
}

type headerRequester struct {
	*baseRequester
	nonceConverter typeConverters.Uint64ByteSliceConverter
}

// NewHeaderRequester returns a new instance of header requester
func NewHeaderRequester(args ArgHeaderRequester) (*headerRequester, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &headerRequester{
		baseRequester:  createBaseRequester(args.ArgBaseRequester),
		nonceConverter: args.NonceConverter,
	}, nil
}

func checkArgs(args ArgHeaderRequester) error {
	err := checkArgBase(args.ArgBaseRequester)
	if err != nil {
		return err
	}
	if check.IfNil(args.NonceConverter) {
		return dataRetriever.ErrNilUint64ByteSliceConverter
	}

	return nil
}

// RequestDataFromNonce requests a header from other peers by having input the hdr nonce
func (requester *headerRequester) RequestDataFromNonce(nonce uint64, epoch uint32) error {
	return requester.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.NonceType,
			Value: requester.nonceConverter.ToByteSlice(nonce),
			Epoch: epoch,
		},
		[][]byte{requester.nonceConverter.ToByteSlice(nonce)},
	)
}

// RequestDataFromEpoch requests a header from other peers having input the epoch
func (requester *headerRequester) RequestDataFromEpoch(identifier []byte) error {
	return requester.SendOnRequestTopic(
		&dataRetriever.RequestData{
			Type:  dataRetriever.EpochType,
			Value: identifier,
		},
		[][]byte{identifier},
	)
}

// IsInterfaceNil returns true if there is no value under the interface
func (requester *headerRequester) IsInterfaceNil() bool {
	return requester == nil
}
