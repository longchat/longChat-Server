package dto

type UserInfo struct {
	NickName  string
	Avatar    string
	Introduce string
}
type CreateUserReq struct {
	UserName string
	PassWord string
	Captcha  string
}

type UpdateInfoReq struct {
	UserInfo
}

type GetUserInfoRsp struct {
	BaseRsp
	Data UserInfo
}
