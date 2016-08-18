package schema

//群组
type Group struct {
	Id        int64 `bson:"_id"`
	Title     string
	Logo      string
	Members   []int64
	Introduce string
}

type GroupMembers struct {
	UserId        int64
	GroupId       int64
	JoinTs        int64
	LastReadMsgId int64
}
