package groupActions

import "errors"

// errUnknownTrigger signals that an unknown trigger was used
var errUnknownTrigger = errors.New("the trigger is unknown")

// errGroupAlreadyRegisteredForTrigger signals that the group is already registered for the trigger
var errGroupAlreadyRegisteredForTrigger = errors.New("the group is already registered for the trigger")

// errNilActionHandler signals that the Action handler is invalid
var errNilActionHandler = errors.New("the Action handler is invalid")

// errNilTriggerType signals that the trigger type is nil
var errNilTriggerType = errors.New("the trigger type is nil")

// errNilTrigger signals that the trigger is nil
var errNilTrigger = errors.New("the trigger is nil")

// errTriggerAlreadyRegistered signals that a registered trigger cannot be overwritten
var errTriggerAlreadyRegistered = errors.New("the trigger is already registered")

// errNilLocker signals that the used locker is nil
var errNilLocker = errors.New("the locker is nil")

// errInvalidGroupID signals that the used group ID is invalid or nil
var errInvalidGroupID = errors.New("the group ID is invalid")

// errGroupMemberAlreadyExists signals that the group member is already in the group
var errGroupMemberAlreadyExists = errors.New("the group member already exists")

// errInvalidTriggerID signals the usage of an invalid trigger ID
var errInvalidTriggerID = errors.New("the trigger ID is invalid")

// errNilCallbackFunction signals the usage of a nil callback function
var errNilCallbackFunction = errors.New("the callback function is nil")
