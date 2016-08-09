package api

import (
	"fmt"
	"net/http"

	"github.com/kataras/iris"
	"github.com/longchat/longChat-Server/apiService/api/dto"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/common/util"
	"github.com/longchat/longChat-Server/idService/generator"
	"github.com/longchat/longChat-Server/storageService/storage"
)

type UserApi struct {
	idGen *generator.IdGenerator
	store *storage.Storage
}

func (ua *UserApi) RegisterRoute(framework *iris.Framework) {
	users := framework.Party("/users")
	users.Post("", ua.createUser)
	users.Put("/:id", ua.updateInfo)
	users.Get("/:id", ua.getInfo)
}

func (ua *UserApi) getInfo(c *iris.Context) {
	uid, err := c.ParamInt64("id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ParameterErrRsp("id"))
		return
	}
	user, err := ua.store.GetUserById(uid)
	if err != nil {
		log.ERROR.Printf("get usser(%d) from storage failed!err:=%v\n", uid, err)
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}
	userRsp := dto.GetUserInfoRsp{BaseRsp: *dto.SuccessRsp()}
	userRsp.Data.User = dto.UserInfo{
		Id:        fmt.Sprintf("%d", user.Id),
		NickName:  user.NickName,
		Avatar:    user.Avatar,
		Introduce: user.Introduce,
	}
	c.JSON(http.StatusOK, &userRsp)

}

func (ua *UserApi) updateInfo(c *iris.Context) {
	var infoReq dto.UpdateInfoReq
	err := c.ReadJSON(&infoReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.PostDataErrRsp("UpdateInfoReq"))
		return
	}
	uid, err := c.ParamInt64("id")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ParameterErrRsp("id"))
		return
	}
	err = ua.store.UpdateUserInfo(uid, infoReq.NickName, infoReq.Avatar, infoReq.Introduce)
	if err != nil {
		log.ERROR.Printf("UpdateUserInfo from storage failed!err:=%v\n", err)
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}
	c.JSON(http.StatusOK, dto.SuccessRsp())
}

func (ua *UserApi) createUser(c *iris.Context) {
	var userReq dto.CreateUserReq
	err := c.ReadJSON(&userReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.PostDataErrRsp("CreateUserReq"))
		return
	}
	id, err := ua.idGen.Generate(generator.GenerateReq_User)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}
	salt := util.RandomString(8)
	hashedPassword := getHashedPassword(userReq.PassWord, salt)
	err = ua.store.CreateUser(id, userReq.UserName, hashedPassword, salt, c.RemoteAddr())
	if err != nil {
		log.ERROR.Printf("CreateUser from storage failed!err:=%v\n", err)
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}
	c.JSON(http.StatusOK, dto.SuccessRsp())
}
