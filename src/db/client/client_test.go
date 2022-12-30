package client_test

import (
	"context"
	"fmt"
	"testing"
	"time"
	"traitor/db/protocol"
	"traitor/db/startup"
	utils "traitor/db/util"
)

func TestSet(t *testing.T) {
	cnn, closer := startup.Startup(context.TODO())
	client := cnn()
	//client.Send(utils.ToCmdLine("SET", "KKK1", "HELLO WORLD"))

	var res = client.Send(utils.ToCmdLine("GET", "KKK1"))

	bulk := res.(*protocol.BulkReply)
	value := string(bulk.Arg)
	fmt.Println(value)
	if value != "HELLO WORLD" {
		t.Fail()
	}
	time.Sleep(time.Second * 1)
	closer()
}
