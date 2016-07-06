package dto

type LoginReq struct {
	UserName string
	Password string
}

type LoginRsp struct {
	BaseRsp
	Data LoginData
}

type LoginData struct {
	User  UserInfo
	Token string
}
