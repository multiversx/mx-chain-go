package groupActions

import (
	"sort"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/groupActions/groupTypes"
)

const (
	notifierOrderStart = 100
)

type groupNotifier struct {
	groupsPerEvent map[string]map[string]groupTypes.GroupActionHandler
	triggers       map[string]groupTypes.RegistrationHandler
	activeTriggers []*groupNotifierTrigger
	sync.RWMutex
}

// NewGroupActionNotifier creates a group notifier instance
func NewGroupActionNotifier() *groupNotifier {
	return &groupNotifier{
		groupsPerEvent: make(map[string]map[string]groupTypes.GroupActionHandler),
		triggers:       make(map[string]groupTypes.RegistrationHandler),
		activeTriggers: make([]*groupNotifierTrigger, 0),
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

	gn.Lock()
	defer gn.Unlock()

	err := gn.registerTrigger(trigger)
	if err != nil {
		return err
	}

	groupsForTrigger, exists := gn.groupsPerEvent[trigger.GetName()]
	if !exists {
		return errUnknownTrigger
	}

	groupID := group.ID()
	_, exists = groupsForTrigger[groupID]
	if exists {
		return errGroupAlreadyRegisteredForTrigger
	}

	groupsForTrigger[groupID] = group

	return nil
}

// registerTrigger registers a new trigger that can be used by groups to be notified on it's triggered events
func (gn *groupNotifier) registerTrigger(trigger groupTypes.RegistrationHandler) error {
	notifyOrder := gn.computeNextNotifyOrder()
	gnt, err := newGroupNotifierTrigger(trigger.GetName(), notifyOrder, gn.notifyGroupsForTrigger)
	if err != nil {
		return err
	}

	err = gn.addTrigger(trigger)
	if err != nil {
		return err
	}

	gn.activeTriggers = append(gn.activeTriggers, gnt)
	trigger.RegisterHandler(gnt)

	return nil
}

// Close cleans up the group notifier
func (gn *groupNotifier) Close() error {
	gn.Lock()
	defer gn.Unlock()

	for _, at := range gn.activeTriggers {
		trig, exists := gn.triggers[at.triggerID]
		if !exists {
			log.Warn("groupNotifier.Close", "trigger", at.triggerID, "err", "trigger not found")
			continue
		}
		trig.UnregisterHandler(at)
	}

	gn.groupsPerEvent = make(map[string]map[string]groupTypes.GroupActionHandler)
	gn.triggers = make(map[string]groupTypes.RegistrationHandler)
	gn.activeTriggers = make([]*groupNotifierTrigger, 0)

	return nil
}

func (gn *groupNotifier) addTrigger(trigger groupTypes.RegistrationHandler) error {
	triggerType := trigger.GetName()
	if len(triggerType) == 0 {
		return errInvalidTriggerID
	}
	_, exists := gn.triggers[triggerType]
	if exists {
		return nil
	}

	gn.triggers[triggerType] = trigger
	gn.groupsPerEvent[triggerType] = make(map[string]groupTypes.GroupActionHandler)
	return nil
}

func (gn *groupNotifier) notifyGroupsForTrigger(triggerID string, header data.HeaderHandler, stage groupTypes.TriggerStage) {
	// Notify all groups for the trigger
	gn.RLock()
	defer gn.RUnlock()

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
