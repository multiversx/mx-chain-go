package slash

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
)

// SlashingData contains the slashable data as well as the severity(slashing level)
// for a possible malicious validator
type SlashingData struct {
	SlashingLevel SlashingLevel
	Data          []process.InterceptedData
}

type slashingHeaders struct {
	slashingLevel SlashingLevel
	headers       []*interceptedBlocks.InterceptedHeader
}

func convertInterceptedDataToHeader(data []process.InterceptedData) ([]*interceptedBlocks.InterceptedHeader, error) {
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
