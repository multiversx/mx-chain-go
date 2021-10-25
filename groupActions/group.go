package groupActions

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/groupActions/groupTypes"
)

var log = logger.GetOrCreate("groupActions")

type Locker interface {
	Lock()
	Unlock()
	RLock()
	RUnlock()
}

type group struct {
	groupID string
	members []groupTypes.ActionHandler
	locker  Locker
}

// NewGroup creates a new group with external lock
func NewGroup(locker Locker, groupID string) (*group, error) {
	if locker == nil {
		return nil, errNilLocker
	}
	if len(groupID) == 0 {
		return nil, errInvalidGroupID
	}
	return &group{
		groupID: groupID,
		members: nil,
		locker:  locker,
	}, nil
}

// NewGroupWithDefaultLock creates a new group
func NewGroupWithDefaultLock(groupID string) (*group, error) {
	if len(groupID) == 0 {
		return nil, errInvalidGroupID
	}
	return &group{
		groupID: groupID,
		members: nil,
		locker:  &sync.RWMutex{},
	}, nil
}

// AddToGroup  adds a new member to group
func (g *group) AddToGroup(member groupTypes.ActionHandler) error {
	if check.IfNil(member) {
		return errNilActionHandler
	}

	g.locker.Lock()
	defer g.locker.Unlock()

	// allow every member only once
	for i := range g.members {
		if g.members[i] == member {
			return errGroupMemberAlreadyExists
		}
	}

	g.members = append(g.members, member)
	return nil
}

// HandleAction handles the group Action, returning the last error if any or nil otherwise
func (g *group) HandleAction(triggerData interface{}, stage groupTypes.TriggerStage) error {
	g.locker.RLock()
	defer g.locker.RUnlock()
	var lastErr error

	for i := range g.members {
		err := g.members[i].HandleAction(triggerData, stage)
		log.LogIfError(err)
		if err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// GroupID returns the group ID
func (g *group) GroupID() string {
	return g.groupID
}

// IsInterfaceNil returns true if there is no value under the interface
func (g *group) IsInterfaceNil() bool {
	return g == nil
}
