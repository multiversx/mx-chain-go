package groupActions

import (
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/groupActions/groupTypes"
)

const (
	// TODO: move to elrond-go-core
	notifierOrderStart = 100
)

type groupNotifierTriggerData struct {
	trigger            groupTypes.RegistrationHandler
	registeredNotifier *groupNotifierTrigger
}

type groupNotifier struct {
	groupsPerEvent map[string]map[string]groupTypes.GroupActionHandler
	triggers       map[string]*groupNotifierTriggerData
	mut sync.RWMutex
}

// NewGroupActionNotifier creates a group notifier instance
func NewGroupActionNotifier() *groupNotifier {
	return &groupNotifier{
		groupsPerEvent: make(map[string]map[string]groupTypes.GroupActionHandler),
		triggers:       make(map[string]*groupNotifierTriggerData),
	}
}

// Register adds a groupAction together with the Action trigger
func (gn *groupNotifier) Register(group groupTypes.GroupActionHandler, trigger groupTypes.RegistrationHandler) error {
	if check.IfNil(group) {
		return errNilGroupActionHandler
	}
	if check.IfNil(trigger) {
		return errNilTrigger
	}
	if len(trigger.GetName()) == 0 {
		return errInvalidTriggerID
	}

	gn.mut.Lock()
	defer gn.mut.Unlock()

	err := gn.registerTrigger(trigger)
	if err != nil {
		return err
	}

	groupsForTrigger := gn.groupsPerEvent[trigger.GetName()]
	groupID := group.ID()
	_, exists := groupsForTrigger[groupID]
	if exists {
		return errGroupAlreadyRegisteredForTrigger
	}

	groupsForTrigger[groupID] = group

	return nil
}

// Close cleans up the group notifier
func (gn *groupNotifier) Close() error {
	gn.mut.Lock()
	defer gn.mut.Unlock()

	for _, at := range gn.triggers {
		at.trigger.UnregisterHandler(at.registeredNotifier)
	}

	gn.groupsPerEvent = make(map[string]map[string]groupTypes.GroupActionHandler)
	gn.triggers = make(map[string]*groupNotifierTriggerData)
	return nil
}

// ReceiveNotification is the registered notification function for a specific trigger
func (gn *groupNotifier) ReceiveNotification(triggerID string, header data.HeaderHandler, stage groupTypes.TriggerStage) {
	gn.mut.RLock()
	defer gn.mut.RUnlock()

	groups, exists := gn.groupsPerEvent[triggerID]
	if !exists {
		log.Warn("no groups found", "trigger", triggerID)
		return
	}

	orderedKeys := getOrderedGroupsKeys(groups)

	for _, key := range orderedKeys {
		gr, ok := groups[key]
		if !ok {
			log.Warn("group not found", "key", key)
			continue
		}

		td := &groupTypes.TriggerData{
			TriggerID: triggerID,
			Data:      header,
		}

		_ = gr.HandleAction(td, stage)
	}
}

// IsInterfaceNil returns true if the receiver is nil, false otherwise
func (gn *groupNotifier) IsInterfaceNil() bool {
	return gn == nil
}

// registerTrigger registers a new trigger that can be used by groups to be notified on it's triggered events
// the parameter validation needs to be done by caller
func (gn *groupNotifier) registerTrigger(trigger groupTypes.RegistrationHandler) error {
	notifyOrder := gn.computeNextNotifyOrder()
	gnt, err := newGroupNotifierTrigger(trigger.GetName(), notifyOrder, gn)
	if err != nil {
		return err
	}

	return gn.addTriggerWithGroupNotifier(trigger, gnt)
}

// parameter validation needs to be done by caller
func (gn *groupNotifier) addTriggerWithGroupNotifier(trigger groupTypes.RegistrationHandler, notifierTrigger *groupNotifierTrigger) error {
	triggerType := trigger.GetName()
	_, exists := gn.triggers[triggerType]
	if exists {
		return nil
	}

	gn.triggers[triggerType] = &groupNotifierTriggerData{
		trigger:            trigger,
		registeredNotifier: notifierTrigger,
	}
	gn.groupsPerEvent[triggerType] = make(map[string]groupTypes.GroupActionHandler)
	trigger.RegisterHandler(notifierTrigger)

	return nil
}

// parameter validation needs to be done by caller
func getOrderedGroupsKeys(groups map[string]groupTypes.GroupActionHandler) []string {
	orderedKeys := make([]string, 0, len(groups))
	for key := range groups {
		orderedKeys = append(orderedKeys, key)
	}

	// Always keep the same ordering of notifying the groups waiting on the same trigger
	sort.Strings(orderedKeys)

	return orderedKeys
}

func (gn *groupNotifier) computeNextNotifyOrder() uint32 {
	notifyOrder := notifierOrderStart + len(gn.triggers)
	return uint32(notifyOrder)
}
