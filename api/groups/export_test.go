package groups

import "github.com/multiversx/mx-chain-go/process"

// CreateSCQuery -
func (vvg *vmValuesGroup) CreateSCQuery(request *VMValueRequest) (*process.SCQuery, error) {
	return vvg.createSCQuery(request)
}

// VmValuesFacadeHandler exported the vm values facade handler interface for testing purposes
type VmValuesFacadeHandler interface {
	vmValuesFacadeHandler
}
