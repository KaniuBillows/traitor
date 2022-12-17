package database

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"
	"traitor/db/aof"
	"traitor/db/config"
	"traitor/db/interface/database"
	"traitor/db/interface/redis"
	"traitor/db/protocol"
	"traitor/logger"
)

type DbServer struct {
	dbSet []*atomic.Value

	aofHandler *aof.Handler
}

// NewStandaloneServer creates a standalone redis server, with multi database and all other funtions
func NewStandaloneServer() *DbServer {
	server := &DbServer{}

	server.dbSet = make([]*atomic.Value, 1)
	for i := range server.dbSet {
		singleDB := makeDB()
		singleDB.index = i
		holder := &atomic.Value{}
		holder.Store(singleDB)
		server.dbSet[i] = holder
	}
	server.initAof()
	return server
}

func (ds *DbServer) selectDB(dbIndex int) (*DB, *protocol.StandardErrReply) {
	if dbIndex >= len(ds.dbSet) || dbIndex < 0 {
		return nil, protocol.MakeErrReply("ERR DB index is out of range")
	}
	return ds.dbSet[dbIndex].Load().(*DB), nil
}
func (ds *DbServer) mustSelectDB(dbIndex int) *DB {
	selectedDB, err := ds.selectDB(dbIndex)
	if err != nil {
		panic(err)
	}
	return selectedDB
}

func MakeTempServer() *DbServer {
	ds := &DbServer{}
	ds.dbSet = make([]*atomic.Value, config.Properties.Databases)
	for i := range ds.dbSet {
		holder := &atomic.Value{}
		holder.Store(makeBasicDB())
		ds.dbSet[i] = holder
	}
	return ds
}

// Exec executes command
// parameter `cmdLine` contains command and its arguments, for example: "set key value"
func (ds *DbServer) Exec(c redis.Connection, cmdLine [][]byte) (result redis.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &protocol.UnknownErrReply{}
		}
	}()
	// normal commands
	dbIndex := c.GetDBIndex()
	selectedDB, errReply := ds.selectDB(dbIndex)
	if errReply != nil {
		return errReply
	}
	return selectedDB.Exec(c, cmdLine)
}

// ForEach traverses all the keys in the given database
func (ds *DbServer) ForEach(dbIndex int, cb func(key string, data *database.DataEntity, expiration *time.Time) bool) {
	ds.mustSelectDB(dbIndex).ForEach(cb)
}

// ExecWithLock executes normal commands, invoker should provide locks
func (ds *DbServer) ExecWithLock(conn redis.Connection, cmdLine [][]byte) redis.Reply {
	db, errReply := ds.selectDB(conn.GetDBIndex())
	if errReply != nil {
		return errReply
	}
	return db.execWithLock(cmdLine)
}

// ExecMulti executes multi commands transaction Atomically and Isolated
func (ds *DbServer) ExecMulti(conn redis.Connection, watching map[string]uint32, cmdLines []CmdLine) redis.Reply {
	selectedDB, errReply := ds.selectDB(conn.GetDBIndex())
	if errReply != nil {
		return errReply
	}
	return selectedDB.ExecMulti(conn, watching, cmdLines)
}

// RWLocks lock keys for writing and reading
func (ds *DbServer) RWLocks(dbIndex int, writeKeys []string, readKeys []string) {
	ds.mustSelectDB(dbIndex).RWLocks(writeKeys, readKeys)
}

// RWUnLocks unlock keys for writing and reading
func (ds *DbServer) RWUnLocks(dbIndex int, writeKeys []string, readKeys []string) {
	ds.mustSelectDB(dbIndex).RWUnLocks(writeKeys, readKeys)
}

// GetUndoLogs return rollback commands
func (ds *DbServer) GetUndoLogs(dbIndex int, cmdLine [][]byte) []CmdLine {
	return ds.mustSelectDB(dbIndex).GetUndoLogs(cmdLine)
}

// GetDBSize returns keys count and ttl key count
func (ds *DbServer) GetDBSize(dbIndex int) (int, int) {
	db := ds.mustSelectDB(dbIndex)
	return db.data.Len(), db.ttlMap.Len()
}

// AfterClientClose does some clean after client close connection
func (ds *DbServer) AfterClientClose(c redis.Connection) {

}

// Close graceful shutdown database
func (ds *DbServer) Close() {
	// stop slaveStatus first
	if ds.aofHandler != nil {
		ds.aofHandler.Close()
	}
}
func (ds *DbServer) initAof() {
	aofHandler, err := aof.NewAOFHandler(ds, func() database.EmbedDB {
		return MakeTempServer()
	})
	if err != nil {
		panic(err)
	}
	ds.aofHandler = aofHandler
	for _, db := range ds.dbSet {
		singleDB := db.Load().(*DB)
		singleDB.addAof = func(line CmdLine) {
			ds.aofHandler.AddAof(singleDB.index, line)
		}
	}
}

// Ping the server
func Ping(db *DB, args [][]byte) redis.Reply {
	if len(args) == 0 {
		return &protocol.PongReply{}
	} else if len(args) == 1 {
		return protocol.MakeStatusReply(string(args[0]))
	} else {
		return protocol.MakeErrReply("ERR wrong number of arguments for 'ping' command")
	}
}

func init() {
	RegisterCommand("ping", Ping, noPrepare, nil, -1, flagReadOnly)
}
