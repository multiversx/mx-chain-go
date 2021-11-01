package groupActions

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/groupActions/groupTypes"
	"github.com/ElrondNetwork/elrond-go/testscommon/groupActionsMocks"
	"github.com/stretchr/testify/require"
)

func Test_newGroupNotifierTriggerInvalidGroupID(t *testing.T) {
	t.Parallel()

	subscriber := &groupActionsMocks.SubscriberStub{}
	grTrigger, err := newGroupNotifierTrigger("", 1, subscriber)

	require.Nil(t, grTrigger)
	require.Equal(t, errInvalidTriggerID, err)
}

func Test_newGroupNotifierTriggerNilCallback(t *testing.T) {
	t.Parallel()

	grTrigger, err := newGroupNotifierTrigger("triggerID", 1, nil)

	require.Nil(t, grTrigger)
	require.Equal(t, errNilSubscriber, err)
}

func Test_newGroupNotifierTrigger(t *testing.T) {
	t.Parallel()

	subscriber := &groupActionsMocks.SubscriberStub{}
	trigID := "triggerID"
	grTrigger, err := newGroupNotifierTrigger("triggerID", 1, subscriber)

	require.Nil(t, err)
	require.Equal(t, trigID, grTrigger.triggerID)
	require.NotNil(t, grTrigger.subscriber)
}

func Test_groupNotifierTrigger_Prepare(t *testing.T) {
	t.Parallel()

	type cbData struct {
		triggerID string
		header    data.HeaderHandler
		stage     groupTypes.TriggerStage
	}

	expectedData := cbData{
		triggerID: "triggerID",
		header:    &block.Header{ShardID: 100},
		stage:     groupTypes.Prepare,
	}

	actual := cbData{}

	subscriber := &groupActionsMocks.SubscriberStub{
		ReceiveNotificationCalled: func(triggerID string, header data.HeaderHandler, stage groupTypes.TriggerStage) {
			actual.header = header
			actual.triggerID = triggerID
			actual.stage = stage
		},
	}

	grTrigger, _ := newGroupNotifierTrigger(expectedData.triggerID, 1, subscriber)
	grTrigger.Prepare(expectedData.header, nil)

	require.Equal(t, expectedData, actual)
}

func Test_groupNotifierTrigger_Action(t *testing.T) {
	t.Parallel()

	type cbData struct {
		triggerID string
		header    data.HeaderHandler
		stage     groupTypes.TriggerStage
	}

	expectedData := cbData{
		triggerID: "triggerID",
		header:    &block.Header{ShardID: 100},
		stage:     groupTypes.Action,
	}

	actual := cbData{}

	subscriber := &groupActionsMocks.SubscriberStub{
		ReceiveNotificationCalled: func(triggerID string, header data.HeaderHandler, stage groupTypes.TriggerStage) {
			actual.header = header
			actual.triggerID = triggerID
			actual.stage = stage
		},
	}

	grTrigger, _ := newGroupNotifierTrigger(expectedData.triggerID, 1, subscriber)
	grTrigger.Action(expectedData.header)

	require.Equal(t, expectedData, actual)
}

func Test_groupNotifierTrigger_NotifyOrder(t *testing.T) {
	t.Parallel()

	subscriber := &groupActionsMocks.SubscriberStub{}
	notifyOrder := uint32(1)
	grTrigger, _ := newGroupNotifierTrigger("triggerID", 1, subscriber)

	require.Equal(t, notifyOrder, grTrigger.NotifyOrder())
}
