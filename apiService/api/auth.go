package api

import (
	"github.com/kataras/iris"
	"github.com/longchat/longChat-Server/storageService/storage"
)

type AuthApi struct {
	store *storage.Storage
}

func (au *AuthApi) RegisterRoute(framework *iris.Framework) {
	framework.Post("/login", au.login)
	framework.Post("/logout", au.login)
}

func (au *AuthApi) login(c *iris.Context) {

}

func (au *AuthApi) logout(c *iris.Context) {

}
