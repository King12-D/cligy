package cache

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, data []byte)
	Delete(key string)
	Clear()
}
