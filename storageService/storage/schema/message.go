package schema



type GroupMessage struct {
	Id      int64 `bson:"_id"`
	Content string
	Type    int
	From    int64
	GroupId int64
}


type UserMessage struct {
	Id      int64 `bson:"_id"`
	From    int64
	To      int64
	Content string
	Type    int
	IsRead bool
}

