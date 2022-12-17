package storeage

import (
	"context"
	"fmt"
	"testing"
	"traitor/db/startup"
)

func TestDel(t *testing.T) {
	cnnFac, cl := startup.Startup(context.TODO())

	cmd := encodeCmd("DEL", "k1")
	client := cnnFac()
	var reply = client.Send(cmd)

	fmt.Println(reply)
	cl()
}
