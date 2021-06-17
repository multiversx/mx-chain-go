package view

import "errors"

// ErrNilPresenterInterface will be returned when a nil PresenterInterface is passed as parameter
var ErrNilPresenterInterface = errors.New("nil presenter interface")

// ErrNilChanNodeIsStarting signals that a nil channel for node starting has been provided
var ErrNilChanNodeIsStarting = errors.New("nil node starting channel")

// ErrNilGrid will be returned when a nil grid is returned
var ErrNilGrid = errors.New("nil grid")

// ErrInvalidRefreshTimeInMilliseconds signals that an invalid time in milliseconds was provided
var ErrInvalidRefreshTimeInMilliseconds = errors.New("invalid refresh time in milliseconds")
