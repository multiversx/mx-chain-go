package vm

import "errors"

// ErrNilVmContainerMetaCreator signal that a nil vm container meta handler has been provided
var ErrNilVmContainerMetaCreator = errors.New("nil vm container meta creator")

// ErrNilVmContainerShardCreator signal that a nil vm container shard handler has been provided
var ErrNilVmContainerShardCreator = errors.New("nil vm container shard creator")
