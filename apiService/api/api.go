package api

import (
	"fmt"
	"unsafe"

	"github.com/kataras/iris"
	"github.com/longchat/longChat-Server/idService/generator"
	"github.com/longchat/longChat-Server/storageService/storage"
)

func Iint(framework *iris.Framework, idGen *generator.IdGenerator, store *storage.Storage) {
	framework.Config.Render.Rest.Gzip = true
	framework.Config.Render.Template.Gzip = true
	ua := UserApi{idGen: idGen, store: store}
	fmt.Println(unsafe.Pointer(&ua))
	ua.RegisterRoute(framework)
}
