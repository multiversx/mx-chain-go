package slash

import (
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
)

type multipleSigningProof struct {
	slashableHeaders map[string]slashingHeaders
	pubKeys          [][]byte
}

func NewMultipleSigningProof(
	slashableData map[string]SlashingData,
) (MultipleSigningProofHandler, error) {
	slashableHeaders, pubKeys, err := convertData(slashableData)
	if err != nil {
		return nil, err
	}

	return &multipleSigningProof{
		pubKeys:          pubKeys,
		slashableHeaders: slashableHeaders,
	}, nil
}

// GetType - gets the slashing proofs type
func (msp *multipleSigningProof) GetType() SlashingType {
	return MultipleSigning
}

// GetLevel - gets the slashing proofs level
func (msp *multipleSigningProof) GetLevel(pubKey []byte) SlashingLevel {
	if _, exists := msp.slashableHeaders[string(pubKey)]; exists {
		return msp.slashableHeaders[string(pubKey)].slashingLevel
	}
	return Level0
}

func (msp *multipleSigningProof) GetHeaders(pubKey []byte) []*interceptedBlocks.InterceptedHeader {
	if _, exists := msp.slashableHeaders[string(pubKey)]; exists {
		return msp.slashableHeaders[string(pubKey)].headers
	}
	return nil
}

func (msp *multipleSigningProof) GetPubKeys() [][]byte {
	return msp.pubKeys
}

func convertData(data map[string]SlashingData) (map[string]slashingHeaders, [][]byte, error) {
	slashableHeaders := make(map[string]slashingHeaders)
	pubKeys := make([][]byte, 0, len(data))

	for pubKey, slashableData := range data {
		headers, err := convertInterceptedDataToHeader(slashableData.Data)
		if err != nil {
			return nil, nil, err
		}

		slashableHeaders[pubKey] = slashingHeaders{
			slashingLevel: slashableData.SlashingLevel,
			headers:       headers,
		}

		pubKeys = append(pubKeys, []byte(pubKey))
	}

	return slashableHeaders, pubKeys, nil
}
