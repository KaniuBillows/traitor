package aof

import (
	"io"
	"os"
	"strconv"
	"sync"
	"traitor/db/config"
	"traitor/db/interface/database"
	"traitor/db/protocol"
	redis "traitor/db/redis/connection"
	"traitor/db/redis/parser"
	utils "traitor/db/util"
	"traitor/logger"
)

type CmdLine = [][]byte

const (
	aofQueueSize = 1 << 16
)

type payload struct {
	cmdLine CmdLine
	dbIndex int
}

type Handler struct {
	db          database.EmbedDB
	currentDB   int
	tmpDBMaker  func() database.EmbedDB
	aofChan     chan *payload
	aofFile     *os.File
	aofFilename string
	aofFinished chan struct{}
	pausingAof  sync.RWMutex
}

// NewAOFHandler creates a new aof.Handler
func NewAOFHandler(db database.EmbedDB, tmpDbMaker func() database.EmbedDB) (*Handler, error) {
	handler := &Handler{}
	handler.aofFilename = config.Properties.AppendFilename
	handler.db = db
	handler.tmpDBMaker = tmpDbMaker
	handler.LoadAof(0)
	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	handler.aofChan = make(chan *payload, aofQueueSize)
	handler.aofFinished = make(chan struct{})
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

// handleAof listen aof channel and write into file
func (handler *Handler) handleAof() {
	// serialized execution
	handler.currentDB = 0
	for p := range handler.aofChan {
		handler.pausingAof.RLock() // prevent other goroutines from pausing aof
		if p.dbIndex != handler.currentDB {
			// select database
			data := protocol.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Warn(err)
				continue // skip this command
			}
			handler.currentDB = p.dbIndex
		}
		data := protocol.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Warn(err)
		}
		handler.pausingAof.RUnlock()
	}
	handler.aofFinished <- struct{}{}
}

// AddAof send command to aof goroutine through channel
func (handler *Handler) AddAof(dbIndex int, cmdLine CmdLine) {
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payload{
			cmdLine: cmdLine,
			dbIndex: dbIndex,
		}
	}
}

// LoadAof read aof file
func (handler *Handler) LoadAof(maxBytes int) {
	// delete aofChan to prevent write again
	aofChan := handler.aofChan
	handler.aofChan = nil
	defer func(aofChan chan *payload) {
		handler.aofChan = aofChan
	}(aofChan)

	file, err := os.Open(handler.aofFilename)

	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return
		}
		logger.Warn(err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			logger.Error("file close ERR")
		}
	}(file)

	var reader io.Reader
	if maxBytes > 0 {
		reader = io.LimitReader(file, int64(maxBytes))
	} else {
		reader = file
	}
	ch := parser.ParseStream(reader)
	fakeConn := &redis.Connection{} // only used for save dbIndex
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			}
			logger.Error("parse error: " + p.Err.Error())
			continue
		}
		if p.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := p.Data.(*protocol.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk protocol")
			continue
		}
		ret := handler.db.Exec(fakeConn, r.Args)
		if protocol.IsErrorReply(ret) {
			logger.Error("exec err", ret.ToBytes())
		}
	}
}

// Close gracefully stops aof persistence procedure
func (handler *Handler) Close() {
	if handler.aofFile != nil {
		close(handler.aofChan)
		<-handler.aofFinished // wait for aof finished
		err := handler.aofFile.Close()
		if err != nil {
			logger.Warn(err)
		}
	}
}
