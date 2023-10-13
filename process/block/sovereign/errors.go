package sovereign

import "errors"

var errNoSubscribedAddresses = errors.New("no subscribed addresses provided")

var errNoSubscribedIdentifier = errors.New("no subscribed identifier provided")

var errNoSubscribedEvent = errors.New("no subscribed event provided")
