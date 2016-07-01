package schema


//用户
type User struct {
	Id            int64 `bson:"_id"`
	UserName      string
	Password      string
	Salt          string
	NickName      string
	Avatar        string
	Introduce     string
	LastLoginIp   string
	JoinedGroups   []struct{
		Id int64
		JoinTs int
		LastReadMsgId int64
	}`bson:",omitempty"`
}

