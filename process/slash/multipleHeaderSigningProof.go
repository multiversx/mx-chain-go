package slash

import "github.com/ElrondNetwork/elrond-go/process"

type multipleSigningProof struct {
	slashableHeaders map[string]SlashingResult
	pubKeys          [][]byte
}

// NewMultipleSigningProof - creates a new multiple signing proof, which contains a list of
// validators which signed multiple headers
func NewMultipleSigningProof(slashResult map[string]SlashingResult) (MultipleSigningProofHandler, error) {
	if slashResult == nil {
		return nil, process.ErrNilSlashResult
	}

	return &multipleSigningProof{
		pubKeys:          getPubKeys(slashResult),
		slashableHeaders: slashResult,
	}, nil
}

// GetType - returns MultipleSigning
func (msp *multipleSigningProof) GetType() SlashingType {
	return MultipleSigning
}

// GetLevel - returns the slashing proof level for a given validator, if exists, Low otherwise
func (msp *multipleSigningProof) GetLevel(pubKey []byte) ThreatLevel {
	if _, exists := msp.slashableHeaders[string(pubKey)]; exists {
		return msp.slashableHeaders[string(pubKey)].SlashingLevel
	}
	return Low
}

// GetHeaders - returns the slashing proofs headers for a given validator, if exists, nil otherwise
func (msp *multipleSigningProof) GetHeaders(pubKey []byte) HeaderInfoList {
	if _, exists := msp.slashableHeaders[string(pubKey)]; exists {
		return msp.slashableHeaders[string(pubKey)].Headers
	}
	return nil
}

// GetPubKeys - returns all validator's public keys which signed multiple headers
func (msp *multipleSigningProof) GetPubKeys() [][]byte {
	return msp.pubKeys
}

func getPubKeys(data map[string]SlashingResult) [][]byte {
	pubKeys := make([][]byte, 0, len(data))

	for pubKey := range data {
		pubKeys = append(pubKeys, []byte(pubKey))
	}

	return pubKeys
}
