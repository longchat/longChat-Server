package storage

import (
	"time"

	"github.com/go-xorm/xorm"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/storageService/storage/schema"
)

func createUser(db *xorm.Engine, id int64, userName string, password string, salt string, lastLoginIp string) error {
	user := schema.User{Id: id, UserName: userName, Password: password, Salt: salt, LastLoginIp: lastLoginIp}
	_, err := db.InsertOne(&user)
	if err != nil {
		log.ERROR.Printf("insert user(%v) into db failed!err:=%v\n", user, err)
		return err
	}
	return nil
}

func updateUserInfo(db *xorm.Engine, id int64, nickName string, avatar string, intro string) error {
	user := schema.User{Id: id, NickName: nickName, Avatar: avatar, Introduce: intro}
	_, err := db.Where("Id=?", id).Update(&user)
	if err != nil {
		log.ERROR.Printf("update user(%d) from db failed!err:=%v\n", id, err)
		return err
	}
	return nil
}

func getUserByUserName(db *xorm.Engine, userName string) (*schema.User, error) {
	var user schema.User
	_, err := db.Where("UserName=?", userName).Get(&user)
	if err != nil {
		log.ERROR.Printf("findone user(%s) from db failed!err:=%v\n", userName, err)
		return nil, err
	}
	return &user, nil
}

func getUserById(db *xorm.Engine, id int64) (*schema.User, error) {
	var user schema.User
	_, err := db.Where("Id=?", id).Get(&user)
	if err != nil {
		log.ERROR.Printf("findone user(%s) from db failed!err:=%v\n", id, err)
		return nil, err
	}
	return &user, nil
}

func addUserGroup(db *xorm.Engine, groupId int64, userId int64) error {
	userGroup := schema.GroupMembers{
		UserId:  userId,
		GroupId: groupId,
		JoinTs:  time.Now().Unix(),
	}
	_, err := db.InsertOne(&userGroup)
	if err != nil {
		log.ERROR.Printf("add a member to a group failed!err:=%v\n", err)
		return err
	}
	return nil
}
func getUsersByIds(db *xorm.Engine, ids []int64) ([]schema.User, error) {
	var users []schema.User
	_, err := db.In("Id", ids).Get(&users)
	if err != nil {
		log.ERROR.Printf("find group by id(%d) from db failed!err:=%v\n", ids, err)
		return nil, err
	}
	return users, nil
}
