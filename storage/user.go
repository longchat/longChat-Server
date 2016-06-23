package storage

import (
	"fmt"
	"longChat-Server/schema"
)

func CreateUser(id int64, userName string, password string, lastLoginIp int) {
	err := DB.C("User").Insert(&schema.User{Id: id, UserName: userName, Password: password, LastLoginIp: lastLoginIp})
	if err != nil {
		fmt.Println("insert user failed!err:=", err)
	}
}
