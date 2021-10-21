package slash

type multipleSigningProof struct {
	slashableHeaders map[string]slashingHeaders
	pubKeys          [][]byte
}

// NewMultipleSigningProof - creates a new multiple signing proof, which contains a list of
// validators which signed multiple headers
func NewMultipleSigningProof(
	slashableData map[string]SlashingResult,
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

// GetType - returns MultipleSigning
func (msp *multipleSigningProof) GetType() SlashingType {
	return MultipleSigning
}

// GetLevel - returns the slashing proof level for a given validator, if exists, Low otherwise
func (msp *multipleSigningProof) GetLevel(pubKey []byte) ThreatLevel {
	if _, exists := msp.slashableHeaders[string(pubKey)]; exists {
		return msp.slashableHeaders[string(pubKey)].slashingLevel
	}
	return Low
}

// GetHeaders - returns the slashing proofs headers for a given validator, if exists, nil otherwise
func (msp *multipleSigningProof) GetHeaders(pubKey []byte) HeaderInfoList {
	if _, exists := msp.slashableHeaders[string(pubKey)]; exists {
		return msp.slashableHeaders[string(pubKey)].headers
	}
	return nil
}

// GetPubKeys - returns all validator's public keys which signed multiple headers
func (msp *multipleSigningProof) GetPubKeys() [][]byte {
	return msp.pubKeys
}

func convertData(data map[string]SlashingResult) (map[string]slashingHeaders, [][]byte, error) {
	slashableHeaders := make(map[string]slashingHeaders)
	pubKeys := make([][]byte, 0, len(data))

	for pubKey, slashableData := range data {
		slashableHeaders[pubKey] = slashingHeaders{
			slashingLevel: slashableData.SlashingLevel,
			headers:       slashableData.Data,
		}

		pubKeys = append(pubKeys, []byte(pubKey))
	}

	return slashableHeaders, pubKeys, nil
}
