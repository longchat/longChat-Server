package api

import (
	"github.com/longchat/longChat-Server/idService/generator"
	"github.com/longchat/longChat-Server/storageService/storage"

	"github.com/kataras/iris"
)

func Iint(framework *iris.Framework, idGen *generator.IdGenerator, store *storage.Storage) {
	framework.Config.Render.Rest.Gzip = true
	framework.Config.Render.Template.Gzip = true

	ua := UserApi{idGen: idGen, store: store}
	ua.RegisterRoute(framework)
}
