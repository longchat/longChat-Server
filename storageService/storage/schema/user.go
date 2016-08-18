package schema

//用户
type User struct {
	Id          int64
	UserName    string
	Password    string
	Salt        string
	NickName    string
	Avatar      string
	Introduce   string
	LastLoginIp string
}
