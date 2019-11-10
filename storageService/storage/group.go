package storage

import (
	"xorm.io/xorm"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/storageService/storage/schema"
)

func getGroupById(db *xorm.Engine, id int64) (*schema.Group, error) {
	var group schema.Group
	_, err := db.Where("Id=?", id).Get(&group)
	if err != nil {
		log.ERROR.Printf("find group by id(%d) from db failed!err:=%v\n", id, err)
		return nil, err
	}
	return &group, nil
}

func getGroupsByIds(db *xorm.Engine, id []int64) ([]schema.Group, error) {
	var groups []schema.Group
	_, err := db.In("Id", id).Get(&groups)
	if err != nil {
		log.ERROR.Printf("find group by id(%d) from db failed!err:=%v\n", id, err)
		return nil, err
	}
	return groups, nil
}
func getGroupsByOrderIdx(db *xorm.Engine, id int64, limit int) ([]schema.Group, error) {
	var groups []schema.Group
	err := db.Where("Id <?", id).Limit(int(limit)).Find(&groups)
	if err != nil {
		return nil, err
	}
	return groups, nil
}
