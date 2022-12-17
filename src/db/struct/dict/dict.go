package dict

type Consumer func(key string, val any) bool

type Dict interface {
	Get(key string) (val any, exists bool)
	Len() int
	// Put
	//if exists,cover the value and return 0,else return 1.
	Put(key string, val any) (result int)
	PutIfAbsent(key string, val any) (result int)
	PutIfExists(key string, val any) (result int)
	Remove(key string) (result int)
	ForEach(consumer Consumer)
	Keys() []string
	RandomKeys(limit int) []string
	RandomDistinctKeys(limit int) []string
	Clear()
}
