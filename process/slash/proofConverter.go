package slash

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
)

// ToProtoMultipleHeaderProposal converts a MultipleProposalProofHandler to its specific proto structure
func ToProtoMultipleHeaderProposal(proof MultipleProposalProofHandler) (*coreSlash.MultipleHeaderProposalProof, error) {
	headers, err := getHeadersFromInterceptedHeaders(proof.GetHeaders())
	if err != nil {
		return nil, err
	}

	return &coreSlash.MultipleHeaderProposalProof{
		Level:   coreSlash.ThreatLevel(proof.GetLevel()),
		Headers: headers,
	}, nil
}

// ToProtoMultipleHeaderSign converts a MultipleSigningProofHandler to its specific proto structure
func ToProtoMultipleHeaderSign(proof MultipleSigningProofHandler) (*coreSlash.MultipleHeaderSigningProof, error) {
	levels := make(map[string]coreSlash.ThreatLevel)
	headers := make(map[string]coreSlash.Headers)

	for _, pubKey := range proof.GetPubKeys() {
		currHeaders, err := getHeadersFromInterceptedHeaders(proof.GetHeaders(pubKey))
		if err != nil {
			return nil, err
		}

		headers[string(pubKey)] = currHeaders
		levels[string(pubKey)] = coreSlash.ThreatLevel(proof.GetLevel(pubKey))
	}

	return &coreSlash.MultipleHeaderSigningProof{
		PubKeys: proof.GetPubKeys(),
		Levels:  levels,
		Headers: headers,
	}, nil
}

func getHeadersFromInterceptedHeaders(interceptedHeaders []*interceptedBlocks.InterceptedHeader) (coreSlash.Headers, error) {
	headers := make([]data.HeaderHandler, 0, len(interceptedHeaders))
	ret := coreSlash.Headers{}

	for _, interceptedHeader := range interceptedHeaders {
		if check.IfNil(interceptedHeader) || check.IfNil(interceptedHeader.HeaderHandler()) {
			return ret, process.ErrNilHeaderHandler
		}

		headers = append(headers, interceptedHeader.HeaderHandler())
	}

	err := ret.SetHeaders(headers)
	return ret, err
}
