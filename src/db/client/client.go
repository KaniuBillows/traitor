package client

import (
	"context"
	"io"
	"runtime/debug"
	"strings"
	"sync"
	"time"
	"traitor/db/interface/redis"
	"traitor/db/protocol"
	"traitor/db/redis/parser"
	"traitor/db/util/wait"
	"traitor/logger"
)

const chanSize = 256

type Client struct {
	receiveBuf  io.Reader
	sendBuf     io.Writer
	pendingReqs chan *Request
	waitingReqs chan *Request
	ctx         context.Context
	cancelFunc  context.CancelFunc
	writing     *sync.WaitGroup
}

func CreateClient(connection redis.Connection) *Client {
	c := &Client{
		receiveBuf:  connection.GetClientReceiveBuff(),
		sendBuf:     connection.GetClientSendBuff(),
		writing:     &sync.WaitGroup{},
		pendingReqs: make(chan *Request),
		waitingReqs: make(chan *Request),
	}
	go c.handleWrite()
	go c.handleRead()

	return c
}

type Request struct {
	id       uint64
	args     [][]byte
	reply    redis.Reply
	waiting  *wait.Wait
	err      error
	callback func(reply redis.Reply)
}

func (client *Client) Send(args [][]byte) redis.Reply {
	request := &Request{
		args:    args,
		waiting: &wait.Wait{},
	}
	request.waiting.Add(1)
	client.writing.Add(1)
	defer client.writing.Done()
	client.pendingReqs <- request
	timeout := request.waiting.WaitTimeOut(time.Second * 500)
	if timeout {
		return protocol.MakeErrReply("server time out")
	}
	if request.err != nil {
		return protocol.MakeErrReply("Request failed")
	}
	return request.reply
}
func (client *Client) SendAsync(args [][]byte, cb func(reply redis.Reply)) {
	request := &Request{
		args:     args,
		waiting:  &wait.Wait{},
		callback: cb,
	}
	request.waiting.Add(1)
	client.writing.Add(1)
	defer client.writing.Done()
	client.pendingReqs <- request
}

func (client *Client) handleWrite() {
	for req := range client.pendingReqs {
		client.doRequest(req)
	}
}

func (client *Client) doRequest(req *Request) {
	if req == nil || len(req.args) == 0 {
		return
	}
	re := protocol.MakeMultiBulkReply(req.args)
	bytes := re.ToBytes()
	var err error
	// retry
	for i := 0; i < 3; i++ {
		_, err = client.sendBuf.Write(bytes)
		if err == nil || (!strings.Contains(err.Error(), "timeout") &&
			!strings.Contains(err.Error(), "deadline exceeded")) {
			break
		}
	}
	if err == nil {
		client.waitingReqs <- req
	} else {
		req.err = err
		req.waiting.Done()
	}
}
func (client *Client) finishRequest(reply redis.Reply) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			logger.Error(err)
		}
	}()
	request := <-client.waitingReqs
	if request == nil {
		return
	}
	request.reply = reply
	if request.waiting != nil {
		request.waiting.Done()
	}
	if request.callback != nil {
		go request.callback(request.reply)
	}
}

func (client *Client) Close() {
	close(client.pendingReqs)

	client.writing.Wait()

	client.cancelFunc()
	close(client.waitingReqs)
}

func (client *Client) handleRead() error {
	ch := parser.ParseStream(client.receiveBuf)
	for payload := range ch {
		if payload.Err != nil {
			client.finishRequest(protocol.MakeErrReply(payload.Err.Error()))
			continue
		}
		client.finishRequest(payload.Data)
	}
	return nil
}
