package lock

import (
	"sort"
	"sync"
)

const prime32 = uint32(16777619)

type Locks struct {
	table []*sync.RWMutex
}

/*
spread
Choose the Shard.
*/
func (locks *Locks) spread(hashCode uint32) uint32 {
	if locks == nil {
		panic("locks can not be nil")
	}
	tableSize := uint32(len(locks.table))
	return (tableSize - 1) & hashCode
}

/*
Make
create an instance of the locks
*/
func Make(tableSize int) *Locks {
	table := make([]*sync.RWMutex, tableSize)
	for i := 0; i < tableSize; i++ {
		table[i] = &sync.RWMutex{}
	}
	return &Locks{
		table: table,
	}
}

func (locks *Locks) Lock(key string) {
	index := locks.spread(hash(key))
	mu := locks.table[index]
	mu.Lock()
}

func (locks *Locks) UnLock(key string) {
	index := locks.spread(hash(key))
	mu := locks.table[index]
	mu.Unlock()
}

func (locks *Locks) RLock(key string) {
	index := locks.spread(hash(key))
	mu := locks.table[index]
	mu.RLock()
}

func (locks *Locks) RUnlock(key string) {
	index := locks.spread(hash(key))
	mu := locks.table[index]
	mu.RUnlock()
}

/*
toLockIndices

find the locks corresponding to these keys,
and sort the locks indices to ensure that each coroutine locks in the same order
*/
func (locks *Locks) toLockIndices(keys []string, reverse bool) []uint32 {
	indexMap := make(map[uint32]bool)
	for _, key := range keys {
		index := locks.spread(hash(key))
		indexMap[index] = true
	}
	indices := make([]uint32, 0, len(indexMap))
	for index := range indexMap {
		indices = append(indices, index)
	}
	sort.Slice(indices, func(i, j int) bool {
		if !reverse {
			return indices[i] < indices[j]
		}
		return indices[i] > indices[j]
	})
	return indices
}

/*
RWLocks

Batch lock
*/
func (locks *Locks) RWLocks(writeKeys []string, readKeys []string) {
	keys := append(writeKeys, readKeys...)
	indices := locks.toLockIndices(keys, false)

	writeIndexSet := make(map[uint32]struct{})
	for _, wKey := range writeKeys {
		idx := locks.spread(hash(wKey))
		writeIndexSet[idx] = struct{}{}
	}

	for _, index := range indices {
		_, w := writeIndexSet[index]
		mu := locks.table[index]
		if w {
			mu.Lock()
		} else {
			mu.RLock()
		}
	}
}

func (locks *Locks) RWUnLocks(writeKeys []string, readKeys []string) {
	keys := append(writeKeys, readKeys...)
	indices := locks.toLockIndices(keys, true)

	writeIndexSet := make(map[uint32]struct{})
	for _, wKey := range writeKeys {
		idx := locks.spread(hash(wKey))
		writeIndexSet[idx] = struct{}{}
	}

	for _, index := range indices {
		_, w := writeIndexSet[index]
		mu := locks.table[index]
		if w {
			mu.Unlock()
		} else {
			mu.RUnlock()
		}
	}
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
