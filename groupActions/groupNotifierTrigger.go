package groupActions

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/groupActions/groupTypes"
)

type notifierCallback func(triggerID string, header data.HeaderHandler, stage groupTypes.TriggerStage)

type groupNotifierTrigger struct {
	triggerID   string
	notifyOrder uint32
	callback    notifierCallback
}

func newGroupNotifierTrigger(triggerID string, notifyOrder uint32, callback notifierCallback) (*groupNotifierTrigger, error) {
	if len(triggerID) ==0 {
		return nil, errInvalidTriggerID
	}
	if callback == nil {
		return nil, errNilCallbackFunction
	}

	return &groupNotifierTrigger{
		triggerID:   triggerID,
		notifyOrder: notifyOrder,
		callback:    callback,
	}, nil
}

// Prepare -
func (nt *groupNotifierTrigger) Prepare(hdr data.HeaderHandler, _ data.BodyHandler) {
	nt.callback(nt.triggerID, hdr, groupTypes.Prepare)
}

// Action -
func (nt *groupNotifierTrigger) Action(hdr data.HeaderHandler) {
	nt.callback(nt.triggerID, hdr, groupTypes.Action)
}

// NotifyOrder -
func (nt *groupNotifierTrigger) NotifyOrder() uint32 {
	return nt.notifyOrder
}
