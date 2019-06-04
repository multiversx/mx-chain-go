package marshal

import "errors"

// ErrMarshallingProto is raised when the object does not implement proto.Message
var ErrMarshallingProto = errors.New("can not serialize the object")

// ErrUnmarshallingProto is raised when the object that needs to be unmarshaled does not implement proto.Message
var ErrUnmarshallingProto = errors.New("obj does not implement proto.Message")
