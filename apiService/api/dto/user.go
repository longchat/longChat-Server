package dto

type UserInfo struct {
	Id        string
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
	Data GetUserInfoData
}

type GetUserInfoData struct {
	User UserInfo
}

type GetUserServerAddrRsp struct {
	BaseRsp
	Data GetUserServerAddrData
}
type GetUserServerAddrData struct {
	Addr string
}
