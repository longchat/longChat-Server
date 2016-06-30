package api

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/kataras/iris"
	"github.com/longchat/longChat-Server/apiService/api/dto"
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
	users.Post("", ua.CreateUser)
}

func getHashedPassword(raw string, salt string) string {
	sha := sha256.Sum256([]byte(raw + salt))
	return hex.EncodeToString(sha[:])
}

func (ua *UserApi) CreateUser(c *iris.Context) {
	var userReq dto.CreateUserReq
	err := c.ReadJSON(&userReq)
	if err != nil {
		c.JSON(400, dto.PostDataErrRsp("CreateUserReq"))
		return
	}
	id, err := ua.idGen.Generate(generator.GenerateReq_User)
	if err != nil {
		c.JSON(500, dto.InternalErrRsp())
		return
	}
	salt := util.RandomString(8)
	hashedPassword := getHashedPassword(userReq.PassWord, salt)
	err = ua.store.CreateUser(id, userReq.UserName, hashedPassword, salt, c.RemoteAddr())
	if err != nil {
		c.JSON(500, dto.InternalErrRsp())
		return
	}
	c.JSON(200, dto.SuccessResponse())
}
