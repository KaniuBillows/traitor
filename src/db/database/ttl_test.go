package database

import (
	"testing"
	"time"
	"traitor/db/interface/database"
)

func TestDB_Expire(t *testing.T) {
	var db = makeDB()
	const key = "hello"

	db.PutEntity(key, &database.DataEntity{Data: nil})
	db.Expire(key, time.Now().Add(time.Second*1))
	expired := db.IsExpired(key)
	if expired {
		panic("the key should not be expired.")
	}
	time.Sleep(time.Second * 2)
	existed := db.Exists(key)
	if existed == true {
		panic("the key should has been removed.")
	}
}
