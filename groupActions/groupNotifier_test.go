package groupActions

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	es "github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/groupActions/groupTypes"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochStart"
	"github.com/ElrondNetwork/elrond-go/testscommon/groupActionsMocks"
	"github.com/stretchr/testify/require"
)

func TestNewGroupActionNotifier(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	require.NotNil(t, gan)
	require.Len(t, gan.groupsPerEvent, 0)
	require.Len(t, gan.triggers, 0)
	require.Len(t, gan.activeTriggers, 0)
}

func TestGroupNotifier_addTriggerOK(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()

	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	err := gan.addTrigger(epochStartTrigger)
	require.Len(t, gan.triggers, 1)
	require.Len(t, gan.groupsPerEvent, 1)
	require.Len(t, gan.activeTriggers, 0)

	require.Nil(t, err)
}

func TestGroupNotifier_addTriggerAlreadyPresentOK(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()

	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.addTrigger(epochStartTrigger)
	err := gan.addTrigger(epochStartTrigger)
	require.Len(t, gan.triggers, 1)
	require.Len(t, gan.groupsPerEvent, 1)
	require.Len(t, gan.activeTriggers, 0)
	require.Nil(t, err)
}

func TestGroupNotifier_registerTriggerOK(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}

	err := gan.registerTrigger(epochStartTrigger)
	require.Equal(t, epochStartTrigger, gan.triggers[epochStartTrigger.GetName()])
	require.Len(t, gan.activeTriggers, 1)
	require.Len(t, gan.groupsPerEvent, 1)
	require.Nil(t, err)
}

func TestGroupNotifier_registerTriggerAlreadyPresentOK(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.registerTrigger(epochStartTrigger)

	err := gan.registerTrigger(epochStartTrigger)
	require.Nil(t, err)
}

func TestGroupNotifier_registerTriggerSecond(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	epochStartTrigger1 := &epochStart.EpochStartNotifierStub{
		GetNameCalled: func() string {
			return "event1"
		},
	}

	epochStartTrigger2 := &epochStart.EpochStartNotifierStub{
		GetNameCalled: func() string {
			return "event2"
		},
	}
	_ = gan.registerTrigger(epochStartTrigger1)
	err := gan.registerTrigger(epochStartTrigger2)
	require.Nil(t, err)
	require.Len(t, gan.triggers, 2 )
	require.Equal(t, gan.triggers[epochStartTrigger1.GetName()], epochStartTrigger1)
	require.Equal(t, gan.triggers[epochStartTrigger2.GetName()], epochStartTrigger2)
}

func TestGroupNotifier_Register(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.registerTrigger(epochStartTrigger)
	groupAction, _ := NewGroupWithDefaultLock("grID")

	err := gan.Register(groupAction, epochStartTrigger)
	require.Nil(t, err)
}

func TestGroupNotifier_RegisterNilActionHandler(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}

	err := gan.Register(nil, epochStartTrigger)
	require.Equal(t, errNilGroupActionHandler, err)
}

func TestGroupNotifier_RegisterInvalidTriggerID(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{
		GetNameCalled: func() string {
			return ""
		},
	}

	groupAction, _ := NewGroupWithDefaultLock("grID")
	err := gan.Register(groupAction, epochStartTrigger)
	require.Equal(t, errInvalidTriggerID, err)
}

func TestGroupNotifier_RegisterNilTrigger(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	groupAction, _ := NewGroupWithDefaultLock("grID")
	err := gan.Register(groupAction, nil)
	require.Equal(t, errNilTrigger, err)
}

func TestGroupNotifier_RegisterAlreadyRegisteredGroup(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	groupAction, _ := NewGroupWithDefaultLock("grID")
	_ = gan.Register(groupAction, epochStartTrigger)
	err := gan.Register(groupAction, epochStartTrigger)
	require.Equal(t, errGroupAlreadyRegisteredForTrigger, err)
}

func TestGroupNotifier_RegisterTwoGroupsSameTrigger(t *testing.T) {
	t.Parallel()

	gr1 := "grID1"
	gr2 := "grID2"
	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	group1Action, _ := NewGroupWithDefaultLock("grID1")
	group2Action, _ := NewGroupWithDefaultLock("grID2")

	_ = gan.Register(group1Action, epochStartTrigger)
	err := gan.Register(group2Action, epochStartTrigger)

	require.Nil(t, err)
	require.Len(t, gan.groupsPerEvent[epochStartTrigger.GetName()], 2)
	require.Equal(t, group1Action, gan.groupsPerEvent[epochStartTrigger.GetName()][gr1])
	require.Equal(t, group2Action, gan.groupsPerEvent[epochStartTrigger.GetName()][gr2])
}

func TestGroupNotifier_notifyGroupsForTriggerNoGroupRegisteredShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should not have panicked", r)
		}
	}()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.registerTrigger(epochStartTrigger)
	gan.notifyGroupsForTrigger(epochStartTrigger.GetName(), &block.Header{}, groupTypes.Prepare)
}

func TestGroupNotifier_notifyGroupsForTriggerTriggerMissingShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should not have panicked", r)
		}
	}()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	gan.notifyGroupsForTrigger(epochStartTrigger.GetName(), &block.Header{}, groupTypes.Prepare)
}

func TestGroupNotifier_notifyGroupsForTriggerOneGroup(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	var handleActionCalled atomic.Flag

	groupAction := &groupActionsMocks.GroupActionStub{
		HandleActionCalled: func(triggerData interface{}, stage groupTypes.TriggerStage) error {
			handleActionCalled.Set()
			return nil
		},
	}

	_ = gan.Register(groupAction, epochStartTrigger)

	require.False(t, handleActionCalled.IsSet())
	gan.notifyGroupsForTrigger(epochStartTrigger.GetName(), &block.Header{}, groupTypes.Prepare)
	require.True(t, handleActionCalled.IsSet())
}

func TestGroupNotifier_notifyGroupsForTriggerTwoGroups(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	var handleAction1Called, handleAction2Called atomic.Flag

	group1Action := &groupActionsMocks.GroupActionStub{
		HandleActionCalled: func(triggerData interface{}, stage groupTypes.TriggerStage) error {
			handleAction1Called.Set()
			return nil
		},
		GroupIDCalled: func() string {
			return "grID1"
		},
	}

	group2Action := &groupActionsMocks.GroupActionStub{
		HandleActionCalled: func(triggerData interface{}, stage groupTypes.TriggerStage) error {
			handleAction2Called.Set()
			return nil
		},
		GroupIDCalled: func() string {
			return "grID2"
		},
	}

	_ = gan.Register(group1Action, epochStartTrigger)
	_ = gan.Register(group2Action, epochStartTrigger)

	require.False(t, handleAction1Called.IsSet())
	require.False(t, handleAction2Called.IsSet())
	gan.notifyGroupsForTrigger(epochStartTrigger.GetName(), &block.Header{}, groupTypes.Prepare)
	require.True(t, handleAction1Called.IsSet())
	require.True(t, handleAction2Called.IsSet())
}

func TestGroupNotifier_notifyGroupsForTriggerHandleActionOneFailed(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	var action1Trigger, action2Trigger atomic.String

	group1Action := &groupActionsMocks.GroupActionStub{
		HandleActionCalled: func(triggerData interface{}, stage groupTypes.TriggerStage) error {
			td, _ := triggerData.(*groupTypes.TriggerData)
			action1Trigger.Set(td.TriggerID)
			return errors.New("error")
		},
		GroupIDCalled: func() string {
			return "grID1"
		},
	}

	group2Action := &groupActionsMocks.GroupActionStub{
		HandleActionCalled: func(triggerData interface{}, stage groupTypes.TriggerStage) error {
			td, _ := triggerData.(*groupTypes.TriggerData)
			action2Trigger.Set(td.TriggerID)
			return nil
		},
		GroupIDCalled: func() string {
			return "grID2"
		},
	}

	_ = gan.Register(group1Action, epochStartTrigger)
	_ = gan.Register(group2Action, epochStartTrigger)

	require.Equal(t, "", action1Trigger.Get())
	require.Equal(t, "", action2Trigger.Get())
	gan.notifyGroupsForTrigger(epochStartTrigger.GetName(), &block.Header{}, groupTypes.Prepare)
	require.Equal(t, epochStartTrigger.GetName(), action1Trigger.Get())
	require.Equal(t, epochStartTrigger.GetName(), action2Trigger.Get())
}

func TestGroupNotifier_CloseNoEventNoGroup(t *testing.T) {
	t.Parallel()

	gan1 := NewGroupActionNotifier()
	gan2 := NewGroupActionNotifier()

	err := gan2.Close()
	require.Nil(t, err)
	require.Equal(t, gan1, gan2)
}

func TestGroupNotifier_CloseOneEventNoGroup(t *testing.T) {
	t.Parallel()

	gan1 := NewGroupActionNotifier()
	gan2 := NewGroupActionNotifier()

	var unregisterCalled atomic.Flag
	epochStartTrigger := &epochStart.EpochStartNotifierStub{
		UnregisterHandlerCalled: func(handler es.ActionHandler) {
			unregisterCalled.Set()
		},
	}
	_ = gan2.registerTrigger(epochStartTrigger)
	err := gan2.Close()

	require.Nil(t, err)
	require.Equal(t, gan1, gan2)
	require.True(t, unregisterCalled.IsSet())
}

func TestGroupNotifier_CloseOneEventOneGroup(t *testing.T) {
	t.Parallel()

	gan1 := NewGroupActionNotifier()
	gan2 := NewGroupActionNotifier()

	var unregisterCalled atomic.Flag
	epochStartTrigger := &epochStart.EpochStartNotifierStub{
		UnregisterHandlerCalled: func(handler es.ActionHandler) {
			unregisterCalled.Set()
		},
	}
	group1Action := &groupActionsMocks.GroupActionStub{
		HandleActionCalled: func(triggerData interface{}, stage groupTypes.TriggerStage) error {
			return nil
		},
	}

	_ = gan2.Register(group1Action, epochStartTrigger)

	err := gan2.Close()
	require.Nil(t, err)
	require.Equal(t, gan1, gan2)
	require.True(t, unregisterCalled.IsSet())
}

func TestGroupNotifier_CloseTwoEventsTwoGroups(t *testing.T) {
	t.Parallel()

	gan1 := NewGroupActionNotifier()
	gan2 := NewGroupActionNotifier()

	var unregister1Called, unregister2Called atomic.Flag
	epochStartTrigger1 := &epochStart.EpochStartNotifierStub{
		UnregisterHandlerCalled: func(handler es.ActionHandler) {
			unregister1Called.Set()
		},
		GetNameCalled: func() string {
			return "event1"
		},
	}
	epochStartTrigger2 := &epochStart.EpochStartNotifierStub{
		UnregisterHandlerCalled: func(handler es.ActionHandler) {
			unregister2Called.Set()
		},
		GetNameCalled: func() string {
			return "event2"
		},
	}

	group1Action := &groupActionsMocks.GroupActionStub{
		HandleActionCalled: func(triggerData interface{}, stage groupTypes.TriggerStage) error {
			return nil
		},
		GroupIDCalled: func() string {
			return "gr1"
		},
	}

	group2Action := &groupActionsMocks.GroupActionStub{
		HandleActionCalled: func(triggerData interface{}, stage groupTypes.TriggerStage) error {
			return nil
		},
		GroupIDCalled: func() string {
			return "gr2"
		},
	}

	_ = gan2.Register(group1Action, epochStartTrigger1)
	_ = gan2.Register(group2Action, epochStartTrigger2)

	err := gan2.Close()
	require.Nil(t, err)
	require.Equal(t, gan1, gan2)
	require.True(t, unregister1Called.IsSet())
	require.True(t, unregister2Called.IsSet())
}

func TestGroupNotifier_getOrderedGroupsKeys(t *testing.T) {
	t.Parallel()

	groups := make(map[string]groupTypes.GroupActionHandler)

	for i := 0; i < 9; i++ {
		groupID := fmt.Sprintf("%d", i)
		groups[groupID], _ = NewGroupWithDefaultLock(groupID)
	}

	orderedKeys := getOrderedGroupsKeys(groups)
	for i := 0; i < 9; i++ {
		keyInt, _ := strconv.Atoi(orderedKeys[i])
		require.Equal(t, i, keyInt)
	}
}

func TestGroupNotifier_RegisterTrigger(t *testing.T) {
	t.Parallel()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}

	err := gan.registerTrigger(epochStartTrigger)
	require.Equal(t, epochStartTrigger, gan.triggers[epochStartTrigger.GetName()])
	require.Len(t, gan.activeTriggers, 1)
	require.Len(t, gan.groupsPerEvent, 1)
	require.Nil(t, err)

	epochStartTrigger2 := &epochStart.EpochStartNotifierStub{GetNameCalled: func() string {
		return ""
	}}
	err = gan.registerTrigger(epochStartTrigger2)
	require.Equal(t, errInvalidTriggerID, err)
	require.Len(t, gan.groupsPerEvent, 1)
	require.Len(t, gan.triggers, 1)
	require.Len(t, gan.activeTriggers, 1)

	epochStartTrigger3 := &epochStart.EpochStartNotifierStub{GetNameCalled: func() string {
		return "33"
	}}
	err = gan.registerTrigger(epochStartTrigger3)
	require.Equal(t, nil, err)
	require.Len(t, gan.groupsPerEvent, 2)
	require.Len(t, gan.triggers, 2)
	require.Len(t, gan.activeTriggers, 2)
}
