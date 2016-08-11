package storage

import (
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/storageService/storage/schema"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func createGroupMessage(db *mgo.Database, id int64, groupId int64, from int64, content string, msgType int) error {
	msg := schema.GroupMessage{Id: id, From: from, GroupId: groupId, Content: content, Type: msgType}
	err := db.C("GroupMessage").Insert(&msg)
	if err != nil {
		log.ERROR.Printf("insert groupmessage(%v) into db failed!err:=%v\n", msg, err)
		return err
	}
	return nil
}

func createUserMessage(db *mgo.Database, id int64, from int64, to int64, content string, msgType int) error {
	msg := schema.UserMessage{Id: id, From: from, To: to, Content: content, Type: msgType, IsRead: false}
	err := db.C("UserMessage").Insert(&msg)
	if err != nil {
		log.ERROR.Printf("insert usermessage(%v) into db failed!err:=%v\n", msg, err)
		return err
	}
	return nil
}

func markUserMessageRead(db *mgo.Database, id int64) error {
	err := db.C("UserMessage").UpdateId(id, bson.M{"$set": bson.M{"isread": true}})
	if err != nil && err != mgo.ErrNotFound {
		log.ERROR.Printf("update user(%d) from db failed!err:=%v\n", id, err)
		return err
	}
	return nil
}
