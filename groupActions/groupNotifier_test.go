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
	gan := NewGroupActionNotifier()
	require.NotNil(t, gan)
	require.Equal(t, 0, len(gan.groupsPerEvent))
	require.Equal(t, 0, len(gan.triggers))
	require.Equal(t, 0, len(gan.activeTriggers))
}

func TestGroupNotifier_addTriggerOK(t *testing.T) {
	gan := NewGroupActionNotifier()

	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	err := gan.addTrigger(epochStartTrigger)
	require.Equal(t, 1, len(gan.triggers))
	require.Equal(t, 1, len(gan.groupsPerEvent))
	require.Equal(t, 0, len(gan.activeTriggers))

	require.Nil(t, err)
}

func TestGroupNotifier_addTriggerAlreadyPresent(t *testing.T) {
	gan := NewGroupActionNotifier()

	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.addTrigger(epochStartTrigger)
	err := gan.addTrigger(epochStartTrigger)
	require.Equal(t, 1, len(gan.triggers))
	require.Equal(t, 1, len(gan.groupsPerEvent))
	require.Equal(t, 0, len(gan.activeTriggers))
	require.Equal(t, errTriggerAlreadyRegistered, err)
}

func TestGroupNotifier_RegisterTriggerNilTrigger(t *testing.T) {
	gan := NewGroupActionNotifier()

	err := gan.RegisterTrigger(nil)
	require.Equal(t, errNilTrigger, err)
}

func TestGroupNotifier_RegisterTriggerOK(t *testing.T) {
	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}

	err := gan.RegisterTrigger(epochStartTrigger)
	require.Equal(t, epochStartTrigger, gan.triggers[epochStartTrigger.GetName()])
	require.Equal(t, 1, len(gan.activeTriggers))
	require.Equal(t, 1, len(gan.groupsPerEvent))
	require.Nil(t, err)
}

func TestGroupNotifier_RegisterTriggerAlreadyPresent(t *testing.T) {
	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.RegisterTrigger(epochStartTrigger)

	err := gan.RegisterTrigger(epochStartTrigger)
	require.Equal(t, errTriggerAlreadyRegistered, err)
}

func TestGroupNotifier_RegisterTriggerSecond(t *testing.T) {
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
	_ = gan.RegisterTrigger(epochStartTrigger1)
	err := gan.RegisterTrigger(epochStartTrigger2)
	require.Nil(t, err)
	require.Equal(t, 2, len(gan.triggers))
}

func TestGroupNotifier_RegisterGroupAction(t *testing.T) {
	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.RegisterTrigger(epochStartTrigger)
	groupAction, _ := NewGroupWithDefaultLock("grID")

	err := gan.RegisterGroupAction(groupAction, epochStartTrigger.GetName())
	require.Nil(t, err)
}

func TestGroupNotifier_RegisterGroupActionNilActionHandler(t *testing.T) {
	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.RegisterTrigger(epochStartTrigger)

	err := gan.RegisterGroupAction(nil, epochStartTrigger.GetName())
	require.Equal(t, errNilActionHandler, err)
}

func TestGroupNotifier_RegisterGroupActionInvalidTriggerID(t *testing.T) {
	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.RegisterTrigger(epochStartTrigger)

	groupAction, _ := NewGroupWithDefaultLock("grID")
	err := gan.RegisterGroupAction(groupAction, "")
	require.Equal(t, errNilTriggerType, err)
}

func TestGroupNotifier_RegisterGroupActionUnknownTrigger(t *testing.T) {
	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}

	groupAction, _ := NewGroupWithDefaultLock("grID")
	err := gan.RegisterGroupAction(groupAction, epochStartTrigger.GetName())
	require.Equal(t, errUnknownTrigger, err)
}

func TestGroupNotifier_RegisterGroupActionAlreadyRegisteredGroup(t *testing.T) {
	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.RegisterTrigger(epochStartTrigger)

	groupAction, _ := NewGroupWithDefaultLock("grID")
	_ = gan.RegisterGroupAction(groupAction, epochStartTrigger.GetName())
	err := gan.RegisterGroupAction(groupAction, epochStartTrigger.GetName())
	require.Equal(t, errGroupAlreadyRegisteredForTrigger, err)
}

func TestGroupNotifier_RegisterGroupActionTwoGroupsSameTrigger(t *testing.T) {
	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.RegisterTrigger(epochStartTrigger)
	group1Action, _ := NewGroupWithDefaultLock("grID1")
	group2Action, _ := NewGroupWithDefaultLock("grID2")

	_ = gan.RegisterGroupAction(group1Action, epochStartTrigger.GetName())
	err := gan.RegisterGroupAction(group2Action, epochStartTrigger.GetName())

	require.Nil(t, err)
	require.Equal(t, 2, len(gan.groupsPerEvent[epochStartTrigger.GetName()]))
}

func TestGroupNotifier_notifyGroupsForTriggerNoGroupRegisteredShouldNotPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			require.Fail(t, "should not have panicked", r)
		}
	}()

	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.RegisterTrigger(epochStartTrigger)
	gan.notifyGroupsForTrigger(epochStartTrigger.GetName(), &block.Header{}, groupTypes.Prepare)
}

func TestGroupNotifier_notifyGroupsForTriggerTriggerMissingShouldNotPanic(t *testing.T) {
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
	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.RegisterTrigger(epochStartTrigger)
	var handleActionCalled atomic.Flag

	groupAction := &groupActionsMocks.GroupActionStub{
		HandleActionCalled: func(triggerData interface{}, stage groupTypes.TriggerStage) error {
			handleActionCalled.Set()
			return nil
		},
	}

	_ = gan.RegisterGroupAction(groupAction, epochStartTrigger.GetName())

	require.False(t, handleActionCalled.IsSet())
	gan.notifyGroupsForTrigger(epochStartTrigger.GetName(), &block.Header{}, groupTypes.Prepare)
	require.True(t, handleActionCalled.IsSet())
}

func TestGroupNotifier_notifyGroupsForTriggerTwoGroups(t *testing.T) {
	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.RegisterTrigger(epochStartTrigger)
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

	_ = gan.RegisterGroupAction(group1Action, epochStartTrigger.GetName())
	_ = gan.RegisterGroupAction(group2Action, epochStartTrigger.GetName())

	require.False(t, handleAction1Called.IsSet())
	require.False(t, handleAction2Called.IsSet())
	gan.notifyGroupsForTrigger(epochStartTrigger.GetName(), &block.Header{}, groupTypes.Prepare)
	require.True(t, handleAction1Called.IsSet())
	require.True(t, handleAction2Called.IsSet())
}

func TestGroupNotifier_notifyGroupsForTriggerHandleActionOneFailed(t *testing.T) {
	gan := NewGroupActionNotifier()
	epochStartTrigger := &epochStart.EpochStartNotifierStub{}
	_ = gan.RegisterTrigger(epochStartTrigger)
	var handleAction1Called, handleAction2Called atomic.Flag

	group1Action := &groupActionsMocks.GroupActionStub{
		HandleActionCalled: func(triggerData interface{}, stage groupTypes.TriggerStage) error {
			handleAction1Called.Set()
			return errors.New("error")
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

	_ = gan.RegisterGroupAction(group1Action, epochStartTrigger.GetName())
	_ = gan.RegisterGroupAction(group2Action, epochStartTrigger.GetName())

	require.False(t, handleAction1Called.IsSet())
	require.False(t, handleAction2Called.IsSet())
	gan.notifyGroupsForTrigger(epochStartTrigger.GetName(), &block.Header{}, groupTypes.Prepare)
	require.True(t, handleAction1Called.IsSet())
	require.True(t, handleAction2Called.IsSet())
}

func TestGroupNotifier_CloseNoEventNoGroup(t *testing.T) {
	gan1 := NewGroupActionNotifier()
	gan2 := NewGroupActionNotifier()

	err := gan2.Close()
	require.Nil(t, err)
	require.Equal(t, gan1, gan2)
}

func TestGroupNotifier_CloseOneEventNoGroup(t *testing.T) {
	gan1 := NewGroupActionNotifier()
	gan2 := NewGroupActionNotifier()

	var unregisterCalled atomic.Flag
	epochStartTrigger := &epochStart.EpochStartNotifierStub{
		UnregisterHandlerCalled: func(handler es.ActionHandler) {
			unregisterCalled.Set()
		},
	}
	_ = gan2.RegisterTrigger(epochStartTrigger)
	err := gan2.Close()

	require.Nil(t, err)
	require.Equal(t, gan1, gan2)
	require.True(t, unregisterCalled.IsSet())
}

func TestGroupNotifier_CloseOneEventOneGroup(t *testing.T) {
	gan1 := NewGroupActionNotifier()
	gan2 := NewGroupActionNotifier()

	var unregisterCalled atomic.Flag
	epochStartTrigger := &epochStart.EpochStartNotifierStub{
		UnregisterHandlerCalled: func(handler es.ActionHandler) {
			unregisterCalled.Set()
		},
	}
	_ = gan2.RegisterTrigger(epochStartTrigger)
	group1Action := &groupActionsMocks.GroupActionStub{
		HandleActionCalled: func(triggerData interface{}, stage groupTypes.TriggerStage) error {
			return nil
		},
	}

	_ = gan2.RegisterGroupAction(group1Action, epochStartTrigger.GetName())

	err := gan2.Close()
	require.Nil(t, err)
	require.Equal(t, gan1, gan2)
	require.True(t, unregisterCalled.IsSet())
}

func TestGroupNotifier_CloseTwoEventsTwoGroups(t *testing.T) {
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

	_ = gan2.RegisterTrigger(epochStartTrigger1)
	_ = gan2.RegisterTrigger(epochStartTrigger2)
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

	_ = gan2.RegisterGroupAction(group1Action, epochStartTrigger1.GetName())
	_ = gan2.RegisterGroupAction(group2Action, epochStartTrigger2.GetName())

	err := gan2.Close()
	require.Nil(t, err)
	require.Equal(t, gan1, gan2)
	require.True(t, unregister1Called.IsSet())
	require.True(t, unregister2Called.IsSet())
}

func TestGroupNotifier_getOrderedGroupsKeys(t *testing.T) {
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
