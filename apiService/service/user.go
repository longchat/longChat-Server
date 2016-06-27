package service

import (
	"fmt"
	idService "longChat-Server/idService/service"

	"longChat-Server/storageService/storageService"
)

func main() {
	err := storage.InitDB()
	if err != nil {
		fmt.Println("init db failed!err:=%v", err)
	}
	createUser()
}

func createUser() {
	id, _ := idService.Generate(idService.GenerateReq_User)
	storage.CreateUser(id, "longxia", "longxia110", "127.0.0.1")
}

func CreateUser() (int64, error) {
	return idService.Generate(idService.GenerateReq_User)
}
