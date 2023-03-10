package redis

import "io"

// Connection represents a connection with redis client
type Connection interface {
	Write([]byte) (int, error)
	//SetPassword(string)
	//GetPassword() string

	//// client should keep its subscribing channels
	//Subscribe(channel string)
	//UnSubscribe(channel string)
	//SubsCount() int
	//GetChannels() []string
	GetClientSendBuff() io.Writer
	GetClientReceiveBuff() io.Reader
	GetReceiveBuff() io.Reader
	GetSendBuff() io.Writer

	// used for `Multi` command
	InMultiState() bool
	SetMultiState(bool)
	GetQueuedCmdLine() [][][]byte
	EnqueueCmd([][]byte)
	ClearQueuedCmds()
	GetWatching() map[string]uint32
	AddTxError(err error)
	GetTxErrors() []error

	// used for multi database
	GetDBIndex() int
	SelectDB(int)
	// returns role of conn, such as connection with client, connection with master node
	GetRole() int32
	SetRole(int32)
}
