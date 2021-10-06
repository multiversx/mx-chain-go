package slash

import (
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
)

type multipleSigningProof struct {
	slashableHeaders map[string]slashingHeaders
	pubKeys          [][]byte
}

// NewMultipleSigningProof - creates a new multiple signing proof, which contains a list of
// validators which signed multiple headers
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

// GetType -returns MultipleSigning
func (msp *multipleSigningProof) GetType() SlashingType {
	return MultipleSigning
}

// GetLevel - returns the slashing proof level for a given validator, if exists, Level0 otherwise
func (msp *multipleSigningProof) GetLevel(pubKey []byte) SlashingLevel {
	if _, exists := msp.slashableHeaders[string(pubKey)]; exists {
		return msp.slashableHeaders[string(pubKey)].slashingLevel
	}
	return Level0
}

// GetHeaders - returns the slashing proofs headers for a given validator, if exists, nil otherwise
func (msp *multipleSigningProof) GetHeaders(pubKey []byte) []*interceptedBlocks.InterceptedHeader {
	if _, exists := msp.slashableHeaders[string(pubKey)]; exists {
		return msp.slashableHeaders[string(pubKey)].headers
	}
	return nil
}

// GetPubKeys - returns all validator's public keys which signed multiple headers
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
