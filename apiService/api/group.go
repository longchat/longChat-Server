package api

import (
	"fmt"
	"net/http"

	"github.com/kataras/iris"
	"github.com/longchat/longChat-Server/apiService/api/dto"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/idService/generator"
	"github.com/longchat/longChat-Server/storageService/storage"
)

type GroupApi struct {
	idGen *generator.IdGenerator
	store *storage.Storage
}

func (ga *GroupApi) RegisterRoute(framework *iris.Framework) {
	users := framework.Party("/groups")
	users.Get("", ga.getGroupList)
	users.Get("/:id", ga.getGroupDetail)
	users.Post("/:id/members/:uid", ga.joinGroup)

}

func (ga *GroupApi) joinGroup(c *iris.Context) {
	gId, err := c.ParamInt64("id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ParameterErrRsp("id"))
		return
	}
	uId, err := c.ParamInt64("uid")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ParameterErrRsp("uid"))
		return
	}
	err = ga.store.AddGroupMember(gId, uId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}
	err = ga.store.AddUserGroup(uId, gId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}
	rsp := dto.SuccessRsp()
	c.JSON(200, &rsp)
}

func (ga *GroupApi) getGroupDetail(c *iris.Context) {
	gId, err := c.ParamInt64("id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ParameterErrRsp("id"))
		return
	}
	group, err := ga.store.GetGroupById(gId)
	if err != nil {
		log.ERROR.Printf("getGroupById(%d) from storage failed!err:=%v\n", gId, err)
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}
	users, err := ga.store.GetUsersById(group.Members)
	rsp := dto.GetGroupDetailRsp{BaseRsp: *dto.SuccessRsp()}
	rsp.Data.Group = dto.GroupDetail{
		Id:        fmt.Sprintf("%d", group.Id),
		Title:     group.Title,
		Logo:      group.Logo,
		Introduce: group.Introduce,
		OrderIdx:  fmt.Sprintf("%d", group.Id),
	}
	var usersDto []dto.UserInfo
	for i := range users {
		data := &users[i]
		userDto := dto.UserInfo{
			Id:        fmt.Sprintf("%d", data.Id),
			NickName:  data.NickName,
			Avatar:    data.Avatar,
			Introduce: data.Introduce,
		}
		usersDto = append(usersDto, userDto)
	}
	rsp.Data.Group.Members = usersDto
	c.JSON(200, &rsp)
}

func (ga *GroupApi) getGroupList(c *iris.Context) {
	orderIdx, err := c.URLParamInt64("orderidx")
	if err != nil {
		orderIdx = 0
	}
	limit, err := c.URLParamInt("limit")
	if err != nil {
		limit = 15
	}
	groups, err := ga.store.GetGroupsByOrderIdx(orderIdx, limit)
	if err != nil {
		log.ERROR.Printf("GetGroupsByOrderIdx from storage failed!err:=%v\n", err)
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}
	groupsDto := make([]dto.Group, limit)
	for i := range groups {
		data := &groups[i]
		groupDto := dto.Group{
			Id:        fmt.Sprintf("%d", data.Id),
			Title:     data.Title,
			Logo:      data.Logo,
			Introduce: data.Introduce,
			OrderIdx:  fmt.Sprintf("%d", data.Id),
		}
		groupsDto[i] = groupDto
	}
	rsp := dto.GetGroupListRsp{
		BaseRsp: *dto.SuccessRsp(),
	}
	rsp.Data.Groups = groupsDto[:len(groups)]
	c.JSON(200, &rsp)
}
