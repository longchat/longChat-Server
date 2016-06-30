package storage

import (
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/storageService/storage/schema"

	"gopkg.in/mgo.v2"
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
