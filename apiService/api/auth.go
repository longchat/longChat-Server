package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/kataras/iris"
	"github.com/longchat/longChat-Server/apiService/api/dto"
	"github.com/longchat/longChat-Server/common/config"
	"github.com/longchat/longChat-Server/common/consts"
	"github.com/longchat/longChat-Server/common/log"
	"github.com/longchat/longChat-Server/common/util"
	"github.com/longchat/longChat-Server/storageService/storage"
)

type AuthApi struct {
	store *storage.Storage
}

func (au *AuthApi) RegisterRoute(framework *iris.Framework) {
	framework.Post("/login", au.login)
	framework.Post("/logout", au.login)
}

func getHashedPassword(raw string, salt string) string {
	sha := sha256.Sum256([]byte(raw + salt))
	return hex.EncodeToString(sha[:])
}

func newToken(id int64) (string, error) {
	privateToken, err := config.GetConfigString(consts.PrivateToken)
	if err != nil {
		log.ERROR.Printf(consts.ErrGetConfigFailed(consts.PrivateToken, err))
		return "", err
	}
	return util.NewToken(id, privateToken, time.Hour*12), nil
}

func (au *AuthApi) login(c *iris.Context) {
	var loginReq dto.LoginReq
	err := c.ReadJSON(&loginReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.PostDataErrRsp("LoginReq"))
		return
	}
	user, err := au.store.GetUserByUserName(loginReq.UserName)
	if err != nil {
		log.ERROR.Printf("GetUserByUserName(%s) from storage failed!err:=%v\n", loginReq.UserName, err)
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}
	if getHashedPassword(loginReq.Password, user.Salt) != user.Password {
		c.JSON(http.StatusUnauthorized, dto.PasswordNotMatchErrRsp())
		return
	}

	err = c.Session().Set("Id", user.Id)
	if err != nil {
		log.ERROR.Printf("get session from redis failed!err:=%v\n", err)
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}
	var userDto dto.UserInfo
	userDto.Id = fmt.Sprintf("%d", user.Id)
	userDto.Avatar = user.Avatar
	userDto.Introduce = user.Introduce
	userDto.NickName = user.NickName
	token, err := newToken(user.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}
	var rsp dto.LoginRsp
	rsp.Data.Token = token
	rsp.Data.User = userDto
	c.JSON(http.StatusOK, &rsp)
}

func (au *AuthApi) logout(c *iris.Context) {

}
