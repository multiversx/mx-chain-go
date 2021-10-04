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
