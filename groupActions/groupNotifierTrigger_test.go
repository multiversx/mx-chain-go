package groupActions

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/groupActions/groupTypes"
	"github.com/stretchr/testify/require"
)

func Test_newGroupNotifierTriggerInvalidGroupID(t *testing.T) {
	t.Parallel()

	cbFunc := func(_ string, _ data.HeaderHandler, _ groupTypes.TriggerStage) {}
	grTrigger, err := newGroupNotifierTrigger("", 1, cbFunc)

	require.Nil(t, grTrigger)
	require.Equal(t, errInvalidTriggerID, err)
}

func Test_newGroupNotifierTriggerNilCallback(t *testing.T) {
	t.Parallel()

	grTrigger, err := newGroupNotifierTrigger("triggerID", 1, nil)

	require.Nil(t, grTrigger)
	require.Equal(t, errNilCallbackFunction, err)
}

func Test_newGroupNotifierTrigger(t *testing.T) {
	t.Parallel()

	cbFunc := func(_ string, _ data.HeaderHandler, _ groupTypes.TriggerStage) {}
	trigID := "triggerID"
	grTrigger, err := newGroupNotifierTrigger("triggerID", 1, cbFunc)

	require.Nil(t, err)
	require.Equal(t, trigID, grTrigger.triggerID)
	require.NotNil(t, grTrigger.callback)
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
	cbFunc := func(triggerID string, header data.HeaderHandler, stage groupTypes.TriggerStage) {
		actual.header = header
		actual.triggerID = triggerID
		actual.stage = stage
	}

	grTrigger, _ := newGroupNotifierTrigger(expectedData.triggerID, 1, cbFunc)
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
	cbFunc := func(triggerID string, header data.HeaderHandler, stage groupTypes.TriggerStage) {
		actual.header = header
		actual.triggerID = triggerID
		actual.stage = stage
	}

	grTrigger, _ := newGroupNotifierTrigger(expectedData.triggerID, 1, cbFunc)
	grTrigger.Action(expectedData.header)

	require.Equal(t, expectedData, actual)
}

func Test_groupNotifierTrigger_NotifyOrder(t *testing.T) {
	t.Parallel()

	cbFunc := func(_ string, _ data.HeaderHandler, _ groupTypes.TriggerStage) {}
	notifyOrder := uint32(1)
	grTrigger, _ := newGroupNotifierTrigger("triggerID", 1, cbFunc)

	require.Equal(t, notifyOrder, grTrigger.NotifyOrder())
}
