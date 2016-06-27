package main

import (
	"fmt"
	apiService "longChat-Server/apiService/service"
	idService "longChat-Server/idService/service"
)

func main() {
	idService.InitIdService(false)
	id, err := apiService.CreateUser()
	fmt.Println(id, "error:", err)
}
