package cache

type Cache interface {
	Set(key string, value any)
	Get(key string) (any, bool)
	Delete(key string)
	Clear()
}
