package schema


//最近联系人
type Contacts struct{
	Id            int64 `bson:"_id"`
	From          string
	To 	      string
	UnreadMsgs    []struct {
		Id      int64
		Content string
		Type    string
	} `bson:",omitempty"`
}
