package storage

import (
	"gopkg.in/mgo.v2"
)

var DB *mgo.Database

func InitDB() error {
	session, err := mgo.Dial("mongodb://longchat:123456@127.0.0.1:27017/longchat")
	if err != nil {
		return err
	}
	err = session.Ping()
	if err != nil {
		return err
	}
	session.SetMode(mgo.Monotonic, true)

	DB = session.DB("longchat")
	return nil
}

func CloseDB() {
	DB.Session.Close()
}
