package container

import "sync"

// MutexMap represents a concurrent safe map
type MutexMap struct {
	mut    sync.RWMutex
	values map[interface{}]interface{}
}

// NewMutexMap returns a new instance of a mutex map
func NewMutexMap() *MutexMap {
	return &MutexMap{
		values: make(map[interface{}]interface{}),
	}
}

// Get returns the element stored with provided key
func (mm *MutexMap) Get(key interface{}) (interface{}, bool) {
	mm.mut.RLock()
	val, ok := mm.values[key]
	mm.mut.RUnlock()

	return val, ok
}

// Insert adds the (key, val) tuple if the key does not exist
// returns true operation succeeded
func (mm *MutexMap) Insert(key interface{}, val interface{}) bool {
	mm.mut.Lock()

	_, ok := mm.values[key]
	if !ok {
		mm.values[key] = val
	}

	mm.mut.Unlock()

	return !ok
}

// Set stores the (key, val) tuple, rewriting data if existing
func (mm *MutexMap) Set(key interface{}, val interface{}) {
	mm.mut.Lock()
	mm.values[key] = val
	mm.mut.Unlock()
}

// Remove deletes a (key, val) tuple (if exists)
func (mm *MutexMap) Remove(key interface{}) {
	mm.mut.Lock()
	delete(mm.values, key)
	mm.mut.Unlock()
}

// Len returns the inner map size
func (mm *MutexMap) Len() int {
	mm.mut.RLock()
	defer mm.mut.RUnlock()

	return len(mm.values)
}

// Keys returns all stored keys. The order is not guaranteed
func (mm *MutexMap) Keys() []interface{} {
	mm.mut.RLock()
	keys := make([]interface{}, 0, len(mm.values))
	for k := range mm.values {
		keys = append(keys, k)
	}
	mm.mut.RUnlock()

	return keys
}

// Values returns all stored values. The order is not guaranteed
func (mm *MutexMap) Values() []interface{} {
	mm.mut.RLock()
	values := make([]interface{}, 0, len(mm.values))
	for _, value := range mm.values {
		values = append(values, value)
	}
	mm.mut.RUnlock()

	return values
}
