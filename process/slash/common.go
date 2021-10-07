package slash

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
)

// SlashingResult contains the slashable data as well as the severity(slashing level)
// for a possible malicious validator
type SlashingResult struct {
	SlashingLevel ThreatLevel
	Data          []process.InterceptedData
}

type slashingHeaders struct {
	slashingLevel ThreatLevel
	headers       []*interceptedBlocks.InterceptedHeader
}

func convertInterceptedDataToInterceptedHeaders(data []process.InterceptedData) ([]*interceptedBlocks.InterceptedHeader, error) {
	headers := make([]*interceptedBlocks.InterceptedHeader, 0, len(data))

	for _, d := range data {
		header, castOk := d.(*interceptedBlocks.InterceptedHeader)
		if !castOk {
			return nil, process.ErrCannotCastInterceptedDataToHeader
		}
		headers = append(headers, header)
	}

	return headers, nil
}
