package slash

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
)

type DataWithSlashingLevel struct {
	SlashingLevel SlashingLevel
	Data          []process.InterceptedData
}

type headersWithSlashingLevel struct {
	slashingLevel SlashingLevel
	headers       []*interceptedBlocks.InterceptedHeader
}

type multipleHeaderSigningProof struct {
	slashableHeaders map[string]headersWithSlashingLevel
	pubKeys          [][]byte
}

func NewMultipleHeaderSigningProof(
	headersWithSlashing map[string]DataWithSlashingLevel,
) (MultipleSigningProofHandler, error) {
	slashableHeaders, pubKeys, err := convertData(headersWithSlashing)
	if err != nil {
		return nil, err
	}

	return &multipleHeaderSigningProof{
		pubKeys:          pubKeys,
		slashableHeaders: slashableHeaders,
	}, nil
}

// GetType - gets the slashing proofs type
func (msp *multipleHeaderSigningProof) GetType() SlashingType {
	return MultipleSigning
}

// GetLevel - gets the slashing proofs level
func (msp *multipleHeaderSigningProof) GetLevel(pubKey []byte) SlashingLevel {
	if _, exists := msp.slashableHeaders[string(pubKey)]; exists {
		return msp.slashableHeaders[string(pubKey)].slashingLevel
	}
	return Level0
}

func (msp *multipleHeaderSigningProof) GetHeaders(pubKey []byte) []*interceptedBlocks.InterceptedHeader {
	if _, exists := msp.slashableHeaders[string(pubKey)]; exists {
		return msp.slashableHeaders[string(pubKey)].headers
	}
	return nil
}

func (msp *multipleHeaderSigningProof) GetPubKeys() [][]byte {
	return msp.pubKeys
}

func convertData(data map[string]DataWithSlashingLevel) (map[string]headersWithSlashingLevel, [][]byte, error) {
	slashableHeaders := make(map[string]headersWithSlashingLevel)
	pubKeys := make([][]byte, 0, len(data))
	idx := uint64(0)

	for pubKey, slashableData := range data {
		headers, err := convertInterceptedDataToHeader(slashableData.Data)
		if err != nil {
			return nil, nil, err
		}

		slashableHeaders[pubKey] = headersWithSlashingLevel{
			slashingLevel: slashableData.SlashingLevel,
			headers:       headers,
		}

		pubKeys[idx] = []byte(pubKey)
		idx++
	}

	return slashableHeaders, pubKeys, nil
}
