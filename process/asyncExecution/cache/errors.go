package cache

import "errors"

var (
	ErrCacheIsFull = errors.New("async execution headers's cache is full")
)
