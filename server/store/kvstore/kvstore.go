package kvstore

// KVStore interface extended to support basic Get and Set operations.
type KVStore interface {
	Get(key string) (interface{}, error)
	Set(key string, value string) error
}
