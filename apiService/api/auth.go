package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
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

func (au *AuthApi) RegisterRoute(framework *iris.Application) {
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

func (au *AuthApi) login(c iris.Context) {
	var loginReq dto.LoginReq
	err := c.ReadJSON(&loginReq)
	if err != nil {
		c.StatusCode(iris.StatusBadRequest)
		c.JSON(dto.PostDataErrRsp("LoginReq"))
		return
	}
	user, err := au.store.GetUserByUserName(loginReq.UserName)
	if err != nil {
		log.ERROR.Printf("GetUserByUserName(%s) from storage failed!err:=%v\n", loginReq.UserName, err)
		c.StatusCode(iris.StatusInternalServerError)
		c.JSON(dto.InternalErrRsp())
		return
	}
	if getHashedPassword(loginReq.Password, user.Salt) != user.Password {
		c.StatusCode(iris.StatusUnauthorized)
		c.JSON(dto.PasswordNotMatchErrRsp())
		return
	}

	sessions.Get(c).Set("Id", user.Id)

	var userDto dto.UserInfo
	userDto.Id = fmt.Sprintf("%d", user.Id)
	userDto.Avatar = user.Avatar
	userDto.Introduce = user.Introduce
	userDto.NickName = user.NickName

	var rsp dto.LoginRsp
	rsp.Data.User = userDto
	c.JSON(&rsp)
}

func (au *AuthApi) logout(c iris.Context) {

}
