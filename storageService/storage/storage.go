package storage

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kataras/iris/core/memstore"

	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/storageService/storage/schema"
	"gopkg.in/mgo.v2"
	"gopkg.in/redis.v4"
)

type Storage struct {
	db    *mgo.Database
	redis *redis.ClusterClient

	mysqlDb       [][]*xorm.Engine
	sessionPrefix string
}

func (s *Storage) initRedis() error {
	redisAddrs, err := config.GetConfigString(consts.RedisAddress)
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.RedisAddress, err))
	}
	addrslice := strings.Split(redisAddrs, ",")
	s.redis = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:          addrslice,
		MaxRedirects:   8,
		ReadOnly:       true,
		RouteByLatency: true,
		DialTimeout:    5 * time.Second,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		IdleTimeout:    10 * time.Second,
		PoolSize:       32,
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

func (s *Storage) initMysql() error {
	cluster1, err := config.GetConfigString("mysql.group.0.cluster.0")
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.SessionPrefix, err))
	}
	cluster2, err := config.GetConfigString("mysql.group.0.cluster.1")
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.SessionPrefix, err))
	}
	cluster3, err := config.GetConfigString("mysql.group.1.cluster.0")
	if err != nil {
		return errors.New(consts.ErrGetConfigFailed(consts.SessionPrefix, err))
	}
	s.mysqlDb = make([][]*xorm.Engine, 2)
	db, err := loadMysqlDb(cluster1)
	if err != nil {
		return nil
	}
	s.mysqlDb[0] = append(s.mysqlDb[0], db)
	db, err = loadMysqlDb(cluster2)
	if err != nil {
		return nil
	}
	s.mysqlDb[0] = append(s.mysqlDb[0], db)
	db, err = loadMysqlDb(cluster3)
	if err != nil {
		return nil
	}
	s.mysqlDb[1] = append(s.mysqlDb[0], db)
	return nil
}

func loadMysqlDb(addrs string) (*xorm.Engine, error) {
	dbuser, err := config.GetConfigString("mysql.db.user")
	if err != nil {
		return nil, err
	}
	dbpasswd, err := config.GetConfigString("mysql.db.passwd")
	if err != nil {
		return nil, err
	}
	dbaddr, err := config.GetConfigString(addrs)
	if err != nil {
		return nil, err
	}
	dbname, err := config.GetConfigString("mysql.db.name")
	if err != nil {
		return nil, err
	}
	connstr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4", dbuser, dbpasswd, dbaddr, dbname)
	db, err := xorm.NewEngine("mysql", connstr)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(50)
	db.SetMapper(core.SameMapper{})
	return db, nil
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
	err = s.initMysql()
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

func (s *Storage) getDb(groupId int64, id int64) *xorm.Engine {
	hash := id % 1024
	var engine *xorm.Engine
	if hash < 512 {
		engine = s.mysqlDb[0][0]
	} else {
		engine = s.mysqlDb[0][1]
	}
	return engine
}

type Slot struct {
	db  *xorm.Engine
	ids []int64
}

func (s *Storage) getDbsByIds(groupId int64, ids []int64) map[int]Slot {
	engineMap := make(map[int]Slot)
	for i := range ids {
		hash := ids[i] % 1024
		if hash < 512 {
			slot, isok := engineMap[0]
			if !isok {
				slot.db = s.mysqlDb[groupId][0]
			}
			slot.ids = append(slot.ids, ids[i])
			engineMap[0] = slot
		} else {
			slot, isok := engineMap[1]
			if !isok {
				slot.db = s.mysqlDb[groupId][1]
			}
			slot.ids = append(slot.ids, ids[i])
			engineMap[1] = slot
		}
	}
	return engineMap
}

func (s *Storage) AddUserGroup(userId int64, groupId int64) error {
	return addUserGroup(s.mysqlDb[1][0], groupId, userId)
}

func (s *Storage) UpdateUserInfo(id int64, nickName string, avatar string, intro string) error {
	return updateUserInfo(s.getDb(0, id), id, nickName, avatar, intro)
}

func (s *Storage) CreateUser(id int64, userName string, password string, salt string, lastLoginIp string) error {
	return createUser(s.getDb(0, id), id, userName, password, salt, lastLoginIp)
}

func (s *Storage) GetUserByUserName(userName string) (*schema.User, error) {
	for i := 0; i <= 1; i++ {
		user, err := getUserByUserName(s.mysqlDb[0][i], userName)
		if err != nil {
			return nil, err
		}
		if user.Id > 0 {
			return user, nil
		}
	}
	return nil, nil
}

func (s *Storage) GetUserById(id int64) (*schema.User, error) {
	return getUserById(s.getDb(0, id), id)
}

func (s *Storage) GetGroupsByOrderId(id int64, limit int) ([]schema.Group, error) {
	var groups []schema.Group
	for i := 0; i <= 1; i++ {
		gs, err := getGroupsByOrderIdx(s.mysqlDb[0][i], id, limit)
		if err != nil {
			return nil, err
		}
		groups = append(groups, gs...)
	}
	return groups, nil
}
func (s *Storage) GetUsersByIds(ids []int64) ([]schema.User, error) {
	dbs := s.getDbsByIds(0, ids)
	var users []schema.User
	for _, data := range dbs {
		g, err := getUsersByIds(data.db, data.ids)
		if err != nil {
			return nil, err
		}
		users = append(users, g...)
	}
	return users, nil
}
func (s *Storage) GetGroupById(id int64) (*schema.Group, error) {
	return getGroupById(s.getDb(0, id), id)
}

func (s *Storage) GetSessionValue(key string) (memstore.Store, error) {
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
