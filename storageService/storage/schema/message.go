package schema


//为每一个group创建一个collection存放message，所以不需要groupId
type GroupMessage struct {
	Id      int64 `bson:"_id"`
	Content string
	Type    string
	From    string
}

//默认每隔15天新建一个collection
type UserMessage struct {
	Id      int64 `bson:"_id"`
	From    string
	To      string
	Content string
	Type    string
}
