package main

import (
	"fmt"
	"longChat-Server/idGenerator/generator"
	"longChat-Server/storage"
)

func main() {
	err := storage.InitDB()
	if err != nil {
		fmt.Println("init db failed!err:=%v", err)
	}
	createUser()
}

func createUser() {
	storage.CreateUser(generator.Generate(generator.UserTypeId), "longxia", "longxia110", 1)
}
