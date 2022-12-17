package database

import (
	"strconv"
	"traitor/db/interface/database"
	"traitor/db/interface/redis"
	"traitor/db/protocol"
	"traitor/db/struct/list"
	utils "traitor/db/util"
)

func (db *DB) getAsList(key string) (list.List, protocol.ErrorReply) {
	entity, ok := db.GetEntity(key)
	if !ok {
		return nil, nil
	}
	l, ok := entity.Data.(list.List)
	if !ok {
		return nil, &protocol.WrongTypeErrReply{}
	}
	return l, nil
}

func (db *DB) getOrInitList(key string) (l list.List, isNew bool, errReply protocol.ErrorReply) {
	l, errReply = db.getAsList(key)
	if errReply != nil {
		return nil, false, errReply
	}
	isNew = false
	if l == nil {
		l = list.NewQuickList()
		db.PutEntity(key, &database.DataEntity{
			Data: l,
		})
		isNew = true
	}
	return l, isNew, nil
}

// execLIndex gets element of list at given list
func execLIndex(db *DB, args [][]byte) redis.Reply {
	// parse args
	key := string(args[0])
	index64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	index := int(index64)

	// get entity
	l, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if l == nil {
		return &protocol.NullBulkReply{}
	}

	size := l.Len() // assert: size > 0
	if index < -1*size {
		return &protocol.NullBulkReply{}
	} else if index < 0 {
		index = size + index
	} else if index >= size {
		return &protocol.NullBulkReply{}
	}

	val, _ := l.Get(index).([]byte)
	return protocol.MakeBulkReply(val)
}

// execLLen gets length of list
func execLLen(db *DB, args [][]byte) redis.Reply {
	// parse args
	key := string(args[0])

	l, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if l == nil {
		return protocol.MakeIntReply(0)
	}

	size := int64(l.Len())
	return protocol.MakeIntReply(size)
}

// execLPop removes the first element of list, and return it
func execLPop(db *DB, args [][]byte) redis.Reply {
	// parse args
	key := string(args[0])

	// get data
	l, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if l == nil {
		return &protocol.NullBulkReply{}
	}

	val, _ := l.Remove(0).([]byte)
	if l.Len() == 0 {
		db.Remove(key)
	}
	db.addAof(utils.ToCmdLine3("lpop", args...))
	return protocol.MakeBulkReply(val)
}

var lPushCmd = []byte("LPUSH")

func undoLPop(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	l, errReply := db.getAsList(key)
	if errReply != nil {
		return nil
	}
	if l == nil || l.Len() == 0 {
		return nil
	}
	element, _ := l.Get(0).([]byte)
	return []CmdLine{
		{
			lPushCmd,
			args[0],
			element,
		},
	}
}

// execLPush inserts element at head of list
func execLPush(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	values := args[1:]

	// get or init entity
	l, _, errReply := db.getOrInitList(key)
	if errReply != nil {
		return errReply
	}

	// insert
	for _, value := range values {
		l.Insert(0, value)
	}

	db.addAof(utils.ToCmdLine3("lpush", args...))
	return protocol.MakeIntReply(int64(l.Len()))
}

func undoLPush(_ *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	count := len(args) - 1
	cmdLines := make([]CmdLine, 0, count)
	for i := 0; i < count; i++ {
		cmdLines = append(cmdLines, utils.ToCmdLine("LPOP", key))
	}
	return cmdLines
}

// execLPushX inserts element at head of list, only if list exists
func execLPushX(db *DB, args [][]byte) redis.Reply {
	key := string(args[0])
	values := args[1:]

	// get or init entity
	l, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if l == nil {
		return protocol.MakeIntReply(0)
	}

	// insert
	for _, value := range values {
		l.Insert(0, value)
	}
	db.addAof(utils.ToCmdLine3("lpushx", args...))
	return protocol.MakeIntReply(int64(l.Len()))
}

// execLRange gets elements of list in given range
func execLRange(db *DB, args [][]byte) redis.Reply {
	// parse args
	key := string(args[0])
	start64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	start := int(start64)
	stop64, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	stop := int(stop64)

	// get data
	l, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if l == nil {
		return &protocol.EmptyMultiBulkReply{}
	}

	// compute index
	size := l.Len() // assert: size > 0
	if start < -1*size {
		start = 0
	} else if start < 0 {
		start = size + start
	} else if start >= size {
		return &protocol.EmptyMultiBulkReply{}
	}
	if stop < -1*size {
		stop = 0
	} else if stop < 0 {
		stop = size + stop + 1
	} else if stop < size {
		stop = stop + 1
	} else {
		stop = size
	}
	if stop < start {
		stop = start
	}

	// assert: start in [0, size - 1], stop in [start, size]
	slice := l.Range(start, stop)
	result := make([][]byte, len(slice))
	for i, raw := range slice {
		bytes, _ := raw.([]byte)
		result[i] = bytes
	}
	return protocol.MakeMultiBulkReply(result)
}

// execLRem removes element of list at specified index
func execLRem(db *DB, args [][]byte) redis.Reply {
	// parse args
	key := string(args[0])
	count64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	count := int(count64)
	value := args[2]

	// get data entity
	l, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if l == nil {
		return protocol.MakeIntReply(0)
	}

	var removed int
	if count == 0 {
		removed = l.RemoveAllByVal(func(a interface{}) bool {
			return utils.Equals(a, value)
		})
	} else if count > 0 {
		removed = l.RemoveByVal(func(a any) bool {
			return utils.Equals(a, value)
		}, count)
	} else {
		removed = l.ReverseRemoveByVal(func(a interface{}) bool {
			return utils.Equals(a, value)
		}, -count)
	}

	if l.Len() == 0 {
		db.Remove(key)
	}
	if removed > 0 {
		db.addAof(utils.ToCmdLine3("lrem", args...))
	}

	return protocol.MakeIntReply(int64(removed))
}

// execLSet puts element at specified index of list
func execLSet(db *DB, args [][]byte) redis.Reply {
	// parse args
	key := string(args[0])
	index64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return protocol.MakeErrReply("ERR value is not an integer or out of range")
	}
	index := int(index64)
	value := args[2]

	// get data
	l, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if l == nil {
		return protocol.MakeErrReply("ERR no such key")
	}

	size := l.Len() // assert: size > 0
	if index < -1*size {
		return protocol.MakeErrReply("ERR index out of range")
	} else if index < 0 {
		index = size + index
	} else if index >= size {
		return protocol.MakeErrReply("ERR index out of range")
	}

	l.Set(index, value)
	db.addAof(utils.ToCmdLine3("lset", args...))
	return &protocol.OkReply{}
}

func undoLSet(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	index64, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return nil
	}
	index := int(index64)
	l, errReply := db.getAsList(key)
	if errReply != nil {
		return nil
	}
	if l == nil {
		return nil
	}
	size := l.Len() // assert: size > 0
	if index < -1*size {
		return nil
	} else if index < 0 {
		index = size + index
	} else if index >= size {
		return nil
	}
	value, _ := l.Get(index).([]byte)
	return []CmdLine{
		{
			[]byte("LSET"),
			args[0],
			args[1],
			value,
		},
	}
}

// execRPop removes last element of list then return it
func execRPop(db *DB, args [][]byte) redis.Reply {
	// parse args
	key := string(args[0])

	// get data
	l, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if l == nil {
		return &protocol.NullBulkReply{}
	}

	val, _ := l.RemoveLast().([]byte)
	if l.Len() == 0 {
		db.Remove(key)
	}
	db.addAof(utils.ToCmdLine3("rpop", args...))
	return protocol.MakeBulkReply(val)
}

var rPushCmd = []byte("RPUSH")

func undoRPop(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	l, errReply := db.getAsList(key)
	if errReply != nil {
		return nil
	}
	if l == nil || l.Len() == 0 {
		return nil
	}
	element, _ := l.Get(l.Len() - 1).([]byte)
	return []CmdLine{
		{
			rPushCmd,
			args[0],
			element,
		},
	}
}

func prepareRPopLPush(args [][]byte) ([]string, []string) {
	return []string{
		string(args[0]),
		string(args[1]),
	}, nil
}

// execRPopLPush pops last element of list-A then insert it to the head of list-B
func execRPopLPush(db *DB, args [][]byte) redis.Reply {
	sourceKey := string(args[0])
	destKey := string(args[1])

	// get source entity
	sourceList, errReply := db.getAsList(sourceKey)
	if errReply != nil {
		return errReply
	}
	if sourceList == nil {
		return &protocol.NullBulkReply{}
	}

	// get dest entity
	destList, _, errReply := db.getOrInitList(destKey)
	if errReply != nil {
		return errReply
	}

	// pop and push
	val, _ := sourceList.RemoveLast().([]byte)
	destList.Insert(0, val)

	if sourceList.Len() == 0 {
		db.Remove(sourceKey)
	}

	db.addAof(utils.ToCmdLine3("rpoplpush", args...))
	return protocol.MakeBulkReply(val)
}

func undoRPopLPush(db *DB, args [][]byte) []CmdLine {
	sourceKey := string(args[0])
	l, errReply := db.getAsList(sourceKey)
	if errReply != nil {
		return nil
	}
	if l == nil || l.Len() == 0 {
		return nil
	}
	element, _ := l.Get(l.Len() - 1).([]byte)
	return []CmdLine{
		{
			rPushCmd,
			args[0],
			element,
		},
		{
			[]byte("LPOP"),
			args[1],
		},
	}
}

// execRPush inserts element at last of list
func execRPush(db *DB, args [][]byte) redis.Reply {
	// parse args
	key := string(args[0])
	values := args[1:]

	// get or init entity
	l, _, errReply := db.getOrInitList(key)
	if errReply != nil {
		return errReply
	}

	// put list
	for _, value := range values {
		l.Add(value)
	}
	db.addAof(utils.ToCmdLine3("rpush", args...))
	return protocol.MakeIntReply(int64(l.Len()))
}

func undoRPush(_ *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	count := len(args) - 1
	cmdLines := make([]CmdLine, 0, count)
	for i := 0; i < count; i++ {
		cmdLines = append(cmdLines, utils.ToCmdLine("RPOP", key))
	}
	return cmdLines
}

// execRPushX inserts element at last of list only if list exists
func execRPushX(db *DB, args [][]byte) redis.Reply {
	if len(args) < 2 {
		return protocol.MakeErrReply("ERR wrong number of arguments for 'rpush' command")
	}
	key := string(args[0])
	values := args[1:]

	// get or init entity
	l, errReply := db.getAsList(key)
	if errReply != nil {
		return errReply
	}
	if l == nil {
		return protocol.MakeIntReply(0)
	}

	// put l
	for _, value := range values {
		l.Add(value)
	}
	db.addAof(utils.ToCmdLine3("rpushx", args...))

	return protocol.MakeIntReply(int64(l.Len()))
}

func init() {
	RegisterCommand("LPush", execLPush, writeFirstKey, undoLPush, -3, flagWrite)
	RegisterCommand("LPushX", execLPushX, writeFirstKey, undoLPush, -3, flagWrite)
	RegisterCommand("RPush", execRPush, writeFirstKey, undoRPush, -3, flagWrite)
	RegisterCommand("RPushX", execRPushX, writeFirstKey, undoRPush, -3, flagWrite)
	RegisterCommand("LPop", execLPop, writeFirstKey, undoLPop, 2, flagWrite)
	RegisterCommand("RPop", execRPop, writeFirstKey, undoRPop, 2, flagWrite)
	RegisterCommand("RPopLPush", execRPopLPush, prepareRPopLPush, undoRPopLPush, 3, flagWrite)
	RegisterCommand("LRem", execLRem, writeFirstKey, rollbackFirstKey, 4, flagWrite)
	RegisterCommand("LLen", execLLen, readFirstKey, nil, 2, flagReadOnly)
	RegisterCommand("LIndex", execLIndex, readFirstKey, nil, 3, flagReadOnly)
	RegisterCommand("LSet", execLSet, writeFirstKey, undoLSet, 4, flagWrite)
	RegisterCommand("LRange", execLRange, readFirstKey, nil, 4, flagReadOnly)
}
