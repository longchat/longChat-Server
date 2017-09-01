package api

import (
	"fmt"

	"github.com/kataras/iris"
	"github.com/longchat/longChat-Server/apiService/api/dto"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/common/util"
	"github.com/longchat/longChat-Server/graphService/graph"
	"github.com/longchat/longChat-Server/idService/generator"
	"github.com/longchat/longChat-Server/storageService/storage"
)

type UserApi struct {
	idGen       *generator.IdGenerator
	store       *storage.Storage
	serverAddrs []string
}

func (ua *UserApi) RegisterRoute(framework *iris.Application) {
	users := framework.Party("/users")
	users.Post("", ua.createUser)
	users.Put("/{id:long}", ua.updateInfo)
	users.Get("/{id:long}", ua.getInfo)
	users.Get("/{id:long}/serveraddr", ua.getserverAddr)
}

func (ua *UserApi) getserverAddr(c iris.Context) {
	uid, _ := c.Params().GetInt64("id")
	clusterId, err := graph.GetClusterByUserId(uid)
	if err != nil {
		log.ERROR.Printf("get user  cluster id from graph failed!err:=%v\n", uid, err)
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(dto.InternalErrRsp())
		return
	}

	id := clusterId % len(ua.serverAddrs)
	userRsp := dto.GetUserServerAddrRsp{BaseRsp: *dto.SuccessRsp()}
	userRsp.Data.Addr = ua.serverAddrs[id]
	c.JSON(&userRsp)
}

func (ua *UserApi) getInfo(c iris.Context) {
	uid, _ := c.Params().GetInt64("id")
	user, err := ua.store.GetUserById(uid)
	if err != nil {
		log.ERROR.Printf("get usser(%d) from storage failed!err:=%v\n", uid, err)
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(dto.InternalErrRsp())
		return
	}
	userRsp := dto.GetUserInfoRsp{BaseRsp: *dto.SuccessRsp()}
	userRsp.Data.User = dto.UserInfo{
		Id:        fmt.Sprintf("%d", user.Id),
		NickName:  user.NickName,
		Avatar:    user.Avatar,
		Introduce: user.Introduce,
	}
	c.JSON(&userRsp)
}

func (ua *UserApi) updateInfo(c iris.Context) {
	var infoReq dto.UpdateInfoReq
	err := c.ReadJSON(&infoReq)
	if err != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(dto.PostDataErrRsp("UpdateInfoReq"))
		return
	}
	uid, _ := c.Params().GetInt64("id")
	err = ua.store.UpdateUserInfo(uid, infoReq.NickName, infoReq.Avatar, infoReq.Introduce)
	if err != nil {
		log.ERROR.Printf("UpdateUserInfo from storage failed!err:=%v\n", err)
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(dto.InternalErrRsp())
		return
	}
	c.JSON(dto.SuccessRsp())
}

func (ua *UserApi) createUser(c iris.Context) {
	var userReq dto.CreateUserReq
	err := c.ReadJSON(&userReq)
	if err != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(dto.PostDataErrRsp("CreateUserReq"))
		return
	}
	id, err := ua.idGen.Generate(generator.GenerateReq_User)
	if err != nil {
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(dto.InternalErrRsp())
		return
	}
	salt := util.RandomString(8)
	hashedPassword := getHashedPassword(userReq.PassWord, salt)
	err = ua.store.CreateUser(id, userReq.UserName, hashedPassword, salt, c.RemoteAddr())
	if err != nil {
		log.ERROR.Printf("CreateUser from storage failed!err:=%v\n", err)
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(dto.InternalErrRsp())
		return
	}
	c.JSON(dto.SuccessRsp())
}
