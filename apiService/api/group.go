package api

import (
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
}
func (ga *GroupApi) getGroupList(c *iris.Context) {
	orderIdx, err := c.URLParamInt64("orderidx")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ParameterErrRsp("orderidx"))
		return
	}
	limit, err := c.URLParamInt("limit")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ParameterErrRsp("limit"))
		return
	}
	groups, err := ga.store.GetGroupsByOrderIdx(orderIdx, limit)
	if err != nil {
		log.ERROR.Printf("GetGroupsByOrderIdx from storage failed!err:=%v\n", err)
		c.JSON(http.StatusInternalServerError, dto.InternalErrRsp())
		return
	}

	println(groups)
}
