package storage

import (
	"errors"
	"fmt"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"

	"gopkg.in/mgo.v2"
)

type Storage struct {
	db *mgo.Database
}

func (s *Storage) Init() error {
	dbName, err := config.GetConfigString(consts.StorageServiceDbName)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.StorageServiceDbName, err))
	}
	dbAddr, err := config.GetConfigString(consts.StorageServiceDbAddr)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.StorageServiceDbAddr, err))
	}
	dbUser, err := config.GetConfigString(consts.StorageServiceDbUser)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.StorageServiceDbUser, err))
	}
	dbPsw, err := config.GetConfigString(consts.StorageServiceDbPsw)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.StorageServiceDbPsw, err))
	}
	var session *mgo.Session
	if dbUser == "" {
		session, err = mgo.Dial(dbAddr)
	} else {
		session, err = mgo.Dial(fmt.Sprintf("mongodb://%s:%s@%s/%s", dbUser, dbPsw, dbAddr, dbName))
	}
	if err != nil {
		return err
	}
	err = session.Ping()
	if err != nil {
		return err
	}
	session.SetMode(mgo.Monotonic, true)

	s.db = session.DB(dbName)
	return nil
}

func (s *Storage) Close() {
	if s.db.Session != nil {
		s.db.Session.Close()
	}
}

func (s *Storage) CreateUser(id int64, userName string, password string, salt string, lastLoginIp string) error {
	return createUser(s.db, id, userName, password, salt, lastLoginIp)
}
