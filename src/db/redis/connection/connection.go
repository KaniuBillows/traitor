package redis

import (
	"io"
	"sync"
	"time"
	"traitor/db/util/wait"
)

const (
	// NormalCli is client with user
	NormalCli = iota
	// ReplicationRecvCli is fake client with replication master
	ReplicationRecvCli
)

type blockingBuff chan []byte

func (bf blockingBuff) Write(b []byte) (n int, err error) {
	bf <- b
	return len(b), nil
}
func (bf blockingBuff) Read(b []byte) (n int, err error) {
	var rec = <-bf
	copy(b, rec)
	return len(rec), nil
}

// Connection represents a connection with a redis-cli
type Connection struct {
	receiveBuf blockingBuff
	sendBuf    blockingBuff

	// waiting until protocol finished
	waitingReply wait.Wait
	// lock while server sending response
	mu sync.Mutex

	// queued commands for `multi`
	multiState bool
	queue      [][][]byte
	watching   map[string]uint32
	txErrors   []error

	// selected database
	selectedDB int
	role       int32
}

//// RemoteAddr returns the remote network address
//func (c *Connection) RemoteAddr() net.Addr {
//	return c.conn.RemoteAddr()
//}

// Close disconnect with the client
func (c *Connection) Close() error {
	c.waitingReply.WaitTimeOut(10 * time.Second)
	return nil
}

// NewConn creates Connection instance
func NewConn() *Connection {
	return &Connection{
		receiveBuf: make(chan []byte, 256),
		sendBuf:    make(chan []byte, 256),
	}
}
func (c *Connection) GetClientSendBuff() io.Writer {
	return c.receiveBuf
}
func (c *Connection) GetClientReceiveBuff() io.Reader {
	return c.sendBuf
}
func (c *Connection) GetReceiveBuff() io.Reader {
	return c.receiveBuf
}
func (c *Connection) GetSendBuff() io.Writer {
	return c.sendBuf
}

// Write sends response to client
func (c *Connection) Write(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
	}()

	return c.sendBuf.Write(b)
}

//
//// SetPassword stores password for authentication
//func (c *Connection) SetPassword(password string) {
//	c.password = password
//}
//
//// GetPassword get password for authentication
//func (c *Connection) GetPassword() string {
//	return c.password
//}

// InMultiState tells is connection in an uncommitted transaction
func (c *Connection) InMultiState() bool {
	return c.multiState
}

// SetMultiState sets transaction flag
func (c *Connection) SetMultiState(state bool) {
	if !state { // reset data when cancel multi
		c.watching = nil
		c.queue = nil
	}
	c.multiState = state
}

// GetQueuedCmdLine returns queued commands of current transaction
func (c *Connection) GetQueuedCmdLine() [][][]byte {
	return c.queue
}

// EnqueueCmd  enqueues command of current transaction
func (c *Connection) EnqueueCmd(cmdLine [][]byte) {
	c.queue = append(c.queue, cmdLine)
}

// AddTxError stores syntax error within transaction
func (c *Connection) AddTxError(err error) {
	c.txErrors = append(c.txErrors, err)
}

// GetTxErrors returns syntax error within transaction
func (c *Connection) GetTxErrors() []error {
	return c.txErrors
}

// ClearQueuedCmds clears queued commands of current transaction
func (c *Connection) ClearQueuedCmds() {
	c.queue = nil
}

// GetRole returns role of connection, such as connection with master
func (c *Connection) GetRole() int32 {
	if c == nil {
		return NormalCli
	}
	return c.role
}

func (c *Connection) SetRole(r int32) {
	c.role = r
}

// GetWatching returns watching keys and their version code when started watching
func (c *Connection) GetWatching() map[string]uint32 {
	if c.watching == nil {
		c.watching = make(map[string]uint32)
	}
	return c.watching
}

// GetDBIndex returns selected database
func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

// SelectDB selects a database
func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}
