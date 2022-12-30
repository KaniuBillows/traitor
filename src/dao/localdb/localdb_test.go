package localdb

import (
	"testing"
	"time"
	"traitor/dao/model"
)

func TestName(t *testing.T) {
	var mp = make(map[string]any)
	var s []byte = nil
	mp[model.LastExecTime] = string(s)

	tm, err := time.Parse("2017-08-30 16:40:41", mp[model.LastExecTime].(string))
	if err != nil {
		t.Error(err)
	}
	t.Log(tm)
}
