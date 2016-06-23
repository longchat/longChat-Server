package schema

//用户
type User struct {
	Id            int64 `bson:"_id"`
	UserName      string
	Password      string
	NickName      string
	Introduce     string
	LastReadMsgId int
	LastLoginIp   string
	UnreadUserMsg struct {
		Id      int64
		From    string
		Content string
		Type    string
	}
}

//群组
type Group struct {
	Id        int64 `bson:"_id"`
	Members   []string
	Introduce string
}

//为每一个group创建一个collection存放message，所以不需要groupId
type GroupMessage struct {
	Id      int64 `bson:"_id"`
	Content string
	Type    string
	From    string
}

//默认每隔15天开辟一个新的collection
type UserMessage struct {
	Id      int64 `bson:"_id"`
	From    string
	To      string
	Content string
	Type    string
}

func getTimeFromId(id int64) int64 {
	return id / 1000000
}
