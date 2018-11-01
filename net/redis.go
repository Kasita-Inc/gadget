package net

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis"

	"github.com/Kasita-Inc/gadget/errors"
	"github.com/Kasita-Inc/gadget/stringutil"
)

const prodSlug = "cache.amazonaws.com"

// InvalidRedisAddressError is returned in an invalid address is passed to the Redis Client initializer.
type InvalidRedisAddressError struct {
	Address       string
	InternalError error
}

// NewInvalidRedisAddressError for the passed address and underlying error.
func NewInvalidRedisAddressError(address string, err error) error {
	return &InvalidRedisAddressError{Address: address, InternalError: err}
}

func (err *InvalidRedisAddressError) Error() string {
	return fmt.Sprintf("invalid redis address '%s': %s", err.Address, err.InternalError)
}

// EmptyAddressError  is returned when an empty string is passed as the address
type EmptyAddressError struct{ trace []string }

func (err *EmptyAddressError) Error() string {
	return "address was empty"
}

// Trace returns the stack trace for the error
func (err *EmptyAddressError) Trace() []string {
	return err.trace
}

// NewEmptyAddressError instantiates a EmptyAddressError with a stack trace
func NewEmptyAddressError() errors.TracerError {
	return &EmptyAddressError{trace: errors.GetStackTrace()}
}

// NameSpacedRedis wraps certain redis commands with a namespace around key names to avoid collisions
type NameSpacedRedis interface {
	// Exists verifies the key is in Redis
	Exists(keys ...string) *redis.IntCmd
	// Del removes a key/value from Redis
	Del(keys ...string) *redis.IntCmd
	// Set add a key/value with an optional expiration (0 is no expiration) into Redis
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	// Get a value for a given key from Redis
	Get(key string) *redis.StringCmd
	// SAdd adds the specified members to the set stored at key
	SAdd(key string, members ...interface{}) *redis.IntCmd
	// SCard returns the cardinality of the set stored at key
	SCard(key string) *redis.IntCmd
	// SRem removes the specified members from the set stored at key
	SRem(key string, members ...interface{}) *redis.IntCmd
	// SRandMember When called with just the key argument, return a random element from the set value stored at key
	SRandMember(key string) *redis.StringCmd
	// SPop removes and returns one or more random elements from the set value store at key
	SPop(key string) *redis.StringCmd
	// FlushAll deletes all the keys of all the existing databases, not just the currently selected one
	FlushAll() *redis.StatusCmd
}

type nameSpacedRedis struct {
	client    redis.Cmdable
	namespace string
	mutex     sync.Mutex
}

func (nsr *nameSpacedRedis) namespaceKey(key string) string {
	return fmt.Sprintf("%s_%s", nsr.namespace, key)
}

func (nsr *nameSpacedRedis) Exists(keys ...string) *redis.IntCmd {
	for i, key := range keys {
		keys[i] = nsr.namespaceKey(key)
	}
	return nsr.client.Exists(keys...)
}

func (nsr *nameSpacedRedis) Del(keys ...string) *redis.IntCmd {
	for i, key := range keys {
		keys[i] = nsr.namespaceKey(key)
	}
	nsr.mutex.Lock()
	defer nsr.mutex.Unlock()
	return nsr.client.Del(keys...)
}

func (nsr *nameSpacedRedis) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	nsr.mutex.Lock()
	defer nsr.mutex.Unlock()
	return nsr.client.Set(nsr.namespaceKey(key), value, expiration)
}

func (nsr *nameSpacedRedis) Get(key string) *redis.StringCmd {
	return nsr.client.Get(nsr.namespaceKey(key))
}

func (nsr *nameSpacedRedis) SAdd(key string, members ...interface{}) *redis.IntCmd {
	nsr.mutex.Lock()
	defer nsr.mutex.Unlock()
	return nsr.client.SAdd(nsr.namespaceKey(key), members...)
}

func (nsr *nameSpacedRedis) SCard(key string) *redis.IntCmd {
	nsr.mutex.Lock()
	defer nsr.mutex.Unlock()
	return nsr.client.SCard(nsr.namespaceKey(key))
}

func (nsr *nameSpacedRedis) SRem(key string, members ...interface{}) *redis.IntCmd {
	nsr.mutex.Lock()
	defer nsr.mutex.Unlock()
	return nsr.client.SRem(nsr.namespaceKey(key), members...)
}

func (nsr *nameSpacedRedis) SRandMember(key string) *redis.StringCmd {
	return nsr.client.SRandMember(nsr.namespaceKey(key))
}

func (nsr *nameSpacedRedis) SPop(key string) *redis.StringCmd {
	nsr.mutex.Lock()
	defer nsr.mutex.Unlock()
	return nsr.client.SPop(nsr.namespaceKey(key))
}

func (nsr *nameSpacedRedis) FlushAll() *redis.StatusCmd {
	return nsr.client.FlushAll()
}

// NewRedisClient that is appropriate for the passed address.
func NewRedisClient(address string, namespace string) (NameSpacedRedis, error) {
	inst := &nameSpacedRedis{namespace: namespace}
	if stringutil.IsWhiteSpace(address) {
		return nil, NewInvalidRedisAddressError(address, NewEmptyAddressError())
	}
	host := address
	if strings.Contains(address, ":") {
		host = strings.Split(address, ":")[0]
	}
	_, err := net.LookupHost(host)
	if nil != err {
		return nil, NewInvalidRedisAddressError(address, err)
	}

	if strings.Contains(address, prodSlug) {
		inst.client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{address},
		})
	} else {
		inst.client = redis.NewClient(&redis.Options{
			Addr: address,
		})
	}
	return inst, nil
}
