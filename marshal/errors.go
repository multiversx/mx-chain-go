package marshal

import "errors"

// ErrMarshallingProto is raised when the object does not implement proto.Message
var ErrMarshallingProto = errors.New("can not serialize the object")

// ErrUnmarshallingProto is raised when the object that needs to be unmarshaled does not implement proto.Message
var ErrUnmarshallingProto = errors.New("obj does not implement proto.Message")

// ErrUnmarshallingBadSize is raised when the provided serialized data size exceeds the re-serialized data size
// plus an additional provided delta
var ErrUnmarshallingBadSize = errors.New("imput buffer too long")

// ErrUnknownMarshalizer signals that the specified marshalizer does not have a ready to use implementation
var ErrUnknownMarshalizer = errors.New("unknown marshalizer")
