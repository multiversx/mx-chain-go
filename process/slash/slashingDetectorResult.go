package slash

import "github.com/ElrondNetwork/elrond-go/process"

// slashingDetectorResult - represents the result of a slash detection and contains the duplicate data
// that can be double proposals or double signatures
type slashingDetectorResult struct {
	message SlashingType
	data1   process.InterceptedData
	data2   process.InterceptedData
}

// NewSlashingDetectorResult - creates a new slashingDetectorResult containing data required to generate a slashing proof
func NewSlashingDetectorResult(
	message SlashingType,
	data1 process.InterceptedData,
	data2 process.InterceptedData,
) SlashingDetectorResultHandler {
	return &slashingDetectorResult{
		message: message,
		data1:   data1,
		data2:   data2,
	}
}

// GetType - gets the result type of the current slash detection
func (sdr *slashingDetectorResult) GetType() SlashingType {
	return sdr.message
}

// GetData1 - gets the result of the first duplicate data of the slash detection
func (sdr *slashingDetectorResult) GetData1() process.InterceptedData {
	return sdr.data1
}

// GetData2 - gets the result of the second duplicate data of the slash detection
func (sdr *slashingDetectorResult) GetData2() process.InterceptedData {
	return sdr.data2
}
