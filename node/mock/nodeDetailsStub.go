package mock

type NodeDetailsStub struct {
	DetailsMapCalled     func() (map[string]interface{}, error)
	IsInterfaceNilCalled func() bool
}

func (nds *NodeDetailsStub) DetailsMap() (map[string]interface{}, error) {
	return nds.DetailsMapCalled()
}

func (nds *NodeDetailsStub) IsInterfaceNil() bool {
	if nds == nil {
		return true
	}
	return false
}
