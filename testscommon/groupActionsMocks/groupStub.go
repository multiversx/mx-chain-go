package groupActionsMocks

import (
	"github.com/ElrondNetwork/elrond-go/groupActions/groupTypes"
)

// GroupActionStub -
type GroupActionStub struct {
	HandleActionCalled func(triggerData interface{}, stage groupTypes.TriggerStage) error
	AddToGroupCalled   func(member groupTypes.ActionHandler) error
	GroupIDCalled      func() string
}

// HandleAction -
func (gas *GroupActionStub) HandleAction(triggerData interface{}, stage groupTypes.TriggerStage) error {
	if gas.HandleActionCalled != nil {
		return gas.HandleActionCalled(triggerData, stage)
	}

	return nil
}

// AddToGroup -
func (gas *GroupActionStub) Add(member groupTypes.ActionHandler) error {
	if gas.AddToGroupCalled != nil {
		return gas.AddToGroupCalled(member)
	}
	return nil
}

// GroupID -
func (gas *GroupActionStub) ID() string {
	if gas.GroupIDCalled != nil {
		return gas.GroupIDCalled()
	}
	return "groupID"
}

// IsInterfaceNil -
func (gas *GroupActionStub) IsInterfaceNil() bool {
	return gas == nil
}
