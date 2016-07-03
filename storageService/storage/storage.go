package storage

import (
	"errors"
	"fmt"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/storageService/storage/schema"

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
func (s *Storage) UpdateUserInfo(id int64, nickName string, avatar string, intro string) error {
	return updateUserInfo(s.db, id, nickName, avatar, intro)
}
func (s *Storage) CreateUser(id int64, userName string, password string, salt string, lastLoginIp string) error {
	return createUser(s.db, id, userName, password, salt, lastLoginIp)
}

func (s *Storage) GetUserByUserName(userName string) (*schema.User, error) {
	return getUserByUserName(s.db, userName)
}

func (s *Storage) GetUserById(id int64) (*schema.User, error) {
	return getUserById(s.db, id)
}

func (s *Storage) GetUsersById(ids []int64) ([]schema.User, error) {
	return getUsersById(s.db, ids)
}

func (s *Storage) GetGroupsByOrderIdx(orderIdx int64, limit int) ([]schema.Group, error) {
	return getGroupsByOrderIdx(s.db, orderIdx, limit)
}

func (s *Storage) GetGroupById(id int64) (*schema.Group, error) {
	return getGroupById(s.db, id)
}
