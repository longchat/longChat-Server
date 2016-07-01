package api

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/longchat/longChat-Server/common/log"

	"github.com/kataras/iris"
	"github.com/longchat/longChat-Server/apiService/api/dto"
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
	err = c.Session().Set("UserName", loginReq.UserName)
	if err != nil {
		log.ERROR.Printf("get session from redis failed!err:=%v\n", err)
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}
	fmt.Println(c.Session().Get("UserName").(string)+"  id:", c.Session().ID())
	c.JSON(http.StatusOK, dto.SuccessRsp())
}

func (au *AuthApi) logout(c *iris.Context) {

}
