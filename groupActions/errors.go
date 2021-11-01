package groupActions

import "errors"

// errGroupAlreadyRegisteredForTrigger signals that the group is already registered for the trigger
var errGroupAlreadyRegisteredForTrigger = errors.New("the group is already registered for the trigger")

// errNilActionHandler signals that the action handler is nil
var errNilActionHandler = errors.New("the action handler is nil")

// errNilGroupActionHandler signals that the group action handler is nil
var errNilGroupActionHandler = errors.New("the group action handler is nil")

// errNilTrigger signals that the trigger is nil
var errNilTrigger = errors.New("the trigger is nil")

// errNilLocker signals that the used locker is nil
var errNilLocker = errors.New("the locker is nil")

// errInvalidGroupID signals that the used group ID is invalid
var errInvalidGroupID = errors.New("the group ID is invalid")

// errGroupMemberAlreadyExists signals that the group member is already in the group
var errGroupMemberAlreadyExists = errors.New("the group member already exists")

// errInvalidTriggerID signals the usage of an invalid trigger ID
var errInvalidTriggerID = errors.New("the trigger ID is invalid")

// errNilSubscriber signals the usage of a nil subscriber
var errNilSubscriber = errors.New("the subscriber is nil")
