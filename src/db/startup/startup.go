package startup

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"traitor/db/client"
	database2 "traitor/db/database"
	"traitor/db/interface/database"
	"traitor/db/protocol"
	"traitor/db/redis/parser"
	"traitor/logger"

	redis "traitor/db/redis/connection"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

type handler struct {
	activeConn    sync.Map
	db            database.DB
	closing       atomic.Bool
	blockingQueue chan *redis.Connection
}
type ConnectionFactory func() *client.Client
type Closer func()

func CreateConnection(h *handler) *client.Client {
	var cnn = redis.NewConn()
	var c = client.CreateClient(cnn)
	h.blockingQueue <- cnn
	return c
}
func (h *handler) close(cnn *redis.Connection) {
	_ = cnn.Close()
	h.activeConn.Delete(cnn)
}
func Startup(ctx context.Context) (ConnectionFactory, Closer) {
	var h = handler{
		activeConn:    sync.Map{},
		db:            database2.NewStandaloneServer(),
		blockingQueue: make(chan *redis.Connection),
	}
	go func() {
		for cnn := range h.blockingQueue {
			go h.handleServer(ctx, cnn)
		}
	}()

	var factory = func() *client.Client {
		return CreateConnection(&h)
	}
	var closer = func() {
		h.Close()
	}
	return factory, closer
}

func (h *handler) handleServer(ctx context.Context, cnn *redis.Connection) {
	var bf = cnn.GetReceiveBuff()
	ch := parser.ParseStream(bf)
	for payload := range ch {
		if payload.Err != nil {
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF {
				h.close(cnn)
				return
			}
			// protocol err
			errReply := protocol.MakeErrReply(payload.Err.Error())
			_, err := cnn.Write(errReply.ToBytes())
			if err != nil {
				h.close(cnn)
				logger.Info("connection closed: with err: " + err.Error())
				return
			}
			continue
		}
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := payload.Data.(*protocol.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk protocol")
			continue
		}
		result := h.db.Exec(cnn, r.Args)
		if result != nil {
			_, _ = cnn.Write(result.ToBytes())
		} else {
			_, _ = cnn.Write(unknownErrReplyBytes)
		}
	}
}
func (h *handler) Close() {
	h.closing.Store(true)
	h.activeConn.Range(func(key any, val any) bool {
		c := key.(*redis.Connection)
		_ = c.Close()
		return true
	})
	h.db.Close()
}
