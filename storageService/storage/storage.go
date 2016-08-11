package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/storageService/storage/schema"
	"gopkg.in/mgo.v2"
	"gopkg.in/redis.v4"
)

type Storage struct {
	db    *mgo.Database
	redis *redis.Client

	sessionPrefix string
}

func (s *Storage) initRedis() error {
	redisAddr, err := config.GetConfigString(consts.RedisAddress)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.RedisAddress, err))
	}
	redisPsw, err := config.GetConfigString(consts.RedisPassword)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.RedisPassword, err))
	}
	redisDb, err := config.GetConfigInt(consts.RedisDb)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.RedisDb, err))
	}
	s.redis = redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		Password:     redisPsw, // no password set
		DB:           redisDb,  // use default DB
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  10 * time.Second,
		PoolSize:     128,
	})
	_, err = s.redis.Ping().Result()
	if err != nil {
		return errors.New(fmt.Sprintf("ping redis failed!err:=%v", err))
	}

	s.sessionPrefix, err = config.GetConfigString(consts.SessionPrefix)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.SessionPrefix, err))
	}
	return nil
}

func (s *Storage) initMongo() error {
	dbName, err := config.GetConfigString(consts.MongoDbName)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.MongoDbName, err))
	}
	dbAddr, err := config.GetConfigString(consts.MongoDbAddr)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.MongoDbAddr, err))
	}
	dbUser, err := config.GetConfigString(consts.MongoDbUser)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.MongoDbUser, err))
	}
	dbPsw, err := config.GetConfigString(consts.MongoDbPsw)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.MongoDbPsw, err))
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

func NewStorage() (*Storage, error) {
	s := new(Storage)
	err := s.initMongo()
	if err != nil {
		return nil, err
	}
	err = s.initRedis()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) Close() {
	if s.db.Session != nil {
		s.db.Session.Close()
	}
}

func (s *Storage) AddUserGroup(userId int64, groupId int64) error {
	return addUserGroup(s.db, groupId, userId)
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

func (s *Storage) AddGroupMember(groupId int64, userId int64) error {
	return addGroupMember(s.db, groupId, userId)
}

func (s *Storage) GetGroupsByOrderIdx(orderIdx int64, limit int) ([]schema.Group, error) {
	return getGroupsByOrderIdx(s.db, orderIdx, limit)
}

func (s *Storage) GetGroupById(id int64) (*schema.Group, error) {
	return getGroupById(s.db, id)
}

func (s *Storage) GetSessionValue(key string) (map[string]interface{}, error) {
	return getSessionValue(s.redis, s.sessionPrefix+key)
}

func (s *Storage) CreateUserMessage(id int64, from int64, to int64, content string, msgType int) error {
	return createUserMessage(s.db, id, from, to, content, msgType)
}

func (s *Storage) CreateGroupMessage(id int64, from int64, groupId int64, content string, msgType int) error {
	return createGroupMessage(s.db, id, groupId, from, content, msgType)
}

func (s *Storage) MarkUserMessageRead(id int64) error {
	return markUserMessageRead(s.db, id)
}
