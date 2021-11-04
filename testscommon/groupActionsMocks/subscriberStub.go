package groupActionsMocks

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/groupActions/groupTypes"
)

// SubscriberStub -
type SubscriberStub struct {
	ReceiveNotificationCalled func(triggerID string, header data.HeaderHandler, stage groupTypes.TriggerStage)
}

// ReceiveNotification -
func (ss *SubscriberStub) ReceiveNotification(triggerID string, header data.HeaderHandler, stage groupTypes.TriggerStage) {
	if ss.ReceiveNotificationCalled != nil {
		ss.ReceiveNotificationCalled(triggerID, header, stage)
	}
}

// IsInterfaceNil -
func (ss *SubscriberStub) IsInterfaceNil() bool {
	return ss == nil
}
