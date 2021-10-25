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

// RegisterGroupAction adds a groupAction together with the Action trigger
func (gn *groupNotifier) RegisterGroupAction(group groupTypes.GroupActionHandler, triggerType string) error {
	if check.IfNil(group) {
		return errNilActionHandler
	}
	if len(triggerType) == 0 {
		return errNilTriggerType
	}

	gn.Lock()
	defer gn.Unlock()

	groupsForTrigger, ok := gn.groupsPerEvent[triggerType]
	if !ok {
		return errUnknownTrigger
	}

	groupID := group.GroupID()
	_, ok = groupsForTrigger[groupID]
	if ok {
		return errGroupAlreadyRegisteredForTrigger
	}

	groupsForTrigger[groupID] = group

	return nil
}

// RegisterTrigger registers a new trigger that can be used by groups to be notified on it's triggered events
func (gn *groupNotifier) RegisterTrigger(trigger groupTypes.RegistrationHandler) error {
	if check.IfNil(trigger) {
		return errNilTrigger
	}

	gn.Lock()
	defer gn.Unlock()

	err := gn.addTrigger(trigger)
	if err != nil {
		return err
	}

	notifyOrder := gn.computeNextNotifyOrder()
	gnt, err := newGroupNotifierTrigger(trigger.GetName(), notifyOrder, gn.notifyGroupsForTrigger)
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
		trig, ok := gn.triggers[at.triggerID]
		if !ok {
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
	_, ok := gn.triggers[triggerType]
	if ok {
		return errTriggerAlreadyRegistered
	}

	gn.triggers[triggerType] = trigger
	gn.groupsPerEvent[triggerType] = make(map[string]groupTypes.GroupActionHandler)
	return nil
}

func (gn *groupNotifier) notifyGroupsForTrigger(triggerID string, header data.HeaderHandler, stage groupTypes.TriggerStage) {
	// Notify all groups for the trigger
	gn.RLock()
	defer gn.RUnlock()

	groups, ok := gn.groupsPerEvent[triggerID]
	if !ok {
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

		err := gr.HandleAction(header, stage)
		if err != nil {
			log.Warn("error trigger Action", "trigger", triggerID, "group", gr.GroupID(), "error", err)
		}
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
