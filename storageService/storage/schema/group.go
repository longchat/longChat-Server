package schema

//群组
type Group struct {
	Id        int64 `bson:"_id"`
	Title     string
	Logo	  string
	Members   []int64
	Introduce string
}

