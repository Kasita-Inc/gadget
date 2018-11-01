package storage

import (
	"reflect"

	"github.com/Kasita-Inc/gadget/errors"
)

// ---------- Error Types ---------
const (
	// NoEntryForKey storage error msg
	NoEntryForKey string = "no entry for key"
)

type memoryProvider struct {
	shards map[string]Provider
	values map[string]interface{}
}

// NewMemoryStorage for storing and retrieving key values.
func NewMemoryStorage() Provider {
	return &memoryProvider{values: make(map[string]interface{}), shards: make(map[string]Provider)}
}

// Shard returns a new instance of the storage provider with a seperate key space from this
// storage provider. The shard should be accessible by calling this function with the same
// shard name.
func (mp *memoryProvider) Shard(name string) (Provider, error) {
	shard, ok := mp.shards[name]
	if !ok {
		shard = NewMemoryStorage()
		mp.shards[name] = shard
	}
	return shard, nil
}

func (mp *memoryProvider) Read(key string, target interface{}) error {
	obj, ok := mp.values[key]
	if !ok {
		return errors.New(NoEntryForKey)
	}
	objValue := reflect.ValueOf(obj)
	// We are putting the value in the pointer obj into the target
	// 1. Get the value of target which is ptr/instance/slice to something we don't know about
	targetValue := reflect.ValueOf(target)
	if objValue.Kind() != targetValue.Kind() {
		return errors.New("target (%d) and value (%d) are not of the same 'Kind'", targetValue.Kind(),
			objValue.Kind())
	}
	// 2. Get the value of obj which is ptr/instance/slice of the unknown thing
	// We could support Slice, Array, and Struct as well, but we don't need it right now.
	switch objValue.Kind() {
	case reflect.Interface:
		// interface works the same as pointer
		fallthrough
	case reflect.Ptr:
		// 3. Set the target value Elem to the objects Elem and we are done
		targetValue.Elem().Set(objValue.Elem())
	default:
		return errors.New("reflect.Kind %d is not supported", objValue.Kind())
	}
	return nil
}

func (mp *memoryProvider) Write(key string, target interface{}) error {
	mp.values[key] = target
	return nil
}

func (mp *memoryProvider) Exists(key string) bool {
	_, ok := mp.values[key]
	return ok
}

func (mp *memoryProvider) Delete(key string) error {
	delete(mp.values, key)
	return nil
}
