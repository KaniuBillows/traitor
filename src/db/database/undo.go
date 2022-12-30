package database

import (
	"strconv"
	"time"
	"traitor/db/aof"
	"traitor/db/protocol"
	utils "traitor/db/util"
)

func rollbackFirstKey(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	return rollbackGivenKeys(db, key)
}

func rollbackHashFields(db *DB, key string, fields ...string) []CmdLine {
	var undoCmdLines [][][]byte
	dict, errReply := db.getAsDict(key)
	if errReply != nil {
		return nil
	}
	if dict == nil {
		undoCmdLines = append(undoCmdLines,
			utils.ToCmdLine("DEL", key),
		)
		return undoCmdLines
	}
	for _, field := range fields {
		entity, ok := dict.Get(field)
		if !ok {
			undoCmdLines = append(undoCmdLines,
				utils.ToCmdLine("HDEL", key, field),
			)
		} else {
			value, _ := entity.([]byte)
			undoCmdLines = append(undoCmdLines,
				utils.ToCmdLine("HSET", key, field, string(value)),
			)
		}
	}
	return undoCmdLines
}

func rollbackGivenKeys(db *DB, keys ...string) []CmdLine {
	var undoCmdLines [][][]byte
	for _, key := range keys {
		entity, ok := db.GetEntity(key)
		if !ok {
			undoCmdLines = append(undoCmdLines,
				utils.ToCmdLine("DEL", key),
			)
		} else {
			undoCmdLines = append(undoCmdLines,
				utils.ToCmdLine("DEL", key), // clean existed first
				aof.EntityToCmd(key, entity).Args,
				toTTLCmd(db, key).Args,
			)
		}
	}
	return undoCmdLines
}
func rollbackZSetFields(db *DB, key string, fields ...string) []CmdLine {
	var undoCmdLines [][][]byte
	zset, errReply := db.getAsSortedSet(key)
	if errReply != nil {
		return nil
	}
	if zset == nil {
		undoCmdLines = append(undoCmdLines,
			utils.ToCmdLine("DEL", key),
		)
		return undoCmdLines
	}
	for _, field := range fields {
		elem, ok := zset.Get(field)
		if !ok {
			undoCmdLines = append(undoCmdLines,
				utils.ToCmdLine("ZREM", key, field),
			)
		} else {
			score := strconv.FormatFloat(elem.Score, 'f', -1, 64)
			undoCmdLines = append(undoCmdLines,
				utils.ToCmdLine("ZADD", key, score, field),
			)
		}
	}
	return undoCmdLines
}

// undoSetChange rollbacks SADD and SREM command
func undoSetChange(db *DB, args [][]byte) []CmdLine {
	key := string(args[0])
	memberArgs := args[1:]
	members := make([]string, len(memberArgs))
	for i, mem := range memberArgs {
		members[i] = string(mem)
	}
	return rollbackSetMembers(db, key, members...)
}
func rollbackSetMembers(db *DB, key string, members ...string) []CmdLine {
	var undoCmdLines [][][]byte
	set, errReply := db.getAsSet(key)
	if errReply != nil {
		return nil
	}
	if set == nil {
		undoCmdLines = append(undoCmdLines,
			utils.ToCmdLine("DEL", key),
		)
		return undoCmdLines
	}
	for _, member := range members {
		ok := set.Has(member)
		if !ok {
			undoCmdLines = append(undoCmdLines,
				utils.ToCmdLine("SREM", key, member),
			)
		} else {
			undoCmdLines = append(undoCmdLines,
				utils.ToCmdLine("SADD", key, member),
			)
		}
	}
	return undoCmdLines
}
func toTTLCmd(db *DB, key string) *protocol.MultiBulkReply {
	raw, exists := db.ttlMap.Get(key)
	if !exists {
		// has no TTL
		return protocol.MakeMultiBulkReply(utils.ToCmdLine("PERSIST", key))
	}
	expireTime, _ := raw.(time.Time)
	timestamp := strconv.FormatInt(expireTime.UnixNano()/1000/1000, 10)
	return protocol.MakeMultiBulkReply(utils.ToCmdLine("PEXPIREAT", key, timestamp))
}
