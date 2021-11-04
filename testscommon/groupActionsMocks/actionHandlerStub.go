package groupActionsMocks

import "github.com/ElrondNetwork/elrond-go/groupActions/groupTypes"

// ActionHandlerStub -
type ActionHandlerStub struct {
	HandleActionCalled func(triggerData interface{}, stage groupTypes.TriggerStage) error
}

// HandleAction -
func (ahs *ActionHandlerStub) HandleAction(triggerData interface{}, stage groupTypes.TriggerStage) error {
	if ahs.HandleActionCalled != nil {
		return ahs.HandleActionCalled(triggerData, stage)
	}
	return nil
}

// IsInterfaceNil -
func (ahs *ActionHandlerStub) IsInterfaceNil() bool {
	return ahs == nil
}
