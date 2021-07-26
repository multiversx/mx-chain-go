package logging

import "errors"

// ErrUnsupportedLogLifeSpanType is raised when an unsupported log life span type is provided
var ErrUnsupportedLogLifeSpanType = errors.New("unsupported log life span type")

// ErrCreateLogLifeSpanner is raised when a log life spanner cannot be created by the factory
var ErrCreateLogLifeSpanner = errors.New("create log lifespan failed")
