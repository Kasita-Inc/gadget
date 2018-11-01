package storage

// Provider is a struct that implements the functions necessary to work with
// the persist package.
type Provider interface {
	Shard(name string) (Provider, error)
	Read(key string, target interface{}) error
	Write(key string, target interface{}) error
	Exists(key string) bool
	Delete(key string) error
}
