package dict

import (
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
)

const prime32 = uint32(16777619)

/*
	ConcurrentDict

the key would be hashed into one of the table
in the same table,keys use the same lock.
the Concurrency depends on the number of locks
*/
type ConcurrentDict struct {
	table      []*Shard
	count      int32
	shardCount int
}

type Shard struct {
	mp    map[string]any
	mutex sync.RWMutex
}

/*
computeCapacity
*/
func computeCapacity(param int) (size int) {
	if param <= 16 {
		return 16
	}
	n := param - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	if n < 0 {
		return math.MaxInt32
	} else {
		return int(n + 1)
	}
}

/*
MakeConcurrentDict
init Concurrent Dict. allocate memory for each table.
*/
func MakeConcurrentDict(shardCount int) *ConcurrentDict {
	shardCount = computeCapacity(shardCount)
	table := make([]*Shard, shardCount)
	for i := 0; i < shardCount; i++ {
		table[i] = &Shard{
			mp: make(map[string]any),
		}
	}
	return &ConcurrentDict{
		table:      table,
		count:      0,
		shardCount: shardCount,
	}
}

/*
spread
Choose the Shard.
*/
func (dict *ConcurrentDict) spread(hashCode uint32) uint32 {
	if dict == nil {
		panic("dict can not be nil")
	}
	tableSize := uint32(len(dict.table))
	return (tableSize - 1) & hashCode
}

/*
Get the Shard according to index.
*/
func (dict *ConcurrentDict) getShard(index uint32) *Shard {
	if dict == nil {
		panic("dict is nil")
	}
	return dict.table[index]
}

/*
Get

try to get the value based on the key.
*/
func (dict *ConcurrentDict) Get(key string) (val any, exists bool) {
	if dict == nil {
		panic("dict is nil")
	}
	hashCode := hash(key)          // calculate hash.
	index := dict.spread(hashCode) // get shard index.
	shard := dict.getShard(index)
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()
	val, exists = shard.mp[key]
	return
}

/*
Len

get the length of the table.
*/
func (dict *ConcurrentDict) Len() int {
	if dict == nil {
		panic("dict is nil")
	}
	return int(atomic.LoadInt32(&dict.count))
}

func (dict *ConcurrentDict) addCount() {
	if dict == nil {
		panic("dict is nil")
	}
	atomic.AddInt32(&dict.count, 1)
}
func (dict *ConcurrentDict) reduceCount() {
	if dict == nil {
		panic("dict is nil")
	}
	atomic.AddInt32(&dict.count, -1)
}

/*
Put

put the key and value into the dict.
if the key is already exists , would replace the value and return 0.
otherwise the key and value would be added into the dict directly and return 1.
*/
func (dict *ConcurrentDict) Put(key string, val any) (result int) {
	if dict == nil {
		panic("dict is nil")
	}
	hashCode := hash(key)
	index := dict.spread(hashCode) // get shard index.
	shard := dict.getShard(index)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if _, ok := shard.mp[key]; ok {
		shard.mp[key] = val
		return 0
	} else {
		shard.mp[key] = val
		dict.addCount()
		return 1
	}
}

/*
PutIfAbsent

if the key doesn't exist,put it into the map.
*/
func (dict *ConcurrentDict) PutIfAbsent(key string, val any) (result int) {
	if dict == nil {
		panic("dict is nil")
	}
	hashCode := hash(key)
	index := dict.spread(hashCode)
	shard := dict.getShard(index)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if _, ok := shard.mp[key]; ok {
		return 0
	}
	shard.mp[key] = val
	dict.addCount()
	return 1
}

func (dict *ConcurrentDict) PutIfExists(key string, val any) (result int) {
	if dict == nil {
		panic("dict is nil")
	}
	hashCode := hash(key)
	index := dict.spread(hashCode)
	shard := dict.getShard(index)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()

	if _, ok := shard.mp[key]; ok {
		shard.mp[key] = val
		return 1
	}
	return 0
}

/*
Remove

remove the key and value from the dict.
if the key is exists, the key would be removed and return true
otherwise return false.
*/
func (dict *ConcurrentDict) Remove(key string) int {
	if dict == nil {
		panic("dict is nil")
	}
	hashCode := hash(key)
	index := dict.spread(hashCode) // get shard index.
	shard := dict.getShard(index)
	shard.mutex.Lock()
	defer shard.mutex.Unlock()
	if _, ok := shard.mp[key]; ok {
		delete(shard.mp, key)
		return 1
	} else {
		return 0
	}
}

func (dict *ConcurrentDict) ForEach(consumer Consumer) {
	if dict == nil {
		panic("dict is nil")
	}

	for _, shard := range dict.table {
		shard.mutex.RLock()
		func() {
			defer shard.mutex.RUnlock()
			for key, value := range shard.mp {
				continues := consumer(key, value)
				if continues == false {
					return
				}
			}
		}()
	}
}

func (shard *Shard) RandomKey() string {
	if shard == nil {
		panic("shard is nil")
	}
	shard.mutex.RLock()
	defer shard.mutex.RUnlock()

	for key := range shard.mp {
		return key
	}
	return ""
}

func (dict *ConcurrentDict) RandomKeys(limit int) []string {
	size := dict.Len()
	if limit >= size {
		return dict.Keys()
	}
	shardCount := len(dict.table)

	result := make([]string, limit)
	for i := 0; i < limit; {
		shard := dict.getShard(uint32(rand.Intn(shardCount)))
		if shard == nil {
			continue
		}
		key := shard.RandomKey()
		if key != "" {
			result[i] = key
			i++
		}
	}
	return result
}

func (dict *ConcurrentDict) RandomDistinctKeys(limit int) []string {
	size := dict.Len()
	if limit >= size {
		return dict.Keys()
	}

	shardCount := len(dict.table)
	result := make(map[string]bool)
	for len(result) < limit {
		shardIndex := uint32(rand.Intn(shardCount))
		shard := dict.getShard(shardIndex)
		if shard == nil {
			continue
		}
		key := shard.RandomKey()
		if key != "" {
			result[key] = true
		}
	}
	var arr = make([]string, len(result))
	var i = 0
	for k := range result {
		arr[i] = k
	}
	return arr
}
func (dict *ConcurrentDict) Keys() []string {
	keys := make([]string, dict.Len())
	i := 0

	dict.ForEach(func(key string, val any) bool {
		if i < len(keys) {
			keys[i] = key
			i++
		} else {
			keys = append(keys, key)
		}
		return true
	})
	return keys
}
func (dict *ConcurrentDict) Clear() {
	*dict = *MakeConcurrentDict(dict.shardCount)
}

/*
hash
FNV to calculate hash value
*/
func hash(key string) uint32 {
	hash := uint32(2166136261)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
