package groupActions

import (
	"errors"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go/groupActions/groupTypes"
	"github.com/ElrondNetwork/elrond-go/testscommon/groupActionsMocks"
	"github.com/stretchr/testify/require"
)

func TestNewGroupNilLocker(t *testing.T) {
	t.Parallel()

	gr, err := NewGroup(nil, "grID")
	require.Equal(t, errNilLocker, err)
	require.Nil(t, gr)
}

func TestNewGroupInvalidGroupID(t *testing.T) {
	t.Parallel()

	gr, err := NewGroup(&sync.RWMutex{}, "")
	require.Equal(t, errInvalidGroupID, err)
	require.Nil(t, gr)
}
func TestNewGroup(t *testing.T) {
	t.Parallel()

	gr, err := NewGroup(&sync.RWMutex{}, "grID")
	require.Nil(t, err)
	require.NotNil(t, gr)
}

func TestNewGroupWithDefaultLockInvalidGroupID(t *testing.T) {
	t.Parallel()

	gr, err := NewGroupWithDefaultLock("")
	require.Equal(t, errInvalidGroupID, err)
	require.Nil(t, gr)
}

func TestNewGroupWithDefaultLock(t *testing.T) {
	t.Parallel()

	gr, err := NewGroupWithDefaultLock("grID")
	require.Nil(t, err)
	require.NotNil(t, gr)
}

func TestGroup_AddToGroupFirstMember(t *testing.T) {
	t.Parallel()

	gr, _ := NewGroupWithDefaultLock("grID")
	m1 := &groupActionsMocks.ActionHandlerStub{}

	err := gr.AddToGroup(m1)
	require.Nil(t, err)
	require.Equal(t, 1, len(gr.members))
	require.Equal(t, m1, gr.members[0])
}

func TestGroup_AddToGroupExistingMember(t *testing.T) {
	t.Parallel()

	gr, _ := NewGroupWithDefaultLock("grID")
	m1 := &groupActionsMocks.ActionHandlerStub{}

	_ = gr.AddToGroup(m1)
	err := gr.AddToGroup(m1)

	require.Equal(t, errGroupMemberAlreadyExists, err)
	require.Equal(t, 1, len(gr.members))
	require.Equal(t, m1, gr.members[0])
}

func TestGroup_AddToGroupSecondMember(t *testing.T) {
	t.Parallel()

	gr, _ := NewGroupWithDefaultLock("grID")
	m1 := &groupActionsMocks.ActionHandlerStub{}
	m2 := &groupActionsMocks.ActionHandlerStub{}

	_ = gr.AddToGroup(m1)
	err := gr.AddToGroup(m2)

	require.Nil(t, err)
	require.Equal(t, 2, len(gr.members))
	require.Equal(t, m1, gr.members[0])
	require.Equal(t, m2, gr.members[1])
}

func TestGroup_GroupID_NewGroup(t *testing.T) {
	t.Parallel()

	groupID := "grID1"
	gr, _ := NewGroup(&sync.RWMutex{}, groupID)

	require.Equal(t, groupID, gr.GroupID())
}

func TestGroup_GroupID_NewGroupWithDefaultLock(t *testing.T) {
	t.Parallel()

	groupID := "grID1"
	gr, _ := NewGroupWithDefaultLock(groupID)

	require.Equal(t, groupID, gr.GroupID())
}

func TestGroup_HandleActionNoMembers(t *testing.T) {
	t.Parallel()

	groupID := "grID1"
	gr, _ := NewGroupWithDefaultLock(groupID)

	err := gr.HandleAction(interface{}(1), groupTypes.Prepare)

	require.Nil(t, err)
}

func TestGroup_HandleActionOneMember(t *testing.T) {
	t.Parallel()

	groupID := "grID1"
	gr, _ := NewGroupWithDefaultLock(groupID)
	var m1Called atomic.Flag
	m1 := &groupActionsMocks.ActionHandlerStub{
		HandleActionCalled: func(_ interface{}, _ groupTypes.TriggerStage) error {
			m1Called.Set()
			return nil
		},
	}

	_ = gr.AddToGroup(m1)
	err := gr.HandleAction(interface{}(1), groupTypes.Prepare)

	require.Nil(t, err)
	require.True(t, m1Called.IsSet())
}

func TestGroup_HandleActionTwoMembers(t *testing.T) {
	t.Parallel()

	groupID := "grID1"
	gr, _ := NewGroupWithDefaultLock(groupID)
	var m1Called, m2Called atomic.Flag
	m1 := &groupActionsMocks.ActionHandlerStub{
		HandleActionCalled: func(_ interface{}, _ groupTypes.TriggerStage) error {
			m1Called.Set()
			return nil
		},
	}
	m2 := &groupActionsMocks.ActionHandlerStub{
		HandleActionCalled: func(_ interface{}, _ groupTypes.TriggerStage) error {
			m2Called.Set()
			return nil
		},
	}

	_ = gr.AddToGroup(m1)
	_ = gr.AddToGroup(m2)
	err := gr.HandleAction(interface{}(1), groupTypes.Prepare)

	require.Nil(t, err)
	require.True(t, m1Called.IsSet())
	require.True(t, m2Called.IsSet())
}

func TestGroup_HandleActionTwoMembersOneWithError(t *testing.T) {
	t.Parallel()

	groupID := "grID1"
	gr, _ := NewGroupWithDefaultLock(groupID)
	var m1Called, m2Called atomic.Flag
	expectedError := errors.New("expectedErrors")

	m1 := &groupActionsMocks.ActionHandlerStub{
		HandleActionCalled: func(_ interface{}, _ groupTypes.TriggerStage) error {
			m1Called.Set()
			return nil
		},
	}
	m2 := &groupActionsMocks.ActionHandlerStub{
		HandleActionCalled: func(_ interface{}, _ groupTypes.TriggerStage) error {
			m2Called.Set()
			return expectedError
		},
	}

	_ = gr.AddToGroup(m1)
	_ = gr.AddToGroup(m2)
	err := gr.HandleAction(interface{}(1), groupTypes.Prepare)

	require.Equal(t, expectedError, err)
	require.True(t, m1Called.IsSet())
	require.True(t, m2Called.IsSet())
}
