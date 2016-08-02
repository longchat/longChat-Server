package storage

import (
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/storageService/storage/schema"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func getGroupsByOrderIdx(db *mgo.Database, orderIdx int64, limit int) ([]schema.Group, error) {
	var groups []schema.Group
	err := db.C("Group").Find(bson.M{"_id": bson.M{"$gt": orderIdx}}).Sort("_id").Limit(limit).All(&groups)
	if err != nil {
		log.ERROR.Printf("find groups from db failed!err:=%v", err)
		return nil, nil
	}
	return groups, nil
}
func getGroupById(db *mgo.Database, id int64) (*schema.Group, error) {
	var group schema.Group
	err := db.C("Group").FindId(id).One(&group)
	if err != nil && err != mgo.ErrNotFound {
		log.ERROR.Printf("find group by id(%d) from db failed!err:=%v\n", id, err)
		return nil, err
	}
	return &group, nil
}

func addGroupMember(db *mgo.Database, groupId int64, userId int64) error {
	err := db.C("Group").Update(bson.M{"_id": groupId}, bson.M{"$push": bson.M{"members": userId}})
	if err != nil && err != mgo.ErrNotFound {
		log.ERROR.Printf("add a member to a group failed!err:=%v\n", err)
		return err
	}
	return nil
}
