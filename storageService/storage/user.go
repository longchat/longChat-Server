package storage

import (
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/storageService/storage/schema"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func createUser(db *mgo.Database, id int64, userName string, password string, salt string, lastLoginIp string) error {
	user := schema.User{Id: id, UserName: userName, Password: password, Salt: salt, LastLoginIp: lastLoginIp}
	err := db.C("User").Insert(&user)
	if err != nil {
		log.ERROR.Printf("insert user(%v) into db failed!err:=%v\n", user, err)
		return err
	}
	return nil
}

func updateUserInfo(db *mgo.Database, id int64, nickName string, avatar string, intro string) error {
	err := db.C("User").UpdateId(id, bson.M{"$set": bson.M{"nickname": nickName, "avatar": avatar, "introduce": intro}})
	if err != nil && err != mgo.ErrNotFound {
		log.ERROR.Printf("update user(%d) from db failed!err:=%v\n", id, err)
		return err
	}
	return nil
}

func getUserByUserName(db *mgo.Database, userName string) (*schema.User, error) {
	var user schema.User
	err := db.C("User").Find(bson.M{"username": userName}).One(&user)
	if err != nil && err != mgo.ErrNotFound {
		log.ERROR.Printf("findone user(%s) from db failed!err:=%v\n", userName, err)
		return nil, err
	}
	return &user, nil
}

func getUserById(db *mgo.Database, id int64) (*schema.User, error) {
	var user schema.User
	err := db.C("User").FindId(id).One(&user)
	if err != nil && err != mgo.ErrNotFound {
		log.ERROR.Printf("findone user(%d) from db failed!err:=%v\n", id, err)
		return nil, err
	}
	return &user, nil
}
