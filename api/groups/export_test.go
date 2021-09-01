package groups

import "github.com/ElrondNetwork/elrond-go/process"

const ExecManualTrigger = execManualTrigger
const ExecBroadcastTrigger = execBroadcastTrigger

// CreateSCQuery -
func (vvg *vmValuesGroup) CreateSCQuery(request *VMValueRequest) (*process.SCQuery, error) {
	return vvg.createSCQuery(request)
}
