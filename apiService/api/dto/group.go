package dto

type GetGroupListRsp struct {
	BaseRsp
	Data GetGroupListData
}

type GetGroupListData struct {
	Groups []Group
}

type Group struct {
	Id        string
	Title     string
	Logo      string
	Introduce string
	OrderIdx  string
}

type GroupDetail struct {
	Id        string
	Title     string
	Logo      string
	Introduce string
	OrderIdx  string
	Members   []UserInfo
}

type GetGroupDetailRsp struct {
	BaseRsp
	Data GetGroupDetailData
}

type GetGroupDetailData struct {
	Group GroupDetail
}
